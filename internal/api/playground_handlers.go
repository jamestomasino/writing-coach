package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/review"
)

func (s Server) handlePlaygroundReview(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	var payload struct {
		Content          string `json:"content"`
		WritingLanguage  string `json:"writing_language"`
		WritingType      string `json:"writing_type"`
		AssignmentFormat string `json:"assignment_format"`
		CoachingBrief    string `json:"coaching_brief"`
	}
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	content := strings.TrimSpace(payload.Content)
	if content == "" {
		writeError(w, http.StatusBadRequest, errMissing("content"))
		return
	}

	runtime, err := s.resolveLLMRuntime(r.Context(), appContext.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	result := s.Reviews.WithClient(runtime.Client, runtime.ProviderKind).ReviewSubmissionDetailedWithOptions(r.Context(), domain.Submission{
		UserID:    appContext.UserID,
		TreeID:    appContext.TreeID,
		Content:   content,
		WordCount: db.CountWords(content),
	}, nil, nil, review.Options{
		AnalyzerContext: analyzer.ContextOptions{
			WritingLanguage:  payload.WritingLanguage,
			WritingType:      payload.WritingType,
			AssignmentFormat: payload.AssignmentFormat,
		},
		CoachingBrief: strings.TrimSpace(payload.CoachingBrief),
		AllowUnscoped: true,
	})

	if result.Review.ReviewKind == runtime.ProviderKind {
		result.Review.ProviderNote = formatProviderNote(runtime.ProviderKind, runtime.ReviewModel)
	}
	if result.Review.ReviewKind == "deterministic-fallback" {
		s.logAIProviderEvent("generation_fallback", runtime.ProviderKind, appContext.UserID, map[string]any{
			"artifact": "playground_review",
			"reason":   strings.TrimSpace(result.Review.ProviderNote),
		})
	}
	result.Review.SkillScores = append([]domain.SkillScore(nil), result.Scores...)

	writeJSON(w, http.StatusOK, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"review":  toReviewResponse(result.Review),
	})
}

func errMissing(field string) error {
	return fmt.Errorf("%s is required", field)
}
