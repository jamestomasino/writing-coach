package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/curriculum"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/prompt"
	"github.com/tomasino/writing-coach/internal/review"
)

func TestDashboardEndpoint(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/api/dashboard?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get dashboard: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var payload struct {
		CompletedAssignments int `json:"completed_assignments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}
	if payload.CompletedAssignments != 0 {
		t.Fatalf("expected empty dashboard count, got %d", payload.CompletedAssignments)
	}
}

func TestDashboardReportsCompletedAssignments(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	testServer := newTestServerWithStore(t, harness.Store, "", "")
	defer testServer.Close()

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	tree, err := harness.Store.TreeBySlug(context.Background(), "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("lookup tree: %v", err)
	}

	exerciseID, err := harness.Store.SaveExercise(context.Background(), domain.Exercise{
		UserID:          user.ID,
		TreeID:          tree.ID,
		Title:           "Assignment One",
		Brief:           "Write something.",
		Constraints:     []string{"one"},
		FocusSkills:     []string{"prose precision"},
		TGOCodes:        []string{"prose-precision"},
		SuccessCriteria: []string{"clear result"},
		GenerationKind:  "openai",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	submissionID, err := harness.Store.SaveSubmission(context.Background(), domain.Submission{
		UserID:     user.ID,
		TreeID:     tree.ID,
		ExerciseID: exerciseID,
		Content:    "A finished draft.",
		WordCount:  3,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	if _, err := harness.Store.SaveReview(context.Background(), domain.Review{
		UserID:           user.ID,
		TreeID:           tree.ID,
		SubmissionID:     submissionID,
		ReviewKind:       "coach",
		Summary:          "Solid.",
		Strengths:        []string{"clear"},
		Weaknesses:       []string{"tighten"},
		AnalyzerFindings: []string{},
		NextFocus:        "prose precision",
		MetricWordCount:  3,
		TGOAssessments: []domain.TGOAssessment{
			{TGOCode: "causal-clarity", Status: "developing", Evidence: "n/a"},
			{TGOCode: "scene-architecture", Status: "developing", Evidence: "n/a"},
			{TGOCode: "prose-precision", Status: "developing", Evidence: "n/a"},
		},
	}, nil); err != nil {
		t.Fatalf("save review: %v", err)
	}

	resp, err := http.Get(testServer.URL + "/api/dashboard?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get dashboard: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}

	var payload struct {
		CompletedAssignments int `json:"completed_assignments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode dashboard: %v", err)
	}
	if payload.CompletedAssignments != 1 {
		t.Fatalf("completed assignments = %d", payload.CompletedAssignments)
	}
}

func TestAccountResetClearsUserDataButKeepsAccount(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	testServer := newTestServerWithStore(t, harness.Store, "", "")
	defer testServer.Close()

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	tree, err := harness.Store.TreeBySlug(context.Background(), "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("lookup tree: %v", err)
	}
	if err := harness.Store.SaveOnboardingProfile(context.Background(), domain.OnboardingProfile{
		UserID:              user.ID,
		WritingType:         "marketing",
		AssignmentFormat:    "landing page",
		TargetAudience:      "buyers",
		SubjectMatter:       "product launch",
		ExperienceLevel:     "intermediate",
		DesiredTone:         "clear",
		BiggestWeaknesses:   []string{"sentence economy"},
		DesiredOutcomes:     []string{"improve professional communication"},
		DifficultyIntensity: "steady",
		WritingGoals:        "Write sharper marketing copy.",
		GeneratedTreeSlug:   "tester-track",
		TemplateKey:         "professional-writing",
	}); err != nil {
		t.Fatalf("save onboarding: %v", err)
	}
	exerciseID, err := harness.Store.SaveExercise(context.Background(), domain.Exercise{
		UserID:          user.ID,
		TreeID:          tree.ID,
		Title:           "Assignment One",
		Brief:           "Write something.",
		Constraints:     []string{"one"},
		FocusSkills:     []string{"prose precision"},
		TGOCodes:        []string{"prose-precision"},
		SuccessCriteria: []string{"clear result"},
		GenerationKind:  "openai",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	submissionID, err := harness.Store.SaveSubmission(context.Background(), domain.Submission{
		UserID:     user.ID,
		TreeID:     tree.ID,
		ExerciseID: exerciseID,
		Content:    "A finished draft.",
		WordCount:  3,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	if _, err := harness.Store.SaveReview(context.Background(), domain.Review{
		UserID:           user.ID,
		TreeID:           tree.ID,
		SubmissionID:     submissionID,
		ReviewKind:       "coach",
		Summary:          "Solid.",
		Strengths:        []string{"clear"},
		Weaknesses:       []string{"tighten"},
		AnalyzerFindings: []string{},
		NextFocus:        "prose precision",
		MetricWordCount:  3,
		TGOAssessments: []domain.TGOAssessment{
			{TGOCode: "causal-clarity", Status: "developing", Evidence: "n/a"},
			{TGOCode: "scene-architecture", Status: "developing", Evidence: "n/a"},
			{TGOCode: "prose-precision", Status: "developing", Evidence: "n/a"},
		},
	}, nil); err != nil {
		t.Fatalf("save review: %v", err)
	}

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/account/reset?user=tester", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("new reset request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("reset request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("reset status = %d", resp.StatusCode)
	}

	if _, err := harness.Store.OnboardingProfileByUserID(context.Background(), user.ID); err == nil {
		t.Fatal("expected onboarding profile to be cleared")
	}
	exercises, err := harness.Store.ListExercises(context.Background(), user.ID, tree.ID, 10)
	if err != nil {
		t.Fatalf("list exercises: %v", err)
	}
	if len(exercises) != 0 {
		t.Fatalf("expected exercises cleared, got %d", len(exercises))
	}
	userAfter, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user after reset: %v", err)
	}
	if userAfter.ID != user.ID {
		t.Fatalf("expected user account to remain, got id %d", userAfter.ID)
	}
	if userAfter.ActiveTreeSlug != "" {
		t.Fatalf("expected active tree slug cleared, got %q", userAfter.ActiveTreeSlug)
	}
}

func TestReadyEndpoint(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/api/ready")
	if err != nil {
		t.Fatalf("get ready: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("ready status: %d", resp.StatusCode)
	}
}

func TestRecoveryMiddlewareReturnsJSON500(t *testing.T) {
	handler := withRecovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/panic", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); !strings.Contains(got, "application/json") {
		t.Fatalf("content-type = %q", got)
	}
	var payload errorResponse
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error payload: %v", err)
	}
	if payload.Error != "internal server error" {
		t.Fatalf("error payload = %q", payload.Error)
	}
}

func TestTreeAndEnrollmentEndpoints(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/api/trees?include_tgos=1")
	if err != nil {
		t.Fatalf("get trees: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("trees status: %d", resp.StatusCode)
	}
	var treesPayload struct {
		Trees []struct {
			Slug string `json:"slug"`
			TGOs []struct {
				Code string `json:"code"`
			} `json:"tgos"`
		} `json:"trees"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&treesPayload); err != nil {
		t.Fatalf("decode trees: %v", err)
	}
	if len(treesPayload.Trees) < 2 {
		t.Fatalf("tree count = %d", len(treesPayload.Trees))
	}
	if len(treesPayload.Trees[0].TGOs) == 0 {
		t.Fatal("expected embedded TGO metadata")
	}

	customTreeResp, err := http.Post(testServer.URL+"/api/trees", "application/json", strings.NewReader(`{
		"slug":"essay-basics",
		"title":"Essay Basics",
		"description":"A simple expository track",
		"seed_codes":["claim-clarity","sentence-flow","paragraph-focus"],
		"priority_skills":["clarity and coherence","paragraph control"],
		"tgos":[
			{"code":"claim-clarity","title":"Claim Clarity","description":"State the main point clearly.","stage":"foundation","stage_order":1,"mastery_hint":"The thesis is immediately legible."},
			{"code":"sentence-flow","title":"Sentence Flow","description":"Keep prose readable.","stage":"foundation","stage_order":2,"mastery_hint":"Sentences connect cleanly."},
			{"code":"paragraph-focus","title":"Paragraph Focus","description":"Each paragraph serves one purpose.","stage":"foundation","stage_order":3,"mastery_hint":"Paragraphs stay unified."}
		]
	}`))
	if err != nil {
		t.Fatalf("create tree: %v", err)
	}
	defer customTreeResp.Body.Close()
	if customTreeResp.StatusCode != http.StatusCreated {
		t.Fatalf("custom tree status: %d", customTreeResp.StatusCode)
	}

	createResp, err := http.Post(testServer.URL+"/api/enrollments", "application/json", strings.NewReader(`{"user_slug":"kid","user_name":"Kid","tree_slug":"youth-writing-foundations"}`))
	if err != nil {
		t.Fatalf("create enrollment: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create enrollment status: %d", createResp.StatusCode)
	}

	listResp, err := http.Get(testServer.URL + "/api/enrollments")
	if err != nil {
		t.Fatalf("list enrollments: %v", err)
	}
	defer listResp.Body.Close()
	if listResp.StatusCode != http.StatusOK {
		t.Fatalf("enrollment list status: %d", listResp.StatusCode)
	}
	var enrollmentsPayload struct {
		Enrollments []struct {
			UserSlug string `json:"user_slug"`
			TreeSlug string `json:"tree_slug"`
		} `json:"enrollments"`
	}
	if err := json.NewDecoder(listResp.Body).Decode(&enrollmentsPayload); err != nil {
		t.Fatalf("decode enrollments: %v", err)
	}
	found := false
	for _, enrollment := range enrollmentsPayload.Enrollments {
		if enrollment.UserSlug == "kid" && enrollment.TreeSlug == "youth-writing-foundations" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected kid enrollment")
	}

	userResp, err := http.Get(testServer.URL + "/api/users/tester")
	if err != nil {
		t.Fatalf("get user: %v", err)
	}
	defer userResp.Body.Close()
	if userResp.StatusCode != http.StatusOK {
		t.Fatalf("user status: %d", userResp.StatusCode)
	}

	treeResp, err := http.Get(testServer.URL + "/api/trees/youth-writing-foundations")
	if err != nil {
		t.Fatalf("get tree: %v", err)
	}
	defer treeResp.Body.Close()
	if treeResp.StatusCode != http.StatusOK {
		t.Fatalf("tree status: %d", treeResp.StatusCode)
	}

	customTreeGetResp, err := http.Get(testServer.URL + "/api/trees/essay-basics")
	if err != nil {
		t.Fatalf("get custom tree: %v", err)
	}
	defer customTreeGetResp.Body.Close()
	if customTreeGetResp.StatusCode != http.StatusOK {
		t.Fatalf("custom tree get status: %d", customTreeGetResp.StatusCode)
	}

	versionsResp, err := http.Get(testServer.URL + "/api/trees/essay-basics/versions")
	if err != nil {
		t.Fatalf("get tree versions: %v", err)
	}
	defer versionsResp.Body.Close()
	if versionsResp.StatusCode != http.StatusOK {
		t.Fatalf("versions status: %d", versionsResp.StatusCode)
	}

	boardResp, err := http.Get(testServer.URL + "/api/enrollments/2/board")
	if err != nil {
		t.Fatalf("get board: %v", err)
	}
	defer boardResp.Body.Close()
	if boardResp.StatusCode != http.StatusOK {
		t.Fatalf("board status: %d", boardResp.StatusCode)
	}
}

func TestTreeUpdateCreatesVersion(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	createResp, err := http.Post(testServer.URL+"/api/trees", "application/json", strings.NewReader(`{
		"slug":"revision-track",
		"title":"Revision Track",
		"description":"Version one",
		"seed_codes":["draft-control","clarity","revision-discipline"],
		"priority_skills":["prose precision"],
		"tgos":[
			{"code":"draft-control","title":"Draft Control","description":"Hold shape.","stage":"core","stage_order":1,"mastery_hint":"shape holds"},
			{"code":"clarity","title":"Clarity","description":"Stay readable.","stage":"core","stage_order":2,"mastery_hint":"clear"},
			{"code":"revision-discipline","title":"Revision Discipline","description":"Revise deliberately.","stage":"core","stage_order":3,"mastery_hint":"deliberate"}
		]
	}`))
	if err != nil {
		t.Fatalf("create tree: %v", err)
	}
	defer createResp.Body.Close()
	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("create status: %d", createResp.StatusCode)
	}

	req, err := http.NewRequest(http.MethodPut, testServer.URL+"/api/trees/revision-track", strings.NewReader(`{
		"title":"Revision Track",
		"description":"Version two",
		"seed_codes":["draft-control","clarity","revision-discipline"],
		"priority_skills":["prose precision","narrative clarity"],
		"tgos":[
			{"code":"draft-control","title":"Draft Control","description":"Hold shape.","stage":"core","stage_order":1,"mastery_hint":"shape holds"},
			{"code":"clarity","title":"Clarity","description":"Stay readable.","stage":"core","stage_order":2,"mastery_hint":"clear"},
			{"code":"revision-discipline","title":"Revision Discipline","description":"Revise deliberately.","stage":"core","stage_order":3,"mastery_hint":"deliberate"}
		]
	}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	updateResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("update tree: %v", err)
	}
	defer updateResp.Body.Close()
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update status: %d", updateResp.StatusCode)
	}

	versionsResp, err := http.Get(testServer.URL + "/api/trees/revision-track/versions")
	if err != nil {
		t.Fatalf("versions request: %v", err)
	}
	defer versionsResp.Body.Close()
	if versionsResp.StatusCode != http.StatusOK {
		t.Fatalf("versions status: %d", versionsResp.StatusCode)
	}
	var versionsPayload struct {
		Versions []struct {
			Version int `json:"version"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(versionsResp.Body).Decode(&versionsPayload); err != nil {
		t.Fatalf("decode versions: %v", err)
	}
	if len(versionsPayload.Versions) < 2 {
		t.Fatalf("version count = %d", len(versionsPayload.Versions))
	}

	diffResp, err := http.Get(testServer.URL + "/api/trees/revision-track/diff?from=1&to=2")
	if err != nil {
		t.Fatalf("diff request: %v", err)
	}
	defer diffResp.Body.Close()
	if diffResp.StatusCode != http.StatusOK {
		t.Fatalf("diff status: %d", diffResp.StatusCode)
	}

	versionResp, err := http.Get(testServer.URL + "/api/trees/revision-track/versions/1")
	if err != nil {
		t.Fatalf("version get request: %v", err)
	}
	defer versionResp.Body.Close()
	if versionResp.StatusCode != http.StatusOK {
		t.Fatalf("version get status: %d", versionResp.StatusCode)
	}

	restoreResp, err := http.Post(testServer.URL+"/api/trees/revision-track/versions/1/restore", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("restore request: %v", err)
	}
	defer restoreResp.Body.Close()
	if restoreResp.StatusCode != http.StatusOK {
		t.Fatalf("restore status: %d", restoreResp.StatusCode)
	}
}

func TestExerciseSubmissionAndReviewEndpoints(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	promptResp, err := http.Post(testServer.URL+"/api/prompts/next?user=tester&tree=mythic-tragedy-apprenticeship", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("prompt next: %v", err)
	}
	defer promptResp.Body.Close()
	if promptResp.StatusCode != http.StatusOK {
		t.Fatalf("prompt status: %d", promptResp.StatusCode)
	}
	var promptPayload struct {
		Exercise struct {
			ID int64 `json:"id"`
		} `json:"exercise"`
	}
	if err := json.NewDecoder(promptResp.Body).Decode(&promptPayload); err != nil {
		t.Fatalf("decode prompt: %v", err)
	}

	exercisesResp, err := http.Get(testServer.URL + "/api/exercises?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("list exercises: %v", err)
	}
	defer exercisesResp.Body.Close()
	if exercisesResp.StatusCode != http.StatusOK {
		t.Fatalf("exercise list status: %d", exercisesResp.StatusCode)
	}

	exerciseResp, err := http.Get(testServer.URL + "/api/exercises/" + int64String(promptPayload.Exercise.ID) + "?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get exercise: %v", err)
	}
	defer exerciseResp.Body.Close()
	if exerciseResp.StatusCode != http.StatusOK {
		t.Fatalf("exercise get status: %d", exerciseResp.StatusCode)
	}

	submitResp, err := http.Post(testServer.URL+"/api/submissions?user=tester&tree=mythic-tragedy-apprenticeship", "application/json", strings.NewReader(`{"exercise_id":`+int64String(promptPayload.Exercise.ID)+`,"content":"The bell cracked as the prince chose the gate that had no hinge."}`))
	if err != nil {
		t.Fatalf("create submission: %v", err)
	}
	defer submitResp.Body.Close()
	if submitResp.StatusCode != http.StatusCreated {
		t.Fatalf("submission status: %d", submitResp.StatusCode)
	}
	var submissionPayload struct {
		Submission struct {
			ID int64 `json:"id"`
		} `json:"submission"`
	}
	if err := json.NewDecoder(submitResp.Body).Decode(&submissionPayload); err != nil {
		t.Fatalf("decode submission: %v", err)
	}

	submissionsResp, err := http.Get(testServer.URL + "/api/submissions?user=tester&tree=mythic-tragedy-apprenticeship&exercise_id=" + int64String(promptPayload.Exercise.ID))
	if err != nil {
		t.Fatalf("list submissions: %v", err)
	}
	defer submissionsResp.Body.Close()
	if submissionsResp.StatusCode != http.StatusOK {
		t.Fatalf("submission list status: %d", submissionsResp.StatusCode)
	}

	reviewResp, err := http.Post(testServer.URL+"/api/reviews?user=tester&tree=mythic-tragedy-apprenticeship", "application/json", strings.NewReader(`{"submission_id":`+int64String(submissionPayload.Submission.ID)+`}`))
	if err != nil {
		t.Fatalf("create review: %v", err)
	}
	defer reviewResp.Body.Close()
	if reviewResp.StatusCode != http.StatusCreated {
		t.Fatalf("review status: %d", reviewResp.StatusCode)
	}
	var reviewPayload struct {
		Review struct {
			ID int64 `json:"id"`
		} `json:"review"`
	}
	if err := json.NewDecoder(reviewResp.Body).Decode(&reviewPayload); err != nil {
		t.Fatalf("decode review: %v", err)
	}

	reviewsResp, err := http.Get(testServer.URL + "/api/reviews?user=tester&tree=mythic-tragedy-apprenticeship&submission_id=" + int64String(submissionPayload.Submission.ID))
	if err != nil {
		t.Fatalf("list reviews: %v", err)
	}
	defer reviewsResp.Body.Close()
	if reviewsResp.StatusCode != http.StatusOK {
		t.Fatalf("review list status: %d", reviewsResp.StatusCode)
	}

	singleReviewResp, err := http.Get(testServer.URL + "/api/reviews/" + int64String(reviewPayload.Review.ID) + "?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get review: %v", err)
	}
	defer singleReviewResp.Body.Close()
	if singleReviewResp.StatusCode != http.StatusOK {
		t.Fatalf("review get status: %d", singleReviewResp.StatusCode)
	}
	var singleReviewPayload struct {
		Review struct {
			Annotations []struct {
				Quote string `json:"quote"`
			} `json:"annotations"`
			Artifacts struct {
				AnalyzerReport map[string]any `json:"analyzer_report"`
				Recommendation map[string]any `json:"recommendation"`
				Annotations    []struct {
					Quote string `json:"quote"`
				} `json:"annotations"`
			} `json:"artifacts"`
		} `json:"review"`
	}
	if err := json.NewDecoder(singleReviewResp.Body).Decode(&singleReviewPayload); err != nil {
		t.Fatalf("decode single review: %v", err)
	}
	if len(singleReviewPayload.Review.Artifacts.AnalyzerReport) == 0 {
		t.Fatal("expected analyzer report artifact")
	}
	if len(singleReviewPayload.Review.Artifacts.Recommendation) == 0 {
		t.Fatal("expected recommendation artifact")
	}
	if len(singleReviewPayload.Review.Annotations) == 0 {
		t.Fatal("expected review annotations")
	}
	if len(singleReviewPayload.Review.Artifacts.Annotations) == 0 {
		t.Fatal("expected annotation artifacts")
	}
}

func TestPromptNextAcceptsSelectedTGOs(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	resp, err := http.Post(
		testServer.URL+"/api/prompts/next?user=tester&tree=mythic-tragedy-apprenticeship",
		"application/json",
		strings.NewReader(`{"tgo_codes":["scene-architecture","prose-precision","causal-clarity"]}`),
	)
	if err != nil {
		t.Fatalf("prompt next with selection: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("prompt status: %d", resp.StatusCode)
	}

	var payload struct {
		Exercise struct {
			TGOCodes []string `json:"tgo_codes"`
		} `json:"exercise"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode prompt response: %v", err)
	}
	if len(payload.Exercise.TGOCodes) != 3 {
		t.Fatalf("selected TGO count = %d", len(payload.Exercise.TGOCodes))
	}
	if payload.Exercise.TGOCodes[0] != "scene-architecture" || payload.Exercise.TGOCodes[2] != "causal-clarity" {
		t.Fatalf("expected selected TGO to persist, got %v", payload.Exercise.TGOCodes)
	}
}

func TestOnboardingCreatesAndActivatesGeneratedTrack(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	resp, err := http.Post(
		testServer.URL+"/api/onboarding?user=tester",
		"application/json",
		strings.NewReader(`{
			"writing_type":"thought leadership",
			"assignment_format":"blog post",
			"target_audience":"startup founders",
			"subject_matter":"AI product strategy",
			"experience_level":"intermediate",
			"desired_tone":"analytical and decisive",
			"biggest_weaknesses":["sentence economy","claim clarity"],
			"desired_outcomes":["write thought leadership with authority","develop a distinctive voice"],
			"difficulty_intensity":"steady",
			"writing_goals":"I want to publish stronger essays with clearer arguments."
		}`),
	)
	if err != nil {
		t.Fatalf("post onboarding: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("onboarding status: %d", resp.StatusCode)
	}
	var onboardingPayload struct {
		OnboardingComplete bool     `json:"onboarding_complete"`
		StarterTGOCodes    []string `json:"starter_tgo_codes"`
		RecommendedRegions []string `json:"recommended_regions"`
		Tree               struct {
			Slug string `json:"slug"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&onboardingPayload); err != nil {
		t.Fatalf("decode onboarding: %v", err)
	}
	if !onboardingPayload.OnboardingComplete {
		t.Fatal("expected onboarding to complete")
	}
	expectedTree := domain.GenerateTreeDefinition("tester", "Tester", domain.OnboardingProfile{
		WritingType:         "thought leadership",
		AssignmentFormat:    "blog post",
		TargetAudience:      "startup founders",
		SubjectMatter:       "AI product strategy",
		ExperienceLevel:     "intermediate",
		DesiredTone:         "analytical and decisive",
		BiggestWeaknesses:   []string{"sentence economy", "claim clarity"},
		DesiredOutcomes:     []string{"write thought leadership with authority", "develop a distinctive voice"},
		DifficultyIntensity: "steady",
		WritingGoals:        "I want to publish stronger essays with clearer arguments.",
	})
	if onboardingPayload.Tree.Slug != expectedTree.Slug {
		t.Fatalf("tree slug = %q", onboardingPayload.Tree.Slug)
	}
	if len(onboardingPayload.StarterTGOCodes) != 3 {
		t.Fatalf("starter codes = %#v", onboardingPayload.StarterTGOCodes)
	}
	if len(onboardingPayload.RecommendedRegions) == 0 {
		t.Fatal("expected recommended regions")
	}

	sessionResp, err := http.Get(testServer.URL + "/api/auth/session?user=tester")
	if err != nil {
		t.Fatalf("get auth session: %v", err)
	}
	defer sessionResp.Body.Close()
	if sessionResp.StatusCode != http.StatusOK {
		t.Fatalf("auth session status: %d", sessionResp.StatusCode)
	}
	var sessionPayload struct {
		OnboardingComplete bool   `json:"onboarding_complete"`
		ActiveTreeSlug     string `json:"active_tree_slug"`
		Context            struct {
			TreeSlug string `json:"tree_slug"`
		} `json:"context"`
	}
	if err := json.NewDecoder(sessionResp.Body).Decode(&sessionPayload); err != nil {
		t.Fatalf("decode auth session: %v", err)
	}
	if !sessionPayload.OnboardingComplete {
		t.Fatal("expected auth session to report onboarding complete")
	}
	if sessionPayload.ActiveTreeSlug != expectedTree.Slug {
		t.Fatalf("active tree slug = %q", sessionPayload.ActiveTreeSlug)
	}
	if sessionPayload.Context.TreeSlug != expectedTree.Slug {
		t.Fatalf("context tree slug = %q", sessionPayload.Context.TreeSlug)
	}

	getResp, err := http.Get(testServer.URL + "/api/onboarding?user=tester")
	if err != nil {
		t.Fatalf("get onboarding: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get onboarding status: %d", getResp.StatusCode)
	}
	var getPayload struct {
		Profile struct {
			AssignmentFormat string `json:"assignment_format"`
			TargetAudience   string `json:"target_audience"`
			SubjectMatter    string `json:"subject_matter"`
		} `json:"profile"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&getPayload); err != nil {
		t.Fatalf("decode onboarding get: %v", err)
	}
	if getPayload.Profile.AssignmentFormat != "blog post" {
		t.Fatalf("assignment format = %q", getPayload.Profile.AssignmentFormat)
	}
	if getPayload.Profile.TargetAudience != "startup founders" {
		t.Fatalf("target audience = %q", getPayload.Profile.TargetAudience)
	}
	if getPayload.Profile.SubjectMatter != "AI product strategy" {
		t.Fatalf("subject matter = %q", getPayload.Profile.SubjectMatter)
	}
}

func TestOnboardingUpdatesExistingProfilePromptSeedFields(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	initialPayload := `{
		"writing_type":"fiction",
		"assignment_format":"scene",
		"target_audience":"fantasy readers",
		"subject_matter":"mythic conflict",
		"experience_level":"intermediate",
		"desired_tone":"mythic and grave",
		"biggest_weaknesses":["scene architecture","word choice"],
		"desired_outcomes":["publish stronger fiction","develop a distinctive voice"],
		"difficulty_intensity":"steady",
		"writing_goals":"I want to write stronger scenes."
	}`
	resp, err := http.Post(
		testServer.URL+"/api/onboarding?user=tester",
		"application/json",
		strings.NewReader(initialPayload),
	)
	if err != nil {
		t.Fatalf("post initial onboarding: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("initial onboarding status: %d", resp.StatusCode)
	}

	updatePayload := `{
		"writing_type":"marketing",
		"assignment_format":"landing page",
		"target_audience":"product-led growth teams",
		"subject_matter":"B2B SaaS launches",
		"experience_level":"advanced",
		"desired_tone":"clear and persuasive",
		"biggest_weaknesses":["sentence economy","claim clarity"],
		"desired_outcomes":["improve professional communication","write thought leadership with authority"],
		"difficulty_intensity":"ambitious",
		"writing_goals":"I want sharper conversion-focused drafts."
	}`
	updateResp, err := http.Post(
		testServer.URL+"/api/onboarding?user=tester",
		"application/json",
		strings.NewReader(updatePayload),
	)
	if err != nil {
		t.Fatalf("post updated onboarding: %v", err)
	}
	defer updateResp.Body.Close()
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("updated onboarding status: %d", updateResp.StatusCode)
	}

	getResp, err := http.Get(testServer.URL + "/api/onboarding?user=tester")
	if err != nil {
		t.Fatalf("get onboarding after update: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get onboarding after update status: %d", getResp.StatusCode)
	}
	var getPayload struct {
		Profile struct {
			WritingType      string `json:"writing_type"`
			AssignmentFormat string `json:"assignment_format"`
			TargetAudience   string `json:"target_audience"`
			SubjectMatter    string `json:"subject_matter"`
			DesiredTone      string `json:"desired_tone"`
			WritingGoals     string `json:"writing_goals"`
		} `json:"profile"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&getPayload); err != nil {
		t.Fatalf("decode onboarding after update: %v", err)
	}
	if getPayload.Profile.WritingType != "marketing" {
		t.Fatalf("writing type = %q", getPayload.Profile.WritingType)
	}
	if getPayload.Profile.AssignmentFormat != "landing page" {
		t.Fatalf("assignment format = %q", getPayload.Profile.AssignmentFormat)
	}
	if getPayload.Profile.TargetAudience != "product-led growth teams" {
		t.Fatalf("target audience = %q", getPayload.Profile.TargetAudience)
	}
	if getPayload.Profile.SubjectMatter != "B2B SaaS launches" {
		t.Fatalf("subject matter = %q", getPayload.Profile.SubjectMatter)
	}
	if getPayload.Profile.DesiredTone != "clear and persuasive" {
		t.Fatalf("desired tone = %q", getPayload.Profile.DesiredTone)
	}
	if getPayload.Profile.WritingGoals != "I want sharper conversion-focused drafts." {
		t.Fatalf("writing goals = %q", getPayload.Profile.WritingGoals)
	}
}

func TestOnboardingValidationNamesMissingFields(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	resp, err := http.Post(
		testServer.URL+"/api/onboarding?user=tester",
		"application/json",
		strings.NewReader(`{
			"writing_type":"marketing",
			"assignment_format":"landing page",
			"target_audience":"",
			"subject_matter":"",
			"experience_level":"advanced",
			"desired_tone":"",
			"biggest_weaknesses":[],
			"desired_outcomes":[],
			"difficulty_intensity":"steady",
			"writing_goals":""
		}`),
	)
	if err != nil {
		t.Fatalf("post invalid onboarding: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("invalid onboarding status: %d", resp.StatusCode)
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode invalid onboarding response: %v", err)
	}
	expectedParts := []string{
		"target audience",
		"typical subject matter",
		"tone target",
		"biggest weaknesses",
		"desired outcomes",
		"writing goals",
	}
	for _, part := range expectedParts {
		if !strings.Contains(payload.Error, part) {
			t.Fatalf("expected error to contain %q, got %q", part, payload.Error)
		}
	}
}

func TestSkillGraphEndpoint(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/api/skill-graph?user=tester")
	if err != nil {
		t.Fatalf("get skill graph: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("skill graph status: %d", resp.StatusCode)
	}
	var payload struct {
		Graph struct {
			Slug    string `json:"slug"`
			Regions []struct {
				Slug string `json:"slug"`
			} `json:"regions"`
			Nodes []struct {
				Code           string   `json:"code"`
				SourceTreeSlug string   `json:"source_tree_slug"`
				Unlocks        []string `json:"unlocks"`
			} `json:"nodes"`
		} `json:"graph"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode skill graph: %v", err)
	}
	if payload.Graph.Slug != domain.GlobalSkillGraphSlug {
		t.Fatalf("graph slug = %q", payload.Graph.Slug)
	}
	if len(payload.Graph.Regions) < 9 {
		t.Fatalf("region count = %d", len(payload.Graph.Regions))
	}
	if len(payload.Graph.Nodes) < 450 {
		t.Fatalf("node count = %d", len(payload.Graph.Nodes))
	}
	if payload.Graph.Nodes[0].Code == "" || payload.Graph.Nodes[0].SourceTreeSlug == "" {
		t.Fatalf("unexpected first node: %#v", payload.Graph.Nodes[0])
	}
}

func TestAPIAuthMiddleware(t *testing.T) {
	testServer := newTestServerWithToken(t, "secret-token")
	defer testServer.Close()

	unauthorizedResp, err := http.Get(testServer.URL + "/api/dashboard?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("unauthorized request: %v", err)
	}
	defer unauthorizedResp.Body.Close()
	if unauthorizedResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d", unauthorizedResp.StatusCode)
	}

	req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/dashboard?user=tester&tree=mythic-tragedy-apprenticeship", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer secret-token")
	authorizedResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("authorized request: %v", err)
	}
	defer authorizedResp.Body.Close()
	if authorizedResp.StatusCode != http.StatusOK {
		t.Fatalf("authorized status = %d", authorizedResp.StatusCode)
	}
}

func TestTreeCreateRequiresAdminAuthorization(t *testing.T) {
	testServer := newTestServerWithToken(t, "secret-token")
	defer testServer.Close()

	resp, err := http.Post(testServer.URL+"/api/trees", "application/json", strings.NewReader(`{"slug":"blocked","title":"Blocked","tgos":[{"code":"one","title":"One","description":"One","stage":"core","stage_order":1}]}`))
	if err != nil {
		t.Fatalf("post tree: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("unauthorized tree create status = %d", resp.StatusCode)
	}

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/trees", strings.NewReader(`{"slug":"allowed","title":"Allowed","tgos":[{"code":"one","title":"One","description":"One","stage":"core","stage_order":1},{"code":"two","title":"Two","description":"Two","stage":"core","stage_order":2},{"code":"three","title":"Three","description":"Three","stage":"core","stage_order":3}]}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("Content-Type", "application/json")
	authResp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("authorized tree create: %v", err)
	}
	defer authResp.Body.Close()
	if authResp.StatusCode != http.StatusCreated {
		t.Fatalf("authorized tree create status = %d", authResp.StatusCode)
	}

	adminsReq, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/admins", nil)
	if err != nil {
		t.Fatalf("admins request: %v", err)
	}
	adminsReq.Header.Set("Authorization", "Bearer secret-token")
	adminsResp, err := http.DefaultClient.Do(adminsReq)
	if err != nil {
		t.Fatalf("admins response: %v", err)
	}
	defer adminsResp.Body.Close()
	if adminsResp.StatusCode != http.StatusOK {
		t.Fatalf("admins status = %d", adminsResp.StatusCode)
	}
}

func TestKratosWhoamiAuthMiddleware(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	tree, err := harness.Store.TreeBySlug(context.Background(), "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("lookup tree: %v", err)
	}
	testerExerciseID, err := harness.Store.SaveExercise(context.Background(), domain.Exercise{
		UserID:             user.ID,
		TreeID:             tree.ID,
		Title:              "Hidden Exercise",
		Brief:              "Secret",
		Constraints:        []string{"single scene"},
		FocusSkills:        []string{"scene architecture"},
		TGOCodes:           []string{"scene-architecture"},
		SuccessCriteria:    []string{"a turn occurs"},
		GenerationKind:     "deterministic",
		SourceSubmissionID: 0,
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	testerSubmissionID, err := harness.Store.SaveSubmission(context.Background(), domain.Submission{
		UserID:     user.ID,
		TreeID:     tree.ID,
		ExerciseID: testerExerciseID,
		Content:    "The king refused the warning.",
		WordCount:  5,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}

	kratos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/sessions/whoami" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("X-Session-Token") != "session-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"identity": map[string]any{
				"id": "7b5f6fc1-36d3-4f48-8ef0-7cfa2d4fb613",
				"traits": map[string]any{
					"email": "writer@example.com",
					"name": map[string]any{
						"first": "Writer",
						"last":  "Coach",
					},
				},
			},
		})
	}))
	defer kratos.Close()

	testServer := newTestServerWithStore(t, harness.Store, "", kratos.URL)
	defer testServer.Close()

	req, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/context?tree=mythic-tragedy-apprenticeship", nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("X-Session-Token", "session-token")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var payload struct {
		UserSlug string `json:"user_slug"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if payload.UserSlug != "kratos-7b5f6fc136d3" {
		t.Fatalf("user slug = %q", payload.UserSlug)
	}

	sessionReq, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/auth/session?tree=mythic-tragedy-apprenticeship", nil)
	if err != nil {
		t.Fatalf("session request: %v", err)
	}
	sessionReq.Header.Set("X-Session-Token", "session-token")
	sessionResp, err := http.DefaultClient.Do(sessionReq)
	if err != nil {
		t.Fatalf("session response: %v", err)
	}
	defer sessionResp.Body.Close()
	if sessionResp.StatusCode != http.StatusOK {
		t.Fatalf("session status = %d", sessionResp.StatusCode)
	}

	foreignReq, err := http.NewRequest(http.MethodGet, testServer.URL+"/api/submissions/"+int64String(testerSubmissionID)+"?user=tester&tree=mythic-tragedy-apprenticeship", nil)
	if err != nil {
		t.Fatalf("foreign request: %v", err)
	}
	foreignReq.Header.Set("X-Session-Token", "session-token")
	foreignResp, err := http.DefaultClient.Do(foreignReq)
	if err != nil {
		t.Fatalf("foreign response: %v", err)
	}
	defer foreignResp.Body.Close()
	if foreignResp.StatusCode != http.StatusNotFound {
		t.Fatalf("foreign status = %d", foreignResp.StatusCode)
	}
}

func TestKratosAdminEmailCanManageTrees(t *testing.T) {
	kratos := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"identity": map[string]any{
				"id": "admin-user-1",
				"traits": map[string]any{
					"email": "writer@example.com",
					"name":  map[string]any{"first": "Writer"},
				},
			},
		})
	}))
	defer kratos.Close()

	harness := newTestHarnessWithAuth(t, "", "")
	if err := harness.Store.EnsureAdminEmails(context.Background(), []string{"writer@example.com"}); err != nil {
		t.Fatalf("seed admin email: %v", err)
	}
	cfg := config.Default(t.TempDir())
	cfg.KratosPublicURL = kratos.URL
	testServer := newTestServerWithConfig(t, harness.Store, cfg)
	defer testServer.Close()

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/trees", strings.NewReader(`{"slug":"kratos-admin","title":"Kratos Admin","tgos":[{"code":"one-a","title":"One","description":"One","stage":"core","stage_order":1},{"code":"two-a","title":"Two","description":"Two","stage":"core","stage_order":2},{"code":"three-a","title":"Three","description":"Three","stage":"core","stage_order":3}]}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("X-Session-Token", "session-token")
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("kratos admin request: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("kratos admin status = %d", resp.StatusCode)
	}

	addAdminReq, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/admins", strings.NewReader(`{"email":"second-admin@example.com"}`))
	if err != nil {
		t.Fatalf("new admin request: %v", err)
	}
	addAdminReq.Header.Set("X-Session-Token", "session-token")
	addAdminReq.Header.Set("Content-Type", "application/json")
	addAdminResp, err := http.DefaultClient.Do(addAdminReq)
	if err != nil {
		t.Fatalf("add admin response: %v", err)
	}
	defer addAdminResp.Body.Close()
	if addAdminResp.StatusCode != http.StatusCreated {
		t.Fatalf("add admin status = %d", addAdminResp.StatusCode)
	}
}

func newTestServer(t *testing.T) *httptest.Server {
	return newTestServerWithAuth(t, "", "")
}

func newTestServerWithToken(t *testing.T, apiToken string) *httptest.Server {
	return newTestServerWithAuth(t, apiToken, "")
}

func newTestServerWithAuth(t *testing.T, apiToken, kratosPublicURL string) *httptest.Server {
	harness := newTestHarnessWithAuth(t, apiToken, kratosPublicURL)
	cfg := config.Default(t.TempDir())
	cfg.APIToken = apiToken
	cfg.KratosPublicURL = kratosPublicURL
	return newTestServerWithConfig(t, harness.Store, cfg)
}

type testHarness struct {
	Store *db.Store
}

func newTestHarnessWithAuth(t *testing.T, apiToken, kratosPublicURL string) testHarness {
	t.Helper()
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "test.db"))
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
	if _, _, _, err := store.EnsureDefaultUserTree(ctx, "tester", "Tester", "mythic-tragedy-apprenticeship"); err != nil {
		t.Fatalf("default user tree: %v", err)
	}
	t.Cleanup(func() {
		_ = store.Close()
	})
	return testHarness{Store: store}
}

func newTestServerWithStore(t *testing.T, store *db.Store, apiToken, kratosPublicURL string) *httptest.Server {
	t.Helper()
	cfg := config.Default(t.TempDir())
	cfg.APIToken = apiToken
	cfg.KratosPublicURL = kratosPublicURL
	return newTestServerWithConfig(t, store, cfg)
}

func newTestServerWithConfig(t *testing.T, store *db.Store, cfg config.Config) *httptest.Server {
	t.Helper()

	server := Server{
		Config:     cfg,
		Store:      store,
		Prompts:    prompt.NewService(nil),
		Reviews:    review.NewService(nil, analyzer.Service{}),
		Curriculum: curriculum.NewService(),
	}
	testServer := httptest.NewServer(server.routes())
	t.Cleanup(func() {
		testServer.Close()
	})
	return testServer
}

func int64String(value int64) string {
	return strconv.FormatInt(value, 10)
}
