package domain

const WriterTrackName = "Mythopoeic Tragic Apprenticeship"

var PrioritySkills = []string{
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

func SkillPriority(skill string) int {
	for idx, value := range PrioritySkills {
		if value == skill {
			return len(PrioritySkills) - idx
		}
	}
	return 0
}
