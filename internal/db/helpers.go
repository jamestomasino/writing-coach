package db

import (
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

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
	var closedAt sql.NullTime
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
		&exercise.ProviderNote,
		&exercise.SourceSubmissionID,
		&closedAt,
		&exercise.CreatedAt,
	); err != nil {
		return domain.Exercise{}, err
	}
	if closedAt.Valid {
		exercise.ClosedAt = closedAt.Time
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
		&review.ProviderNote,
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

func nullTime(value time.Time) any {
	if value.IsZero() {
		return nil
	}
	return value
}

func boolToInt(value bool) int {
	if value {
		return 1
	}
	return 0
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
