package db

import (
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

func objectiveScoresToSkillScores(scores []domain.ObjectiveScore) []domain.SkillScore {
	if len(scores) == 0 {
		return nil
	}
	type agg struct {
		total    int
		count    int
		source   string
		version  string
		evidence string
	}
	bySkill := map[string]agg{}
	for _, score := range scores {
		if strings.TrimSpace(score.ScoreSource) != "deterministic" {
			continue
		}
		skill := strings.TrimSpace(domain.TGOCodeToSkill[score.TGOCode])
		if skill == "" {
			continue
		}
		item := bySkill[skill]
		item.total += score.Score
		item.count++
		if item.source == "" {
			item.source = strings.TrimSpace(score.ScoreSource)
		}
		if item.version == "" {
			item.version = strings.TrimSpace(score.ScoreVersion)
		}
		if item.evidence == "" {
			item.evidence = strings.TrimSpace(score.ScoreEvidenceJSON)
		}
		bySkill[skill] = item
	}
	if len(bySkill) == 0 {
		return nil
	}
	skills := make([]string, 0, len(bySkill))
	for skill := range bySkill {
		skills = append(skills, skill)
	}
	sort.Strings(skills)
	out := make([]domain.SkillScore, 0, len(skills))
	for _, skill := range skills {
		item := bySkill[skill]
		if item.count == 0 {
			continue
		}
		avg := int(float64(item.total)/float64(item.count) + 0.5)
		if avg < 1 {
			avg = 1
		}
		if avg > 5 {
			avg = 5
		}
		out = append(out, domain.SkillScore{
			Skill:             skill,
			Score:             avg,
			ScoreSource:       item.source,
			ScoreVersion:      item.version,
			ScoreEvidenceJSON: item.evidence,
		})
	}
	return out
}
