package review

import (
	"encoding/json"
	"fmt"
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
	Skill              string   `json:"skill"`
	BaselineScore      int      `json:"baseline_score"`
	CurrentScore       int      `json:"current_score"`
	Delta              int      `json:"delta"`
	Direction          string   `json:"direction"`
	EvidenceQuotes     []string `json:"evidence_quotes,omitempty"`
	BaselineRules      []string `json:"baseline_rules,omitempty"`
	CurrentRules       []string `json:"current_rules,omitempty"`
	DeltaExplanation   string   `json:"delta_explanation,omitempty"`
	DeterministicDelta bool     `json:"deterministic_delta"`
}

func CompareSubmissions(current, baseline domain.Submission, currentReview domain.Review, baselineReview domain.Review) Comparison {
	added, removed := diffWordSets(baseline.Content, current.Content)
	persisting, addressed := weaknessDelta(baselineReview.Weaknesses, currentReview.Weaknesses)
	var skillDeltas []SkillDelta
	var skillSetMismatch bool
	switch {
	case len(currentReview.ObjectiveScores) > 0 && len(baselineReview.ObjectiveScores) > 0:
		skillDeltas, skillSetMismatch = objectiveScoreDelta(baselineReview.ObjectiveScores, currentReview.ObjectiveScores, currentReview.Annotations)
	case len(currentReview.ObjectiveScores) > 0 && len(baselineReview.SkillScores) > 0:
		allowedSkills := allowedSignalSkills(currentReview.TGOAssessments)
		skillDeltas, skillSetMismatch = scoreDelta(
			baselineReview.SkillScores,
			objectiveScoresToSkillScores(currentReview.ObjectiveScores),
			currentReview.Annotations,
			allowedSkills,
		)
	case len(currentReview.SkillScores) > 0 && len(baselineReview.ObjectiveScores) > 0:
		allowedSkills := allowedSignalSkills(currentReview.TGOAssessments)
		skillDeltas, skillSetMismatch = scoreDelta(
			objectiveScoresToSkillScores(baselineReview.ObjectiveScores),
			currentReview.SkillScores,
			currentReview.Annotations,
			allowedSkills,
		)
	default:
		allowedSkills := allowedSignalSkills(currentReview.TGOAssessments)
		skillDeltas, skillSetMismatch = scoreDelta(baselineReview.SkillScores, currentReview.SkillScores, currentReview.Annotations, allowedSkills)
	}

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

func scoreDelta(baseline, current []domain.SkillScore, annotations []domain.ReviewAnnotation, allowed map[string]bool) ([]SkillDelta, bool) {
	baselineMap := authoritativeSkillMap(baseline)
	currentMap := authoritativeSkillMap(current)
	if len(allowed) > 0 {
		baselineMap = filterSkillMap(baselineMap, allowed)
		currentMap = filterSkillMap(currentMap, allowed)
	}
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
		delta := cur.Score - base.Score
		direction := "flat"
		switch {
		case delta > 0:
			direction = "up"
		case delta < 0:
			direction = "down"
		}
		baselineRules := appliedRules(base.ScoreEvidenceJSON)
		currentRules := appliedRules(cur.ScoreEvidenceJSON)
		out = append(out, SkillDelta{
			Skill:              skill,
			BaselineScore:      base.Score,
			CurrentScore:       cur.Score,
			Delta:              delta,
			Direction:          direction,
			EvidenceQuotes:     evidenceQuotesForSkill(skill, annotations),
			BaselineRules:      baselineRules,
			CurrentRules:       currentRules,
			DeltaExplanation:   deltaExplanation(delta, baselineRules, currentRules),
			DeterministicDelta: strings.TrimSpace(base.ScoreSource) == "deterministic" || strings.TrimSpace(cur.ScoreSource) == "deterministic",
		})
	}
	return out, skillSetMismatch
}

func objectiveScoreDelta(baseline, current []domain.ObjectiveScore, annotations []domain.ReviewAnnotation) ([]SkillDelta, bool) {
	baselineMap := authoritativeObjectiveMap(baseline)
	currentMap := authoritativeObjectiveMap(current)
	if len(currentMap) == 0 {
		return nil, false
	}
	skillSetMismatch := !sameObjectiveKeys(baselineMap, currentMap)
	keys := make([]string, 0, len(currentMap))
	for code := range currentMap {
		if _, ok := baselineMap[code]; ok {
			keys = append(keys, code)
		}
	}
	sort.Strings(keys)
	if len(keys) > 8 {
		keys = keys[:8]
	}
	out := make([]SkillDelta, 0, len(keys))
	for _, code := range keys {
		cur := currentMap[code]
		base := baselineMap[code]
		delta := cur.Score - base.Score
		direction := "flat"
		switch {
		case delta > 0:
			direction = "up"
		case delta < 0:
			direction = "down"
		}
		baselineRules := appliedRules(base.ScoreEvidenceJSON)
		currentRules := appliedRules(cur.ScoreEvidenceJSON)
		label := code
		if tgo, ok := domain.TGOByCode(code); ok {
			label = tgo.Title
		}
		out = append(out, SkillDelta{
			Skill:              label,
			BaselineScore:      base.Score,
			CurrentScore:       cur.Score,
			Delta:              delta,
			Direction:          direction,
			EvidenceQuotes:     objectiveEvidenceQuotes(code, annotations),
			BaselineRules:      baselineRules,
			CurrentRules:       currentRules,
			DeltaExplanation:   deltaExplanation(delta, baselineRules, currentRules),
			DeterministicDelta: strings.TrimSpace(base.ScoreSource) == "deterministic" || strings.TrimSpace(cur.ScoreSource) == "deterministic",
		})
	}
	return out, skillSetMismatch
}

func filterSkillMap(input map[string]domain.SkillScore, allowed map[string]bool) map[string]domain.SkillScore {
	if len(input) == 0 || len(allowed) == 0 {
		return input
	}
	out := make(map[string]domain.SkillScore, len(input))
	for skill, score := range input {
		if !allowed[strings.ToLower(strings.TrimSpace(skill))] {
			continue
		}
		out[skill] = score
	}
	return out
}

func allowedSignalSkills(assessments []domain.TGOAssessment) map[string]bool {
	if len(assessments) == 0 {
		return nil
	}
	out := map[string]bool{}
	for _, assessment := range assessments {
		skill := strings.TrimSpace(domain.TGOCodeToSkill[assessment.TGOCode])
		if skill == "" {
			continue
		}
		out[strings.ToLower(skill)] = true
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func objectiveScoresToSkillScores(scores []domain.ObjectiveScore) []domain.SkillScore {
	if len(scores) == 0 {
		return nil
	}
	authoritative := authoritativeObjectiveMap(scores)
	if len(authoritative) == 0 {
		return nil
	}
	codes := make([]string, 0, len(authoritative))
	for code := range authoritative {
		codes = append(codes, code)
	}
	sort.Strings(codes)
	type agg struct {
		total    int
		count    int
		source   string
		version  string
		evidence string
	}
	bySkill := map[string]agg{}
	for _, code := range codes {
		score := authoritative[code]
		skill := strings.TrimSpace(domain.TGOCodeToSkill[code])
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

func authoritativeObjectiveMap(scores []domain.ObjectiveScore) map[string]domain.ObjectiveScore {
	primary := map[string]domain.ObjectiveScore{}
	fallback := map[string]domain.ObjectiveScore{}
	for _, score := range scores {
		code := strings.TrimSpace(score.TGOCode)
		if code == "" {
			continue
		}
		if strings.TrimSpace(score.ScoreSource) == "deterministic" {
			primary[code] = score
			continue
		}
		if _, ok := fallback[code]; !ok {
			fallback[code] = score
		}
	}
	if len(primary) > 0 {
		return primary
	}
	return fallback
}

func authoritativeSkillMap(scores []domain.SkillScore) map[string]domain.SkillScore {
	primary := map[string]domain.SkillScore{}
	legacy := map[string]domain.SkillScore{}
	fallback := map[string]domain.SkillScore{}
	for _, score := range scores {
		skill := strings.TrimSpace(score.Skill)
		if skill == "" {
			continue
		}
		switch {
		case score.ScoreSource == "deterministic":
			primary[skill] = score
		case strings.Contains(score.ScoreSource, "legacy") || strings.TrimSpace(score.ScoreSource) == "":
			legacy[skill] = score
		default:
			if _, ok := fallback[skill]; !ok {
				fallback[skill] = score
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

func appliedRules(rawEvidence string) []string {
	rawEvidence = strings.TrimSpace(rawEvidence)
	if rawEvidence == "" || rawEvidence == "{}" {
		return nil
	}
	var payload struct {
		AppliedRules []string `json:"applied_rules"`
	}
	if err := json.Unmarshal([]byte(rawEvidence), &payload); err != nil {
		return nil
	}
	if len(payload.AppliedRules) == 0 {
		return nil
	}
	if len(payload.AppliedRules) > 4 {
		return append([]string{}, payload.AppliedRules[len(payload.AppliedRules)-4:]...)
	}
	return append([]string{}, payload.AppliedRules...)
}

func deltaExplanation(delta int, baselineRules, currentRules []string) string {
	if delta == 0 {
		return "Score held flat across drafts."
	}
	topGate := func(rules []string) string {
		for i := len(rules) - 1; i >= 0; i-- {
			if strings.Contains(strings.ToLower(rules[i]), "top score gate") {
				return rules[i]
			}
		}
		return ""
	}
	if gate := topGate(currentRules); gate != "" {
		if delta > 0 {
			return "Improved: " + gate
		}
		return "Regressed: " + gate
	}
	if len(currentRules) > 0 {
		if delta > 0 {
			return fmt.Sprintf("Improved by %+d from current deterministic rules.", delta)
		}
		return fmt.Sprintf("Dropped by %d from current deterministic rules.", -delta)
	}
	if len(baselineRules) > 0 {
		if delta > 0 {
			return fmt.Sprintf("Improved by %+d versus baseline rule state.", delta)
		}
		return fmt.Sprintf("Dropped by %d versus baseline rule state.", -delta)
	}
	if delta > 0 {
		return fmt.Sprintf("Improved by %+d.", delta)
	}
	return fmt.Sprintf("Dropped by %d.", -delta)
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

func objectiveEvidenceQuotes(code string, annotations []domain.ReviewAnnotation) []string {
	key := strings.TrimSpace(code)
	if key == "" || len(annotations) == 0 {
		return nil
	}
	out := make([]string, 0, 2)
	seen := map[string]bool{}
	for _, ann := range annotations {
		if strings.TrimSpace(ann.TGOCode) != key {
			continue
		}
		quote := strings.TrimSpace(ann.Quote)
		if quote == "" || seen[quote] {
			continue
		}
		out = append(out, quote)
		seen[quote] = true
		if len(out) == 2 {
			break
		}
	}
	return out
}

func sameSkillKeys(left, right map[string]domain.SkillScore) bool {
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

func sameObjectiveKeys(left, right map[string]domain.ObjectiveScore) bool {
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
