package domain

import (
	"slices"
	"testing"
)

func TestBuiltInTreesHaveDepthAndSeeds(t *testing.T) {
	if len(BuiltInTrees) < 18 {
		t.Fatalf("expected at least 18 built-in trees, got %d", len(BuiltInTrees))
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

func TestPublicBuiltInTreesMatchWritingDomains(t *testing.T) {
	options := AvailableOnboardingOptions()
	publicBySlug := PublicBuiltInTreeSlugs()
	if len(PublicBuiltInTrees) != len(options.WritingDomains)-1 {
		t.Fatalf("public built-in tree count = %d, want %d", len(PublicBuiltInTrees), len(options.WritingDomains)-1)
	}

	expected := []string{
		storyCraftTree.Slug,
		fantasyFictionTree.Slug,
		scienceFictionTree.Slug,
		romanceFictionTree.Slug,
		literaryFictionTree.Slug,
		mysteryThrillerTree.Slug,
		thoughtLeadershipTree.Slug,
		professionalWritingTree.Slug,
		marketingWritingTree.Slug,
		contentMarketingTree.Slug,
		journalismReportingTree.Slug,
		educationalWritingTree.Slug,
		grantWritingTree.Slug,
		academicEssayTree.Slug,
		technicalWritingTree.Slug,
		persuasiveWritingTree.Slug,
		memoirNarrativeTree.Slug,
	}
	for _, slug := range expected {
		if !publicBySlug[slug] {
			t.Fatalf("public built-in trees missing %s", slug)
		}
	}
	if publicBySlug[youthFoundationsTree.Slug] {
		t.Fatalf("youth foundations should not be public")
	}
}

func TestPublicBuiltInTreesStartWithTwoCoreAndOneDomainSeed(t *testing.T) {
	for _, tree := range PublicBuiltInTrees {
		coreCount := 0
		domainCount := 0
		for _, code := range tree.SeedCodes {
			tier := SkillTierForName(TGOCodeToSkill[code])
			switch tier {
			case SkillTierCore:
				coreCount++
			case SkillTierDomain:
				domainCount++
			default:
				t.Fatalf("tree %s seed %s has unsupported seed tier %q", tree.Slug, code, tier)
			}
		}
		if coreCount != 2 || domainCount != 1 {
			t.Fatalf("tree %s seeds = %#v, want 2 core and 1 domain", tree.Slug, tree.SeedCodes)
		}
	}
}

func TestLiteraryFictionSeedsLeadWithImageWork(t *testing.T) {
	if got, want := literaryFictionTree.SeedCodes, []string{
		"literary-story-causal-clarity",
		"literary-story-scene-architecture",
		"literary-story-description-focus",
	}; !slices.Equal(got, want) {
		t.Fatalf("literary fiction seeds = %#v, want %#v", got, want)
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

func TestBuiltInTreesValidateAsReachableDAGs(t *testing.T) {
	if err := ValidateBuiltInTrees(); err != nil {
		t.Fatalf("validate built-in trees: %v", err)
	}
}

func TestBuiltInTreesCanBeCheckedForPlanarity(t *testing.T) {
	checked := 0
	var nonPlanar []string
	for _, tree := range BuiltInTrees {
		if !TreeIsPlanar(tree) {
			nonPlanar = append(nonPlanar, tree.Slug)
		}
		checked++
	}
	if !TreeIsPlanar(GlobalSkillGraphDefinition()) {
		nonPlanar = append(nonPlanar, GlobalSkillGraphSlug)
	}
	if checked != len(BuiltInTrees) {
		t.Fatalf("checked %d trees, want %d", checked, len(BuiltInTrees))
	}
	if len(nonPlanar) > 0 {
		t.Logf("non-planar trees: %v", nonPlanar)
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

func TestRecommendedStarterCodesForBroaderFictionTemplates(t *testing.T) {
	profile := OnboardingProfile{
		WritingType:         "fantasy fiction",
		ExperienceLevel:     "advanced",
		DesiredTone:         "serious and emotional",
		BiggestWeaknesses:   []string{"scene work"},
		DesiredOutcomes:     []string{"novel"},
		DifficultyIntensity: "steady",
		WritingGoals:        "write stronger fantasy scenes",
	}

	starter := RecommendedStarterCodes(profile)
	if len(starter) != 3 {
		t.Fatalf("starter codes = %#v", starter)
	}
	if starter[0] != "fantasy-story-causal-clarity" {
		t.Fatalf("starter codes = %#v", starter)
	}
	regions := RecommendedRegionSlugs(profile)
	if len(regions) == 0 || regions[0] != fantasyFictionTree.Slug {
		t.Fatalf("regions = %#v", regions)
	}
}

func TestExpandedFictionTreesAreAvailable(t *testing.T) {
	trees := []TGOTreeDefinition{
		fantasyFictionTree,
		scienceFictionTree,
		romanceFictionTree,
		literaryFictionTree,
		mysteryThrillerTree,
	}

	for _, tree := range trees {
		if len(tree.TGOs) < 50 {
			t.Fatalf("tree %s has only %d TGOs", tree.Slug, len(tree.TGOs))
		}
		if len(tree.SeedCodes) != 3 {
			t.Fatalf("tree %s seed codes = %#v", tree.Slug, tree.SeedCodes)
		}
		if len(tree.PrioritySkills) < 6 {
			t.Fatalf("tree %s priority skills = %#v", tree.Slug, tree.PrioritySkills)
		}
	}
}

func TestExpandedNonfictionTreesAreAvailable(t *testing.T) {
	trees := []TGOTreeDefinition{
		marketingWritingTree,
		contentMarketingTree,
		journalismReportingTree,
		educationalWritingTree,
		grantWritingTree,
	}

	for _, tree := range trees {
		if len(tree.TGOs) < 45 {
			t.Fatalf("tree %s has only %d TGOs", tree.Slug, len(tree.TGOs))
		}
		if len(tree.SeedCodes) != 3 {
			t.Fatalf("tree %s seed codes = %#v", tree.Slug, tree.SeedCodes)
		}
		if len(tree.PrioritySkills) < 6 {
			t.Fatalf("tree %s priority skills = %#v", tree.Slug, tree.PrioritySkills)
		}
	}
}

func TestRecommendedStarterCodesForExpandedNonfictionTemplates(t *testing.T) {
	profile := OnboardingProfile{
		WritingType:         "marketing writing",
		ExperienceLevel:     "advanced",
		DesiredTone:         "clear and persuasive",
		BiggestWeaknesses:   []string{"positioning"},
		DesiredOutcomes:     []string{"stronger launch copy"},
		DifficultyIntensity: "steady",
		WritingGoals:        "write stronger marketing campaigns",
	}

	starter := RecommendedStarterCodes(profile)
	if len(starter) != 3 {
		t.Fatalf("starter codes = %#v", starter)
	}
	if starter[0] != "marketing-claim-clarity" {
		t.Fatalf("starter codes = %#v", starter)
	}
	regions := RecommendedRegionSlugs(profile)
	if len(regions) == 0 || regions[0] != marketingWritingTree.Slug {
		t.Fatalf("regions = %#v", regions)
	}
}

func TestEveryWritingTypeMapsToAlignedTemplate(t *testing.T) {
	cases := []struct {
		writingType  string
		templateKey  string
		baseTreeSlug string
	}{
		{writingType: "fiction", templateKey: "story-craft", baseTreeSlug: storyCraftTree.Slug},
		{writingType: "fantasy fiction", templateKey: "fantasy-fiction", baseTreeSlug: fantasyFictionTree.Slug},
		{writingType: "science fiction", templateKey: "science-fiction", baseTreeSlug: scienceFictionTree.Slug},
		{writingType: "romance", templateKey: "romance-fiction", baseTreeSlug: romanceFictionTree.Slug},
		{writingType: "literary fiction", templateKey: "literary-fiction", baseTreeSlug: literaryFictionTree.Slug},
		{writingType: "mystery", templateKey: "mystery-thriller", baseTreeSlug: mysteryThrillerTree.Slug},
		{writingType: "thought leadership", templateKey: "thought-leadership", baseTreeSlug: thoughtLeadershipTree.Slug},
		{writingType: "professional writing", templateKey: "professional-writing", baseTreeSlug: professionalWritingTree.Slug},
		{writingType: "marketing writing", templateKey: "marketing-writing", baseTreeSlug: marketingWritingTree.Slug},
		{writingType: "content marketing", templateKey: "content-marketing", baseTreeSlug: contentMarketingTree.Slug},
		{writingType: "journalism", templateKey: "journalism-reporting", baseTreeSlug: journalismReportingTree.Slug},
		{writingType: "educational writing", templateKey: "educational-writing", baseTreeSlug: educationalWritingTree.Slug},
		{writingType: "grant writing", templateKey: "grant-writing", baseTreeSlug: grantWritingTree.Slug},
		{writingType: "academic writing", templateKey: "academic-essay", baseTreeSlug: academicEssayTree.Slug},
		{writingType: "technical writing", templateKey: "technical-writing", baseTreeSlug: technicalWritingTree.Slug},
		{writingType: "persuasive writing", templateKey: "persuasive-writing", baseTreeSlug: persuasiveWritingTree.Slug},
		{writingType: "memoir", templateKey: "memoir-personal-narrative", baseTreeSlug: memoirNarrativeTree.Slug},
	}

	for _, tc := range cases {
		t.Run(tc.writingType, func(t *testing.T) {
			profile := OnboardingProfile{
				WritingType:         tc.writingType,
				AssignmentFormat:    "essay",
				TargetAudience:      "readers",
				SubjectMatter:       "real work",
				ExperienceLevel:     "intermediate",
				DesiredTone:         "clear",
				BiggestWeaknesses:   []string{"structure"},
				DesiredOutcomes:     []string{"stronger writing"},
				DifficultyIntensity: "steady",
				WritingGoals:        "improve craft",
			}

			if tc.writingType == "fiction" {
				profile.AssignmentFormat = "scene"
			}

			if got := TemplateKeyForProfile(profile); got != tc.templateKey {
				t.Fatalf("template key = %q, want %q", got, tc.templateKey)
			}

			base := treeForTemplateKey(tc.templateKey)
			if base.Slug != tc.baseTreeSlug {
				t.Fatalf("base tree slug = %q, want %q", base.Slug, tc.baseTreeSlug)
			}

			generated := GenerateTreeDefinition("writer", "Writer", profile)
			if len(generated.TGOs) != len(base.TGOs) {
				t.Fatalf("generated tree size = %d, want %d", len(generated.TGOs), len(base.TGOs))
			}
			if len(generated.SeedCodes) != len(base.SeedCodes) {
				t.Fatalf("generated seed count = %d, want %d", len(generated.SeedCodes), len(base.SeedCodes))
			}
			for i, seed := range base.SeedCodes {
				if generated.SeedCodes[i] != seed {
					t.Fatalf("generated seed %d = %q, want %q", i, generated.SeedCodes[i], seed)
				}
			}
			if len(generated.PrioritySkills) != len(base.PrioritySkills) {
				t.Fatalf("generated priority skill count = %d, want %d", len(generated.PrioritySkills), len(base.PrioritySkills))
			}
			for i, skill := range base.PrioritySkills {
				if generated.PrioritySkills[i] != skill {
					t.Fatalf("generated priority skill %d = %q, want %q", i, generated.PrioritySkills[i], skill)
				}
			}

			regions := RecommendedRegionSlugs(profile)
			if len(regions) == 0 || regions[0] != tc.baseTreeSlug {
				t.Fatalf("recommended regions = %#v, want first region %q", regions, tc.baseTreeSlug)
			}
		})
	}
}
