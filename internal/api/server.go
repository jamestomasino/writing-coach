package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"time"

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
}

func (s *Server) Serve(ctx context.Context) error {
	s.startBackgroundWorkers(ctx)

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

func (s *Server) routes() http.Handler {
	if s.validationLimiter == nil {
		s.validationLimiter = newAIValidationLimiter(s.Config.AIValidateLimitPerMinute, s.Config.AIValidateGlobalLimitPerMinute)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", s.handleHealth)
	mux.HandleFunc("GET /api/ready", s.handleReady)
	mux.HandleFunc("GET /api/auth/session", s.handleAuthSession)
	mux.HandleFunc("GET /api/ai/settings", s.handleAISettingsGet)
	mux.HandleFunc("PUT /api/ai/settings", s.handleAISettingsUpsert)
	mux.HandleFunc("DELETE /api/ai/settings", s.handleAISettingsDelete)
	mux.HandleFunc("POST /api/ai/settings/validate", s.handleAISettingsValidate)
	mux.HandleFunc("POST /api/account/reset", s.handleAccountReset)
	mux.HandleFunc("GET /api/skill-graph", s.handleSkillGraph)
	mux.HandleFunc("GET /api/onboarding/options", s.handleOnboardingOptions)
	mux.HandleFunc("GET /api/onboarding", s.handleOnboardingGet)
	mux.HandleFunc("POST /api/onboarding", s.handleOnboardingUpsert)
	mux.HandleFunc("GET /api/admins", s.handleAdminsList)
	mux.HandleFunc("POST /api/admins", s.handleAdminsCreate)
	mux.HandleFunc("DELETE /api/admins/{email}", s.handleAdminsDelete)
	mux.HandleFunc("GET /api/admin/ai-provider-events", s.handleAdminAIProviderEvents)
	mux.HandleFunc("GET /api/users", s.handleUsersList)
	mux.HandleFunc("POST /api/users", s.handleUsersCreate)
	mux.HandleFunc("GET /api/users/{slug}", s.handleUserGet)
	mux.HandleFunc("GET /api/trees", s.handleTreesList)
	mux.HandleFunc("POST /api/trees", s.handleTreeCreate)
	mux.HandleFunc("GET /api/trees/{slug}/versions", s.handleTreeVersionsList)
	mux.HandleFunc("GET /api/trees/{slug}/versions/{version}", s.handleTreeVersionGet)
	mux.HandleFunc("GET /api/trees/{slug}/diff", s.handleTreeDiff)
	mux.HandleFunc("POST /api/trees/{slug}/versions/{version}/restore", s.handleTreeVersionRestore)
	mux.HandleFunc("GET /api/trees/{slug}", s.handleTreeGet)
	mux.HandleFunc("PUT /api/trees/{slug}", s.handleTreeUpdate)
	mux.HandleFunc("GET /api/enrollments", s.handleEnrollmentsList)
	mux.HandleFunc("POST /api/enrollments", s.handleEnrollmentsCreate)
	mux.HandleFunc("GET /api/enrollments/{id}/board", s.handleEnrollmentBoard)
	mux.HandleFunc("GET /api/tracks", s.handleTracksList)
	mux.HandleFunc("PUT /api/tracks/active", s.handleTracksActiveUpdate)
	mux.HandleFunc("POST /api/tracks/{slug}/archive", s.handleTracksArchive)
	mux.HandleFunc("GET /api/context", s.handleContext)
	mux.HandleFunc("GET /api/dashboard", s.handleDashboard)
	mux.HandleFunc("GET /api/assignments", s.handleAssignmentsList)
	mux.HandleFunc("GET /api/assignments/{id}", s.handleAssignmentGet)
	mux.HandleFunc("GET /api/exercises", s.handleExercisesList)
	mux.HandleFunc("GET /api/exercises/{id}", s.handleExerciseGet)
	mux.HandleFunc("POST /api/prompts/next", s.handlePromptNext)
	mux.HandleFunc("POST /api/prompts/accept", s.handlePromptAccept)
	mux.HandleFunc("POST /api/prompts/revise", s.handlePromptRevise)
	mux.HandleFunc("GET /api/submissions", s.handleSubmissionsList)
	mux.HandleFunc("POST /api/submissions", s.handleSubmissionCreate)
	mux.HandleFunc("GET /api/submissions/{id}", s.handleSubmissionGet)
	mux.HandleFunc("GET /api/review-jobs", s.handleReviewJobGet)
	mux.HandleFunc("GET /api/reviews", s.handleReviewsList)
	mux.HandleFunc("POST /api/reviews", s.handleReviewCreate)
	mux.HandleFunc("GET /api/reviews/{id}", s.handleReviewGet)
	mux.HandleFunc("GET /api/compare", s.handleCompare)
	return withServerLogging(withRecovery(withCORS(withAuth(mux, s.Config.APIToken, s.Config.KratosPublicURL))))
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
	CategoryCounts      []aiProviderEventCountResponse `json:"category_counts"`
}

type aiProviderEventFiltersResponse struct {
	Hours     int      `json:"hours"`
	Provider  string   `json:"provider,omitempty"`
	Event     string   `json:"event,omitempty"`
	Providers []string `json:"providers,omitempty"`
	Events    []string `json:"events,omitempty"`
}

type userResponse struct {
	ID             int64  `json:"id"`
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	ActiveTreeSlug string `json:"active_tree_slug,omitempty"`
	CreatedAt      string `json:"created_at"`
}

type treeResponse struct {
	ID             int64         `json:"id"`
	Slug           string        `json:"slug"`
	Title          string        `json:"title"`
	Description    string        `json:"description"`
	SeedCodes      []string      `json:"seed_codes,omitempty"`
	PrioritySkills []string      `json:"priority_skills,omitempty"`
	TGOs           []tgoResponse `json:"tgos,omitempty"`
	CreatedAt      string        `json:"created_at,omitempty"`
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

type treeVersionResponse struct {
	ID          int64  `json:"id"`
	TreeID      int64  `json:"tree_id"`
	TreeSlug    string `json:"tree_slug"`
	Version     int    `json:"version"`
	Title       string `json:"title"`
	Description string `json:"description"`
	CreatedAt   string `json:"created_at"`
}

type enrollmentResponse struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	TreeID    int64  `json:"tree_id"`
	UserSlug  string `json:"user_slug"`
	TreeSlug  string `json:"tree_slug"`
	CreatedAt string `json:"created_at"`
}

type curriculumStateResponse struct {
	ID              int64  `json:"id"`
	CurrentFocus    string `json:"current_focus"`
	DifficultyLevel int    `json:"difficulty_level"`
	LastReviewID    int64  `json:"last_review_id"`
	UpdatedAt       string `json:"updated_at"`
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
	LatestStepID      string                   `json:"latest_step_id,omitempty"`
	Steps             []assignmentStepResponse `json:"steps"`
}

type assignmentListItemResponse struct {
	RootExerciseID    int64    `json:"root_exercise_id"`
	CurrentExerciseID int64    `json:"current_exercise_id"`
	Title             string   `json:"title"`
	LatestActivity    string   `json:"latest_activity"`
	LatestStepLabel   string   `json:"latest_step_label"`
	ExerciseCount     int      `json:"exercise_count"`
	DraftCount        int      `json:"draft_count"`
	ReviewCount       int      `json:"review_count"`
	RevisionCount     int      `json:"revision_count"`
	TGOs              []string `json:"tgos,omitempty"`
	IsCurrent         bool     `json:"is_current,omitempty"`
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
	Stage          string   `json:"stage"`
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
	Skill string `json:"skill"`
	Score int    `json:"score"`
}

type tgoAssessmentResponse struct {
	TGOCode  string `json:"tgo_code"`
	TGOTitle string `json:"tgo_title,omitempty"`
	Status   string `json:"status"`
	Evidence string `json:"evidence"`
}

type reviewResponse struct {
	ID                 int64                   `json:"id"`
	SubmissionID       int64                   `json:"submission_id"`
	ReviewKind         string                  `json:"review_kind"`
	ProviderNote       string                  `json:"provider_note,omitempty"`
	Summary            string                  `json:"summary"`
	Strengths          []string                `json:"strengths"`
	Weaknesses         []string                `json:"weaknesses"`
	AnalyzerFindings   []string                `json:"analyzer_findings"`
	NextFocus          string                  `json:"next_focus"`
	MetricWordCount    int                     `json:"metric_word_count"`
	SkillScores        []scoreResponse         `json:"skill_scores"`
	TGOAssessments     []tgoAssessmentResponse `json:"tgo_assessments"`
	CompletedTGOChecks []tgoAssessmentResponse `json:"completed_tgo_checks"`
	Annotations        []annotationResponse    `json:"annotations,omitempty"`
	Artifacts          *reviewArtifactsPayload `json:"artifacts,omitempty"`
}

type comparisonResponse struct {
	Summary              string   `json:"summary"`
	WordDelta            int      `json:"word_delta"`
	AddedWords           []string `json:"added_words"`
	RemovedWords         []string `json:"removed_words"`
	AddressedWeaknesses  []string `json:"addressed_weaknesses"`
	PersistingWeaknesses []string `json:"persisting_weaknesses"`
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
				Stage:         node.Stage,
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

func (s Server) handleUsersList(w http.ResponseWriter, r *http.Request) {
	users, err := s.Store.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": toUserResponses(users)})
}

func (s Server) handleUsersCreate(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
		Tree string `json:"tree"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
		return
	}
	if strings.TrimSpace(payload.Slug) == "" || strings.TrimSpace(payload.Name) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("slug and name are required"))
		return
	}
	treeSlug := firstNonEmpty(payload.Tree, s.Config.DefaultTreeSlug)
	userID, treeID, enrollmentID, err := s.Store.EnsureDefaultUserTree(r.Context(), payload.Slug, payload.Name, treeSlug)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"user": map[string]any{
			"id":   userID,
			"slug": payload.Slug,
			"name": payload.Name,
		},
		"enrollment": map[string]any{
			"id":        enrollmentID,
			"user_id":   userID,
			"tree_id":   treeID,
			"user_slug": payload.Slug,
			"tree_slug": treeSlug,
		},
	})
}

func (s Server) handleUserGet(w http.ResponseWriter, r *http.Request) {
	user, err := s.Store.UserBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": toUserResponses([]domain.User{user})[0]})
}

func (s Server) handleTreesList(w http.ResponseWriter, r *http.Request) {
	trees, err := s.Store.ListTrees(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	includeTGOs := r.URL.Query().Get("include_tgos") == "1"
	writeJSON(w, http.StatusOK, map[string]any{"trees": s.toTreeResponses(r.Context(), trees, includeTGOs)})
}

func (s Server) handleTreeCreate(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	var payload struct {
		Slug           string        `json:"slug"`
		Title          string        `json:"title"`
		Description    string        `json:"description"`
		SeedCodes      []string      `json:"seed_codes"`
		PrioritySkills []string      `json:"priority_skills"`
		TGOs           []tgoResponse `json:"tgos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
		return
	}
	if strings.TrimSpace(payload.Slug) == "" || strings.TrimSpace(payload.Title) == "" || len(payload.TGOs) == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("slug, title, and at least one TGO are required"))
		return
	}
	def := domain.TGOTreeDefinition{
		Slug:           payload.Slug,
		Title:          payload.Title,
		Description:    payload.Description,
		SeedCodes:      append([]string(nil), payload.SeedCodes...),
		PrioritySkills: append([]string(nil), payload.PrioritySkills...),
	}
	for _, item := range payload.TGOs {
		def.TGOs = append(def.TGOs, domain.TGO{
			Code:          item.Code,
			Title:         item.Title,
			Description:   item.Description,
			Stage:         item.Stage,
			StageOrder:    item.StageOrder,
			Prerequisites: append([]string(nil), item.Prerequisites...),
			MasteryHint:   item.MasteryHint,
			ProgressMode:  item.ProgressMode,
		})
	}
	if len(def.SeedCodes) == 0 && len(def.TGOs) >= 3 {
		def.SeedCodes = []string{def.TGOs[0].Code, def.TGOs[1].Code, def.TGOs[2].Code}
	}
	if err := s.Store.SaveTreeDefinition(r.Context(), def); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tree, err := s.Store.TreeBySlug(r.Context(), def.Slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"tree": s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)[0]})
}

func (s Server) handleTreeGet(w http.ResponseWriter, r *http.Request) {
	tree, err := s.Store.TreeBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	response := s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)[0]
	if appContext, err := s.resolveSession(r.Context(), r); err == nil {
		response = s.applyGeneratedTreeProfileDisplay(r.Context(), appContext, tree, response)
	}
	writeJSON(w, http.StatusOK, map[string]any{"tree": response})
}

func (s Server) handleTreeUpdate(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	var payload struct {
		Title          string        `json:"title"`
		Description    string        `json:"description"`
		SeedCodes      []string      `json:"seed_codes"`
		PrioritySkills []string      `json:"priority_skills"`
		TGOs           []tgoResponse `json:"tgos"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
		return
	}
	slug := r.PathValue("slug")
	if strings.TrimSpace(payload.Title) == "" || len(payload.TGOs) == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("title and at least one TGO are required"))
		return
	}
	def := domain.TGOTreeDefinition{
		Slug:           slug,
		Title:          payload.Title,
		Description:    payload.Description,
		SeedCodes:      append([]string(nil), payload.SeedCodes...),
		PrioritySkills: append([]string(nil), payload.PrioritySkills...),
	}
	for _, item := range payload.TGOs {
		def.TGOs = append(def.TGOs, domain.TGO{
			Code:          item.Code,
			Title:         item.Title,
			Description:   item.Description,
			Stage:         item.Stage,
			StageOrder:    item.StageOrder,
			Prerequisites: append([]string(nil), item.Prerequisites...),
			MasteryHint:   item.MasteryHint,
			ProgressMode:  item.ProgressMode,
		})
	}
	if len(def.SeedCodes) == 0 && len(def.TGOs) >= 3 {
		def.SeedCodes = []string{def.TGOs[0].Code, def.TGOs[1].Code, def.TGOs[2].Code}
	}
	if err := s.Store.SaveTreeDefinition(r.Context(), def); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tree, err := s.Store.TreeBySlug(r.Context(), slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tree": s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)[0]})
}

func (s Server) handleTreeVersionsList(w http.ResponseWriter, r *http.Request) {
	versions, err := s.Store.ListTreeVersions(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var items []treeVersionResponse
	for _, version := range versions {
		items = append(items, treeVersionResponse{
			ID:          version.ID,
			TreeID:      version.TreeID,
			TreeSlug:    version.TreeSlug,
			Version:     version.Version,
			Title:       version.Title,
			Description: version.Description,
			CreatedAt:   db.Since(version.CreatedAt),
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"versions": items})
}

func (s Server) handleTreeVersionGet(w http.ResponseWriter, r *http.Request) {
	version, err := strconv.Atoi(r.PathValue("version"))
	if err != nil || version <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid version"))
		return
	}
	meta, def, err := s.Store.TreeVersionByNumber(r.Context(), r.PathValue("slug"), version)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version": treeVersionResponse{
			ID:          meta.ID,
			TreeID:      meta.TreeID,
			TreeSlug:    meta.TreeSlug,
			Version:     meta.Version,
			Title:       meta.Title,
			Description: meta.Description,
			CreatedAt:   db.Since(meta.CreatedAt),
		},
		"tree": treeResponse{
			Slug:           def.Slug,
			Title:          def.Title,
			Description:    def.Description,
			SeedCodes:      append([]string(nil), def.SeedCodes...),
			PrioritySkills: append([]string(nil), def.PrioritySkills...),
			TGOs:           toTGOResponses(def.TGOs),
		},
	})
}

func (s Server) handleTreeDiff(w http.ResponseWriter, r *http.Request) {
	fromVersion, err := strconv.Atoi(r.URL.Query().Get("from"))
	if err != nil || fromVersion <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid from version"))
		return
	}
	toVersion, err := strconv.Atoi(r.URL.Query().Get("to"))
	if err != nil || toVersion <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid to version"))
		return
	}
	_, fromDef, err := s.Store.TreeVersionByNumber(r.Context(), r.PathValue("slug"), fromVersion)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	_, toDef, err := s.Store.TreeVersionByNumber(r.Context(), r.PathValue("slug"), toVersion)
	if err != nil {
		writeError(w, http.StatusNotFound, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"from_version": fromVersion,
		"to_version":   toVersion,
		"diff":         treeDefinitionDiff(fromDef, toDef),
	})
}

func (s Server) handleTreeVersionRestore(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	version, err := strconv.Atoi(r.PathValue("version"))
	if err != nil || version <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid version"))
		return
	}
	_, def, err := s.Store.TreeVersionByNumber(r.Context(), r.PathValue("slug"), version)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if err := s.Store.SaveTreeDefinition(r.Context(), def); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tree, err := s.Store.TreeBySlug(r.Context(), def.Slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"restored_version": version,
		"tree":             s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)[0],
	})
}

func (s Server) handleEnrollmentsList(w http.ResponseWriter, r *http.Request) {
	enrollments, err := s.Store.ListEnrollments(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"enrollments": toEnrollmentResponses(enrollments)})
}

func (s Server) handleEnrollmentsCreate(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UserSlug string `json:"user_slug"`
		UserName string `json:"user_name"`
		TreeSlug string `json:"tree_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
		return
	}
	if strings.TrimSpace(payload.UserSlug) == "" || strings.TrimSpace(payload.TreeSlug) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("user_slug and tree_slug are required"))
		return
	}
	userName := firstNonEmpty(payload.UserName, payload.UserSlug)
	userID, treeID, enrollmentID, err := s.Store.EnsureDefaultUserTree(r.Context(), payload.UserSlug, userName, payload.TreeSlug)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"enrollment": map[string]any{
			"id":        enrollmentID,
			"user_id":   userID,
			"tree_id":   treeID,
			"user_slug": payload.UserSlug,
			"tree_slug": payload.TreeSlug,
		},
	})
}

func (s Server) handleEnrollmentBoard(w http.ResponseWriter, r *http.Request) {
	enrollmentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || enrollmentID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid enrollment id"))
		return
	}
	enrollment, err := s.Store.EnrollmentByID(r.Context(), enrollmentID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	s.writeEnrollmentBoard(r.Context(), w, session.Context{
		UserID:       enrollment.UserID,
		TreeID:       enrollment.TreeID,
		EnrollmentID: enrollment.ID,
		UserSlug:     enrollment.UserSlug,
		TreeSlug:     enrollment.TreeSlug,
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
		if strings.TrimSpace(s.Config.APIToken) == "" && strings.TrimSpace(s.Config.KratosPublicURL) == "" {
			return true
		}
	}
	writeJSON(w, http.StatusForbidden, errorResponse{Error: "admin access required"})
	return false
}

func (s Server) reviewComparisonPayload(ctx context.Context, sub domain.Submission, currentReview domain.Review) map[string]any {
	previous, err := s.Store.PreviousSubmission(ctx, sub)
	if err != nil {
		return nil
	}
	previousReview, err := s.Store.LatestReviewForSubmission(ctx, previous.ID)
	if err != nil {
		return nil
	}
	comparison := review.CompareSubmissions(sub, previous, currentReview, previousReview)
	return map[string]any{
		"summary":               comparison.Summary,
		"word_delta":            comparison.WordDelta,
		"added_words":           comparison.AddedWords,
		"removed_words":         comparison.RemovedWords,
		"addressed_weaknesses":  comparison.AddressedWeaknesses,
		"persisting_weaknesses": comparison.PersistingWeaknesses,
	}
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
	runtime, err := s.resolveLLMRuntime(ctx, job.UserID)
	if err != nil {
		return fmt.Errorf("resolve provider: %w", err)
	}

	reviewResult := s.Reviews.WithClient(runtime.Client, runtime.ProviderKind).ReviewSubmissionDetailed(ctx, sub, activeTGOs, completedTGOs)
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
	recommendation, err := s.Curriculum.SyncTGOs(ctx, s.Store, treeSlug, job.EnrollmentID, reviewResult.Review)
	if err != nil {
		return fmt.Errorf("sync curriculum: %w", err)
	}
	reviewResult.Review.NextFocus = recommendation.Focus
	reviewID, err := s.Store.SaveReview(ctx, reviewResult.Review, reviewResult.Scores)
	if err != nil {
		return fmt.Errorf("save review: %w", err)
	}
	if err := s.Store.SaveReviewArtifacts(ctx, domain.ReviewArtifacts{
		ReviewID:           reviewID,
		AnalyzerReportJSON: mustJSON(reviewResult.AnalyzerReport),
		RecommendationJSON: mustJSON(recommendation),
		ComparisonJSON:     mustJSON(s.reviewComparisonPayload(ctx, sub, reviewResult.Review)),
		AnnotationsJSON:    mustJSON(reviewResult.Review.Annotations),
	}); err != nil {
		return fmt.Errorf("save review artifacts: %w", err)
	}
	if err := s.Store.UpdateCurriculumState(ctx, job.EnrollmentID, recommendation.Focus, recommendation.Difficulty, reviewID); err != nil {
		return fmt.Errorf("update curriculum state: %w", err)
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

func treeDefinitionDiff(fromDef, toDef domain.TGOTreeDefinition) map[string]any {
	fromMap := map[string]domain.TGO{}
	toMap := map[string]domain.TGO{}
	for _, tgo := range fromDef.TGOs {
		fromMap[tgo.Code] = tgo
	}
	for _, tgo := range toDef.TGOs {
		toMap[tgo.Code] = tgo
	}
	var added []string
	var removed []string
	var changed []string
	for code, tgo := range toMap {
		if _, ok := fromMap[code]; !ok {
			added = append(added, code)
			continue
		}
		if fromMap[code].Title != tgo.Title || fromMap[code].Description != tgo.Description || fromMap[code].Stage != tgo.Stage || fromMap[code].StageOrder != tgo.StageOrder || strings.Join(fromMap[code].Prerequisites, ",") != strings.Join(tgo.Prerequisites, ",") || fromMap[code].MasteryHint != tgo.MasteryHint {
			changed = append(changed, code)
		}
	}
	for code := range fromMap {
		if _, ok := toMap[code]; !ok {
			removed = append(removed, code)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	sort.Strings(changed)
	return map[string]any{
		"title_changed":           fromDef.Title != toDef.Title,
		"description_changed":     fromDef.Description != toDef.Description,
		"seed_codes_changed":      strings.Join(fromDef.SeedCodes, ",") != strings.Join(toDef.SeedCodes, ","),
		"priority_skills_changed": strings.Join(fromDef.PrioritySkills, ",") != strings.Join(toDef.PrioritySkills, ","),
		"added_tgos":              added,
		"removed_tgos":            removed,
		"changed_tgos":            changed,
	}
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
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-API-Token, X-Session-Token, X-Writing-Coach-User, X-Writing-Coach-Tree")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(b)
}

func withServerLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		if recorder.status >= 500 {
			log.Printf("api %s %s -> %d", r.Method, r.URL.Path, recorder.status)
		}
	})
}

func withRecovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("api panic: %s %s: %v\n%s", r.Method, r.URL.Path, recovered, debug.Stack())
				writeError(w, http.StatusInternalServerError, fmt.Errorf("internal server error"))
			}
		}()
		next.ServeHTTP(w, r)
	})
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

func toUserResponses(users []domain.User) []userResponse {
	var out []userResponse
	for _, user := range users {
		out = append(out, userResponse{
			ID:             user.ID,
			Slug:           user.Slug,
			Name:           user.Name,
			ActiveTreeSlug: user.ActiveTreeSlug,
			CreatedAt:      db.Since(user.CreatedAt),
		})
	}
	return out
}

func toOnboardingProfileResponse(profile domain.OnboardingProfile) *onboardingProfileResponse {
	return &onboardingProfileResponse{
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

func (s Server) toTreeResponses(ctx context.Context, trees []domain.TGOTree, includeTGOs bool) []treeResponse {
	var out []treeResponse
	for _, tree := range trees {
		item := treeResponse{
			ID:          tree.ID,
			Slug:        tree.Slug,
			Title:       tree.Title,
			Description: tree.Description,
			CreatedAt:   db.Since(tree.CreatedAt),
		}
		if includeTGOs {
			if def, err := s.Store.TreeDefinitionBySlug(ctx, tree.Slug); err == nil {
				item.SeedCodes = append([]string(nil), def.SeedCodes...)
				item.PrioritySkills = append([]string(nil), def.PrioritySkills...)
				item.TGOs = toTGOResponses(def.TGOs)
			}
		}
		out = append(out, item)
	}
	return out
}

func (s Server) applyGeneratedTreeProfileDisplay(ctx context.Context, appContext session.Context, tree domain.TGOTree, response treeResponse) treeResponse {
	profile, err := s.Store.OnboardingProfileByEnrollmentID(ctx, appContext.EnrollmentID)
	if err != nil || profile.GeneratedTreeSlug != tree.Slug {
		return response
	}
	user, err := s.Store.UserBySlug(ctx, appContext.UserSlug)
	if err != nil {
		return response
	}
	response.Title, response.Description = domain.GeneratedTreeDisplay(user.Name, profile, tree.Title, tree.Description)
	response.PrioritySkills = nil
	return response
}

func toEnrollmentResponses(enrollments []domain.Enrollment) []enrollmentResponse {
	var out []enrollmentResponse
	for _, enrollment := range enrollments {
		out = append(out, enrollmentResponse{
			ID:        enrollment.ID,
			UserID:    enrollment.UserID,
			TreeID:    enrollment.TreeID,
			UserSlug:  enrollment.UserSlug,
			TreeSlug:  enrollment.TreeSlug,
			CreatedAt: db.Since(enrollment.CreatedAt),
		})
	}
	return out
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
			ID:             tgo.ID,
			Code:           tgo.Code,
			Title:          tgo.Title,
			Description:    tgo.Description,
			Stage:          tgo.Stage,
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
	for _, exercise := range chain {
		exerciseCopy := exercise
		steps = append(steps, assignmentStepResponse{
			ID:         fmt.Sprintf("exercise-%d", exercise.ID),
			Kind:       "exercise",
			Title:      exerciseTitle(exercise),
			Label:      stepLabel("exercise", exercise, domain.Submission{}),
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
			steps = append(steps, assignmentStepResponse{
				ID:           fmt.Sprintf("submission-%d", submission.ID),
				Kind:         "submission",
				Title:        fmt.Sprintf("Draft %d", submission.DraftNumber),
				Label:        stepLabel("submission", exercise, submission),
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
			steps = append(steps, assignmentStepResponse{
				ID:           fmt.Sprintf("review-%d", reviewResult.ID),
				Kind:         "review",
				Title:        fmt.Sprintf("Feedback %d", submission.DraftNumber),
				Label:        stepLabel("review", exercise, submission),
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
	latestExercise := exercises[0]
	for _, exercise := range exercises[1:] {
		if exercise.CreatedAt.After(latestExercise.CreatedAt) || (exercise.CreatedAt.Equal(latestExercise.CreatedAt) && exercise.ID > latestExercise.ID) {
			latestExercise = exercise
		}
	}
	currentRootID := rootExerciseID(latestExercise, exerciseByID, submissionByID)
	latestStepID := ""
	if len(steps) > 0 {
		latestStepID = steps[len(steps)-1].ID
	}
	return assignmentResponse{
		RootExerciseID:    rootID,
		CurrentExerciseID: current.ID,
		Title:             title,
		IsCurrent:         rootID == currentRootID,
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
	}
	chains := map[int64]*chainSummary{}
	latestExercise := exercises[0]
	for _, exercise := range exercises {
		if exercise.CreatedAt.After(latestExercise.CreatedAt) || (exercise.CreatedAt.Equal(latestExercise.CreatedAt) && exercise.ID > latestExercise.ID) {
			latestExercise = exercise
		}
		rootID := rootExerciseID(exercise, exerciseByID, submissionByID)
		chain := chains[rootID]
		if chain == nil {
			chain = &chainSummary{
				rootExerciseID:  rootID,
				title:           exerciseTitle(exerciseByID[rootID]),
				latestActivity:  exercise.CreatedAt,
				latestStepLabel: stepLabel("exercise", exercise, domain.Submission{}),
				tgos:            titlesForTGOCodes(exerciseByID[rootID].TGOCodes),
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

	currentRootID := rootExerciseID(latestExercise, exerciseByID, submissionByID)
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

func stepLabel(kind string, exercise domain.Exercise, submission domain.Submission) string {
	switch kind {
	case "exercise":
		if exercise.SourceSubmissionID != 0 {
			return "Revision brief"
		}
		return "Prompt"
	case "submission":
		return fmt.Sprintf("Draft %d", submission.DraftNumber)
	case "review":
		return fmt.Sprintf("Feedback %d", submission.DraftNumber)
	default:
		return ""
	}
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
		SkillScores:      toScoreResponses(reviewResult.SkillScores),
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

func toScoreResponses(scores []domain.SkillScore) []scoreResponse {
	var out []scoreResponse
	for _, score := range scores {
		out = append(out, scoreResponse{Skill: score.Skill, Score: score.Score})
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

func tgoTitleForCode(code string) string {
	if tgo, ok := domain.TGOByCode(code); ok {
		return tgo.Title
	}
	return strings.ReplaceAll(code, "-", " ")
}
