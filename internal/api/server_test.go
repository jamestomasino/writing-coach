package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/curriculum"
	"github.com/tomasino/writing-coach/internal/db"
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

	boardResp, err := http.Get(testServer.URL + "/api/enrollments/2/board")
	if err != nil {
		t.Fatalf("get board: %v", err)
	}
	defer boardResp.Body.Close()
	if boardResp.StatusCode != http.StatusOK {
		t.Fatalf("board status: %d", boardResp.StatusCode)
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

func newTestServer(t *testing.T) *httptest.Server {
	return newTestServerWithToken(t, "")
}

func newTestServerWithToken(t *testing.T, apiToken string) *httptest.Server {
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

	cfg := config.Default(root)
	cfg.APIToken = apiToken

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
		_ = store.Close()
	})
	return testServer
}
