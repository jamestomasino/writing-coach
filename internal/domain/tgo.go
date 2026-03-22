package domain

import "math"

const (
	ProgressModeStage   = "stage"
	ProgressModePercent = "percent"
)

type TGO struct {
	ID             int64
	Code           string
	Title          string
	Description    string
	Stage          string
	StageOrder     int
	ActiveSlot     int
	Prerequisites  []string
	MasteryHint    string
	ProgressMode   string
	MasteryStage   string
	MasteryPercent int
	EvidenceCount  int
}

type TGOAssessment struct {
	TGOCode  string
	Status   string
	Evidence string
	ReviewID int64
}

type TGOTreeDefinition struct {
	Slug           string
	Title          string
	Description    string
	SeedCodes      []string
	PrioritySkills []string
	TGOs           []TGO
}

type TGOMasterySignal struct {
	ProgressMode  string
	Percent       int
	Stage         string
	EvidenceCount int
	Ready         bool
}

func InferProgressMode(skill string) string {
	switch skill {
	case "word choice", "sentence variety", "sentence economy", "sentence complexity", "paragraph control", "narrative sequencing",
		"claim clarity", "evidence integration", "spelling and mechanics", "grammar control", "actionability", "clarity and coherence",
		"causal clarity", "sentence flow", "reasoning quality", "professional format", "scannability":
		return ProgressModePercent
	default:
		return ProgressModeStage
	}
}

func ComputeMasterySignal(progressMode string, statuses []string) TGOMasterySignal {
	if progressMode == "" {
		progressMode = ProgressModeStage
	}
	if len(statuses) == 0 {
		return TGOMasterySignal{
			ProgressMode:  progressMode,
			Percent:       0,
			Stage:         "emerging",
			EvidenceCount: 0,
			Ready:         false,
		}
	}

	var totalWeight, weighted float64
	recentStrong := 0
	recentMastered := 0
	recentWindow := minInt(len(statuses), 3)
	for i, status := range statuses {
		weight := float64(maxInt(1, 5-i))
		totalWeight += weight
		score := statusScore(status)
		weighted += score * weight
		if i < recentWindow {
			if status == "secure" || status == "mastered" {
				recentStrong++
			}
			if status == "mastered" {
				recentMastered++
			}
		}
	}

	avg := weighted / totalWeight
	ready := len(statuses) >= 3 && recentStrong == recentWindow && recentMastered >= 2 && avg >= 0.88
	percent := int(math.Round(avg * 100))
	switch {
	case ready:
		percent = maxInt(percent, 100)
	case len(statuses) < 2 && percent > 70:
		percent = 70
	case len(statuses) < 3 && percent > 85:
		percent = 85
	case percent > 95:
		percent = 95
	}

	stage := "emerging"
	switch {
	case ready:
		stage = "mastery evidence"
	case avg >= 0.78:
		stage = "strong control"
	case avg >= 0.5:
		stage = "developing"
	}

	return TGOMasterySignal{
		ProgressMode:  progressMode,
		Percent:       percent,
		Stage:         stage,
		EvidenceCount: len(statuses),
		Ready:         ready,
	}
}

func statusScore(status string) float64 {
	switch status {
	case "mastered":
		return 1.0
	case "secure":
		return 0.72
	case "developing":
		return 0.35
	default:
		return 0.5
	}
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func AllTGOs() []TGO {
	var out []TGO
	for _, tree := range BuiltInTrees {
		out = append(out, tree.TGOs...)
	}
	return out
}

func BuiltInTreeBySlug(slug string) (TGOTreeDefinition, bool) {
	slug = NormalizeTreeSlug(slug)
	for _, tree := range BuiltInTrees {
		if tree.Slug == slug {
			return tree, true
		}
	}
	return TGOTreeDefinition{}, false
}

func TGOByCode(code string) (TGO, bool) {
	for _, tgo := range AllTGOs() {
		if tgo.Code == code {
			return tgo, true
		}
	}
	return TGO{}, false
}

func SeedTGOs(treeSlug string) []string {
	tree, ok := BuiltInTreeBySlug(treeSlug)
	if !ok {
		return storyCraftTree.SeedCodes
	}
	return SeedCodesForDefinition(tree)
}

func NextUnlockedTGOs(treeSlug string, completed map[string]bool, active map[string]bool, limit int) []TGO {
	tree, ok := BuiltInTreeBySlug(treeSlug)
	if !ok {
		tree = storyCraftTree
	}
	return NextUnlockedFromDefinition(tree, completed, active, limit)
}

func SeedCodesForDefinition(tree TGOTreeDefinition) []string {
	return append([]string(nil), tree.SeedCodes...)
}

func PrioritySkillsForDefinition(tree TGOTreeDefinition) []string {
	return append([]string(nil), tree.PrioritySkills...)
}

func NextUnlockedFromDefinition(tree TGOTreeDefinition, completed map[string]bool, active map[string]bool, limit int) []TGO {
	var out []TGO
	for _, tgo := range tree.TGOs {
		if completed[tgo.Code] || active[tgo.Code] || !prereqsMet(tgo, completed) {
			continue
		}
		out = append(out, tgo)
		if len(out) == limit {
			break
		}
	}
	return out
}

func prereqsMet(tgo TGO, completed map[string]bool) bool {
	for _, prereq := range tgo.Prerequisites {
		if !completed[prereq] {
			return false
		}
	}
	return true
}
