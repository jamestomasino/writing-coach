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

	for _, tgo := range domain.TGOCatalog {
		if _, err := s.SQL.ExecContext(ctx, `
			INSERT INTO tgo_catalog (code, title, description, stage, stage_order)
			SELECT ?, ?, ?, ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM tgo_catalog WHERE code = ?)
		`, tgo.Code, tgo.Title, tgo.Description, tgo.Stage, tgo.StageOrder, tgo.Code); err != nil {
			return err
		}
	}

	initial := domain.SeedTGOs()
	for idx, code := range initial {
		slot := idx + 1
		if _, err := s.SQL.ExecContext(ctx, `
			INSERT INTO active_tgos (slot, tgo_code)
			SELECT ?, ?
			WHERE NOT EXISTS (SELECT 1 FROM active_tgos WHERE slot = ?)
		`, slot, code, slot); err != nil {
			return err
		}
	}

	return nil
}

func (s *Store) GetCurriculumState(ctx context.Context) (domain.CurriculumState, error) {
	var state domain.CurriculumState
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, current_focus, difficulty_level, COALESCE(last_review_id, 0), updated_at
		FROM curriculum_state
		WHERE id = 1
	`).Scan(&state.ID, &state.CurrentFocus, &state.DifficultyLevel, &state.LastReviewID, &state.UpdatedAt)
	if err != nil {
		return domain.CurriculumState{}, err
	}
	return state, nil
}

func (s *Store) SaveExercise(ctx context.Context, ex domain.Exercise) (int64, error) {
	const query = `
		INSERT INTO exercises (title, brief, constraints_json, focus_skills_json, success_criteria_json, generation_kind)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	res, err := s.SQL.ExecContext(
		ctx,
		query,
		ex.Title,
		ex.Brief,
		mustJSON(ex.Constraints),
		mustJSON(ex.FocusSkills),
		mustJSON(ex.TGOCodes),
		mustJSON(ex.SuccessCriteria),
		ex.GenerationKind,
	)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (s *Store) SaveSubmission(ctx context.Context, sub domain.Submission) (int64, error) {
	if sub.DraftNumber == 0 {
		nextDraft, err := s.NextDraftNumber(ctx, sub.ExerciseID)
		if err != nil {
			return 0, err
		}
		sub.DraftNumber = nextDraft
	}
	res, err := s.SQL.ExecContext(ctx, `
		INSERT INTO submissions (exercise_id, parent_submission_id, draft_number, content, word_count)
		VALUES (?, ?, ?, ?, ?)
	`, sub.ExerciseID, nullableID(sub.ParentSubmissionID), sub.DraftNumber, sub.Content, sub.WordCount)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (s *Store) GetSubmission(ctx context.Context, submissionID int64) (domain.Submission, error) {
	var sub domain.Submission
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE id = ?
	`, submissionID).Scan(&sub.ID, &sub.ExerciseID, &sub.ParentSubmissionID, &sub.DraftNumber, &sub.Content, &sub.WordCount, &sub.CreatedAt)
	if err != nil {
		return domain.Submission{}, err
	}

	return sub, nil
}

func (s *Store) NextDraftNumber(ctx context.Context, exerciseID int64) (int, error) {
	var next int
	err := s.SQL.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(draft_number), 0) + 1
		FROM submissions
		WHERE exercise_id = ?
	`, exerciseID).Scan(&next)
	if err != nil {
		return 0, err
	}
	if next == 0 {
		next = 1
	}
	return next, nil
}

func (s *Store) LatestSubmissionForExercise(ctx context.Context, exerciseID int64) (domain.Submission, error) {
	var sub domain.Submission
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE exercise_id = ?
		ORDER BY draft_number DESC, id DESC
		LIMIT 1
	`, exerciseID).Scan(&sub.ID, &sub.ExerciseID, &sub.ParentSubmissionID, &sub.DraftNumber, &sub.Content, &sub.WordCount, &sub.CreatedAt)
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
		SELECT id, exercise_id, COALESCE(parent_submission_id, 0), COALESCE(draft_number, 1), content, word_count, created_at
		FROM submissions
		WHERE exercise_id = ? AND draft_number < ?
		ORDER BY draft_number DESC, id DESC
		LIMIT 1
	`, sub.ExerciseID, sub.DraftNumber).Scan(&prev.ID, &prev.ExerciseID, &prev.ParentSubmissionID, &prev.DraftNumber, &prev.Content, &prev.WordCount, &prev.CreatedAt)
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
			submission_id,
			review_kind,
			summary,
			strengths_json,
			weaknesses_json,
			analyzer_findings_json,
			next_focus,
			metric_word_count
		)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`,
		review.SubmissionID,
		review.ReviewKind,
		review.Summary,
		mustJSON(review.Strengths),
		mustJSON(review.Weaknesses),
		mustJSON(review.AnalyzerFindings),
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

func (s *Store) ActiveTGOs(ctx context.Context) ([]domain.TGO, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT c.id, c.code, c.title, c.description, c.stage, c.stage_order, a.slot
		FROM active_tgos a
		JOIN tgo_catalog c ON c.code = a.tgo_code
		ORDER BY a.slot ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tgos []domain.TGO
	for rows.Next() {
		var tgo domain.TGO
		if err := rows.Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder, &tgo.ActiveSlot); err != nil {
			return nil, err
		}
		if canonical, ok := domain.TGOByCode(tgo.Code); ok {
			tgo.Prerequisites = canonical.Prerequisites
			tgo.MasteryHint = canonical.MasteryHint
		}
		tgos = append(tgos, tgo)
	}
	return tgos, rows.Err()
}

func (s *Store) CompletedTGOs(ctx context.Context) ([]domain.TGO, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT c.id, c.code, c.title, c.description, c.stage, c.stage_order, 0
		FROM completed_tgos x
		JOIN tgo_catalog c ON c.code = x.tgo_code
		ORDER BY x.completed_at ASC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tgos []domain.TGO
	for rows.Next() {
		var tgo domain.TGO
		if err := rows.Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder, &tgo.ActiveSlot); err != nil {
			return nil, err
		}
		if canonical, ok := domain.TGOByCode(tgo.Code); ok {
			tgo.Prerequisites = canonical.Prerequisites
			tgo.MasteryHint = canonical.MasteryHint
		}
		tgos = append(tgos, tgo)
	}
	return tgos, rows.Err()
}

func (s *Store) RecentTGOStatuses(ctx context.Context, code string, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT status
		FROM review_tgo_assessments
		WHERE tgo_code = ?
		ORDER BY id DESC
		LIMIT ?
	`, code, limit)
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

func (s *Store) ReplaceActiveTGO(ctx context.Context, slot int, completedCode string, nextCode string) error {
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
		INSERT INTO completed_tgos (tgo_code)
		SELECT ?
		WHERE NOT EXISTS (SELECT 1 FROM completed_tgos WHERE tgo_code = ?)
	`, completedCode, completedCode); err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		UPDATE active_tgos SET tgo_code = ?, activated_at = CURRENT_TIMESTAMP WHERE slot = ?
	`, nextCode, slot); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) NextAvailableTGO(ctx context.Context) (domain.TGO, error) {
	var tgo domain.TGO
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, code, title, description, stage, stage_order
		FROM tgo_catalog
		WHERE code NOT IN (SELECT tgo_code FROM active_tgos)
		  AND code NOT IN (SELECT tgo_code FROM completed_tgos)
		ORDER BY stage_order ASC
		LIMIT 1
	`).Scan(&tgo.ID, &tgo.Code, &tgo.Title, &tgo.Description, &tgo.Stage, &tgo.StageOrder)
	if err != nil {
		return domain.TGO{}, err
	}
	return tgo, nil
}

func (s *Store) LatestReviewForSubmission(ctx context.Context, submissionID int64) (domain.Review, error) {
	var review domain.Review
	var strengthsJSON, weaknessesJSON, findingsJSON string
	err := s.SQL.QueryRowContext(ctx, `
		SELECT id, submission_id, review_kind, summary, strengths_json, weaknesses_json, analyzer_findings_json, next_focus, metric_word_count, created_at
		FROM reviews
		WHERE submission_id = ?
		ORDER BY id DESC
		LIMIT 1
	`, submissionID).Scan(
		&review.ID,
		&review.SubmissionID,
		&review.ReviewKind,
		&review.Summary,
		&strengthsJSON,
		&weaknessesJSON,
		&findingsJSON,
		&review.NextFocus,
		&review.MetricWordCount,
		&review.CreatedAt,
	)
	if err != nil {
		return domain.Review{}, err
	}
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
	return review, nil
}

func (s *Store) UpdateCurriculumState(ctx context.Context, focus string, difficulty int, reviewID int64) error {
	_, err := s.SQL.ExecContext(ctx, `
		UPDATE curriculum_state
		SET current_focus = ?, difficulty_level = ?, last_review_id = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = 1
	`, focus, difficulty, reviewID)
	return err
}

func (s *Store) SkillAverages(ctx context.Context, limit int) (map[string]float64, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		WITH recent_submissions AS (
			SELECT DISTINCT submission_id
			FROM submission_skill_scores
			ORDER BY submission_id DESC
			LIMIT ?
		)
		SELECT skill_name, AVG(score)
		FROM submission_skill_scores
		WHERE submission_id IN (SELECT submission_id FROM recent_submissions)
		GROUP BY skill_name
	`, limit)
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

func (s *Store) RecentSkillScores(ctx context.Context, skill string, limit int) ([]int, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT score
		FROM submission_skill_scores
		WHERE skill_name = ?
		ORDER BY submission_id DESC
		LIMIT ?
	`, skill, limit)
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

func (s *Store) ProgressReport(ctx context.Context, limit int) ([]string, error) {
	averages, err := s.SkillAverages(ctx, limit)
	if err != nil {
		return nil, err
	}

	var lines []string
	for _, skill := range domain.PrioritySkills {
		avg, ok := averages[skill]
		if !ok {
			continue
		}
		recent, err := s.RecentSkillScores(ctx, skill, 2)
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

func (s *Store) RecurringWeaknesses(ctx context.Context, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT weaknesses_json
		FROM reviews
		ORDER BY id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectRecurringJSONStrings(rows, 3)
}

func (s *Store) RecurringAnalyzerFindings(ctx context.Context, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT analyzer_findings_json
		FROM reviews
		ORDER BY id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectRecurringJSONStrings(rows, 4)
}

func (s *Store) StrongestWeakestSkills(ctx context.Context, limit int) ([]string, []string, error) {
	averages, err := s.SkillAverages(ctx, limit)
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
			return domain.SkillPriority(pairs[i].skill) > domain.SkillPriority(pairs[j].skill)
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

func (s *Store) History(ctx context.Context) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT e.id, e.title, s.id, r.id, COALESCE(r.next_focus, '')
		FROM exercises e
		LEFT JOIN submissions s ON s.exercise_id = e.id
		LEFT JOIN reviews r ON r.submission_id = s.id
		ORDER BY e.id DESC
		LIMIT 10
	`)
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

func (s *Store) RecentExerciseTitles(ctx context.Context, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT title
		FROM exercises
		ORDER BY id DESC
		LIMIT ?
	`, limit)
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
