package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

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
	treeSlug = domain.NormalizeTreeSlug(treeSlug)
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
		WHERE NOT EXISTS (SELECT 1 FROM user_tree_enrollments WHERE user_id = ? AND tree_id = ? AND archived_at IS NULL)
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
	if err := s.reconcileEnrollmentSeedSet(ctx, enrollmentID, treeDef); err != nil {
		return 0, 0, 0, err
	}
	if user.ActiveTreeSlug == "" {
		if err := s.SetUserActiveTree(ctx, user.ID, treeSlug); err != nil {
			return 0, 0, 0, err
		}
	}

	return user.ID, tree.ID, enrollmentID, nil
}

func (s *Store) reconcileEnrollmentSeedSet(ctx context.Context, enrollmentID int64, treeDef domain.TGOTreeDefinition) error {
	completed, err := s.CompletedTGOs(ctx, enrollmentID)
	if err != nil {
		return err
	}
	if len(completed) > 0 {
		return nil
	}

	active, err := s.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		return err
	}
	if len(active) != len(treeDef.SeedCodes) {
		return nil
	}

	currentCodes := make([]string, 0, len(active))
	currentSet := make(map[string]bool, len(active))
	for _, tgo := range active {
		currentCodes = append(currentCodes, tgo.Code)
		currentSet[tgo.Code] = true
	}

	desiredSet := make(map[string]bool, len(treeDef.SeedCodes))
	var missing []string
	for _, code := range treeDef.SeedCodes {
		desiredSet[code] = true
		if !currentSet[code] {
			missing = append(missing, code)
		}
	}
	if len(missing) == 0 {
		return nil
	}

	type candidate struct {
		index    int
		percent  int
		evidence int
	}
	var replaceable []candidate
	for idx, tgo := range active {
		if desiredSet[tgo.Code] {
			continue
		}
		signal, err := s.TGOMasterySignal(ctx, enrollmentID, tgo, "")
		if err != nil {
			return err
		}
		replaceable = append(replaceable, candidate{
			index:    idx,
			percent:  signal.Percent,
			evidence: signal.EvidenceCount,
		})
	}
	if len(replaceable) == 0 {
		return nil
	}

	slices.SortFunc(replaceable, func(a, b candidate) int {
		if a.evidence != b.evidence {
			return a.evidence - b.evidence
		}
		if a.percent != b.percent {
			return a.percent - b.percent
		}
		return a.index - b.index
	})

	updated := append([]string(nil), currentCodes...)
	for i, code := range missing {
		if i >= len(replaceable) {
			break
		}
		updated[replaceable[i].index] = code
	}
	return s.SetActiveTGOs(ctx, enrollmentID, updated)
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
