package api

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
)

type aiProviderEventRecorder struct {
	store         *db.Store
	retentionDays int
	queue         chan domain.AIProviderEvent
	once          sync.Once
}

func newAIProviderEventRecorder(store *db.Store, cfg config.Config) *aiProviderEventRecorder {
	return &aiProviderEventRecorder{
		store:         store,
		retentionDays: cfg.AIProviderEventRetentionDays,
		queue:         make(chan domain.AIProviderEvent, 256),
	}
}

func (r *aiProviderEventRecorder) start(ctx context.Context) {
	if r == nil {
		return
	}
	r.once.Do(func() {
		go r.run(ctx)
	})
}

func (r *aiProviderEventRecorder) record(event domain.AIProviderEvent) {
	if r == nil {
		return
	}
	select {
	case r.queue <- event:
	default:
		log.Printf("ai_provider_event_dropped provider=%s event=%s user=%d", event.Provider, event.Event, event.UserID)
	}
}

func (r *aiProviderEventRecorder) run(ctx context.Context) {
	r.cleanup(ctx)
	ticker := time.NewTicker(6 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-r.queue:
			if err := r.store.SaveAIProviderEvent(ctx, event); err != nil {
				log.Printf("ai_provider_event_store_failed provider=%s event=%s user=%d err=%v", event.Provider, event.Event, event.UserID, err)
			}
		case <-ticker.C:
			r.cleanup(ctx)
		}
	}
}

func (r *aiProviderEventRecorder) cleanup(ctx context.Context) {
	if r == nil || r.store == nil || r.retentionDays <= 0 {
		return
	}
	cutoff := time.Now().UTC().Add(-time.Duration(r.retentionDays) * 24 * time.Hour)
	if err := r.store.DeleteAIProviderEventsOlderThan(ctx, cutoff); err != nil {
		log.Printf("ai_provider_event_retention_failed cutoff=%s err=%v", cutoff.Format(time.RFC3339), err)
	}
}

func (s *Server) startBackgroundWorkers(ctx context.Context) {
	if s.validationLimiter == nil {
		s.validationLimiter = newAIValidationLimiter(s.Config.AIValidateLimitPerMinute, s.Config.AIValidateGlobalLimitPerMinute)
	}
	if s.eventRecorder == nil {
		s.eventRecorder = newAIProviderEventRecorder(s.Store, s.Config)
	}
	s.eventRecorder.start(ctx)
	go s.runReviewWorker(ctx)
}
