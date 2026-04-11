package db

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIsSQLiteBusyError(t *testing.T) {
	if !isSQLiteBusyError(errors.New("database is locked (5) (SQLITE_BUSY)")) {
		t.Fatal("expected SQLITE_BUSY text to be treated as busy")
	}
	if !isSQLiteBusyError(errors.New("database is locked")) {
		t.Fatal("expected database-is-locked text to be treated as busy")
	}
	if isSQLiteBusyError(errors.New("some other write error")) {
		t.Fatal("expected non-busy errors not to be treated as busy")
	}
}

func TestWithSQLiteBusyRetryRetriesAndSucceeds(t *testing.T) {
	ctx := context.Background()
	attempts := 0
	err := withSQLiteBusyRetry(ctx, func() error {
		attempts++
		if attempts < 3 {
			return errors.New("database is locked (5) (SQLITE_BUSY)")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("with retry: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("attempt count = %d, want 3", attempts)
	}
}

func TestWithSQLiteBusyRetryStopsWhenContextCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	attempts := 0
	err := withSQLiteBusyRetry(ctx, func() error {
		attempts++
		return errors.New("database is locked (5) (SQLITE_BUSY)")
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
	if attempts != 1 {
		t.Fatalf("attempt count = %d, want 1", attempts)
	}
}

func TestWithSQLiteBusyRetryStopsAtMaxAttempts(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	attempts := 0
	err := withSQLiteBusyRetry(ctx, func() error {
		attempts++
		return errors.New("database is locked (5) (SQLITE_BUSY)")
	})
	if err == nil {
		t.Fatal("expected busy error")
	}
	if attempts != sqliteBusyMaxAttempts {
		t.Fatalf("attempt count = %d, want %d", attempts, sqliteBusyMaxAttempts)
	}
}
