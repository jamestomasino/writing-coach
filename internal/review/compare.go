package review

import (
	"sort"
	"strings"

	"github.com/tomasino/writing-coach/internal/domain"
)

type Comparison struct {
	CurrentID            int64
	BaselineID           int64
	WordDelta            int
	AddedWords           []string
	RemovedWords         []string
	PersistingWeaknesses []string
	AddressedWeaknesses  []string
	SkillSetMismatch     bool
	SkillDeltas          []SkillDelta
	Summary              string
}

type SkillDelta struct {
	Skill          string   `json:"skill"`
	BaselineScore  int      `json:"baseline_score"`
	CurrentScore   int      `json:"current_score"`
	Delta          int      `json:"delta"`
	Direction      string   `json:"direction"`
	EvidenceQuotes []string `json:"evidence_quotes,omitempty"`
}

func CompareSubmissions(current, baseline domain.Submission, currentReview domain.Review, baselineReview domain.Review) Comparison {
	added, removed := diffWordSets(baseline.Content, current.Content)
	persisting, addressed := weaknessDelta(baselineReview.Weaknesses, currentReview.Weaknesses)
	skillDeltas, skillSetMismatch := scoreDelta(baselineReview.SkillScores, currentReview.SkillScores, currentReview.Annotations)

	summary := "Revision changes are mixed."
	switch {
	case len(addressed) > len(persisting):
		summary = "Revision appears to address more weaknesses than it preserves."
	case len(persisting) > 0:
		summary = "Revision still carries forward several earlier weaknesses."
	case len(addressed) > 0:
		summary = "Revision appears to have addressed earlier weaknesses."
	}

	return Comparison{
		CurrentID:            current.ID,
		BaselineID:           baseline.ID,
		WordDelta:            current.WordCount - baseline.WordCount,
		AddedWords:           added,
		RemovedWords:         removed,
		PersistingWeaknesses: persisting,
		AddressedWeaknesses:  addressed,
		SkillSetMismatch:     skillSetMismatch,
		SkillDeltas:          skillDeltas,
		Summary:              summary,
	}
}

func diffWordSets(before, after string) ([]string, []string) {
	beforeSet := wordSet(before)
	afterSet := wordSet(after)
	var added []string
	var removed []string
	for word := range afterSet {
		if !beforeSet[word] {
			added = append(added, word)
		}
	}
	for word := range beforeSet {
		if !afterSet[word] {
			removed = append(removed, word)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	if len(added) > 10 {
		added = added[:10]
	}
	if len(removed) > 10 {
		removed = removed[:10]
	}
	return added, removed
}

func wordSet(text string) map[string]bool {
	set := map[string]bool{}
	for _, word := range strings.Fields(strings.ToLower(text)) {
		word = strings.Trim(word, ".,;:!?\"'()[]{}")
		if len(word) < 4 {
			continue
		}
		set[word] = true
	}
	return set
}

func weaknessDelta(before, after []string) ([]string, []string) {
	var persisting []string
	var addressed []string
	joinedAfter := strings.ToLower(strings.Join(after, " || "))
	for _, weakness := range before {
		key := normalizeWeaknessKey(weakness)
		if key == "" {
			continue
		}
		if strings.Contains(joinedAfter, key) {
			persisting = append(persisting, weakness)
		} else {
			addressed = append(addressed, weakness)
		}
	}
	if len(persisting) > 3 {
		persisting = persisting[:3]
	}
	if len(addressed) > 3 {
		addressed = addressed[:3]
	}
	return persisting, addressed
}

func normalizeWeaknessKey(value string) string {
	value = strings.ToLower(value)
	parts := strings.Fields(value)
	if len(parts) > 6 {
		parts = parts[:6]
	}
	return strings.Join(parts, " ")
}

func scoreDelta(baseline, current []domain.SkillScore, annotations []domain.ReviewAnnotation) ([]SkillDelta, bool) {
	baselineMap := authoritativeSkillMap(baseline)
	currentMap := authoritativeSkillMap(current)
	if len(currentMap) == 0 {
		return nil, false
	}
	skillSetMismatch := !sameSkillKeys(baselineMap, currentMap)
	keys := make([]string, 0, len(currentMap))
	for skill := range currentMap {
		if _, ok := baselineMap[skill]; ok {
			keys = append(keys, skill)
		}
	}
	sort.Strings(keys)
	if len(keys) > 8 {
		keys = keys[:8]
	}
	out := make([]SkillDelta, 0, len(keys))
	for _, skill := range keys {
		cur := currentMap[skill]
		base := baselineMap[skill]
		delta := cur - base
		direction := "flat"
		switch {
		case delta > 0:
			direction = "up"
		case delta < 0:
			direction = "down"
		}
		out = append(out, SkillDelta{
			Skill:          skill,
			BaselineScore:  base,
			CurrentScore:   cur,
			Delta:          delta,
			Direction:      direction,
			EvidenceQuotes: evidenceQuotesForSkill(skill, annotations),
		})
	}
	return out, skillSetMismatch
}

func authoritativeSkillMap(scores []domain.SkillScore) map[string]int {
	primary := map[string]int{}
	legacy := map[string]int{}
	fallback := map[string]int{}
	for _, score := range scores {
		skill := strings.TrimSpace(score.Skill)
		if skill == "" {
			continue
		}
		switch {
		case score.ScoreSource == "deterministic":
			primary[skill] = score.Score
		case strings.Contains(score.ScoreSource, "legacy") || strings.TrimSpace(score.ScoreSource) == "":
			legacy[skill] = score.Score
		default:
			if _, ok := fallback[skill]; !ok {
				fallback[skill] = score.Score
			}
		}
	}
	if len(primary) > 0 {
		return primary
	}
	if len(legacy) > 0 {
		return legacy
	}
	return fallback
}

func evidenceQuotesForSkill(skill string, annotations []domain.ReviewAnnotation) []string {
	parts := strings.Fields(strings.ToLower(skill))
	keywords := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.Trim(part, ".,;:!?\"'()[]{}")
		if len(part) >= 4 {
			keywords = append(keywords, part)
		}
	}
	if len(keywords) == 0 || len(annotations) == 0 {
		return nil
	}
	out := make([]string, 0, 2)
	seen := map[string]bool{}
	for _, ann := range annotations {
		haystack := strings.ToLower(strings.TrimSpace(ann.Category + " " + ann.Comment + " " + ann.TGOCode))
		for _, key := range keywords {
			if strings.Contains(haystack, key) {
				quote := strings.TrimSpace(ann.Quote)
				if quote != "" && !seen[quote] {
					out = append(out, quote)
					seen[quote] = true
				}
				break
			}
		}
		if len(out) == 2 {
			break
		}
	}
	return out
}

func sameSkillKeys(left, right map[string]int) bool {
	if len(left) != len(right) {
		return false
	}
	for key := range left {
		if _, ok := right[key]; !ok {
			return false
		}
	}
	return true
}
