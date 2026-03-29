package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
)

type calibrationTrackLearningResponse struct {
	TreeSlug                string   `json:"tree_slug"`
	Domain                  string   `json:"domain"`
	SubmissionCount         int      `json:"submission_count"`
	DeterministicScoreCount int      `json:"deterministic_score_count"`
	TopScoreRate            float64  `json:"top_score_rate"`
	AverageScore            float64  `json:"average_score"`
	Issues                  []string `json:"issues"`
}

type calibrationDomainLearningResponse struct {
	Domain                  string   `json:"domain"`
	TrackCount              int      `json:"track_count"`
	SubmissionCount         int      `json:"submission_count"`
	DeterministicScoreCount int      `json:"deterministic_score_count"`
	TopScoreRate            float64  `json:"top_score_rate"`
	AverageScore            float64  `json:"average_score"`
	Issues                  []string `json:"issues"`
}

type calibrationRunResponse struct {
	ID                      int64                               `json:"id"`
	RunKind                 string                              `json:"run_kind"`
	Status                  string                              `json:"status"`
	MinSamples              int                                 `json:"min_samples"`
	LimitPerTrack           int                                 `json:"limit_per_track"`
	SubmissionCount         int                                 `json:"submission_count"`
	DeterministicScoreCount int                                 `json:"deterministic_score_count"`
	TrackLearnings          []calibrationTrackLearningResponse  `json:"track_learnings"`
	DomainLearnings         []calibrationDomainLearningResponse `json:"domain_learnings"`
	Highlights              []string                            `json:"highlights"`
	Recommendations         []string                            `json:"recommendations"`
	ErrorText               string                              `json:"error_text,omitempty"`
	StartedAt               string                              `json:"started_at"`
	CompletedAt             string                              `json:"completed_at,omitempty"`
	CreatedAt               string                              `json:"created_at"`
}

type adminNotificationResponse struct {
	ID           int64          `json:"id"`
	Kind         string         `json:"kind"`
	Title        string         `json:"title"`
	Body         string         `json:"body"`
	RelatedRunID int64          `json:"related_run_id,omitempty"`
	IsRead       bool           `json:"is_read"`
	CreatedAt    string         `json:"created_at"`
	ReadAt       string         `json:"read_at,omitempty"`
	Payload      map[string]any `json:"payload,omitempty"`
}

func (s Server) handleAdminCalibrationDashboard(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	limit := listLimit(r, 20, 100)
	notificationsLimit := 50
	if raw := strings.TrimSpace(r.URL.Query().Get("notifications_limit")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil && value > 0 {
			if value > 200 {
				value = 200
			}
			notificationsLimit = value
		}
	}

	runs, err := s.Store.ListRecentCalibrationRuns(r.Context(), limit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	notifications, err := s.Store.ListRecentAdminNotifications(r.Context(), notificationsLimit)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	unreadCount, err := s.Store.CountUnreadAdminNotifications(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	intervalHours := int(s.Config.CalibrationMaintenanceInterval / time.Hour)
	if intervalHours <= 0 {
		intervalHours = 24 * 30
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"runs":          toCalibrationRunResponses(runs),
		"notifications": toAdminNotificationResponses(notifications),
		"unread_count":  unreadCount,
		"settings": map[string]any{
			"enabled":           s.Config.CalibrationMaintenanceEnabled,
			"interval_hours":    intervalHours,
			"min_samples":       s.Config.CalibrationMinSamples,
			"limit_per_track":   s.Config.CalibrationLimitPerTrack,
			"background_active": s.calibration != nil,
			"running":           s.calibration != nil && s.calibration.IsRunning(),
		},
	})
}

func (s *Server) handleAdminCalibrationRun(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	if s.calibration == nil {
		s.calibration = newCalibrationMaintainer(s.Store, s.Config)
	}
	run, err := s.calibration.RunOnce(r.Context(), calibrationRunKindManual, 0)
	if err != nil {
		writeError(w, http.StatusConflict, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"run": toCalibrationRunResponse(run)})
}

func (s Server) handleAdminCalibrationNotificationRead(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	notificationID, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
	if err != nil || notificationID <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid notification id"))
		return
	}
	if err := s.Store.MarkAdminNotificationRead(r.Context(), notificationID); err != nil {
		status := http.StatusBadRequest
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": notificationID, "read": true})
}

func (s Server) handleAdminCalibrationRunRead(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	runID, err := strconv.ParseInt(strings.TrimSpace(r.PathValue("id")), 10, 64)
	if err != nil || runID <= 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid run id"))
		return
	}
	if err := s.Store.MarkCalibrationRunNotificationsRead(r.Context(), runID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"run_id": runID, "read": true})
}

func toCalibrationRunResponses(runs []domain.CalibrationRun) []calibrationRunResponse {
	items := make([]calibrationRunResponse, 0, len(runs))
	for _, run := range runs {
		items = append(items, toCalibrationRunResponse(run))
	}
	return items
}

func toCalibrationRunResponse(run domain.CalibrationRun) calibrationRunResponse {
	item := calibrationRunResponse{
		ID:                      run.ID,
		RunKind:                 run.RunKind,
		Status:                  run.Status,
		MinSamples:              run.MinSamples,
		LimitPerTrack:           run.LimitPerTrack,
		SubmissionCount:         run.SubmissionCount,
		DeterministicScoreCount: run.DeterministicScoreCount,
		TrackLearnings:          make([]calibrationTrackLearningResponse, 0, len(run.TrackLearnings)),
		DomainLearnings:         make([]calibrationDomainLearningResponse, 0, len(run.DomainLearnings)),
		Highlights:              append([]string{}, run.Highlights...),
		Recommendations:         append([]string{}, run.Recommendations...),
		ErrorText:               run.ErrorText,
		StartedAt:               db.Since(run.StartedAt),
		CreatedAt:               db.Since(run.CreatedAt),
	}
	if !run.CompletedAt.IsZero() {
		item.CompletedAt = db.Since(run.CompletedAt)
	}
	for _, track := range run.TrackLearnings {
		item.TrackLearnings = append(item.TrackLearnings, calibrationTrackLearningResponse{
			TreeSlug:                track.TreeSlug,
			Domain:                  track.Domain,
			SubmissionCount:         track.SubmissionCount,
			DeterministicScoreCount: track.DeterministicScoreCount,
			TopScoreRate:            track.TopScoreRate,
			AverageScore:            track.AverageScore,
			Issues:                  append([]string{}, track.Issues...),
		})
	}
	for _, domainLearning := range run.DomainLearnings {
		item.DomainLearnings = append(item.DomainLearnings, calibrationDomainLearningResponse{
			Domain:                  domainLearning.Domain,
			TrackCount:              domainLearning.TrackCount,
			SubmissionCount:         domainLearning.SubmissionCount,
			DeterministicScoreCount: domainLearning.DeterministicScoreCount,
			TopScoreRate:            domainLearning.TopScoreRate,
			AverageScore:            domainLearning.AverageScore,
			Issues:                  append([]string{}, domainLearning.Issues...),
		})
	}
	return item
}

func toAdminNotificationResponses(items []domain.AdminNotification) []adminNotificationResponse {
	result := make([]adminNotificationResponse, 0, len(items))
	for _, item := range items {
		response := adminNotificationResponse{
			ID:           item.ID,
			Kind:         item.Kind,
			Title:        item.Title,
			Body:         item.Body,
			RelatedRunID: item.RelatedRunID,
			IsRead:       item.IsRead,
			CreatedAt:    db.Since(item.CreatedAt),
		}
		if !item.ReadAt.IsZero() {
			response.ReadAt = db.Since(item.ReadAt)
		}
		if strings.TrimSpace(item.PayloadJSON) != "" {
			var payload map[string]any
			if err := json.Unmarshal([]byte(item.PayloadJSON), &payload); err == nil && len(payload) > 0 {
				response.Payload = payload
			}
		}
		result = append(result, response)
	}
	return result
}
