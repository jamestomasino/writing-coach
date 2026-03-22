package domain

var HouseCurriculum = map[string]string{
	"narrative clarity":     "Clarify causal sequence and relationship stakes without overexposition.",
	"scene architecture":    "Stage turns clearly so conflict remains legible under pressure.",
	"prose precision":       "Replace soft modifiers with exact nouns and verbs.",
	"emotional compression": "Condense feeling into image, gesture, and consequence.",
	"dialogue intelligence": "Make speech reveal rank, motive, and fracture under restraint.",
	"image freshness":       "Prefer singular, earned imagery over stock language.",
	"worldbuilding economy": "Imply history through pressure, not encyclopedic explanation.",
	"story development":     "Build character want, contradiction, and change through consequence.",
}

func PrioritySkillsForTree(treeSlug string) []string {
	tree, ok := BuiltInTreeBySlug(treeSlug)
	if !ok || len(tree.PrioritySkills) == 0 {
		return append([]string(nil), storyCraftTree.PrioritySkills...)
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
