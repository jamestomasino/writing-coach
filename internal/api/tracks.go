package api

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/session"
)

type trackResponse struct {
	EnrollmentID         int64  `json:"enrollment_id"`
	TreeID               int64  `json:"tree_id"`
	TreeSlug             string `json:"tree_slug"`
	Title                string `json:"title"`
	Description          string `json:"description"`
	IsActive             bool   `json:"is_active"`
	CreatedAt            string `json:"created_at"`
	AssignmentCount      int    `json:"assignment_count"`
	CurrentAssignment    string `json:"current_assignment,omitempty"`
	LatestAssignmentTime string `json:"latest_assignment_time,omitempty"`
}

func (s Server) handleTracksList(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tracks, err := s.Store.ListUserTracks(r.Context(), appContext.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeTrackPayload(w, http.StatusOK, appContext, s.trackResponses(r.Context(), appContext, tracks))
}

func (s Server) handleTracksActiveUpdate(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	var payload struct {
		TreeSlug string `json:"tree_slug"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON body"))
		return
	}
	treeSlug := strings.TrimSpace(payload.TreeSlug)
	if treeSlug == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("tree_slug is required"))
		return
	}
	if err := s.ensureTrackEnrollment(r.Context(), appContext.UserID, treeSlug); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	if err := s.Store.SetUserActiveTree(r.Context(), appContext.UserID, treeSlug); err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	s.writeTrackCollection(w, r, appContext.UserID)
}

func (s Server) handleTracksArchive(w http.ResponseWriter, r *http.Request) {
	appContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	treeSlug := strings.TrimSpace(r.PathValue("slug"))
	if treeSlug == "" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("track slug is required"))
		return
	}
	tracks, err := s.Store.ListUserTracks(r.Context(), appContext.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	target, nextActive, err := chooseArchivedTrackOutcome(tracks, treeSlug)
	if err != nil {
		status := http.StatusBadRequest
		if errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	if target.IsActive {
		if err := s.Store.SetUserActiveTree(r.Context(), appContext.UserID, nextActive.TreeSlug); err != nil {
			writeError(w, http.StatusInternalServerError, err)
			return
		}
	}
	if err := s.Store.ArchiveUserTrack(r.Context(), appContext.UserID, treeSlug); err != nil {
		status := http.StatusInternalServerError
		if db.IsNotFound(err) || errors.Is(err, sql.ErrNoRows) {
			status = http.StatusNotFound
		}
		writeError(w, status, err)
		return
	}
	s.writeTrackCollection(w, r, appContext.UserID)
}

func (s Server) writeTrackCollection(w http.ResponseWriter, r *http.Request, userID int64) {
	nextContext, err := s.resolveSession(r.Context(), r)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	tracks, err := s.Store.ListUserTracks(r.Context(), userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}
	writeTrackPayload(w, http.StatusOK, nextContext, s.trackResponses(r.Context(), nextContext, tracks))
}

func writeTrackPayload(w http.ResponseWriter, status int, appContext session.Context, tracks []trackResponse) {
	writeJSON(w, status, map[string]any{
		"context": requestContextResponse{
			UserSlug: appContext.UserSlug,
			TreeSlug: appContext.TreeSlug,
			UserID:   appContext.UserID,
			TreeID:   appContext.TreeID,
		},
		"tracks": tracks,
	})
}

func (s Server) ensureTrackEnrollment(ctx context.Context, userID int64, treeSlug string) error {
	enrollments, err := s.Store.ListEnrollmentsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, enrollment := range enrollments {
		if enrollment.TreeSlug == treeSlug {
			return nil
		}
	}
	return fmt.Errorf("tree %q is not enrolled for this user", treeSlug)
}

func chooseArchivedTrackOutcome(tracks []domain.UserTrack, treeSlug string) (domain.UserTrack, domain.UserTrack, error) {
	if len(tracks) <= 1 {
		return domain.UserTrack{}, domain.UserTrack{}, fmt.Errorf("you must keep at least one active track")
	}
	var target *domain.UserTrack
	var nextActive *domain.UserTrack
	for i := range tracks {
		track := &tracks[i]
		if track.TreeSlug == treeSlug {
			target = track
			continue
		}
		if nextActive == nil {
			nextActive = track
		}
	}
	if target == nil {
		return domain.UserTrack{}, domain.UserTrack{}, sql.ErrNoRows
	}
	if target.IsActive && nextActive == nil {
		return domain.UserTrack{}, domain.UserTrack{}, fmt.Errorf("you must keep at least one active track")
	}
	if nextActive == nil {
		return *target, domain.UserTrack{}, nil
	}
	return *target, *nextActive, nil
}

func (s Server) trackResponses(ctx context.Context, appContext session.Context, tracks []domain.UserTrack) []trackResponse {
	user, _ := s.Store.UserBySlug(ctx, appContext.UserSlug)
	out := make([]trackResponse, 0, len(tracks))
	for _, track := range tracks {
		item := trackResponse{
			EnrollmentID: track.EnrollmentID,
			TreeID:       track.TreeID,
			TreeSlug:     track.TreeSlug,
			Title:        track.Title,
			Description:  track.Description,
			IsActive:     track.IsActive,
			CreatedAt:    db.Since(track.CreatedAt),
		}
		if profile, err := s.Store.OnboardingProfileByEnrollmentID(ctx, track.EnrollmentID); err == nil && profile.GeneratedTreeSlug == track.TreeSlug {
			item.Title, item.Description = domain.GeneratedTreeDisplay(user.Name, profile, item.Title, item.Description)
		}
		summary, err := s.trackActivitySummary(ctx, appContext.UserID, track)
		if err == nil {
			item.AssignmentCount = summary.AssignmentCount
			item.CurrentAssignment = summary.CurrentAssignment
			if !summary.LatestAssignmentTime.IsZero() {
				item.LatestAssignmentTime = db.Since(summary.LatestAssignmentTime)
			}
		}
		out = append(out, item)
	}
	return out
}

func (s Server) trackActivitySummary(ctx context.Context, userID int64, track domain.UserTrack) (domain.TrackActivitySummary, error) {
	exercises, err := s.Store.ListExercises(ctx, userID, track.TreeID, 500)
	if err != nil {
		return domain.TrackActivitySummary{}, err
	}
	if len(exercises) == 0 {
		return domain.TrackActivitySummary{}, nil
	}
	submissions, err := s.Store.ListSubmissions(ctx, userID, track.TreeID, 0, 500)
	if err != nil {
		return domain.TrackActivitySummary{}, err
	}
	reviews, err := s.Store.ListReviews(ctx, userID, track.TreeID, 0, 500)
	if err != nil {
		return domain.TrackActivitySummary{}, err
	}

	exerciseByID := make(map[int64]domain.Exercise, len(exercises))
	for _, exercise := range exercises {
		exerciseByID[exercise.ID] = exercise
	}
	submissionByID := make(map[int64]domain.Submission, len(submissions))
	for _, submission := range submissions {
		submissionByID[submission.ID] = submission
	}
	reviewBySubmission := make(map[int64]domain.Review, len(reviews))
	for _, reviewResult := range reviews {
		if _, exists := reviewBySubmission[reviewResult.SubmissionID]; !exists {
			reviewBySubmission[reviewResult.SubmissionID] = reviewResult
		}
	}

	type chainSummary struct {
		title          string
		latestActivity time.Time
	}
	chains := map[int64]*chainSummary{}
	latestExercise := exercises[0]
	for _, exercise := range exercises {
		if exercise.CreatedAt.After(latestExercise.CreatedAt) || (exercise.CreatedAt.Equal(latestExercise.CreatedAt) && exercise.ID > latestExercise.ID) {
			latestExercise = exercise
		}
		rootID := rootExerciseID(exercise, exerciseByID, submissionByID)
		chain := chains[rootID]
		if chain == nil {
			chains[rootID] = &chainSummary{
				title:          exerciseTitle(exerciseByID[rootID]),
				latestActivity: exercise.CreatedAt,
			}
			continue
		}
		if exercise.CreatedAt.After(chain.latestActivity) {
			chain.latestActivity = exercise.CreatedAt
		}
	}
	for _, submission := range submissions {
		exercise, ok := exerciseByID[submission.ExerciseID]
		if !ok {
			continue
		}
		rootID := rootExerciseID(exercise, exerciseByID, submissionByID)
		chain := chains[rootID]
		if chain == nil {
			continue
		}
		if submission.CreatedAt.After(chain.latestActivity) {
			chain.latestActivity = submission.CreatedAt
		}
		if reviewResult, ok := reviewBySubmission[submission.ID]; ok && reviewResult.CreatedAt.After(chain.latestActivity) {
			chain.latestActivity = reviewResult.CreatedAt
		}
	}

	currentRootID := rootExerciseID(latestExercise, exerciseByID, submissionByID)
	current := chains[currentRootID]
	summary := domain.TrackActivitySummary{
		AssignmentCount: len(chains),
	}
	if current != nil {
		summary.CurrentAssignment = current.title
		summary.LatestAssignmentTime = current.latestActivity
	}
	for _, chain := range chains {
		if chain.latestActivity.After(summary.LatestAssignmentTime) {
			summary.LatestAssignmentTime = chain.latestActivity
		}
	}
	return summary, nil
}
