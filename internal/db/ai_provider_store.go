package db

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

func (s *Store) SaveAIProviderEvent(ctx context.Context, event domain.AIProviderEvent) error {
	_, err := s.SQL.ExecContext(ctx, `
		INSERT INTO ai_provider_events (user_id, provider, event, category, status_code, detail_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`,
		event.UserID,
		strings.TrimSpace(event.Provider),
		strings.TrimSpace(event.Event),
		strings.TrimSpace(event.Category),
		event.StatusCode,
		strings.TrimSpace(event.DetailJSON),
		event.CreatedAt.UTC(),
	)
	return err
}

func (s *Store) DeleteAIProviderEventsOlderThan(ctx context.Context, cutoff time.Time) error {
	_, err := s.SQL.ExecContext(ctx, `
		DELETE FROM ai_provider_events
		WHERE created_at < ?
	`, cutoff.UTC())
	return err
}

func (s *Store) ListRecentAIProviderEvents(ctx context.Context, limit int, since time.Time, provider, event string) ([]domain.AIProviderEvent, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `
		SELECT e.id, e.user_id, COALESCE(u.slug, ''), e.provider, e.event, e.category, e.status_code, e.detail_json, e.created_at
		FROM ai_provider_events e
		LEFT JOIN users u ON u.id = e.user_id
		WHERE e.created_at >= ?
	`
	args := []any{since.UTC()}
	if strings.TrimSpace(provider) != "" {
		query += ` AND e.provider = ?`
		args = append(args, strings.TrimSpace(provider))
	}
	if strings.TrimSpace(event) != "" {
		query += ` AND e.event = ?`
		args = append(args, strings.TrimSpace(event))
	}
	query += `
		ORDER BY e.created_at DESC, e.id DESC
		LIMIT ?
	`
	args = append(args, limit)
	rows, err := s.SQL.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var events []domain.AIProviderEvent
	for rows.Next() {
		var event domain.AIProviderEvent
		if err := rows.Scan(
			&event.ID,
			&event.UserID,
			&event.UserSlug,
			&event.Provider,
			&event.Event,
			&event.Category,
			&event.StatusCode,
			&event.DetailJSON,
			&event.CreatedAt,
		); err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func (s *Store) SummarizeAIProviderEventsSince(ctx context.Context, since time.Time, provider, event string) (domain.AIProviderEventSummary, error) {
	summary := domain.AIProviderEventSummary{Since: since.UTC()}

	query := `
		SELECT provider, event, category, COUNT(*)
		FROM ai_provider_events
		WHERE created_at >= ?
	`
	args := []any{since.UTC()}
	if strings.TrimSpace(provider) != "" {
		query += ` AND provider = ?`
		args = append(args, strings.TrimSpace(provider))
	}
	if strings.TrimSpace(event) != "" {
		query += ` AND event = ?`
		args = append(args, strings.TrimSpace(event))
	}
	query += `
		GROUP BY provider, event, category
	`
	rows, err := s.SQL.QueryContext(ctx, query, args...)
	if err != nil {
		return summary, err
	}
	defer rows.Close()

	providerCounts := make(map[string]int)
	categoryCounts := make(map[string]int)
	for rows.Next() {
		var provider, event, category string
		var count int
		if err := rows.Scan(&provider, &event, &category, &count); err != nil {
			return summary, err
		}
		summary.Total += count
		if strings.Contains(event, "validate_failed") || strings.Contains(event, "save_failed") {
			summary.ValidationFailures += count
		}
		if strings.Contains(event, "rate_limited") {
			summary.ValidationRateLimit += count
		}
		if event == "generation_fallback" {
			summary.Fallbacks += count
		}
		providerKey := strings.TrimSpace(provider)
		if providerKey == "" {
			providerKey = "unknown"
		}
		providerCounts[providerKey] += count
		categoryKey := strings.TrimSpace(category)
		if categoryKey == "" {
			categoryKey = "none"
		}
		categoryCounts[categoryKey] += count
	}
	if err := rows.Err(); err != nil {
		return summary, err
	}

	summary.ProviderCounts = sortEventCounts(providerCounts)
	summary.CategoryCounts = sortEventCounts(categoryCounts)
	return summary, nil
}

func sortEventCounts(raw map[string]int) []domain.AIProviderEventCount {
	items := make([]domain.AIProviderEventCount, 0, len(raw))
	for label, count := range raw {
		items = append(items, domain.AIProviderEventCount{Label: label, Count: count})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].Count == items[j].Count {
			return items[i].Label < items[j].Label
		}
		return items[i].Count > items[j].Count
	})
	return items
}
