package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/curriculum"
	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/prompt"
	"github.com/tomasino/writing-coach/internal/review"
	"github.com/tomasino/writing-coach/internal/secrets"
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

func TestOnboardingOptionsEndpoint(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/api/onboarding/options")
	if err != nil {
		t.Fatalf("get onboarding options: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("options status: %d", resp.StatusCode)
	}

	var payload struct {
		WritingDomains []struct {
			Value string `json:"value"`
			Label string `json:"label"`
		} `json:"writing_domains"`
		AssignmentFormats []struct {
			Value string `json:"value"`
		} `json:"assignment_formats"`
		Weaknesses []struct {
			Value string `json:"value"`
		} `json:"weaknesses"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode onboarding options: %v", err)
	}
	if len(payload.WritingDomains) < 10 {
		t.Fatalf("writing domains = %#v", payload.WritingDomains)
	}
	if payload.WritingDomains[0].Value == "" || payload.WritingDomains[0].Label == "" {
		t.Fatalf("first writing domain = %#v", payload.WritingDomains[0])
	}
	if len(payload.AssignmentFormats) == 0 || payload.AssignmentFormats[0].Value == "" {
		t.Fatalf("assignment formats = %#v", payload.AssignmentFormats)
	}
	if len(payload.Weaknesses) == 0 || payload.Weaknesses[0].Value == "" {
		t.Fatalf("weaknesses = %#v", payload.Weaknesses)
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
		Exercise exerciseResponse `json:"exercise"`
	}
	if err := json.NewDecoder(promptResp.Body).Decode(&promptPayload); err != nil {
		t.Fatalf("decode prompt: %v", err)
	}
	acceptResp, err := http.Post(
		testServer.URL+"/api/prompts/accept?user=tester&tree=mythic-tragedy-apprenticeship",
		"application/json",
		strings.NewReader(mustJSONString(map[string]any{
			"title":            promptPayload.Exercise.Title,
			"brief":            promptPayload.Exercise.Brief,
			"constraints":      promptPayload.Exercise.Constraints,
			"focus_skills":     promptPayload.Exercise.FocusSkills,
			"tgo_codes":        promptPayload.Exercise.TGOCodes,
			"success_criteria": promptPayload.Exercise.SuccessCriteria,
			"generation_kind":  promptPayload.Exercise.GenerationKind,
			"provider_note":    promptPayload.Exercise.ProviderNote,
		})),
	)
	if err != nil {
		t.Fatalf("accept prompt: %v", err)
	}
	defer acceptResp.Body.Close()
	if acceptResp.StatusCode != http.StatusOK {
		t.Fatalf("accept prompt status: %d", acceptResp.StatusCode)
	}
	if err := json.NewDecoder(acceptResp.Body).Decode(&promptPayload); err != nil {
		t.Fatalf("decode accepted prompt: %v", err)
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
	if reviewResp.StatusCode != http.StatusAccepted {
		t.Fatalf("review status: %d", reviewResp.StatusCode)
	}
	var reviewPayload struct {
		Job struct {
			ID     int64  `json:"id"`
			Status string `json:"status"`
		} `json:"job"`
	}
	if err := json.NewDecoder(reviewResp.Body).Decode(&reviewPayload); err != nil {
		t.Fatalf("decode review: %v", err)
	}
	if reviewPayload.Job.ID == 0 || reviewPayload.Job.Status == "" {
		t.Fatalf("expected queued review job, got %+v", reviewPayload.Job)
	}

	reviewID := waitForReview(t, testServer.URL, submissionPayload.Submission.ID)

	reviewsResp, err := http.Get(testServer.URL + "/api/reviews?user=tester&tree=mythic-tragedy-apprenticeship&submission_id=" + int64String(submissionPayload.Submission.ID))
	if err != nil {
		t.Fatalf("list reviews: %v", err)
	}
	defer reviewsResp.Body.Close()
	if reviewsResp.StatusCode != http.StatusOK {
		t.Fatalf("review list status: %d", reviewsResp.StatusCode)
	}

	singleReviewResp, err := http.Get(testServer.URL + "/api/reviews/" + int64String(reviewID) + "?user=tester&tree=mythic-tragedy-apprenticeship")
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

func TestAssignmentTimelineEndpoint(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	testServer := newTestServerWithStore(t, harness.Store, "", "")
	defer testServer.Close()

	ctx := context.Background()
	user, err := harness.Store.UserBySlug(ctx, "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	tree, err := harness.Store.TreeBySlug(ctx, "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("lookup tree: %v", err)
	}

	rootExerciseID, err := harness.Store.SaveExercise(ctx, domain.Exercise{
		UserID:          user.ID,
		TreeID:          tree.ID,
		Title:           "Original Assignment",
		Brief:           "Draft the omen scene.",
		Constraints:     []string{"under 600 words"},
		FocusSkills:     []string{"causal clarity"},
		TGOCodes:        []string{"causal-clarity"},
		SuccessCriteria: []string{"the sequence is easy to follow"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save root exercise: %v", err)
	}
	rootSubmissionID, err := harness.Store.SaveSubmission(ctx, domain.Submission{
		UserID:     user.ID,
		TreeID:     tree.ID,
		ExerciseID: rootExerciseID,
		Content:    "The priest heard the bell before the gate split.",
		WordCount:  10,
	})
	if err != nil {
		t.Fatalf("save root submission: %v", err)
	}
	if _, err := harness.Store.SaveReview(ctx, domain.Review{
		UserID:           user.ID,
		TreeID:           tree.ID,
		SubmissionID:     rootSubmissionID,
		ReviewKind:       "coach",
		Summary:          "The draft establishes the omen clearly.",
		Strengths:        []string{"clear inciting image"},
		Weaknesses:       []string{"consequences arrive too abstractly"},
		AnalyzerFindings: []string{"one passive construction"},
		NextFocus:        "make the consequence concrete",
		MetricWordCount:  10,
		TGOAssessments: []domain.TGOAssessment{
			{TGOCode: "causal-clarity", Status: "developing", Evidence: "The omen lands before the consequence clarifies."},
		},
	}, []domain.SkillScore{{SubmissionID: rootSubmissionID, Skill: "causal clarity", Score: 3}}); err != nil {
		t.Fatalf("save root review: %v", err)
	}

	revisionExerciseID, err := harness.Store.SaveExercise(ctx, domain.Exercise{
		UserID:             user.ID,
		TreeID:             tree.ID,
		Title:              "Revision Assignment",
		Brief:              "Revise the omen scene with sharper consequence.",
		Constraints:        []string{"preserve the same core event"},
		FocusSkills:        []string{"causal clarity"},
		TGOCodes:           []string{"causal-clarity"},
		SuccessCriteria:    []string{"the consequence becomes concrete"},
		GenerationKind:     "revision",
		SourceSubmissionID: rootSubmissionID,
	})
	if err != nil {
		t.Fatalf("save revision exercise: %v", err)
	}
	revisionSubmissionID, err := harness.Store.SaveSubmission(ctx, domain.Submission{
		UserID:             user.ID,
		TreeID:             tree.ID,
		ExerciseID:         revisionExerciseID,
		ParentSubmissionID: rootSubmissionID,
		Content:            "The priest heard the bell, then watched the gate split and pin the guard beneath it.",
		WordCount:          15,
	})
	if err != nil {
		t.Fatalf("save revision submission: %v", err)
	}
	if _, err := harness.Store.SaveReview(ctx, domain.Review{
		UserID:           user.ID,
		TreeID:           tree.ID,
		SubmissionID:     revisionSubmissionID,
		ReviewKind:       "coach",
		Summary:          "The revision makes the consequence visible.",
		Strengths:        []string{"cause and effect read cleanly"},
		Weaknesses:       []string{"ending image could resonate longer"},
		AnalyzerFindings: []string{},
		NextFocus:        "strengthen the closing image",
		MetricWordCount:  15,
		TGOAssessments: []domain.TGOAssessment{
			{TGOCode: "causal-clarity", Status: "mastered", Evidence: "The consequence now lands in-scene."},
		},
	}, []domain.SkillScore{{SubmissionID: revisionSubmissionID, Skill: "causal clarity", Score: 4}}); err != nil {
		t.Fatalf("save revision review: %v", err)
	}
	if _, err := harness.Store.SaveExercise(ctx, domain.Exercise{
		UserID:          user.ID,
		TreeID:          tree.ID,
		Title:           "New Assignment",
		Brief:           "Start something new.",
		Constraints:     []string{"under 600 words"},
		FocusSkills:     []string{"scene architecture"},
		TGOCodes:        []string{"scene-architecture"},
		SuccessCriteria: []string{"the scene advances cleanly"},
		GenerationKind:  "deterministic",
	}); err != nil {
		t.Fatalf("save later exercise: %v", err)
	}

	resp, err := http.Get(testServer.URL + "/api/assignments/" + int64String(revisionExerciseID) + "?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get assignment timeline: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("assignment timeline status: %d", resp.StatusCode)
	}

	var payload struct {
		Assignment struct {
			RootExerciseID    int64 `json:"root_exercise_id"`
			CurrentExerciseID int64 `json:"current_exercise_id"`
			IsCurrent         bool  `json:"is_current"`
			Steps             []struct {
				Kind         string `json:"kind"`
				ExerciseID   int64  `json:"exercise_id"`
				SubmissionID int64  `json:"submission_id"`
				ReviewID     int64  `json:"review_id"`
				Label        string `json:"label"`
			} `json:"steps"`
		} `json:"assignment"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode assignment timeline: %v", err)
	}
	if payload.Assignment.RootExerciseID != rootExerciseID {
		t.Fatalf("root exercise id = %d", payload.Assignment.RootExerciseID)
	}
	if payload.Assignment.CurrentExerciseID != revisionExerciseID {
		t.Fatalf("current exercise id = %d", payload.Assignment.CurrentExerciseID)
	}
	if payload.Assignment.IsCurrent {
		t.Fatal("expected assignment chain to be historical after a newer assignment was created")
	}
	if len(payload.Assignment.Steps) != 6 {
		t.Fatalf("step count = %d", len(payload.Assignment.Steps))
	}
	if payload.Assignment.Steps[0].Kind != "exercise" || payload.Assignment.Steps[0].ExerciseID != rootExerciseID {
		t.Fatalf("first step = %#v", payload.Assignment.Steps[0])
	}
	if payload.Assignment.Steps[2].Kind != "review" || payload.Assignment.Steps[2].Label != "Feedback 1" {
		t.Fatalf("third step = %#v", payload.Assignment.Steps[2])
	}
	if payload.Assignment.Steps[3].Kind != "exercise" || payload.Assignment.Steps[3].ExerciseID != revisionExerciseID {
		t.Fatalf("fourth step = %#v", payload.Assignment.Steps[3])
	}
	if payload.Assignment.Steps[5].Kind != "review" || payload.Assignment.Steps[5].SubmissionID != revisionSubmissionID {
		t.Fatalf("last step = %#v", payload.Assignment.Steps[5])
	}
}

func TestAssignmentsListEndpoint(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	testServer := newTestServerWithStore(t, harness.Store, "", "")
	defer testServer.Close()

	ctx := context.Background()
	user, err := harness.Store.UserBySlug(ctx, "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	tree, err := harness.Store.TreeBySlug(ctx, "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("lookup tree: %v", err)
	}

	pastExerciseID, err := harness.Store.SaveExercise(ctx, domain.Exercise{
		UserID:          user.ID,
		TreeID:          tree.ID,
		Title:           "Past Assignment",
		Brief:           "Write the omen.",
		Constraints:     []string{"under 500 words"},
		FocusSkills:     []string{"causal clarity"},
		TGOCodes:        []string{"causal-clarity"},
		SuccessCriteria: []string{"clear cause and effect"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save past exercise: %v", err)
	}
	pastSubmissionID, err := harness.Store.SaveSubmission(ctx, domain.Submission{
		UserID:     user.ID,
		TreeID:     tree.ID,
		ExerciseID: pastExerciseID,
		Content:    "The bell rang and the gate split.",
		WordCount:  7,
	})
	if err != nil {
		t.Fatalf("save past submission: %v", err)
	}
	if _, err := harness.Store.SaveReview(ctx, domain.Review{
		UserID:           user.ID,
		TreeID:           tree.ID,
		SubmissionID:     pastSubmissionID,
		ReviewKind:       "coach",
		Summary:          "Past review.",
		Strengths:        []string{"clear event"},
		Weaknesses:       []string{"thin consequence"},
		AnalyzerFindings: []string{},
		NextFocus:        "make consequence vivid",
		MetricWordCount:  7,
		TGOAssessments: []domain.TGOAssessment{
			{TGOCode: "causal-clarity", Status: "developing", Evidence: "The cause is present."},
		},
	}, nil); err != nil {
		t.Fatalf("save past review: %v", err)
	}

	currentExerciseID, err := harness.Store.SaveExercise(ctx, domain.Exercise{
		UserID:          user.ID,
		TreeID:          tree.ID,
		Title:           "Current Assignment",
		Brief:           "Write the aftermath.",
		Constraints:     []string{"under 500 words"},
		FocusSkills:     []string{"scene architecture"},
		TGOCodes:        []string{"scene-architecture"},
		SuccessCriteria: []string{"scene moves cleanly"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save current exercise: %v", err)
	}

	resp, err := http.Get(testServer.URL + "/api/assignments?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get assignments list: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("assignments list status: %d", resp.StatusCode)
	}

	var payload struct {
		Assignments []struct {
			RootExerciseID    int64  `json:"root_exercise_id"`
			CurrentExerciseID int64  `json:"current_exercise_id"`
			Title             string `json:"title"`
			IsCurrent         bool   `json:"is_current"`
			ReviewCount       int    `json:"review_count"`
		} `json:"assignments"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode assignments list: %v", err)
	}
	if len(payload.Assignments) != 2 {
		t.Fatalf("assignment count = %d", len(payload.Assignments))
	}
	if payload.Assignments[0].CurrentExerciseID != currentExerciseID || !payload.Assignments[0].IsCurrent {
		t.Fatalf("first assignment = %#v", payload.Assignments[0])
	}
	if payload.Assignments[1].RootExerciseID != pastExerciseID || payload.Assignments[1].ReviewCount != 1 {
		t.Fatalf("second assignment = %#v", payload.Assignments[1])
	}
}

func TestAISettingsLifecycleEndpoint(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	testServer := newTestServerWithStore(t, harness.Store, "", "")
	defer testServer.Close()

	getResp, err := http.Get(testServer.URL + "/api/ai/settings?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get ai settings: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("initial ai settings status: %d", getResp.StatusCode)
	}
	var initial struct {
		Settings struct {
			Provider                         string `json:"provider"`
			HasKey                           bool   `json:"has_key"`
			EffectiveProvider                string `json:"effective_provider"`
			SystemFallback                   bool   `json:"system_fallback"`
			PersonalProviderStorageAvailable bool   `json:"personal_provider_storage_available"`
			Ready                            bool   `json:"ready"`
		} `json:"settings"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&initial); err != nil {
		t.Fatalf("decode initial ai settings: %v", err)
	}
	if initial.Settings.Provider != "" || initial.Settings.HasKey {
		t.Fatalf("unexpected initial settings: %#v", initial.Settings)
	}
	if initial.Settings.EffectiveProvider != "system/openai" {
		t.Fatalf("effective provider = %q", initial.Settings.EffectiveProvider)
	}
	if !initial.Settings.PersonalProviderStorageAvailable {
		t.Fatal("expected personal provider storage to be available in default test config")
	}

	var authHeader string
	fakeProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader = r.Header.Get("Authorization")
		if r.URL.Path != "/models" {
			t.Fatalf("provider path = %q", r.URL.Path)
		}
		_, _ = io.WriteString(w, `{"data":[]}`)
	}))
	defer fakeProvider.Close()

	validateReq, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/ai/settings/validate?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(fmt.Sprintf(`{"provider":"openai","api_key":"sk-test-1234","base_url_override":%q,"prompt_model_override":"gpt-5.4","review_model_override":"gpt-5.4-mini","enabled":true}`, fakeProvider.URL)))
	if err != nil {
		t.Fatalf("new validate request: %v", err)
	}
	validateReq.Header.Set("Content-Type", "application/json")
	validateResp, err := http.DefaultClient.Do(validateReq)
	if err != nil {
		t.Fatalf("validate ai settings: %v", err)
	}
	defer validateResp.Body.Close()
	if validateResp.StatusCode != http.StatusOK {
		t.Fatalf("validate ai settings status: %d", validateResp.StatusCode)
	}
	var validation struct {
		Valid    bool `json:"valid"`
		Settings struct {
			KeyLast4 string `json:"key_last4"`
			HasKey   bool   `json:"has_key"`
		} `json:"settings"`
	}
	if err := json.NewDecoder(validateResp.Body).Decode(&validation); err != nil {
		t.Fatalf("decode validation response: %v", err)
	}
	if !validation.Valid || !validation.Settings.HasKey || validation.Settings.KeyLast4 != "1234" {
		t.Fatalf("validation payload = %#v", validation)
	}
	if authHeader != "Bearer sk-test-1234" {
		t.Fatalf("authorization = %q", authHeader)
	}

	putReq, err := http.NewRequest(http.MethodPut, testServer.URL+"/api/ai/settings?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(fmt.Sprintf(`{"provider":"openai","api_key":"sk-test-1234","base_url_override":%q,"prompt_model_override":"gpt-5.4","review_model_override":"gpt-5.4-mini","enabled":true}`, fakeProvider.URL)))
	if err != nil {
		t.Fatalf("new ai settings request: %v", err)
	}
	putReq.Header.Set("Content-Type", "application/json")
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		t.Fatalf("save ai settings: %v", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode != http.StatusOK {
		t.Fatalf("save ai settings status: %d", putResp.StatusCode)
	}
	var stored struct {
		Settings struct {
			Provider            string `json:"provider"`
			KeyLast4            string `json:"key_last4"`
			HasKey              bool   `json:"has_key"`
			PromptModelOverride string `json:"prompt_model_override"`
			ReviewModelOverride string `json:"review_model_override"`
			EffectiveProvider   string `json:"effective_provider"`
		} `json:"settings"`
	}
	if err := json.NewDecoder(putResp.Body).Decode(&stored); err != nil {
		t.Fatalf("decode stored ai settings: %v", err)
	}
	if stored.Settings.Provider != "openai" || stored.Settings.KeyLast4 != "1234" || !stored.Settings.HasKey {
		t.Fatalf("stored settings payload = %#v", stored.Settings)
	}
	if stored.Settings.EffectiveProvider != "user" {
		t.Fatalf("effective provider = %q", stored.Settings.EffectiveProvider)
	}

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	record, err := harness.Store.AIProviderSettingsByUserID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("stored provider settings lookup: %v", err)
	}
	if record.APIKeyEncrypted == "" || record.APIKeyEncrypted == "sk-test-1234" {
		t.Fatalf("expected encrypted key, got %#v", record)
	}

	updateReq, err := http.NewRequest(http.MethodPut, testServer.URL+"/api/ai/settings?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(fmt.Sprintf(`{"provider":"openai","api_key":"","base_url_override":%q,"prompt_model_override":"gpt-5-mini","review_model_override":"gpt-5-mini","enabled":false}`, fakeProvider.URL)))
	if err != nil {
		t.Fatalf("new update request: %v", err)
	}
	updateReq.Header.Set("Content-Type", "application/json")
	updateResp, err := http.DefaultClient.Do(updateReq)
	if err != nil {
		t.Fatalf("update ai settings: %v", err)
	}
	defer updateResp.Body.Close()
	if updateResp.StatusCode != http.StatusOK {
		t.Fatalf("update ai settings status: %d", updateResp.StatusCode)
	}
	record, err = harness.Store.AIProviderSettingsByUserID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("lookup updated provider settings: %v", err)
	}
	if record.Enabled {
		t.Fatal("expected provider settings to be disabled")
	}
	if record.APIKeyEncrypted == "" {
		t.Fatal("expected saved key to be preserved")
	}
	if record.APIKeyLast4 != "1234" {
		t.Fatalf("key last4 = %q", record.APIKeyLast4)
	}

	secondUserID, _, _, err := harness.Store.EnsureDefaultUserTree(context.Background(), "other", "Other", "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("ensure other user: %v", err)
	}
	if _, err := harness.Store.AIProviderSettingsByUserID(context.Background(), secondUserID); !db.IsNotFound(err) {
		t.Fatalf("expected other user to have no provider settings, got %v", err)
	}

	deleteReq, err := http.NewRequest(http.MethodDelete, testServer.URL+"/api/ai/settings?user=tester&tree=mythic-tragedy-apprenticeship", nil)
	if err != nil {
		t.Fatalf("new delete request: %v", err)
	}
	deleteResp, err := http.DefaultClient.Do(deleteReq)
	if err != nil {
		t.Fatalf("delete ai settings: %v", err)
	}
	defer deleteResp.Body.Close()
	if deleteResp.StatusCode != http.StatusOK {
		t.Fatalf("delete ai settings status: %d", deleteResp.StatusCode)
	}
	if _, err := harness.Store.AIProviderSettingsByUserID(context.Background(), user.ID); !db.IsNotFound(err) {
		t.Fatalf("expected deleted provider settings, got %v", err)
	}
}

func TestAISettingsDisablePersonalProviderStorageWithoutSecret(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	cfg := config.Default(t.TempDir())
	testServer := newTestServerWithConfig(t, harness.Store, cfg)
	defer testServer.Close()

	getResp, err := http.Get(testServer.URL + "/api/ai/settings?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get ai settings: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get ai settings status: %d", getResp.StatusCode)
	}
	var getPayload struct {
		Settings struct {
			PersonalProviderStorageAvailable bool `json:"personal_provider_storage_available"`
			SystemFallback                   bool `json:"system_fallback"`
		} `json:"settings"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&getPayload); err != nil {
		t.Fatalf("decode ai settings: %v", err)
	}
	if getPayload.Settings.PersonalProviderStorageAvailable {
		t.Fatal("expected personal provider storage to be unavailable without AI key secret")
	}

	validateReq, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/ai/settings/validate?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(`{"provider":"openai","api_key":"sk-test-1234","enabled":true}`))
	if err != nil {
		t.Fatalf("new validate request: %v", err)
	}
	validateReq.Header.Set("Content-Type", "application/json")
	validateResp, err := http.DefaultClient.Do(validateReq)
	if err != nil {
		t.Fatalf("validate ai settings: %v", err)
	}
	defer validateResp.Body.Close()
	if validateResp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("validate status = %d", validateResp.StatusCode)
	}

	putReq, err := http.NewRequest(http.MethodPut, testServer.URL+"/api/ai/settings?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(`{"provider":"openai","api_key":"sk-test-1234","enabled":true}`))
	if err != nil {
		t.Fatalf("new put request: %v", err)
	}
	putReq.Header.Set("Content-Type", "application/json")
	putResp, err := http.DefaultClient.Do(putReq)
	if err != nil {
		t.Fatalf("save ai settings: %v", err)
	}
	defer putResp.Body.Close()
	if putResp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("put status = %d", putResp.StatusCode)
	}

	sessionResp, err := http.Get(testServer.URL + "/api/auth/session?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get auth session: %v", err)
	}
	defer sessionResp.Body.Close()
	if sessionResp.StatusCode != http.StatusOK {
		t.Fatalf("auth session status = %d", sessionResp.StatusCode)
	}
	var sessionPayload struct {
		AIPersonalProviderStorageAvailable bool `json:"ai_personal_provider_storage_available"`
	}
	if err := json.NewDecoder(sessionResp.Body).Decode(&sessionPayload); err != nil {
		t.Fatalf("decode auth session: %v", err)
	}
	if sessionPayload.AIPersonalProviderStorageAvailable {
		t.Fatal("expected auth session to report unavailable personal provider storage")
	}
}

func TestAISettingsValidateRejectsInvalidCredentials(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	fakeProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error":{"message":"invalid api key"}}`)
	}))
	defer fakeProvider.Close()

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/ai/settings/validate?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(fmt.Sprintf(`{"provider":"openai","api_key":"sk-bad","base_url_override":%q,"enabled":true}`, fakeProvider.URL)))
	if err != nil {
		t.Fatalf("new validate request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("validate bad credentials: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode validation error: %v", err)
	}
	if !strings.Contains(strings.ToLower(payload.Error), "rejected this api key") {
		t.Fatalf("unexpected error = %q", payload.Error)
	}
}

func TestAISettingsValidateRateLimitsRepeatedChecksPerUser(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	cfg.AIValidateLimitPerMinute = 1
	cfg.AIValidateGlobalLimitPerMinute = 10
	testServer := newTestServerWithConfig(t, harness.Store, cfg)
	defer testServer.Close()

	hits := 0
	fakeProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":[{"id":"gpt-5-mini"}]}`)
	}))
	defer fakeProvider.Close()

	requestBody := fmt.Sprintf(`{"provider":"openai","api_key":"sk-test-1234","base_url_override":%q,"enabled":true}`, fakeProvider.URL)

	firstReq, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/ai/settings/validate?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("new first validate request: %v", err)
	}
	firstReq.Header.Set("Content-Type", "application/json")
	firstResp, err := http.DefaultClient.Do(firstReq)
	if err != nil {
		t.Fatalf("first validate request: %v", err)
	}
	defer firstResp.Body.Close()
	if firstResp.StatusCode != http.StatusOK {
		t.Fatalf("first validate status = %d", firstResp.StatusCode)
	}

	secondReq, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/ai/settings/validate?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("new second validate request: %v", err)
	}
	secondReq.Header.Set("Content-Type", "application/json")
	secondResp, err := http.DefaultClient.Do(secondReq)
	if err != nil {
		t.Fatalf("second validate request: %v", err)
	}
	defer secondResp.Body.Close()
	if secondResp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("second validate status = %d", secondResp.StatusCode)
	}
	if secondResp.Header.Get("Retry-After") == "" {
		t.Fatal("expected retry-after header")
	}

	var payload struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(secondResp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode rate limit error: %v", err)
	}
	if !strings.Contains(strings.ToLower(payload.Error), "rate-limiting") {
		t.Fatalf("unexpected rate limit error = %q", payload.Error)
	}
	if hits != 1 {
		t.Fatalf("provider validation hits = %d", hits)
	}
}

func TestAISettingsValidateMapsQuotaError(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	fakeProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = io.WriteString(w, `{"error":{"message":"insufficient_quota"}}`)
	}))
	defer fakeProvider.Close()

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/ai/settings/validate?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(fmt.Sprintf(`{"provider":"openai","api_key":"sk-quota","base_url_override":%q,"enabled":true}`, fakeProvider.URL)))
	if err != nil {
		t.Fatalf("new validate request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("validate quota error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadGateway {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode quota error: %v", err)
	}
	if !strings.Contains(strings.ToLower(payload.Error), "out of quota") {
		t.Fatalf("unexpected error = %q", payload.Error)
	}
}

func TestAISettingsValidateRateLimitDoesNotBlockOtherUsers(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	if _, _, _, err := harness.Store.EnsureDefaultUserTree(context.Background(), "second-writer", "Second Writer", "mythic-tragedy-apprenticeship"); err != nil {
		t.Fatalf("ensure second user tree: %v", err)
	}
	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	cfg.AIValidateLimitPerMinute = 1
	cfg.AIValidateGlobalLimitPerMinute = 10
	testServer := newTestServerWithConfig(t, harness.Store, cfg)
	defer testServer.Close()

	hits := 0
	fakeProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":[{"id":"gpt-5-mini"}]}`)
	}))
	defer fakeProvider.Close()

	requestBody := fmt.Sprintf(`{"provider":"openai","api_key":"sk-test-1234","base_url_override":%q,"enabled":true}`, fakeProvider.URL)

	requestFor := func(userSlug string) *http.Response {
		t.Helper()
		req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/ai/settings/validate?user="+userSlug+"&tree=mythic-tragedy-apprenticeship", strings.NewReader(requestBody))
		if err != nil {
			t.Fatalf("new validate request for %s: %v", userSlug, err)
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("validate request for %s: %v", userSlug, err)
		}
		return resp
	}

	firstResp := requestFor("tester")
	defer firstResp.Body.Close()
	if firstResp.StatusCode != http.StatusOK {
		t.Fatalf("first tester validate status = %d", firstResp.StatusCode)
	}

	repeatResp := requestFor("tester")
	defer repeatResp.Body.Close()
	if repeatResp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("repeat tester validate status = %d", repeatResp.StatusCode)
	}

	secondUserResp := requestFor("second-writer")
	defer secondUserResp.Body.Close()
	if secondUserResp.StatusCode != http.StatusOK {
		t.Fatalf("second user validate status = %d", secondUserResp.StatusCode)
	}

	if hits != 2 {
		t.Fatalf("provider validation hits = %d", hits)
	}
}

func TestAISettingsSaveSharesValidationBudget(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	cfg.AIValidateLimitPerMinute = 1
	cfg.AIValidateGlobalLimitPerMinute = 10
	testServer := newTestServerWithConfig(t, harness.Store, cfg)
	defer testServer.Close()

	hits := 0
	fakeProvider := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"data":[{"id":"gpt-5-mini"}]}`)
	}))
	defer fakeProvider.Close()

	requestBody := fmt.Sprintf(`{"provider":"openai","api_key":"sk-test-1234","base_url_override":%q,"enabled":true}`, fakeProvider.URL)

	validateReq, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/ai/settings/validate?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("new validate request: %v", err)
	}
	validateReq.Header.Set("Content-Type", "application/json")
	validateResp, err := http.DefaultClient.Do(validateReq)
	if err != nil {
		t.Fatalf("validate request: %v", err)
	}
	defer validateResp.Body.Close()
	if validateResp.StatusCode != http.StatusOK {
		t.Fatalf("validate status = %d", validateResp.StatusCode)
	}

	saveReq, err := http.NewRequest(http.MethodPut, testServer.URL+"/api/ai/settings?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(requestBody))
	if err != nil {
		t.Fatalf("new save request: %v", err)
	}
	saveReq.Header.Set("Content-Type", "application/json")
	saveResp, err := http.DefaultClient.Do(saveReq)
	if err != nil {
		t.Fatalf("save request: %v", err)
	}
	defer saveResp.Body.Close()
	if saveResp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("save status = %d", saveResp.StatusCode)
	}
	if hits != 1 {
		t.Fatalf("provider validation hits = %d", hits)
	}
}

func TestAISettingsUpdateRequiresNewKeyWhenChangingProvider(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	testServer := newTestServerWithStore(t, harness.Store, "", "")
	defer testServer.Close()

	encrypted, err := secrets.EncryptString("test-ai-key-secret", "sk-openai-1234")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	if err := harness.Store.SaveAIProviderSettings(context.Background(), domain.AIProviderSettings{
		UserID:          user.ID,
		Provider:        "openai",
		APIKeyEncrypted: encrypted,
		APIKeyLast4:     "1234",
		Enabled:         true,
		ValidatedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save ai provider settings: %v", err)
	}

	req, err := http.NewRequest(http.MethodPut, testServer.URL+"/api/ai/settings?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(`{"provider":"groq","api_key":"","enabled":true}`))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("update provider without key: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !strings.Contains(payload.Error, "api key is required") {
		t.Fatalf("unexpected error = %q", payload.Error)
	}
}

func TestAuthSessionReportsAIProviderReadiness(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	testServer := newTestServerWithConfig(t, harness.Store, cfg)
	defer testServer.Close()

	resp, err := http.Get(testServer.URL + "/api/auth/session?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get auth session: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d", resp.StatusCode)
	}
	var initial struct {
		AIProviderReady     bool   `json:"ai_provider_ready"`
		AIEffectiveProvider string `json:"ai_effective_provider"`
		AISystemFallback    bool   `json:"ai_system_fallback"`
		AIHasPersonalKey    bool   `json:"ai_has_personal_key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&initial); err != nil {
		t.Fatalf("decode auth session: %v", err)
	}
	if initial.AIProviderReady {
		t.Fatal("expected ai provider readiness to be false without fallback or personal key")
	}
	if initial.AISystemFallback {
		t.Fatal("expected system fallback to be false")
	}
	if initial.AIEffectiveProvider != "system/openai" {
		t.Fatalf("effective provider = %q", initial.AIEffectiveProvider)
	}
	if initial.AIHasPersonalKey {
		t.Fatal("expected personal key flag to be false")
	}

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	encrypted, err := secrets.EncryptString(cfg.AIKeySecret, "sk-user-1234")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if err := harness.Store.SaveAIProviderSettings(context.Background(), domain.AIProviderSettings{
		UserID:          user.ID,
		Provider:        "openai",
		APIKeyEncrypted: encrypted,
		APIKeyLast4:     "1234",
		Enabled:         true,
		ValidatedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save ai provider settings: %v", err)
	}

	resp, err = http.Get(testServer.URL + "/api/auth/session?user=tester&tree=mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("get auth session after provider save: %v", err)
	}
	defer resp.Body.Close()
	var updated struct {
		AIProviderReady     bool   `json:"ai_provider_ready"`
		AIEffectiveProvider string `json:"ai_effective_provider"`
		AIHasPersonalKey    bool   `json:"ai_has_personal_key"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&updated); err != nil {
		t.Fatalf("decode updated auth session: %v", err)
	}
	if !updated.AIProviderReady {
		t.Fatal("expected ai provider readiness after saving provider")
	}
	if updated.AIEffectiveProvider != "user" {
		t.Fatalf("effective provider = %q", updated.AIEffectiveProvider)
	}
	if !updated.AIHasPersonalKey {
		t.Fatal("expected personal key flag after saving provider")
	}
}

func TestAuthSessionReportsSetupStep(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	testServer := newTestServerWithConfig(t, harness.Store, cfg)
	defer testServer.Close()

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}

	readSession := func() struct {
		SetupStep          string `json:"setup_step"`
		OnboardingComplete bool   `json:"onboarding_complete"`
		AIProviderReady    bool   `json:"ai_provider_ready"`
	} {
		t.Helper()
		resp, err := http.Get(testServer.URL + "/api/auth/session?user=tester&tree=mythic-tragedy-apprenticeship")
		if err != nil {
			t.Fatalf("get auth session: %v", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("auth session status = %d", resp.StatusCode)
		}
		var payload struct {
			SetupStep          string `json:"setup_step"`
			OnboardingComplete bool   `json:"onboarding_complete"`
			AIProviderReady    bool   `json:"ai_provider_ready"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			t.Fatalf("decode auth session: %v", err)
		}
		return payload
	}

	initial := readSession()
	if initial.SetupStep != "needs_ai_setup" {
		t.Fatalf("initial setup_step = %q", initial.SetupStep)
	}

	encrypted, err := secrets.EncryptString(cfg.AIKeySecret, "sk-user-1234")
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}
	if err := harness.Store.SaveAIProviderSettings(context.Background(), domain.AIProviderSettings{
		UserID:          user.ID,
		Provider:        "openai",
		APIKeyEncrypted: encrypted,
		APIKeyLast4:     "1234",
		Enabled:         true,
		ValidatedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save ai provider settings: %v", err)
	}

	afterAI := readSession()
	if afterAI.SetupStep != "needs_first_track" {
		t.Fatalf("setup_step after ai = %q", afterAI.SetupStep)
	}

	enrollmentID, err := harness.Store.ActiveEnrollmentIDByUserID(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("lookup active enrollment: %v", err)
	}
	if err := harness.Store.SaveOnboardingProfile(context.Background(), domain.OnboardingProfile{
		EnrollmentID:        enrollmentID,
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
		GeneratedTreeSlug:   "mythic-tragedy-apprenticeship",
		TemplateKey:         "professional-writing",
	}); err != nil {
		t.Fatalf("save onboarding profile: %v", err)
	}

	afterTrack := readSession()
	if afterTrack.SetupStep != "needs_first_assignment" {
		t.Fatalf("setup_step after track = %q", afterTrack.SetupStep)
	}

	tree, err := harness.Store.TreeBySlug(context.Background(), "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("lookup tree: %v", err)
	}
	if _, err := harness.Store.SaveExercise(context.Background(), domain.Exercise{
		UserID:          user.ID,
		TreeID:          tree.ID,
		Title:           "Assignment One",
		Brief:           "Write something.",
		Constraints:     []string{"one"},
		FocusSkills:     []string{"prose precision"},
		TGOCodes:        []string{"prose-precision"},
		SuccessCriteria: []string{"clear result"},
		GenerationKind:  "openai",
	}); err != nil {
		t.Fatalf("save exercise: %v", err)
	}

	ready := readSession()
	if ready.SetupStep != "ready" {
		t.Fatalf("final setup_step = %q", ready.SetupStep)
	}
}

func TestAISettingsRejectUnsupportedProvider(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/ai/settings/validate?user=tester&tree=mythic-tragedy-apprenticeship", strings.NewReader(`{"provider":"mistral","api_key":"sk-test-1234","enabled":true}`))
	if err != nil {
		t.Fatalf("new validate request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("validate unsupported provider: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("unsupported provider status = %d", resp.StatusCode)
	}
	var payload struct {
		Error string `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode unsupported provider: %v", err)
	}
	if !strings.Contains(payload.Error, "unsupported provider") {
		t.Fatalf("unexpected error = %q", payload.Error)
	}
}

func TestPromptNextUsesUserProviderSettings(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	var authHeader string
	fakeProvider := newFakeOpenAIResponsesServer(t, &authHeader)
	defer fakeProvider.Close()

	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	server := newTestServerWithConfig(t, harness.Store, cfg)
	defer server.Close()

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	encrypted, err := secrets.EncryptString(cfg.AIKeySecret, "sk-user-9876")
	if err != nil {
		t.Fatalf("encrypt user key: %v", err)
	}
	if err := harness.Store.SaveAIProviderSettings(context.Background(), domain.AIProviderSettings{
		UserID:          user.ID,
		Provider:        "xai",
		APIKeyEncrypted: encrypted,
		APIKeyLast4:     "9876",
		BaseURLOverride: fakeProvider.URL,
		Enabled:         true,
		ValidatedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save ai provider settings: %v", err)
	}

	resp, err := http.Post(server.URL+"/api/prompts/next?user=tester&tree=mythic-tragedy-apprenticeship", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("prompt next: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("prompt next status: %d", resp.StatusCode)
	}

	var payload struct {
		Exercise struct {
			Title          string `json:"title"`
			GenerationKind string `json:"generation_kind"`
			ProviderNote   string `json:"provider_note"`
		} `json:"exercise"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode prompt payload: %v", err)
	}
	if payload.Exercise.Title != "Provider Draft" {
		t.Fatalf("exercise title = %q", payload.Exercise.Title)
	}
	if payload.Exercise.GenerationKind != "user/xai" {
		t.Fatalf("generation kind = %q", payload.Exercise.GenerationKind)
	}
	if payload.Exercise.ProviderNote != "user/xai • gpt-5-mini" {
		t.Fatalf("provider note = %q", payload.Exercise.ProviderNote)
	}
	if authHeader != "Bearer sk-user-9876" {
		t.Fatalf("authorization header = %q", authHeader)
	}
}

func TestReviewWorkerUsesUserProviderSettings(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	var authHeader string
	fakeProvider := newFakeOpenAIResponsesServer(t, &authHeader)
	defer fakeProvider.Close()

	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	server := Server{
		Config:     cfg,
		Store:      harness.Store,
		Prompts:    prompt.NewService(nil),
		Reviews:    review.NewService(nil, analyzer.Service{}),
		Curriculum: curriculum.NewService(),
	}

	ctx := context.Background()
	user, err := harness.Store.UserBySlug(ctx, "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	tree, err := harness.Store.TreeBySlug(ctx, "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("lookup tree: %v", err)
	}
	encrypted, err := secrets.EncryptString(cfg.AIKeySecret, "sk-review-4321")
	if err != nil {
		t.Fatalf("encrypt review key: %v", err)
	}
	if err := harness.Store.SaveAIProviderSettings(ctx, domain.AIProviderSettings{
		UserID:          user.ID,
		Provider:        "groq",
		APIKeyEncrypted: encrypted,
		APIKeyLast4:     "4321",
		BaseURLOverride: fakeProvider.URL,
		Enabled:         true,
		ValidatedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save ai provider settings: %v", err)
	}

	exerciseID, err := harness.Store.SaveExercise(ctx, domain.Exercise{
		UserID:          user.ID,
		TreeID:          tree.ID,
		Title:           "Provider Review Exercise",
		Brief:           "Write a paragraph.",
		Constraints:     []string{"under 200 words"},
		FocusSkills:     []string{"causal clarity"},
		TGOCodes:        []string{"causal-clarity"},
		SuccessCriteria: []string{"clear result"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	submissionID, err := harness.Store.SaveSubmission(ctx, domain.Submission{
		UserID:     user.ID,
		TreeID:     tree.ID,
		ExerciseID: exerciseID,
		Content:    "The bell rang and the gate split in the same instant.",
		WordCount:  11,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	enrollmentID, err := harness.Store.EnrollmentID(ctx, user.ID, tree.ID)
	if err != nil {
		t.Fatalf("enrollment id: %v", err)
	}
	jobRecord, err := harness.Store.EnqueueReviewJob(ctx, domain.ReviewJob{
		UserID:       user.ID,
		TreeID:       tree.ID,
		EnrollmentID: enrollmentID,
		SubmissionID: submissionID,
		MaxAttempts:  3,
	})
	if err != nil {
		t.Fatalf("enqueue review job: %v", err)
	}

	job, err := harness.Store.ClaimNextReviewJob(ctx)
	if err != nil {
		t.Fatalf("claim review job: %v", err)
	}
	if job.ID != jobRecord.ID {
		t.Fatalf("claimed job id = %d", job.ID)
	}
	if err := server.processReviewJob(ctx, job); err != nil {
		t.Fatalf("process review job: %v", err)
	}

	savedReview, err := harness.Store.LatestReviewForSubmission(ctx, submissionID)
	if err != nil {
		t.Fatalf("latest review: %v", err)
	}
	if savedReview.ReviewKind != "user/groq" {
		t.Fatalf("review kind = %q", savedReview.ReviewKind)
	}
	if savedReview.ProviderNote != "user/groq • gpt-5-mini" {
		t.Fatalf("provider note = %q", savedReview.ProviderNote)
	}
	if authHeader != "Bearer sk-review-4321" {
		t.Fatalf("authorization header = %q", authHeader)
	}
}

func TestPromptNextUsesAnthropicProviderSettings(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	var apiKey string
	fakeProvider := newFakeAnthropicServer(t, &apiKey)
	defer fakeProvider.Close()

	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	server := newTestServerWithConfig(t, harness.Store, cfg)
	defer server.Close()

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	encrypted, err := secrets.EncryptString(cfg.AIKeySecret, "sk-ant-5555")
	if err != nil {
		t.Fatalf("encrypt user key: %v", err)
	}
	if err := harness.Store.SaveAIProviderSettings(context.Background(), domain.AIProviderSettings{
		UserID:          user.ID,
		Provider:        "anthropic",
		APIKeyEncrypted: encrypted,
		APIKeyLast4:     "5555",
		BaseURLOverride: fakeProvider.URL,
		Enabled:         true,
		ValidatedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save ai provider settings: %v", err)
	}

	resp, err := http.Post(server.URL+"/api/prompts/next?user=tester&tree=mythic-tragedy-apprenticeship", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("prompt next: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("prompt next status: %d", resp.StatusCode)
	}
	var payload struct {
		Exercise struct {
			Title          string `json:"title"`
			GenerationKind string `json:"generation_kind"`
			ProviderNote   string `json:"provider_note"`
		} `json:"exercise"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Exercise.Title != "Anthropic Draft" {
		t.Fatalf("title = %q", payload.Exercise.Title)
	}
	if payload.Exercise.GenerationKind != "user/anthropic" {
		t.Fatalf("generation kind = %q", payload.Exercise.GenerationKind)
	}
	if payload.Exercise.ProviderNote != "user/anthropic • claude-sonnet-4-20250514" {
		t.Fatalf("provider note = %q", payload.Exercise.ProviderNote)
	}
	if apiKey != "sk-ant-5555" {
		t.Fatalf("api key header = %q", apiKey)
	}
}

func TestReviewWorkerUsesAnthropicProviderSettings(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	var apiKey string
	fakeProvider := newFakeAnthropicServer(t, &apiKey)
	defer fakeProvider.Close()

	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	server := Server{
		Config:     cfg,
		Store:      harness.Store,
		Prompts:    prompt.NewService(nil),
		Reviews:    review.NewService(nil, analyzer.Service{}),
		Curriculum: curriculum.NewService(),
	}

	ctx := context.Background()
	user, err := harness.Store.UserBySlug(ctx, "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	tree, err := harness.Store.TreeBySlug(ctx, "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("lookup tree: %v", err)
	}
	encrypted, err := secrets.EncryptString(cfg.AIKeySecret, "sk-ant-review")
	if err != nil {
		t.Fatalf("encrypt review key: %v", err)
	}
	if err := harness.Store.SaveAIProviderSettings(ctx, domain.AIProviderSettings{
		UserID:          user.ID,
		Provider:        "anthropic",
		APIKeyEncrypted: encrypted,
		APIKeyLast4:     "view",
		BaseURLOverride: fakeProvider.URL,
		Enabled:         true,
		ValidatedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save ai provider settings: %v", err)
	}

	exerciseID, err := harness.Store.SaveExercise(ctx, domain.Exercise{
		UserID:          user.ID,
		TreeID:          tree.ID,
		Title:           "Anthropic Review Exercise",
		Brief:           "Write a paragraph.",
		Constraints:     []string{"under 200 words"},
		FocusSkills:     []string{"causal clarity"},
		TGOCodes:        []string{"causal-clarity"},
		SuccessCriteria: []string{"clear result"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	submissionID, err := harness.Store.SaveSubmission(ctx, domain.Submission{
		UserID:     user.ID,
		TreeID:     tree.ID,
		ExerciseID: exerciseID,
		Content:    "The king refused the warning, and the room changed around him.",
		WordCount:  11,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	job, err := harness.Store.EnqueueReviewJob(ctx, domain.ReviewJob{
		SubmissionID: submissionID,
		UserID:       user.ID,
		TreeID:       tree.ID,
		EnrollmentID: 1,
		Status:       "queued",
		MaxAttempts:  3,
	})
	if err != nil {
		t.Fatalf("queue review job: %v", err)
	}
	if err := server.processReviewJob(ctx, job); err != nil {
		t.Fatalf("process review job: %v", err)
	}
	savedReview, err := harness.Store.LatestReviewForSubmission(ctx, submissionID)
	if err != nil {
		t.Fatalf("load review: %v", err)
	}
	if savedReview.ReviewKind != "user/anthropic" {
		t.Fatalf("review kind = %q", savedReview.ReviewKind)
	}
	if savedReview.ProviderNote != "user/anthropic • claude-sonnet-4-20250514" {
		t.Fatalf("provider note = %q", savedReview.ProviderNote)
	}
	if apiKey != "sk-ant-review" {
		t.Fatalf("api key header = %q", apiKey)
	}
}

func TestPromptNextUsesGeminiProviderSettings(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	var apiKey string
	fakeProvider := newFakeGeminiServer(t, &apiKey)
	defer fakeProvider.Close()

	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	server := newTestServerWithConfig(t, harness.Store, cfg)
	defer server.Close()

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	encrypted, err := secrets.EncryptString(cfg.AIKeySecret, "sk-gem-1234")
	if err != nil {
		t.Fatalf("encrypt user key: %v", err)
	}
	if err := harness.Store.SaveAIProviderSettings(context.Background(), domain.AIProviderSettings{
		UserID:          user.ID,
		Provider:        "gemini",
		APIKeyEncrypted: encrypted,
		APIKeyLast4:     "1234",
		BaseURLOverride: fakeProvider.URL,
		Enabled:         true,
		ValidatedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save ai provider settings: %v", err)
	}

	resp, err := http.Post(server.URL+"/api/prompts/next?user=tester&tree=mythic-tragedy-apprenticeship", "application/json", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("prompt next: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("prompt next status: %d", resp.StatusCode)
	}
	var payload struct {
		Exercise struct {
			Title          string `json:"title"`
			GenerationKind string `json:"generation_kind"`
			ProviderNote   string `json:"provider_note"`
		} `json:"exercise"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Exercise.Title != "Gemini Draft" {
		t.Fatalf("title = %q", payload.Exercise.Title)
	}
	if payload.Exercise.GenerationKind != "user/gemini" {
		t.Fatalf("generation kind = %q", payload.Exercise.GenerationKind)
	}
	if payload.Exercise.ProviderNote != "user/gemini • gemini-2.5-flash" {
		t.Fatalf("provider note = %q", payload.Exercise.ProviderNote)
	}
	if apiKey != "sk-gem-1234" {
		t.Fatalf("api key query = %q", apiKey)
	}
}

func TestReviewWorkerUsesGeminiProviderSettings(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	var apiKey string
	fakeProvider := newFakeGeminiServer(t, &apiKey)
	defer fakeProvider.Close()

	cfg := config.Default(t.TempDir())
	cfg.AIKeySecret = "test-ai-key-secret"
	server := Server{
		Config:     cfg,
		Store:      harness.Store,
		Prompts:    prompt.NewService(nil),
		Reviews:    review.NewService(nil, analyzer.Service{}),
		Curriculum: curriculum.NewService(),
	}

	ctx := context.Background()
	user, err := harness.Store.UserBySlug(ctx, "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	tree, err := harness.Store.TreeBySlug(ctx, "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("lookup tree: %v", err)
	}
	encrypted, err := secrets.EncryptString(cfg.AIKeySecret, "sk-gem-review")
	if err != nil {
		t.Fatalf("encrypt review key: %v", err)
	}
	if err := harness.Store.SaveAIProviderSettings(ctx, domain.AIProviderSettings{
		UserID:          user.ID,
		Provider:        "gemini",
		APIKeyEncrypted: encrypted,
		APIKeyLast4:     "view",
		BaseURLOverride: fakeProvider.URL,
		Enabled:         true,
		ValidatedAt:     time.Now().UTC(),
	}); err != nil {
		t.Fatalf("save ai provider settings: %v", err)
	}

	exerciseID, err := harness.Store.SaveExercise(ctx, domain.Exercise{
		UserID:          user.ID,
		TreeID:          tree.ID,
		Title:           "Gemini Review Exercise",
		Brief:           "Write a paragraph.",
		Constraints:     []string{"under 200 words"},
		FocusSkills:     []string{"causal clarity"},
		TGOCodes:        []string{"causal-clarity"},
		SuccessCriteria: []string{"clear result"},
		GenerationKind:  "deterministic",
	})
	if err != nil {
		t.Fatalf("save exercise: %v", err)
	}
	submissionID, err := harness.Store.SaveSubmission(ctx, domain.Submission{
		UserID:     user.ID,
		TreeID:     tree.ID,
		ExerciseID: exerciseID,
		Content:    "The warning landed too late, and the silence did the rest.",
		WordCount:  11,
	})
	if err != nil {
		t.Fatalf("save submission: %v", err)
	}
	job, err := harness.Store.EnqueueReviewJob(ctx, domain.ReviewJob{
		SubmissionID: submissionID,
		UserID:       user.ID,
		TreeID:       tree.ID,
		EnrollmentID: 1,
		Status:       "queued",
		MaxAttempts:  3,
	})
	if err != nil {
		t.Fatalf("queue review job: %v", err)
	}
	if err := server.processReviewJob(ctx, job); err != nil {
		t.Fatalf("process review job: %v", err)
	}
	savedReview, err := harness.Store.LatestReviewForSubmission(ctx, submissionID)
	if err != nil {
		t.Fatalf("load review: %v", err)
	}
	if savedReview.ReviewKind != "user/gemini" {
		t.Fatalf("review kind = %q", savedReview.ReviewKind)
	}
	if savedReview.ProviderNote != "user/gemini • gemini-2.5-flash" {
		t.Fatalf("provider note = %q", savedReview.ProviderNote)
	}
	if apiKey != "sk-gem-review" {
		t.Fatalf("api key query = %q", apiKey)
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

func TestPromptNextUsesSelectedTGOsForSuccessCriteria(t *testing.T) {
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
			SuccessCriteria []string `json:"success_criteria"`
		} `json:"exercise"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode prompt response: %v", err)
	}

	want := []string{
		"Stage turns, entrances, exits, and power shifts cleanly.",
		"Replace soft modifiers with exact nouns and verbs.",
		"Make action and consequence legible beat by beat.",
	}
	if len(payload.Exercise.SuccessCriteria) != len(want) {
		t.Fatalf("success criteria count = %d, want %d (%v)", len(payload.Exercise.SuccessCriteria), len(want), payload.Exercise.SuccessCriteria)
	}
	for i := range want {
		if payload.Exercise.SuccessCriteria[i] != want[i] {
			t.Fatalf("success criterion %d = %q, want %q", i, payload.Exercise.SuccessCriteria[i], want[i])
		}
	}
	for _, criterion := range payload.Exercise.SuccessCriteria {
		if strings.Contains(strings.ToLower(criterion), "word") {
			t.Fatalf("unexpected non-TGO success criterion leaked into rubric: %q", criterion)
		}
	}
}

func TestPromptPreviewDoesNotPersistUntilAccepted(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	testServer := newTestServerWithStore(t, harness.Store, "", "")
	defer testServer.Close()

	resp, err := http.Post(
		testServer.URL+"/api/prompts/next?user=tester",
		"application/json",
		strings.NewReader(`{"tgo_codes":["scene-architecture","prose-precision","causal-clarity"]}`),
	)
	if err != nil {
		t.Fatalf("preview prompt: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("preview status: %d", resp.StatusCode)
	}

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	tree, err := harness.Store.TreeBySlug(context.Background(), "mythic-tragedy-apprenticeship")
	if err != nil {
		t.Fatalf("lookup tree: %v", err)
	}
	exercises, err := harness.Store.ListExercises(context.Background(), user.ID, tree.ID, 10)
	if err != nil {
		t.Fatalf("list exercises after preview: %v", err)
	}
	if len(exercises) != 0 {
		t.Fatalf("expected no saved exercises after preview, got %d", len(exercises))
	}

	var payload struct {
		Exercise exerciseResponse `json:"exercise"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode preview payload: %v", err)
	}

	acceptResp, err := http.Post(
		testServer.URL+"/api/prompts/accept?user=tester",
		"application/json",
		strings.NewReader(mustJSONString(map[string]any{
			"title":            payload.Exercise.Title,
			"brief":            payload.Exercise.Brief,
			"constraints":      payload.Exercise.Constraints,
			"focus_skills":     payload.Exercise.FocusSkills,
			"tgo_codes":        payload.Exercise.TGOCodes,
			"success_criteria": payload.Exercise.SuccessCriteria,
			"generation_kind":  payload.Exercise.GenerationKind,
			"provider_note":    payload.Exercise.ProviderNote,
		})),
	)
	if err != nil {
		t.Fatalf("accept prompt: %v", err)
	}
	defer acceptResp.Body.Close()
	if acceptResp.StatusCode != http.StatusOK {
		t.Fatalf("accept status: %d", acceptResp.StatusCode)
	}

	exercises, err = harness.Store.ListExercises(context.Background(), user.ID, tree.ID, 10)
	if err != nil {
		t.Fatalf("list exercises after accept: %v", err)
	}
	if len(exercises) != 1 {
		t.Fatalf("expected one saved exercise after accept, got %d", len(exercises))
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

func TestTracksListAndActiveSwitchUseTrackScopedProfiles(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	firstPayload := `{
		"mode":"edit",
		"writing_type":"fiction",
		"assignment_format":"scene",
		"target_audience":"fantasy readers",
		"subject_matter":"mythic conflict",
		"experience_level":"intermediate",
		"desired_tone":"grave",
		"biggest_weaknesses":["scene architecture","word choice"],
		"desired_outcomes":["publish stronger fiction","develop a distinctive voice"],
		"difficulty_intensity":"steady",
		"writing_goals":"I want to write stronger scenes."
	}`
	firstResp, err := http.Post(testServer.URL+"/api/onboarding?user=tester", "application/json", strings.NewReader(firstPayload))
	if err != nil {
		t.Fatalf("post first onboarding: %v", err)
	}
	defer firstResp.Body.Close()
	if firstResp.StatusCode != http.StatusOK {
		t.Fatalf("first onboarding status: %d", firstResp.StatusCode)
	}
	var firstBody struct {
		Tree struct {
			Slug string `json:"slug"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(firstResp.Body).Decode(&firstBody); err != nil {
		t.Fatalf("decode first onboarding: %v", err)
	}

	initialTracksResp, err := http.Get(testServer.URL + "/api/tracks?user=tester")
	if err != nil {
		t.Fatalf("get initial tracks: %v", err)
	}
	defer initialTracksResp.Body.Close()
	if initialTracksResp.StatusCode != http.StatusOK {
		t.Fatalf("initial tracks status: %d", initialTracksResp.StatusCode)
	}
	var initialTracksBody struct {
		Tracks []struct {
			TreeSlug string `json:"tree_slug"`
			IsActive bool   `json:"is_active"`
		} `json:"tracks"`
	}
	if err := json.NewDecoder(initialTracksResp.Body).Decode(&initialTracksBody); err != nil {
		t.Fatalf("decode initial tracks: %v", err)
	}
	if len(initialTracksBody.Tracks) != 1 {
		t.Fatalf("expected one track after first onboarding, got %#v", initialTracksBody.Tracks)
	}
	if initialTracksBody.Tracks[0].TreeSlug != firstBody.Tree.Slug {
		t.Fatalf("initial track slug = %q", initialTracksBody.Tracks[0].TreeSlug)
	}
	if !initialTracksBody.Tracks[0].IsActive {
		t.Fatal("expected first generated track to be active")
	}

	secondPayload := `{
		"mode":"create",
		"writing_type":"technical writing",
		"assignment_format":"how-to guide",
		"target_audience":"API integrators",
		"subject_matter":"developer tooling",
		"experience_level":"advanced",
		"desired_tone":"clear",
		"biggest_weaknesses":["sentence economy","paragraph control"],
		"desired_outcomes":["improve professional communication","write clearer essays"],
		"difficulty_intensity":"ambitious",
		"writing_goals":"I want cleaner technical drafts."
	}`
	secondResp, err := http.Post(testServer.URL+"/api/onboarding?user=tester", "application/json", strings.NewReader(secondPayload))
	if err != nil {
		t.Fatalf("post second onboarding: %v", err)
	}
	defer secondResp.Body.Close()
	if secondResp.StatusCode != http.StatusOK {
		t.Fatalf("second onboarding status: %d", secondResp.StatusCode)
	}
	var secondBody struct {
		Tree struct {
			Slug string `json:"slug"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(secondResp.Body).Decode(&secondBody); err != nil {
		t.Fatalf("decode second onboarding: %v", err)
	}
	if secondBody.Tree.Slug == firstBody.Tree.Slug {
		t.Fatalf("expected unique track slug, got %q", secondBody.Tree.Slug)
	}

	tracksResp, err := http.Get(testServer.URL + "/api/tracks?user=tester")
	if err != nil {
		t.Fatalf("get tracks: %v", err)
	}
	defer tracksResp.Body.Close()
	if tracksResp.StatusCode != http.StatusOK {
		t.Fatalf("tracks status: %d", tracksResp.StatusCode)
	}
	var tracksBody struct {
		Tracks []struct {
			TreeSlug string `json:"tree_slug"`
			IsActive bool   `json:"is_active"`
		} `json:"tracks"`
	}
	if err := json.NewDecoder(tracksResp.Body).Decode(&tracksBody); err != nil {
		t.Fatalf("decode tracks: %v", err)
	}
	if len(tracksBody.Tracks) != 2 {
		t.Fatalf("track count = %d", len(tracksBody.Tracks))
	}
	foundFirst := false
	foundSecond := false
	for _, track := range tracksBody.Tracks {
		if track.TreeSlug == firstBody.Tree.Slug {
			foundFirst = true
		}
		if track.TreeSlug == secondBody.Tree.Slug {
			foundSecond = true
		}
	}
	if !foundFirst || !foundSecond {
		t.Fatalf("expected both generated tracks in list, got %#v", tracksBody.Tracks)
	}

	switchReq, err := http.NewRequest(http.MethodPut, testServer.URL+"/api/tracks/active?user=tester", strings.NewReader(fmt.Sprintf(`{"tree_slug":%q}`, firstBody.Tree.Slug)))
	if err != nil {
		t.Fatalf("new switch request: %v", err)
	}
	switchReq.Header.Set("Content-Type", "application/json")
	switchResp, err := http.DefaultClient.Do(switchReq)
	if err != nil {
		t.Fatalf("switch active track: %v", err)
	}
	defer switchResp.Body.Close()
	if switchResp.StatusCode != http.StatusOK {
		t.Fatalf("switch status: %d", switchResp.StatusCode)
	}

	getResp, err := http.Get(testServer.URL + "/api/onboarding?user=tester")
	if err != nil {
		t.Fatalf("get onboarding after switch: %v", err)
	}
	defer getResp.Body.Close()
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("get onboarding after switch status: %d", getResp.StatusCode)
	}
	var getBody struct {
		Context struct {
			TreeSlug string `json:"tree_slug"`
		} `json:"context"`
		Profile struct {
			AssignmentFormat string `json:"assignment_format"`
		} `json:"profile"`
	}
	if err := json.NewDecoder(getResp.Body).Decode(&getBody); err != nil {
		t.Fatalf("decode onboarding after switch: %v", err)
	}
	if getBody.Context.TreeSlug != firstBody.Tree.Slug {
		t.Fatalf("active context tree slug = %q", getBody.Context.TreeSlug)
	}
	if getBody.Profile.AssignmentFormat != "scene" {
		t.Fatalf("profile assignment format = %q", getBody.Profile.AssignmentFormat)
	}
}

func TestArchiveTrackRemovesItAndSwitchesActiveContext(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	firstPayload := `{
		"mode":"edit",
		"writing_type":"fiction",
		"assignment_format":"scene",
		"target_audience":"fantasy readers",
		"subject_matter":"mythic conflict",
		"experience_level":"intermediate",
		"desired_tone":"grave",
		"biggest_weaknesses":["scene architecture","word choice"],
		"desired_outcomes":["publish stronger fiction","develop a distinctive voice"],
		"difficulty_intensity":"steady",
		"writing_goals":"I want to write stronger scenes."
	}`
	firstResp, err := http.Post(testServer.URL+"/api/onboarding?user=tester", "application/json", strings.NewReader(firstPayload))
	if err != nil {
		t.Fatalf("post first onboarding: %v", err)
	}
	defer firstResp.Body.Close()
	if firstResp.StatusCode != http.StatusOK {
		t.Fatalf("first onboarding status: %d", firstResp.StatusCode)
	}
	var firstBody struct {
		Tree struct {
			Slug string `json:"slug"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(firstResp.Body).Decode(&firstBody); err != nil {
		t.Fatalf("decode first onboarding: %v", err)
	}

	secondPayload := `{
		"mode":"create",
		"writing_type":"technical writing",
		"assignment_format":"how-to guide",
		"target_audience":"API integrators",
		"subject_matter":"developer tooling",
		"experience_level":"advanced",
		"desired_tone":"clear",
		"biggest_weaknesses":["sentence economy","paragraph control"],
		"desired_outcomes":["improve professional communication","write clearer essays"],
		"difficulty_intensity":"ambitious",
		"writing_goals":"I want cleaner technical drafts."
	}`
	secondResp, err := http.Post(testServer.URL+"/api/onboarding?user=tester", "application/json", strings.NewReader(secondPayload))
	if err != nil {
		t.Fatalf("post second onboarding: %v", err)
	}
	defer secondResp.Body.Close()
	if secondResp.StatusCode != http.StatusOK {
		t.Fatalf("second onboarding status: %d", secondResp.StatusCode)
	}
	var secondBody struct {
		Tree struct {
			Slug string `json:"slug"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(secondResp.Body).Decode(&secondBody); err != nil {
		t.Fatalf("decode second onboarding: %v", err)
	}

	archiveReq, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/tracks/"+secondBody.Tree.Slug+"/archive?user=tester", strings.NewReader(`{}`))
	if err != nil {
		t.Fatalf("new archive request: %v", err)
	}
	archiveReq.Header.Set("Content-Type", "application/json")
	archiveResp, err := http.DefaultClient.Do(archiveReq)
	if err != nil {
		t.Fatalf("archive track: %v", err)
	}
	defer archiveResp.Body.Close()
	if archiveResp.StatusCode != http.StatusOK {
		t.Fatalf("archive status: %d", archiveResp.StatusCode)
	}

	var archiveBody struct {
		Context struct {
			TreeSlug string `json:"tree_slug"`
		} `json:"context"`
		Tracks []struct {
			TreeSlug string `json:"tree_slug"`
		} `json:"tracks"`
	}
	if err := json.NewDecoder(archiveResp.Body).Decode(&archiveBody); err != nil {
		t.Fatalf("decode archive response: %v", err)
	}
	if archiveBody.Context.TreeSlug == secondBody.Tree.Slug {
		t.Fatalf("active tree after archive = %q", archiveBody.Context.TreeSlug)
	}
	foundFirst := false
	for _, track := range archiveBody.Tracks {
		if track.TreeSlug == firstBody.Tree.Slug {
			foundFirst = true
		}
		if track.TreeSlug == secondBody.Tree.Slug {
			t.Fatalf("archived track still present in tracks list: %q", track.TreeSlug)
		}
	}
	if !foundFirst {
		t.Fatalf("expected first generated track to remain after archive: %#v", archiveBody.Tracks)
	}
}

func TestTreeGetUsesProfileDisplayForGeneratedTrack(t *testing.T) {
	testServer := newTestServer(t)
	defer testServer.Close()

	resp, err := http.Post(
		testServer.URL+"/api/onboarding?user=tester",
		"application/json",
		strings.NewReader(`{
			"writing_type":"fiction",
			"assignment_format":"scene",
			"target_audience":"fantasy readers",
			"subject_matter":"succession fights and sacred relics",
			"experience_level":"advanced",
			"desired_tone":"serious and emotional, literary, philosophical",
			"biggest_weaknesses":["symbol control","line pressure"],
			"desired_outcomes":["stronger long-form fiction","stronger symbolic control"],
			"difficulty_intensity":"intensive",
			"writing_goals":"I want to become a great author at mythopoeic literature, epic fantasy, and world building."
		}`),
	)
	if err != nil {
		t.Fatalf("post onboarding: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("onboarding status: %d", resp.StatusCode)
	}

	treeResp, err := http.Get(testServer.URL + "/api/trees/tester-track?user=tester")
	if err != nil {
		t.Fatalf("get generated tree: %v", err)
	}
	defer treeResp.Body.Close()
	if treeResp.StatusCode != http.StatusOK {
		t.Fatalf("tree status: %d", treeResp.StatusCode)
	}

	var payload struct {
		Tree struct {
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"tree"`
	}
	if err := json.NewDecoder(treeResp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode tree payload: %v", err)
	}
	if payload.Tree.Title != "Tester's Fiction Track" {
		t.Fatalf("tree title = %q", payload.Tree.Title)
	}
	if strings.Contains(payload.Tree.Description, "Advanced mythopoeic tragic fiction track.") {
		t.Fatalf("tree description leaked template branding: %q", payload.Tree.Description)
	}
	if !strings.Contains(payload.Tree.Description, "Skill track for fiction, with scene assignments for fantasy readers.") {
		t.Fatalf("tree description = %q", payload.Tree.Description)
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

func TestAdminAIProviderEventsEndpointReturnsSummary(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	testServer := newTestServerWithStore(t, harness.Store, "", "")
	defer testServer.Close()

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	now := time.Now().UTC()
	events := []domain.AIProviderEvent{
		{
			UserID:     user.ID,
			Provider:   "openai",
			Event:      "settings_validate_failed",
			Category:   "auth",
			StatusCode: http.StatusBadRequest,
			DetailJSON: `{"status":400}`,
			CreatedAt:  now,
		},
		{
			UserID:     user.ID,
			Provider:   "openai",
			Event:      "settings_validate_rate_limited",
			Category:   "local_rate_limit",
			StatusCode: http.StatusTooManyRequests,
			DetailJSON: `{"status":429}`,
			CreatedAt:  now.Add(-time.Minute),
		},
		{
			UserID:     user.ID,
			Provider:   "anthropic",
			Event:      "generation_fallback",
			Category:   "quota",
			StatusCode: http.StatusBadGateway,
			DetailJSON: `{"kind":"prompt_next"}`,
			CreatedAt:  now.Add(-2 * time.Minute),
		},
	}
	for _, event := range events {
		if err := harness.Store.SaveAIProviderEvent(context.Background(), event); err != nil {
			t.Fatalf("save ai provider event: %v", err)
		}
	}

	resp, err := http.Get(testServer.URL + "/api/admin/ai-provider-events?limit=10&hours=24")
	if err != nil {
		t.Fatalf("get admin ai provider events: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("admin ai provider events status = %d", resp.StatusCode)
	}

	var payload struct {
		Summary struct {
			Total               int `json:"total"`
			ValidationFailures  int `json:"validation_failures"`
			ValidationRateLimit int `json:"validation_rate_limit"`
			Fallbacks           int `json:"fallbacks"`
			ProviderCounts      []struct {
				Label string `json:"label"`
				Count int    `json:"count"`
			} `json:"provider_counts"`
		} `json:"summary"`
		Events []struct {
			UserSlug string         `json:"user_slug"`
			Event    string         `json:"event"`
			Details  map[string]any `json:"details"`
		} `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode admin ai provider events: %v", err)
	}
	if payload.Summary.Total != 3 {
		t.Fatalf("summary total = %d", payload.Summary.Total)
	}
	if payload.Summary.ValidationFailures != 1 {
		t.Fatalf("validation failures = %d", payload.Summary.ValidationFailures)
	}
	if payload.Summary.ValidationRateLimit != 1 {
		t.Fatalf("validation rate limit = %d", payload.Summary.ValidationRateLimit)
	}
	if payload.Summary.Fallbacks != 1 {
		t.Fatalf("fallbacks = %d", payload.Summary.Fallbacks)
	}
	if len(payload.Events) != 3 {
		t.Fatalf("event count = %d", len(payload.Events))
	}
	if payload.Events[0].UserSlug != "tester" {
		t.Fatalf("first event user slug = %q", payload.Events[0].UserSlug)
	}
	if payload.Events[0].Event == "" {
		t.Fatal("expected event label")
	}
	if len(payload.Summary.ProviderCounts) == 0 {
		t.Fatal("expected provider counts")
	}
}

func TestAdminAIProviderEventsEndpointFiltersResults(t *testing.T) {
	harness := newTestHarnessWithAuth(t, "", "")
	testServer := newTestServerWithStore(t, harness.Store, "", "")
	defer testServer.Close()

	user, err := harness.Store.UserBySlug(context.Background(), "tester")
	if err != nil {
		t.Fatalf("lookup user: %v", err)
	}
	now := time.Now().UTC()
	for _, event := range []domain.AIProviderEvent{
		{UserID: user.ID, Provider: "openai", Event: "settings_validate_failed", Category: "auth", StatusCode: 400, CreatedAt: now},
		{UserID: user.ID, Provider: "anthropic", Event: "generation_fallback", Category: "quota", StatusCode: 502, CreatedAt: now},
	} {
		if err := harness.Store.SaveAIProviderEvent(context.Background(), event); err != nil {
			t.Fatalf("save ai provider event: %v", err)
		}
	}

	resp, err := http.Get(testServer.URL + "/api/admin/ai-provider-events?limit=10&hours=24&provider=openai&event=settings_validate_failed")
	if err != nil {
		t.Fatalf("get filtered admin ai provider events: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("filtered admin ai provider events status = %d", resp.StatusCode)
	}

	var payload struct {
		Summary struct {
			Total int `json:"total"`
		} `json:"summary"`
		Events []struct {
			Provider string `json:"provider"`
			Event    string `json:"event"`
		} `json:"events"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		t.Fatalf("decode filtered admin ai provider events: %v", err)
	}
	if payload.Summary.Total != 1 {
		t.Fatalf("filtered summary total = %d", payload.Summary.Total)
	}
	if len(payload.Events) != 1 {
		t.Fatalf("filtered event count = %d", len(payload.Events))
	}
	if payload.Events[0].Provider != "openai" || payload.Events[0].Event != "settings_validate_failed" {
		t.Fatalf("filtered event = %#v", payload.Events[0])
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
	cfg.AIKeySecret = "test-ai-key-secret"
	return newTestServerWithConfig(t, harness.Store, cfg)
}

type testHarness struct {
	Store *db.Store
}

func configureTestSQLite(t *testing.T, store *db.Store) {
	t.Helper()

	pragmas := []string{
		"PRAGMA journal_mode = MEMORY",
		"PRAGMA synchronous = OFF",
		"PRAGMA temp_store = MEMORY",
	}
	for _, pragma := range pragmas {
		if _, err := store.SQL.Exec(pragma); err != nil {
			t.Fatalf("configure sqlite pragma %q: %v", pragma, err)
		}
	}
}

func newTestHarnessWithAuth(t *testing.T, apiToken, kratosPublicURL string) testHarness {
	t.Helper()
	root := t.TempDir()
	store, err := db.Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	configureTestSQLite(t, store)

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
	cfg.AIKeySecret = "test-ai-key-secret"
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
	ctx, cancel := context.WithCancel(context.Background())
	server.startBackgroundWorkers(ctx)
	testServer := httptest.NewServer(server.routes())
	t.Cleanup(func() {
		cancel()
		testServer.Close()
	})
	return testServer
}

func int64String(value int64) string {
	return strconv.FormatInt(value, 10)
}

func mustJSONString(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(data)
}

func newFakeOpenAIResponsesServer(t *testing.T, authHeader *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authHeader != nil {
			*authHeader = r.Header.Get("Authorization")
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read fake provider body: %v", err)
		}
		_ = r.Body.Close()

		var payload string
		switch {
		case strings.Contains(string(body), "submission_review"):
			payload = `{"summary":"Provider review summary","strengths":["Provider strength"],"weaknesses":["Provider weakness"],"next_focus":"causal clarity","skill_scores":[{"skill":"causal clarity","score":4}],"tgo_assessments":[{"code":"causal-clarity","status":"secure","evidence":"Provider evidence"},{"code":"scene-architecture","status":"secure","evidence":"Provider evidence"},{"code":"prose-precision","status":"secure","evidence":"Provider evidence"}],"completed_tgo_checks":[],"annotations":[]}`
		default:
			payload = `{"title":"Provider Draft","brief":"Generated through the user provider.","constraints":["Keep the draft focused."],"focus_skills":["causal clarity"],"success_criteria":["Make the chain of events easy to follow."]}`
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"output": []map[string]any{
				{
					"content": []map[string]any{
						{
							"type": "output_text",
							"text": payload,
						},
					},
				},
			},
		})
	}))
}

func newFakeAnthropicServer(t *testing.T, apiKey *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey != nil {
			*apiKey = r.Header.Get("x-api-key")
		}
		switch r.URL.Path {
		case "/models":
			_, _ = io.WriteString(w, `{"data":[]}`)
			return
		case "/messages":
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read body: %v", err)
			}
			if strings.Contains(string(body), "submission_review") {
				_, _ = io.WriteString(w, `{"content":[{"type":"tool_use","input":{"summary":"Anthropic review summary","strengths":["Clear turn","Concrete consequence"],"weaknesses":["Push the beats harder","Tighten the prose"],"next_focus":"narrative clarity","skill_scores":[{"skill":"scene architecture","score":4},{"skill":"narrative clarity","score":3},{"skill":"prose precision","score":3}],"tgo_assessments":[{"code":"causal-clarity","status":"secure","evidence":"Cause and effect remain visible."},{"code":"scene-architecture","status":"developing","evidence":"The middle turn could land harder."},{"code":"prose-precision","status":"developing","evidence":"Some lines can tighten."}],"completed_tgo_checks":[],"annotations":[{"quote":"the room changed around him","tgo_code":"causal-clarity","category":"strength","comment":"The consequence lands on the page.","severity":"info"}]}}]}`)
				return
			}
			_, _ = io.WriteString(w, `{"content":[{"type":"tool_use","input":{"title":"Anthropic Draft","brief":"Write the scene from scratch.","constraints":["Keep the focus tight.","Make the turn visible."],"focus_skills":["scene architecture","narrative clarity"],"success_criteria":["The shift is easy to follow.","The ending lands cleanly."]}}]}`)
			return
		default:
			http.NotFound(w, r)
		}
	}))
}

func newFakeGeminiServer(t *testing.T, apiKey *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if apiKey != nil {
			*apiKey = r.URL.Query().Get("key")
		}
		switch r.URL.Path {
		case "/models":
			_, _ = io.WriteString(w, `{"models":[]}`)
			return
		default:
			if strings.HasSuffix(r.URL.Path, ":generateContent") {
				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("read body: %v", err)
				}
				if strings.Contains(string(body), "Submission ID:") {
					_, _ = io.WriteString(w, `{"candidates":[{"content":{"parts":[{"text":"{\"summary\":\"Gemini review summary\",\"strengths\":[\"Clear turn\",\"Concrete pressure\"],\"weaknesses\":[\"Tighten the midsection\",\"Sharpen the closing line\"],\"next_focus\":\"narrative clarity\",\"skill_scores\":[{\"skill\":\"scene architecture\",\"score\":4},{\"skill\":\"narrative clarity\",\"score\":3},{\"skill\":\"prose precision\",\"score\":3}],\"tgo_assessments\":[{\"code\":\"causal-clarity\",\"status\":\"secure\",\"evidence\":\"The draft preserves consequence.\"},{\"code\":\"scene-architecture\",\"status\":\"developing\",\"evidence\":\"The midpoint turn can sharpen.\"},{\"code\":\"prose-precision\",\"status\":\"developing\",\"evidence\":\"Some lines can tighten.\"}],\"completed_tgo_checks\":[],\"annotations\":[{\"quote\":\"the silence did the rest\",\"tgo_code\":\"causal-clarity\",\"category\":\"strength\",\"comment\":\"The effect lands on the page.\",\"severity\":\"info\"}]}"}]}}]}`)
					return
				}
				_, _ = io.WriteString(w, `{"candidates":[{"content":{"parts":[{"text":"{\"title\":\"Gemini Draft\",\"brief\":\"Write the scene from scratch.\",\"constraints\":[\"Keep the focus tight.\",\"Make the turn visible.\"],\"focus_skills\":[\"scene architecture\",\"narrative clarity\"],\"success_criteria\":[\"The shift is easy to follow.\",\"The ending lands cleanly.\"]}"}]}}]}`)
				return
			}
			http.NotFound(w, r)
		}
	}))
}

func waitForReview(t *testing.T, baseURL string, submissionID int64) int64 {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := http.Get(baseURL + "/api/review-jobs?user=tester&tree=mythic-tragedy-apprenticeship&submission_id=" + int64String(submissionID))
		if err != nil {
			t.Fatalf("get review job: %v", err)
		}
		var payload struct {
			Job struct {
				ReviewID int64  `json:"review_id"`
				Status   string `json:"status"`
			} `json:"job"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			resp.Body.Close()
			t.Fatalf("decode review job: %v", err)
		}
		resp.Body.Close()
		if payload.Job.Status == "completed" && payload.Job.ReviewID != 0 {
			return payload.Job.ReviewID
		}
		if payload.Job.Status == "failed" {
			t.Fatalf("review job failed for submission %d", submissionID)
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatalf("timed out waiting for review job for submission %d", submissionID)
	return 0
}
