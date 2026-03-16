package domain

import "testing"

func TestBuiltInTreesHaveDepthAndSeeds(t *testing.T) {
	if len(BuiltInTrees) < 9 {
		t.Fatalf("expected at least 9 built-in trees, got %d", len(BuiltInTrees))
	}
	for _, tree := range BuiltInTrees {
		if len(tree.TGOs) < 50 {
			t.Fatalf("tree %s has only %d TGOs", tree.Slug, len(tree.TGOs))
		}
		if len(tree.SeedCodes) != 3 {
			t.Fatalf("tree %s should expose exactly 3 seed codes", tree.Slug)
		}
		if len(tree.PrioritySkills) < 6 {
			t.Fatalf("tree %s should expose a richer priority skill list", tree.Slug)
		}
	}
}

func TestBuiltInTreesHaveUniqueCodesAndValidPrerequisites(t *testing.T) {
	allCodes := map[string]string{}
	for _, tree := range BuiltInTrees {
		inTree := map[string]bool{}
		for _, tgo := range tree.TGOs {
			if inTree[tgo.Code] {
				t.Fatalf("duplicate code %s inside tree %s", tgo.Code, tree.Slug)
			}
			inTree[tgo.Code] = true
			if existingTree, exists := allCodes[tgo.Code]; exists && existingTree != tree.Slug {
				t.Fatalf("duplicate code %s shared by trees %s and %s", tgo.Code, existingTree, tree.Slug)
			}
			allCodes[tgo.Code] = tree.Slug
			if TGOCodeToSkill[tgo.Code] == "" {
				t.Fatalf("missing skill mapping for %s", tgo.Code)
			}
		}
		for _, seed := range tree.SeedCodes {
			if !inTree[seed] {
				t.Fatalf("seed code %s missing from tree %s", seed, tree.Slug)
			}
		}
		for _, tgo := range tree.TGOs {
			for _, prereq := range tgo.Prerequisites {
				if !inTree[prereq] {
					t.Fatalf("tree %s references missing prerequisite %s from %s", tree.Slug, prereq, tgo.Code)
				}
			}
		}
	}
}

func TestGeneratedTemplatesStayLarge(t *testing.T) {
	cases := []OnboardingProfile{
		{WritingType: "fiction", ExperienceLevel: "advanced", DesiredTone: "mythic tragic fantasy", BiggestWeaknesses: []string{"symbolism"}, DesiredOutcomes: []string{"novel"}, DifficultyIntensity: "high", WritingGoals: "write mythic tragedy"},
		{WritingType: "fiction", ExperienceLevel: "intermediate", DesiredTone: "literary", BiggestWeaknesses: []string{"scene work"}, DesiredOutcomes: []string{"stories"}, DifficultyIntensity: "medium", WritingGoals: "write stronger fiction"},
		{WritingType: "thought leadership", ExperienceLevel: "intermediate", DesiredTone: "clear and sharp", BiggestWeaknesses: []string{"structure"}, DesiredOutcomes: []string{"essays"}, DifficultyIntensity: "medium", WritingGoals: "publish thought leadership"},
		{WritingType: "professional writing", ExperienceLevel: "intermediate", DesiredTone: "direct", BiggestWeaknesses: []string{"clarity"}, DesiredOutcomes: []string{"memos"}, DifficultyIntensity: "medium", WritingGoals: "improve workplace communication"},
		{WritingType: "fiction", ExperienceLevel: "beginner", DesiredTone: "playful", BiggestWeaknesses: []string{"spelling"}, DesiredOutcomes: []string{"school stories"}, DifficultyIntensity: "low", WritingGoals: "learn to write better stories"},
		{WritingType: "academic writing", ExperienceLevel: "intermediate", DesiredTone: "analytical", BiggestWeaknesses: []string{"thesis"}, DesiredOutcomes: []string{"papers"}, DifficultyIntensity: "medium", WritingGoals: "write stronger essays"},
		{WritingType: "technical writing", ExperienceLevel: "intermediate", DesiredTone: "clear", BiggestWeaknesses: []string{"docs structure"}, DesiredOutcomes: []string{"documentation"}, DifficultyIntensity: "medium", WritingGoals: "write better docs"},
		{WritingType: "persuasive writing", ExperienceLevel: "intermediate", DesiredTone: "confident", BiggestWeaknesses: []string{"argument"}, DesiredOutcomes: []string{"op-eds"}, DifficultyIntensity: "medium", WritingGoals: "write stronger persuasive pieces"},
		{WritingType: "memoir", ExperienceLevel: "intermediate", DesiredTone: "reflective", BiggestWeaknesses: []string{"reflection"}, DesiredOutcomes: []string{"essays"}, DifficultyIntensity: "medium", WritingGoals: "write personal narrative"},
	}
	for _, profile := range cases {
		def := GenerateTreeDefinition("writer", "Writer", profile)
		if len(def.TGOs) < 50 {
			t.Fatalf("generated tree for %s has only %d TGOs", profile.WritingType, len(def.TGOs))
		}
	}
}

func TestGlobalSkillGraphDefinition(t *testing.T) {
	graph := GlobalSkillGraphDefinition()
	if graph.Slug != GlobalSkillGraphSlug {
		t.Fatalf("graph slug = %q", graph.Slug)
	}
	if len(graph.TGOs) < 450 {
		t.Fatalf("global graph too small: %d", len(graph.TGOs))
	}
	if len(graph.SeedCodes) != 3 {
		t.Fatalf("global graph seeds = %#v", graph.SeedCodes)
	}
}

func TestRecommendedStarterCodes(t *testing.T) {
	profile := OnboardingProfile{
		WritingType:         "technical writing",
		ExperienceLevel:     "intermediate",
		DesiredTone:         "clear",
		BiggestWeaknesses:   []string{"structure"},
		DesiredOutcomes:     []string{"documentation"},
		DifficultyIntensity: "steady",
		WritingGoals:        "write better docs",
	}
	starter := RecommendedStarterCodes(profile)
	if len(starter) != 3 {
		t.Fatalf("starter codes = %#v", starter)
	}
	if starter[0] != "technical-user-goal" {
		t.Fatalf("starter codes = %#v", starter)
	}
	regions := RecommendedRegionSlugs(profile)
	if len(regions) == 0 || regions[0] != technicalWritingTree.Slug {
		t.Fatalf("regions = %#v", regions)
	}
}
