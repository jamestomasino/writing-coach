package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/curriculum"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/prompt"
	"github.com/tomasino/writing-coach/internal/review"
	"github.com/tomasino/writing-coach/internal/session"
)

type Server struct {
	Config     config.Config
	Store      *db.Store
	Prompts    prompt.Service
	Reviews    review.Service
	Curriculum curriculum.Service
}

func (s Server) Serve(ctx context.Context) error {
	server := &http.Server{
		Addr:    s.Config.HTTPAddr,
		Handler: s.routes(),
	}

	go func() {
		<-ctx.Done()
		_ = server.Shutdown(context.Background())
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/context", s.handleContext)
	mux.HandleFunc("GET /api/dashboard", s.handleDashboard)
	mux.HandleFunc("POST /api/prompts/next", s.handlePromptNext)
	mux.HandleFunc("POST /api/prompts/revise", s.handlePromptRevise)
	mux.HandleFunc("POST /api/submissions", s.handleSubmissionCreate)
	mux.HandleFunc("GET /api/submissions/{id}", s.handleSubmissionGet)
	mux.HandleFunc("POST /api/reviews", s.handleReviewCreate)
	mux.HandleFunc("GET /api/compare", s.handleCompare)
	return withCORS(mux)
}

type errorResponse struct {
	Error string `json:"error"`
}

type requestContextResponse struct {
	UserSlug string `json:"user_slug"`
	TreeSlug string `json:"tree_slug"`
	UserID   int64  `json:"user_id"`
	TreeID   int64  `json:"tree_id"`
}

type curriculumStateResponse struct {
	ID              int64  `json:"id"`
	CurrentFocus    string `json:"current_focus"`
	DifficultyLevel int    `json:"difficulty_level"`
	LastReviewID    int64  `json:"last_review_id"`
	UpdatedAt       string `json:"updated_at"`
}

type tgoResponse struct {
	ID            int64    `json:"id"`
	Code          string   `json:"code"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Stage         string   `json:"stage"`
	StageOrder    int      `json:"stage_order"`
	ActiveSlot    int      `json:"active_slot,omitempty"`
	Prerequisites []string `json:"prerequisites,omitempty"`
	MasteryHint   string   `json:"mastery_hint,omitempty"`
}

type exerciseResponse struct {
	ID              int64    `json:"id"`
	Title           string   `json:"title"`
	Brief           string   `json:"brief"`
	Constraints     []string `json:"constraints"`
	FocusSkills     []string `json:"focus_skills"`
	TGOCodes        []string `json:"tgo_codes"`
	SuccessCriteria []string `json:"success_criteria"`
	GenerationKind  string   `json:"generation_kind"`
	ProviderNote    string   `json:"provider_note,omitempty"`
}

type submissionResponse struct {
	ID                 int64  `json:"id"`
	ExerciseID         int64  `json:"exercise_id"`
	ParentSubmissionID int64  `json:"parent_submission_id,omitempty"`
	DraftNumber        int    `json:"draft_number"`
	Content            string `json:"content"`
	WordCount          int    `json:"word_count"`
	CreatedAt          string `json:"created_at"`
}

type scoreResponse struct {
	Skill string `json:"skill"`
	Score int    `json:"score"`
}

type tgoAssessmentResponse struct {
	TGOCode  string `json:"tgo_code"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

type reviewResponse struct {
	ID               int64                   `json:"id"`
	SubmissionID     int64                   `json:"submission_id"`
	ReviewKind       string                  `json:"review_kind"`
	ProviderNote     string                  `json:"provider_note,omitempty"`
	Summary          string                  `json:"summary"`
	Strengths        []string                `json:"strengths"`
	Weaknesses       []string                `json:"weaknesses"`
	AnalyzerFindings []string                `json:"analyzer_findings"`
	NextFocus        string                  `json:"next_focus"`
	MetricWordCount  int                     `json:"metric_word_count"`
	TGOAssessments   []tgoAssessmentResponse `json:"tgo_assessments"`
}

type comparisonResponse struct {
	Summary              string   `json:"summary"`
	WordDelta            int      `json:"word_delta"`
	AddedWords           []string `json:"added_words"`
	RemovedWords         []string `json:"removed_words"`
	AddressedWeaknesses  []string `json:"addressed_weaknesses"`
	PersistingWeaknesses []string `json:"persisting_weaknesses"`
}

func (s Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s Server) handleContext(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, requestContextResponse{
		UserSlug: appContext.UserSlug,
		TreeSlug: appContext.TreeSlug,
		UserID:   appContext.UserID,
		TreeID:   appContext.TreeID,
	})
}

func (s Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	state, err := s.Store.GetCurriculumState(r.Context(), appContext.EnrollmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	activeTGOs, err := s.Store.ActiveTGOs(r.Context(), appContext.EnrollmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	completedTGOs, err := s.Store.CompletedTGOs(r.Context(), appContext.EnrollmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	progress, err := s.Store.ProgressReport(r.Context(), appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	strongest, weakest, err := s.Store.StrongestWeakestSkills(r.Context(), appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	recurringWeaknesses, err := s.Store.RecurringWeaknesses(r.Context(), appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	recurringFindings, err := s.Store.RecurringAnalyzerFindings(r.Context(), appContext.UserID, appContext.TreeID, 5)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	history, err := s.Store.History(r.Context(), appContext.UserID, appContext.TreeID)
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
	upcoming := domain.NextUnlockedTGOs(completedSet, activeSet, 3)

	writeJSON(w, http.StatusOK, map[string]any{
		"context":              requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"curriculum_state":     toCurriculumStateResponse(state),
		"active_tgos":          toTGOResponses(activeTGOs),
		"completed_tgos":       toTGOResponses(completedTGOs),
		"upcoming_tgos":        toTGOResponses(upcoming),
		"progress_lines":       progress,
		"strongest_skills":     strongest,
		"weakest_skills":       weakest,
		"recurring_weaknesses": recurringWeaknesses,
		"recurring_findings":   recurringFindings,
		"history":              history,
	})
}

func (s Server) handlePromptNext(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	ex, err := s.createNextExercise(r.Context(), appContext)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
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
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
		return
	}
	if payload.SubmissionID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("submission_id is required"))
		return
	}
	ex, err := s.createRevisionExercise(r.Context(), appContext, payload.SubmissionID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
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
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
		return
	}
	if payload.ExerciseID == 0 || strings.TrimSpace(payload.Content) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("exercise_id and content are required"))
		return
	}
	sub := domain.Submission{
		UserID:             appContext.UserID,
		TreeID:             appContext.TreeID,
		ExerciseID:         payload.ExerciseID,
		ParentSubmissionID: payload.ParentSubmissionID,
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

func (s Server) handleSubmissionGet(w http.ResponseWriter, r *http.Request) {
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
	writeJSON(w, http.StatusOK, map[string]any{"submission": toSubmissionResponse(sub)})
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
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
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
	activeTGOs, err := s.Store.ActiveTGOs(r.Context(), appContext.EnrollmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	reviewResult, scores := s.Reviews.ReviewSubmission(r.Context(), sub, activeTGOs)
	reviewResult.UserID = appContext.UserID
	reviewResult.TreeID = appContext.TreeID
	recommendation, err := s.Curriculum.SyncTGOs(r.Context(), s.Store, appContext.EnrollmentID, reviewResult)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	reviewResult.NextFocus = recommendation.Focus
	reviewID, err := s.Store.SaveReview(r.Context(), reviewResult, scores)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.Store.UpdateCurriculumState(r.Context(), appContext.EnrollmentID, recommendation.Focus, recommendation.Difficulty, reviewID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	reviewResult.ID = reviewID
	writeJSON(w, http.StatusCreated, map[string]any{
		"context":        requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"review":         toReviewResponse(reviewResult),
		"skill_scores":   toScoreResponses(scores),
		"recommendation": recommendation,
	})
}

func (s Server) handleCompare(w http.ResponseWriter, r *http.Request) {
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

func (s Server) createNextExercise(ctx context.Context, appContext session.Context) (domain.Exercise, error) {
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

	ex := s.Prompts.NextExercise(ctx, prompt.Context{
		CurriculumState:   state,
		ActiveTGOs:        activeTGOs,
		RecentTitles:      recentTitles,
		RecentWeaknesses:  recentWeaknesses,
		RecurringFindings: recurringFindings,
	})
	ex.UserID = appContext.UserID
	ex.TreeID = appContext.TreeID
	id, err := s.Store.SaveExercise(ctx, ex)
	if err != nil {
		return domain.Exercise{}, err
	}
	ex.ID = id
	return ex, nil
}

func (s Server) createRevisionExercise(ctx context.Context, appContext session.Context, submissionID int64) (domain.Exercise, error) {
	sub, err := s.Store.GetSubmission(ctx, submissionID)
	if err != nil {
		return domain.Exercise{}, err
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

	var cmp *review.Comparison
	if previous, err := s.Store.PreviousSubmission(ctx, sub); err == nil {
		if previousReview, err := s.Store.LatestReviewForSubmission(ctx, previous.ID); err == nil {
			comparison := review.CompareSubmissions(sub, previous, reviewResult, previousReview)
			cmp = &comparison
		}
	}

	ex := s.Prompts.RevisionExercise(ctx, prompt.Context{
		CurriculumState:    state,
		ActiveTGOs:         activeTGOs,
		RecentTitles:       recentTitles,
		RecentWeaknesses:   recentWeaknesses,
		RecurringFindings:  recurringFindings,
		RevisionOf:         &sub,
		RevisionReview:     &reviewResult,
		RevisionComparison: cmp,
	})
	ex.UserID = appContext.UserID
	ex.TreeID = appContext.TreeID
	id, err := s.Store.SaveExercise(ctx, ex)
	if err != nil {
		return domain.Exercise{}, err
	}
	ex.ID = id
	return ex, nil
}

func (s Server) resolveSession(ctx context.Context, r *http.Request) (session.Context, error) {
	userSlug := firstNonEmpty(r.URL.Query().Get("user"), r.Header.Get("X-Writing-Coach-User"), s.Config.DefaultUserSlug)
	treeSlug := firstNonEmpty(r.URL.Query().Get("tree"), r.Header.Get("X-Writing-Coach-Tree"), s.Config.DefaultTreeSlug)
	userName := firstNonEmpty(r.URL.Query().Get("user_name"), s.Config.WriterName)

	userID, treeID, enrollmentID, err := s.Store.EnsureDefaultUserTree(ctx, userSlug, userName, treeSlug)
	if err != nil {
		return session.Context{}, err
	}
	return session.Context{
		UserID:       userID,
		TreeID:       treeID,
		EnrollmentID: enrollmentID,
		UserSlug:     userSlug,
		TreeSlug:     treeSlug,
	}, nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, errorResponse{Error: err.Error()})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Writing-Coach-User, X-Writing-Coach-Tree")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func toExerciseResponse(ex domain.Exercise) exerciseResponse {
	return exerciseResponse{
		ID:              ex.ID,
		Title:           ex.Title,
		Brief:           ex.Brief,
		Constraints:     ex.Constraints,
		FocusSkills:     ex.FocusSkills,
		TGOCodes:        ex.TGOCodes,
		SuccessCriteria: ex.SuccessCriteria,
		GenerationKind:  ex.GenerationKind,
		ProviderNote:    ex.ProviderNote,
	}
}

func toCurriculumStateResponse(state domain.CurriculumState) curriculumStateResponse {
	return curriculumStateResponse{
		ID:              state.ID,
		CurrentFocus:    state.CurrentFocus,
		DifficultyLevel: state.DifficultyLevel,
		LastReviewID:    state.LastReviewID,
		UpdatedAt:       db.Since(state.UpdatedAt),
	}
}

func toTGOResponses(tgos []domain.TGO) []tgoResponse {
	var out []tgoResponse
	for _, tgo := range tgos {
		out = append(out, tgoResponse{
			ID:            tgo.ID,
			Code:          tgo.Code,
			Title:         tgo.Title,
			Description:   tgo.Description,
			Stage:         tgo.Stage,
			StageOrder:    tgo.StageOrder,
			ActiveSlot:    tgo.ActiveSlot,
			Prerequisites: tgo.Prerequisites,
			MasteryHint:   tgo.MasteryHint,
		})
	}
	return out
}

func toSubmissionResponse(sub domain.Submission) submissionResponse {
	return submissionResponse{
		ID:                 sub.ID,
		ExerciseID:         sub.ExerciseID,
		ParentSubmissionID: sub.ParentSubmissionID,
		DraftNumber:        sub.DraftNumber,
		Content:            sub.Content,
		WordCount:          sub.WordCount,
		CreatedAt:          db.Since(sub.CreatedAt),
	}
}

func toReviewResponse(reviewResult domain.Review) reviewResponse {
	out := reviewResponse{
		ID:               reviewResult.ID,
		SubmissionID:     reviewResult.SubmissionID,
		ReviewKind:       reviewResult.ReviewKind,
		ProviderNote:     reviewResult.ProviderNote,
		Summary:          reviewResult.Summary,
		Strengths:        reviewResult.Strengths,
		Weaknesses:       reviewResult.Weaknesses,
		AnalyzerFindings: reviewResult.AnalyzerFindings,
		NextFocus:        reviewResult.NextFocus,
		MetricWordCount:  reviewResult.MetricWordCount,
	}
	for _, assessment := range reviewResult.TGOAssessments {
		out.TGOAssessments = append(out.TGOAssessments, tgoAssessmentResponse{
			TGOCode:  assessment.TGOCode,
			Status:   assessment.Status,
			Evidence: assessment.Evidence,
		})
	}
	return out
}

func toScoreResponses(scores []domain.SkillScore) []scoreResponse {
	var out []scoreResponse
	for _, score := range scores {
		out = append(out, scoreResponse{Skill: score.Skill, Score: score.Score})
	}
	return out
}
