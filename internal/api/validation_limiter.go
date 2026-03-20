package api

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type aiValidationLimiter struct {
	mu         sync.Mutex
	window     time.Duration
	perUserMax int
	globalMax  int
	globalHits []time.Time
	userHits   map[string][]time.Time
}

type aiValidationLimitError struct {
	RetryAfter time.Duration
	Scope      string
}

func (e *aiValidationLimitError) Error() string {
	if e == nil {
		return ""
	}
	seconds := int(e.RetryAfter.Round(time.Second) / time.Second)
	if seconds < 1 {
		seconds = 1
	}
	return fmt.Sprintf("provider validation is rate-limiting requests right now. Wait about %d seconds and try again", seconds)
}

func newAIValidationLimiter(perUserMax, globalMax int) *aiValidationLimiter {
	return &aiValidationLimiter{
		window:     time.Minute,
		perUserMax: perUserMax,
		globalMax:  globalMax,
		userHits:   make(map[string][]time.Time),
	}
}

func (l *aiValidationLimiter) Allow(now time.Time, userID int64, provider string) error {
	if l == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.globalHits = pruneValidationHits(l.globalHits, now, l.window)
	if l.globalMax > 0 && len(l.globalHits) >= l.globalMax {
		return &aiValidationLimitError{
			RetryAfter: retryAfterForValidationHits(l.globalHits, now, l.window),
			Scope:      "global",
		}
	}

	key := validationLimiterKey(userID, provider)
	hits := pruneValidationHits(l.userHits[key], now, l.window)
	if l.perUserMax > 0 && len(hits) >= l.perUserMax {
		l.userHits[key] = hits
		return &aiValidationLimitError{
			RetryAfter: retryAfterForValidationHits(hits, now, l.window),
			Scope:      "user",
		}
	}

	l.globalHits = append(l.globalHits, now)
	hits = append(hits, now)
	l.userHits[key] = hits
	return nil
}

func validationLimiterKey(userID int64, provider string) string {
	return fmt.Sprintf("%d:%s", userID, strings.TrimSpace(normalizeProvider(provider)))
}

func pruneValidationHits(hits []time.Time, now time.Time, window time.Duration) []time.Time {
	if len(hits) == 0 {
		return nil
	}
	cutoff := now.Add(-window)
	index := 0
	for index < len(hits) && !hits[index].After(cutoff) {
		index++
	}
	if index >= len(hits) {
		return nil
	}
	return hits[index:]
}

func retryAfterForValidationHits(hits []time.Time, now time.Time, window time.Duration) time.Duration {
	if len(hits) == 0 {
		return time.Second
	}
	retryAfter := hits[0].Add(window).Sub(now)
	if retryAfter < time.Second {
		return time.Second
	}
	return retryAfter
}
