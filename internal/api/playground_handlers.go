package api

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/session"
)

type playgroundSessionPayload struct {
	Content          string `json:"content"`
	WritingLanguage  string `json:"writing_language"`
	WritingType      string `json:"writing_type"`
	AssignmentFormat string `json:"assignment_format"`
	CoachingBrief    string `json:"coaching_brief"`
}

func (s Server) handlePlaygroundSessionsList(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sessions, err := s.Store.ListPlaygroundSessions(r.Context(), appContext.UserID, appContext.TreeID, listLimit(r, 50, 200))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	items := make([]playgroundSessionResponse, 0, len(sessions))
	for _, session := range sessions {
		items = append(items, toPlaygroundSessionResponse(session))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context":  requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"sessions": items,
	})
}

func (s Server) handlePlaygroundSessionCreate(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	session, err := decodePlaygroundSessionPayload(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	session.UserID = appContext.UserID
	session.TreeID = appContext.TreeID
	sessionID, err := s.Store.SavePlaygroundSession(r.Context(), session)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	saved, err := s.Store.GetPlaygroundSession(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"session": toPlaygroundSessionResponse(saved),
	})
}

func (s Server) handlePlaygroundSessionGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sessionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || sessionID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid session id"))
		return
	}
	session, err := s.Store.GetPlaygroundSession(r.Context(), sessionID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(session.UserID, session.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("session not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"session": toPlaygroundSessionResponse(session),
	})
}

func (s Server) handlePlaygroundSessionUpdate(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sessionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || sessionID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid session id"))
		return
	}
	existing, err := s.Store.GetPlaygroundSession(r.Context(), sessionID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(existing.UserID, existing.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("session not found"))
		return
	}
	session, err := decodePlaygroundSessionPayload(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	session.ID = existing.ID
	session.UserID = existing.UserID
	session.TreeID = existing.TreeID
	if err := s.Store.UpdatePlaygroundSession(r.Context(), session); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	saved, err := s.Store.GetPlaygroundSession(r.Context(), existing.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"session": toPlaygroundSessionResponse(saved),
	})
}

func (s Server) handlePlaygroundSessionReviewCreate(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sessionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || sessionID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid session id"))
		return
	}
	session, err := s.Store.GetPlaygroundSession(r.Context(), sessionID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(session.UserID, session.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("session not found"))
		return
	}
	if strings.TrimSpace(session.Content) == "" {
		writeError(w, http.StatusBadRequest, errMissing("content"))
		return
	}
	job, err := s.enqueuePlaygroundReviewJob(r.Context(), appContext, session.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"job":     s.toAIJobResponse(r.Context(), job),
	})
}

func (s Server) handlePlaygroundSessionReviewsList(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	sessionID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || sessionID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid session id"))
		return
	}
	session, err := s.Store.GetPlaygroundSession(r.Context(), sessionID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(session.UserID, session.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("session not found"))
		return
	}
	reviews, err := s.Store.ListPlaygroundReviews(r.Context(), appContext.UserID, appContext.TreeID, sessionID, listLimit(r, 20, 100))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	items := make([]playgroundReviewResponse, 0, len(reviews))
	for _, item := range reviews {
		items = append(items, toPlaygroundReviewResponse(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"reviews": items,
	})
}

func (s Server) handlePlaygroundReviewGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	reviewID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || reviewID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid review id"))
		return
	}
	item, err := s.Store.GetPlaygroundReview(r.Context(), reviewID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(item.UserID, item.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("review not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"review":  toPlaygroundReviewResponse(item),
	})
}

func (s Server) handlePlaygroundReview(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	session, err := decodePlaygroundSessionPayload(w, r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	session.UserID = appContext.UserID
	session.TreeID = appContext.TreeID
	sessionID, err := s.Store.SavePlaygroundSession(r.Context(), session)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	job, err := s.enqueuePlaygroundReviewJob(r.Context(), appContext, sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	saved, err := s.Store.GetPlaygroundSession(r.Context(), sessionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusAccepted, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"job":     s.toAIJobResponse(r.Context(), job),
		"session": toPlaygroundSessionResponse(saved),
	})
}

func (s Server) enqueuePlaygroundReviewJob(ctx context.Context, appContext session.Context, sessionID int64) (domain.AIJob, error) {
	return s.Store.EnqueueAIJob(ctx, domain.AIJob{
		UserID:       appContext.UserID,
		TreeID:       appContext.TreeID,
		EnrollmentID: appContext.EnrollmentID,
		Kind:         aiJobKindPlaygroundReview,
		MaxAttempts:  3,
		PayloadJSON:  mustJSON(playgroundReviewJobPayload{SessionID: sessionID}),
	})
}

func decodePlaygroundSessionPayload(w http.ResponseWriter, r *http.Request) (domain.PlaygroundSession, error) {
	var payload playgroundSessionPayload
	if err := decodeJSONBody(w, r, &payload); err != nil {
		return domain.PlaygroundSession{}, err
	}
	content := strings.TrimSpace(payload.Content)
	if content == "" {
		return domain.PlaygroundSession{}, errMissing("content")
	}
	return domain.PlaygroundSession{
		Title:            playgroundSessionTitle(content, payload.AssignmentFormat, payload.WritingType),
		Content:          content,
		WritingLanguage:  strings.TrimSpace(payload.WritingLanguage),
		WritingType:      strings.TrimSpace(payload.WritingType),
		AssignmentFormat: strings.TrimSpace(payload.AssignmentFormat),
		CoachingBrief:    strings.TrimSpace(payload.CoachingBrief),
	}, nil
}

func playgroundSessionTitle(content, assignmentFormat, writingType string) string {
	if value := strings.TrimSpace(assignmentFormat); value != "" {
		return value
	}
	if value := strings.TrimSpace(writingType); value != "" {
		return value
	}
	content = strings.TrimSpace(content)
	if content == "" {
		return "Untitled"
	}
	line := strings.Split(content, "\n")[0]
	words := strings.Fields(line)
	if len(words) > 8 {
		words = words[:8]
	}
	title := strings.Join(words, " ")
	title = strings.TrimSpace(title)
	title = strings.Trim(title, " .,:;!?-")
	if title == "" {
		return "Untitled"
	}
	return title
}

func errMissing(field string) error {
	return fmt.Errorf("%s is required", field)
}
