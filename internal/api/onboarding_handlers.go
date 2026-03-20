package api

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/session"
)

type onboardingResponse struct {
	Profile            *onboardingProfileResponse `json:"profile,omitempty"`
	OnboardingComplete bool                       `json:"onboarding_complete"`
	Tree               *treeResponse              `json:"tree,omitempty"`
	StarterTGOCodes    []string                   `json:"starter_tgo_codes,omitempty"`
	RecommendedRegions []string                   `json:"recommended_regions,omitempty"`
	Context            *requestContextResponse    `json:"context,omitempty"`
}

type onboardingOptionResponse struct {
	Value string `json:"value"`
	Label string `json:"label"`
}

type onboardingOptionsResponse struct {
	WritingDomains    []onboardingOptionResponse `json:"writing_domains"`
	AssignmentFormats []onboardingOptionResponse `json:"assignment_formats"`
	ExperienceLevels  []onboardingOptionResponse `json:"experience_levels"`
	DifficultyLevels  []onboardingOptionResponse `json:"difficulty_levels"`
	Weaknesses        []onboardingOptionResponse `json:"weaknesses"`
	DesiredOutcomes   []onboardingOptionResponse `json:"desired_outcomes"`
}

type onboardingProfileResponse struct {
	WritingType         string   `json:"writing_type"`
	AssignmentFormat    string   `json:"assignment_format"`
	TargetAudience      string   `json:"target_audience"`
	SubjectMatter       string   `json:"subject_matter"`
	ExperienceLevel     string   `json:"experience_level"`
	DesiredTone         string   `json:"desired_tone"`
	BiggestWeaknesses   []string `json:"biggest_weaknesses"`
	DesiredOutcomes     []string `json:"desired_outcomes"`
	DifficultyIntensity string   `json:"difficulty_intensity"`
	WritingGoals        string   `json:"writing_goals"`
	GeneratedTreeSlug   string   `json:"generated_tree_slug"`
	TemplateKey         string   `json:"template_key"`
}

func (s Server) handleOnboardingOptions(w http.ResponseWriter, r *http.Request) {
	options := domain.AvailableOnboardingOptions()
	writeJSON(w, http.StatusOK, onboardingOptionsResponse{
		WritingDomains:    toOnboardingOptionResponses(options.WritingDomains),
		AssignmentFormats: toOnboardingOptionResponses(options.AssignmentFormats),
		ExperienceLevels:  toOnboardingOptionResponses(options.ExperienceLevels),
		DifficultyLevels:  toOnboardingOptionResponses(options.DifficultyLevels),
		Weaknesses:        toOnboardingOptionResponses(options.Weaknesses),
		DesiredOutcomes:   toOnboardingOptionResponses(options.DesiredOutcomes),
	})
}

func (s Server) handleOnboardingGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	profile, err := s.Store.OnboardingProfileByEnrollmentID(r.Context(), appContext.EnrollmentID)
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
	treeResponse = s.applyGeneratedTreeProfileDisplay(r.Context(), appContext, tree, treeResponse)
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
		Mode                string   `json:"mode"`
		WritingType         string   `json:"writing_type"`
		AssignmentFormat    string   `json:"assignment_format"`
		TargetAudience      string   `json:"target_audience"`
		SubjectMatter       string   `json:"subject_matter"`
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
		WritingType:         strings.TrimSpace(payload.WritingType),
		AssignmentFormat:    strings.TrimSpace(payload.AssignmentFormat),
		TargetAudience:      strings.TrimSpace(payload.TargetAudience),
		SubjectMatter:       strings.TrimSpace(payload.SubjectMatter),
		ExperienceLevel:     strings.TrimSpace(payload.ExperienceLevel),
		DesiredTone:         strings.TrimSpace(payload.DesiredTone),
		BiggestWeaknesses:   sanitizeStringList(payload.BiggestWeaknesses),
		DesiredOutcomes:     sanitizeStringList(payload.DesiredOutcomes),
		DifficultyIntensity: strings.TrimSpace(payload.DifficultyIntensity),
		WritingGoals:        strings.TrimSpace(payload.WritingGoals),
	}
	if !profile.Complete() {
		writeError(w, http.StatusBadRequest, fmt.Errorf("missing onboarding fields: %s", strings.Join(profile.MissingFields(), ", ")))
		return
	}

	user, err := s.Store.UserBySlug(r.Context(), appContext.UserSlug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	mode := strings.ToLower(strings.TrimSpace(payload.Mode))
	if mode == "" {
		mode = "edit"
	}
	firstTrackSetup := !s.userHasTrackProfile(r.Context(), appContext.UserID)
	bootstrapTrackIsEmpty := false
	if firstTrackSetup {
		if exercises, err := s.Store.ListExercises(r.Context(), appContext.UserID, appContext.TreeID, 1); err == nil && len(exercises) == 0 {
			bootstrapTrackIsEmpty = true
		}
	}
	profile.TemplateKey = domain.TemplateKeyForProfile(profile)
	treeDef := domain.GenerateTreeDefinition(appContext.UserSlug, user.Name, profile)
	starterCodes := domain.RecommendedStarterCodes(profile)
	recommendedRegions := domain.RecommendedRegionSlugs(profile)

	targetContext := appContext
	switch mode {
	case "create":
		treeDef.Slug = s.uniqueGeneratedTreeSlug(r.Context(), treeDef.Slug)
		profile.GeneratedTreeSlug = treeDef.Slug
	case "edit":
		if existing, err := s.Store.OnboardingProfileByEnrollmentID(r.Context(), appContext.EnrollmentID); err == nil && existing.GeneratedTreeSlug == appContext.TreeSlug {
			treeDef.Slug = appContext.TreeSlug
			profile.GeneratedTreeSlug = treeDef.Slug
			profile.EnrollmentID = appContext.EnrollmentID
			profile.UserID = appContext.UserID
		} else {
			treeDef.Slug = s.uniqueGeneratedTreeSlug(r.Context(), treeDef.Slug)
			profile.GeneratedTreeSlug = treeDef.Slug
		}
	default:
		writeError(w, http.StatusBadRequest, fmt.Errorf("unsupported onboarding mode"))
		return
	}

	if err := s.Store.SaveTreeDefinition(r.Context(), treeDef); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	_, treeID, enrollmentID, err := s.Store.EnsureDefaultUserTree(r.Context(), appContext.UserSlug, user.Name, treeDef.Slug)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	profile.EnrollmentID = enrollmentID
	profile.UserID = appContext.UserID
	if err := s.Store.SaveOnboardingProfile(r.Context(), profile); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.Store.SetUserActiveTree(r.Context(), appContext.UserID, treeDef.Slug); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if firstTrackSetup && bootstrapTrackIsEmpty && mode == "edit" && appContext.TreeSlug != treeDef.Slug {
		if err := s.Store.ArchiveUserTrack(r.Context(), appContext.UserID, appContext.TreeSlug); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	targetContext = session.Context{
		UserID:       appContext.UserID,
		TreeID:       treeID,
		EnrollmentID: enrollmentID,
		UserSlug:     appContext.UserSlug,
		TreeSlug:     treeDef.Slug,
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
		Context:            &requestContextResponse{UserSlug: targetContext.UserSlug, TreeSlug: targetContext.TreeSlug, UserID: targetContext.UserID, TreeID: targetContext.TreeID},
	})
}
