package db

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

func (s *Store) SkillAverages(ctx context.Context, userID, treeID int64, limit int) (map[string]float64, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		WITH recent_submissions AS (
			SELECT DISTINCT sss.submission_id
			FROM submission_skill_scores sss
			JOIN submissions s ON s.id = sss.submission_id
			WHERE s.user_id = ? AND s.tree_id = ? AND sss.score_source = 'deterministic'
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
		WHERE sss.skill_name = ? AND s.user_id = ? AND s.tree_id = ? AND sss.score_source = 'deterministic'
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
		WHERE submission_id = ? AND score_source = 'deterministic'
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
		SELECT title, tgo_codes_json
		FROM exercises
		WHERE user_id = ? AND tree_id = ?
		ORDER BY id DESC
		LIMIT 10
	`, userID, treeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []string
	for rows.Next() {
		var (
			title        string
			tgoCodesJSON string
		)

		if err := rows.Scan(&title, &tgoCodesJSON); err != nil {
			return nil, err
		}

		title = strings.TrimSpace(title)
		if title == "" {
			title = "Untitled assignment"
		}

		item := "Assignment: " + title
		if tgoCodes, err := DecodeStringSlice(tgoCodesJSON); err == nil && len(tgoCodes) > 0 {
			item += " | " + strings.Join(tgoCodes, ", ")
		}
		items = append(items, item)
	}

	return items, rows.Err()
}

func (s *Store) HistoryItems(ctx context.Context, userID, treeID int64) ([]domain.Exercise, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT id, user_id, tree_id, title, brief, constraints_json, focus_skills_json, tgo_codes_json, success_criteria_json, generation_kind, provider_note, COALESCE(source_submission_id, 0), closed_at, created_at
		FROM exercises
		WHERE user_id = ? AND tree_id = ?
		ORDER BY id DESC
		LIMIT 10
	`, userID, treeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.Exercise
	for rows.Next() {
		exercise, err := scanExercise(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, exercise)
	}

	return items, rows.Err()
}

func (s *Store) CompletedAssignmentCount(ctx context.Context, userID, treeID int64) (int, error) {
	var count int
	err := s.SQL.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT e.id)
		FROM exercises e
		WHERE e.user_id = ? AND e.tree_id = ?
		  AND EXISTS (
			SELECT 1
			FROM submissions s
			JOIN reviews r ON r.submission_id = s.id AND r.user_id = s.user_id AND r.tree_id = s.tree_id
			WHERE s.exercise_id = e.id AND s.user_id = e.user_id AND s.tree_id = e.tree_id
		  )
	`, userID, treeID).Scan(&count)
	return count, err
}

func (s *Store) ResetUserData(ctx context.Context, userID int64) error {
	tx, err := s.SQL.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	statements := []string{
		`DELETE FROM review_artifacts WHERE review_id IN (SELECT id FROM reviews WHERE user_id = ?)`,
		`DELETE FROM review_tgo_assessments WHERE review_id IN (SELECT id FROM reviews WHERE user_id = ?)`,
		`DELETE FROM submission_skill_scores WHERE submission_id IN (SELECT id FROM submissions WHERE user_id = ?)`,
		`DELETE FROM reviews WHERE user_id = ?`,
		`DELETE FROM submissions WHERE user_id = ?`,
		`DELETE FROM exercises WHERE user_id = ?`,
		`DELETE FROM enrollment_completed_tgos WHERE enrollment_id IN (SELECT id FROM user_tree_enrollments WHERE user_id = ?)`,
		`DELETE FROM enrollment_active_tgos WHERE enrollment_id IN (SELECT id FROM user_tree_enrollments WHERE user_id = ?)`,
		`DELETE FROM enrollment_onboarding_profiles WHERE enrollment_id IN (SELECT id FROM user_tree_enrollments WHERE user_id = ?)`,
		`DELETE FROM user_curriculum_state WHERE enrollment_id IN (SELECT id FROM user_tree_enrollments WHERE user_id = ?)`,
		`DELETE FROM user_tree_enrollments WHERE user_id = ?`,
		`DELETE FROM user_onboarding_profiles WHERE user_id = ?`,
		`UPDATE users SET active_tree_slug = '' WHERE id = ?`,
	}
	for _, statement := range statements {
		if _, err = tx.ExecContext(ctx, statement, userID); err != nil {
			return err
		}
	}

	return tx.Commit()
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

func (s *Store) RecentExerciseSummaries(ctx context.Context, userID, treeID int64, limit int) ([]string, error) {
	rows, err := s.SQL.QueryContext(ctx, `
		SELECT title, brief
		FROM exercises
		WHERE user_id = ? AND tree_id = ?
		ORDER BY id DESC
		LIMIT ?
	`, userID, treeID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	summaries := make([]string, 0, limit)
	for rows.Next() {
		var title, brief string
		if err := rows.Scan(&title, &brief); err != nil {
			return nil, err
		}
		title = strings.TrimSpace(title)
		brief = strings.TrimSpace(brief)
		if title == "" && brief == "" {
			continue
		}
		if title == "" {
			summaries = append(summaries, brief)
			continue
		}
		if brief == "" {
			summaries = append(summaries, title)
			continue
		}
		summaries = append(summaries, title+": "+brief)
	}
	return summaries, rows.Err()
}
