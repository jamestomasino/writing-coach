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

	boardResp, err := http.Get(testServer.URL + "/api/enrollments/2/board")
	if err != nil {
		t.Fatalf("get board: %v", err)
	}
	defer boardResp.Body.Close()
	if boardResp.StatusCode != http.StatusOK {
		t.Fatalf("board status: %d", boardResp.StatusCode)
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
			Artifacts struct {
				AnalyzerReport map[string]any `json:"analyzer_report"`
				Recommendation map[string]any `json:"recommendation"`
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

func newTestServer(t *testing.T) *httptest.Server {
	return newTestServerWithAuth(t, "", "")
}

func newTestServerWithToken(t *testing.T, apiToken string) *httptest.Server {
	return newTestServerWithAuth(t, apiToken, "")
}

func newTestServerWithAuth(t *testing.T, apiToken, kratosPublicURL string) *httptest.Server {
	harness := newTestHarnessWithAuth(t, apiToken, kratosPublicURL)
	return newTestServerWithStore(t, harness.Store, apiToken, kratosPublicURL)
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
	root := t.TempDir()
	cfg := config.Default(root)
	cfg.APIToken = apiToken
	cfg.KratosPublicURL = kratosPublicURL

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
