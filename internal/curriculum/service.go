package curriculum

import (
	"context"
	"sort"

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
		return "Focus on scene architecture next."
	}
	return "Focus on " + state.CurrentFocus + " next."
}

func (Service) RecommendNextFocus(ctx context.Context, store *db.Store, state domain.CurriculumState, suggested string, scores []domain.SkillScore) (Recommendation, error) {
	averages, err := store.SkillAverages(ctx, 5)
	if err != nil {
		return Recommendation{}, err
	}

	latest := map[string]int{}
	for _, score := range scores {
		latest[score.Skill] = score.Score
	}

	type candidate struct {
		skill  string
		weight float64
	}
	var candidates []candidate
	for _, skill := range domain.PrioritySkills {
		avg := averages[skill]
		if avg == 0 {
			avg = 3
		}
		current := latest[skill]
		if current == 0 {
			current = int(avg + 0.5)
		}

		weight := (6.0 - avg) + (6.0 - float64(current))
		if skill == suggested {
			weight += 1.5
		}
		if skill == state.CurrentFocus {
			weight -= 0.75
		}

		recent, err := store.RecentSkillScores(ctx, skill, 2)
		if err != nil {
			return Recommendation{}, err
		}
		if len(recent) >= 2 && recent[0] <= recent[1] {
			weight += 0.75
		}

		weight += float64(domain.SkillPriority(skill)) * 0.15
		candidates = append(candidates, candidate{skill: skill, weight: weight})
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].weight == candidates[j].weight {
			return domain.SkillPriority(candidates[i].skill) > domain.SkillPriority(candidates[j].skill)
		}
		return candidates[i].weight > candidates[j].weight
	})

	focus := suggested
	if len(candidates) > 0 {
		focus = candidates[0].skill
	}
	if focus == "" {
		focus = "scene architecture"
	}

	return Recommendation{
		Focus:      focus,
		Difficulty: db.NextDifficulty(focus),
		Rationale:  domain.HouseCurriculum[focus],
	}, nil
}
