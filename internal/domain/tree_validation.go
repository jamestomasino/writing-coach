package domain

import (
	"fmt"
	"slices"
	"strings"

	gigraph "github.com/Tom-Johnston/mamba/graph"
)

func ValidateTreeDefinition(tree TGOTreeDefinition) error {
	if strings.TrimSpace(tree.Slug) == "" {
		return fmt.Errorf("tree slug is required")
	}
	if strings.TrimSpace(tree.Title) == "" {
		return fmt.Errorf("tree %s title is required", tree.Slug)
	}
	if len(tree.TGOs) == 0 {
		return fmt.Errorf("tree %s must contain at least one TGO", tree.Slug)
	}
	if len(tree.SeedCodes) != 3 {
		return fmt.Errorf("tree %s must expose exactly 3 seed codes", tree.Slug)
	}

	codeToTGO := make(map[string]TGO, len(tree.TGOs))
	indexByCode := make(map[string]int, len(tree.TGOs))
	for i, tgo := range tree.TGOs {
		if strings.TrimSpace(tgo.Code) == "" {
			return fmt.Errorf("tree %s contains a TGO with empty code", tree.Slug)
		}
		if _, exists := codeToTGO[tgo.Code]; exists {
			return fmt.Errorf("tree %s contains duplicate code %s", tree.Slug, tgo.Code)
		}
		codeToTGO[tgo.Code] = tgo
		indexByCode[tgo.Code] = i
	}

	for _, seed := range tree.SeedCodes {
		tgo, ok := codeToTGO[seed]
		if !ok {
			return fmt.Errorf("tree %s references missing seed %s", tree.Slug, seed)
		}
		if len(tgo.Prerequisites) > 0 {
			return fmt.Errorf("tree %s seed %s cannot have prerequisites", tree.Slug, seed)
		}
	}

	for _, tgo := range tree.TGOs {
		seenPrereqs := make(map[string]bool, len(tgo.Prerequisites))
		for _, prereq := range tgo.Prerequisites {
			if prereq == tgo.Code {
				return fmt.Errorf("tree %s node %s cannot depend on itself", tree.Slug, tgo.Code)
			}
			if seenPrereqs[prereq] {
				return fmt.Errorf("tree %s node %s repeats prerequisite %s", tree.Slug, tgo.Code, prereq)
			}
			seenPrereqs[prereq] = true

			prereqNode, ok := codeToTGO[prereq]
			if !ok {
				return fmt.Errorf("tree %s references missing prerequisite %s from %s", tree.Slug, prereq, tgo.Code)
			}
			_ = prereqNode
		}
	}

	if cycle := firstCycle(tree); len(cycle) > 0 {
		return fmt.Errorf("tree %s contains a cycle: %s", tree.Slug, strings.Join(cycle, " -> "))
	}

	if tree.Slug != GlobalSkillGraphSlug {
		unreachable := unreachableNodes(tree)
		if len(unreachable) > 0 {
			slices.Sort(unreachable)
			return fmt.Errorf("tree %s contains nodes unreachable from seeds: %s", tree.Slug, strings.Join(unreachable, ", "))
		}
	}

	return nil
}

func ValidateBuiltInTrees() error {
	allCodes := map[string]string{}
	for _, tree := range BuiltInTrees {
		if err := ValidateTreeDefinition(tree); err != nil {
			return err
		}
		for _, tgo := range tree.TGOs {
			if existingTree, exists := allCodes[tgo.Code]; exists && existingTree != tree.Slug {
				return fmt.Errorf("duplicate code %s shared by trees %s and %s", tgo.Code, existingTree, tree.Slug)
			}
			allCodes[tgo.Code] = tree.Slug
		}
	}
	if err := ValidateTreeDefinition(GlobalSkillGraphDefinition()); err != nil {
		return err
	}
	return nil
}

func MustValidateBuiltInTrees() {
	if err := ValidateBuiltInTrees(); err != nil {
		panic(err)
	}
}

func firstCycle(tree TGOTreeDefinition) []string {
	codeToTGO := make(map[string]TGO, len(tree.TGOs))
	for _, tgo := range tree.TGOs {
		codeToTGO[tgo.Code] = tgo
	}

	const (
		unseen = iota
		visiting
		done
	)
	state := make(map[string]int, len(tree.TGOs))
	var stack []string
	var cycle []string

	var visit func(string) bool
	visit = func(code string) bool {
		state[code] = visiting
		stack = append(stack, code)
		for _, prereq := range codeToTGO[code].Prerequisites {
			switch state[prereq] {
			case unseen:
				if visit(prereq) {
					return true
				}
			case visiting:
				start := slices.Index(stack, prereq)
				cycle = append(append([]string(nil), stack[start:]...), prereq)
				return true
			}
		}
		stack = stack[:len(stack)-1]
		state[code] = done
		return false
	}

	for _, tgo := range tree.TGOs {
		if state[tgo.Code] == unseen && visit(tgo.Code) {
			return cycle
		}
	}
	return nil
}

func unreachableNodes(tree TGOTreeDefinition) []string {
	unlocks := make(map[string][]string, len(tree.TGOs))
	for _, tgo := range tree.TGOs {
		for _, prereq := range tgo.Prerequisites {
			unlocks[prereq] = append(unlocks[prereq], tgo.Code)
		}
	}

	reachable := make(map[string]bool, len(tree.TGOs))
	queue := append([]string(nil), tree.SeedCodes...)
	for len(queue) > 0 {
		code := queue[0]
		queue = queue[1:]
		if reachable[code] {
			continue
		}
		reachable[code] = true
		queue = append(queue, unlocks[code]...)
	}

	var unreachable []string
	for _, tgo := range tree.TGOs {
		if !reachable[tgo.Code] {
			unreachable = append(unreachable, tgo.Code)
		}
	}
	return unreachable
}

func isPlanar(tree TGOTreeDefinition) bool {
	g := gigraph.NewDense(len(tree.TGOs), nil)
	indexByCode := make(map[string]int, len(tree.TGOs))
	for i, tgo := range tree.TGOs {
		indexByCode[tgo.Code] = i
	}
	for _, tgo := range tree.TGOs {
		target := indexByCode[tgo.Code]
		for _, prereq := range tgo.Prerequisites {
			source := indexByCode[prereq]
			if source == target || g.IsEdge(source, target) {
				continue
			}
			g.AddEdge(source, target)
		}
	}
	return gigraph.IsPlanar(g)
}

func TreeIsPlanar(tree TGOTreeDefinition) bool {
	return isPlanar(tree)
}
