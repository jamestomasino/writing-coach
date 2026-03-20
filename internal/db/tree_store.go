package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

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
		SELECT c.id, c.code, c.title, c.description, c.stage, c.stage_order, c.progress_mode, tt.prerequisites_json, tt.mastery_hint
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
		if err := rows.Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder, &tgo.ProgressMode, &prereqsJSON, &tgo.MasteryHint); err != nil {
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
		progressMode := tgo.ProgressMode
		if progressMode == "" {
			progressMode = domain.ProgressModeStage
		}
		if _, err = tx.ExecContext(ctx, `
			INSERT INTO tgo_catalog (code, title, description, stage, stage_order, progress_mode)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(code) DO UPDATE SET
				title = excluded.title,
				description = excluded.description,
				stage = excluded.stage,
				stage_order = excluded.stage_order,
				progress_mode = excluded.progress_mode
		`, tgo.Code, tgo.Title, tgo.Description, tgo.Stage, tgo.StageOrder, progressMode); err != nil {
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
