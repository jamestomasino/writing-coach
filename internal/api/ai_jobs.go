package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/review"
	"github.com/tomasino/writing-coach/internal/session"
)

const (
	aiJobKindPromptNext       = "prompt_next"
	aiJobKindPromptRevise     = "prompt_revise"
	aiJobKindReviewSubmission = "review_submission"
	aiJobKindPlaygroundReview = "playground_review"
)

type promptReviseJobPayload struct {
	SubmissionID int64 `json:"submission_id"`
}

type reviewSubmissionJobPayload struct {
	SubmissionID int64 `json:"submission_id"`
}

type playgroundReviewJobPayload struct {
	SessionID        int64  `json:"session_id"`
	Content          string `json:"content"`
	WritingLanguage  string `json:"writing_language"`
	WritingType      string `json:"writing_type"`
	AssignmentFormat string `json:"assignment_format"`
	CoachingBrief    string `json:"coaching_brief"`
}

func (s *Server) runAIJobWorker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.Store.RequeueStaleAIJobs(ctx, 3*time.Minute); err != nil {
				log.Printf("ai job worker: requeue stale jobs failed: %v", err)
			}
			if err := s.processNextAIJob(ctx); err != nil && !errors.Is(err, sql.ErrNoRows) {
				log.Printf("ai job worker: process failed: %v", err)
			}
		}
	}
}

func (s Server) processNextAIJob(ctx context.Context) error {
	job, err := s.Store.ClaimNextAIJob(ctx)
	if err != nil {
		return err
	}
	log.Printf("ai job started: job=%d kind=%s attempt=%d", job.ID, job.Kind, job.AttemptCount)
	if err := s.processAIJob(ctx, job); err != nil {
		log.Printf("ai job failed: job=%d kind=%s attempt=%d err=%v", job.ID, job.Kind, job.AttemptCount, err)
		if failErr := s.Store.FailAIJob(ctx, job, err.Error()); failErr != nil {
			log.Printf("ai job failure update failed: job=%d err=%v", job.ID, failErr)
		}
		return err
	}
	log.Printf("ai job completed: job=%d kind=%s", job.ID, job.Kind)
	return nil
}

func (s Server) processAIJob(ctx context.Context, job domain.AIJob) error {
	switch job.Kind {
	case aiJobKindPromptNext:
		return s.processPromptNextJob(ctx, job)
	case aiJobKindPromptRevise:
		return s.processPromptReviseJob(ctx, job)
	case aiJobKindReviewSubmission:
		return s.processReviewSubmissionJob(ctx, job)
	case aiJobKindPlaygroundReview:
		return s.processPlaygroundReviewJob(ctx, job)
	default:
		return fmt.Errorf("unsupported ai job kind %q", job.Kind)
	}
}

func (s Server) processPromptNextJob(ctx context.Context, job domain.AIJob) error {
	appContext, err := s.jobAppContext(ctx, job)
	if err != nil {
		return err
	}
	ex, err := s.generateNextExercise(ctx, appContext)
	if err != nil {
		return err
	}
	payloadResult := aiJobResultResponse{Exercise: ptrExerciseResponse(toExerciseResponse(ex))}
	return s.Store.CompleteAIJob(ctx, job.ID, 0, 0, mustJSON(payloadResult))
}

func (s Server) processPromptReviseJob(ctx context.Context, job domain.AIJob) error {
	var payload promptReviseJobPayload
	if err := json.Unmarshal([]byte(job.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("decode prompt revise payload: %w", err)
	}
	if payload.SubmissionID == 0 {
		return fmt.Errorf("submission_id is required")
	}
	appContext, err := s.jobAppContext(ctx, job)
	if err != nil {
		return err
	}
	ex, err := s.createRevisionExercise(ctx, appContext, payload.SubmissionID)
	if err != nil {
		return err
	}
	return s.Store.CompleteAIJob(ctx, job.ID, ex.ID, 0, "")
}

func (s Server) processReviewSubmissionJob(ctx context.Context, job domain.AIJob) error {
	submissionID := job.SubmissionID
	if submissionID == 0 {
		var payload reviewSubmissionJobPayload
		if err := json.Unmarshal([]byte(job.PayloadJSON), &payload); err != nil {
			return fmt.Errorf("decode review payload: %w", err)
		}
		submissionID = payload.SubmissionID
	}
	if submissionID == 0 {
		return fmt.Errorf("submission_id is required")
	}

	sub, err := s.Store.GetSubmission(ctx, submissionID)
	if err != nil {
		return fmt.Errorf("load submission: %w", err)
	}
	if existing, err := s.Store.LatestReviewForSubmission(ctx, sub.ID); err == nil {
		return s.Store.CompleteAIJob(ctx, job.ID, 0, existing.ID, "")
	}
	activeTGOs, err := s.Store.ActiveTGOs(ctx, job.EnrollmentID)
	if err != nil {
		return fmt.Errorf("load active tgos: %w", err)
	}
	completedTGOs, err := s.Store.CompletedTGOs(ctx, job.EnrollmentID)
	if err != nil {
		return fmt.Errorf("load completed tgos: %w", err)
	}
	treeSlug, err := s.treeSlugForEnrollment(ctx, job.EnrollmentID)
	if err != nil {
		return fmt.Errorf("load tree slug: %w", err)
	}
	analyzerContext := analyzer.ContextOptions{TreeSlug: treeSlug}
	if profile, err := s.Store.OnboardingProfileByEnrollmentID(ctx, job.EnrollmentID); err == nil {
		analyzerContext = analyzer.ContextFromProfile(treeSlug, profile)
	}
	runtime, err := s.resolveLLMRuntime(ctx, job.UserID)
	if err != nil {
		return fmt.Errorf("resolve provider: %w", err)
	}

	reviewResult := s.Reviews.WithClient(runtime.Client, runtime.ProviderKind).ReviewSubmissionDetailedWithOptions(ctx, sub, activeTGOs, completedTGOs, review.Options{
		AnalyzerContext: analyzerContext,
	})
	if reviewResult.Review.ReviewKind == runtime.ProviderKind {
		reviewResult.Review.ProviderNote = formatProviderNote(runtime.ProviderKind, runtime.ReviewModel)
	}
	if reviewResult.Review.ReviewKind == "deterministic-fallback" {
		s.logAIProviderEvent("generation_fallback", runtime.ProviderKind, job.UserID, map[string]any{
			"artifact": "review",
			"reason":   strings.TrimSpace(reviewResult.Review.ProviderNote),
		})
	}
	reviewResult.Review.UserID = job.UserID
	reviewResult.Review.TreeID = job.TreeID
	recommendation, err := s.Curriculum.SyncTGOs(ctx, s.Store, treeSlug, job.EnrollmentID, reviewResult.Review)
	if err != nil {
		return fmt.Errorf("sync curriculum: %w", err)
	}
	reviewResult.Review.NextFocus = recommendation.Focus
	reviewID, err := s.Store.SaveReview(ctx, reviewResult.Review, reviewResult.Scores)
	if err != nil {
		return fmt.Errorf("save review: %w", err)
	}
	if err := s.Store.SaveReviewArtifacts(ctx, domain.ReviewArtifacts{
		ReviewID:           reviewID,
		AnalyzerReportJSON: mustJSON(reviewResult.AnalyzerReport),
		RecommendationJSON: mustJSON(recommendation),
		ComparisonJSON:     mustJSON(s.reviewComparisonPayload(ctx, sub, reviewResult.Review)),
		AnnotationsJSON:    mustJSON(reviewResult.Review.Annotations),
	}); err != nil {
		return fmt.Errorf("save review artifacts: %w", err)
	}
	if err := s.Store.UpdateCurriculumState(ctx, job.EnrollmentID, recommendation.Focus, recommendation.Difficulty, reviewID); err != nil {
		return fmt.Errorf("update curriculum state: %w", err)
	}
	return s.Store.CompleteAIJob(ctx, job.ID, 0, reviewID, "")
}

func (s Server) processPlaygroundReviewJob(ctx context.Context, job domain.AIJob) error {
	var payload playgroundReviewJobPayload
	if err := json.Unmarshal([]byte(job.PayloadJSON), &payload); err != nil {
		return fmt.Errorf("decode playground payload: %w", err)
	}
	var sessionID int64
	var draft domain.PlaygroundDraft
	content := strings.TrimSpace(payload.Content)
	writingLanguage := strings.TrimSpace(payload.WritingLanguage)
	writingType := strings.TrimSpace(payload.WritingType)
	assignmentFormat := strings.TrimSpace(payload.AssignmentFormat)
	coachingBrief := strings.TrimSpace(payload.CoachingBrief)
	if payload.SessionID != 0 {
		item, err := s.Store.GetPlaygroundSession(ctx, payload.SessionID)
		if err != nil {
			return fmt.Errorf("load playground session: %w", err)
		}
		sessionID = item.ID
		draft, err = s.Store.EnsurePlaygroundDraftSnapshot(ctx, item)
		if err != nil {
			return fmt.Errorf("save playground draft snapshot: %w", err)
		}
		content = strings.TrimSpace(draft.Content)
		writingLanguage = strings.TrimSpace(item.WritingLanguage)
		writingType = strings.TrimSpace(item.WritingType)
		assignmentFormat = strings.TrimSpace(item.AssignmentFormat)
		coachingBrief = strings.TrimSpace(item.CoachingBrief)
	}
	if content == "" {
		return fmt.Errorf("content is required")
	}

	runtime, err := s.resolveLLMRuntime(ctx, job.UserID)
	if err != nil {
		return fmt.Errorf("resolve provider: %w", err)
	}

	result := s.Reviews.WithClient(runtime.Client, runtime.ProviderKind).ReviewSubmissionDetailedWithOptions(ctx, domain.Submission{
		UserID:    job.UserID,
		TreeID:    job.TreeID,
		Content:   content,
		WordCount: db.CountWords(content),
	}, nil, nil, review.Options{
		AnalyzerContext: analyzer.ContextOptions{
			WritingLanguage:  writingLanguage,
			WritingType:      writingType,
			AssignmentFormat: assignmentFormat,
		},
		CoachingBrief: coachingBrief,
		AllowUnscoped: true,
	})
	if result.Review.ReviewKind == runtime.ProviderKind {
		result.Review.ProviderNote = formatProviderNote(runtime.ProviderKind, runtime.ReviewModel)
	}
	if result.Review.ReviewKind == "deterministic-fallback" {
		s.logAIProviderEvent("generation_fallback", runtime.ProviderKind, job.UserID, map[string]any{
			"artifact": "playground_review",
			"reason":   strings.TrimSpace(result.Review.ProviderNote),
		})
	}
	result.Review.SkillScores = append([]domain.SkillScore(nil), result.Scores...)
	if sessionID != 0 {
		comparisonJSON := ""
		if draft.ParentDraftID != 0 {
			if previousDraft, err := s.Store.GetPlaygroundDraft(ctx, draft.ParentDraftID); err == nil {
				if previousReview, err := s.Store.LatestPlaygroundReviewForDraft(ctx, previousDraft.ID); err == nil {
					currentSub := domain.Submission{Content: draft.Content, WordCount: draft.WordCount}
					previousSub := domain.Submission{Content: previousDraft.Content, WordCount: previousDraft.WordCount}
					comparison := review.CompareSubmissions(currentSub, previousSub, result.Review, previousReview.Review)
					if comparison.SkillSetMismatch {
						log.Printf("playground compare: skill set mismatch session=%d draft=%d parent_draft=%d", sessionID, draft.ID, previousDraft.ID)
					}
					comparisonJSON = mustJSON(map[string]any{
						"summary":               comparison.Summary,
						"word_delta":            comparison.WordDelta,
						"added_words":           comparison.AddedWords,
						"removed_words":         comparison.RemovedWords,
						"addressed_weaknesses":  comparison.AddressedWeaknesses,
						"persisting_weaknesses": comparison.PersistingWeaknesses,
						"skill_set_mismatch":    comparison.SkillSetMismatch,
						"skill_deltas":          comparison.SkillDeltas,
					})
				}
			}
		}
		reviewID, err := s.Store.SavePlaygroundReview(ctx, domain.PlaygroundReview{
			SessionID:          sessionID,
			DraftID:            draft.ID,
			UserID:             job.UserID,
			TreeID:             job.TreeID,
			Review:             result.Review,
			AnalyzerReportJSON: mustJSON(result.AnalyzerReport),
			ComparisonJSON:     comparisonJSON,
		})
		if err != nil {
			return fmt.Errorf("save playground review: %w", err)
		}
		saved, err := s.Store.GetPlaygroundReview(ctx, reviewID)
		if err != nil {
			return fmt.Errorf("load playground review: %w", err)
		}
		payloadResult := aiJobResultResponse{Review: ptrReviewResponse(toPlaygroundReviewResponse(saved).Review)}
		return s.Store.CompleteAIJob(ctx, job.ID, 0, 0, mustJSON(payloadResult))
	}
	payloadResult := aiJobResultResponse{Review: ptrReviewResponse(toReviewResponse(result.Review))}
	return s.Store.CompleteAIJob(ctx, job.ID, 0, 0, mustJSON(payloadResult))
}

func (s Server) treeSlugForEnrollment(ctx context.Context, enrollmentID int64) (string, error) {
	var slug string
	err := s.Store.SQL.QueryRowContext(ctx, `
		SELECT t.slug
		FROM user_tree_enrollments e
		JOIN tgo_trees t ON t.id = e.tree_id
		WHERE e.id = ?
	`, enrollmentID).Scan(&slug)
	return slug, err
}

func (s Server) jobAppContext(ctx context.Context, job domain.AIJob) (session.Context, error) {
	treeSlug, err := s.treeSlugForEnrollment(ctx, job.EnrollmentID)
	if err != nil {
		return session.Context{}, err
	}
	var userSlug string
	err = s.Store.SQL.QueryRowContext(ctx, `SELECT slug FROM users WHERE id = ?`, job.UserID).Scan(&userSlug)
	if err != nil {
		return session.Context{}, err
	}
	return session.Context{
		UserID:       job.UserID,
		TreeID:       job.TreeID,
		EnrollmentID: job.EnrollmentID,
		UserSlug:     userSlug,
		TreeSlug:     treeSlug,
	}, nil
}

func ptrReviewResponse(review reviewResponse) *reviewResponse {
	return &review
}

func (s Server) handleAIJobGet(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	id, err := parseOptionalInt64(r.PathValue("id"))
	if err != nil || id == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid job id"))
		return
	}
	job, err := s.Store.AIJobByID(r.Context(), id)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if !belongsToContext(job.UserID, job.TreeID, appContext) {
		writeError(w, http.StatusNotFound, fmt.Errorf("job not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"context": requestContextResponse{UserSlug: appContext.UserSlug, TreeSlug: appContext.TreeSlug, UserID: appContext.UserID, TreeID: appContext.TreeID},
		"job":     s.toAIJobResponse(r.Context(), job),
	})
}
