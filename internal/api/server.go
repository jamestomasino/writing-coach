package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/curriculum"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/prompt"
	"github.com/tomasino/writing-coach/internal/review"
	"github.com/tomasino/writing-coach/internal/session"
)

type Server struct {
	Config            config.Config
	Store             *db.Store
	Prompts           prompt.Service
	Reviews           review.Service
	Curriculum        curriculum.Service
	validationLimiter *aiValidationLimiter
	eventRecorder     *aiProviderEventRecorder
	calibration       *calibrationMaintainer
}

const serverShutdownTimeout = 10 * time.Second
const serverReadHeaderTimeout = 5 * time.Second
const serverReadTimeout = 30 * time.Second
const serverWriteTimeout = 120 * time.Second
const serverIdleTimeout = 120 * time.Second

func (s *Server) Serve(ctx context.Context) error {
	s.startBackgroundWorkers(ctx)

	server := &http.Server{
		Addr:              s.Config.HTTPAddr,
		Handler:           s.routes(),
		ReadHeaderTimeout: serverReadHeaderTimeout,
		ReadTimeout:       serverReadTimeout,
		WriteTimeout:      serverWriteTimeout,
		IdleTimeout:       serverIdleTimeout,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

type errorResponse struct {
	Error string `json:"error"`
}

type aiProviderSettingsResponse struct {
	Provider                         string `json:"provider,omitempty"`
	BaseURLOverride                  string `json:"base_url_override,omitempty"`
	PromptModelOverride              string `json:"prompt_model_override,omitempty"`
	ReviewModelOverride              string `json:"review_model_override,omitempty"`
	Enabled                          bool   `json:"enabled"`
	HasKey                           bool   `json:"has_key"`
	KeyLast4                         string `json:"key_last4,omitempty"`
	ValidatedAt                      string `json:"validated_at,omitempty"`
	LastValidationError              string `json:"last_validation_error,omitempty"`
	EffectiveProvider                string `json:"effective_provider"`
	SystemFallback                   bool   `json:"system_fallback"`
	PersonalProviderStorageAvailable bool   `json:"personal_provider_storage_available"`
	Ready                            bool   `json:"ready"`
}

type aiProviderSettingsPayload struct {
	Provider            string `json:"provider"`
	APIKey              string `json:"api_key"`
	BaseURLOverride     string `json:"base_url_override"`
	PromptModelOverride string `json:"prompt_model_override"`
	ReviewModelOverride string `json:"review_model_override"`
	Enabled             bool   `json:"enabled"`
}

type aiProviderEventResponse struct {
	ID         int64          `json:"id"`
	UserID     int64          `json:"user_id"`
	UserSlug   string         `json:"user_slug,omitempty"`
	Provider   string         `json:"provider"`
	Event      string         `json:"event"`
	Category   string         `json:"category,omitempty"`
	StatusCode int            `json:"status_code,omitempty"`
	Details    map[string]any `json:"details,omitempty"`
	CreatedAt  string         `json:"created_at"`
}

type aiProviderEventCountResponse struct {
	Label string `json:"label"`
	Count int    `json:"count"`
}

type aiProviderEventSummaryResponse struct {
	Since               string                         `json:"since"`
	Total               int                            `json:"total"`
	ValidationFailures  int                            `json:"validation_failures"`
	ValidationRateLimit int                            `json:"validation_rate_limit"`
	Fallbacks           int                            `json:"fallbacks"`
	ProviderCounts      []aiProviderEventCountResponse `json:"provider_counts"`
	EventCounts         []aiProviderEventCountResponse `json:"event_counts"`
	CategoryCounts      []aiProviderEventCountResponse `json:"category_counts"`
}

type aiProviderEventFiltersResponse struct {
	Hours     int      `json:"hours"`
	Provider  string   `json:"provider,omitempty"`
	Event     string   `json:"event,omitempty"`
	Providers []string `json:"providers,omitempty"`
	Events    []string `json:"events,omitempty"`
}

type skillGraphRegionResponse struct {
	Slug           string   `json:"slug"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	SeedCodes      []string `json:"seed_codes"`
	PrioritySkills []string `json:"priority_skills"`
	NodeCodes      []string `json:"node_codes"`
}

type skillGraphNodeResponse struct {
	tgoResponse
	SourceTreeSlug  string   `json:"source_tree_slug"`
	SourceTreeTitle string   `json:"source_tree_title"`
	Unlocks         []string `json:"unlocks"`
}

type curriculumStateResponse struct {
	ID                    int64  `json:"id"`
	CurrentFocus          string `json:"current_focus"`
	DifficultyLevel       int    `json:"difficulty_level"`
	LastReviewID          int64  `json:"last_review_id"`
	ProgressionHoldActive bool   `json:"progression_hold_active"`
	ProgressionHoldReason string `json:"progression_hold_reason_code,omitempty"`
	HoldTriggerReviewID   int64  `json:"hold_trigger_review_id,omitempty"`
	HoldClearedReviewID   int64  `json:"hold_cleared_review_id,omitempty"`
	HoldUpdatedAt         string `json:"hold_updated_at,omitempty"`
	UpdatedAt             string `json:"updated_at"`
}

type historyItemResponse struct {
	Title string   `json:"title"`
	TGOs  []string `json:"tgos,omitempty"`
}

type assignmentResponse struct {
	RootExerciseID    int64                    `json:"root_exercise_id"`
	CurrentExerciseID int64                    `json:"current_exercise_id"`
	Title             string                   `json:"title"`
	IsCurrent         bool                     `json:"is_current,omitempty"`
	IsClosed          bool                     `json:"is_closed,omitempty"`
	LatestStepID      string                   `json:"latest_step_id,omitempty"`
	Steps             []assignmentStepResponse `json:"steps"`
}

type assignmentListItemResponse struct {
	RootExerciseID    int64    `json:"root_exercise_id"`
	CurrentExerciseID int64    `json:"current_exercise_id"`
	Title             string   `json:"title"`
	LatestActivity    string   `json:"latest_activity"`
	LatestStepLabel   string   `json:"latest_step_label"`
	IsCurrent         bool     `json:"is_current,omitempty"`
	IsClosed          bool     `json:"is_closed,omitempty"`
	ExerciseCount     int      `json:"exercise_count"`
	DraftCount        int      `json:"draft_count"`
	ReviewCount       int      `json:"review_count"`
	RevisionCount     int      `json:"revision_count"`
	TGOs              []string `json:"tgos,omitempty"`
}

type assignmentStepResponse struct {
	ID           string              `json:"id"`
	Kind         string              `json:"kind"`
	Title        string              `json:"title"`
	Label        string              `json:"label"`
	CreatedAt    string              `json:"created_at"`
	ExerciseID   int64               `json:"exercise_id,omitempty"`
	SubmissionID int64               `json:"submission_id,omitempty"`
	ReviewID     int64               `json:"review_id,omitempty"`
	DraftNumber  int                 `json:"draft_number,omitempty"`
	Exercise     *exerciseResponse   `json:"exercise,omitempty"`
	Submission   *submissionResponse `json:"submission,omitempty"`
	Review       *reviewResponse     `json:"review,omitempty"`
}

type tgoResponse struct {
	ID             int64    `json:"id"`
	Code           string   `json:"code"`
	Title          string   `json:"title"`
	Description    string   `json:"description"`
	SkillName      string   `json:"skill_name,omitempty"`
	Stage          string   `json:"stage"`
	SkillTier      string   `json:"skill_tier,omitempty"`
	StageOrder     int      `json:"stage_order"`
	ActiveSlot     int      `json:"active_slot,omitempty"`
	Prerequisites  []string `json:"prerequisites,omitempty"`
	MasteryHint    string   `json:"mastery_hint,omitempty"`
	ProgressMode   string   `json:"progress_mode,omitempty"`
	MasteryStage   string   `json:"mastery_stage,omitempty"`
	MasteryPercent int      `json:"mastery_percent,omitempty"`
	EvidenceCount  int      `json:"mastery_evidence_count,omitempty"`
}

type exerciseResponse struct {
	ID                 int64    `json:"id"`
	Title              string   `json:"title"`
	Brief              string   `json:"brief"`
	Constraints        []string `json:"constraints"`
	FocusSkills        []string `json:"focus_skills"`
	TGOCodes           []string `json:"tgo_codes"`
	SuccessCriteria    []string `json:"success_criteria"`
	GenerationKind     string   `json:"generation_kind"`
	ProviderNote       string   `json:"provider_note,omitempty"`
	SourceSubmissionID int64    `json:"source_submission_id,omitempty"`
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
	Skill         string         `json:"skill"`
	Score         int            `json:"score"`
	ScoreSource   string         `json:"score_source,omitempty"`
	ScoreVersion  string         `json:"score_version,omitempty"`
	ScoreEvidence map[string]any `json:"score_evidence,omitempty"`
}

type objectiveScoreResponse struct {
	TGOCode       string         `json:"tgo_code"`
	TGOTitle      string         `json:"tgo_title,omitempty"`
	Score         int            `json:"score"`
	ScoreSource   string         `json:"score_source,omitempty"`
	ScoreVersion  string         `json:"score_version,omitempty"`
	ScoreEvidence map[string]any `json:"score_evidence,omitempty"`
}

type tgoAssessmentResponse struct {
	TGOCode  string `json:"tgo_code"`
	TGOTitle string `json:"tgo_title,omitempty"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

type reviewResponse struct {
	ID                 int64                    `json:"id"`
	SubmissionID       int64                    `json:"submission_id"`
	ReviewKind         string                   `json:"review_kind"`
	ProviderNote       string                   `json:"provider_note,omitempty"`
	Summary            string                   `json:"summary"`
	Strengths          []string                 `json:"strengths"`
	Weaknesses         []string                 `json:"weaknesses"`
	AnalyzerFindings   []string                 `json:"analyzer_findings"`
	NextFocus          string                   `json:"next_focus"`
	MetricWordCount    int                      `json:"metric_word_count"`
	SkillScores        []scoreResponse          `json:"skill_scores"`
	ObjectiveScores    []objectiveScoreResponse `json:"objective_scores,omitempty"`
	TGOAssessments     []tgoAssessmentResponse  `json:"tgo_assessments"`
	CompletedTGOChecks []tgoAssessmentResponse  `json:"completed_tgo_checks"`
	Annotations        []annotationResponse     `json:"annotations,omitempty"`
	Artifacts          *reviewArtifactsPayload  `json:"artifacts,omitempty"`
}

type comparisonResponse struct {
	Summary              string              `json:"summary"`
	WordDelta            int                 `json:"word_delta"`
	AddedWords           []string            `json:"added_words"`
	RemovedWords         []string            `json:"removed_words"`
	AddressedWeaknesses  []string            `json:"addressed_weaknesses"`
	PersistingWeaknesses []string            `json:"persisting_weaknesses"`
	SkillSetMismatch     bool                `json:"skill_set_mismatch,omitempty"`
	SkillDeltas          []review.SkillDelta `json:"skill_deltas,omitempty"`
}

type reviewArtifactsPayload struct {
	AnalyzerReport map[string]any       `json:"analyzer_report,omitempty"`
	Recommendation map[string]any       `json:"recommendation,omitempty"`
	Comparison     map[string]any       `json:"comparison,omitempty"`
	Annotations    []annotationResponse `json:"annotations,omitempty"`
}

type annotationResponse struct {
	Quote    string `json:"quote"`
	TGOCode  string `json:"tgo_code"`
	TGOTitle string `json:"tgo_title,omitempty"`
	Category string `json:"category"`
	Comment  string `json:"comment"`
	Severity string `json:"severity"`
}

type reviewJobResponse struct {
	ID           int64  `json:"id"`
	SubmissionID int64  `json:"submission_id"`
	ReviewID     int64  `json:"review_id,omitempty"`
	Status       string `json:"status"`
	AttemptCount int    `json:"attempt_count"`
	MaxAttempts  int    `json:"max_attempts"`
	LastError    string `json:"last_error,omitempty"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type aiJobResultResponse struct {
	Exercise *exerciseResponse `json:"exercise,omitempty"`
	Review   *reviewResponse   `json:"review,omitempty"`
}

type aiJobResponse struct {
	ID           int64                `json:"id"`
	Kind         string               `json:"kind"`
	SubmissionID int64                `json:"submission_id,omitempty"`
	ExerciseID   int64                `json:"exercise_id,omitempty"`
	ReviewID     int64                `json:"review_id,omitempty"`
	Status       string               `json:"status"`
	AttemptCount int                  `json:"attempt_count"`
	MaxAttempts  int                  `json:"max_attempts"`
	LastError    string               `json:"last_error,omitempty"`
	CreatedAt    string               `json:"created_at"`
	UpdatedAt    string               `json:"updated_at"`
	Result       *aiJobResultResponse `json:"result,omitempty"`
}

type playgroundDraftResponse struct {
	ID            int64  `json:"id"`
	SessionID     int64  `json:"session_id"`
	ParentDraftID int64  `json:"parent_draft_id,omitempty"`
	Content       string `json:"content"`
	WordCount     int    `json:"word_count"`
	CreatedAt     string `json:"created_at"`
}

type playgroundSessionResponse struct {
	ID               int64  `json:"id"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	WritingLanguage  string `json:"writing_language"`
	WritingType      string `json:"writing_type,omitempty"`
	AssignmentFormat string `json:"assignment_format,omitempty"`
	CoachingBrief    string `json:"coaching_brief,omitempty"`
	LatestDraftID    int64  `json:"latest_draft_id,omitempty"`
	LatestReviewID   int64  `json:"latest_review_id,omitempty"`
	LatestReviewAt   string `json:"latest_review_at,omitempty"`
	DraftCount       int    `json:"draft_count"`
	ReviewCount      int    `json:"review_count"`
	CreatedAt        string `json:"created_at"`
	UpdatedAt        string `json:"updated_at"`
}

type playgroundReviewResponse struct {
	ID        int64          `json:"id"`
	SessionID int64          `json:"session_id"`
	DraftID   int64          `json:"draft_id,omitempty"`
	CreatedAt string         `json:"created_at"`
	Review    reviewResponse `json:"review"`
}

func (s Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (s Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := s.Store.SQL.PingContext(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"ok": false, "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "db": "ready"})
}

func (s Server) handleSkillGraph(w http.ResponseWriter, r *http.Request) {
	graph := domain.SkillGraphFromBuiltIns()
	regions := make([]skillGraphRegionResponse, 0, len(graph.Regions))
	for _, region := range graph.Regions {
		regions = append(regions, skillGraphRegionResponse{
			Slug:           region.Slug,
			Title:          region.Title,
			Description:    region.Description,
			SeedCodes:      append([]string(nil), region.SeedCodes...),
			PrioritySkills: append([]string(nil), region.PrioritySkills...),
			NodeCodes:      append([]string(nil), region.NodeCodes...),
		})
	}
	nodes := make([]skillGraphNodeResponse, 0, len(graph.Nodes))
	for _, node := range graph.Nodes {
		nodes = append(nodes, skillGraphNodeResponse{
			tgoResponse: tgoResponse{
				ID:            node.ID,
				Code:          node.Code,
				Title:         node.Title,
				Description:   node.Description,
				SkillName:     strings.TrimSpace(domain.TGOCodeToSkill[node.Code]),
				Stage:         node.Stage,
				SkillTier:     skillTierForTGOCode(node.Code),
				StageOrder:    node.StageOrder,
				ActiveSlot:    node.ActiveSlot,
				Prerequisites: append([]string(nil), node.Prerequisites...),
				MasteryHint:   node.MasteryHint,
				ProgressMode:  node.ProgressMode,
			},
			SourceTreeSlug:  node.SourceTreeSlug,
			SourceTreeTitle: node.SourceTreeTitle,
			Unlocks:         append([]string(nil), node.Unlocks...),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"graph": map[string]any{
			"slug":        graph.Slug,
			"title":       graph.Title,
			"description": graph.Description,
			"regions":     regions,
			"nodes":       nodes,
		},
	})
}

func (s Server) handleAccountReset(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.Store.ResetUserData(r.Context(), appContext.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"ok": true,
	})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func listLimit(r *http.Request, fallback, max int) int {
	raw := strings.TrimSpace(r.URL.Query().Get("limit"))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}

func parseOptionalInt64(raw string) (int64, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, nil
	}
	return strconv.ParseInt(raw, 10, 64)
}

func belongsToContext(userID, treeID int64, appContext session.Context) bool {
	return userID == appContext.UserID && treeID == appContext.TreeID
}

func prereqsMetForSelection(tgo domain.TGO, completed map[string]bool) bool {
	for _, prereq := range tgo.Prerequisites {
		if !completed[prereq] {
			return false
		}
	}
	return true
}

func sanitizeStringList(items []string) []string {
	var out []string
	seen := map[string]bool{}
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" || seen[item] {
			continue
		}
		seen[item] = true
		out = append(out, item)
	}
	return out
}

func nextSeedCodes(treeDef domain.TGOTreeDefinition, completed map[string]bool) []string {
	var selected []string
	active := map[string]bool{}
	for _, code := range treeDef.SeedCodes {
		if completed[code] {
			continue
		}
		selected = append(selected, code)
		active[code] = true
		if len(selected) == 3 {
			return selected
		}
	}
	for _, tgo := range domain.NextUnlockedFromDefinition(treeDef, completed, active, 10) {
		if completed[tgo.Code] || active[tgo.Code] {
			continue
		}
		selected = append(selected, tgo.Code)
		active[tgo.Code] = true
		if len(selected) == 3 {
			break
		}
	}
	if len(selected) < 3 {
		for _, tgo := range treeDef.TGOs {
			if completed[tgo.Code] || active[tgo.Code] {
				continue
			}
			selected = append(selected, tgo.Code)
			if len(selected) == 3 {
				break
			}
		}
	}
	return selected
}

func (s Server) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	mode := authModeFromContext(r.Context())
	switch mode {
	case "api_token":
		return true
	case "kratos":
		ident, ok := identityFromContext(r.Context())
		if ok {
			allowed, err := s.Store.IsAdminEmail(r.Context(), ident.Email)
			if err == nil && allowed {
				return true
			}
		}
	case "none":
		if s.Config.AllowInsecureAuth && strings.TrimSpace(s.Config.APIToken) == "" && strings.TrimSpace(s.Config.KratosPublicURL) == "" {
			return true
		}
	}
	writeJSON(w, http.StatusForbidden, errorResponse{Error: "admin access required"})
	return false
}

func (s Server) reviewComparisonPayload(ctx context.Context, sub domain.Submission, currentReview domain.Review) map[string]any {
	comparison := s.reviewComparison(ctx, sub, currentReview)
	if comparison == nil {
		return nil
	}
	return reviewComparisonMap(*comparison)
}

func (s Server) reviewComparison(ctx context.Context, sub domain.Submission, currentReview domain.Review) *review.Comparison {
	previous, err := s.Store.PreviousSubmission(ctx, sub)
	if err != nil {
		return nil
	}
	previousReview, err := s.Store.LatestReviewForSubmission(ctx, previous.ID)
	if err != nil {
		return nil
	}
	comparison := review.CompareSubmissions(sub, previous, currentReview, previousReview)
	if comparison.SkillSetMismatch {
		log.Printf("comparison skill set mismatch: current_submission=%d baseline_submission=%d", sub.ID, previous.ID)
	}
	return &comparison
}

func reviewComparisonMap(comparison review.Comparison) map[string]any {
	return map[string]any{
		"summary":               comparison.Summary,
		"word_delta":            comparison.WordDelta,
		"added_words":           comparison.AddedWords,
		"removed_words":         comparison.RemovedWords,
		"addressed_weaknesses":  comparison.AddressedWeaknesses,
		"persisting_weaknesses": comparison.PersistingWeaknesses,
		"skill_set_mismatch":    comparison.SkillSetMismatch,
		"skill_deltas":          comparison.SkillDeltas,
	}
}

func recommendationArtifactPayload(recommendation curriculum.Recommendation, interventions []review.Intervention) map[string]any {
	return map[string]any{
		"focus":            recommendation.Focus,
		"difficulty":       recommendation.Difficulty,
		"rationale":        recommendation.Rationale,
		"hold_active":      recommendation.HoldActive,
		"hold_reason_code": recommendation.HoldReasonCode,
		"interventions":    interventions,
	}
}

func recommendationArtifactPayloadWithOutcomes(recommendation curriculum.Recommendation, interventions []review.Intervention, outcomes []review.InterventionOutcome) map[string]any {
	payload := recommendationArtifactPayload(recommendation, interventions)
	payload["intervention_outcomes"] = outcomes
	return payload
}

type reviewInterventionContext struct {
	Comparison            *review.Comparison
	PreviousReviewID      int64
	PreviousInterventions []review.Intervention
}

func (s Server) reviewInterventionData(ctx context.Context, sub domain.Submission, currentReview domain.Review) reviewInterventionContext {
	previous, err := s.Store.PreviousSubmission(ctx, sub)
	if err != nil {
		return reviewInterventionContext{}
	}
	previousReview, err := s.Store.LatestReviewForSubmission(ctx, previous.ID)
	if err != nil {
		return reviewInterventionContext{}
	}
	comparison := review.CompareSubmissions(sub, previous, currentReview, previousReview)
	if comparison.SkillSetMismatch {
		log.Printf("comparison skill set mismatch: current_submission=%d baseline_submission=%d", sub.ID, previous.ID)
	}
	context := reviewInterventionContext{
		Comparison:       &comparison,
		PreviousReviewID: previousReview.ID,
	}
	if artifacts, err := s.Store.GetReviewArtifacts(ctx, previousReview.ID); err == nil {
		context.PreviousInterventions = decodeInterventionsFromRecommendationJSON(artifacts.RecommendationJSON)
	}
	return context
}

func decodeInterventionsFromRecommendationJSON(raw string) []review.Intervention {
	if strings.TrimSpace(raw) == "" {
		return nil
	}
	var payload struct {
		Interventions []review.Intervention `json:"interventions"`
	}
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil
	}
	return payload.Interventions
}

func comparisonPayload(ctx reviewInterventionContext) map[string]any {
	if ctx.Comparison == nil {
		return nil
	}
	return reviewComparisonMap(*ctx.Comparison)
}

func (s Server) runReviewWorker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Store.RequeueStaleReviewJobs(ctx, 3*time.Minute); err != nil {
				log.Printf("review worker: requeue stale jobs failed: %v", err)
			}
			if err := s.processNextReviewJob(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Printf("review worker: process failed: %v", err)
			}
		}
	}
}

func (s Server) processNextReviewJob(ctx context.Context) error {
	job, err := s.Store.ClaimNextReviewJob(ctx)
	if err != nil {
		return err
	}
	log.Printf("review job started: job=%d submission=%d attempt=%d", job.ID, job.SubmissionID, job.AttemptCount)
	if err := s.processReviewJob(ctx, job); err != nil {
		log.Printf("review job failed: job=%d submission=%d attempt=%d err=%v", job.ID, job.SubmissionID, job.AttemptCount, err)
		if failErr := s.Store.FailReviewJob(ctx, job, err.Error()); failErr != nil {
			log.Printf("review job failure update failed: job=%d err=%v", job.ID, failErr)
		}
		return err
	}
	log.Printf("review job completed: job=%d submission=%d", job.ID, job.SubmissionID)
	return nil
}

func (s Server) processReviewJob(ctx context.Context, job domain.ReviewJob) error {
	sub, err := s.Store.GetSubmission(ctx, job.SubmissionID)
	if err != nil {
		return fmt.Errorf("load submission: %w", err)
	}
	if existing, err := s.Store.LatestReviewForSubmission(ctx, sub.ID); err == nil {
		return s.Store.CompleteReviewJob(ctx, job.ID, existing.ID)
	}
	activeTGOs, err := s.Store.ActiveTGOs(ctx, job.EnrollmentID)
	if err != nil {
		return fmt.Errorf("load active tgos: %w", err)
	}
	completedTGOs, err := s.Store.CompletedTGOs(ctx, job.EnrollmentID)
	if err != nil {
		return fmt.Errorf("load completed tgos: %w", err)
	}
	treeSlug, err := s.treeSlugForJob(ctx, job)
	if err != nil {
		return fmt.Errorf("load tree slug: %w", err)
	}
	analyzerContext := analyzer.ContextOptions{TreeSlug: treeSlug}
	if profile, err := s.Store.OnboardingProfileByEnrollmentID(ctx, job.EnrollmentID); err == nil {
		analyzerContext = analyzer.ContextFromProfile(treeSlug, profile)
	}
	runtime, err := s.resolveLLMRuntime(ctx, job.UserID)
	if err != nil {
		return fmt.Errorf("resolve provider: %w", err)
	}

	reviewResult := s.Reviews.WithClient(runtime.Client, runtime.ProviderKind).ReviewSubmissionDetailedWithOptions(ctx, sub, activeTGOs, completedTGOs, review.Options{
		AnalyzerContext: analyzerContext,
	})
	if reviewResult.Review.ReviewKind == runtime.ProviderKind {
		reviewResult.Review.ProviderNote = formatProviderNote(runtime.ProviderKind, runtime.ReviewModel)
	}
	if reviewResult.Review.ReviewKind == "deterministic-fallback" {
		s.logAIProviderEvent("generation_fallback", runtime.ProviderKind, job.UserID, map[string]any{
			"artifact": "review",
			"reason":   strings.TrimSpace(reviewResult.Review.ProviderNote),
		})
	}
	reviewResult.Review.UserID = job.UserID
	reviewResult.Review.TreeID = job.TreeID
	stateBefore, err := s.Store.GetCurriculumState(ctx, job.EnrollmentID)
	if err != nil {
		return fmt.Errorf("load curriculum state before review sync: %w", err)
	}
	recommendation, err := s.Curriculum.SyncTGOs(ctx, s.Store, treeSlug, job.EnrollmentID, reviewResult.Review)
	if err != nil {
		return fmt.Errorf("sync curriculum: %w", err)
	}
	reviewResult.Review.NextFocus = recommendation.Focus
	objectiveScores := review.BuildObjectiveScores(sub.ID, activeTGOs, reviewResult.Review.TGOAssessments, reviewResult.Scores)
	reviewID, err := s.Store.SaveReviewWithObjectiveScores(ctx, reviewResult.Review, reviewResult.Scores, objectiveScores)
	if err != nil {
		return fmt.Errorf("save review: %w", err)
	}
	interventionContext := s.reviewInterventionData(ctx, sub, reviewResult.Review)
	interventions := review.PrioritizeInterventions(reviewResult.Review, interventionContext.Comparison)
	outcomes := review.BuildInterventionOutcomes(interventions, interventionContext.PreviousInterventions, interventionContext.Comparison, interventionContext.PreviousReviewID, reviewID)
	if err := s.Store.SaveReviewArtifacts(ctx, domain.ReviewArtifacts{
		ReviewID:           reviewID,
		AnalyzerReportJSON: mustJSON(reviewResult.AnalyzerReport),
		RecommendationJSON: mustJSON(recommendationArtifactPayloadWithOutcomes(recommendation, interventions, outcomes)),
		ComparisonJSON:     mustJSON(comparisonPayload(interventionContext)),
		AnnotationsJSON:    mustJSON(reviewResult.Review.Annotations),
	}); err != nil {
		return fmt.Errorf("save review artifacts: %w", err)
	}
	if err := s.Store.UpdateCurriculumState(ctx, job.EnrollmentID, recommendation.Focus, recommendation.Difficulty, reviewID); err != nil {
		return fmt.Errorf("update curriculum state: %w", err)
	}
	if err := s.Store.UpdateProgressionHoldState(ctx, job.EnrollmentID, recommendation.HoldActive, recommendation.HoldReasonCode, reviewID, s.Config.ProgressionHoldClearStreak); err != nil {
		return fmt.Errorf("update progression hold state: %w", err)
	}
	stateAfter, err := s.Store.GetCurriculumState(ctx, job.EnrollmentID)
	if err != nil {
		return fmt.Errorf("load curriculum state after review sync: %w", err)
	}
	reviewResult.Review.ID = reviewID
	if err := s.emitReviewDecisionEvents(ctx, job.UserID, job.TreeID, job.EnrollmentID, reviewResult.Review, recommendation, len(reviewResult.Scores)); err != nil {
		return fmt.Errorf("save decision events: %w", err)
	}
	if err := s.emitProgressionHoldTransitionEvent(ctx, job.UserID, job.TreeID, job.EnrollmentID, reviewResult.Review, stateBefore.ProgressionHoldActive, stateAfter.ProgressionHoldActive, stateAfter.ProgressionHoldReasonCode); err != nil {
		return fmt.Errorf("save progression hold transition event: %w", err)
	}
	if err := s.Store.CompleteReviewJob(ctx, job.ID, reviewID); err != nil {
		return fmt.Errorf("complete review job: %w", err)
	}
	return nil
}

func (s Server) treeSlugForJob(ctx context.Context, job domain.ReviewJob) (string, error) {
	var slug string
	err := s.Store.SQL.QueryRowContext(ctx, `
		SELECT t.slug
		FROM user_tree_enrollments e
		JOIN tgo_trees t ON t.id = e.tree_id
		WHERE e.id = ?
	`, job.EnrollmentID).Scan(&slug)
	return slug, err
}

func decodeReviewArtifacts(artifacts domain.ReviewArtifacts) *reviewArtifactsPayload {
	payload := &reviewArtifactsPayload{}
	_ = json.Unmarshal([]byte(artifacts.AnalyzerReportJSON), &payload.AnalyzerReport)
	_ = json.Unmarshal([]byte(artifacts.RecommendationJSON), &payload.Recommendation)
	_ = json.Unmarshal([]byte(artifacts.ComparisonJSON), &payload.Comparison)
	_ = json.Unmarshal([]byte(artifacts.AnnotationsJSON), &payload.Annotations)
	for i := range payload.Annotations {
		payload.Annotations[i].TGOTitle = tgoTitleForCode(payload.Annotations[i].TGOCode)
	}
	return payload
}

func mustJSON(v any) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}

func toExerciseResponse(ex domain.Exercise) exerciseResponse {
	return exerciseResponse{
		ID:                 ex.ID,
		Title:              ex.Title,
		Brief:              ex.Brief,
		Constraints:        ex.Constraints,
		FocusSkills:        ex.FocusSkills,
		TGOCodes:           ex.TGOCodes,
		SuccessCriteria:    ex.SuccessCriteria,
		GenerationKind:     ex.GenerationKind,
		ProviderNote:       ex.ProviderNote,
		SourceSubmissionID: ex.SourceSubmissionID,
	}
}

func toOnboardingProfileResponse(profile domain.OnboardingProfile) *onboardingProfileResponse {
	return &onboardingProfileResponse{
		WritingLanguage:     domain.NormalizeWritingLanguage(profile.WritingLanguage),
		WritingType:         profile.WritingType,
		AssignmentFormat:    profile.AssignmentFormat,
		TargetAudience:      profile.TargetAudience,
		SubjectMatter:       profile.SubjectMatter,
		ExperienceLevel:     profile.ExperienceLevel,
		DesiredTone:         profile.DesiredTone,
		BiggestWeaknesses:   append([]string(nil), profile.BiggestWeaknesses...),
		DesiredOutcomes:     append([]string(nil), profile.DesiredOutcomes...),
		DifficultyIntensity: profile.DifficultyIntensity,
		WritingGoals:        profile.WritingGoals,
		GeneratedTreeSlug:   profile.GeneratedTreeSlug,
		TemplateKey:         profile.TemplateKey,
	}
}

func toOnboardingOptionResponses(values []domain.OnboardingOption) []onboardingOptionResponse {
	out := make([]onboardingOptionResponse, 0, len(values))
	for _, value := range values {
		out = append(out, onboardingOptionResponse{
			Value: value.Value,
			Label: value.Label,
		})
	}
	return out
}

func toCurriculumStateResponse(state domain.CurriculumState) curriculumStateResponse {
	holdUpdatedAt := ""
	if !state.HoldUpdatedAt.IsZero() {
		holdUpdatedAt = db.Since(state.HoldUpdatedAt)
	}
	return curriculumStateResponse{
		ID:                    state.ID,
		CurrentFocus:          state.CurrentFocus,
		DifficultyLevel:       state.DifficultyLevel,
		LastReviewID:          state.LastReviewID,
		ProgressionHoldActive: state.ProgressionHoldActive,
		ProgressionHoldReason: state.ProgressionHoldReasonCode,
		HoldTriggerReviewID:   state.HoldTriggerReviewID,
		HoldClearedReviewID:   state.HoldClearedReviewID,
		HoldUpdatedAt:         holdUpdatedAt,
		UpdatedAt:             db.Since(state.UpdatedAt),
	}
}

func toTGOResponses(tgos []domain.TGO) []tgoResponse {
	var out []tgoResponse
	for _, tgo := range tgos {
		out = append(out, tgoResponse{
			ID:             tgo.ID,
			Code:           tgo.Code,
			Title:          tgo.Title,
			Description:    tgo.Description,
			SkillName:      strings.TrimSpace(domain.TGOCodeToSkill[tgo.Code]),
			Stage:          tgo.Stage,
			SkillTier:      skillTierForTGOCode(tgo.Code),
			StageOrder:     tgo.StageOrder,
			ActiveSlot:     tgo.ActiveSlot,
			Prerequisites:  tgo.Prerequisites,
			MasteryHint:    tgo.MasteryHint,
			ProgressMode:   tgo.ProgressMode,
			MasteryStage:   tgo.MasteryStage,
			MasteryPercent: tgo.MasteryPercent,
			EvidenceCount:  tgo.EvidenceCount,
		})
	}
	return out
}

func skillTierForTGOCode(code string) string {
	skill := strings.TrimSpace(domain.TGOCodeToSkill[code])
	return string(domain.SkillTierForName(skill))
}

func toHistoryResponses(exercises []domain.Exercise) []historyItemResponse {
	items := make([]historyItemResponse, 0, len(exercises))
	for _, exercise := range exercises {
		item := historyItemResponse{
			Title: strings.TrimSpace(exercise.Title),
		}
		if item.Title == "" {
			item.Title = "Untitled assignment"
		}
		for _, code := range exercise.TGOCodes {
			if tgo, ok := domain.TGOByCode(code); ok {
				item.TGOs = append(item.TGOs, tgo.Title)
				continue
			}
			if value := strings.TrimSpace(code); value != "" {
				item.TGOs = append(item.TGOs, value)
			}
		}
		items = append(items, item)
	}
	return items
}

func (s Server) assignmentTimeline(ctx context.Context, appContext session.Context, exerciseID int64) (assignmentResponse, error) {
	current, err := s.Store.GetExercise(ctx, exerciseID)
	if err != nil {
		return assignmentResponse{}, err
	}
	if !belongsToContext(current.UserID, current.TreeID, appContext) {
		return assignmentResponse{}, sql.ErrNoRows
	}

	exercises, err := s.Store.ListExercises(ctx, appContext.UserID, appContext.TreeID, 500)
	if err != nil {
		return assignmentResponse{}, err
	}
	submissions, err := s.Store.ListSubmissions(ctx, appContext.UserID, appContext.TreeID, 0, 500)
	if err != nil {
		return assignmentResponse{}, err
	}
	reviews, err := s.Store.ListReviews(ctx, appContext.UserID, appContext.TreeID, 0, 500)
	if err != nil {
		return assignmentResponse{}, err
	}

	exerciseByID := make(map[int64]domain.Exercise, len(exercises))
	for _, exercise := range exercises {
		exerciseByID[exercise.ID] = exercise
	}
	submissionByID := make(map[int64]domain.Submission, len(submissions))
	submissionsByExercise := make(map[int64][]domain.Submission)
	for _, submission := range submissions {
		submissionByID[submission.ID] = submission
		submissionsByExercise[submission.ExerciseID] = append(submissionsByExercise[submission.ExerciseID], submission)
	}
	reviewBySubmission := make(map[int64]domain.Review, len(reviews))
	for _, reviewResult := range reviews {
		if _, exists := reviewBySubmission[reviewResult.SubmissionID]; !exists {
			reviewBySubmission[reviewResult.SubmissionID] = reviewResult
		}
	}

	rootID := rootExerciseID(current, exerciseByID, submissionByID)
	chain := make([]domain.Exercise, 0, len(exercises))
	for _, exercise := range exercises {
		if rootExerciseID(exercise, exerciseByID, submissionByID) == rootID {
			chain = append(chain, exercise)
		}
	}
	sort.Slice(chain, func(i, j int) bool {
		if chain[i].CreatedAt.Equal(chain[j].CreatedAt) {
			return chain[i].ID < chain[j].ID
		}
		return chain[i].CreatedAt.Before(chain[j].CreatedAt)
	})

	steps := make([]assignmentStepResponse, 0, len(chain)*3)
	revisionNumberByExercise := make(map[int64]int, len(chain))
	draftNumberBySubmission := make(map[int64]int, len(submissions))
	revisionNumber := 0
	draftNumber := 0
	for _, exercise := range chain {
		if exercise.SourceSubmissionID != 0 {
			revisionNumber++
			revisionNumberByExercise[exercise.ID] = revisionNumber
		}
		exerciseSubmissions := append([]domain.Submission(nil), submissionsByExercise[exercise.ID]...)
		sort.Slice(exerciseSubmissions, func(i, j int) bool {
			if exerciseSubmissions[i].DraftNumber == exerciseSubmissions[j].DraftNumber {
				return exerciseSubmissions[i].ID < exerciseSubmissions[j].ID
			}
			return exerciseSubmissions[i].DraftNumber < exerciseSubmissions[j].DraftNumber
		})
		for _, submission := range exerciseSubmissions {
			draftNumber++
			draftNumberBySubmission[submission.ID] = draftNumber
		}
	}

	for _, exercise := range chain {
		exerciseCopy := exercise
		exerciseLabel := assignmentStepLabel("exercise", exercise, domain.Submission{}, revisionNumberByExercise[exercise.ID], 0)
		steps = append(steps, assignmentStepResponse{
			ID:         fmt.Sprintf("exercise-%d", exercise.ID),
			Kind:       "exercise",
			Title:      exerciseTitle(exercise),
			Label:      exerciseLabel,
			CreatedAt:  db.Since(exercise.CreatedAt),
			ExerciseID: exercise.ID,
			Exercise:   ptrExerciseResponse(toExerciseResponse(exerciseCopy)),
		})

		exerciseSubmissions := append([]domain.Submission(nil), submissionsByExercise[exercise.ID]...)
		sort.Slice(exerciseSubmissions, func(i, j int) bool {
			if exerciseSubmissions[i].DraftNumber == exerciseSubmissions[j].DraftNumber {
				return exerciseSubmissions[i].ID < exerciseSubmissions[j].ID
			}
			return exerciseSubmissions[i].DraftNumber < exerciseSubmissions[j].DraftNumber
		})
		for _, submission := range exerciseSubmissions {
			submissionCopy := submission
			stepNumber := draftNumberBySubmission[submission.ID]
			submissionLabel := assignmentStepLabel("submission", exercise, submission, revisionNumberByExercise[exercise.ID], stepNumber)
			steps = append(steps, assignmentStepResponse{
				ID:           fmt.Sprintf("submission-%d", submission.ID),
				Kind:         "submission",
				Title:        submissionLabel,
				Label:        submissionLabel,
				CreatedAt:    db.Since(submission.CreatedAt),
				ExerciseID:   exercise.ID,
				SubmissionID: submission.ID,
				DraftNumber:  submission.DraftNumber,
				Submission:   ptrSubmissionResponse(toSubmissionResponse(submissionCopy)),
			})

			reviewResult, ok := reviewBySubmission[submission.ID]
			if !ok {
				continue
			}
			reviewResponse := toReviewResponse(reviewResult)
			if artifacts, err := s.Store.GetReviewArtifacts(ctx, reviewResult.ID); err == nil {
				reviewResponse.Artifacts = decodeReviewArtifacts(artifacts)
				if len(reviewResponse.Annotations) == 0 {
					reviewResponse.Annotations = append(reviewResponse.Annotations, reviewResponse.Artifacts.Annotations...)
				}
			}
			reviewLabel := assignmentStepLabel("review", exercise, submission, revisionNumberByExercise[exercise.ID], stepNumber)
			steps = append(steps, assignmentStepResponse{
				ID:           fmt.Sprintf("review-%d", reviewResult.ID),
				Kind:         "review",
				Title:        reviewLabel,
				Label:        reviewLabel,
				CreatedAt:    db.Since(reviewResult.CreatedAt),
				ExerciseID:   exercise.ID,
				SubmissionID: submission.ID,
				ReviewID:     reviewResult.ID,
				DraftNumber:  submission.DraftNumber,
				Review:       &reviewResponse,
			})
		}
	}

	title := exerciseTitle(exerciseByID[rootID])
	if title == "" {
		title = exerciseTitle(current)
	}
	currentRootID := latestOpenRootExerciseID(exercises, exerciseByID, submissionByID)
	latestStepID := ""
	if len(steps) > 0 {
		latestStepID = steps[len(steps)-1].ID
	}
	return assignmentResponse{
		RootExerciseID:    rootID,
		CurrentExerciseID: current.ID,
		Title:             title,
		IsCurrent:         rootID == currentRootID,
		IsClosed:          assignmentClosed(rootID, exerciseByID),
		LatestStepID:      latestStepID,
		Steps:             steps,
	}, nil
}

func (s Server) assignmentSummaries(ctx context.Context, appContext session.Context) ([]assignmentListItemResponse, error) {
	exercises, err := s.Store.ListExercises(ctx, appContext.UserID, appContext.TreeID, 500)
	if err != nil {
		return nil, err
	}
	submissions, err := s.Store.ListSubmissions(ctx, appContext.UserID, appContext.TreeID, 0, 500)
	if err != nil {
		return nil, err
	}
	reviews, err := s.Store.ListReviews(ctx, appContext.UserID, appContext.TreeID, 0, 500)
	if err != nil {
		return nil, err
	}
	if len(exercises) == 0 {
		return nil, nil
	}

	exerciseByID := make(map[int64]domain.Exercise, len(exercises))
	for _, exercise := range exercises {
		exerciseByID[exercise.ID] = exercise
	}
	submissionByID := make(map[int64]domain.Submission, len(submissions))
	for _, submission := range submissions {
		submissionByID[submission.ID] = submission
	}
	reviewBySubmission := make(map[int64]domain.Review, len(reviews))
	for _, reviewResult := range reviews {
		if _, exists := reviewBySubmission[reviewResult.SubmissionID]; !exists {
			reviewBySubmission[reviewResult.SubmissionID] = reviewResult
		}
	}

	type chainSummary struct {
		rootExerciseID    int64
		currentExerciseID int64
		title             string
		latestActivity    time.Time
		latestStepLabel   string
		exerciseCount     int
		draftCount        int
		reviewCount       int
		revisionCount     int
		tgos              []string
		isClosed          bool
	}
	chains := map[int64]*chainSummary{}
	for _, exercise := range exercises {
		rootID := rootExerciseID(exercise, exerciseByID, submissionByID)
		chain := chains[rootID]
		if chain == nil {
			chain = &chainSummary{
				rootExerciseID:  rootID,
				title:           exerciseTitle(exerciseByID[rootID]),
				latestActivity:  exercise.CreatedAt,
				latestStepLabel: stepLabel("exercise", exercise, domain.Submission{}),
				tgos:            titlesForTGOCodes(exerciseByID[rootID].TGOCodes),
				isClosed:        assignmentClosed(rootID, exerciseByID),
			}
			chains[rootID] = chain
		}
		chain.exerciseCount++
		if exercise.SourceSubmissionID != 0 {
			chain.revisionCount++
		}
		if exercise.CreatedAt.After(chain.latestActivity) || (exercise.CreatedAt.Equal(chain.latestActivity) && exercise.ID > chain.currentExerciseID) {
			chain.latestActivity = exercise.CreatedAt
			chain.latestStepLabel = stepLabel("exercise", exercise, domain.Submission{})
		}
		if exercise.CreatedAt.After(exerciseByID[chain.currentExerciseID].CreatedAt) || chain.currentExerciseID == 0 || (exercise.CreatedAt.Equal(exerciseByID[chain.currentExerciseID].CreatedAt) && exercise.ID > chain.currentExerciseID) {
			chain.currentExerciseID = exercise.ID
		}
	}

	for _, submission := range submissions {
		exercise, ok := exerciseByID[submission.ExerciseID]
		if !ok {
			continue
		}
		rootID := rootExerciseID(exercise, exerciseByID, submissionByID)
		chain := chains[rootID]
		if chain == nil {
			continue
		}
		chain.draftCount++
		if submission.CreatedAt.After(chain.latestActivity) {
			chain.latestActivity = submission.CreatedAt
			chain.latestStepLabel = stepLabel("submission", exercise, submission)
		}
		if reviewResult, ok := reviewBySubmission[submission.ID]; ok {
			chain.reviewCount++
			if reviewResult.CreatedAt.After(chain.latestActivity) {
				chain.latestActivity = reviewResult.CreatedAt
				chain.latestStepLabel = stepLabel("review", exercise, submission)
			}
		}
	}

	currentRootID := latestOpenRootExerciseID(exercises, exerciseByID, submissionByID)
	items := make([]assignmentListItemResponse, 0, len(chains))
	for _, chain := range chains {
		items = append(items, assignmentListItemResponse{
			RootExerciseID:    chain.rootExerciseID,
			CurrentExerciseID: chain.currentExerciseID,
			Title:             chain.title,
			LatestActivity:    db.Since(chain.latestActivity),
			LatestStepLabel:   chain.latestStepLabel,
			ExerciseCount:     chain.exerciseCount,
			DraftCount:        chain.draftCount,
			ReviewCount:       chain.reviewCount,
			RevisionCount:     chain.revisionCount,
			TGOs:              chain.tgos,
			IsCurrent:         chain.rootExerciseID == currentRootID,
			IsClosed:          chain.isClosed,
		})
	}
	sort.Slice(items, func(i, j int) bool {
		left := chains[items[i].RootExerciseID]
		right := chains[items[j].RootExerciseID]
		if left.latestActivity.Equal(right.latestActivity) {
			return items[i].RootExerciseID > items[j].RootExerciseID
		}
		return left.latestActivity.After(right.latestActivity)
	})
	return items, nil
}

func rootExerciseID(exercise domain.Exercise, exerciseByID map[int64]domain.Exercise, submissionByID map[int64]domain.Submission) int64 {
	current := exercise
	seen := map[int64]bool{current.ID: true}
	for current.SourceSubmissionID != 0 {
		submission, ok := submissionByID[current.SourceSubmissionID]
		if !ok {
			break
		}
		parent, ok := exerciseByID[submission.ExerciseID]
		if !ok || seen[parent.ID] {
			break
		}
		current = parent
		seen[current.ID] = true
	}
	return current.ID
}

func assignmentClosed(rootID int64, exerciseByID map[int64]domain.Exercise) bool {
	root, ok := exerciseByID[rootID]
	return ok && !root.ClosedAt.IsZero()
}

func latestOpenRootExerciseID(exercises []domain.Exercise, exerciseByID map[int64]domain.Exercise, submissionByID map[int64]domain.Submission) int64 {
	var latest domain.Exercise
	var found bool
	for _, exercise := range exercises {
		rootID := rootExerciseID(exercise, exerciseByID, submissionByID)
		if assignmentClosed(rootID, exerciseByID) {
			continue
		}
		if !found || exercise.CreatedAt.After(latest.CreatedAt) || (exercise.CreatedAt.Equal(latest.CreatedAt) && exercise.ID > latest.ID) {
			latest = exercise
			found = true
		}
	}
	if !found {
		return 0
	}
	return rootExerciseID(latest, exerciseByID, submissionByID)
}

func exerciseTitle(exercise domain.Exercise) string {
	title := strings.TrimSpace(exercise.Title)
	if title == "" {
		return "Untitled assignment"
	}
	return title
}

func titlesForTGOCodes(codes []string) []string {
	out := make([]string, 0, len(codes))
	for _, code := range codes {
		if tgo, ok := domain.TGOByCode(code); ok {
			out = append(out, tgo.Title)
			continue
		}
		value := strings.TrimSpace(code)
		if value != "" {
			out = append(out, value)
		}
	}
	return out
}

func assignmentStepLabel(kind string, exercise domain.Exercise, submission domain.Submission, revisionNumber int, draftNumber int) string {
	switch kind {
	case "exercise":
		if exercise.SourceSubmissionID != 0 {
			if revisionNumber > 0 {
				return fmt.Sprintf("Revision %d", revisionNumber)
			}
			return "Revision"
		}
		return "Prompt"
	case "submission":
		if draftNumber > 0 {
			return fmt.Sprintf("Draft %d", draftNumber)
		}
		return fmt.Sprintf("Draft %d", submission.DraftNumber)
	case "review":
		if draftNumber > 0 {
			return fmt.Sprintf("Feedback %d", draftNumber)
		}
		return fmt.Sprintf("Feedback %d", submission.DraftNumber)
	default:
		return ""
	}
}

func stepLabel(kind string, exercise domain.Exercise, submission domain.Submission) string {
	return assignmentStepLabel(kind, exercise, submission, 0, 0)
}

func ptrExerciseResponse(value exerciseResponse) *exerciseResponse {
	return &value
}

func ptrSubmissionResponse(value submissionResponse) *submissionResponse {
	return &value
}

func focusSkillsForTGOs(tgos []domain.TGO, fallback []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, tgo := range tgos {
		skill := strings.TrimSpace(domain.TGOCodeToSkill[tgo.Code])
		if skill == "" || seen[skill] {
			continue
		}
		seen[skill] = true
		out = append(out, skill)
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func tgoCodesForExercise(tgos []domain.TGO) []string {
	var out []string
	for _, tgo := range tgos {
		code := strings.TrimSpace(tgo.Code)
		if code == "" {
			continue
		}
		out = append(out, code)
	}
	return out
}

func successCriteriaForTGOs(tgos []domain.TGO, fallback []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, tgo := range tgos {
		criterion := strings.TrimSpace(tgo.Description)
		if criterion == "" {
			criterion = strings.TrimSpace(tgo.MasteryHint)
		}
		if criterion == "" {
			continue
		}
		criterion = strings.ToUpper(criterion[:1]) + criterion[1:]
		if !strings.HasSuffix(criterion, ".") && !strings.HasSuffix(criterion, "!") && !strings.HasSuffix(criterion, "?") {
			criterion += "."
		}
		if seen[criterion] {
			continue
		}
		seen[criterion] = true
		out = append(out, criterion)
	}
	if len(out) == 0 {
		return fallback
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
		SkillScores:      toScoreResponses(reviewResult.SkillScores, reviewResult.TGOAssessments),
		ObjectiveScores:  toObjectiveScoreResponses(reviewResult.ObjectiveScores),
	}
	for _, assessment := range reviewResult.TGOAssessments {
		out.TGOAssessments = append(out.TGOAssessments, tgoAssessmentResponse{
			TGOCode:  assessment.TGOCode,
			TGOTitle: tgoTitleForCode(assessment.TGOCode),
			Status:   assessment.Status,
			Evidence: assessment.Evidence,
		})
	}
	for _, check := range reviewResult.CompletedTGOChecks {
		out.CompletedTGOChecks = append(out.CompletedTGOChecks, tgoAssessmentResponse{
			TGOCode:  check.TGOCode,
			TGOTitle: tgoTitleForCode(check.TGOCode),
			Status:   check.Status,
			Evidence: check.Evidence,
		})
	}
	for _, annotation := range reviewResult.Annotations {
		out.Annotations = append(out.Annotations, annotationResponse{
			Quote:    annotation.Quote,
			TGOCode:  annotation.TGOCode,
			TGOTitle: tgoTitleForCode(annotation.TGOCode),
			Category: annotation.Category,
			Comment:  annotation.Comment,
			Severity: annotation.Severity,
		})
	}
	return out
}

func toPlaygroundSessionResponse(session domain.PlaygroundSession) playgroundSessionResponse {
	latestReviewAt := ""
	if !session.LatestReviewAt.IsZero() {
		latestReviewAt = db.Since(session.LatestReviewAt)
	}
	return playgroundSessionResponse{
		ID:               session.ID,
		Title:            strings.TrimSpace(session.Title),
		Content:          session.Content,
		WritingLanguage:  session.WritingLanguage,
		WritingType:      session.WritingType,
		AssignmentFormat: session.AssignmentFormat,
		CoachingBrief:    session.CoachingBrief,
		LatestDraftID:    session.LatestDraftID,
		LatestReviewID:   session.LatestReviewID,
		LatestReviewAt:   latestReviewAt,
		DraftCount:       session.DraftCount,
		ReviewCount:      session.ReviewCount,
		CreatedAt:        db.Since(session.CreatedAt),
		UpdatedAt:        db.Since(session.UpdatedAt),
	}
}

func toPlaygroundDraftResponse(item domain.PlaygroundDraft) playgroundDraftResponse {
	return playgroundDraftResponse{
		ID:            item.ID,
		SessionID:     item.SessionID,
		ParentDraftID: item.ParentDraftID,
		Content:       item.Content,
		WordCount:     item.WordCount,
		CreatedAt:     db.Since(item.CreatedAt),
	}
}

func toPlaygroundReviewResponse(item domain.PlaygroundReview) playgroundReviewResponse {
	response := toReviewResponse(item.Review)
	var analyzerReport map[string]any
	if strings.TrimSpace(item.AnalyzerReportJSON) != "" {
		_ = json.Unmarshal([]byte(item.AnalyzerReportJSON), &analyzerReport)
	}
	var comparison map[string]any
	if strings.TrimSpace(item.ComparisonJSON) != "" {
		_ = json.Unmarshal([]byte(item.ComparisonJSON), &comparison)
	}
	response.Artifacts = &reviewArtifactsPayload{
		AnalyzerReport: analyzerReport,
		Comparison:     comparison,
		Annotations:    response.Annotations,
	}
	return playgroundReviewResponse{
		ID:        item.ID,
		SessionID: item.SessionID,
		DraftID:   item.DraftID,
		CreatedAt: db.Since(item.CreatedAt),
		Review:    response,
	}
}

func toScoreResponses(scores []domain.SkillScore, assessments []domain.TGOAssessment) []scoreResponse {
	allowedSkills := focusedSkillsFromAssessments(assessments)
	filtered := make([]domain.SkillScore, 0, len(scores))
	for _, score := range scores {
		if score.ScoreSource == "deterministic" {
			filtered = append(filtered, score)
		}
	}
	if len(filtered) == 0 {
		for _, score := range scores {
			if strings.Contains(score.ScoreSource, "legacy") || score.ScoreSource == "" {
				filtered = append(filtered, score)
			}
		}
	}
	if len(filtered) == 0 {
		filtered = scores
	}
	if len(allowedSkills) > 0 {
		focused := make([]domain.SkillScore, 0, len(filtered))
		for _, score := range filtered {
			key := strings.ToLower(strings.TrimSpace(score.Skill))
			if key == "" || !allowedSkills[key] {
				continue
			}
			focused = append(focused, score)
		}
		filtered = focused
	}

	var out []scoreResponse
	for _, score := range filtered {
		item := scoreResponse{
			Skill:        score.Skill,
			Score:        score.Score,
			ScoreSource:  score.ScoreSource,
			ScoreVersion: score.ScoreVersion,
		}
		if strings.TrimSpace(score.ScoreEvidenceJSON) != "" && score.ScoreEvidenceJSON != "{}" {
			var evidence map[string]any
			if err := json.Unmarshal([]byte(score.ScoreEvidenceJSON), &evidence); err == nil && len(evidence) > 0 {
				item.ScoreEvidence = evidence
			}
		}
		out = append(out, item)
	}
	return out
}

func toObjectiveScoreResponses(scores []domain.ObjectiveScore) []objectiveScoreResponse {
	if len(scores) == 0 {
		return nil
	}
	out := make([]objectiveScoreResponse, 0, len(scores))
	for _, score := range scores {
		item := objectiveScoreResponse{
			TGOCode:      score.TGOCode,
			TGOTitle:     tgoTitleForCode(score.TGOCode),
			Score:        score.Score,
			ScoreSource:  score.ScoreSource,
			ScoreVersion: score.ScoreVersion,
		}
		if strings.TrimSpace(score.ScoreEvidenceJSON) != "" && score.ScoreEvidenceJSON != "{}" {
			var evidence map[string]any
			if err := json.Unmarshal([]byte(score.ScoreEvidenceJSON), &evidence); err == nil && len(evidence) > 0 {
				item.ScoreEvidence = evidence
			}
		}
		out = append(out, item)
	}
	return out
}

func focusedSkillsFromAssessments(assessments []domain.TGOAssessment) map[string]bool {
	if len(assessments) == 0 {
		return nil
	}
	out := map[string]bool{}
	for _, assessment := range assessments {
		skill := strings.TrimSpace(domain.TGOCodeToSkill[assessment.TGOCode])
		if skill == "" {
			continue
		}
		out[strings.ToLower(skill)] = true
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func toReviewJobResponse(job domain.ReviewJob) reviewJobResponse {
	return reviewJobResponse{
		ID:           job.ID,
		SubmissionID: job.SubmissionID,
		ReviewID:     job.ReviewID,
		Status:       job.Status,
		AttemptCount: job.AttemptCount,
		MaxAttempts:  job.MaxAttempts,
		LastError:    job.LastError,
		CreatedAt:    db.Since(job.CreatedAt),
		UpdatedAt:    db.Since(job.UpdatedAt),
	}
}

func toReviewJobResponseFromAIJob(job domain.AIJob) reviewJobResponse {
	return reviewJobResponse{
		ID:           job.ID,
		SubmissionID: job.SubmissionID,
		ReviewID:     job.ReviewID,
		Status:       job.Status,
		AttemptCount: job.AttemptCount,
		MaxAttempts:  job.MaxAttempts,
		LastError:    job.LastError,
		CreatedAt:    db.Since(job.CreatedAt),
		UpdatedAt:    db.Since(job.UpdatedAt),
	}
}

func (s Server) toAIJobResponse(ctx context.Context, job domain.AIJob) aiJobResponse {
	out := aiJobResponse{
		ID:           job.ID,
		Kind:         job.Kind,
		SubmissionID: job.SubmissionID,
		ExerciseID:   job.ExerciseID,
		ReviewID:     job.ReviewID,
		Status:       job.Status,
		AttemptCount: job.AttemptCount,
		MaxAttempts:  job.MaxAttempts,
		LastError:    job.LastError,
		CreatedAt:    db.Since(job.CreatedAt),
		UpdatedAt:    db.Since(job.UpdatedAt),
	}
	var result aiJobResultResponse
	if job.ExerciseID != 0 {
		if exercise, err := s.Store.GetExercise(ctx, job.ExerciseID); err == nil {
			exerciseResponse := toExerciseResponse(exercise)
			result.Exercise = &exerciseResponse
		}
	}
	if job.ReviewID != 0 {
		if reviewResult, err := s.Store.GetReview(ctx, job.ReviewID); err == nil {
			response := toReviewResponse(reviewResult)
			if artifacts, err := s.Store.GetReviewArtifacts(ctx, reviewResult.ID); err == nil {
				response.Artifacts = decodeReviewArtifacts(artifacts)
				if len(response.Annotations) == 0 {
					response.Annotations = append(response.Annotations, response.Artifacts.Annotations...)
				}
			}
			result.Review = &response
		}
	} else if strings.TrimSpace(job.ResultJSON) != "" {
		_ = json.Unmarshal([]byte(job.ResultJSON), &result)
	}
	if result.Exercise != nil || result.Review != nil {
		out.Result = &result
	}
	return out
}

func tgoTitleForCode(code string) string {
	if tgo, ok := domain.TGOByCode(code); ok {
		return tgo.Title
	}
	return strings.ReplaceAll(code, "-", " ")
}
