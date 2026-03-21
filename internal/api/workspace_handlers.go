package api

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/prompt"
	"github.com/tomasino/writing-coach/internal/review"
	"github.com/tomasino/writing-coach/internal/session"
)

var errAssignmentClosed = errors.New("assignment is closed")

func (s Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	state, activeTGOs, completedTGOs, err := s.dashboardState(r.Context(), appContext)
	if err != nil {
		log.Printf("dashboard: load failed for user=%d tree=%d enrollment=%d: %v", appContext.UserID, appContext.TreeID, appContext.EnrollmentID, err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeDashboardPayload(r.Context(), w, appContext, state, activeTGOs, completedTGOs)
}

func (s Server) handleAssignmentsList(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	assignments, err := s.assignmentSummaries(r.Context(), appContext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context":     requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"assignments": assignments,
	})
}

func (s Server) handleAssignmentGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid assignment id"))
		return
	}
	assignment, err := s.assignmentTimeline(r.Context(), appContext, id)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context":    requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"assignment": assignment,
	})
}

func (s Server) handleAssignmentClose(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid assignment id"))
		return
	}
	rootID, _, err := s.assignmentRootID(r.Context(), appContext, id)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if err := s.Store.CloseExercise(r.Context(), appContext.UserID, appContext.TreeID, rootID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true})
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s Server) handleExercisesList(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	exercises, err := s.Store.ListExercises(r.Context(), appContext.UserID, appContext.TreeID, listLimit(r, 20, 100))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	items := make([]exerciseResponse, 0, len(exercises))
	for _, exercise := range exercises {
		items = append(items, toExerciseResponse(exercise))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context":   requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"exercises": items,
	})
}

func (s Server) handleExerciseGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid exercise id"))
		return
	}
	exercise, err := s.Store.GetExercise(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(exercise.UserID, exercise.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("exercise not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context":  requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"exercise": toExerciseResponse(exercise),
	})
}

func (s Server) handlePromptNext(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		log.Printf("prompt next: resolve session failed: %v", err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if r.ContentLength != 0 {
		var payload struct {
			TGOCodes []string `json:"tgo_codes"`
		}
		if err := decodeJSONBody(w, r, &payload); err != nil {
			writeError(w, http.StatusBadRequest, err)
			return
		}
		if len(payload.TGOCodes) > 0 {
			if err := s.setActiveTGOsForSelection(r.Context(), appContext, payload.TGOCodes); err != nil {
				log.Printf("prompt next: set active skills failed for user=%d tree=%d enrollment=%d codes=%v: %v", appContext.UserID, appContext.TreeID, appContext.EnrollmentID, payload.TGOCodes, err)
				writeError(w, http.StatusBadRequest, err)
				return
			}
		}
	}
	ex, err := s.generateNextExercise(r.Context(), appContext)
	if err != nil {
		log.Printf("prompt next: create exercise failed for user=%d tree=%d enrollment=%d: %v", appContext.UserID, appContext.TreeID, appContext.EnrollmentID, err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context":  requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"exercise": toExerciseResponse(ex),
	})
}

func (s Server) handlePromptAccept(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var payload struct {
		Title           string   `json:"title"`
		Brief           string   `json:"brief"`
		Constraints     []string `json:"constraints"`
		FocusSkills     []string `json:"focus_skills"`
		TGOCodes        []string `json:"tgo_codes"`
		SuccessCriteria []string `json:"success_criteria"`
		GenerationKind  string   `json:"generation_kind"`
		ProviderNote    string   `json:"provider_note"`
	}
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	ex := domain.Exercise{
		UserID:          appContext.UserID,
		TreeID:          appContext.TreeID,
		Title:           strings.TrimSpace(payload.Title),
		Brief:           strings.TrimSpace(payload.Brief),
		Constraints:     sanitizeStringList(payload.Constraints),
		FocusSkills:     sanitizeStringList(payload.FocusSkills),
		TGOCodes:        sanitizeStringList(payload.TGOCodes),
		SuccessCriteria: sanitizeStringList(payload.SuccessCriteria),
		GenerationKind:  strings.TrimSpace(payload.GenerationKind),
		ProviderNote:    strings.TrimSpace(payload.ProviderNote),
	}
	if ex.Title == "" || ex.Brief == "" || len(ex.Constraints) == 0 || len(ex.SuccessCriteria) == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("preview exercise is incomplete"))
		return
	}
	if ex.GenerationKind == "" {
		ex.GenerationKind = "accepted-preview"
	}
	id, err := s.saveExercise(r.Context(), ex)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	ex.ID = id
	writeJSON(w, http.StatusOK, map[string]any{
		"context":  requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"exercise": toExerciseResponse(ex),
	})
}

func (s Server) handlePromptRevise(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var payload struct {
		SubmissionID int64 `json:"submission_id"`
	}
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if payload.SubmissionID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("submission_id is required"))
		return
	}
	ex, err := s.createRevisionExercise(r.Context(), appContext, payload.SubmissionID)
	if err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, errAssignmentClosed) {
			status = http.StatusConflict
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context":  requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"exercise": toExerciseResponse(ex),
	})
}

func (s Server) handleSubmissionCreate(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var payload struct {
		ExerciseID         int64  `json:"exercise_id"`
		ParentSubmissionID int64  `json:"parent_submission_id"`
		Content            string `json:"content"`
	}
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if payload.ExerciseID == 0 || strings.TrimSpace(payload.Content) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("exercise_id and content are required"))
		return
	}
	exercise, err := s.Store.GetExercise(r.Context(), payload.ExerciseID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(exercise.UserID, exercise.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("exercise not found"))
		return
	}
	if closed, err := s.isExerciseChainClosed(r.Context(), appContext, exercise); err == nil && closed {
		writeError(w, http.StatusConflict, fmt.Errorf("assignment is closed"))
		return
	}
	draftNumber := 1
	if payload.ParentSubmissionID != 0 {
		parent, err := s.Store.GetSubmission(r.Context(), payload.ParentSubmissionID)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid parent submission"))
			return
		}
		if !belongsToContext(parent.UserID, parent.TreeID, appContext) || parent.ExerciseID != exercise.ID {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid parent submission"))
			return
		}
		draftNumber = parent.DraftNumber + 1
	}
	sub := domain.Submission{
		UserID:             appContext.UserID,
		TreeID:             appContext.TreeID,
		ExerciseID:         exercise.ID,
		ParentSubmissionID: payload.ParentSubmissionID,
		DraftNumber:        draftNumber,
		Content:            payload.Content,
		WordCount:          db.CountWords(payload.Content),
	}
	id, err := s.Store.SaveSubmission(r.Context(), sub)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	saved, err := s.Store.GetSubmission(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"context":    requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"submission": toSubmissionResponse(saved),
	})
}

func (s Server) handleSubmissionsList(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	exerciseID, err := parseOptionalInt64(r.URL.Query().Get("exercise_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid exercise_id"))
		return
	}
	if exerciseID != 0 {
		exercise, err := s.Store.GetExercise(r.Context(), exerciseID)
		if err != nil {
			status := http.StatusInternalServerError
			if db.IsNotFound(err) {
				status = http.StatusNotFound
			}
			log.Printf("submissions list: exercise lookup failed for user=%d tree=%d exercise=%d: %v", appContext.UserID, appContext.TreeID, exerciseID, err)
			writeError(w, status, err)
			return
		}
		if !belongsToContext(exercise.UserID, exercise.TreeID, appContext) {
			writeError(w, http.StatusNotFound, fmt.Errorf("exercise not found"))
			return
		}
	}
	submissions, err := s.Store.ListSubmissions(r.Context(), appContext.UserID, appContext.TreeID, exerciseID, listLimit(r, 20, 100))
	if err != nil {
		log.Printf("submissions list: load failed for user=%d tree=%d exercise=%d: %v", appContext.UserID, appContext.TreeID, exerciseID, err)
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	items := make([]submissionResponse, 0, len(submissions))
	for _, submission := range submissions {
		items = append(items, toSubmissionResponse(submission))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context":     requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"submissions": items,
	})
}

func (s Server) dashboardState(ctx context.Context, appContext session.Context) (domain.CurriculumState, []domain.TGO, []domain.TGO, error) {
	state, err := s.Store.GetCurriculumState(ctx, appContext.EnrollmentID)
	if errors.Is(err, sql.ErrNoRows) {
		if err := s.reseedEnrollmentState(ctx, appContext); err != nil {
			return domain.CurriculumState{}, nil, nil, err
		}
		state, err = s.Store.GetCurriculumState(ctx, appContext.EnrollmentID)
	}
	if err != nil {
		return domain.CurriculumState{}, nil, nil, err
	}
	activeTGOs, err := s.Store.ActiveTGOs(ctx, appContext.EnrollmentID)
	if err != nil {
		return domain.CurriculumState{}, nil, nil, err
	}
	completedTGOs, err := s.Store.CompletedTGOs(ctx, appContext.EnrollmentID)
	if err != nil {
		return domain.CurriculumState{}, nil, nil, err
	}
	return state, activeTGOs, completedTGOs, nil
}

func (s Server) reseedEnrollmentState(ctx context.Context, appContext session.Context) error {
	user, err := s.Store.UserBySlug(ctx, appContext.UserSlug)
	if err != nil {
		return err
	}
	_, _, _, err = s.Store.EnsureDefaultUserTree(ctx, appContext.UserSlug, user.Name, appContext.TreeSlug)
	return err
}

func (s Server) handleSubmissionGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid submission id"))
		return
	}
	sub, err := s.Store.GetSubmission(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(sub.UserID, sub.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("submission not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context":    requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"submission": toSubmissionResponse(sub),
	})
}

func (s Server) handleReviewCreate(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var payload struct {
		SubmissionID int64 `json:"submission_id"`
	}
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if payload.SubmissionID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("submission_id is required"))
		return
	}
	sub, err := s.Store.GetSubmission(r.Context(), payload.SubmissionID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(sub.UserID, sub.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("submission not found"))
		return
	}
	if existing, err := s.Store.LatestReviewForSubmission(r.Context(), sub.ID); err == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
			"review":  toReviewResponse(existing),
			"job": reviewJobResponse{
				SubmissionID: sub.ID,
				ReviewID:     existing.ID,
				Status:       "completed",
			},
		})
		return
	}
	job, err := s.Store.EnqueueReviewJob(r.Context(), domain.ReviewJob{
		UserID:       appContext.UserID,
		TreeID:       appContext.TreeID,
		EnrollmentID: appContext.EnrollmentID,
		SubmissionID: sub.ID,
		MaxAttempts:  3,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	log.Printf("review job queued: job=%d submission=%d user=%d tree=%d", job.ID, job.SubmissionID, job.UserID, job.TreeID)
	writeJSON(w, http.StatusAccepted, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"job":     toReviewJobResponse(job),
	})
}

func (s Server) handleReviewJobGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	submissionID, err := parseOptionalInt64(r.URL.Query().Get("submission_id"))
	if err != nil || submissionID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("submission_id is required"))
		return
	}
	sub, err := s.Store.GetSubmission(r.Context(), submissionID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(sub.UserID, sub.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("submission not found"))
		return
	}
	job, err := s.Store.ReviewJobBySubmission(r.Context(), appContext.UserID, appContext.TreeID, submissionID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"job":     toReviewJobResponse(job),
	})
}

func (s Server) handleReviewsList(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	submissionID, err := parseOptionalInt64(r.URL.Query().Get("submission_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid submission_id"))
		return
	}
	if submissionID != 0 {
		sub, err := s.Store.GetSubmission(r.Context(), submissionID)
		if err != nil {
			status := http.StatusInternalServerError
			if db.IsNotFound(err) {
				status = http.StatusNotFound
			}
			writeError(w, status, err)
			return
		}
		if !belongsToContext(sub.UserID, sub.TreeID, appContext) {
			writeError(w, http.StatusNotFound, fmt.Errorf("submission not found"))
			return
		}
	}
	reviews, err := s.Store.ListReviews(r.Context(), appContext.UserID, appContext.TreeID, submissionID, listLimit(r, 20, 100))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	items := make([]reviewResponse, 0, len(reviews))
	for _, review := range reviews {
		response := toReviewResponse(review)
		if artifacts, err := s.Store.GetReviewArtifacts(r.Context(), review.ID); err == nil {
			response.Artifacts = decodeReviewArtifacts(artifacts)
			if len(response.Annotations) == 0 {
				response.Annotations = append(response.Annotations, response.Artifacts.Annotations...)
			}
		}
		items = append(items, response)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"reviews": items,
	})
}

func (s Server) handleReviewGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid review id"))
		return
	}
	reviewResult, err := s.Store.GetReview(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(reviewResult.UserID, reviewResult.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("review not found"))
		return
	}
	artifacts, err := s.Store.GetReviewArtifacts(r.Context(), reviewResult.ID)
	if err == nil {
		response := toReviewResponse(reviewResult)
		response.Artifacts = decodeReviewArtifacts(artifacts)
		if len(response.Annotations) == 0 {
			response.Annotations = append(response.Annotations, response.Artifacts.Annotations...)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
			"review":  response,
		})
		return
	} else if !db.IsNotFound(err) {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"review":  toReviewResponse(reviewResult),
	})
}

func (s Server) handleCompare(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	submissionID, err := strconv.ParseInt(r.URL.Query().Get("submission_id"), 10, 64)
	if err != nil || submissionID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("submission_id is required"))
		return
	}
	current, err := s.Store.GetSubmission(r.Context(), submissionID)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	if !belongsToContext(current.UserID, current.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("submission not found"))
		return
	}
	var baseline domain.Submission
	if raw := r.URL.Query().Get("against"); raw != "" {
		againstID, err := strconv.ParseInt(raw, 10, 64)
		if err != nil || againstID == 0 {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid against submission id"))
			return
		}
		baseline, err = s.Store.GetSubmission(r.Context(), againstID)
		if err != nil {
			writeError(w, http.StatusNotFound, err)
			return
		}
		if !belongsToContext(baseline.UserID, baseline.TreeID, appContext) {
			writeError(w, http.StatusNotFound, fmt.Errorf("baseline submission not found"))
			return
		}
	} else {
		baseline, err = s.Store.PreviousSubmission(r.Context(), current)
		if err != nil {
			writeError(w, http.StatusNotFound, fmt.Errorf("no baseline submission available"))
			return
		}
	}
	currentReview, err := s.Store.LatestReviewForSubmission(r.Context(), current.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Errorf("current submission has no review yet"))
		return
	}
	baselineReview, err := s.Store.LatestReviewForSubmission(r.Context(), baseline.ID)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Errorf("baseline submission has no review yet"))
		return
	}
	comparison := review.CompareSubmissions(current, baseline, currentReview, baselineReview)
	writeJSON(w, http.StatusOK, map[string]any{
		"current_submission_id":  current.ID,
		"baseline_submission_id": baseline.ID,
		"comparison": comparisonResponse{
			Summary:              comparison.Summary,
			WordDelta:            comparison.WordDelta,
			AddedWords:           comparison.AddedWords,
			RemovedWords:         comparison.RemovedWords,
			AddressedWeaknesses:  comparison.AddressedWeaknesses,
			PersistingWeaknesses: comparison.PersistingWeaknesses,
		},
	})
}

func (s Server) generateNextExercise(ctx context.Context, appContext session.Context) (domain.Exercise, error) {
	profile := s.onboardingProfile(ctx, appContext.EnrollmentID)
	coachingBrief := s.coachingBrief(ctx, appContext.EnrollmentID)
	state, err := s.Store.GetCurriculumState(ctx, appContext.EnrollmentID)
	if err != nil {
		log.Printf("create exercise: curriculum state lookup failed for enrollment=%d: %v", appContext.EnrollmentID, err)
		return domain.Exercise{}, err
	}
	recentTitles, err := s.Store.RecentExerciseTitles(ctx, appContext.UserID, appContext.TreeID, 3)
	if err != nil {
		log.Printf("create exercise: recent titles lookup failed for user=%d tree=%d: %v", appContext.UserID, appContext.TreeID, err)
		return domain.Exercise{}, err
	}
	recentWeaknesses, err := s.Store.RecurringWeaknesses(ctx, appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		log.Printf("create exercise: recurring weaknesses lookup failed for user=%d tree=%d: %v", appContext.UserID, appContext.TreeID, err)
		return domain.Exercise{}, err
	}
	recurringFindings, err := s.Store.RecurringAnalyzerFindings(ctx, appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		log.Printf("create exercise: recurring findings lookup failed for user=%d tree=%d: %v", appContext.UserID, appContext.TreeID, err)
		return domain.Exercise{}, err
	}
	activeTGOs, err := s.Store.ActiveTGOs(ctx, appContext.EnrollmentID)
	if err != nil {
		log.Printf("create exercise: active skills lookup failed for enrollment=%d: %v", appContext.EnrollmentID, err)
		return domain.Exercise{}, err
	}
	runtime, err := s.resolveLLMRuntime(ctx, appContext.UserID)
	if err != nil {
		log.Printf("create exercise: provider resolution failed for user=%d: %v", appContext.UserID, err)
		return domain.Exercise{}, err
	}

	ex := s.Prompts.WithClient(runtime.Client, runtime.ProviderKind).NextExercise(ctx, prompt.Context{
		CurriculumState:   state,
		ActiveTGOs:        activeTGOs,
		OnboardingProfile: profile,
		RecentTitles:      recentTitles,
		RecentWeaknesses:  recentWeaknesses,
		RecurringFindings: recurringFindings,
		CoachingBrief:     coachingBrief,
	})
	if ex.GenerationKind == runtime.ProviderKind {
		ex.ProviderNote = formatProviderNote(runtime.ProviderKind, runtime.PromptModel)
	}
	if ex.GenerationKind == "deterministic-fallback" {
		s.logAIProviderEvent("generation_fallback", runtime.ProviderKind, appContext.UserID, map[string]any{
			"artifact": "exercise",
			"reason":   strings.TrimSpace(ex.ProviderNote),
		})
	}
	if len(activeTGOs) > 0 {
		ex.FocusSkills = focusSkillsForTGOs(activeTGOs, ex.FocusSkills)
		ex.TGOCodes = tgoCodesForExercise(activeTGOs)
		ex.SuccessCriteria = successCriteriaForTGOs(activeTGOs, ex.SuccessCriteria)
	}
	return ex, nil
}

func (s Server) saveExercise(ctx context.Context, ex domain.Exercise) (int64, error) {
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	id, err := s.Store.SaveExercise(saveCtx, ex)
	if err != nil {
		log.Printf("create exercise: save failed for user=%d tree=%d title=%q generation=%q: %v", ex.UserID, ex.TreeID, ex.Title, ex.GenerationKind, err)
		return 0, err
	}
	return id, nil
}

func (s Server) createNextExercise(ctx context.Context, appContext session.Context) (domain.Exercise, error) {
	ex, err := s.generateNextExercise(ctx, appContext)
	if err != nil {
		return domain.Exercise{}, err
	}
	ex.UserID = appContext.UserID
	ex.TreeID = appContext.TreeID
	id, err := s.saveExercise(ctx, ex)
	if err != nil {
		return domain.Exercise{}, err
	}
	ex.ID = id
	return ex, nil
}

func (s Server) setActiveTGOsForSelection(ctx context.Context, appContext session.Context, codes []string) error {
	if len(codes) != 3 {
		return fmt.Errorf("exactly 3 TGOs must be selected")
	}
	seen := make(map[string]bool, len(codes))
	selected := make([]string, 0, len(codes))
	for _, code := range codes {
		code = strings.TrimSpace(code)
		if code == "" {
			return fmt.Errorf("TGO codes cannot be empty")
		}
		if seen[code] {
			return fmt.Errorf("duplicate TGO code: %s", code)
		}
		seen[code] = true
		selected = append(selected, code)
	}

	treeDef, err := s.Store.TreeDefinitionBySlug(ctx, appContext.TreeSlug)
	if err != nil {
		return err
	}
	activeTGOs, err := s.Store.ActiveTGOs(ctx, appContext.EnrollmentID)
	if err != nil {
		return err
	}
	completedTGOs, err := s.Store.CompletedTGOs(ctx, appContext.EnrollmentID)
	if err != nil {
		return err
	}

	completedSet := make(map[string]bool, len(completedTGOs))
	selectable := make(map[string]bool, len(treeDef.TGOs))
	for _, tgo := range completedTGOs {
		completedSet[tgo.Code] = true
	}
	for _, tgo := range activeTGOs {
		selectable[tgo.Code] = true
	}
	for _, tgo := range treeDef.TGOs {
		if completedSet[tgo.Code] {
			continue
		}
		if selectable[tgo.Code] || prereqsMetForSelection(tgo, completedSet) {
			selectable[tgo.Code] = true
		}
	}
	for _, code := range selected {
		if !selectable[code] {
			return fmt.Errorf("TGO %q is not unlocked for selection", code)
		}
	}
	return s.Store.SetActiveTGOs(ctx, appContext.EnrollmentID, selected)
}

func (s Server) createRevisionExercise(ctx context.Context, appContext session.Context, submissionID int64) (domain.Exercise, error) {
	coachingBrief := s.coachingBrief(ctx, appContext.EnrollmentID)
	var profile *domain.OnboardingProfile
	if loadedProfile, err := s.Store.OnboardingProfileByEnrollmentID(ctx, appContext.EnrollmentID); err == nil {
		profile = &loadedProfile
	}
	sub, err := s.Store.GetSubmission(ctx, submissionID)
	if err != nil {
		return domain.Exercise{}, err
	}
	existingExercise, err := s.Store.GetExercise(ctx, sub.ExerciseID)
	if err != nil {
		return domain.Exercise{}, err
	}
	if closed, err := s.isExerciseChainClosed(ctx, appContext, existingExercise); err == nil && closed {
		return domain.Exercise{}, errAssignmentClosed
	}
	reviewResult, err := s.Store.LatestReviewForSubmission(ctx, sub.ID)
	if err != nil {
		return domain.Exercise{}, fmt.Errorf("submission %d has no review yet", sub.ID)
	}
	state, err := s.Store.GetCurriculumState(ctx, appContext.EnrollmentID)
	if err != nil {
		return domain.Exercise{}, err
	}
	recentTitles, err := s.Store.RecentExerciseTitles(ctx, appContext.UserID, appContext.TreeID, 3)
	if err != nil {
		return domain.Exercise{}, err
	}
	recentWeaknesses, err := s.Store.RecurringWeaknesses(ctx, appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		return domain.Exercise{}, err
	}
	recurringFindings, err := s.Store.RecurringAnalyzerFindings(ctx, appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		return domain.Exercise{}, err
	}
	activeTGOs, err := s.Store.ActiveTGOs(ctx, appContext.EnrollmentID)
	if err != nil {
		return domain.Exercise{}, err
	}
	runtime, err := s.resolveLLMRuntime(ctx, appContext.UserID)
	if err != nil {
		return domain.Exercise{}, err
	}

	var cmp *review.Comparison
	if previous, err := s.Store.PreviousSubmission(ctx, sub); err == nil {
		if previousReview, err := s.Store.LatestReviewForSubmission(ctx, previous.ID); err == nil {
			comparison := review.CompareSubmissions(sub, previous, reviewResult, previousReview)
			cmp = &comparison
		}
	}

	ex := s.Prompts.WithClient(runtime.Client, runtime.ProviderKind).RevisionExercise(ctx, prompt.Context{
		CurriculumState:    state,
		ActiveTGOs:         activeTGOs,
		OnboardingProfile:  profile,
		RecentTitles:       recentTitles,
		RecentWeaknesses:   recentWeaknesses,
		RecurringFindings:  recurringFindings,
		CoachingBrief:      coachingBrief,
		RevisionOf:         &sub,
		RevisionReview:     &reviewResult,
		RevisionComparison: cmp,
	})
	if ex.GenerationKind == runtime.ProviderKind {
		ex.ProviderNote = formatProviderNote(runtime.ProviderKind, runtime.PromptModel)
	}
	if ex.GenerationKind == "deterministic-fallback" {
		s.logAIProviderEvent("generation_fallback", runtime.ProviderKind, appContext.UserID, map[string]any{
			"artifact": "revision_exercise",
			"reason":   strings.TrimSpace(ex.ProviderNote),
		})
	}
	ex.UserID = appContext.UserID
	ex.TreeID = appContext.TreeID
	saveCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
	defer cancel()
	id, err := s.Store.SaveExercise(saveCtx, ex)
	if err != nil {
		return domain.Exercise{}, err
	}
	ex.ID = id
	return ex, nil
}

func (s Server) assignmentRootID(ctx context.Context, appContext session.Context, exerciseID int64) (int64, map[int64]domain.Exercise, error) {
	current, err := s.Store.GetExercise(ctx, exerciseID)
	if err != nil {
		return 0, nil, err
	}
	if !belongsToContext(current.UserID, current.TreeID, appContext) {
		return 0, nil, sql.ErrNoRows
	}
	exercises, err := s.Store.ListExercises(ctx, appContext.UserID, appContext.TreeID, 500)
	if err != nil {
		return 0, nil, err
	}
	submissions, err := s.Store.ListSubmissions(ctx, appContext.UserID, appContext.TreeID, 0, 500)
	if err != nil {
		return 0, nil, err
	}
	exerciseByID := make(map[int64]domain.Exercise, len(exercises))
	for _, exercise := range exercises {
		exerciseByID[exercise.ID] = exercise
	}
	submissionByID := make(map[int64]domain.Submission, len(submissions))
	for _, submission := range submissions {
		submissionByID[submission.ID] = submission
	}
	return rootExerciseID(current, exerciseByID, submissionByID), exerciseByID, nil
}

func (s Server) isExerciseChainClosed(ctx context.Context, appContext session.Context, exercise domain.Exercise) (bool, error) {
	rootID, exerciseByID, err := s.assignmentRootID(ctx, appContext, exercise.ID)
	if err != nil {
		return false, err
	}
	return assignmentClosed(rootID, exerciseByID), nil
}

func (s Server) writeEnrollmentBoard(ctx context.Context, w http.ResponseWriter, appContext session.Context) {
	state, err := s.Store.GetCurriculumState(ctx, appContext.EnrollmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	activeTGOs, err := s.Store.ActiveTGOs(ctx, appContext.EnrollmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	completedTGOs, err := s.Store.CompletedTGOs(ctx, appContext.EnrollmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeDashboardPayload(ctx, w, appContext, state, activeTGOs, completedTGOs)
}

func (s Server) writeDashboardPayload(ctx context.Context, w http.ResponseWriter, appContext session.Context, state domain.CurriculumState, activeTGOs, completedTGOs []domain.TGO) {
	treeDef, err := s.Store.TreeDefinitionBySlug(ctx, appContext.TreeSlug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	progress, err := s.Store.ProgressReport(ctx, appContext.UserID, appContext.TreeID, treeDef.PrioritySkills, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	strongest, weakest, err := s.Store.StrongestWeakestSkills(ctx, appContext.UserID, appContext.TreeID, treeDef.PrioritySkills, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	exercises, err := s.Store.ListExercises(ctx, appContext.UserID, appContext.TreeID, 500)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	submissions, err := s.Store.ListSubmissions(ctx, appContext.UserID, appContext.TreeID, 0, 500)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	reviews, err := s.Store.ListReviews(ctx, appContext.UserID, appContext.TreeID, 0, 500)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	totalAssignments, draftCount, completedAssignments := dashboardAssignmentStats(exercises, submissions, reviews)
	recurringWeaknesses, err := s.Store.RecurringWeaknesses(ctx, appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	recurringFindings, err := s.Store.RecurringAnalyzerFindings(ctx, appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	recurringSlips, err := s.Store.RecurringCompletedTGOSlips(ctx, appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	recentExercises, err := s.Store.HistoryItems(ctx, appContext.UserID, appContext.TreeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	completedSet := map[string]bool{}
	for _, tgo := range completedTGOs {
		completedSet[tgo.Code] = true
	}
	activeSet := map[string]bool{}
	for _, tgo := range activeTGOs {
		activeSet[tgo.Code] = true
	}
	upcoming := domain.NextUnlockedFromDefinition(treeDef, completedSet, activeSet, 3)
	for i := range activeTGOs {
		signal, err := s.Store.TGOMasterySignal(ctx, appContext.EnrollmentID, activeTGOs[i], "")
		if err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
		activeTGOs[i].MasteryStage = signal.Stage
		activeTGOs[i].MasteryPercent = signal.Percent
		activeTGOs[i].EvidenceCount = signal.EvidenceCount
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"context":                   requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"curriculum_state":          toCurriculumStateResponse(state),
		"active_tgos":               toTGOResponses(activeTGOs),
		"completed_tgos":            toTGOResponses(completedTGOs),
		"upcoming_tgos":             toTGOResponses(upcoming),
		"total_assignments":         totalAssignments,
		"completed_assignments":     completedAssignments,
		"draft_count":               draftCount,
		"revision_count":            0,
		"progress_lines":            progress,
		"strongest_skills":          strongest,
		"weakest_skills":            weakest,
		"recurring_weaknesses":      recurringWeaknesses,
		"recurring_findings":        recurringFindings,
		"recurring_completed_slips": recurringSlips,
		"history":                   toHistoryResponses(recentExercises),
	})
}

func dashboardAssignmentStats(exercises []domain.Exercise, submissions []domain.Submission, reviews []domain.Review) (totalAssignments, draftCount, completedAssignments int) {
	exerciseByID := make(map[int64]domain.Exercise, len(exercises))
	for _, exercise := range exercises {
		exerciseByID[exercise.ID] = exercise
	}
	submissionByID := make(map[int64]domain.Submission, len(submissions))
	for _, submission := range submissions {
		submissionByID[submission.ID] = submission
		draftCount++
	}
	totalRoots := make(map[int64]bool)
	for _, exercise := range exercises {
		rootID := rootExerciseID(exercise, exerciseByID, submissionByID)
		if rootID != 0 {
			totalRoots[rootID] = true
		}
	}

	completedRoots := make(map[int64]bool)
	for _, reviewResult := range reviews {
		submission, ok := submissionByID[reviewResult.SubmissionID]
		if !ok {
			continue
		}
		exercise, ok := exerciseByID[submission.ExerciseID]
		if !ok {
			continue
		}
		rootID := rootExerciseID(exercise, exerciseByID, submissionByID)
		if rootID != 0 {
			completedRoots[rootID] = true
		}
	}
	return len(totalRoots), draftCount, len(completedRoots)
}
