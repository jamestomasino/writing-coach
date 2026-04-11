package db

import (
	"context"
	"strings"
	"time"
)

const (
	sqliteBusyMaxAttempts = 5
	sqliteBusyBaseDelay   = 25 * time.Millisecond
)

func withSQLiteBusyRetry(ctx context.Context, fn func() error) error {
	var err error
	for attempt := 1; attempt <= sqliteBusyMaxAttempts; attempt++ {
		err = fn()
		if err == nil {
			return nil
		}
		if !isSQLiteBusyError(err) || attempt == sqliteBusyMaxAttempts {
			return err
		}
		delay := time.Duration(attempt*attempt) * sqliteBusyBaseDelay
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return err
}

func isSQLiteBusyError(err error) bool {
	if err == nil {
		return false
	}
	lower := strings.ToLower(err.Error())
	return strings.Contains(lower, "sqlite_busy") || strings.Contains(lower, "database is locked")
}
