package domain

const WriterTrackName = "Mythopoeic Tragic Apprenticeship"

var HouseCurriculum = map[string]string{
	"tragic inevitability":  "Build scenes where choice closes futures rather than opening them.",
	"symbolic control":      "Let objects carry fate without explaining their meaning.",
	"mythic tone":           "Sustain gravity without drifting into inflated diction.",
	"emotional compression": "Condense feeling into image, gesture, and consequence.",
	"scene architecture":    "Stage turns clearly so ritual and conflict remain legible.",
	"narrative clarity":     "Clarify causal sequence and relationship stakes without overexposition.",
	"prose precision":       "Replace soft modifiers with exact nouns and verbs.",
	"worldbuilding economy": "Imply history through pressure, not encyclopedic explanation.",
	"dialogue intelligence": "Make speech reveal rank, motive, and fracture under restraint.",
	"image freshness":       "Prefer singular, earned imagery over fantasy stock language.",
}

func PrioritySkillsForTree(treeSlug string) []string {
	tree, ok := BuiltInTreeBySlug(treeSlug)
	if !ok || len(tree.PrioritySkills) == 0 {
		return []string{
			"tragic inevitability",
			"symbolic control",
			"mythic tone",
			"emotional compression",
			"scene architecture",
			"narrative clarity",
			"prose precision",
			"worldbuilding economy",
			"dialogue intelligence",
			"image freshness",
		}
	}
	return append([]string(nil), tree.PrioritySkills...)
}

func SkillPriority(treeSlug, skill string) int {
	prioritySkills := PrioritySkillsForTree(treeSlug)
	for idx, value := range prioritySkills {
		if value == skill {
			return len(prioritySkills) - idx
		}
	}
	return 0
}
