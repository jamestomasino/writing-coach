package api

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/session"
)

type userResponse struct {
	ID             int64  `json:"id"`
	Slug           string `json:"slug"`
	Name           string `json:"name"`
	ActiveTreeSlug string `json:"active_tree_slug,omitempty"`
	CreatedAt      string `json:"created_at"`
}

type enrollmentResponse struct {
	ID        int64  `json:"id"`
	UserID    int64  `json:"user_id"`
	TreeID    int64  `json:"tree_id"`
	UserSlug  string `json:"user_slug"`
	TreeSlug  string `json:"tree_slug"`
	CreatedAt string `json:"created_at"`
}

func (s Server) handleUsersList(w http.ResponseWriter, r *http.Request) {
	users, err := s.Store.ListUsers(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"users": toUserResponses(users)})
}

func (s Server) handleUsersCreate(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Slug string `json:"slug"`
		Name string `json:"name"`
		Tree string `json:"tree"`
	}
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(payload.Slug) == "" || strings.TrimSpace(payload.Name) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("slug and name are required"))
		return
	}
	treeSlug := firstNonEmpty(payload.Tree, s.Config.DefaultTreeSlug)
	userID, treeID, enrollmentID, err := s.Store.EnsureDefaultUserTree(r.Context(), payload.Slug, payload.Name, treeSlug)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"user": map[string]any{
			"id":   userID,
			"slug": payload.Slug,
			"name": payload.Name,
		},
		"enrollment": map[string]any{
			"id":        enrollmentID,
			"user_id":   userID,
			"tree_id":   treeID,
			"user_slug": payload.Slug,
			"tree_slug": treeSlug,
		},
	})
}

func (s Server) handleUserGet(w http.ResponseWriter, r *http.Request) {
	user, err := s.Store.UserBySlug(r.Context(), r.PathValue("slug"))
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": toUserResponses([]domain.User{user})[0]})
}

func (s Server) handleEnrollmentsList(w http.ResponseWriter, r *http.Request) {
	enrollments, err := s.Store.ListEnrollments(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"enrollments": toEnrollmentResponses(enrollments)})
}

func (s Server) handleEnrollmentsCreate(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		UserSlug string `json:"user_slug"`
		UserName string `json:"user_name"`
		TreeSlug string `json:"tree_slug"`
	}
	if err := decodeJSONBody(w, r, &payload); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if strings.TrimSpace(payload.UserSlug) == "" || strings.TrimSpace(payload.TreeSlug) == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("user_slug and tree_slug are required"))
		return
	}
	userName := firstNonEmpty(payload.UserName, payload.UserSlug)
	userID, treeID, enrollmentID, err := s.Store.EnsureDefaultUserTree(r.Context(), payload.UserSlug, userName, payload.TreeSlug)
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{
		"enrollment": map[string]any{
			"id":        enrollmentID,
			"user_id":   userID,
			"tree_id":   treeID,
			"user_slug": payload.UserSlug,
			"tree_slug": payload.TreeSlug,
		},
	})
}

func (s Server) handleEnrollmentBoard(w http.ResponseWriter, r *http.Request) {
	enrollmentID, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || enrollmentID == 0 {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid enrollment id"))
		return
	}
	enrollment, err := s.Store.EnrollmentByID(r.Context(), enrollmentID)
	if err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	s.writeEnrollmentBoard(r.Context(), w, session.Context{
		UserID:       enrollment.UserID,
		TreeID:       enrollment.TreeID,
		EnrollmentID: enrollment.ID,
		UserSlug:     enrollment.UserSlug,
		TreeSlug:     enrollment.TreeSlug,
	})
}

func toUserResponses(users []domain.User) []userResponse {
	out := make([]userResponse, 0, len(users))
	for _, user := range users {
		out = append(out, userResponse{
			ID:             user.ID,
			Slug:           user.Slug,
			Name:           user.Name,
			ActiveTreeSlug: user.ActiveTreeSlug,
			CreatedAt:      db.Since(user.CreatedAt),
		})
	}
	return out
}

func toEnrollmentResponses(enrollments []domain.Enrollment) []enrollmentResponse {
	out := make([]enrollmentResponse, 0, len(enrollments))
	for _, enrollment := range enrollments {
		out = append(out, enrollmentResponse{
			ID:        enrollment.ID,
			UserID:    enrollment.UserID,
			TreeID:    enrollment.TreeID,
			UserSlug:  enrollment.UserSlug,
			TreeSlug:  enrollment.TreeSlug,
			CreatedAt: db.Since(enrollment.CreatedAt),
		})
	}
	return out
}
