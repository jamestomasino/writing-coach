package curriculum

import (
	"context"

	"github.com/tomasino/writing-coach/internal/db"
	"github.com/tomasino/writing-coach/internal/domain"
)

type Service struct{}

type Recommendation struct {
	Focus      string
	Difficulty int
	Rationale  string
}

func NewService() Service {
	return Service{}
}

func (Service) DescribeNextStep(state domain.CurriculumState) string {
	if state.CurrentFocus == "" {
		return "Active TGOs drive the next assignment."
	}
	return "Primary pressure remains on " + state.CurrentFocus + "."
}

func (Service) SyncTGOs(ctx context.Context, store *db.Store, treeSlug string, enrollmentID int64, review domain.Review) (Recommendation, error) {
	active, err := store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		return Recommendation{}, err
	}
	completed, err := store.CompletedTGOs(ctx, enrollmentID)
	if err != nil {
		return Recommendation{}, err
	}
	completedSet := make(map[string]bool, len(completed))
	for _, tgo := range completed {
		completedSet[tgo.Code] = true
	}
	activeSet := make(map[string]bool, len(active))
	for _, tgo := range active {
		activeSet[tgo.Code] = true
	}

	for _, check := range review.CompletedTGOChecks {
		if check.Status != "slipping" {
			continue
		}
		if slipped, ok := domain.TGOByCode(check.TGOCode); ok {
			return Recommendation{
				Focus:      slipped.Title,
				Difficulty: 1,
				Rationale:  "Hold advancement while a completed TGO slips: " + slipped.Title + ". Evidence: " + check.Evidence,
			}, nil
		}
		return Recommendation{
			Focus:      review.NextFocus,
			Difficulty: 1,
			Rationale:  "Hold advancement while a completed TGO slips. Evidence: " + check.Evidence,
		}, nil
	}

	for _, assessment := range review.TGOAssessments {
		if assessment.Status != "mastered" {
			continue
		}
		tgo, ok := findTGO(active, assessment.TGOCode)
		if !ok {
			continue
		}
		signal, err := store.TGOMasterySignal(ctx, enrollmentID, tgo, assessment.Status)
		if err != nil {
			return Recommendation{}, err
		}
		if !signal.Ready {
			continue
		}
		completedSet[assessment.TGOCode] = true
		if err := store.MarkTGOCompleted(ctx, enrollmentID, assessment.TGOCode); err != nil {
			return Recommendation{}, err
		}
	}

	active, err = store.ActiveTGOs(ctx, enrollmentID)
	if err != nil {
		return Recommendation{}, err
	}
	primary := ""
	rationale := "Active TGOs stay fixed for the full assignment chain. Choose replacements when you start the next assignment."
	if len(active) > 0 {
		primary = active[0].Title
		rationale = active[0].Description + " Mastery marker: " + active[0].MasteryHint
	}

	return Recommendation{
		Focus:      primary,
		Difficulty: 2,
		Rationale:  rationale,
	}, nil
}

func findTGO(active []domain.TGO, code string) (domain.TGO, bool) {
	for _, tgo := range active {
		if tgo.Code == code {
			return tgo, true
		}
	}
	return domain.TGO{}, false
}
