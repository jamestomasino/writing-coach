package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/review"
)

const calibrationRunKindScheduled = "scheduled"
const calibrationRunKindManual = "manual"

const calibrationStatusSucceeded = "succeeded"
const calibrationStatusFailed = "failed"

const adminNotificationKindCalibrationCompleted = "calibration_run_completed"

var evaluateObjectiveCalibrationCorpus = func() (review.ObjectiveEvalResult, error) {
	return review.EvaluateObjectiveScoreCorpus(review.DefaultObjectiveEvalCorpus())
}

type calibrationMaintainer struct {
	store *db.Store
	cfg   config.Config

	once sync.Once
	mu   sync.Mutex

	running bool
}

func newCalibrationMaintainer(store *db.Store, cfg config.Config) *calibrationMaintainer {
	return &calibrationMaintainer{store: store, cfg: cfg}
}

func (m *calibrationMaintainer) start(ctx context.Context) {
	if m == nil || m.store == nil || !m.cfg.CalibrationMaintenanceEnabled {
		return
	}
	m.once.Do(func() {
		go m.run(ctx)
	})
}

func (m *calibrationMaintainer) run(ctx context.Context) {
	interval := m.cfg.CalibrationMaintenanceInterval
	if interval <= 0 {
		interval = 30 * 24 * time.Hour
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if _, err := m.RunOnce(ctx, calibrationRunKindScheduled, 0); err != nil {
				log.Printf("calibration maintenance run failed: %v", err)
			}
		}
	}
}

func (m *calibrationMaintainer) IsRunning() bool {
	if m == nil {
		return false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.running
}

func (m *calibrationMaintainer) RunOnce(ctx context.Context, runKind string, triggeredByUserID int64) (domain.CalibrationRun, error) {
	if m == nil || m.store == nil {
		return domain.CalibrationRun{}, fmt.Errorf("calibration maintainer unavailable")
	}

	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return domain.CalibrationRun{}, fmt.Errorf("calibration run already in progress")
	}
	m.running = true
	m.mu.Unlock()
	defer func() {
		m.mu.Lock()
		m.running = false
		m.mu.Unlock()
	}()

	run, err := m.store.CreateCalibrationRun(ctx, runKind, triggeredByUserID, m.cfg.CalibrationMinSamples, m.cfg.CalibrationLimitPerTrack)
	if err != nil {
		return domain.CalibrationRun{}, err
	}

	runCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()
	snapshots, err := m.store.ListCalibrationTrackSnapshots(runCtx, m.cfg.DefaultTreeSlug, run.LimitPerTrack)
	if err != nil {
		run.Status = calibrationStatusFailed
		run.ErrorText = err.Error()
		run.CompletedAt = time.Now().UTC()
		_ = m.store.FinalizeCalibrationRun(ctx, run)
		_ = m.publishRunNotification(ctx, run)
		return run, err
	}
	hybridSignals, err := m.store.ListCalibrationHybridSignalSnapshots(runCtx, m.cfg.DefaultTreeSlug, run.LimitPerTrack)
	if err != nil {
		run.Status = calibrationStatusFailed
		run.ErrorText = err.Error()
		run.CompletedAt = time.Now().UTC()
		_ = m.store.FinalizeCalibrationRun(ctx, run)
		_ = m.publishRunNotification(ctx, run)
		return run, err
	}

	run.Status = calibrationStatusSucceeded
	run.CompletedAt = time.Now().UTC()
	run.TrackLearnings, run.DomainLearnings, run.Highlights, run.Recommendations, run.DataAdequate = analyzeCalibrationSnapshots(snapshots, hybridSignals, run.MinSamples)
	run.TrackLearnings, run.Highlights, run.Recommendations, run.DataAdequate = applyObjectiveEvalCalibrationGate(run.TrackLearnings, run.Highlights, run.Recommendations, run.DataAdequate)
	for _, item := range run.TrackLearnings {
		run.SubmissionCount += item.SubmissionCount
		run.DeterministicScoreCount += item.DeterministicScoreCount
	}
	if len(run.TrackLearnings) == 0 {
		run.Highlights = append(run.Highlights, "No deterministic calibration data is available yet.")
		run.Recommendations = append(run.Recommendations, "Wait for more reviewed submissions, then rerun calibration.")
	}
	if err := m.store.FinalizeCalibrationRun(ctx, run); err != nil {
		return run, err
	}
	if err := m.publishRunNotification(ctx, run); err != nil {
		log.Printf("calibration run notification failed run=%d err=%v", run.ID, err)
	}
	return run, nil
}

func (m *calibrationMaintainer) publishRunNotification(ctx context.Context, run domain.CalibrationRun) error {
	highlight := ""
	if len(run.Highlights) > 0 {
		highlight = run.Highlights[0]
	}
	title := "Calibration maintenance run completed"
	if run.Status == calibrationStatusFailed {
		title = "Calibration maintenance run failed"
		if strings.TrimSpace(run.ErrorText) != "" {
			highlight = run.ErrorText
		}
	}
	payload, err := json.Marshal(map[string]any{
		"run_id":          run.ID,
		"status":          run.Status,
		"highlights":      run.Highlights,
		"recommendations": run.Recommendations,
	})
	if err != nil {
		return err
	}
	return m.store.SaveAdminNotification(ctx, domain.AdminNotification{
		Kind:         adminNotificationKindCalibrationCompleted,
		Title:        title,
		Body:         highlight,
		PayloadJSON:  string(payload),
		RelatedRunID: run.ID,
		IsRead:       false,
		CreatedAt:    time.Now().UTC(),
	})
}

func analyzeCalibrationSnapshots(snapshots []domain.CalibrationTrackSnapshot, hybridSignals []domain.CalibrationHybridSignalSnapshot, minSamples int) ([]domain.CalibrationTrackLearning, []domain.CalibrationDomainLearning, []string, []string, bool) {
	tracks := make([]domain.CalibrationTrackLearning, 0, len(snapshots))
	domainRollup := map[string]*domain.CalibrationDomainLearning{}
	hybridByTrack := map[string]domain.CalibrationHybridSignalSnapshot{}
	for _, signal := range hybridSignals {
		hybridByTrack[signal.TreeSlug] = signal
	}

	insufficientTracks := 0
	missingDeterministicTracks := 0
	saturationTracks := 0
	scarcityTracks := 0
	dataAdequateTracks := 0

	for _, snapshot := range snapshots {
		hybrid := hybridByTrack[snapshot.TreeSlug]
		domainName := calibrationDomainForTrack(snapshot.TreeSlug)
		topScoreRate := 0.0
		if snapshot.DeterministicScoreCount > 0 {
			topScoreRate = (float64(snapshot.TopScoreCount) / float64(snapshot.DeterministicScoreCount)) * 100
		}
		issues := make([]string, 0, 4)
		if snapshot.SubmissionCount < minSamples {
			issues = append(issues, "insufficient_samples")
			insufficientTracks++
		}
		if snapshot.DeterministicScoreCount == 0 {
			issues = append(issues, "no_deterministic_scores")
			missingDeterministicTracks++
		}
		if snapshot.DeterministicScoreCount >= minSamples && topScoreRate > 40 {
			issues = append(issues, "top_score_saturation")
			saturationTracks++
		}
		if snapshot.DeterministicScoreCount >= minSamples && topScoreRate < 5 {
			issues = append(issues, "top_score_scarcity")
			scarcityTracks++
		}
		confidence := "low"
		if snapshot.SubmissionCount >= minSamples && snapshot.DeterministicScoreCount >= minSamples {
			confidence = "high"
			dataAdequateTracks++
		} else if snapshot.SubmissionCount >= (minSamples/2) && snapshot.DeterministicScoreCount >= (minSamples/2) {
			confidence = "medium"
		}

		track := domain.CalibrationTrackLearning{
			TreeSlug:                snapshot.TreeSlug,
			Domain:                  domainName,
			SubmissionCount:         snapshot.SubmissionCount,
			DeterministicScoreCount: snapshot.DeterministicScoreCount,
			HybridScoreCount:        hybrid.HybridScoreCount,
			HybridConflictCount:     hybrid.ConflictCount,
			HybridAdjustedCount:     hybrid.AdjustedCount,
			TopScoreRate:            topScoreRate,
			AverageScore:            snapshot.AverageScore,
			Confidence:              confidence,
			Issues:                  issues,
		}
		tracks = append(tracks, track)

		domainAgg, ok := domainRollup[domainName]
		if !ok {
			domainAgg = &domain.CalibrationDomainLearning{Domain: domainName}
			domainRollup[domainName] = domainAgg
		}
		domainAgg.TrackCount++
		domainAgg.SubmissionCount += track.SubmissionCount
		domainAgg.DeterministicScoreCount += track.DeterministicScoreCount
		domainAgg.HybridScoreCount += track.HybridScoreCount
		domainAgg.HybridConflictCount += track.HybridConflictCount
		domainAgg.HybridAdjustedCount += track.HybridAdjustedCount
		domainAgg.TopScoreRate += track.TopScoreRate
		domainAgg.AverageScore += track.AverageScore
	}

	domains := make([]domain.CalibrationDomainLearning, 0, len(domainRollup))
	for _, item := range domainRollup {
		if item.TrackCount > 0 {
			item.TopScoreRate = item.TopScoreRate / float64(item.TrackCount)
			item.AverageScore = item.AverageScore / float64(item.TrackCount)
		}
		if item.SubmissionCount < minSamples {
			item.Issues = append(item.Issues, "insufficient_samples")
		}
		if item.DeterministicScoreCount == 0 {
			item.Issues = append(item.Issues, "no_deterministic_scores")
		}
		if item.SubmissionCount >= minSamples && item.DeterministicScoreCount >= minSamples {
			item.Confidence = "high"
		} else if item.SubmissionCount >= (minSamples/2) && item.DeterministicScoreCount >= (minSamples/2) {
			item.Confidence = "medium"
		} else {
			item.Confidence = "low"
		}
		domains = append(domains, *item)
	}

	sort.Slice(tracks, func(i, j int) bool {
		if tracks[i].SubmissionCount == tracks[j].SubmissionCount {
			return tracks[i].TreeSlug < tracks[j].TreeSlug
		}
		return tracks[i].SubmissionCount > tracks[j].SubmissionCount
	})
	sort.Slice(domains, func(i, j int) bool {
		if domains[i].SubmissionCount == domains[j].SubmissionCount {
			return domains[i].Domain < domains[j].Domain
		}
		return domains[i].SubmissionCount > domains[j].SubmissionCount
	})

	highlights := make([]string, 0, 4)
	if insufficientTracks > 0 {
		highlights = append(highlights, fmt.Sprintf("%d track(s) are below the minimum sample threshold (%d).", insufficientTracks, minSamples))
	}
	if missingDeterministicTracks > 0 {
		highlights = append(highlights, fmt.Sprintf("%d track(s) have no deterministic scores in the sampled window.", missingDeterministicTracks))
	}
	if saturationTracks > 0 {
		highlights = append(highlights, fmt.Sprintf("%d track(s) show high 5/5 saturation (>40%%).", saturationTracks))
	}
	if scarcityTracks > 0 {
		highlights = append(highlights, fmt.Sprintf("%d track(s) show low 5/5 rates (<5%%).", scarcityTracks))
	}
	hybridConflictTracks := 0
	for _, track := range tracks {
		if track.HybridConflictCount > 0 {
			hybridConflictTracks++
		}
	}
	if hybridConflictTracks > 0 {
		highlights = append(highlights, fmt.Sprintf("%d track(s) show bounded-calibration conflicts worth review.", hybridConflictTracks))
	}
	if len(highlights) == 0 {
		highlights = append(highlights, "Calibration signals look stable across sampled tracks.")
	}

	recommendations := make([]string, 0, 5)
	if insufficientTracks > 0 {
		recommendations = append(recommendations, "Collect more submissions before making threshold changes for under-sampled tracks.")
	}
	if missingDeterministicTracks > 0 {
		recommendations = append(recommendations, "Verify deterministic scoring coverage for tracks with missing score rows.")
	}
	if saturationTracks > 0 {
		recommendations = append(recommendations, "Review top-score gates on saturated tracks to reduce false positives.")
	}
	if scarcityTracks > 0 {
		recommendations = append(recommendations, "Review strict gate criteria on low-5/5 tracks to reduce false negatives.")
	}
	if hybridConflictTracks > 0 {
		recommendations = append(recommendations, "Inspect hybrid calibration conflicts before approving rubric adjustments.")
	}
	dataAdequate := dataAdequateTracks > 0 && insufficientTracks == 0 && missingDeterministicTracks == 0
	if !dataAdequate {
		recommendations = append(recommendations, "Data is not adequate for automatic threshold tuning; keep changes in pending review only.")
	}
	if len(recommendations) == 0 {
		recommendations = append(recommendations, "No immediate threshold tuning is recommended.")
	}

	return tracks, domains, highlights, recommendations, dataAdequate
}

func calibrationDomainForTrack(treeSlug string) string {
	opts := analyzer.ContextOptions{TreeSlug: treeSlug, WritingType: writingTypeForTreeSlug(treeSlug)}
	return analyzer.DomainForContext(opts)
}

func writingTypeForTreeSlug(treeSlug string) string {
	if tree, ok := domain.BuiltInTreeBySlug(treeSlug); ok {
		return tree.Title
	}
	return treeSlug
}

func applyObjectiveEvalCalibrationGate(tracks []domain.CalibrationTrackLearning, highlights []string, recommendations []string, dataAdequate bool) ([]domain.CalibrationTrackLearning, []string, []string, bool) {
	result, err := evaluateObjectiveCalibrationCorpus()
	if err != nil {
		highlights = append(highlights, "Deterministic objective-score validation could not run.")
		recommendations = append(recommendations, "Fix objective-score evaluation corpus or loader errors before approving threshold tuning.")
		return tracks, highlights, recommendations, false
	}
	trackFailures := review.ObjectiveEvalTrackFailures(result)
	for i := range tracks {
		slug := strings.TrimSpace(tracks[i].TreeSlug)
		if _, ok := trackFailures[slug]; !ok {
			continue
		}
		if !calibrationIssueExists(tracks[i].Issues, "objective_eval_policy_failed") {
			tracks[i].Issues = append(tracks[i].Issues, "objective_eval_policy_failed")
		}
	}
	if result.PassedPolicyRequirements {
		return tracks, highlights, recommendations, dataAdequate
	}

	highlights = append(highlights, fmt.Sprintf("Deterministic objective-score gate failed (checks=%d pass_rate=%.3f required=%.3f).", result.TotalChecks, result.PassRate, result.RequiredMinPassRate))
	if result.MaxPairwiseTieRate != nil {
		highlights = append(highlights, fmt.Sprintf("Pairwise tie-rate %.3f exceeded maximum %.3f.", result.PairwiseTieRate, *result.MaxPairwiseTieRate))
	}
	for i, item := range result.PolicyFailures {
		if i >= 3 {
			break
		}
		recommendations = append(recommendations, "Objective eval policy failure: "+item)
	}
	if len(trackFailures) > 0 {
		recommendations = append(recommendations, fmt.Sprintf("%d track(s) failed objective policy checks; keep those tracks blocked for tuning unless an explicit override is documented.", len(trackFailures)))
	}
	recommendations = append(recommendations, "Do not approve automatic threshold tuning until objective-score policy checks pass.")
	return tracks, highlights, recommendations, false
}

func calibrationIssueExists(issues []string, issue string) bool {
	target := strings.TrimSpace(issue)
	for _, item := range issues {
		if strings.TrimSpace(item) == target {
			return true
		}
	}
	return false
}
