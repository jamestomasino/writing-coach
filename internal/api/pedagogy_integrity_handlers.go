package api

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
)

type pedagogyIntegrityAlertResponse struct {
	Severity string `json:"severity"`
	Code     string `json:"code"`
	Message  string `json:"message"`
}

type pedagogyIntegrityResponse struct {
	Since                        string                           `json:"since"`
	WindowHours                  int                              `json:"window_hours"`
	TotalReviews                 int                              `json:"total_reviews"`
	ReviewsMissingDecisionEvents int                              `json:"reviews_missing_decision_events"`
	ReviewScoredEvents           int                              `json:"review_scored_events"`
	RecommendationEvents         int                              `json:"recommendation_events"`
	HoldActivationEvents         int                              `json:"hold_activation_events"`
	HoldClearEvents              int                              `json:"hold_clear_events"`
	HoldBlockedEvents            int                              `json:"hold_blocked_events"`
	ActiveHoldEnrollments        int                              `json:"active_hold_enrollments"`
	Alerts                       []pedagogyIntegrityAlertResponse `json:"alerts"`
	Policy                       map[string]int                   `json:"policy"`
}

func (s Server) handleAdminPedagogyIntegrity(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	windowHours := 7 * 24
	if raw := strings.TrimSpace(r.URL.Query().Get("hours")); raw != "" {
		if value, err := strconv.Atoi(raw); err == nil {
			if value < 1 {
				value = 1
			}
			if value > 24*90 {
				value = 24 * 90
			}
			windowHours = value
		}
	}
	since := time.Now().UTC().Add(-time.Duration(windowHours) * time.Hour)
	snapshot, err := s.Store.PedagogyIntegritySnapshot(r.Context(), since, windowHours)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	snapshot.Alerts = pedagogyIntegrityAlerts(snapshot, s.Config.IntegrityHoldActivationClearGap, s.Config.IntegrityActiveHoldWarnCount)
	writeJSON(w, http.StatusOK, map[string]any{
		"integrity": toPedagogyIntegrityResponse(snapshot, s.Config.ProgressionHoldClearStreak, s.Config.IntegrityHoldActivationClearGap, s.Config.IntegrityActiveHoldWarnCount),
	})
}

func pedagogyIntegrityAlerts(snapshot domain.PedagogyIntegritySnapshot, holdActivationClearGap int, activeHoldWarnCount int) []domain.PedagogyIntegrityAlert {
	if holdActivationClearGap <= 0 {
		holdActivationClearGap = 3
	}
	if activeHoldWarnCount <= 0 {
		activeHoldWarnCount = 10
	}
	alerts := make([]domain.PedagogyIntegrityAlert, 0, 5)
	if snapshot.TotalReviews > 0 && snapshot.ReviewsMissingDecisionEvents > 0 {
		alerts = append(alerts, domain.PedagogyIntegrityAlert{
			Severity: "high",
			Code:     "missing_decision_events",
			Message:  "Some reviews are missing required decision events (review_scored or recommendation_issued).",
		})
	}
	if snapshot.HoldActivationEvents > snapshot.HoldClearEvents+holdActivationClearGap {
		alerts = append(alerts, domain.PedagogyIntegrityAlert{
			Severity: "medium",
			Code:     "hold_clearance_lag",
			Message:  "Hold activations are outpacing hold clears in the current window.",
		})
	}
	if snapshot.ActiveHoldEnrollments >= activeHoldWarnCount {
		alerts = append(alerts, domain.PedagogyIntegrityAlert{
			Severity: "medium",
			Code:     "elevated_active_holds",
			Message:  "A high number of enrollments are currently on progression hold.",
		})
	}
	if snapshot.HoldBlockedEvents > 0 {
		alerts = append(alerts, domain.PedagogyIntegrityAlert{
			Severity: "info",
			Code:     "advancement_blocks_observed",
			Message:  "Progression hold has blocked active-objective changes in the current window.",
		})
	}
	if len(alerts) == 0 {
		alerts = append(alerts, domain.PedagogyIntegrityAlert{
			Severity: "ok",
			Code:     "integrity_stable",
			Message:  "No pedagogy integrity alerts in the current window.",
		})
	}
	return alerts
}

func toPedagogyIntegrityResponse(snapshot domain.PedagogyIntegritySnapshot, holdClearStreakRequired int, holdActivationClearGap int, activeHoldWarnCount int) pedagogyIntegrityResponse {
	if holdClearStreakRequired <= 0 {
		holdClearStreakRequired = 2
	}
	if holdActivationClearGap <= 0 {
		holdActivationClearGap = 3
	}
	if activeHoldWarnCount <= 0 {
		activeHoldWarnCount = 10
	}
	alerts := make([]pedagogyIntegrityAlertResponse, 0, len(snapshot.Alerts))
	for _, alert := range snapshot.Alerts {
		alerts = append(alerts, pedagogyIntegrityAlertResponse{
			Severity: strings.TrimSpace(alert.Severity),
			Code:     strings.TrimSpace(alert.Code),
			Message:  strings.TrimSpace(alert.Message),
		})
	}
	return pedagogyIntegrityResponse{
		Since:                        db.Since(snapshot.Since),
		WindowHours:                  snapshot.WindowHours,
		TotalReviews:                 snapshot.TotalReviews,
		ReviewsMissingDecisionEvents: snapshot.ReviewsMissingDecisionEvents,
		ReviewScoredEvents:           snapshot.ReviewScoredEvents,
		RecommendationEvents:         snapshot.RecommendationEvents,
		HoldActivationEvents:         snapshot.HoldActivationEvents,
		HoldClearEvents:              snapshot.HoldClearEvents,
		HoldBlockedEvents:            snapshot.HoldBlockedEvents,
		ActiveHoldEnrollments:        snapshot.ActiveHoldEnrollments,
		Alerts:                       alerts,
		Policy: map[string]int{
			"hold_clear_streak_required": holdClearStreakRequired,
			"hold_activation_clear_gap":  holdActivationClearGap,
			"active_hold_warn_count":     activeHoldWarnCount,
		},
	}
}
