package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

const (
	legacySkillPathUpgradeVersion = "data_0026_legacy_skill_path_upgrade"
	oldMythicTreeSlug             = "mythic-tragedy-apprenticeship"
	newStoryCraftTreeSlug         = "story-craft-track"
)

var retiredTreeSlugAliases = map[string]string{
	oldMythicTreeSlug: newStoryCraftTreeSlug,
}

var legacyTGOCodeAliases = map[string]string{
	"causal-clarity":     "story-causal-clarity",
	"scene-architecture": "story-scene-architecture",
	"prose-precision":    "story-prose-precision",
}

var legacySkillAliases = map[string]string{
	"tragic inevitability": "narrative clarity",
	"symbolic control":     "scene architecture",
	"symbolic discipline":  "image freshness",
	"mythic register":      "prose precision",
}

type migrationTreeTarget struct {
	Slug         string
	TreeID       int64
	TGOs         []domain.TGO
	ValidCodes   map[string]bool
	ByTitle      map[string][]domain.TGO
	BySkill      map[string][]domain.TGO
	SuffixToCode map[string]string
	SeedCodes    []string
}

type migrationContext struct {
	TargetsBySlug map[string]migrationTreeTarget
}

type enrollmentMigrationRow struct {
	EnrollmentID int64
	UserID       int64
	TreeID       int64
	TreeSlug     string
}

type activeCodeRow struct {
	Slot        int
	Code        string
	ActivatedAt time.Time
}

type completedCodeRow struct {
	Code        string
	CompletedAt time.Time
}

type currentFocusState struct {
	Focus         string
	Difficulty    int
	LastReviewID  int64
	HasCurriculum bool
	EnrollmentID  int64
}

func (s *Store) runOneTimeDataMigrations(ctx context.Context) error {
	return s.runDataMigration(ctx, legacySkillPathUpgradeVersion, s.migrateLegacySkillPaths)
}

func (s *Store) runDataMigration(ctx context.Context, version string, fn func(context.Context) error) error {
	var exists int
	if err := s.SQL.QueryRowContext(ctx, `
		SELECT COUNT(1) FROM schema_migrations WHERE version = ?
	`, version).Scan(&exists); err != nil {
		return err
	}
	if exists > 0 {
		return nil
	}
	if err := fn(ctx); err != nil {
		return err
	}
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO schema_migrations (version) VALUES (?)
	`, version)
	return err
}

func (s *Store) migrateLegacySkillPaths(ctx context.Context) (err error) {
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	meta, err := loadMigrationContext(ctx, tx)
	if err != nil {
		return err
	}
	needs, err := legacySkillPathUpgradeNeeded(ctx, tx, meta)
	if err != nil {
		return err
	}
	if !needs {
		return tx.Commit()
	}

	if err = remapGlobalSlugReferences(ctx, tx); err != nil {
		return err
	}
	if err = remapTreeMetadata(ctx, tx, meta); err != nil {
		return err
	}
	if err = remapEnrollments(ctx, tx, meta); err != nil {
		return err
	}
	if err = remapExercises(ctx, tx, meta); err != nil {
		return err
	}
	if err = remapSubmissions(ctx, tx, meta); err != nil {
		return err
	}
	if err = remapReviews(ctx, tx, meta); err != nil {
		return err
	}
	if err = remapReviewAssessments(ctx, tx, meta); err != nil {
		return err
	}
	if err = remapReviewArtifacts(ctx, tx, meta); err != nil {
		return err
	}
	if err = remapReviewJobs(ctx, tx, meta); err != nil {
		return err
	}
	if err = remapSubmissionSkillScores(ctx, tx); err != nil {
		return err
	}
	if err = remapLegacyCurriculumState(ctx, tx); err != nil {
		return err
	}
	if err = cleanupRetiredTrees(ctx, tx); err != nil {
		return err
	}

	return tx.Commit()
}

func loadMigrationContext(ctx context.Context, tx *sql.Tx) (migrationContext, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, slug, title, description, seed_codes_json, priority_skills_json
		FROM tgo_trees
	`)
	if err != nil {
		return migrationContext{}, err
	}
	defer rows.Close()

	type treeMeta struct {
		id           int64
		slug         string
		title        string
		description  string
		seedJSON     string
		priorityJSON string
	}
	var trees []treeMeta
	for rows.Next() {
		var meta treeMeta
		if err := rows.Scan(&meta.id, &meta.slug, &meta.title, &meta.description, &meta.seedJSON, &meta.priorityJSON); err != nil {
			return migrationContext{}, err
		}
		trees = append(trees, meta)
	}
	if err := rows.Err(); err != nil {
		return migrationContext{}, err
	}

	targets := make(map[string]migrationTreeTarget, len(trees))
	for _, tree := range trees {
		def, err := loadTreeDefinitionForMigration(ctx, tx, tree.id, tree.slug, tree.title, tree.description, tree.seedJSON, tree.priorityJSON)
		if err != nil {
			return migrationContext{}, err
		}
		targets[tree.slug] = buildTreeTarget(tree.id, def)
	}
	return migrationContext{TargetsBySlug: targets}, nil
}

func loadTreeDefinitionForMigration(ctx context.Context, tx *sql.Tx, treeID int64, slug, title, description, seedJSON, priorityJSON string) (domain.TGOTreeDefinition, error) {
	def := domain.TGOTreeDefinition{
		Slug:        slug,
		Title:       title,
		Description: description,
	}
	var err error
	if def.SeedCodes, err = DecodeStringSlice(seedJSON); err != nil {
		return domain.TGOTreeDefinition{}, fmt.Errorf("decode seed codes for %s: %w", slug, err)
	}
	if def.PrioritySkills, err = DecodeStringSlice(priorityJSON); err != nil {
		return domain.TGOTreeDefinition{}, fmt.Errorf("decode priority skills for %s: %w", slug, err)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT c.id, c.code, c.title, c.description, c.stage, c.stage_order, c.progress_mode, tt.prerequisites_json, tt.mastery_hint
		FROM tree_tgos tt
		JOIN tgo_catalog c ON c.code = tt.tgo_code
		WHERE tt.tree_id = ?
		ORDER BY c.stage_order ASC, c.id ASC
	`, treeID)
	if err != nil {
		return domain.TGOTreeDefinition{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var tgo domain.TGO
		var prereqsJSON string
		var masteryHint sql.NullString
		if err := rows.Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder, &tgo.ProgressMode, &prereqsJSON, &masteryHint); err != nil {
			return domain.TGOTreeDefinition{}, err
		}
		if tgo.Prerequisites, err = DecodeStringSlice(prereqsJSON); err != nil {
			return domain.TGOTreeDefinition{}, fmt.Errorf("decode prerequisites for %s/%s: %w", slug, tgo.Code, err)
		}
		tgo.MasteryHint = masteryHint.String
		def.TGOs = append(def.TGOs, tgo)
	}
	if err := rows.Err(); err != nil {
		return domain.TGOTreeDefinition{}, err
	}
	return def, nil
}

func buildTreeTarget(treeID int64, tree domain.TGOTreeDefinition) migrationTreeTarget {
	target := migrationTreeTarget{
		Slug:         tree.Slug,
		TreeID:       treeID,
		TGOs:         append([]domain.TGO(nil), tree.TGOs...),
		ValidCodes:   make(map[string]bool, len(tree.TGOs)),
		ByTitle:      make(map[string][]domain.TGO),
		BySkill:      make(map[string][]domain.TGO),
		SuffixToCode: map[string]string{},
		SeedCodes:    append([]string(nil), tree.SeedCodes...),
	}
	for _, tgo := range tree.TGOs {
		target.ValidCodes[tgo.Code] = true
		title := normalizeToken(tgo.Title)
		target.ByTitle[title] = append(target.ByTitle[title], tgo)
		skill := normalizeToken(domain.TGOCodeToSkill[tgo.Code])
		if skill != "" {
			target.BySkill[skill] = append(target.BySkill[skill], tgo)
		}
		for _, suffix := range codeSuffixes(tgo.Code) {
			if existing, ok := target.SuffixToCode[suffix]; ok && existing != tgo.Code {
				target.SuffixToCode[suffix] = ""
				continue
			}
			target.SuffixToCode[suffix] = tgo.Code
		}
	}
	return target
}

func legacySkillPathUpgradeNeeded(ctx context.Context, tx *sql.Tx, meta migrationContext) (bool, error) {
	for oldSlug := range retiredTreeSlugAliases {
		var count int
		if err := tx.QueryRowContext(ctx, `SELECT COUNT(1) FROM tgo_trees WHERE slug = ?`, oldSlug).Scan(&count); err != nil {
			return false, err
		}
		if count > 0 {
			return true, nil
		}
		for _, query := range []string{
			`SELECT COUNT(1) FROM users WHERE active_tree_slug = ?`,
			`SELECT COUNT(1) FROM user_onboarding_profiles WHERE generated_tree_slug = ?`,
			`SELECT COUNT(1) FROM enrollment_onboarding_profiles WHERE generated_tree_slug = ?`,
		} {
			if err := tx.QueryRowContext(ctx, query, oldSlug).Scan(&count); err != nil {
				return false, err
			}
			if count > 0 {
				return true, nil
			}
		}
	}

	enrollments, err := listEnrollmentMigrationRows(ctx, tx)
	if err != nil {
		return false, err
	}
	for _, row := range enrollments {
		target, ok := targetForSlug(meta, row.TreeSlug)
		if !ok {
			continue
		}
		if target.TreeID != row.TreeID {
			return true, nil
		}
		activeRows, err := loadActiveRows(ctx, tx, row.EnrollmentID)
		if err != nil {
			return false, err
		}
		for _, active := range activeRows {
			next := resolveTGOCodeForTarget(active.Code, target)
			if next != active.Code {
				return true, nil
			}
		}
		completedRows, err := loadCompletedRows(ctx, tx, row.EnrollmentID)
		if err != nil {
			return false, err
		}
		for _, completed := range completedRows {
			next := resolveTGOCodeForTarget(completed.Code, target)
			if next != completed.Code {
				return true, nil
			}
		}
		state, err := loadCurrentFocusState(ctx, tx, row.EnrollmentID)
		if err != nil {
			return false, err
		}
		if state.HasCurriculum && resolveCurrentFocus(state.Focus, target) != state.Focus {
			return true, nil
		}
	}

	if stale, err := anyExerciseNeedsUpgrade(ctx, tx, meta); err != nil || stale {
		return stale, err
	}
	if stale, err := anyTreeMetadataNeedsUpgrade(ctx, tx, meta); err != nil || stale {
		return stale, err
	}
	if stale, err := anyReviewNeedsUpgrade(ctx, tx, meta); err != nil || stale {
		return stale, err
	}
	if stale, err := anyReviewAssessmentNeedsUpgrade(ctx, tx, meta); err != nil || stale {
		return stale, err
	}
	if stale, err := anyReviewArtifactNeedsUpgrade(ctx, tx, meta); err != nil || stale {
		return stale, err
	}
	if stale, err := anySubmissionSkillScoreNeedsUpgrade(ctx, tx); err != nil || stale {
		return stale, err
	}
	if stale, err := anyLegacyCurriculumStateNeedsUpgrade(ctx, tx); err != nil || stale {
		return stale, err
	}
	return false, nil
}

func remapGlobalSlugReferences(ctx context.Context, tx *sql.Tx) error {
	for oldSlug, newSlug := range retiredTreeSlugAliases {
		for _, statement := range []struct {
			query string
			args  []any
		}{
			{query: `UPDATE users SET active_tree_slug = ? WHERE active_tree_slug = ?`, args: []any{newSlug, oldSlug}},
			{query: `UPDATE user_onboarding_profiles SET generated_tree_slug = ?, updated_at = CURRENT_TIMESTAMP WHERE generated_tree_slug = ?`, args: []any{newSlug, oldSlug}},
			{query: `UPDATE enrollment_onboarding_profiles SET generated_tree_slug = ?, updated_at = CURRENT_TIMESTAMP WHERE generated_tree_slug = ?`, args: []any{newSlug, oldSlug}},
		} {
			if _, err := tx.ExecContext(ctx, statement.query, statement.args...); err != nil {
				return err
			}
		}
	}
	return nil
}

func remapEnrollments(ctx context.Context, tx *sql.Tx, meta migrationContext) error {
	enrollments, err := listEnrollmentMigrationRows(ctx, tx)
	if err != nil {
		return err
	}
	for _, row := range enrollments {
		target, ok := targetForSlug(meta, row.TreeSlug)
		if !ok {
			continue
		}
		if err := rewriteEnrollmentActiveTGOs(ctx, tx, row.EnrollmentID, target); err != nil {
			return err
		}
		if err := rewriteEnrollmentCompletedTGOs(ctx, tx, row.EnrollmentID, target); err != nil {
			return err
		}
		if err := rewriteEnrollmentCurrentFocus(ctx, tx, row.EnrollmentID, target); err != nil {
			return err
		}
		if target.TreeID == row.TreeID {
			continue
		}
		targetEnrollmentID, err := otherEnrollmentForTree(ctx, tx, row.UserID, target.TreeID, row.EnrollmentID)
		if err != nil {
			return err
		}
		if targetEnrollmentID != 0 {
			if err := mergeEnrollmentIntoTarget(ctx, tx, row.EnrollmentID, targetEnrollmentID, target.Slug); err != nil {
				return err
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE user_tree_enrollments SET tree_id = ? WHERE id = ?
		`, target.TreeID, row.EnrollmentID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			UPDATE enrollment_onboarding_profiles
			SET generated_tree_slug = ?, updated_at = CURRENT_TIMESTAMP
			WHERE enrollment_id = ?
		`, target.Slug, row.EnrollmentID); err != nil {
			return err
		}
	}
	return nil
}

func remapExercises(ctx context.Context, tx *sql.Tx, meta migrationContext) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT e.id, e.tree_id, t.slug, e.tgo_codes_json, e.focus_skills_json
		FROM exercises e
		JOIN tgo_trees t ON t.id = e.tree_id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type update struct {
		id        int64
		treeID    int64
		tgoJSON   string
		focusJSON string
	}
	var updates []update
	for rows.Next() {
		var id, treeID int64
		var slug, tgoJSON, focusJSON string
		if err := rows.Scan(&id, &treeID, &slug, &tgoJSON, &focusJSON); err != nil {
			return err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok {
			continue
		}
		codes, err := DecodeStringSlice(tgoJSON)
		if err != nil {
			return err
		}
		focusSkills, err := DecodeStringSlice(focusJSON)
		if err != nil {
			return err
		}
		remappedCodes := remapCodeSliceForTarget(codes, target)
		remappedSkills := remapExerciseFocusSkills(focusSkills, remappedCodes)
		if target.TreeID == treeID && slices.Equal(remappedCodes, codes) && slices.Equal(remappedSkills, focusSkills) {
			continue
		}
		updates = append(updates, update{
			id:        id,
			treeID:    target.TreeID,
			tgoJSON:   mustJSON(remappedCodes),
			focusJSON: mustJSON(remappedSkills),
		})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range updates {
		if _, err := tx.ExecContext(ctx, `
			UPDATE exercises
			SET tree_id = ?, tgo_codes_json = ?, focus_skills_json = ?
			WHERE id = ?
		`, item.treeID, item.tgoJSON, item.focusJSON, item.id); err != nil {
			return err
		}
	}
	return nil
}

func remapTreeMetadata(ctx context.Context, tx *sql.Tx, meta migrationContext) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, slug, seed_codes_json, priority_skills_json
		FROM tgo_trees
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type treeUpdate struct {
		id           int64
		seedJSON     string
		priorityJSON string
	}
	var treeUpdates []treeUpdate
	for rows.Next() {
		var id int64
		var slug, seedJSON, priorityJSON string
		if err := rows.Scan(&id, &slug, &seedJSON, &priorityJSON); err != nil {
			return err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok {
			continue
		}
		seeds, err := DecodeStringSlice(seedJSON)
		if err != nil {
			return err
		}
		priorities, err := DecodeStringSlice(priorityJSON)
		if err != nil {
			return err
		}
		nextSeeds := remapCodeSliceForTarget(seeds, target)
		nextPriorities := remapSkillSlice(priorities)
		if slices.Equal(nextSeeds, seeds) && slices.Equal(nextPriorities, priorities) {
			continue
		}
		treeUpdates = append(treeUpdates, treeUpdate{
			id:           id,
			seedJSON:     mustJSON(nextSeeds),
			priorityJSON: mustJSON(nextPriorities),
		})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range treeUpdates {
		if _, err := tx.ExecContext(ctx, `
			UPDATE tgo_trees
			SET seed_codes_json = ?, priority_skills_json = ?
			WHERE id = ?
		`, item.seedJSON, item.priorityJSON, item.id); err != nil {
			return err
		}
	}

	versionRows, err := tx.QueryContext(ctx, `
		SELECT id, tree_id, seed_codes_json, priority_skills_json, tgos_json
		FROM tree_versions
	`)
	if err != nil {
		return err
	}
	defer versionRows.Close()

	type versionUpdate struct {
		id           int64
		seedJSON     string
		priorityJSON string
		tgosJSON     string
	}
	var versionUpdates []versionUpdate
	for versionRows.Next() {
		var id, treeID int64
		var seedJSON, priorityJSON, tgosJSON string
		if err := versionRows.Scan(&id, &treeID, &seedJSON, &priorityJSON, &tgosJSON); err != nil {
			return err
		}
		target, ok := targetForTreeID(meta, treeID)
		if !ok {
			continue
		}
		seeds, err := DecodeStringSlice(seedJSON)
		if err != nil {
			return err
		}
		priorities, err := DecodeStringSlice(priorityJSON)
		if err != nil {
			return err
		}
		nextSeeds := remapCodeSliceForTarget(seeds, target)
		nextPriorities := remapSkillSlice(priorities)
		nextTGOs, changedTGOs, err := remapVersionTGOsJSONForTarget(tgosJSON, target)
		if err != nil {
			return err
		}
		if slices.Equal(nextSeeds, seeds) && slices.Equal(nextPriorities, priorities) && !changedTGOs {
			continue
		}
		versionUpdates = append(versionUpdates, versionUpdate{
			id:           id,
			seedJSON:     mustJSON(nextSeeds),
			priorityJSON: mustJSON(nextPriorities),
			tgosJSON:     nextTGOs,
		})
	}
	if err := versionRows.Err(); err != nil {
		return err
	}
	for _, item := range versionUpdates {
		if _, err := tx.ExecContext(ctx, `
			UPDATE tree_versions
			SET seed_codes_json = ?, priority_skills_json = ?, tgos_json = ?
			WHERE id = ?
		`, item.seedJSON, item.priorityJSON, item.tgosJSON, item.id); err != nil {
			return err
		}
	}
	return nil
}

func remapSubmissions(ctx context.Context, tx *sql.Tx, meta migrationContext) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT s.id, s.tree_id, t.slug
		FROM submissions s
		JOIN tgo_trees t ON t.id = s.tree_id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type update struct {
		id     int64
		treeID int64
	}
	var updates []update
	for rows.Next() {
		var id, treeID int64
		var slug string
		if err := rows.Scan(&id, &treeID, &slug); err != nil {
			return err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok || target.TreeID == treeID {
			continue
		}
		updates = append(updates, update{id: id, treeID: target.TreeID})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range updates {
		if _, err := tx.ExecContext(ctx, `
			UPDATE submissions SET tree_id = ? WHERE id = ?
		`, item.treeID, item.id); err != nil {
			return err
		}
	}
	return nil
}

func remapReviews(ctx context.Context, tx *sql.Tx, meta migrationContext) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT r.id, r.tree_id, t.slug, r.completed_tgo_checks_json, r.next_focus
		FROM reviews r
		JOIN tgo_trees t ON t.id = r.tree_id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type update struct {
		id         int64
		treeID     int64
		checksJSON string
		nextFocus  string
	}
	var updates []update
	for rows.Next() {
		var id, treeID int64
		var slug, checksJSON, nextFocus string
		if err := rows.Scan(&id, &treeID, &slug, &checksJSON, &nextFocus); err != nil {
			return err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok {
			continue
		}
		nextChecksJSON, changedChecks, err := remapReviewChecksJSONForTarget(checksJSON, target)
		if err != nil {
			return err
		}
		remappedNextFocus := remapSkillName(nextFocus)
		if target.TreeID == treeID && !changedChecks && remappedNextFocus == nextFocus {
			continue
		}
		updates = append(updates, update{
			id:         id,
			treeID:     target.TreeID,
			checksJSON: nextChecksJSON,
			nextFocus:  remappedNextFocus,
		})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range updates {
		if _, err := tx.ExecContext(ctx, `
			UPDATE reviews
			SET tree_id = ?, completed_tgo_checks_json = ?, next_focus = ?
			WHERE id = ?
		`, item.treeID, item.checksJSON, item.nextFocus, item.id); err != nil {
			return err
		}
	}
	return nil
}

func remapReviewAssessments(ctx context.Context, tx *sql.Tx, meta migrationContext) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT a.id, a.tgo_code, t.slug
		FROM review_tgo_assessments a
		JOIN reviews r ON r.id = a.review_id
		JOIN tgo_trees t ON t.id = r.tree_id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type update struct {
		id   int64
		code string
	}
	var updates []update
	for rows.Next() {
		var id int64
		var code, slug string
		if err := rows.Scan(&id, &code, &slug); err != nil {
			return err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok {
			continue
		}
		next := resolveTGOCodeForTarget(code, target)
		if next == code {
			continue
		}
		updates = append(updates, update{id: id, code: next})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range updates {
		if _, err := tx.ExecContext(ctx, `
			UPDATE review_tgo_assessments SET tgo_code = ? WHERE id = ?
		`, item.code, item.id); err != nil {
			return err
		}
	}
	return nil
}

func remapReviewArtifacts(ctx context.Context, tx *sql.Tx, meta migrationContext) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT a.review_id, a.annotations_json, t.slug
		FROM review_artifacts a
		JOIN reviews r ON r.id = a.review_id
		JOIN tgo_trees t ON t.id = r.tree_id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type update struct {
		reviewID int64
		raw      string
	}
	var updates []update
	for rows.Next() {
		var reviewID int64
		var raw, slug string
		if err := rows.Scan(&reviewID, &raw, &slug); err != nil {
			return err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok {
			continue
		}
		next, changed, err := remapReviewAnnotationsJSONForTarget(raw, target)
		if err != nil {
			return err
		}
		if !changed {
			continue
		}
		updates = append(updates, update{reviewID: reviewID, raw: next})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range updates {
		if _, err := tx.ExecContext(ctx, `
			UPDATE review_artifacts SET annotations_json = ? WHERE review_id = ?
		`, item.raw, item.reviewID); err != nil {
			return err
		}
	}
	return nil
}

func remapReviewJobs(ctx context.Context, tx *sql.Tx, meta migrationContext) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT j.id, j.tree_id, t.slug
		FROM review_jobs j
		JOIN tgo_trees t ON t.id = j.tree_id
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type update struct {
		id     int64
		treeID int64
	}
	var updates []update
	for rows.Next() {
		var id, treeID int64
		var slug string
		if err := rows.Scan(&id, &treeID, &slug); err != nil {
			return err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok || target.TreeID == treeID {
			continue
		}
		updates = append(updates, update{id: id, treeID: target.TreeID})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range updates {
		if _, err := tx.ExecContext(ctx, `
			UPDATE review_jobs SET tree_id = ? WHERE id = ?
		`, item.treeID, item.id); err != nil {
			return err
		}
	}
	return nil
}

func remapSubmissionSkillScores(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `
		SELECT id, skill_name
		FROM submission_skill_scores
	`)
	if err != nil {
		return err
	}
	defer rows.Close()
	type update struct {
		id    int64
		skill string
	}
	var updates []update
	for rows.Next() {
		var id int64
		var skill string
		if err := rows.Scan(&id, &skill); err != nil {
			return err
		}
		next := remapSkillName(skill)
		if next == skill {
			continue
		}
		updates = append(updates, update{id: id, skill: next})
	}
	if err := rows.Err(); err != nil {
		return err
	}
	for _, item := range updates {
		if _, err := tx.ExecContext(ctx, `
			UPDATE submission_skill_scores SET skill_name = ? WHERE id = ?
		`, item.skill, item.id); err != nil {
			return err
		}
	}
	return nil
}

func remapLegacyCurriculumState(ctx context.Context, tx *sql.Tx) error {
	var currentFocus string
	err := tx.QueryRowContext(ctx, `
		SELECT current_focus FROM curriculum_state WHERE id = 1
	`).Scan(&currentFocus)
	if errors.Is(err, sql.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	next := remapSkillName(currentFocus)
	if next == currentFocus {
		return nil
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE curriculum_state SET current_focus = ?, updated_at = CURRENT_TIMESTAMP WHERE id = 1
	`, next)
	return err
}

func cleanupRetiredTrees(ctx context.Context, tx *sql.Tx) error {
	for oldSlug := range retiredTreeSlugAliases {
		var treeID int64
		err := tx.QueryRowContext(ctx, `SELECT id FROM tgo_trees WHERE slug = ?`, oldSlug).Scan(&treeID)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM tree_versions WHERE tree_id = ?`, treeID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM tree_tgos WHERE tree_id = ?`, treeID); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `DELETE FROM tgo_trees WHERE id = ?`, treeID); err != nil {
			return err
		}
	}
	return nil
}

func listEnrollmentMigrationRows(ctx context.Context, tx *sql.Tx) ([]enrollmentMigrationRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT e.id, e.user_id, e.tree_id, t.slug
		FROM user_tree_enrollments e
		JOIN tgo_trees t ON t.id = e.tree_id
		ORDER BY e.id ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []enrollmentMigrationRow
	for rows.Next() {
		var row enrollmentMigrationRow
		if err := rows.Scan(&row.EnrollmentID, &row.UserID, &row.TreeID, &row.TreeSlug); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func targetForSlug(meta migrationContext, slug string) (migrationTreeTarget, bool) {
	if targetSlug := retiredTreeSlugAliases[slug]; targetSlug != "" {
		target, ok := meta.TargetsBySlug[targetSlug]
		return target, ok
	}
	target, ok := meta.TargetsBySlug[slug]
	return target, ok
}

func targetForTreeID(meta migrationContext, treeID int64) (migrationTreeTarget, bool) {
	for _, target := range meta.TargetsBySlug {
		if target.TreeID == treeID {
			return target, true
		}
	}
	return migrationTreeTarget{}, false
}

func loadActiveRows(ctx context.Context, tx *sql.Tx, enrollmentID int64) ([]activeCodeRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT slot, tgo_code, activated_at
		FROM enrollment_active_tgos
		WHERE enrollment_id = ?
		ORDER BY slot ASC
	`, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []activeCodeRow
	for rows.Next() {
		var row activeCodeRow
		if err := rows.Scan(&row.Slot, &row.Code, &row.ActivatedAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func loadCompletedRows(ctx context.Context, tx *sql.Tx, enrollmentID int64) ([]completedCodeRow, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT tgo_code, completed_at
		FROM enrollment_completed_tgos
		WHERE enrollment_id = ?
		ORDER BY completed_at ASC
	`, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []completedCodeRow
	for rows.Next() {
		var row completedCodeRow
		if err := rows.Scan(&row.Code, &row.CompletedAt); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func loadCurrentFocusState(ctx context.Context, tx *sql.Tx, enrollmentID int64) (currentFocusState, error) {
	var state currentFocusState
	err := tx.QueryRowContext(ctx, `
		SELECT current_focus, difficulty_level, COALESCE(last_review_id, 0)
		FROM user_curriculum_state
		WHERE enrollment_id = ?
	`, enrollmentID).Scan(&state.Focus, &state.Difficulty, &state.LastReviewID)
	if errors.Is(err, sql.ErrNoRows) {
		return currentFocusState{HasCurriculum: false}, nil
	}
	if err != nil {
		return currentFocusState{}, err
	}
	state.HasCurriculum = true
	state.EnrollmentID = enrollmentID
	return state, nil
}

func rewriteEnrollmentActiveTGOs(ctx context.Context, tx *sql.Tx, enrollmentID int64, target migrationTreeTarget) error {
	rows, err := loadActiveRows(ctx, tx, enrollmentID)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	used := map[string]bool{}
	capHint := len(rows)
	if capHint < 3 {
		capHint = 3
	}
	resolved := make([]activeCodeRow, 0, capHint)
	for _, row := range rows {
		code := resolveTGOCodeForTarget(row.Code, target)
		if code == "" || used[code] {
			code = fallbackTGOCode(target, used)
		}
		if code == "" {
			continue
		}
		used[code] = true
		row.Code = code
		resolved = append(resolved, row)
	}
	for len(resolved) < 3 {
		code := fallbackTGOCode(target, used)
		if code == "" {
			break
		}
		used[code] = true
		resolved = append(resolved, activeCodeRow{
			Slot:        len(resolved) + 1,
			Code:        code,
			ActivatedAt: time.Now().UTC(),
		})
	}
	for i := range resolved {
		resolved[i].Slot = i + 1
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM enrollment_active_tgos WHERE enrollment_id = ?`, enrollmentID); err != nil {
		return err
	}
	for _, row := range resolved {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO enrollment_active_tgos (enrollment_id, slot, tgo_code, activated_at)
			VALUES (?, ?, ?, ?)
		`, enrollmentID, row.Slot, row.Code, row.ActivatedAt); err != nil {
			return err
		}
	}
	return nil
}

func rewriteEnrollmentCompletedTGOs(ctx context.Context, tx *sql.Tx, enrollmentID int64, target migrationTreeTarget) error {
	rows, err := loadCompletedRows(ctx, tx, enrollmentID)
	if err != nil {
		return err
	}
	if len(rows) == 0 {
		return nil
	}
	earliest := map[string]time.Time{}
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		code := resolveTGOCodeForTarget(row.Code, target)
		if code == "" {
			continue
		}
		if ts, ok := earliest[code]; !ok {
			earliest[code] = row.CompletedAt
			order = append(order, code)
		} else if row.CompletedAt.Before(ts) {
			earliest[code] = row.CompletedAt
		}
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM enrollment_completed_tgos WHERE enrollment_id = ?`, enrollmentID); err != nil {
		return err
	}
	for _, code := range order {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO enrollment_completed_tgos (enrollment_id, tgo_code, completed_at)
			VALUES (?, ?, ?)
		`, enrollmentID, code, earliest[code]); err != nil {
			return err
		}
	}
	return nil
}

func rewriteEnrollmentCurrentFocus(ctx context.Context, tx *sql.Tx, enrollmentID int64, target migrationTreeTarget) error {
	state, err := loadCurrentFocusState(ctx, tx, enrollmentID)
	if err != nil || !state.HasCurriculum {
		return err
	}
	next := resolveCurrentFocus(state.Focus, target)
	if next == state.Focus {
		return nil
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE user_curriculum_state
		SET current_focus = ?, updated_at = CURRENT_TIMESTAMP
		WHERE enrollment_id = ?
	`, next, enrollmentID)
	return err
}

func otherEnrollmentForTree(ctx context.Context, tx *sql.Tx, userID, treeID, excludeEnrollmentID int64) (int64, error) {
	var enrollmentID int64
	err := tx.QueryRowContext(ctx, `
		SELECT id
		FROM user_tree_enrollments
		WHERE user_id = ? AND tree_id = ? AND id <> ?
		ORDER BY CASE WHEN archived_at IS NULL THEN 0 ELSE 1 END, id ASC
		LIMIT 1
	`, userID, treeID, excludeEnrollmentID).Scan(&enrollmentID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	return enrollmentID, err
}

func mergeEnrollmentIntoTarget(ctx context.Context, tx *sql.Tx, sourceEnrollmentID, targetEnrollmentID int64, targetSlug string) error {
	sourceActive, err := loadActiveRows(ctx, tx, sourceEnrollmentID)
	if err != nil {
		return err
	}
	targetActive, err := loadActiveRows(ctx, tx, targetEnrollmentID)
	if err != nil {
		return err
	}
	if len(targetActive) == 0 && len(sourceActive) > 0 {
		for _, row := range sourceActive {
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO enrollment_active_tgos (enrollment_id, slot, tgo_code, activated_at)
				VALUES (?, ?, ?, ?)
			`, targetEnrollmentID, row.Slot, row.Code, row.ActivatedAt); err != nil {
				return err
			}
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO enrollment_completed_tgos (enrollment_id, tgo_code, completed_at)
		SELECT ?, tgo_code, completed_at
		FROM enrollment_completed_tgos
		WHERE enrollment_id = ?
		ON CONFLICT(enrollment_id, tgo_code) DO NOTHING
	`, targetEnrollmentID, sourceEnrollmentID); err != nil {
		return err
	}

	if err := mergeCurriculumState(ctx, tx, sourceEnrollmentID, targetEnrollmentID); err != nil {
		return err
	}
	if err := mergeEnrollmentOnboardingProfile(ctx, tx, sourceEnrollmentID, targetEnrollmentID, targetSlug); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE review_jobs SET enrollment_id = ? WHERE enrollment_id = ?
	`, targetEnrollmentID, sourceEnrollmentID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM enrollment_active_tgos WHERE enrollment_id = ?`, sourceEnrollmentID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM enrollment_completed_tgos WHERE enrollment_id = ?`, sourceEnrollmentID); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM user_tree_enrollments WHERE id = ?`, sourceEnrollmentID); err != nil {
		return err
	}
	return nil
}

func mergeCurriculumState(ctx context.Context, tx *sql.Tx, sourceEnrollmentID, targetEnrollmentID int64) error {
	source, err := loadCurrentFocusState(ctx, tx, sourceEnrollmentID)
	if err != nil || !source.HasCurriculum {
		return err
	}
	target, err := loadCurrentFocusState(ctx, tx, targetEnrollmentID)
	if err != nil {
		return err
	}
	if !target.HasCurriculum {
		_, err = tx.ExecContext(ctx, `
			UPDATE user_curriculum_state
			SET enrollment_id = ?, updated_at = CURRENT_TIMESTAMP
			WHERE enrollment_id = ?
		`, targetEnrollmentID, sourceEnrollmentID)
		return err
	}

	focus := target.Focus
	if focus == "" {
		focus = source.Focus
	}
	difficulty := target.Difficulty
	if source.Difficulty > difficulty {
		difficulty = source.Difficulty
	}
	lastReviewID := target.LastReviewID
	if lastReviewID == 0 {
		lastReviewID = source.LastReviewID
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE user_curriculum_state
		SET current_focus = ?, difficulty_level = ?, last_review_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE enrollment_id = ?
	`, focus, difficulty, nullableID(lastReviewID), targetEnrollmentID); err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `DELETE FROM user_curriculum_state WHERE enrollment_id = ?`, sourceEnrollmentID)
	return err
}

func mergeEnrollmentOnboardingProfile(ctx context.Context, tx *sql.Tx, sourceEnrollmentID, targetEnrollmentID int64, targetSlug string) error {
	var sourceExists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(1) FROM enrollment_onboarding_profiles WHERE enrollment_id = ?`, sourceEnrollmentID).Scan(&sourceExists); err != nil {
		return err
	}
	if sourceExists == 0 {
		return nil
	}
	var targetExists int
	if err := tx.QueryRowContext(ctx, `SELECT COUNT(1) FROM enrollment_onboarding_profiles WHERE enrollment_id = ?`, targetEnrollmentID).Scan(&targetExists); err != nil {
		return err
	}
	if targetExists == 0 {
		_, err := tx.ExecContext(ctx, `
			UPDATE enrollment_onboarding_profiles
			SET enrollment_id = ?, generated_tree_slug = ?, updated_at = CURRENT_TIMESTAMP
			WHERE enrollment_id = ?
		`, targetEnrollmentID, targetSlug, sourceEnrollmentID)
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE enrollment_onboarding_profiles
		SET generated_tree_slug = ?, updated_at = CURRENT_TIMESTAMP
		WHERE enrollment_id = ?
	`, targetSlug, targetEnrollmentID); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		DELETE FROM enrollment_onboarding_profiles WHERE enrollment_id = ?
	`, sourceEnrollmentID)
	return err
}

func anyExerciseNeedsUpgrade(ctx context.Context, tx *sql.Tx, meta migrationContext) (bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT e.tree_id, t.slug, e.tgo_codes_json, e.focus_skills_json
		FROM exercises e
		JOIN tgo_trees t ON t.id = e.tree_id
	`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var treeID int64
		var slug, tgoJSON, focusJSON string
		if err := rows.Scan(&treeID, &slug, &tgoJSON, &focusJSON); err != nil {
			return false, err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok {
			continue
		}
		codes, err := DecodeStringSlice(tgoJSON)
		if err != nil {
			return false, err
		}
		focus, err := DecodeStringSlice(focusJSON)
		if err != nil {
			return false, err
		}
		if target.TreeID != treeID || !slices.Equal(remapCodeSliceForTarget(codes, target), codes) || !slices.Equal(remapExerciseFocusSkills(focus, codes), focus) {
			return true, nil
		}
	}
	return false, rows.Err()
}

func anyTreeMetadataNeedsUpgrade(ctx context.Context, tx *sql.Tx, meta migrationContext) (bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT slug, seed_codes_json, priority_skills_json
		FROM tgo_trees
	`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var slug, seedJSON, priorityJSON string
		if err := rows.Scan(&slug, &seedJSON, &priorityJSON); err != nil {
			return false, err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok {
			continue
		}
		seeds, err := DecodeStringSlice(seedJSON)
		if err != nil {
			return false, err
		}
		priorities, err := DecodeStringSlice(priorityJSON)
		if err != nil {
			return false, err
		}
		if !slices.Equal(remapCodeSliceForTarget(seeds, target), seeds) || !slices.Equal(remapSkillSlice(priorities), priorities) {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}

	versionRows, err := tx.QueryContext(ctx, `
		SELECT tree_id, seed_codes_json, priority_skills_json, tgos_json
		FROM tree_versions
	`)
	if err != nil {
		return false, err
	}
	defer versionRows.Close()
	for versionRows.Next() {
		var treeID int64
		var seedJSON, priorityJSON, tgosJSON string
		if err := versionRows.Scan(&treeID, &seedJSON, &priorityJSON, &tgosJSON); err != nil {
			return false, err
		}
		target, ok := targetForTreeID(meta, treeID)
		if !ok {
			continue
		}
		seeds, err := DecodeStringSlice(seedJSON)
		if err != nil {
			return false, err
		}
		priorities, err := DecodeStringSlice(priorityJSON)
		if err != nil {
			return false, err
		}
		_, changedTGOs, err := remapVersionTGOsJSONForTarget(tgosJSON, target)
		if err != nil {
			return false, err
		}
		if !slices.Equal(remapCodeSliceForTarget(seeds, target), seeds) || !slices.Equal(remapSkillSlice(priorities), priorities) || changedTGOs {
			return true, nil
		}
	}
	return false, versionRows.Err()
}

func anyReviewNeedsUpgrade(ctx context.Context, tx *sql.Tx, meta migrationContext) (bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT r.tree_id, t.slug, r.completed_tgo_checks_json, r.next_focus
		FROM reviews r
		JOIN tgo_trees t ON t.id = r.tree_id
	`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var treeID int64
		var slug, checksJSON, nextFocus string
		if err := rows.Scan(&treeID, &slug, &checksJSON, &nextFocus); err != nil {
			return false, err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok {
			continue
		}
		_, changed, err := remapReviewChecksJSONForTarget(checksJSON, target)
		if err != nil {
			return false, err
		}
		if target.TreeID != treeID || changed || remapSkillName(nextFocus) != nextFocus {
			return true, nil
		}
	}
	return false, rows.Err()
}

func anyReviewAssessmentNeedsUpgrade(ctx context.Context, tx *sql.Tx, meta migrationContext) (bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT a.tgo_code, t.slug
		FROM review_tgo_assessments a
		JOIN reviews r ON r.id = a.review_id
		JOIN tgo_trees t ON t.id = r.tree_id
	`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var code, slug string
		if err := rows.Scan(&code, &slug); err != nil {
			return false, err
		}
		target, ok := targetForSlug(meta, slug)
		if ok && resolveTGOCodeForTarget(code, target) != code {
			return true, nil
		}
	}
	return false, rows.Err()
}

func anyReviewArtifactNeedsUpgrade(ctx context.Context, tx *sql.Tx, meta migrationContext) (bool, error) {
	rows, err := tx.QueryContext(ctx, `
		SELECT a.annotations_json, t.slug
		FROM review_artifacts a
		JOIN reviews r ON r.id = a.review_id
		JOIN tgo_trees t ON t.id = r.tree_id
	`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var raw, slug string
		if err := rows.Scan(&raw, &slug); err != nil {
			return false, err
		}
		target, ok := targetForSlug(meta, slug)
		if !ok {
			continue
		}
		_, changed, err := remapReviewAnnotationsJSONForTarget(raw, target)
		if err != nil {
			return false, err
		}
		if changed {
			return true, nil
		}
	}
	return false, rows.Err()
}

func anySubmissionSkillScoreNeedsUpgrade(ctx context.Context, tx *sql.Tx) (bool, error) {
	rows, err := tx.QueryContext(ctx, `SELECT skill_name FROM submission_skill_scores`)
	if err != nil {
		return false, err
	}
	defer rows.Close()
	for rows.Next() {
		var skill string
		if err := rows.Scan(&skill); err != nil {
			return false, err
		}
		if remapSkillName(skill) != skill {
			return true, nil
		}
	}
	return false, rows.Err()
}

func anyLegacyCurriculumStateNeedsUpgrade(ctx context.Context, tx *sql.Tx) (bool, error) {
	var currentFocus string
	err := tx.QueryRowContext(ctx, `SELECT current_focus FROM curriculum_state WHERE id = 1`).Scan(&currentFocus)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return remapSkillName(currentFocus) != currentFocus, nil
}

func remapReviewChecksJSONForTarget(raw string, target migrationTreeTarget) (string, bool, error) {
	if raw == "" || raw == "[]" {
		return raw, false, nil
	}
	var checks []domain.TGOAssessment
	if err := json.Unmarshal([]byte(raw), &checks); err != nil {
		return "", false, err
	}
	changed := false
	for i := range checks {
		next := resolveTGOCodeForTarget(checks[i].TGOCode, target)
		if next != checks[i].TGOCode {
			checks[i].TGOCode = next
			changed = true
		}
	}
	if !changed {
		return raw, false, nil
	}
	return mustJSON(checks), true, nil
}

func remapReviewAnnotationsJSONForTarget(raw string, target migrationTreeTarget) (string, bool, error) {
	if raw == "" || raw == "[]" {
		return raw, false, nil
	}
	var annotations []domain.ReviewAnnotation
	if err := json.Unmarshal([]byte(raw), &annotations); err != nil {
		return "", false, err
	}
	changed := false
	for i := range annotations {
		next := resolveTGOCodeForTarget(annotations[i].TGOCode, target)
		if next != annotations[i].TGOCode {
			annotations[i].TGOCode = next
			changed = true
		}
	}
	if !changed {
		return raw, false, nil
	}
	return mustJSON(annotations), true, nil
}

func remapVersionTGOsJSONForTarget(raw string, target migrationTreeTarget) (string, bool, error) {
	if raw == "" || raw == "[]" {
		return raw, false, nil
	}
	var tgos []domain.TGO
	if err := json.Unmarshal([]byte(raw), &tgos); err != nil {
		return "", false, err
	}
	changed := false
	for i := range tgos {
		nextCode := resolveTGOCodeForTarget(tgos[i].Code, target)
		if nextCode != "" && nextCode != tgos[i].Code {
			tgos[i].Code = nextCode
			changed = true
		}
		nextPrereqs := remapCodeSliceForTarget(tgos[i].Prerequisites, target)
		if !slices.Equal(nextPrereqs, tgos[i].Prerequisites) {
			tgos[i].Prerequisites = nextPrereqs
			changed = true
		}
	}
	if !changed {
		return raw, false, nil
	}
	return mustJSON(tgos), true, nil
}

func remapCodeSliceForTarget(codes []string, target migrationTreeTarget) []string {
	out := make([]string, 0, len(codes))
	seen := map[string]bool{}
	for _, code := range codes {
		next := resolveTGOCodeForTarget(code, target)
		if next == "" || seen[next] {
			continue
		}
		seen[next] = true
		out = append(out, next)
	}
	return out
}

func remapExerciseFocusSkills(skills []string, remappedCodes []string) []string {
	if len(remappedCodes) > 0 {
		seen := map[string]bool{}
		out := make([]string, 0, len(remappedCodes))
		for _, code := range remappedCodes {
			skill := domain.TGOCodeToSkill[code]
			if skill == "" || seen[skill] {
				continue
			}
			seen[skill] = true
			out = append(out, skill)
		}
		if len(out) > 0 {
			return out
		}
	}
	out := make([]string, 0, len(skills))
	seen := map[string]bool{}
	for _, skill := range skills {
		next := remapSkillName(skill)
		if next == "" || seen[next] {
			continue
		}
		seen[next] = true
		out = append(out, next)
	}
	return out
}

func remapSkillSlice(skills []string) []string {
	out := make([]string, 0, len(skills))
	seen := map[string]bool{}
	for _, skill := range skills {
		next := remapSkillName(skill)
		if next == "" || seen[next] {
			continue
		}
		seen[next] = true
		out = append(out, next)
	}
	return out
}

func resolveCurrentFocus(value string, target migrationTreeTarget) string {
	if value == "" {
		return value
	}
	if next := resolveTGOCodeForTarget(value, target); next != value {
		return next
	}
	return remapSkillName(value)
}

func resolveTGOCodeForTarget(code string, target migrationTreeTarget) string {
	if code == "" {
		return ""
	}
	if target.ValidCodes[code] {
		return code
	}
	if explicit := legacyTGOCodeAliases[code]; explicit != "" && target.ValidCodes[explicit] {
		return explicit
	}
	if suffix := target.SuffixToCode[code]; suffix != "" {
		return suffix
	}
	for _, suffix := range codeSuffixes(code) {
		if match := target.SuffixToCode[suffix]; match != "" {
			return match
		}
	}
	if tgo, ok := firstMatch(target.ByTitle[normalizeToken(code)]); ok {
		return tgo.Code
	}
	if tgo, ok := firstMatch(target.ByTitle[normalizeToken(strings.ReplaceAll(code, "-", " "))]); ok {
		return tgo.Code
	}
	if tgo, ok := firstMatch(target.BySkill[normalizeToken(legacySkillForCode(code))]); ok {
		return tgo.Code
	}
	return ""
}

func fallbackTGOCode(target migrationTreeTarget, used map[string]bool) string {
	for _, code := range target.SeedCodes {
		if !used[code] {
			return code
		}
	}
	for _, tgo := range target.TGOs {
		if !used[tgo.Code] {
			return tgo.Code
		}
	}
	return ""
}

func legacySkillForCode(code string) string {
	if skill := domain.TGOCodeToSkill[code]; skill != "" {
		return skill
	}
	if alias := legacyTGOCodeAliases[code]; alias != "" {
		return domain.TGOCodeToSkill[alias]
	}
	return ""
}

func remapSkillName(skill string) string {
	if skill == "" {
		return ""
	}
	if next := legacySkillAliases[normalizeToken(skill)]; next != "" {
		return next
	}
	return skill
}

func firstMatch(values []domain.TGO) (domain.TGO, bool) {
	if len(values) == 0 {
		return domain.TGO{}, false
	}
	return values[0], true
}

func codeSuffixes(code string) []string {
	parts := strings.Split(code, "-")
	out := make([]string, 0, len(parts))
	for i := 0; i < len(parts); i++ {
		out = append(out, strings.Join(parts[i:], "-"))
	}
	return out
}

func normalizeToken(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	value = strings.ReplaceAll(value, "_", " ")
	value = strings.ReplaceAll(value, "-", " ")
	value = strings.Join(strings.Fields(value), " ")
	return value
}
