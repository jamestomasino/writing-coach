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

type TGOTreeDefinition struct {
	Slug           string
	Title          string
	Description    string
	SeedCodes      []string
	PrioritySkills []string
	TGOs           []TGO
}

var mythicTragedyTree = TGOTreeDefinition{
	Slug:           "mythic-tragedy-apprenticeship",
	Title:          WriterTrackName,
	Description:    "Advanced mythopoeic tragic fiction track",
	SeedCodes:      []string{"causal-clarity", "scene-architecture", "prose-precision"},
	PrioritySkills: []string{"tragic inevitability", "symbolic control", "mythic tone", "emotional compression", "scene architecture", "narrative clarity", "prose precision", "worldbuilding economy", "dialogue intelligence", "image freshness"},
	TGOs: []TGO{
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
	},
}

var youthFoundationsTree = TGOTreeDefinition{
	Slug:           "youth-writing-foundations",
	Title:          "Youth Writing Foundations",
	Description:    "Foundational writing track for younger writers building sentence, paragraph, and story control.",
	SeedCodes:      []string{"word-choice", "sentence-variety", "sentence-clarity"},
	PrioritySkills: []string{"word choice", "sentence variety", "clarity and coherence", "paragraph control", "narrative sequencing", "descriptive precision", "dialogue basics"},
	TGOs: []TGO{
		{Code: "word-choice", Title: "Word Choice", Description: "Choose specific words instead of vague filler.", Stage: "foundation", StageOrder: 1, MasteryHint: "The draft prefers concrete, accurate words over generic ones like nice, bad, thing, and stuff."},
		{Code: "sentence-variety", Title: "Sentence Variety", Description: "Mix short, medium, and longer sentences so the prose does not drag or stutter.", Stage: "foundation", StageOrder: 2, MasteryHint: "Sentence openings and lengths vary enough to keep the writing moving naturally."},
		{Code: "sentence-clarity", Title: "Sentence Clarity", Description: "Make each sentence easy to follow the first time.", Stage: "foundation", StageOrder: 3, MasteryHint: "Most sentences read cleanly without confusion about who did what."},
		{Code: "paragraph-control", Title: "Paragraph Control", Description: "Group related ideas into readable paragraphs.", Stage: "foundation", StageOrder: 4, Prerequisites: []string{"sentence-clarity"}, MasteryHint: "Paragraphs stay focused on one small unit of action or thought."},
		{Code: "narrative-sequencing", Title: "Narrative Sequencing", Description: "Put events in a clear order with useful transitions.", Stage: "story", StageOrder: 5, Prerequisites: []string{"sentence-clarity", "paragraph-control"}, MasteryHint: "The reader can retell the story events in the right order without guessing."},
		{Code: "descriptive-specificity", Title: "Descriptive Specificity", Description: "Use concrete details that help the reader picture the scene.", Stage: "story", StageOrder: 6, Prerequisites: []string{"word-choice"}, MasteryHint: "Description relies on a few strong details instead of broad labels."},
		{Code: "sentence-complexity", Title: "Sentence Complexity", Description: "Control longer sentences without losing clarity.", Stage: "craft", StageOrder: 7, Prerequisites: []string{"sentence-variety", "sentence-clarity"}, MasteryHint: "Longer sentences still stay grammatically stable and easy to follow."},
		{Code: "dialogue-basics", Title: "Dialogue Basics", Description: "Use dialogue to show speaker intent and keep the reader oriented.", Stage: "story", StageOrder: 8, Prerequisites: []string{"sentence-clarity"}, MasteryHint: "Dialogue is easy to track and does more than fill space."},
	},
}

var BuiltInTrees = []TGOTreeDefinition{
	mythicTragedyTree,
	youthFoundationsTree,
}

var TGOCodeToSkill = map[string]string{
	"causal-clarity":          "narrative clarity",
	"scene-architecture":      "scene architecture",
	"prose-precision":         "prose precision",
	"emotional-compression":   "emotional compression",
	"tragic-inevitability":    "tragic inevitability",
	"symbolic-discipline":     "symbolic control",
	"mythic-register":         "mythic tone",
	"worldbuilding-economy":   "worldbuilding economy",
	"dialogue-under-strain":   "dialogue intelligence",
	"image-freshness":         "image freshness",
	"word-choice":             "word choice",
	"sentence-variety":        "sentence variety",
	"sentence-clarity":        "clarity and coherence",
	"paragraph-control":       "paragraph control",
	"narrative-sequencing":    "narrative sequencing",
	"descriptive-specificity": "descriptive precision",
	"sentence-complexity":     "sentence complexity",
	"dialogue-basics":         "dialogue basics",
}

func AllTGOs() []TGO {
	var out []TGO
	for _, tree := range BuiltInTrees {
		out = append(out, tree.TGOs...)
	}
	return out
}

func BuiltInTreeBySlug(slug string) (TGOTreeDefinition, bool) {
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
		return mythicTragedyTree.SeedCodes
	}
	return SeedCodesForDefinition(tree)
}

func NextUnlockedTGOs(treeSlug string, completed map[string]bool, active map[string]bool, limit int) []TGO {
	tree, ok := BuiltInTreeBySlug(treeSlug)
	if !ok {
		tree = mythicTragedyTree
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
