package db

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/tomasino/writing-coach/internal/domain"
)

type Store struct {
	SQL *sql.DB
}

func Open(databasePath string) (*Store, error) {
	sqlDB, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(1)

	return &Store{SQL: sqlDB}, nil
}

func (s *Store) Close() error {
	if s == nil || s.SQL == nil {
		return nil
	}
	return s.SQL.Close()
}

func (s *Store) EnsureSeedData(ctx context.Context, writerName string) error {
	if _, err := s.SQL.ExecContext(ctx, `
		INSERT INTO writer_profile (name, aesthetic_target)
		SELECT ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM writer_profile)
	`, writerName, domain.WriterTrackName+": epic tragedy in a mythopoeic mode with fantasy influences, disciplined by Herbert/Le Guin depth rather than imitation"); err != nil {
		return err
	}

	if _, err := s.SQL.ExecContext(ctx, `
		INSERT INTO curriculum_state (id, current_focus, difficulty_level, updated_at)
		SELECT 1, 'tragic inevitability', 2, CURRENT_TIMESTAMP
		WHERE NOT EXISTS (SELECT 1 FROM curriculum_state WHERE id = 1)
	`); err != nil {
		return err
	}

	for _, skill := range domain.SupportedSkills {
		if _, err := s.SQL.ExecContext(ctx, `
			INSERT INTO skill_dimensions (name)
			SELECT ?
			WHERE NOT EXISTS (SELECT 1 FROM skill_dimensions WHERE name = ?)
		`, skill, skill); err != nil {
			return err
		}
	}

	for _, tgo := range domain.AllTGOs() {
		if _, err := s.SQL.ExecContext(ctx, `
			INSERT INTO tgo_catalog (code, title, description, stage, stage_order)
			SELECT ?, ?, ?, ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM tgo_catalog WHERE code = ?)
		`, tgo.Code, tgo.Title, tgo.Description, tgo.Stage, tgo.StageOrder, tgo.Code); err != nil {
			return err
		}
		if _, err := s.SQL.ExecContext(ctx, `
			UPDATE tgo_catalog
			SET title = ?, description = ?, stage = ?, stage_order = ?
			WHERE code = ?
		`, tgo.Title, tgo.Description, tgo.Stage, tgo.StageOrder, tgo.Code); err != nil {
			return err
		}
	}

	for _, tree := range domain.BuiltInTrees {
		if err := s.SaveTreeDefinition(ctx, tree); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) EnsureAdminEmails(ctx context.Context, emails []string) error {
	for _, email := range emails {
		email = strings.TrimSpace(strings.ToLower(email))
		if email == "" {
			continue
		}
		if _, err := s.SQL.ExecContext(ctx, `
			INSERT INTO admin_identities (email)
			SELECT ?
			WHERE NOT EXISTS (SELECT 1 FROM admin_identities WHERE email = ?)
		`, email, email); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) EnsureDefaultUserTree(ctx context.Context, userSlug, userName, treeSlug string) (int64, int64, int64, error) {
	treeDef, err := s.TreeDefinitionBySlug(ctx, treeSlug)
	if err != nil {
		if IsNotFound(err) {
			return 0, 0, 0, fmt.Errorf("unknown tree slug %q", treeSlug)
		}
		return 0, 0, 0, err
	}

	if err := s.EnsureUser(ctx, userSlug, userName); err != nil {
		return 0, 0, 0, err
	}

	user, err := s.UserBySlug(ctx, userSlug)
	if err != nil {
		return 0, 0, 0, err
	}
	tree, err := s.TreeBySlug(ctx, treeSlug)
	if err != nil {
		return 0, 0, 0, err
	}

	if _, err := s.SQL.ExecContext(ctx, `
		INSERT INTO user_tree_enrollments (user_id, tree_id)
		SELECT ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM user_tree_enrollments WHERE user_id = ? AND tree_id = ?)
	`, user.ID, tree.ID, user.ID, tree.ID); err != nil {
		return 0, 0, 0, err
	}

	enrollmentID, err := s.EnrollmentID(ctx, user.ID, tree.ID)
	if err != nil {
		return 0, 0, 0, err
	}
	if _, err := s.SQL.ExecContext(ctx, `
		INSERT INTO user_curriculum_state (enrollment_id, current_focus, difficulty_level)
		SELECT ?, ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM user_curriculum_state WHERE enrollment_id = ?)
	`, enrollmentID, treeDef.SeedCodes[0], 2, enrollmentID); err != nil {
		return 0, 0, 0, err
	}
	for idx, code := range domain.SeedCodesForDefinition(treeDef) {
		slot := idx + 1
		if _, err := s.SQL.ExecContext(ctx, `
			INSERT INTO enrollment_active_tgos (enrollment_id, slot, tgo_code)
			SELECT ?, ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM enrollment_active_tgos WHERE enrollment_id = ? AND slot = ?)
		`, enrollmentID, slot, code, enrollmentID, slot); err != nil {
			return 0, 0, 0, err
		}
	}
	if user.ActiveTreeSlug == "" {
		if err := s.SetUserActiveTree(ctx, user.ID, treeSlug); err != nil {
			return 0, 0, 0, err
		}
	}

	return user.ID, tree.ID, enrollmentID, nil
}

func (s *Store) EnsureUser(ctx context.Context, userSlug, userName string) error {
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO users (slug, name)
		SELECT ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM users WHERE slug = ?)
	`, userSlug, userName, userSlug)
	return err
}

func (s *Store) UserBySlug(ctx context.Context, slug string) (domain.User, error) {
	var user domain.User
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, slug, name, active_tree_slug, created_at FROM users WHERE slug = ?
	`, slug).Scan(&user.ID, &user.Slug, &user.Name, &user.ActiveTreeSlug, &user.CreatedAt)
	return user, err
}

func (s *Store) UserActiveTreeSlug(ctx context.Context, userSlug string) (string, error) {
	var activeTreeSlug string
	err := s.SQL.QueryRowContext(ctx, `
		SELECT active_tree_slug
		FROM users
		WHERE slug = ?
	`, userSlug).Scan(&activeTreeSlug)
	return activeTreeSlug, err
}

func (s *Store) SetUserActiveTree(ctx context.Context, userID int64, treeSlug string) error {
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE users
		SET active_tree_slug = ?
		WHERE id = ?
	`, treeSlug, userID)
	return err
}

func (s *Store) IsAdminEmail(ctx context.Context, email string) (bool, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return false, nil
	}
	var exists int
	err := s.SQL.QueryRowContext(ctx, `SELECT 1 FROM admin_identities WHERE email = ?`, email).Scan(&exists)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *Store) ListAdminEmails(ctx context.Context) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT email
		FROM admin_identities
		ORDER BY email ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var emails []string
	for rows.Next() {
		var email string
		if err := rows.Scan(&email); err != nil {
			return nil, err
		}
		emails = append(emails, email)
	}
	return emails, rows.Err()
}

func (s *Store) AddAdminEmail(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return fmt.Errorf("email is required")
	}
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO admin_identities (email)
		VALUES (?)
		ON CONFLICT(email) DO NOTHING
	`, email)
	return err
}

func (s *Store) RemoveAdminEmail(ctx context.Context, email string) error {
	email = strings.TrimSpace(strings.ToLower(email))
	if email == "" {
		return fmt.Errorf("email is required")
	}
	res, err := s.SQL.ExecContext(ctx, `DELETE FROM admin_identities WHERE email = ?`, email)
	if err != nil {
		return err
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *Store) TreeBySlug(ctx context.Context, slug string) (domain.TGOTree, error) {
	var tree domain.TGOTree
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, slug, title, description, created_at FROM tgo_trees WHERE slug = ?
	`, slug).Scan(&tree.ID, &tree.Slug, &tree.Title, &tree.Description, &tree.CreatedAt)
	return tree, err
}

func (s *Store) TreeDefinitionBySlug(ctx context.Context, slug string) (domain.TGOTreeDefinition, error) {
	var def domain.TGOTreeDefinition
	var seedCodesJSON, prioritySkillsJSON string
	err := s.SQL.QueryRowContext(ctx, `
		SELECT slug, title, description, seed_codes_json, priority_skills_json
		FROM tgo_trees
		WHERE slug = ?
	`, slug).Scan(&def.Slug, &def.Title, &def.Description, &seedCodesJSON, &prioritySkillsJSON)
	if err != nil {
		return domain.TGOTreeDefinition{}, err
	}
	if def.SeedCodes, err = DecodeStringSlice(seedCodesJSON); err != nil {
		return domain.TGOTreeDefinition{}, err
	}
	if def.PrioritySkills, err = DecodeStringSlice(prioritySkillsJSON); err != nil {
		return domain.TGOTreeDefinition{}, err
	}
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT c.id, c.code, c.title, c.description, c.stage, c.stage_order, tt.prerequisites_json, tt.mastery_hint
		FROM tree_tgos tt
		JOIN tgo_trees t ON t.id = tt.tree_id
		JOIN tgo_catalog c ON c.code = tt.tgo_code
		WHERE t.slug = ?
		ORDER BY c.stage_order ASC, c.id ASC
	`, slug)
	if err != nil {
		return domain.TGOTreeDefinition{}, err
	}
	defer rows.Close()
	for rows.Next() {
		var tgo domain.TGO
		var prereqsJSON string
		if err := rows.Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder, &prereqsJSON, &tgo.MasteryHint); err != nil {
			return domain.TGOTreeDefinition{}, err
		}
		if tgo.Prerequisites, err = DecodeStringSlice(prereqsJSON); err != nil {
			return domain.TGOTreeDefinition{}, err
		}
		def.TGOs = append(def.TGOs, tgo)
	}
	return def, rows.Err()
}

func (s *Store) SaveTreeDefinition(ctx context.Context, def domain.TGOTreeDefinition) error {
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO tgo_trees (slug, title, description, seed_codes_json, priority_skills_json)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(slug) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			seed_codes_json = excluded.seed_codes_json,
			priority_skills_json = excluded.priority_skills_json
	`, def.Slug, def.Title, def.Description, mustJSON(def.SeedCodes), mustJSON(def.PrioritySkills)); err != nil {
		return err
	}

	var treeID int64
	if err = tx.QueryRowContext(ctx, `SELECT id FROM tgo_trees WHERE slug = ?`, def.Slug).Scan(&treeID); err != nil {
		return err
	}

	for _, tgo := range def.TGOs {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO tgo_catalog (code, title, description, stage, stage_order)
			VALUES (?, ?, ?, ?, ?)
			ON CONFLICT(code) DO UPDATE SET
				title = excluded.title,
				description = excluded.description,
				stage = excluded.stage,
				stage_order = excluded.stage_order
		`, tgo.Code, tgo.Title, tgo.Description, tgo.Stage, tgo.StageOrder); err != nil {
			return err
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO tree_tgos (tree_id, tgo_code, prerequisites_json, mastery_hint)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(tree_id, tgo_code) DO UPDATE SET
				prerequisites_json = excluded.prerequisites_json,
				mastery_hint = excluded.mastery_hint
		`, treeID, tgo.Code, mustJSON(tgo.Prerequisites), tgo.MasteryHint); err != nil {
			return err
		}
	}

	if len(def.TGOs) > 0 {
		placeholders := strings.TrimSuffix(strings.Repeat("?,", len(def.TGOs)), ",")
		args := make([]any, 0, len(def.TGOs)+1)
		args = append(args, treeID)
		for _, tgo := range def.TGOs {
			args = append(args, tgo.Code)
		}
		if _, err = tx.ExecContext(ctx, `
			DELETE FROM tree_tgos
			WHERE tree_id = ? AND tgo_code NOT IN (`+placeholders+`)
		`, args...); err != nil {
			return err
		}
	}

	snapshotJSON := mustJSON(def.TGOs)
	var latestVersion int
	var latestTitle, latestDescription, latestSeedCodesJSON, latestPrioritySkillsJSON, latestTGOsJSON string
	switch err = tx.QueryRowContext(ctx, `
		SELECT version, title, description, seed_codes_json, priority_skills_json, tgos_json
		FROM tree_versions
		WHERE tree_id = ?
		ORDER BY version DESC
		LIMIT 1
	`, treeID).Scan(&latestVersion, &latestTitle, &latestDescription, &latestSeedCodesJSON, &latestPrioritySkillsJSON, &latestTGOsJSON); {
	case errors.Is(err, sql.ErrNoRows):
		latestVersion = 0
		err = nil
	case err != nil:
		return err
	}
	if latestVersion == 0 || latestTitle != def.Title || latestDescription != def.Description || latestSeedCodesJSON != mustJSON(def.SeedCodes) || latestPrioritySkillsJSON != mustJSON(def.PrioritySkills) || latestTGOsJSON != snapshotJSON {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO tree_versions (tree_id, version, title, description, seed_codes_json, priority_skills_json, tgos_json)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`, treeID, latestVersion+1, def.Title, def.Description, mustJSON(def.SeedCodes), mustJSON(def.PrioritySkills), snapshotJSON); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *Store) ListTreeVersions(ctx context.Context, treeSlug string) ([]domain.TreeVersion, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT v.id, v.tree_id, t.slug, v.version, v.title, v.description, v.created_at
		FROM tree_versions v
		JOIN tgo_trees t ON t.id = v.tree_id
		WHERE t.slug = ?
		ORDER BY v.version DESC
	`, treeSlug)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var versions []domain.TreeVersion
	for rows.Next() {
		var version domain.TreeVersion
		if err := rows.Scan(&version.ID, &version.TreeID, &version.TreeSlug, &version.Version, &version.Title, &version.Description, &version.CreatedAt); err != nil {
			return nil, err
		}
		versions = append(versions, version)
	}
	return versions, rows.Err()
}

func (s *Store) TreeVersionByNumber(ctx context.Context, treeSlug string, version int) (domain.TreeVersion, domain.TGOTreeDefinition, error) {
	var meta domain.TreeVersion
	var seedCodesJSON, prioritySkillsJSON, tgosJSON string
	err := s.SQL.QueryRowContext(ctx, `
		SELECT v.id, v.tree_id, t.slug, v.version, v.title, v.description, v.seed_codes_json, v.priority_skills_json, v.tgos_json, v.created_at
		FROM tree_versions v
		JOIN tgo_trees t ON t.id = v.tree_id
		WHERE t.slug = ? AND v.version = ?
	`, treeSlug, version).Scan(&meta.ID, &meta.TreeID, &meta.TreeSlug, &meta.Version, &meta.Title, &meta.Description, &seedCodesJSON, &prioritySkillsJSON, &tgosJSON, &meta.CreatedAt)
	if err != nil {
		return domain.TreeVersion{}, domain.TGOTreeDefinition{}, err
	}
	def := domain.TGOTreeDefinition{
		Slug:        meta.TreeSlug,
		Title:       meta.Title,
		Description: meta.Description,
	}
	if def.SeedCodes, err = DecodeStringSlice(seedCodesJSON); err != nil {
		return domain.TreeVersion{}, domain.TGOTreeDefinition{}, err
	}
	if def.PrioritySkills, err = DecodeStringSlice(prioritySkillsJSON); err != nil {
		return domain.TreeVersion{}, domain.TGOTreeDefinition{}, err
	}
	if err := json.Unmarshal([]byte(tgosJSON), &def.TGOs); err != nil {
		return domain.TreeVersion{}, domain.TGOTreeDefinition{}, err
	}
	return meta, def, nil
}

func (s *Store) ListUsers(ctx context.Context) ([]domain.User, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, slug, name, active_tree_slug, created_at
		FROM users
		ORDER BY slug ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.Slug, &user.Name, &user.ActiveTreeSlug, &user.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

func (s *Store) OnboardingProfileByUserID(ctx context.Context, userID int64) (domain.OnboardingProfile, error) {
	var profile domain.OnboardingProfile
	var weaknessesJSON, outcomesJSON string
	err := s.SQL.QueryRowContext(ctx, `
		SELECT user_id, writing_type, experience_level, desired_tone, biggest_weaknesses_json, desired_outcomes_json, difficulty_intensity, writing_goals, generated_tree_slug, template_key
		FROM user_onboarding_profiles
		WHERE user_id = ?
	`, userID).Scan(
		&profile.UserID,
		&profile.WritingType,
		&profile.ExperienceLevel,
		&profile.DesiredTone,
		&weaknessesJSON,
		&outcomesJSON,
		&profile.DifficultyIntensity,
		&profile.WritingGoals,
		&profile.GeneratedTreeSlug,
		&profile.TemplateKey,
	)
	if err != nil {
		return domain.OnboardingProfile{}, err
	}
	if profile.BiggestWeaknesses, err = DecodeStringSlice(weaknessesJSON); err != nil {
		return domain.OnboardingProfile{}, err
	}
	if profile.DesiredOutcomes, err = DecodeStringSlice(outcomesJSON); err != nil {
		return domain.OnboardingProfile{}, err
	}
	return profile, nil
}

func (s *Store) SaveOnboardingProfile(ctx context.Context, profile domain.OnboardingProfile) error {
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO user_onboarding_profiles (
			user_id, writing_type, experience_level, desired_tone, biggest_weaknesses_json, desired_outcomes_json,
			difficulty_intensity, writing_goals, generated_tree_slug, template_key, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP)
		ON CONFLICT(user_id) DO UPDATE SET
			writing_type = excluded.writing_type,
			experience_level = excluded.experience_level,
			desired_tone = excluded.desired_tone,
			biggest_weaknesses_json = excluded.biggest_weaknesses_json,
			desired_outcomes_json = excluded.desired_outcomes_json,
			difficulty_intensity = excluded.difficulty_intensity,
			writing_goals = excluded.writing_goals,
			generated_tree_slug = excluded.generated_tree_slug,
			template_key = excluded.template_key,
			updated_at = CURRENT_TIMESTAMP
	`, profile.UserID, profile.WritingType, profile.ExperienceLevel, profile.DesiredTone, mustJSON(profile.BiggestWeaknesses), mustJSON(profile.DesiredOutcomes), profile.DifficultyIntensity, profile.WritingGoals, profile.GeneratedTreeSlug, profile.TemplateKey)
	return err
}

func (s *Store) ListTrees(ctx context.Context) ([]domain.TGOTree, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, slug, title, description, created_at
		FROM tgo_trees
		ORDER BY slug ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trees []domain.TGOTree
	for rows.Next() {
		var tree domain.TGOTree
		if err := rows.Scan(&tree.ID, &tree.Slug, &tree.Title, &tree.Description, &tree.CreatedAt); err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	}
	return trees, rows.Err()
}

func (s *Store) ListEnrollments(ctx context.Context) ([]domain.Enrollment, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT e.id, e.user_id, e.tree_id, u.slug, t.slug, e.created_at
		FROM user_tree_enrollments e
		JOIN users u ON u.id = e.user_id
		JOIN tgo_trees t ON t.id = e.tree_id
		ORDER BY u.slug ASC, t.slug ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var enrollments []domain.Enrollment
	for rows.Next() {
		var enrollment domain.Enrollment
		if err := rows.Scan(&enrollment.ID, &enrollment.UserID, &enrollment.TreeID, &enrollment.UserSlug, &enrollment.TreeSlug, &enrollment.CreatedAt); err != nil {
			return nil, err
		}
		enrollments = append(enrollments, enrollment)
	}
	return enrollments, rows.Err()
}

func (s *Store) EnrollmentByID(ctx context.Context, enrollmentID int64) (domain.Enrollment, error) {
	var enrollment domain.Enrollment
	err := s.SQL.QueryRowContext(ctx, `
		SELECT e.id, e.user_id, e.tree_id, u.slug, t.slug, e.created_at
		FROM user_tree_enrollments e
		JOIN users u ON u.id = e.user_id
		JOIN tgo_trees t ON t.id = e.tree_id
		WHERE e.id = ?
	`, enrollmentID).Scan(&enrollment.ID, &enrollment.UserID, &enrollment.TreeID, &enrollment.UserSlug, &enrollment.TreeSlug, &enrollment.CreatedAt)
	return enrollment, err
}

func (s *Store) EnrollmentID(ctx context.Context, userID, treeID int64) (int64, error) {
	var id int64
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id FROM user_tree_enrollments WHERE user_id = ? AND tree_id = ?
	`, userID, treeID).Scan(&id)
	return id, err
}

func (s *Store) GetCurriculumState(ctx context.Context, enrollmentID int64) (domain.CurriculumState, error) {
	var state domain.CurriculumState
	err := s.SQL.QueryRowContext(ctx, `
		SELECT enrollment_id, current_focus, difficulty_level, COALESCE(last_review_id, 0), updated_at
		FROM user_curriculum_state
		WHERE enrollment_id = ?
	`, enrollmentID).Scan(&state.ID, &state.CurrentFocus, &state.DifficultyLevel, &state.LastReviewID, &state.UpdatedAt)
	if err != nil {
		return domain.CurriculumState{}, err
	}
	return state, nil
}

func (s *Store) SaveExercise(ctx context.Context, ex domain.Exercise) (int64, error) {
	const query = `
		INSERT INTO exercises (
			user_id,
			tree_id,
			title,
			brief,
			constraints_json,
			focus_skills_json,
			tgo_codes_json,
			success_criteria_json,
			generation_kind,
			source_submission_id
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	res, err := s.SQL.ExecContext(
		ctx,
		query,
		ex.UserID,
		ex.TreeID,
		ex.Title,
		ex.Brief,
		mustJSON(ex.Constraints),
		mustJSON(ex.FocusSkills),
		mustJSON(ex.TGOCodes),
		mustJSON(ex.SuccessCriteria),
		ex.GenerationKind,
		nullableID(ex.SourceSubmissionID),
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (s *Store) GetExercise(ctx context.Context, exerciseID int64) (domain.Exercise, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, title, brief, constraints_json, focus_skills_json, tgo_codes_json, success_criteria_json, generation_kind, COALESCE(source_submission_id, 0), created_at
		FROM exercises
		WHERE id = ?
	`, exerciseID)
	if err != nil {
		return domain.Exercise{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.Exercise{}, sql.ErrNoRows
	}
	exercise, err := scanExercise(rows)
	if err != nil {
		return domain.Exercise{}, err
	}
	return exercise, rows.Err()
}

func (s *Store) ListExercises(ctx context.Context, userID, treeID int64, limit int) ([]domain.Exercise, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, title, brief, constraints_json, focus_skills_json, tgo_codes_json, success_criteria_json, generation_kind, COALESCE(source_submission_id, 0), created_at
		FROM exercises
		WHERE user_id = ? AND tree_id = ?
		ORDER BY id DESC
		LIMIT ?
	`, userID, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exercises []domain.Exercise
	for rows.Next() {
		exercise, err := scanExercise(rows)
		if err != nil {
			return nil, err
		}
		exercises = append(exercises, exercise)
	}
	return exercises, rows.Err()
}

func (s *Store) SaveSubmission(ctx context.Context, sub domain.Submission) (int64, error) {
	if sub.DraftNumber == 0 {
		nextDraft, err := s.NextDraftNumber(ctx, sub.ExerciseID, sub.UserID, sub.TreeID)
		if err != nil {
			return 0, err
		}
		sub.DraftNumber = nextDraft
	}
	res, err := s.SQL.ExecContext(ctx, `
		INSERT INTO submissions (user_id, tree_id, exercise_id, parent_submission_id, draft_number, content, word_count)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, sub.UserID, sub.TreeID, sub.ExerciseID, nullableID(sub.ParentSubmissionID), sub.DraftNumber, sub.Content, sub.WordCount)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (s *Store) GetSubmission(ctx context.Context, submissionID int64) (domain.Submission, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE id = ?
	`, submissionID)
	if err != nil {
		return domain.Submission{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.Submission{}, sql.ErrNoRows
	}
	submission, err := scanSubmission(rows)
	if err != nil {
		return domain.Submission{}, err
	}
	return submission, rows.Err()
}

func (s *Store) ListSubmissions(ctx context.Context, userID, treeID, exerciseID int64, limit int) ([]domain.Submission, error) {
	query := `
		SELECT id, user_id, tree_id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE user_id = ? AND tree_id = ?
	`
	args := []any{userID, treeID}
	if exerciseID != 0 {
		query += " AND exercise_id = ?"
		args = append(args, exerciseID)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.SQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []domain.Submission
	for rows.Next() {
		submission, err := scanSubmission(rows)
		if err != nil {
			return nil, err
		}
		submissions = append(submissions, submission)
	}
	return submissions, rows.Err()
}

func (s *Store) NextDraftNumber(ctx context.Context, exerciseID, userID, treeID int64) (int, error) {
	var next int
	err := s.SQL.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(draft_number), 0) + 1
		FROM submissions
		WHERE exercise_id = ? AND user_id = ? AND tree_id = ?
	`, exerciseID, userID, treeID).Scan(&next)
	if err != nil {
		return 0, err
	}
	if next == 0 {
		next = 1
	}
	return next, nil
}

func (s *Store) LatestSubmissionForExercise(ctx context.Context, exerciseID, userID, treeID int64) (domain.Submission, error) {
	var sub domain.Submission
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE exercise_id = ? AND user_id = ? AND tree_id = ?
		ORDER BY draft_number DESC, id DESC
		LIMIT 1
	`, exerciseID, userID, treeID).Scan(&sub.ID, &sub.UserID, &sub.TreeID, &sub.ExerciseID, &sub.ParentSubmissionID, &sub.DraftNumber, &sub.Content, &sub.WordCount, &sub.CreatedAt)
	if err != nil {
		return domain.Submission{}, err
	}
	return sub, nil
}

func (s *Store) PreviousSubmission(ctx context.Context, sub domain.Submission) (domain.Submission, error) {
	if sub.ParentSubmissionID != 0 {
		return s.GetSubmission(ctx, sub.ParentSubmissionID)
	}
	var prev domain.Submission
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE exercise_id = ? AND user_id = ? AND tree_id = ? AND draft_number < ?
		ORDER BY draft_number DESC, id DESC
		LIMIT 1
	`, sub.ExerciseID, sub.UserID, sub.TreeID, sub.DraftNumber).Scan(&prev.ID, &prev.UserID, &prev.TreeID, &prev.ExerciseID, &prev.ParentSubmissionID, &prev.DraftNumber, &prev.Content, &prev.WordCount, &prev.CreatedAt)
	if err != nil {
		return domain.Submission{}, err
	}
	return prev, nil
}

func (s *Store) SaveReview(ctx context.Context, review domain.Review, scores []domain.SkillScore) (int64, error) {
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	res, err := tx.ExecContext(ctx, `
		INSERT INTO reviews (
			user_id,
			tree_id,
			submission_id,
			review_kind,
			summary,
			strengths_json,
			weaknesses_json,
			analyzer_findings_json,
			completed_tgo_checks_json,
			next_focus,
			metric_word_count
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		review.UserID,
		review.TreeID,
		review.SubmissionID,
		review.ReviewKind,
		review.Summary,
		mustJSON(review.Strengths),
		mustJSON(review.Weaknesses),
		mustJSON(review.AnalyzerFindings),
		mustJSON(review.CompletedTGOChecks),
		review.NextFocus,
		review.MetricWordCount,
	)
	if err != nil {
		return 0, err
	}

	reviewID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	for _, score := range scores {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO submission_skill_scores (submission_id, skill_name, score)
			VALUES (?, ?, ?)
		`, review.SubmissionID, score.Skill, score.Score); err != nil {
			return 0, err
		}
	}

	for _, assessment := range review.TGOAssessments {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO review_tgo_assessments (review_id, submission_id, tgo_code, status, evidence)
			VALUES (?, ?, ?, ?, ?)
		`, reviewID, review.SubmissionID, assessment.TGOCode, assessment.Status, assessment.Evidence); err != nil {
			return 0, err
		}
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return reviewID, nil
}

func (s *Store) SaveReviewArtifacts(ctx context.Context, artifacts domain.ReviewArtifacts) error {
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO review_artifacts (review_id, analyzer_report_json, recommendation_json, comparison_json, annotations_json)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(review_id) DO UPDATE SET
			analyzer_report_json = excluded.analyzer_report_json,
			recommendation_json = excluded.recommendation_json,
			comparison_json = excluded.comparison_json,
			annotations_json = excluded.annotations_json
	`, artifacts.ReviewID, artifacts.AnalyzerReportJSON, artifacts.RecommendationJSON, artifacts.ComparisonJSON, artifacts.AnnotationsJSON)
	return err
}

func (s *Store) ActiveTGOs(ctx context.Context, enrollmentID int64) ([]domain.TGO, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT c.id, c.code, c.title, c.description, c.stage, c.stage_order, a.slot, tt.prerequisites_json, tt.mastery_hint
		FROM enrollment_active_tgos a
		JOIN tgo_catalog c ON c.code = a.tgo_code
		LEFT JOIN user_tree_enrollments e ON e.id = a.enrollment_id
		LEFT JOIN tree_tgos tt ON tt.tree_id = e.tree_id AND tt.tgo_code = c.code
		WHERE a.enrollment_id = ?
		ORDER BY a.slot ASC
	`, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tgos []domain.TGO
	for rows.Next() {
		var tgo domain.TGO
		var prereqsJSON string
		if err := rows.Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder, &tgo.ActiveSlot, &prereqsJSON, &tgo.MasteryHint); err != nil {
			return nil, err
		}
		if tgo.Prerequisites, err = DecodeStringSlice(prereqsJSON); err != nil {
			return nil, err
		}
		tgos = append(tgos, tgo)
	}
	return tgos, rows.Err()
}

func (s *Store) CompletedTGOs(ctx context.Context, enrollmentID int64) ([]domain.TGO, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT c.id, c.code, c.title, c.description, c.stage, c.stage_order, 0, tt.prerequisites_json, tt.mastery_hint
		FROM enrollment_completed_tgos x
		JOIN tgo_catalog c ON c.code = x.tgo_code
		LEFT JOIN user_tree_enrollments e ON e.id = x.enrollment_id
		LEFT JOIN tree_tgos tt ON tt.tree_id = e.tree_id AND tt.tgo_code = c.code
		WHERE x.enrollment_id = ?
		ORDER BY x.completed_at ASC
	`, enrollmentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tgos []domain.TGO
	for rows.Next() {
		var tgo domain.TGO
		var prereqsJSON string
		if err := rows.Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder, &tgo.ActiveSlot, &prereqsJSON, &tgo.MasteryHint); err != nil {
			return nil, err
		}
		if tgo.Prerequisites, err = DecodeStringSlice(prereqsJSON); err != nil {
			return nil, err
		}
		tgos = append(tgos, tgo)
	}
	return tgos, rows.Err()
}

func (s *Store) RecentTGOStatuses(ctx context.Context, enrollmentID int64, code string, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT status
		FROM review_tgo_assessments
		WHERE tgo_code = ? AND EXISTS (
			SELECT 1 FROM reviews r WHERE r.id = review_tgo_assessments.review_id AND r.user_id = (
				SELECT user_id FROM user_tree_enrollments WHERE id = ?
			) AND r.tree_id = (
				SELECT tree_id FROM user_tree_enrollments WHERE id = ?
			)
		)
		ORDER BY id DESC
		LIMIT ?
	`, code, enrollmentID, enrollmentID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var status string
		if err := rows.Scan(&status); err != nil {
			return nil, err
		}
		out = append(out, status)
	}
	return out, rows.Err()
}

func (s *Store) ReplaceActiveTGO(ctx context.Context, enrollmentID int64, slot int, completedCode string, nextCode string) error {
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `
		INSERT INTO enrollment_completed_tgos (enrollment_id, tgo_code)
		SELECT ?, ?
		WHERE NOT EXISTS (SELECT 1 FROM enrollment_completed_tgos WHERE enrollment_id = ? AND tgo_code = ?)
	`, enrollmentID, completedCode, enrollmentID, completedCode); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE enrollment_active_tgos SET tgo_code = ?, activated_at = CURRENT_TIMESTAMP WHERE enrollment_id = ? AND slot = ?
	`, nextCode, enrollmentID, slot); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) SetActiveTGOs(ctx context.Context, enrollmentID int64, codes []string) error {
	if len(codes) != 3 {
		return fmt.Errorf("exactly 3 TGO codes are required")
	}
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err = tx.ExecContext(ctx, `DELETE FROM enrollment_active_tgos WHERE enrollment_id = ?`, enrollmentID); err != nil {
		return err
	}
	for i, code := range codes {
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO enrollment_active_tgos (enrollment_id, slot, tgo_code)
			VALUES (?, ?, ?)
		`, enrollmentID, i+1, code); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) NextAvailableTGO(ctx context.Context, enrollmentID int64) (domain.TGO, error) {
	var tgo domain.TGO
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, code, title, description, stage, stage_order
		FROM tgo_catalog
		WHERE code NOT IN (SELECT tgo_code FROM enrollment_active_tgos WHERE enrollment_id = ?)
		  AND code NOT IN (SELECT tgo_code FROM enrollment_completed_tgos WHERE enrollment_id = ?)
		ORDER BY stage_order ASC
		LIMIT 1
	`, enrollmentID, enrollmentID).Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder)
	if err != nil {
		return domain.TGO{}, err
	}
	return tgo, nil
}

func (s *Store) LatestReviewForSubmission(ctx context.Context, submissionID int64) (domain.Review, error) {
	var review domain.Review
	var strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON string
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, submission_id, review_kind, summary, strengths_json, weaknesses_json, analyzer_findings_json, completed_tgo_checks_json, next_focus, metric_word_count, created_at
		FROM reviews
		WHERE submission_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, submissionID).Scan(
		&review.ID,
		&review.UserID,
		&review.TreeID,
		&review.SubmissionID,
		&review.ReviewKind,
		&review.Summary,
		&strengthsJSON,
		&weaknessesJSON,
		&findingsJSON,
		&completedChecksJSON,
		&review.NextFocus,
		&review.MetricWordCount,
		&review.CreatedAt,
	)
	if err != nil {
		return domain.Review{}, err
	}
	review, err = hydrateReview(review, strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON)
	if err != nil {
		return domain.Review{}, err
	}
	review.TGOAssessments, err = s.ReviewTGOAssessments(ctx, review.ID)
	if err != nil {
		return domain.Review{}, err
	}
	return review, nil
}

func (s *Store) GetReview(ctx context.Context, reviewID int64) (domain.Review, error) {
	var review domain.Review
	var strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON string
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, user_id, tree_id, submission_id, review_kind, summary, strengths_json, weaknesses_json, analyzer_findings_json, completed_tgo_checks_json, next_focus, metric_word_count, created_at
		FROM reviews
		WHERE id = ?
	`, reviewID).Scan(
		&review.ID,
		&review.UserID,
		&review.TreeID,
		&review.SubmissionID,
		&review.ReviewKind,
		&review.Summary,
		&strengthsJSON,
		&weaknessesJSON,
		&findingsJSON,
		&completedChecksJSON,
		&review.NextFocus,
		&review.MetricWordCount,
		&review.CreatedAt,
	)
	if err != nil {
		return domain.Review{}, err
	}
	review, err = hydrateReview(review, strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON)
	if err != nil {
		return domain.Review{}, err
	}
	review.TGOAssessments, err = s.ReviewTGOAssessments(ctx, review.ID)
	if err != nil {
		return domain.Review{}, err
	}
	return review, nil
}

func (s *Store) ListReviews(ctx context.Context, userID, treeID, submissionID int64, limit int) ([]domain.Review, error) {
	query := `
		SELECT id, user_id, tree_id, submission_id, review_kind, summary, strengths_json, weaknesses_json, analyzer_findings_json, completed_tgo_checks_json, next_focus, metric_word_count, created_at
		FROM reviews
		WHERE user_id = ? AND tree_id = ?
	`
	args := []any{userID, treeID}
	if submissionID != 0 {
		query += " AND submission_id = ?"
		args = append(args, submissionID)
	}
	query += " ORDER BY id DESC LIMIT ?"
	args = append(args, limit)

	rows, err := s.SQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []domain.Review
	for rows.Next() {
		review, err := scanReview(rows)
		if err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range reviews {
		assessments, err := s.ReviewTGOAssessments(ctx, reviews[i].ID)
		if err != nil {
			return nil, err
		}
		reviews[i].TGOAssessments = assessments
	}
	return reviews, nil
}

func (s *Store) ReviewTGOAssessments(ctx context.Context, reviewID int64) ([]domain.TGOAssessment, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT tgo_code, status, evidence
		FROM review_tgo_assessments
		WHERE review_id = ?
		ORDER BY id ASC
	`, reviewID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assessments []domain.TGOAssessment
	for rows.Next() {
		var assessment domain.TGOAssessment
		if err := rows.Scan(&assessment.TGOCode, &assessment.Status, &assessment.Evidence); err != nil {
			return nil, err
		}
		assessment.ReviewID = reviewID
		assessments = append(assessments, assessment)
	}
	return assessments, rows.Err()
}

func (s *Store) GetReviewArtifacts(ctx context.Context, reviewID int64) (domain.ReviewArtifacts, error) {
	var artifacts domain.ReviewArtifacts
	err := s.SQL.QueryRowContext(ctx, `
		SELECT review_id, analyzer_report_json, recommendation_json, comparison_json, annotations_json, created_at
		FROM review_artifacts
		WHERE review_id = ?
	`, reviewID).Scan(
		&artifacts.ReviewID,
		&artifacts.AnalyzerReportJSON,
		&artifacts.RecommendationJSON,
		&artifacts.ComparisonJSON,
		&artifacts.AnnotationsJSON,
		&artifacts.CreatedAt,
	)
	if err != nil {
		return domain.ReviewArtifacts{}, err
	}
	return artifacts, nil
}

func (s *Store) UpdateCurriculumState(ctx context.Context, enrollmentID int64, focus string, difficulty int, reviewID int64) error {
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE user_curriculum_state
		SET current_focus = ?, difficulty_level = ?, last_review_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE enrollment_id = ?
	`, focus, difficulty, reviewID, enrollmentID)
	return err
}

func (s *Store) SkillAverages(ctx context.Context, userID, treeID int64, limit int) (map[string]float64, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		WITH recent_submissions AS (
			SELECT DISTINCT sss.submission_id
			FROM submission_skill_scores sss
			JOIN submissions s ON s.id = sss.submission_id
			WHERE s.user_id = ? AND s.tree_id = ?
			ORDER BY sss.submission_id DESC
			LIMIT ?
		)
		SELECT skill_name, AVG(score)
		FROM submission_skill_scores
		WHERE submission_id IN (SELECT submission_id FROM recent_submissions)
		GROUP BY skill_name
	`, userID, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := map[string]float64{}
	for rows.Next() {
		var skill string
		var avg float64
		if err := rows.Scan(&skill, &avg); err != nil {
			return nil, err
		}
		values[skill] = avg
	}
	return values, rows.Err()
}

func (s *Store) RecentSkillScores(ctx context.Context, userID, treeID int64, skill string, limit int) ([]int, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT sss.score
		FROM submission_skill_scores sss
		JOIN submissions s ON s.id = sss.submission_id
		WHERE sss.skill_name = ? AND s.user_id = ? AND s.tree_id = ?
		ORDER BY sss.submission_id DESC
		LIMIT ?
	`, skill, userID, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scores []int
	for rows.Next() {
		var score int
		if err := rows.Scan(&score); err != nil {
			return nil, err
		}
		scores = append(scores, score)
	}
	return scores, rows.Err()
}

func (s *Store) LatestSkillScores(ctx context.Context, submissionID int64) (map[string]int, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT skill_name, score
		FROM submission_skill_scores
		WHERE submission_id = ?
	`, submissionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	values := map[string]int{}
	for rows.Next() {
		var skill string
		var score int
		if err := rows.Scan(&skill, &score); err != nil {
			return nil, err
		}
		values[skill] = score
	}
	return values, rows.Err()
}

func (s *Store) ProgressReport(ctx context.Context, userID, treeID int64, prioritySkills []string, limit int) ([]string, error) {
	averages, err := s.SkillAverages(ctx, userID, treeID, limit)
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, skill := range prioritySkills {
		avg, ok := averages[skill]
		if !ok {
			continue
		}
		recent, err := s.RecentSkillScores(ctx, userID, treeID, skill, 2)
		if err != nil {
			return nil, err
		}
		trend := "flat"
		if len(recent) >= 2 {
			switch {
			case recent[0] > recent[1]:
				trend = "up"
			case recent[0] < recent[1]:
				trend = "down"
			}
		}
		lines = append(lines, fmt.Sprintf("%s | avg %.2f | trend %s", skill, avg, trend))
	}
	return lines, nil
}

func (s *Store) RecurringWeaknesses(ctx context.Context, userID, treeID int64, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT r.weaknesses_json
		FROM reviews r
		WHERE r.user_id = ? AND r.tree_id = ?
		ORDER BY id DESC
		LIMIT ?
	`, userID, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectRecurringJSONStrings(rows, 3)
}

func (s *Store) RecurringAnalyzerFindings(ctx context.Context, userID, treeID int64, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT r.analyzer_findings_json
		FROM reviews r
		WHERE r.user_id = ? AND r.tree_id = ?
		ORDER BY id DESC
		LIMIT ?
	`, userID, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectRecurringJSONStrings(rows, 4)
}

func (s *Store) StrongestWeakestSkills(ctx context.Context, userID, treeID int64, prioritySkills []string, limit int) ([]string, []string, error) {
	averages, err := s.SkillAverages(ctx, userID, treeID, limit)
	if err != nil {
		return nil, nil, err
	}
	type pair struct {
		skill string
		avg   float64
	}
	var pairs []pair
	for skill, avg := range averages {
		if !domain.IsSupportedSkill(skill) {
			continue
		}
		pairs = append(pairs, pair{skill: skill, avg: avg})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].avg == pairs[j].avg {
			return skillPriority(prioritySkills, pairs[i].skill) > skillPriority(prioritySkills, pairs[j].skill)
		}
		return pairs[i].avg > pairs[j].avg
	})
	var strongest []string
	var weakest []string
	for i := 0; i < len(pairs) && i < 3; i++ {
		strongest = append(strongest, fmt.Sprintf("%s (%.2f)", pairs[i].skill, pairs[i].avg))
	}
	for i := len(pairs) - 1; i >= 0 && len(weakest) < 3; i-- {
		weakest = append(weakest, fmt.Sprintf("%s (%.2f)", pairs[i].skill, pairs[i].avg))
	}
	return strongest, weakest, nil
}

func (s *Store) RecurringCompletedTGOSlips(ctx context.Context, userID, treeID int64, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT r.completed_tgo_checks_json
		FROM reviews r
		WHERE r.user_id = ? AND r.tree_id = ?
		ORDER BY r.id DESC
		LIMIT ?
	`, userID, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var checks []domain.TGOAssessment
		if err := json.Unmarshal([]byte(raw), &checks); err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		for _, check := range checks {
			if check.Status != "slipping" || check.TGOCode == "" || seen[check.TGOCode] {
				continue
			}
			seen[check.TGOCode] = true
			if tgo, ok := domain.TGOByCode(check.TGOCode); ok {
				counts[tgo.Title]++
				continue
			}
			counts[check.TGOCode]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	type pair struct {
		text  string
		count int
	}
	var pairs []pair
	for text, count := range counts {
		pairs = append(pairs, pair{text: text, count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].text < pairs[j].text
		}
		return pairs[i].count > pairs[j].count
	})
	var out []string
	for _, pair := range pairs {
		out = append(out, pair.text)
		if len(out) == 3 {
			break
		}
	}
	return out, nil
}

func (s *Store) History(ctx context.Context, userID, treeID int64) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT e.id, e.title, s.id, r.id, COALESCE(r.next_focus, '')
		FROM exercises e
		LEFT JOIN submissions s ON s.exercise_id = e.id AND s.user_id = e.user_id AND s.tree_id = e.tree_id
		LEFT JOIN reviews r ON r.submission_id = s.id AND r.user_id = e.user_id AND r.tree_id = e.tree_id
		WHERE e.user_id = ? AND e.tree_id = ?
		ORDER BY e.id DESC
		LIMIT 10
	`, userID, treeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var (
			exerciseID   int64
			title        string
			submissionID sql.NullInt64
			reviewID     sql.NullInt64
			nextFocus    string
		)

		if err := rows.Scan(&exerciseID, &title, &submissionID, &reviewID, &nextFocus); err != nil {
			return nil, err
		}

		item := fmt.Sprintf("exercise %d: %s", exerciseID, title)
		if submissionID.Valid {
			item += fmt.Sprintf(" | submission %d", submissionID.Int64)
		}
		if reviewID.Valid {
			item += fmt.Sprintf(" | review %d | next focus: %s", reviewID.Int64, nextFocus)
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *Store) RecentExerciseTitles(ctx context.Context, userID, treeID int64, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT title
		FROM exercises
		WHERE user_id = ? AND tree_id = ?
		ORDER BY id DESC
		LIMIT ?
	`, userID, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var titles []string
	for rows.Next() {
		var title string
		if err := rows.Scan(&title); err != nil {
			return nil, err
		}
		titles = append(titles, title)
	}
	return titles, rows.Err()
}

func mustJSON(v any) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}

func NextDifficulty(nextFocus string) int {
	if strings.Contains(nextFocus, "mythic") || strings.Contains(nextFocus, "symbol") {
		return 2
	}
	if strings.Contains(nextFocus, "tragic") || strings.Contains(nextFocus, "compression") {
		return 2
	}
	return 1
}

func DecodeStringSlice(raw string) ([]string, error) {
	if raw == "" {
		return nil, nil
	}

	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return nil, err
	}
	return values, nil
}

func skillPriority(prioritySkills []string, skill string) int {
	for idx, value := range prioritySkills {
		if value == skill {
			return len(prioritySkills) - idx
		}
	}
	return 0
}

func scanExercise(scanner interface{ Scan(...any) error }) (domain.Exercise, error) {
	var exercise domain.Exercise
	var constraintsJSON, focusSkillsJSON, tgoCodesJSON, successJSON string
	if err := scanner.Scan(
		&exercise.ID,
		&exercise.UserID,
		&exercise.TreeID,
		&exercise.Title,
		&exercise.Brief,
		&constraintsJSON,
		&focusSkillsJSON,
		&tgoCodesJSON,
		&successJSON,
		&exercise.GenerationKind,
		&exercise.SourceSubmissionID,
		&exercise.CreatedAt,
	); err != nil {
		return domain.Exercise{}, err
	}
	var err error
	exercise.Constraints, err = DecodeStringSlice(constraintsJSON)
	if err != nil {
		return domain.Exercise{}, err
	}
	exercise.FocusSkills, err = DecodeStringSlice(focusSkillsJSON)
	if err != nil {
		return domain.Exercise{}, err
	}
	exercise.TGOCodes, err = DecodeStringSlice(tgoCodesJSON)
	if err != nil {
		return domain.Exercise{}, err
	}
	exercise.SuccessCriteria, err = DecodeStringSlice(successJSON)
	if err != nil {
		return domain.Exercise{}, err
	}
	return exercise, nil
}

func scanSubmission(scanner interface{ Scan(...any) error }) (domain.Submission, error) {
	var submission domain.Submission
	if err := scanner.Scan(
		&submission.ID,
		&submission.UserID,
		&submission.TreeID,
		&submission.ExerciseID,
		&submission.ParentSubmissionID,
		&submission.DraftNumber,
		&submission.Content,
		&submission.WordCount,
		&submission.CreatedAt,
	); err != nil {
		return domain.Submission{}, err
	}
	return submission, nil
}

func scanReview(scanner interface{ Scan(...any) error }) (domain.Review, error) {
	var review domain.Review
	var strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON string
	if err := scanner.Scan(
		&review.ID,
		&review.UserID,
		&review.TreeID,
		&review.SubmissionID,
		&review.ReviewKind,
		&review.Summary,
		&strengthsJSON,
		&weaknessesJSON,
		&findingsJSON,
		&completedChecksJSON,
		&review.NextFocus,
		&review.MetricWordCount,
		&review.CreatedAt,
	); err != nil {
		return domain.Review{}, err
	}
	return hydrateReview(review, strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON)
}

func hydrateReview(review domain.Review, strengthsJSON, weaknessesJSON, findingsJSON, completedChecksJSON string) (domain.Review, error) {
	var err error
	review.Strengths, err = DecodeStringSlice(strengthsJSON)
	if err != nil {
		return domain.Review{}, err
	}
	review.Weaknesses, err = DecodeStringSlice(weaknessesJSON)
	if err != nil {
		return domain.Review{}, err
	}
	review.AnalyzerFindings, err = DecodeStringSlice(findingsJSON)
	if err != nil {
		return domain.Review{}, err
	}
	if err := json.Unmarshal([]byte(completedChecksJSON), &review.CompletedTGOChecks); err != nil {
		return domain.Review{}, err
	}
	return review, nil
}

func IsNotFound(err error) bool {
	return errors.Is(err, sql.ErrNoRows)
}

func CountWords(text string) int {
	return len(strings.Fields(text))
}

func Since(t time.Time) string {
	if t.IsZero() {
		return "unknown"
	}
	return t.Format(time.RFC3339)
}

func nullableID(id int64) any {
	if id == 0 {
		return nil
	}
	return id
}

func collectRecurringJSONStrings(rows *sql.Rows, top int) ([]string, error) {
	counts := map[string]int{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		values, err := DecodeStringSlice(raw)
		if err != nil {
			return nil, err
		}
		seen := map[string]bool{}
		for _, value := range values {
			if value == "" || seen[value] {
				continue
			}
			seen[value] = true
			counts[value]++
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	type pair struct {
		text  string
		count int
	}
	var pairs []pair
	for text, count := range counts {
		pairs = append(pairs, pair{text: text, count: count})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].count == pairs[j].count {
			return pairs[i].text < pairs[j].text
		}
		return pairs[i].count > pairs[j].count
	})
	var out []string
	for _, pair := range pairs {
		out = append(out, pair.text)
		if len(out) == top {
			break
		}
	}
	return out, nil
}
