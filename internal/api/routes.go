package api

import "net/http"

func (s *Server) routes() http.Handler {
	if s.validationLimiter == nil {
		s.validationLimiter = newAIValidationLimiter(s.Config.AIValidateLimitPerMinute, s.Config.AIValidateGlobalLimitPerMinute)
	}
	mux := http.NewServeMux()
	s.registerCoreRoutes(mux)
	s.registerAdminRoutes(mux)
	s.registerTreeRoutes(mux)
	s.registerPlaygroundRoutes(mux)
	s.registerAIJobRoutes(mux)
	return withServerLogging(withRecovery(withCORS(withAuth(mux, s.Config.APIToken, s.Config.KratosPublicURL, s.Config.AllowInsecureAuth))))
}

func (s *Server) registerCoreRoutes(mux *http.ServeMux) {
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
	mux.HandleFunc("GET /api/users", s.handleUsersList)
	mux.HandleFunc("POST /api/users", s.handleUsersCreate)
	mux.HandleFunc("GET /api/users/{slug}", s.handleUserGet)
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
	mux.HandleFunc("POST /api/assignments/{id}/close", s.handleAssignmentClose)
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
}

func (s *Server) registerAdminRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/admins", s.handleAdminsList)
	mux.HandleFunc("POST /api/admins", s.handleAdminsCreate)
	mux.HandleFunc("DELETE /api/admins/{email}", s.handleAdminsDelete)
	mux.HandleFunc("GET /api/admin/ai-provider-events", s.handleAdminAIProviderEvents)
	mux.HandleFunc("GET /api/admin/calibration", s.handleAdminCalibrationDashboard)
	mux.HandleFunc("POST /api/admin/calibration/run", s.handleAdminCalibrationRun)
	mux.HandleFunc("POST /api/admin/calibration/notifications/{id}/read", s.handleAdminCalibrationNotificationRead)
	mux.HandleFunc("POST /api/admin/calibration/runs/{id}/read", s.handleAdminCalibrationRunRead)
	mux.HandleFunc("POST /api/admin/calibration/runs/{id}/approval", s.handleAdminCalibrationRunApproval)
}

func (s *Server) registerTreeRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/trees", s.handleTreesList)
	mux.HandleFunc("POST /api/trees", s.handleTreeCreate)
	mux.HandleFunc("GET /api/trees/{slug}/versions", s.handleTreeVersionsList)
	mux.HandleFunc("GET /api/trees/{slug}/versions/{version}", s.handleTreeVersionGet)
	mux.HandleFunc("GET /api/trees/{slug}/diff", s.handleTreeDiff)
	mux.HandleFunc("POST /api/trees/{slug}/versions/{version}/restore", s.handleTreeVersionRestore)
	mux.HandleFunc("GET /api/trees/{slug}", s.handleTreeGet)
	mux.HandleFunc("PUT /api/trees/{slug}", s.handleTreeUpdate)
}

func (s *Server) registerPlaygroundRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/playground/sessions", s.handlePlaygroundSessionsList)
	mux.HandleFunc("POST /api/playground/sessions", s.handlePlaygroundSessionCreate)
	mux.HandleFunc("GET /api/playground/sessions/{id}", s.handlePlaygroundSessionGet)
	mux.HandleFunc("PUT /api/playground/sessions/{id}", s.handlePlaygroundSessionUpdate)
	mux.HandleFunc("GET /api/playground/sessions/{id}/drafts", s.handlePlaygroundSessionDraftsList)
	mux.HandleFunc("POST /api/playground/sessions/{id}/drafts", s.handlePlaygroundSessionDraftCreate)
	mux.HandleFunc("POST /api/playground/sessions/{id}/reviews", s.handlePlaygroundSessionReviewCreate)
	mux.HandleFunc("GET /api/playground/sessions/{id}/reviews", s.handlePlaygroundSessionReviewsList)
	mux.HandleFunc("GET /api/playground/reviews/{id}", s.handlePlaygroundReviewGet)
	mux.HandleFunc("POST /api/playground/review", s.handlePlaygroundReview)
}

func (s *Server) registerAIJobRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/jobs/{id}", s.handleAIJobGet)
}
