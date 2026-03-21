package api

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/session"
)

type requestContextResponse struct {
	UserSlug string `json:"user_slug"`
	TreeSlug string `json:"tree_slug"`
	UserID   int64  `json:"user_id"`
	TreeID   int64  `json:"tree_id"`
}

type authSessionResponse struct {
	Authenticated                      bool                    `json:"authenticated"`
	AuthMode                           string                  `json:"auth_mode"`
	Identity                           *authIdentityResponse   `json:"identity,omitempty"`
	Context                            *requestContextResponse `json:"context,omitempty"`
	OnboardingComplete                 bool                    `json:"onboarding_complete"`
	SetupStep                          string                  `json:"setup_step"`
	ActiveTreeSlug                     string                  `json:"active_tree_slug,omitempty"`
	IsAdmin                            bool                    `json:"is_admin"`
	AIProviderReady                    bool                    `json:"ai_provider_ready"`
	AIEffectiveProvider                string                  `json:"ai_effective_provider,omitempty"`
	AISystemFallback                   bool                    `json:"ai_system_fallback"`
	AIHasPersonalKey                   bool                    `json:"ai_has_personal_key"`
	AIPersonalProviderStorageAvailable bool                    `json:"ai_personal_provider_storage_available"`
}

type authIdentityResponse struct {
	Subject string `json:"subject"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
}

func (s Server) handleAuthSession(w http.ResponseWriter, r *http.Request) {
	mode := authModeFromContext(r.Context())
	fallback := systemFallbackAvailable(s.Config)
	resp := authSessionResponse{
		Authenticated:                      mode != "none",
		AuthMode:                           mode,
		SetupStep:                          "ready",
		AIProviderReady:                    fallback,
		AIEffectiveProvider:                effectiveProviderLabel(false, false),
		AISystemFallback:                   fallback,
		AIPersonalProviderStorageAvailable: personalProviderStorageAvailable(s.Config),
	}
	if ident, ok := identityFromContext(r.Context()); ok {
		resp.Identity = &authIdentityResponse{
			Subject: ident.Subject,
			Email:   ident.Email,
			Name:    ident.Name,
		}
		if ident.Email != "" {
			if allowed, err := s.Store.IsAdminEmail(r.Context(), ident.Email); err == nil {
				resp.IsAdmin = allowed
			}
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
			if profile, err := s.Store.OnboardingProfileByEnrollmentID(r.Context(), appContext.EnrollmentID); err == nil {
				resp.OnboardingComplete = profile.Complete()
			}
			if settings, err := s.Store.AIProviderSettingsByUserID(r.Context(), user.ID); err == nil {
				settingsResponse := s.toAIProviderSettingsResponse(settings, true)
				resp.AIProviderReady = settingsResponse.Ready
				resp.AIEffectiveProvider = settingsResponse.EffectiveProvider
				resp.AISystemFallback = settingsResponse.SystemFallback
				resp.AIHasPersonalKey = settingsResponse.HasKey
			}
			if resp.OnboardingComplete {
				if exercises, err := s.Store.ListExercises(r.Context(), user.ID, appContext.TreeID, 1); err == nil && len(exercises) == 0 {
					resp.SetupStep = "needs_first_assignment"
				}
			}
		}
	}
	if !resp.AIProviderReady {
		resp.SetupStep = "needs_ai_setup"
	} else if !resp.OnboardingComplete {
		resp.SetupStep = "needs_first_track"
	}
	writeJSON(w, http.StatusOK, resp)
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

func (s Server) coachingBrief(ctx context.Context, enrollmentID int64) string {
	profile := s.onboardingProfile(ctx, enrollmentID)
	if profile == nil {
		return ""
	}
	return domain.CoachingBrief(*profile)
}

func (s Server) onboardingProfile(ctx context.Context, enrollmentID int64) *domain.OnboardingProfile {
	profile, err := s.Store.OnboardingProfileByEnrollmentID(ctx, enrollmentID)
	if err != nil {
		return nil
	}
	return &profile
}

func (s Server) resolveSession(ctx context.Context, r *http.Request) (session.Context, error) {
	if ident, ok := identityFromContext(ctx); ok {
		userSlug := slugFromIdentity(ident)
		userName := displayNameFromIdentity(ident)
		return s.resolveUserSession(ctx, r, userSlug, userName)
	}

	userSlug := firstNonEmpty(r.URL.Query().Get("user"), r.Header.Get("X-Writing-Coach-User"), s.Config.DefaultUserSlug)
	userName := firstNonEmpty(r.URL.Query().Get("user_name"), s.Config.WriterName)
	return s.resolveUserSession(ctx, r, userSlug, userName)
}

func (s Server) resolveUserSession(ctx context.Context, r *http.Request, userSlug, userName string) (session.Context, error) {
	if err := s.Store.EnsureUser(ctx, userSlug, userName); err != nil {
		return session.Context{}, err
	}
	user, err := s.Store.UserBySlug(ctx, userSlug)
	if err != nil {
		return session.Context{}, err
	}
	treeSlug := strings.TrimSpace(firstNonEmpty(r.URL.Query().Get("tree"), r.Header.Get("X-Writing-Coach-Tree")))
	if treeSlug == "" {
		if nextUser, err := s.archiveLegacyBootstrapTrack(ctx, user); err != nil {
			log.Printf("session: legacy track cleanup failed for user=%s: %v", userSlug, err)
		} else {
			user = nextUser
		}
		if strings.TrimSpace(user.ActiveTreeSlug) != "" {
			treeSlug = user.ActiveTreeSlug
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

func (s Server) archiveLegacyBootstrapTrack(ctx context.Context, user domain.User) (domain.User, error) {
	tracks, err := s.Store.ListUserTracks(ctx, user.ID)
	if err != nil || len(tracks) < 2 {
		return user, err
	}

	var legacy *domain.UserTrack
	var generated []domain.UserTrack
	for i := range tracks {
		track := tracks[i]
		profile, err := s.Store.OnboardingProfileByEnrollmentID(ctx, track.EnrollmentID)
		if track.TreeSlug == domain.GlobalSkillGraphSlug {
			if err == nil && profile.GeneratedTreeSlug == track.TreeSlug {
				return user, nil
			}
			legacy = &track
			continue
		}
		if err == nil && profile.GeneratedTreeSlug == track.TreeSlug {
			generated = append(generated, track)
		}
	}
	if legacy == nil || len(generated) == 0 {
		return user, nil
	}

	if strings.TrimSpace(user.ActiveTreeSlug) == "" || user.ActiveTreeSlug == legacy.TreeSlug {
		nextActive := generated[0].TreeSlug
		for _, track := range generated {
			if track.IsActive {
				nextActive = track.TreeSlug
				break
			}
		}
		if err := s.Store.SetUserActiveTree(ctx, user.ID, nextActive); err != nil {
			return user, err
		}
	}
	if err := s.Store.ArchiveUserTrack(ctx, user.ID, legacy.TreeSlug); err != nil {
		return user, err
	}
	return s.Store.UserBySlug(ctx, user.Slug)
}

func (s Server) uniqueGeneratedTreeSlug(ctx context.Context, base string) string {
	base = strings.TrimSpace(base)
	if base == "" {
		base = "writer-track"
	}
	slug := base
	for suffix := 2; ; suffix++ {
		if _, err := s.Store.TreeBySlug(ctx, slug); err != nil {
			if db.IsNotFound(err) {
				return slug
			}
			return slug
		}
		slug = fmt.Sprintf("%s-%d", base, suffix)
	}
}

func (s Server) userHasTrackProfile(ctx context.Context, userID int64) bool {
	tracks, err := s.Store.ListUserTracks(ctx, userID)
	if err != nil {
		return false
	}
	for _, track := range tracks {
		if _, err := s.Store.OnboardingProfileByEnrollmentID(ctx, track.EnrollmentID); err == nil {
			return true
		}
	}
	return false
}
