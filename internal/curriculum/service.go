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

func (Service) SyncTGOs(ctx context.Context, store *db.Store, review domain.Review) (Recommendation, error) {
	active, err := store.ActiveTGOs(ctx)
	if err != nil {
		return Recommendation{}, err
	}
	completed, err := store.CompletedTGOs(ctx)
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

	for _, assessment := range review.TGOAssessments {
		if assessment.Status != "mastered" {
			continue
		}
		slot := findSlot(active, assessment.TGOCode)
		if slot == 0 {
			continue
		}
		statuses, err := store.RecentTGOStatuses(ctx, assessment.TGOCode, 2)
		if err != nil {
			return Recommendation{}, err
		}
		if len(statuses) < 2 || statuses[0] != "mastered" || statuses[1] == "developing" {
			continue
		}
		completedSet[assessment.TGOCode] = true
		delete(activeSet, assessment.TGOCode)
		nextOptions := domain.NextUnlockedTGOs(completedSet, activeSet, 1)
		if len(nextOptions) == 0 {
			continue
		}
		next := nextOptions[0]
		if err := store.ReplaceActiveTGO(ctx, slot, assessment.TGOCode, next.Code); err != nil {
			return Recommendation{}, err
		}
		activeSet[next.Code] = true
	}

	active, err = store.ActiveTGOs(ctx)
	if err != nil {
		return Recommendation{}, err
	}
	primary := ""
	rationale := "Maintain exactly three active TGOs and advance only when mastery is stable."
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

func findSlot(active []domain.TGO, code string) int {
	for _, tgo := range active {
		if tgo.Code == code {
			return tgo.ActiveSlot
		}
	}
	return 0
}
