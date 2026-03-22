package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/tomasino/writing-coach/internal/domain"
)

func (s *Store) SavePlaygroundSession(ctx context.Context, session domain.PlaygroundSession) (int64, error) {
	res, err := s.SQL.ExecContext(ctx, `
		INSERT INTO playground_sessions (
			user_id, tree_id, title, content, writing_language, writing_type, assignment_format, coaching_brief
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, session.UserID, session.TreeID, session.Title, session.Content, session.WritingLanguage, session.WritingType, session.AssignmentFormat, session.CoachingBrief)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) UpdatePlaygroundSession(ctx context.Context, session domain.PlaygroundSession) error {
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE playground_sessions
		SET title = ?,
			content = ?,
			writing_language = ?,
			writing_type = ?,
			assignment_format = ?,
			coaching_brief = ?,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND tree_id = ?
	`, session.Title, session.Content, session.WritingLanguage, session.WritingType, session.AssignmentFormat, session.CoachingBrief, session.ID, session.UserID, session.TreeID)
	return err
}

func (s *Store) GetPlaygroundSession(ctx context.Context, sessionID int64) (domain.PlaygroundSession, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, title, content, writing_language, writing_type, assignment_format, coaching_brief,
			COALESCE(latest_draft_id, 0), COALESCE(latest_review_id, 0), latest_review_at, draft_count, review_count, created_at, updated_at
		FROM playground_sessions
		WHERE id = ?
	`, sessionID)
	if err != nil {
		return domain.PlaygroundSession{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.PlaygroundSession{}, sql.ErrNoRows
	}
	return scanPlaygroundSession(rows)
}

func (s *Store) ListPlaygroundSessions(ctx context.Context, userID, treeID int64, limit int) ([]domain.PlaygroundSession, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, title, content, writing_language, writing_type, assignment_format, coaching_brief,
			COALESCE(latest_draft_id, 0), COALESCE(latest_review_id, 0), latest_review_at, draft_count, review_count, created_at, updated_at
		FROM playground_sessions
		WHERE user_id = ? AND tree_id = ?
		ORDER BY updated_at DESC, id DESC
		LIMIT ?
	`, userID, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.PlaygroundSession
	for rows.Next() {
		item, err := scanPlaygroundSession(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) SavePlaygroundDraft(ctx context.Context, item domain.PlaygroundDraft) (int64, error) {
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
		INSERT INTO playground_drafts (
			session_id, user_id, tree_id, parent_draft_id, content, word_count
		)
		VALUES (?, ?, ?, ?, ?, ?)
	`, item.SessionID, item.UserID, item.TreeID, nullableID(item.ParentDraftID), item.Content, item.WordCount)
	if err != nil {
		return 0, err
	}
	draftID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE playground_sessions
		SET latest_draft_id = ?,
			draft_count = draft_count + 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND tree_id = ?
	`, draftID, item.SessionID, item.UserID, item.TreeID); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return draftID, nil
}

func (s *Store) GetPlaygroundDraft(ctx context.Context, draftID int64) (domain.PlaygroundDraft, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, session_id, user_id, tree_id, COALESCE(parent_draft_id, 0), content, word_count, created_at
		FROM playground_drafts
		WHERE id = ?
	`, draftID)
	if err != nil {
		return domain.PlaygroundDraft{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.PlaygroundDraft{}, sql.ErrNoRows
	}
	return scanPlaygroundDraft(rows)
}

func (s *Store) LatestPlaygroundDraftBySession(ctx context.Context, sessionID int64) (domain.PlaygroundDraft, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, session_id, user_id, tree_id, COALESCE(parent_draft_id, 0), content, word_count, created_at
		FROM playground_drafts
		WHERE session_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, sessionID)
	if err != nil {
		return domain.PlaygroundDraft{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.PlaygroundDraft{}, sql.ErrNoRows
	}
	return scanPlaygroundDraft(rows)
}

func (s *Store) ListPlaygroundDrafts(ctx context.Context, userID, treeID, sessionID int64, limit int) ([]domain.PlaygroundDraft, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, session_id, user_id, tree_id, COALESCE(parent_draft_id, 0), content, word_count, created_at
		FROM playground_drafts
		WHERE user_id = ? AND tree_id = ? AND session_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, userID, treeID, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.PlaygroundDraft
	for rows.Next() {
		item, err := scanPlaygroundDraft(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) EnsurePlaygroundDraftSnapshot(ctx context.Context, session domain.PlaygroundSession) (domain.PlaygroundDraft, error) {
	content := session.Content
	if latest, err := s.LatestPlaygroundDraftBySession(ctx, session.ID); err == nil {
		if latest.Content == content {
			return latest, nil
		}
		draftID, err := s.SavePlaygroundDraft(ctx, domain.PlaygroundDraft{
			SessionID:     session.ID,
			UserID:        session.UserID,
			TreeID:        session.TreeID,
			ParentDraftID: latest.ID,
			Content:       content,
			WordCount:     CountWords(content),
		})
		if err != nil {
			return domain.PlaygroundDraft{}, err
		}
		return s.GetPlaygroundDraft(ctx, draftID)
	} else if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return domain.PlaygroundDraft{}, err
	}
	draftID, err := s.SavePlaygroundDraft(ctx, domain.PlaygroundDraft{
		SessionID: session.ID,
		UserID:    session.UserID,
		TreeID:    session.TreeID,
		Content:   content,
		WordCount: CountWords(content),
	})
	if err != nil {
		return domain.PlaygroundDraft{}, err
	}
	return s.GetPlaygroundDraft(ctx, draftID)
}

func (s *Store) SavePlaygroundReview(ctx context.Context, item domain.PlaygroundReview) (int64, error) {
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
		INSERT INTO playground_reviews (
			session_id, draft_id, user_id, tree_id, review_kind, provider_note, summary,
			strengths_json, weaknesses_json, analyzer_findings_json, next_focus, metric_word_count,
			skill_scores_json, tgo_assessments_json, completed_tgo_checks_json, annotations_json,
			analyzer_report_json, comparison_json
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, item.SessionID, nullableID(item.DraftID), item.UserID, item.TreeID, item.Review.ReviewKind, item.Review.ProviderNote, item.Review.Summary,
		mustJSON(item.Review.Strengths), mustJSON(item.Review.Weaknesses), mustJSON(item.Review.AnalyzerFindings), item.Review.NextFocus, item.Review.MetricWordCount,
		mustJSON(item.Review.SkillScores), mustJSON(item.Review.TGOAssessments), mustJSON(item.Review.CompletedTGOChecks), mustJSON(item.Review.Annotations),
		emptyJSONObject(item.AnalyzerReportJSON), emptyJSONObject(item.ComparisonJSON))
	if err != nil {
		return 0, err
	}
	reviewID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE playground_sessions
		SET latest_review_id = ?,
			latest_review_at = CURRENT_TIMESTAMP,
			review_count = review_count + 1,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = ? AND user_id = ? AND tree_id = ?
	`, reviewID, item.SessionID, item.UserID, item.TreeID); err != nil {
		return 0, err
	}
	if err = tx.Commit(); err != nil {
		return 0, err
	}
	return reviewID, nil
}

func (s *Store) GetPlaygroundReview(ctx context.Context, reviewID int64) (domain.PlaygroundReview, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, session_id, COALESCE(draft_id, 0), user_id, tree_id, id, review_kind, provider_note, summary,
			strengths_json, weaknesses_json, analyzer_findings_json, next_focus, metric_word_count,
			skill_scores_json, tgo_assessments_json, completed_tgo_checks_json, annotations_json,
			analyzer_report_json, comparison_json, created_at
		FROM playground_reviews
		WHERE id = ?
	`, reviewID)
	if err != nil {
		return domain.PlaygroundReview{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.PlaygroundReview{}, sql.ErrNoRows
	}
	return scanPlaygroundReview(rows)
}

func (s *Store) ListPlaygroundReviews(ctx context.Context, userID, treeID, sessionID int64, limit int) ([]domain.PlaygroundReview, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, session_id, COALESCE(draft_id, 0), user_id, tree_id, id, review_kind, provider_note, summary,
			strengths_json, weaknesses_json, analyzer_findings_json, next_focus, metric_word_count,
			skill_scores_json, tgo_assessments_json, completed_tgo_checks_json, annotations_json,
			analyzer_report_json, comparison_json, created_at
		FROM playground_reviews
		WHERE user_id = ? AND tree_id = ? AND session_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT ?
	`, userID, treeID, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var items []domain.PlaygroundReview
	for rows.Next() {
		item, err := scanPlaygroundReview(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) LatestPlaygroundReviewForDraft(ctx context.Context, draftID int64) (domain.PlaygroundReview, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, session_id, COALESCE(draft_id, 0), user_id, tree_id, id, review_kind, provider_note, summary,
			strengths_json, weaknesses_json, analyzer_findings_json, next_focus, metric_word_count,
			skill_scores_json, tgo_assessments_json, completed_tgo_checks_json, annotations_json,
			analyzer_report_json, comparison_json, created_at
		FROM playground_reviews
		WHERE draft_id = ?
		ORDER BY created_at DESC, id DESC
		LIMIT 1
	`, draftID)
	if err != nil {
		return domain.PlaygroundReview{}, err
	}
	defer rows.Close()
	if !rows.Next() {
		return domain.PlaygroundReview{}, sql.ErrNoRows
	}
	return scanPlaygroundReview(rows)
}

func emptyJSONObject(raw string) string {
	if raw == "" {
		return "{}"
	}
	var decoded map[string]any
	if err := json.Unmarshal([]byte(raw), &decoded); err != nil {
		return "{}"
	}
	return raw
}
