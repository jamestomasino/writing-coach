package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
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

	job, err := s.Store.EnqueueAIJob(r.Context(), domain.AIJob{
		UserID:       appContext.UserID,
		TreeID:       appContext.TreeID,
		EnrollmentID: appContext.EnrollmentID,
		Kind:         aiJobKindPlaygroundReview,
		MaxAttempts:  3,
		PayloadJSON: mustJSON(playgroundReviewJobPayload{
			Content:          content,
			WritingLanguage:  payload.WritingLanguage,
			WritingType:      payload.WritingType,
			AssignmentFormat: payload.AssignmentFormat,
			CoachingBrief:    strings.TrimSpace(payload.CoachingBrief),
		}),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"job":     s.toAIJobResponse(r.Context(), job),
	})
}

func errMissing(field string) error {
	return fmt.Errorf("%s is required", field)
}
