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

func (s Server) handleAdminAIProviderEvents(w http.ResponseWriter, r *http.Request) {
	if !s.requireAdmin(w, r) {
		return
	}
	limit := listLimit(r, 100, 250)
	hours := 24
	provider := strings.TrimSpace(normalizeProvider(r.URL.Query().Get("provider")))
	eventFilter := strings.TrimSpace(r.URL.Query().Get("event"))
	if raw := strings.TrimSpace(r.URL.Query().Get("hours")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value <= 0 {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid hours"))
			return
		}
		if value > 24*14 {
			value = 24 * 14
		}
		hours = value
	}
	since := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)
	events, err := s.Store.ListRecentAIProviderEvents(r.Context(), limit, since, provider, eventFilter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	summary, err := s.Store.SummarizeAIProviderEventsSince(r.Context(), since, provider, eventFilter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"summary": s.toAIProviderEventSummaryResponse(summary),
		"events":  s.toAIProviderEventResponses(events),
		"filters": aiProviderEventFiltersResponse{
			Hours:     hours,
			Provider:  provider,
			Event:     eventFilter,
			Providers: []string{"anthropic", "gemini", "openai", "groq", "xai"},
			Events: []string{
				"settings_validate_failed",
				"settings_validate_rate_limited",
				"settings_validate_succeeded",
				"settings_save_failed",
				"settings_save_rate_limited",
				"settings_save_succeeded",
				"provider_resolved",
				"provider_missing",
				"provider_resolve_failed",
				"generation_fallback",
			},
		},
	})
}

func (s Server) toAIProviderEventResponses(events []domain.AIProviderEvent) []aiProviderEventResponse {
	items := make([]aiProviderEventResponse, 0, len(events))
	for _, event := range events {
		item := aiProviderEventResponse{
			ID:         event.ID,
			UserID:     event.UserID,
			UserSlug:   event.UserSlug,
			Provider:   event.Provider,
			Event:      event.Event,
			Category:   event.Category,
			StatusCode: event.StatusCode,
			CreatedAt:  db.Since(event.CreatedAt),
		}
		if strings.TrimSpace(event.DetailJSON) != "" {
			var details map[string]any
			if err := json.Unmarshal([]byte(event.DetailJSON), &details); err == nil && len(details) > 0 {
				item.Details = details
			}
		}
		items = append(items, item)
	}
	return items
}

func (s Server) toAIProviderEventSummaryResponse(summary domain.AIProviderEventSummary) aiProviderEventSummaryResponse {
	return aiProviderEventSummaryResponse{
		Since:               db.Since(summary.Since),
		Total:               summary.Total,
		ValidationFailures:  summary.ValidationFailures,
		ValidationRateLimit: summary.ValidationRateLimit,
		Fallbacks:           summary.Fallbacks,
		ProviderCounts:      toAIProviderEventCountResponses(summary.ProviderCounts),
		CategoryCounts:      toAIProviderEventCountResponses(summary.CategoryCounts),
	}
}

func toAIProviderEventCountResponses(counts []domain.AIProviderEventCount) []aiProviderEventCountResponse {
	items := make([]aiProviderEventCountResponse, 0, len(counts))
	for _, count := range counts {
		items = append(items, aiProviderEventCountResponse{
			Label: count.Label,
			Count: count.Count,
		})
	}
	return items
}
