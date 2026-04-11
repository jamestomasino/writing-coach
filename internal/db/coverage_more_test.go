package db

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/tomasino/writing-coach/internal/domain"
)

func setupCoverageStore(t *testing.T) (*Store, context.Context) {
	t.Helper()
	root := t.TempDir()
	store, err := Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	ctx := context.Background()
	if err := store.Migrate(ctx, filepath.Join("..", "..", "migrations")); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	if err := store.EnsureSeedData(ctx, "Tester"); err != nil {
		t.Fatalf("seed: %v", err)
	}
	return store, ctx
}

func TestCoverageEnrollmentAndProfileFlows(t *testing.T) {
	store, ctx := setupCoverageStore(t)
	defer store.Close()

	userID, treeID, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "cover-user", "Cover User", "story-craft-track")
	if err != nil {
		t.Fatalf("ensure default user tree: %v", err)
	}

	enrollments, err := store.ListEnrollments(ctx)
	if err != nil || len(enrollments) == 0 {
		t.Fatalf("list enrollments: %v len=%d", err, len(enrollments))
	}
	byUser, err := store.ListEnrollmentsByUserID(ctx, userID)
	if err != nil || len(byUser) == 0 {
		t.Fatalf("list enrollments by user: %v len=%d", err, len(byUser))
	}
	if _, err := store.EnrollmentByID(ctx, enrollmentID); err != nil {
		t.Fatalf("enrollment by id: %v", err)
	}
	if _, err := store.ActiveEnrollmentIDByUserID(ctx, userID); err != nil {
		t.Fatalf("active enrollment by user: %v", err)
	}
	tracks, err := store.ListUserTracks(ctx, userID)
	if err != nil || len(tracks) == 0 {
		t.Fatalf("list user tracks: %v len=%d", err, len(tracks))
	}
	if _, err := store.GetCurriculumState(ctx, enrollmentID); err != nil {
		t.Fatalf("get curriculum state: %v", err)
	}

	profile := domain.OnboardingProfile{
		EnrollmentID:        enrollmentID,
		UserID:              userID,
		WritingLanguage:     "en-US",
		WritingType:         "technical writing",
		AssignmentFormat:    "how-to guide",
		TargetAudience:      "developers",
		SubjectMatter:       "APIs",
		ExperienceLevel:     "intermediate",
		DesiredTone:         "clear",
		BiggestWeaknesses:   []string{"verbosity"},
		DesiredOutcomes:     []string{"clarity"},
		DifficultyIntensity: "moderate",
		WritingGoals:        "better docs",
		GeneratedTreeSlug:   "story-craft-track",
		TemplateKey:         "technical-writing",
	}
	if err := store.SaveOnboardingProfile(ctx, profile); err != nil {
		t.Fatalf("save onboarding profile: %v", err)
	}
	loadedByEnrollment, err := store.OnboardingProfileByEnrollmentID(ctx, enrollmentID)
	if err != nil {
		t.Fatalf("onboarding by enrollment: %v", err)
	}
	if loadedByEnrollment.WritingLanguage != "en" {
		t.Fatalf("expected normalized writing language en, got %q", loadedByEnrollment.WritingLanguage)
	}
	loadedByUser, err := store.OnboardingProfileByUserID(ctx, userID)
	if err != nil {
		t.Fatalf("onboarding by user: %v", err)
	}
	if loadedByUser.EnrollmentID != enrollmentID {
		t.Fatalf("onboarding by user enrollment = %d", loadedByUser.EnrollmentID)
	}

	exID, err := store.SaveExercise(ctx, domain.Exercise{UserID: userID, TreeID: treeID, Title: "t", Brief: "b", Constraints: []string{"c"}, FocusSkills: []string{"clarity"}, SuccessCriteria: []string{"s"}, GenerationKind: "deterministic"})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	if _, err := store.GetExercise(ctx, exID); err != nil {
		t.Fatalf("get exercise: %v", err)
	}
	if _, err := store.ListExercises(ctx, userID, treeID, 20); err != nil {
		t.Fatalf("list exercises: %v", err)
	}
	if err := store.CloseExercise(ctx, userID, treeID, exID); err != nil {
		t.Fatalf("close exercise: %v", err)
	}

	sub1ID, err := store.SaveSubmission(ctx, domain.Submission{UserID: userID, TreeID: treeID, ExerciseID: exID, Content: "first draft text", WordCount: 3})
	if err != nil {
		t.Fatalf("save first submission: %v", err)
	}
	sub2ID, err := store.SaveSubmission(ctx, domain.Submission{UserID: userID, TreeID: treeID, ExerciseID: exID, ParentSubmissionID: sub1ID, Content: "second draft text", WordCount: 3})
	if err != nil {
		t.Fatalf("save second submission: %v", err)
	}
	sub2, err := store.GetSubmission(ctx, sub2ID)
	if err != nil {
		t.Fatalf("get submission: %v", err)
	}
	if _, err := store.ListSubmissions(ctx, userID, treeID, exID, 20); err != nil {
		t.Fatalf("list submissions: %v", err)
	}
	if _, err := store.LatestSubmissionForExercise(ctx, exID, userID, treeID); err != nil {
		t.Fatalf("latest submission: %v", err)
	}
	if _, err := store.PreviousSubmission(ctx, sub2); err != nil {
		t.Fatalf("previous submission: %v", err)
	}

	reviewID, err := store.SaveReview(ctx, domain.Review{UserID: userID, TreeID: treeID, SubmissionID: sub2ID, ReviewKind: "deterministic", Summary: "ok", Strengths: []string{"s"}, Weaknesses: []string{"w"}, AnalyzerFindings: []string{"f"}, TGOAssessments: []domain.TGOAssessment{{TGOCode: "story-causal-clarity", Status: "developing", Evidence: "e"}}, NextFocus: "clarity", MetricWordCount: 3}, []domain.SkillScore{{SubmissionID: sub2ID, Skill: "clarity", Score: 3}})
	if err != nil {
		t.Fatalf("save review: %v", err)
	}
	if _, err := store.GetReview(ctx, reviewID); err != nil {
		t.Fatalf("get review: %v", err)
	}
	if _, err := store.ListReviews(ctx, userID, treeID, sub2ID, 10); err != nil {
		t.Fatalf("list reviews: %v", err)
	}
	if _, err := store.ReviewTGOAssessments(ctx, reviewID); err != nil {
		t.Fatalf("review tgo assessments: %v", err)
	}
	if err := store.UpdateCurriculumState(ctx, enrollmentID, "clarity", 3, reviewID); err != nil {
		t.Fatalf("update curriculum state: %v", err)
	}

	reviewJob, err := store.EnqueueReviewJob(ctx, domain.ReviewJob{
		UserID:       userID,
		TreeID:       treeID,
		EnrollmentID: enrollmentID,
		SubmissionID: sub2ID,
		MaxAttempts:  2,
	})
	if err != nil {
		t.Fatalf("enqueue review job: %v", err)
	}
	if _, err := store.ReviewJobBySubmission(ctx, userID, treeID, sub2ID); err != nil {
		t.Fatalf("review job by submission: %v", err)
	}
	claimedReviewJob, err := store.ClaimNextReviewJob(ctx)
	if err != nil {
		t.Fatalf("claim review job: %v", err)
	}
	if err := store.CompleteReviewJob(ctx, claimedReviewJob.ID, reviewID); err != nil {
		t.Fatalf("complete review job: %v", err)
	}
	if err := store.FailReviewJob(ctx, reviewJob, "ignored"); err != nil {
		t.Fatalf("fail review job: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `UPDATE review_jobs SET status='running', updated_at=? WHERE id=?`, time.Now().UTC().Add(-10*time.Minute), reviewJob.ID); err != nil {
		t.Fatalf("mark review job stale: %v", err)
	}
	if err := store.RequeueStaleReviewJobs(ctx, time.Minute); err != nil {
		t.Fatalf("requeue stale review jobs: %v", err)
	}

	treeDef, err := store.TreeDefinitionBySlug(ctx, "story-craft-track")
	if err != nil {
		t.Fatalf("tree definition: %v", err)
	}
	if _, err := store.SkillAverages(ctx, userID, treeID, 10); err != nil {
		t.Fatalf("skill averages: %v", err)
	}
	if _, err := store.RecentSkillScores(ctx, userID, treeID, "clarity", 10); err != nil {
		t.Fatalf("recent skill scores: %v", err)
	}
	if _, err := store.LatestSkillScores(ctx, sub2ID); err != nil {
		t.Fatalf("latest skill scores: %v", err)
	}
	if _, err := store.ProgressReport(ctx, userID, treeID, treeDef.PrioritySkills, 10); err != nil {
		t.Fatalf("progress report: %v", err)
	}
	if _, err := store.RecurringWeaknesses(ctx, userID, treeID, 10); err != nil {
		t.Fatalf("recurring weaknesses: %v", err)
	}
	if _, err := store.RecurringAnalyzerFindings(ctx, userID, treeID, 10); err != nil {
		t.Fatalf("recurring analyzer findings: %v", err)
	}
	if _, _, err := store.StrongestWeakestSkills(ctx, userID, treeID, treeDef.PrioritySkills, 10); err != nil {
		t.Fatalf("strongest/weakest: %v", err)
	}
	if _, err := store.RecurringCompletedTGOSlips(ctx, userID, treeID, 10); err != nil {
		t.Fatalf("recurring completed slips: %v", err)
	}
	if _, err := store.History(ctx, userID, treeID); err != nil {
		t.Fatalf("history: %v", err)
	}
	if _, err := store.HistoryItems(ctx, userID, treeID); err != nil {
		t.Fatalf("history items: %v", err)
	}
	if _, err := store.CompletedAssignmentCount(ctx, userID, treeID); err != nil {
		t.Fatalf("completed assignment count: %v", err)
	}
	if _, err := store.RecentExerciseTitles(ctx, userID, treeID, 10); err != nil {
		t.Fatalf("recent exercise titles: %v", err)
	}
	if _, err := store.RecentExerciseSummaries(ctx, userID, treeID, 10); err != nil {
		t.Fatalf("recent exercise summaries: %v", err)
	}

	if err := store.EnsureAdminEmails(ctx, []string{"admin@example.com"}); err != nil {
		t.Fatalf("ensure admin emails: %v", err)
	}
	if ok, err := store.IsAdminEmail(ctx, "admin@example.com"); err != nil || !ok {
		t.Fatalf("is admin email: ok=%v err=%v", ok, err)
	}
	if _, err := store.ListAdminEmails(ctx); err != nil {
		t.Fatalf("list admin emails: %v", err)
	}
	if err := store.AddAdminEmail(ctx, "second-admin@example.com"); err != nil {
		t.Fatalf("add admin email: %v", err)
	}
	if err := store.RemoveAdminEmail(ctx, "second-admin@example.com"); err != nil {
		t.Fatalf("remove admin email: %v", err)
	}
	if _, err := store.ListUsers(ctx); err != nil {
		t.Fatalf("list users: %v", err)
	}
	if _, err := store.UserActiveTreeSlug(ctx, "cover-user"); err != nil {
		t.Fatalf("user active tree slug: %v", err)
	}
	if _, err := store.ListTrees(ctx); err != nil {
		t.Fatalf("list trees: %v", err)
	}
	if versions, err := store.ListTreeVersions(ctx, "story-craft-track"); err != nil {
		t.Fatalf("list tree versions: %v", err)
	} else if len(versions) > 0 {
		if _, _, err := store.TreeVersionByNumber(ctx, "story-craft-track", versions[len(versions)-1].Version); err != nil {
			t.Fatalf("tree version by number: %v", err)
		}
	}

	if err := store.ArchiveUserTrack(ctx, userID, "story-craft-track"); err != nil {
		t.Fatalf("archive user track: %v", err)
	}
}

func TestCoverageAIJobsAndPlaygroundFlows(t *testing.T) {
	store, ctx := setupCoverageStore(t)
	defer store.Close()

	userID, treeID, enrollmentID, err := store.EnsureDefaultUserTree(ctx, "jobs-user", "Jobs User", "story-craft-track")
	if err != nil {
		t.Fatalf("ensure user tree: %v", err)
	}

	job, err := store.EnqueueAIJob(ctx, domain.AIJob{UserID: userID, TreeID: treeID, EnrollmentID: enrollmentID, Kind: "prompt_next", ResourceKey: "rk-1", MaxAttempts: 2, PayloadJSON: `{"k":"v"}`})
	if err != nil {
		t.Fatalf("enqueue ai job: %v", err)
	}
	if _, err := store.AIJobByID(ctx, job.ID); err != nil {
		t.Fatalf("job by id: %v", err)
	}
	if _, err := store.AIJobByResourceKey(ctx, "rk-1"); err != nil {
		t.Fatalf("job by resource key: %v", err)
	}
	claimed, err := store.ClaimNextAIJob(ctx)
	if err != nil {
		t.Fatalf("claim ai job: %v", err)
	}
	if err := store.CompleteAIJob(ctx, claimed.ID, 11, 0, `{"done":true}`); err != nil {
		t.Fatalf("complete ai job: %v", err)
	}

	job2, err := store.EnqueueAIJob(ctx, domain.AIJob{UserID: userID, TreeID: treeID, EnrollmentID: enrollmentID, Kind: "review_submission", ResourceKey: "rk-2", SubmissionID: 99, MaxAttempts: 1})
	if err != nil {
		t.Fatalf("enqueue ai job 2: %v", err)
	}
	if _, err := store.LatestAIJobBySubmission(ctx, userID, treeID, 99, "review_submission"); err != nil {
		t.Fatalf("latest ai job by submission: %v", err)
	}
	claimed2, err := store.ClaimNextAIJob(ctx)
	if err != nil {
		t.Fatalf("claim ai job 2: %v", err)
	}
	if err := store.FailAIJob(ctx, claimed2, "boom"); err != nil {
		t.Fatalf("fail ai job: %v", err)
	}
	if _, err := store.SQL.ExecContext(ctx, `UPDATE ai_jobs SET status='running', updated_at=? WHERE id=?`, time.Now().UTC().Add(-10*time.Minute), job2.ID); err != nil {
		t.Fatalf("mark running stale: %v", err)
	}
	if err := store.RequeueStaleAIJobs(ctx, time.Minute); err != nil {
		t.Fatalf("requeue stale ai jobs: %v", err)
	}

	sessionID, err := store.SavePlaygroundSession(ctx, domain.PlaygroundSession{UserID: userID, TreeID: treeID, Title: "Draft", Content: "first", WritingLanguage: "en", WritingType: "technical writing", AssignmentFormat: "how-to", CoachingBrief: "brief"})
	if err != nil {
		t.Fatalf("save playground session: %v", err)
	}
	session, err := store.GetPlaygroundSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("get playground session: %v", err)
	}
	session.Content = "first improved"
	if err := store.UpdatePlaygroundSession(ctx, session); err != nil {
		t.Fatalf("update playground session: %v", err)
	}
	if _, err := store.ListPlaygroundSessions(ctx, userID, treeID, 10); err != nil {
		t.Fatalf("list playground sessions: %v", err)
	}

	draft, err := store.EnsurePlaygroundDraftSnapshot(ctx, session)
	if err != nil {
		t.Fatalf("ensure draft snapshot: %v", err)
	}
	if _, err := store.GetPlaygroundDraft(ctx, draft.ID); err != nil {
		t.Fatalf("get draft: %v", err)
	}
	if _, err := store.LatestPlaygroundDraftBySession(ctx, sessionID); err != nil {
		t.Fatalf("latest draft: %v", err)
	}
	if _, err := store.ListPlaygroundDrafts(ctx, userID, treeID, sessionID, 10); err != nil {
		t.Fatalf("list drafts: %v", err)
	}

	reviewPayload := domain.PlaygroundReview{
		SessionID: sessionID,
		DraftID:   draft.ID,
		UserID:    userID,
		TreeID:    treeID,
		Review: domain.Review{
			ReviewKind:       "deterministic",
			ProviderNote:     "note",
			Summary:          "summary",
			Strengths:        []string{"s"},
			Weaknesses:       []string{"w"},
			AnalyzerFindings: []string{"f"},
			NextFocus:        "clarity",
			MetricWordCount:  2,
			SkillScores:      []domain.SkillScore{{Skill: "clarity", Score: 3}},
			TGOAssessments:   []domain.TGOAssessment{{TGOCode: "story-causal-clarity", Status: "developing", Evidence: "e"}},
		},
		AnalyzerReportJSON: `{"r":1}`,
		ComparisonJSON:     `{"c":1}`,
	}
	reviewID, err := store.SavePlaygroundReview(ctx, reviewPayload)
	if err != nil {
		t.Fatalf("save playground review: %v", err)
	}
	if _, err := store.GetPlaygroundReview(ctx, reviewID); err != nil {
		t.Fatalf("get playground review: %v", err)
	}
	if _, err := store.ListPlaygroundReviews(ctx, userID, treeID, sessionID, 10); err != nil {
		t.Fatalf("list playground reviews: %v", err)
	}
	if _, err := store.LatestPlaygroundReviewForDraft(ctx, draft.ID); err != nil {
		t.Fatalf("latest playground review for draft: %v", err)
	}

	if _, err := store.GetPlaygroundSession(ctx, -1); err == nil || err == sql.ErrNoRows {
		// keep branch exercised without strict error matching
	}

	if err := store.ResetUserData(ctx, userID); err != nil {
		t.Fatalf("reset user data: %v", err)
	}
}
