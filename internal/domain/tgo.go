package domain

type TGO struct {
	ID            int64
	Code          string
	Title         string
	Description   string
	Stage         string
	StageOrder    int
	ActiveSlot    int
	Prerequisites []string
	MasteryHint   string
}

type TGOAssessment struct {
	TGOCode  string
	Status   string
	Evidence string
	ReviewID int64
}

var TGOCatalog = []TGO{
	{Code: "causal-clarity", Title: "Causal Clarity", Description: "Make action and consequence legible beat by beat.", Stage: "core", StageOrder: 1, MasteryHint: "Readers can follow each decision and its immediate consequence without backtracking."},
	{Code: "scene-architecture", Title: "Scene Architecture", Description: "Stage turns, entrances, exits, and power shifts cleanly.", Stage: "core", StageOrder: 2, MasteryHint: "Scenes remain spatially and dramatically legible even under pressure."},
	{Code: "prose-precision", Title: "Prose Precision", Description: "Replace soft modifiers with exact nouns and verbs.", Stage: "core", StageOrder: 3, MasteryHint: "Line-level revision consistently sharpens verbs, nouns, and cadence without puffing the diction."},
	{Code: "emotional-compression", Title: "Emotional Compression", Description: "Condense feeling into image, gesture, and consequence.", Stage: "core", StageOrder: 4, Prerequisites: []string{"causal-clarity"}, MasteryHint: "Feeling arrives through action and image rather than named emotion."},
	{Code: "tragic-inevitability", Title: "Tragic Inevitability", Description: "Make choices close futures rather than open them.", Stage: "mythic", StageOrder: 5, Prerequisites: []string{"causal-clarity", "scene-architecture"}, MasteryHint: "The ending feels sealed by prior choices, not imposed by authorial summary."},
	{Code: "worldbuilding-economy", Title: "Worldbuilding Economy", Description: "Imply history through pressure, not exposition.", Stage: "mythic", StageOrder: 6, Prerequisites: []string{"scene-architecture"}, MasteryHint: "Setting, history, and politics arrive through conflict-bearing details."},
	{Code: "symbolic-discipline", Title: "Symbolic Discipline", Description: "Let objects carry fate without direct explanation.", Stage: "mythic", StageOrder: 7, Prerequisites: []string{"prose-precision", "emotional-compression"}, MasteryHint: "Symbols recur with pressure and coherence without being glossed for the reader."},
	{Code: "dialogue-under-strain", Title: "Dialogue Under Strain", Description: "Use speech to reveal rank, fracture, and motive under pressure.", Stage: "genre", StageOrder: 8, Prerequisites: []string{"scene-architecture", "prose-precision"}, MasteryHint: "Dialogue carries conflict and hierarchy without flattening into exposition."},
	{Code: "mythic-register", Title: "Mythic Register", Description: "Sustain gravity without inflated diction or pastiche.", Stage: "genre", StageOrder: 9, Prerequisites: []string{"symbolic-discipline", "prose-precision"}, MasteryHint: "The prose feels elevated and disciplined rather than imitative or overwritten."},
	{Code: "image-freshness", Title: "Image Freshness", Description: "Prefer singular earned imagery over fantasy stock language.", Stage: "genre", StageOrder: 10, Prerequisites: []string{"mythic-register"}, MasteryHint: "Images feel specific to the scene rather than drawn from generic fantasy vocabulary."},
}

var TGOCodeToSkill = map[string]string{
	"causal-clarity":        "narrative clarity",
	"scene-architecture":    "scene architecture",
	"prose-precision":       "prose precision",
	"emotional-compression": "emotional compression",
	"tragic-inevitability":  "tragic inevitability",
	"symbolic-discipline":   "symbolic control",
	"mythic-register":       "mythic tone",
	"worldbuilding-economy": "worldbuilding economy",
	"dialogue-under-strain": "dialogue intelligence",
	"image-freshness":       "image freshness",
}

func TGOByCode(code string) (TGO, bool) {
	for _, tgo := range TGOCatalog {
		if tgo.Code == code {
			return tgo, true
		}
	}
	return TGO{}, false
}

func SeedTGOs() []string {
	return []string{"causal-clarity", "scene-architecture", "prose-precision"}
}

func NextUnlockedTGOs(completed map[string]bool, active map[string]bool, limit int) []TGO {
	var out []TGO
	for _, tgo := range TGOCatalog {
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
