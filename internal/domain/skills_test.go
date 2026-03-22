package domain

import "testing"

func TestSkillDefinitionsAreUniqueAndTiered(t *testing.T) {
	seen := map[string]bool{}
	for _, def := range SkillDefinitions {
		if def.Name == "" {
			t.Fatalf("skill definition has empty name")
		}
		if seen[def.Name] {
			t.Fatalf("duplicate skill definition for %q", def.Name)
		}
		seen[def.Name] = true
		switch def.Tier {
		case SkillTierCore, SkillTierDomain, SkillTierSpecialty:
		default:
			t.Fatalf("skill %q has unsupported tier %q", def.Name, def.Tier)
		}
	}
}

func TestExpansionSkillsAreSupported(t *testing.T) {
	expected := []string{
		"lineation",
		"sonic patterning",
		"image logic",
		"stanza movement",
		"visual exposition",
		"beat design",
		"act structure",
		"oral cadence",
		"rhetorical repetition",
		"audience energy",
		"microcopy clarity",
		"error-state guidance",
		"information scent",
	}

	for _, skill := range expected {
		def, ok := SkillByName(skill)
		if !ok {
			t.Fatalf("expected %q to be supported", skill)
		}
		if def.Tier != SkillTierDomain {
			t.Fatalf("skill %q tier = %q, want %q", skill, def.Tier, SkillTierDomain)
		}
	}
}
