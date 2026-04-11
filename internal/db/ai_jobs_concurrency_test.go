package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/tomasino/writing-coach/internal/domain"
)

func TestClaimNextAIJobConcurrentWorkersNoDuplicates(t *testing.T) {
	store, ctx := setupCoverageStore(t)
	defer store.Close()

	userID, treeID, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "concurrency-user", "Concurrency User", "story-craft-track")
	if err != nil {
		t.Fatalf("ensure user tree: %v", err)
	}

	const totalJobs = 24
	for i := range totalJobs {
		resourceKey := fmt.Sprintf("concurrency-rk-%d", i+1)
		if _, err := store.EnqueueAIJob(ctx, domain.AIJob{
			UserID:       userID,
			TreeID:       treeID,
			EnrollmentID: enrollmentID,
			Kind:         "prompt_next",
			ResourceKey:  resourceKey,
			MaxAttempts:  2,
			PayloadJSON:  fmt.Sprintf(`{"seed":%d}`, i+1),
		}); err != nil {
			t.Fatalf("enqueue job %d: %v", i+1, err)
		}
	}

	const workers = 6
	claimedCounts := map[int64]int{}
	var mu sync.Mutex
	var wg sync.WaitGroup

	claimAndComplete := func(ctx context.Context) {
		defer wg.Done()
		for {
			job, claimErr := store.ClaimNextAIJob(ctx)
			if claimErr != nil {
				if errors.Is(claimErr, sql.ErrNoRows) {
					return
				}
				t.Errorf("claim ai job: %v", claimErr)
				return
			}

			mu.Lock()
			claimedCounts[job.ID]++
			mu.Unlock()

			if job.Status != "running" {
				t.Errorf("job %d status = %q, want running", job.ID, job.Status)
				return
			}
			if job.AttemptCount != 1 {
				t.Errorf("job %d attempt_count = %d, want 1", job.ID, job.AttemptCount)
				return
			}

			if completeErr := store.CompleteAIJob(ctx, job.ID, 1000+job.ID, 0, `{"completed":true}`); completeErr != nil {
				t.Errorf("complete ai job %d: %v", job.ID, completeErr)
				return
			}
		}
	}

	for range workers {
		wg.Add(1)
		go claimAndComplete(ctx)
	}
	wg.Wait()

	if len(claimedCounts) != totalJobs {
		t.Fatalf("unique claimed jobs = %d, want %d", len(claimedCounts), totalJobs)
	}
	for jobID, count := range claimedCounts {
		if count != 1 {
			t.Fatalf("job %d claimed %d times", jobID, count)
		}
		record, err := store.AIJobByID(ctx, jobID)
		if err != nil {
			t.Fatalf("lookup claimed job %d: %v", jobID, err)
		}
		if record.Status != "completed" {
			t.Fatalf("job %d final status = %q, want completed", jobID, record.Status)
		}
	}
}
