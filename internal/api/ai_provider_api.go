package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/anthropic"
	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/gemini"
	"github.com/tomasino/writing-coach/internal/llm"
	"github.com/tomasino/writing-coach/internal/openai"
	"github.com/tomasino/writing-coach/internal/secrets"
)

func (s Server) handleAISettingsGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	settings, err := s.Store.AIProviderSettingsByUserID(r.Context(), appContext.UserID)
	if err != nil && !db.IsNotFound(err) {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"settings": s.toAIProviderSettingsResponse(settings, err == nil),
	})
}

func (s *Server) handleAISettingsUpsert(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	var payload aiProviderSettingsPayload
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := validateAIProviderPayload(payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !personalProviderStorageAvailable(s.Config) {
		writeError(w, http.StatusServiceUnavailable, fmt.Errorf("personal provider storage is unavailable because WRITING_COACH_AI_KEY_SECRET is not configured"))
		return
	}
	existing, _ := s.Store.AIProviderSettingsByUserID(r.Context(), appContext.UserID)
	rawKey, encryptedKey, keyLast4, err := s.resolveAIProviderKey(r.Context(), existing, normalizeProvider(payload.Provider), strings.TrimSpace(payload.APIKey))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	now := time.Now().UTC()
	settings := domain.AIProviderSettings{
		UserID:              appContext.UserID,
		Provider:            normalizeProvider(payload.Provider),
		APIKeyLast4:         keyLast4,
		BaseURLOverride:     strings.TrimSpace(payload.BaseURLOverride),
		PromptModelOverride: strings.TrimSpace(payload.PromptModelOverride),
		ReviewModelOverride: strings.TrimSpace(payload.ReviewModelOverride),
		Enabled:             payload.Enabled,
	}
	if err := s.consumeAIProviderValidationAttempt(appContext.UserID, settings.Provider); err != nil {
		s.logAIProviderEvent("settings_save_rate_limited", settings.Provider, appContext.UserID, map[string]any{
			"category": aiProviderErrorCategory(err),
			"status":   statusForAIProviderValidationError(err),
		})
		if retryAfter := retryAfterForAIValidationError(err); retryAfter > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Round(time.Second)/time.Second)))
		}
		writeError(w, statusForAIProviderValidationError(err), userFacingAIProviderError("save", settings.Provider, err))
		return
	}
	if err := s.validateAIProviderSettings(r.Context(), settings, rawKey); err != nil {
		s.logAIProviderEvent("settings_save_failed", settings.Provider, appContext.UserID, map[string]any{
			"category": aiProviderErrorCategory(err),
			"status":   statusForAIProviderValidationError(err),
		})
		writeError(w, statusForAIProviderValidationError(err), userFacingAIProviderError("save", settings.Provider, err))
		return
	}
	s.logAIProviderEvent("settings_save_succeeded", settings.Provider, appContext.UserID, nil)
	settings.APIKeyEncrypted = encryptedKey
	settings.ValidatedAt = now
	settings.LastValidationError = ""
	if err := s.Store.SaveAIProviderSettings(r.Context(), settings); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"settings": s.toAIProviderSettingsResponse(settings, true),
	})
}

func (s Server) handleAISettingsDelete(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	if err := s.Store.DeleteAIProviderSettings(r.Context(), appContext.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"deleted": true,
		"settings": aiProviderSettingsResponse{
			EffectiveProvider: effectiveProviderLabel(false, false),
			SystemFallback:    systemFallbackAvailable(s.Config),
			Ready:             systemFallbackAvailable(s.Config),
		},
	})
}

func (s *Server) handleAISettingsValidate(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var payload aiProviderSettingsPayload
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := validateAIProviderPayload(payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if !personalProviderStorageAvailable(s.Config) {
		writeError(w, http.StatusServiceUnavailable, fmt.Errorf("personal provider storage is unavailable because WRITING_COACH_AI_KEY_SECRET is not configured"))
		return
	}
	existing, _ := s.Store.AIProviderSettingsByUserID(r.Context(), appContext.UserID)
	rawKey, _, keyLast4, err := s.resolveAIProviderKey(r.Context(), existing, normalizeProvider(payload.Provider), strings.TrimSpace(payload.APIKey))
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	settings := domain.AIProviderSettings{
		UserID:              appContext.UserID,
		Provider:            normalizeProvider(payload.Provider),
		APIKeyLast4:         keyLast4,
		BaseURLOverride:     strings.TrimSpace(payload.BaseURLOverride),
		PromptModelOverride: strings.TrimSpace(payload.PromptModelOverride),
		ReviewModelOverride: strings.TrimSpace(payload.ReviewModelOverride),
		Enabled:             payload.Enabled,
		ValidatedAt:         time.Now().UTC(),
	}
	if err := s.consumeAIProviderValidationAttempt(appContext.UserID, settings.Provider); err != nil {
		s.logAIProviderEvent("settings_validate_rate_limited", settings.Provider, appContext.UserID, map[string]any{
			"category": aiProviderErrorCategory(err),
			"status":   statusForAIProviderValidationError(err),
		})
		if retryAfter := retryAfterForAIValidationError(err); retryAfter > 0 {
			w.Header().Set("Retry-After", strconv.Itoa(int(retryAfter.Round(time.Second)/time.Second)))
		}
		writeError(w, statusForAIProviderValidationError(err), userFacingAIProviderError("validate", settings.Provider, err))
		return
	}
	if err := s.validateAIProviderSettings(r.Context(), settings, rawKey); err != nil {
		s.logAIProviderEvent("settings_validate_failed", settings.Provider, appContext.UserID, map[string]any{
			"category": aiProviderErrorCategory(err),
			"status":   statusForAIProviderValidationError(err),
		})
		writeError(w, statusForAIProviderValidationError(err), userFacingAIProviderError("validate", settings.Provider, err))
		return
	}
	s.logAIProviderEvent("settings_validate_succeeded", settings.Provider, appContext.UserID, nil)
	response := s.toAIProviderSettingsResponse(settings, true)
	response.HasKey = true
	response.KeyLast4 = settings.APIKeyLast4
	writeJSON(w, http.StatusOK, map[string]any{
		"valid":    true,
		"settings": response,
	})
}

func supportedAIProvider(provider string) bool {
	switch normalizeProvider(provider) {
	case "openai", "groq", "xai", "anthropic", "gemini":
		return true
	default:
		return false
	}
}

func normalizeProvider(provider string) string {
	return strings.ToLower(strings.TrimSpace(provider))
}

func validateAIProviderPayload(payload aiProviderSettingsPayload) error {
	if !supportedAIProvider(payload.Provider) {
		return fmt.Errorf("unsupported provider")
	}
	return nil
}

func (s Server) validateAIProviderSettings(ctx context.Context, settings domain.AIProviderSettings, apiKey string) error {
	client := s.providerClient(normalizeProvider(settings.Provider), strings.TrimSpace(apiKey), settings.BaseURLOverride, settings.PromptModelOverride, settings.ReviewModelOverride)
	validateCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.ValidateCredentials(validateCtx); err != nil {
		return err
	}
	return nil
}

func aiProviderErrorCategory(err error) string {
	var limitErr *aiValidationLimitError
	if errors.As(err, &limitErr) {
		return "local_rate_limit"
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	var httpErr *llm.HTTPError
	if errors.As(err, &httpErr) {
		switch httpErr.StatusCode {
		case http.StatusUnauthorized, http.StatusForbidden:
			return "auth"
		case http.StatusTooManyRequests:
			message := strings.ToLower(httpErr.Message)
			if strings.Contains(message, "quota") || strings.Contains(message, "insufficient_quota") || strings.Contains(message, "billing") {
				return "quota"
			}
			return "rate_limit"
		case http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return "upstream"
		default:
			return "provider_http"
		}
	}
	var timeoutErr interface{ Timeout() bool }
	if errors.As(err, &timeoutErr) && timeoutErr.Timeout() {
		return "timeout"
	}
	return "unknown"
}

func userFacingAIProviderError(action, provider string, err error) error {
	label := strings.ToUpper(strings.TrimSpace(provider))
	if label == "" {
		label = "provider"
	}
	switch aiProviderErrorCategory(err) {
	case "auth":
		return fmt.Errorf("%s rejected this API key. Check that the key is correct and has access to the selected endpoint", label)
	case "quota":
		return fmt.Errorf("%s cannot be used right now because the account is out of quota or billing is unavailable", label)
	case "local_rate_limit":
		return fmt.Errorf("provider validation is rate-limiting requests right now. Wait a moment and try again")
	case "rate_limit":
		return fmt.Errorf("%s is rate-limiting requests right now. Try again in a moment", label)
	case "timeout":
		return fmt.Errorf("the %s check timed out. Confirm the endpoint and try again", strings.ToLower(label))
	case "upstream":
		return fmt.Errorf("%s is temporarily unavailable. Try again shortly", label)
	default:
		if strings.TrimSpace(action) != "" {
			return fmt.Errorf("could not %s this %s configuration right now", action, label)
		}
		return fmt.Errorf("could not use %s right now", label)
	}
}

func (s Server) logAIProviderEvent(event, provider string, userID int64, fields map[string]any) {
	parts := []string{
		"ai_provider_event=" + strings.TrimSpace(event),
		"provider=" + firstNonEmpty(strings.TrimSpace(provider), "unknown"),
		fmt.Sprintf("user=%d", userID),
	}
	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%v", key, fields[key]))
	}
	log.Printf(strings.Join(parts, " "))

	if s.Store == nil || userID <= 0 {
		return
	}

	detailJSON := ""
	if len(fields) > 0 {
		if payload, err := json.Marshal(fields); err == nil {
			detailJSON = string(payload)
		}
	}
	statusCode := 0
	if raw, ok := fields["status"]; ok {
		switch value := raw.(type) {
		case int:
			statusCode = value
		case int64:
			statusCode = int(value)
		case float64:
			statusCode = int(value)
		}
	}
	category := inferredAIProviderEventCategory(strings.TrimSpace(event), fields)
	eventRecord := domain.AIProviderEvent{
		UserID:     userID,
		Provider:   strings.TrimSpace(provider),
		Event:      strings.TrimSpace(event),
		Category:   category,
		StatusCode: statusCode,
		DetailJSON: detailJSON,
		CreatedAt:  time.Now().UTC(),
	}
	if s.eventRecorder != nil {
		s.eventRecorder.record(eventRecord)
		return
	}
	if err := s.Store.SaveAIProviderEvent(context.Background(), eventRecord); err != nil {
		log.Printf("ai_provider_event_store_failed provider=%s event=%s user=%d err=%v", strings.TrimSpace(provider), strings.TrimSpace(event), userID, err)
	}
}

func inferredAIProviderEventCategory(event string, fields map[string]any) string {
	if raw, ok := fields["category"]; ok {
		if category := strings.TrimSpace(fmt.Sprint(raw)); category != "" && category != "<nil>" {
			return category
		}
	}
	switch {
	case strings.HasPrefix(event, "settings_"):
		return "settings"
	case strings.HasPrefix(event, "provider_"):
		return "provider"
	case strings.HasPrefix(event, "generation_"):
		return "generation"
	default:
		return "uncategorized"
	}
}

func (s Server) resolveAIProviderKey(ctx context.Context, existing domain.AIProviderSettings, provider, apiKey string) (string, string, string, error) {
	_ = ctx
	trimmedKey := strings.TrimSpace(apiKey)
	if trimmedKey != "" {
		encrypted, err := secrets.EncryptString(s.Config.AIKeySecret, trimmedKey)
		if err != nil {
			return "", "", "", err
		}
		return trimmedKey, encrypted, last4(trimmedKey), nil
	}
	if strings.TrimSpace(existing.Provider) == normalizeProvider(provider) && strings.TrimSpace(existing.APIKeyEncrypted) != "" {
		decrypted, err := secrets.DecryptString(s.Config.AIKeySecret, existing.APIKeyEncrypted)
		if err != nil {
			return "", "", "", err
		}
		return decrypted, existing.APIKeyEncrypted, existing.APIKeyLast4, nil
	}
	return "", "", "", fmt.Errorf("api key is required")
}

func statusForAIProviderValidationError(err error) int {
	var limitErr *aiValidationLimitError
	if errors.As(err, &limitErr) {
		return http.StatusTooManyRequests
	}
	var httpErr *llm.HTTPError
	if errors.As(err, &httpErr) {
		if httpErr.StatusCode == http.StatusUnauthorized || httpErr.StatusCode == http.StatusForbidden {
			return http.StatusBadRequest
		}
		return http.StatusBadGateway
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return http.StatusGatewayTimeout
	}
	return http.StatusBadGateway
}

func retryAfterForAIValidationError(err error) time.Duration {
	var limitErr *aiValidationLimitError
	if errors.As(err, &limitErr) {
		return limitErr.RetryAfter
	}
	return 0
}

func (s *Server) consumeAIProviderValidationAttempt(userID int64, provider string) error {
	if s.validationLimiter == nil {
		s.validationLimiter = newAIValidationLimiter(s.Config.AIValidateLimitPerMinute, s.Config.AIValidateGlobalLimitPerMinute)
	}
	return s.validationLimiter.Allow(time.Now().UTC(), userID, provider)
}

func last4(value string) string {
	trimmed := strings.TrimSpace(value)
	if len(trimmed) <= 4 {
		return trimmed
	}
	return trimmed[len(trimmed)-4:]
}

func systemFallbackAvailable(cfg config.Config) bool {
	return strings.TrimSpace(cfg.OpenAIAPIKey) != ""
}

func personalProviderStorageAvailable(cfg config.Config) bool {
	return strings.TrimSpace(cfg.AIKeySecret) != ""
}

func defaultBaseURLForProvider(provider string, cfg config.Config) string {
	switch normalizeProvider(provider) {
	case "anthropic":
		return "https://api.anthropic.com/v1"
	case "gemini":
		return "https://generativelanguage.googleapis.com/v1beta"
	case "groq":
		return "https://api.groq.com/openai/v1"
	case "xai":
		return "https://api.x.ai/v1"
	default:
		if strings.TrimSpace(cfg.OpenAIBaseURL) != "" {
			return cfg.OpenAIBaseURL
		}
		return "https://api.openai.com/v1"
	}
}

func defaultModelForProvider(provider, task string, cfg config.Config) string {
	switch normalizeProvider(provider) {
	case "anthropic":
		return "claude-sonnet-4-20250514"
	case "gemini":
		return "gemini-2.5-flash"
	default:
		if task == "review" {
			return cfg.ReviewModel
		}
		return cfg.PromptModel
	}
}

func effectiveProviderLabel(hasSettings, enabled bool) string {
	if hasSettings && enabled {
		return "user"
	}
	return "system/openai"
}

type llmRuntime struct {
	Client       llm.Client
	ProviderKind string
	PromptModel  string
	ReviewModel  string
}

func formatProviderNote(providerKind, model string) string {
	providerKind = strings.TrimSpace(providerKind)
	model = strings.TrimSpace(model)
	if providerKind == "" {
		return model
	}
	if model == "" {
		return providerKind
	}
	return providerKind + " • " + model
}

func (s Server) resolveLLMRuntime(ctx context.Context, userID int64) (llmRuntime, error) {
	settings, err := s.Store.AIProviderSettingsByUserID(ctx, userID)
	if err != nil && !db.IsNotFound(err) {
		return llmRuntime{}, err
	}
	if err == nil && settings.Enabled {
		decrypted, err := secrets.DecryptString(s.Config.AIKeySecret, settings.APIKeyEncrypted)
		if err != nil {
			s.logAIProviderEvent("provider_resolve_failed", settings.Provider, userID, map[string]any{"category": "decrypt"})
			return llmRuntime{}, err
		}
		normalized := normalizeProvider(settings.Provider)
		promptModel := firstNonEmpty(settings.PromptModelOverride, defaultModelForProvider(normalized, "prompt", s.Config))
		reviewModel := firstNonEmpty(settings.ReviewModelOverride, defaultModelForProvider(normalized, "review", s.Config))
		client := s.providerClient(normalized, decrypted, settings.BaseURLOverride, settings.PromptModelOverride, settings.ReviewModelOverride)
		s.logAIProviderEvent("provider_resolved", settings.Provider, userID, map[string]any{"mode": "personal"})
		return llmRuntime{
			Client:       client,
			ProviderKind: "user/" + normalized,
			PromptModel:  promptModel,
			ReviewModel:  reviewModel,
		}, nil
	}
	if systemFallbackAvailable(s.Config) {
		s.logAIProviderEvent("provider_resolved", "openai", userID, map[string]any{"mode": "system"})
		return llmRuntime{
			Client:       openai.NewClient(s.Config),
			ProviderKind: "system/openai",
			PromptModel:  strings.TrimSpace(s.Config.PromptModel),
			ReviewModel:  strings.TrimSpace(s.Config.ReviewModel),
		}, nil
	}
	s.logAIProviderEvent("provider_missing", "", userID, nil)
	return llmRuntime{}, nil
}

func (s Server) providerClient(provider, apiKey, baseURL, promptModel, reviewModel string) llm.Client {
	baseURL = firstNonEmpty(baseURL, defaultBaseURLForProvider(provider, s.Config))
	promptModel = firstNonEmpty(promptModel, defaultModelForProvider(provider, "prompt", s.Config))
	reviewModel = firstNonEmpty(reviewModel, defaultModelForProvider(provider, "review", s.Config))
	switch normalizeProvider(provider) {
	case "anthropic":
		return anthropic.NewClientWithOptions(anthropic.ClientOptions{
			APIKey:      apiKey,
			BaseURL:     baseURL,
			PromptModel: promptModel,
			ReviewModel: reviewModel,
		})
	case "gemini":
		return gemini.NewClientWithOptions(gemini.ClientOptions{
			APIKey:      apiKey,
			BaseURL:     baseURL,
			PromptModel: promptModel,
			ReviewModel: reviewModel,
		})
	default:
		return openai.NewClientWithOptions(openai.ClientOptions{
			APIKey:      apiKey,
			BaseURL:     baseURL,
			PromptModel: promptModel,
			ReviewModel: reviewModel,
		})
	}
}

func (s Server) toAIProviderSettingsResponse(settings domain.AIProviderSettings, exists bool) aiProviderSettingsResponse {
	fallback := systemFallbackAvailable(s.Config)
	personalStorage := personalProviderStorageAvailable(s.Config)
	if !exists {
		return aiProviderSettingsResponse{
			Enabled:                          false,
			HasKey:                           false,
			EffectiveProvider:                effectiveProviderLabel(false, false),
			SystemFallback:                   fallback,
			PersonalProviderStorageAvailable: personalStorage,
			Ready:                            fallback,
		}
	}
	return aiProviderSettingsResponse{
		Provider:                         settings.Provider,
		BaseURLOverride:                  settings.BaseURLOverride,
		PromptModelOverride:              settings.PromptModelOverride,
		ReviewModelOverride:              settings.ReviewModelOverride,
		Enabled:                          settings.Enabled,
		HasKey:                           strings.TrimSpace(settings.APIKeyEncrypted) != "",
		KeyLast4:                         settings.APIKeyLast4,
		ValidatedAt:                      db.Since(settings.ValidatedAt),
		LastValidationError:              settings.LastValidationError,
		EffectiveProvider:                effectiveProviderLabel(true, settings.Enabled),
		SystemFallback:                   fallback,
		PersonalProviderStorageAvailable: personalStorage,
		Ready:                            settings.Enabled || fallback,
	}
}
