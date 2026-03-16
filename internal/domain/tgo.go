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
