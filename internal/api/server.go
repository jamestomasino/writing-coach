package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
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
	mux.HandleFunc("GET /api/ready", s.handleReady)
	mux.HandleFunc("GET /api/auth/session", s.handleAuthSession)
	mux.HandleFunc("GET /api/skill-graph", s.handleSkillGraph)
	mux.HandleFunc("GET /api/onboarding", s.handleOnboardingGet)
	mux.HandleFunc("POST /api/onboarding", s.handleOnboardingUpsert)
	mux.HandleFunc("GET /api/admins", s.handleAdminsList)
	mux.HandleFunc("POST /api/admins", s.handleAdminsCreate)
	mux.HandleFunc("DELETE /api/admins/{email}", s.handleAdminsDelete)
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
	mux.HandleFunc("GET /api/context", s.handleContext)
	mux.HandleFunc("GET /api/dashboard", s.handleDashboard)
	mux.HandleFunc("GET /api/exercises", s.handleExercisesList)
	mux.HandleFunc("GET /api/exercises/{id}", s.handleExerciseGet)
	mux.HandleFunc("POST /api/prompts/next", s.handlePromptNext)
	mux.HandleFunc("POST /api/prompts/revise", s.handlePromptRevise)
	mux.HandleFunc("GET /api/submissions", s.handleSubmissionsList)
	mux.HandleFunc("POST /api/submissions", s.handleSubmissionCreate)
	mux.HandleFunc("GET /api/submissions/{id}", s.handleSubmissionGet)
	mux.HandleFunc("GET /api/reviews", s.handleReviewsList)
	mux.HandleFunc("POST /api/reviews", s.handleReviewCreate)
	mux.HandleFunc("GET /api/reviews/{id}", s.handleReviewGet)
	mux.HandleFunc("GET /api/compare", s.handleCompare)
	return withCORS(withAuth(mux, s.Config.APIToken, s.Config.KratosPublicURL))
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

type authSessionResponse struct {
	Authenticated      bool                    `json:"authenticated"`
	AuthMode           string                  `json:"auth_mode"`
	Identity           *authIdentityResponse   `json:"identity,omitempty"`
	Context            *requestContextResponse `json:"context,omitempty"`
	OnboardingComplete bool                    `json:"onboarding_complete"`
	ActiveTreeSlug     string                  `json:"active_tree_slug,omitempty"`
}

type authIdentityResponse struct {
	Subject string `json:"subject"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
}

type userResponse struct {
	ID             int64  `json:"id"`
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	ActiveTreeSlug string `json:"active_tree_slug,omitempty"`
	CreatedAt      string `json:"created_at"`
}

type onboardingResponse struct {
	Profile            *onboardingProfileResponse `json:"profile,omitempty"`
	OnboardingComplete bool                       `json:"onboarding_complete"`
	Tree               *treeResponse              `json:"tree,omitempty"`
	StarterTGOCodes    []string                   `json:"starter_tgo_codes,omitempty"`
	RecommendedRegions []string                   `json:"recommended_regions,omitempty"`
	Context            *requestContextResponse    `json:"context,omitempty"`
}

type onboardingProfileResponse struct {
	WritingType         string   `json:"writing_type"`
	ExperienceLevel     string   `json:"experience_level"`
	DesiredTone         string   `json:"desired_tone"`
	BiggestWeaknesses   []string `json:"biggest_weaknesses"`
	DesiredOutcomes     []string `json:"desired_outcomes"`
	DifficultyIntensity string   `json:"difficulty_intensity"`
	WritingGoals        string   `json:"writing_goals"`
	GeneratedTreeSlug   string   `json:"generated_tree_slug"`
	TemplateKey         string   `json:"template_key"`
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
	Category string `json:"category"`
	Comment  string `json:"comment"`
	Severity string `json:"severity"`
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

func (s Server) handleAuthSession(w http.ResponseWriter, r *http.Request) {
	mode := authModeFromContext(r.Context())
	resp := authSessionResponse{
		Authenticated: mode != "none",
		AuthMode:      mode,
	}
	if ident, ok := identityFromContext(r.Context()); ok {
		resp.Identity = &authIdentityResponse{
			Subject: ident.Subject,
			Email:   ident.Email,
			Name:    ident.Name,
		}
	}
	if appContext, err := s.resolveSession(r.Context(), r); err == nil {
		resp.Context = &requestContextResponse{
			UserSlug: appContext.UserSlug,
			TreeSlug: appContext.TreeSlug,
			UserID:   appContext.UserID,
			TreeID:   appContext.TreeID,
		}
		if user, err := s.Store.UserBySlug(r.Context(), appContext.UserSlug); err == nil {
			resp.ActiveTreeSlug = user.ActiveTreeSlug
			if profile, err := s.Store.OnboardingProfileByUserID(r.Context(), user.ID); err == nil {
				resp.OnboardingComplete = profile.Complete()
			}
		}
	}
	writeJSON(w, http.StatusOK, resp)
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

func (s Server) handleOnboardingGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	profile, err := s.Store.OnboardingProfileByUserID(r.Context(), appContext.UserID)
	if err != nil {
		if db.IsNotFound(err) || errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusOK, onboardingResponse{
				OnboardingComplete: false,
				Tree:               &treeResponse{Slug: domain.GlobalSkillGraphSlug, Title: domain.GlobalSkillGraphTitle},
				Context:            &requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
			})
			return
		}
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tree, err := s.Store.TreeBySlug(r.Context(), profile.GeneratedTreeSlug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	treeResponses := s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)
	treeResponse := treeResponses[0]
	writeJSON(w, http.StatusOK, onboardingResponse{
		Profile:            toOnboardingProfileResponse(profile),
		OnboardingComplete: profile.Complete(),
		Tree:               &treeResponse,
		StarterTGOCodes:    domain.RecommendedStarterCodes(profile),
		RecommendedRegions: domain.RecommendedRegionSlugs(profile),
		Context:            &requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
	})
}

func (s Server) handleOnboardingUpsert(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var payload struct {
		WritingType         string   `json:"writing_type"`
		ExperienceLevel     string   `json:"experience_level"`
		DesiredTone         string   `json:"desired_tone"`
		BiggestWeaknesses   []string `json:"biggest_weaknesses"`
		DesiredOutcomes     []string `json:"desired_outcomes"`
		DifficultyIntensity string   `json:"difficulty_intensity"`
		WritingGoals        string   `json:"writing_goals"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
		return
	}
	profile := domain.OnboardingProfile{
		UserID:              appContext.UserID,
		WritingType:         strings.TrimSpace(payload.WritingType),
		ExperienceLevel:     strings.TrimSpace(payload.ExperienceLevel),
		DesiredTone:         strings.TrimSpace(payload.DesiredTone),
		BiggestWeaknesses:   sanitizeStringList(payload.BiggestWeaknesses),
		DesiredOutcomes:     sanitizeStringList(payload.DesiredOutcomes),
		DifficultyIntensity: strings.TrimSpace(payload.DifficultyIntensity),
		WritingGoals:        strings.TrimSpace(payload.WritingGoals),
	}
	if !profile.Complete() {
		writeError(w, http.StatusBadRequest, fmt.Errorf("all onboarding fields are required"))
		return
	}

	user, err := s.Store.UserBySlug(r.Context(), appContext.UserSlug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	profile.TemplateKey = domain.TemplateKeyForProfile(profile)
	treeDef := domain.GlobalSkillGraphDefinition()
	profile.GeneratedTreeSlug = treeDef.Slug
	starterCodes := domain.RecommendedStarterCodes(profile)
	recommendedRegions := domain.RecommendedRegionSlugs(profile)

	if err := s.Store.SaveTreeDefinition(r.Context(), treeDef); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.Store.SaveOnboardingProfile(r.Context(), profile); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.Store.SetUserActiveTree(r.Context(), appContext.UserID, treeDef.Slug); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	_, treeID, enrollmentID, err := s.Store.EnsureDefaultUserTree(r.Context(), appContext.UserSlug, user.Name, treeDef.Slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	completed, err := s.Store.CompletedTGOs(r.Context(), enrollmentID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	activeCodes := sanitizeStringList(starterCodes)
	if len(activeCodes) != 3 {
		completedSet := make(map[string]bool, len(completed))
		for _, tgo := range completed {
			completedSet[tgo.Code] = true
		}
		activeCodes = nextSeedCodes(treeDef, completedSet)
	}
	if err := s.Store.SetActiveTGOs(r.Context(), enrollmentID, activeCodes); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.Store.UpdateCurriculumState(r.Context(), enrollmentID, activeCodes[0], 2, 0); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tree := domain.TGOTree{
		ID:          treeID,
		Slug:        treeDef.Slug,
		Title:       treeDef.Title,
		Description: treeDef.Description,
	}
	treeResponses := s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)
	treeResponse := treeResponses[0]
	writeJSON(w, http.StatusOK, onboardingResponse{
		Profile:            toOnboardingProfileResponse(profile),
		OnboardingComplete: true,
		Tree:               &treeResponse,
		StarterTGOCodes:    starterCodes,
		RecommendedRegions: recommendedRegions,
		Context:            &requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: treeDef.Slug, UserID: appContext.UserID, TreeID: treeID},
	})
}

func (s Server) handleAdminsList(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	emails, err := s.Store.ListAdminEmails(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"admins": emails})
}

func (s Server) handleAdminsCreate(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	var payload struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
		return
	}
	if err := s.Store.AddAdminEmail(r.Context(), payload.Email); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"email": strings.ToLower(strings.TrimSpace(payload.Email))})
}

func (s Server) handleAdminsDelete(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	email := r.PathValue("email")
	if err := s.Store.RemoveAdminEmail(r.Context(), email); err != nil {
		status := http.StatusBadRequest
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": strings.ToLower(strings.TrimSpace(email))})
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
	writeJSON(w, http.StatusOK, map[string]any{"tree": s.toTreeResponses(r.Context(), []domain.TGOTree{tree}, true)[0]})
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
	s.writeDashboardPayload(r.Context(), w, appContext, state, activeTGOs, completedTGOs)
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
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if r.ContentLength != 0 {
		var payload struct {
			TGOCodes []string `json:"tgo_codes"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
			return
		}
		if len(payload.TGOCodes) > 0 {
			if err := s.setActiveTGOsForSelection(r.Context(), appContext, payload.TGOCodes); err != nil {
				writeError(w, http.StatusBadRequest, err)
				return
			}
		}
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
	submissions, err := s.Store.ListSubmissions(r.Context(), appContext.UserID, appContext.TreeID, exerciseID, listLimit(r, 20, 100))
	if err != nil {
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
	if !belongsToContext(sub.UserID, sub.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("submission not found"))
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
	reviewResult := s.Reviews.ReviewSubmissionDetailed(r.Context(), sub, activeTGOs, completedTGOs)
	reviewResult.Review.UserID = appContext.UserID
	reviewResult.Review.TreeID = appContext.TreeID
	recommendation, err := s.Curriculum.SyncTGOs(r.Context(), s.Store, appContext.TreeSlug, appContext.EnrollmentID, reviewResult.Review)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	reviewResult.Review.NextFocus = recommendation.Focus
	reviewID, err := s.Store.SaveReview(r.Context(), reviewResult.Review, reviewResult.Scores)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.Store.SaveReviewArtifacts(r.Context(), domain.ReviewArtifacts{
		ReviewID:           reviewID,
		AnalyzerReportJSON: mustJSON(reviewResult.AnalyzerReport),
		RecommendationJSON: mustJSON(recommendation),
		ComparisonJSON:     mustJSON(s.reviewComparisonPayload(r.Context(), sub, reviewResult.Review)),
		AnnotationsJSON:    mustJSON(reviewResult.Review.Annotations),
	}); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.Store.UpdateCurriculumState(r.Context(), appContext.EnrollmentID, recommendation.Focus, recommendation.Difficulty, reviewID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	reviewResult.Review.ID = reviewID
	writeJSON(w, http.StatusCreated, map[string]any{
		"context":        requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"review":         toReviewResponse(reviewResult.Review),
		"skill_scores":   toScoreResponses(reviewResult.Scores),
		"recommendation": recommendation,
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
	history, err := s.Store.History(ctx, appContext.UserID, appContext.TreeID)
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

	writeJSON(w, http.StatusOK, map[string]any{
		"context":                   requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"curriculum_state":          toCurriculumStateResponse(state),
		"active_tgos":               toTGOResponses(activeTGOs),
		"completed_tgos":            toTGOResponses(completedTGOs),
		"upcoming_tgos":             toTGOResponses(upcoming),
		"progress_lines":            progress,
		"strongest_skills":          strongest,
		"weakest_skills":            weakest,
		"recurring_weaknesses":      recurringWeaknesses,
		"recurring_findings":        recurringFindings,
		"recurring_completed_slips": recurringSlips,
		"history":                   history,
	})
}

func (s Server) resolveSession(ctx context.Context, r *http.Request) (session.Context, error) {
	if ident, ok := identityFromContext(ctx); ok {
		userSlug := slugFromIdentity(ident)
		userName := displayNameFromIdentity(ident)
		if err := s.Store.EnsureUser(ctx, userSlug, userName); err != nil {
			return session.Context{}, err
		}
		treeSlug := strings.TrimSpace(firstNonEmpty(r.URL.Query().Get("tree"), r.Header.Get("X-Writing-Coach-Tree")))
		if treeSlug == "" {
			activeTreeSlug, err := s.Store.UserActiveTreeSlug(ctx, userSlug)
			if err == nil && strings.TrimSpace(activeTreeSlug) != "" {
				treeSlug = activeTreeSlug
			} else {
				treeSlug = s.Config.DefaultTreeSlug
			}
		}

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

	userSlug := firstNonEmpty(r.URL.Query().Get("user"), r.Header.Get("X-Writing-Coach-User"), s.Config.DefaultUserSlug)
	userName := firstNonEmpty(r.URL.Query().Get("user_name"), s.Config.WriterName)
	if err := s.Store.EnsureUser(ctx, userSlug, userName); err != nil {
		return session.Context{}, err
	}
	treeSlug := strings.TrimSpace(firstNonEmpty(r.URL.Query().Get("tree"), r.Header.Get("X-Writing-Coach-Tree")))
	if treeSlug == "" {
		activeTreeSlug, err := s.Store.UserActiveTreeSlug(ctx, userSlug)
		if err == nil && strings.TrimSpace(activeTreeSlug) != "" {
			treeSlug = activeTreeSlug
		} else {
			treeSlug = s.Config.DefaultTreeSlug
		}
	}

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

func decodeReviewArtifacts(artifacts domain.ReviewArtifacts) *reviewArtifactsPayload {
	payload := &reviewArtifactsPayload{}
	_ = json.Unmarshal([]byte(artifacts.AnalyzerReportJSON), &payload.AnalyzerReport)
	_ = json.Unmarshal([]byte(artifacts.RecommendationJSON), &payload.Recommendation)
	_ = json.Unmarshal([]byte(artifacts.ComparisonJSON), &payload.Comparison)
	_ = json.Unmarshal([]byte(artifacts.AnnotationsJSON), &payload.Annotations)
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
	for _, check := range reviewResult.CompletedTGOChecks {
		out.CompletedTGOChecks = append(out.CompletedTGOChecks, tgoAssessmentResponse{
			TGOCode:  check.TGOCode,
			Status:   check.Status,
			Evidence: check.Evidence,
		})
	}
	for _, annotation := range reviewResult.Annotations {
		out.Annotations = append(out.Annotations, annotationResponse{
			Quote:    annotation.Quote,
			TGOCode:  annotation.TGOCode,
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
