package domain

type SkillTier string

const (
	SkillTierCore      SkillTier = "core"
	SkillTierDomain    SkillTier = "domain"
	SkillTierSpecialty SkillTier = "specialty"
)

type SkillDefinition struct {
	Name string
	Tier SkillTier
}

var SkillDefinitions = []SkillDefinition{
	{Name: "clarity and coherence", Tier: SkillTierCore},
	{Name: "claim clarity", Tier: SkillTierCore},
	{Name: "audience alignment", Tier: SkillTierCore},
	{Name: "structural signposting", Tier: SkillTierCore},
	{Name: "evidence integration", Tier: SkillTierCore},
	{Name: "narrative clarity", Tier: SkillTierCore},
	{Name: "scene architecture", Tier: SkillTierCore},
	{Name: "actionability", Tier: SkillTierCore},
	{Name: "scannability", Tier: SkillTierCore},
	{Name: "voice presence", Tier: SkillTierCore},

	{Name: "sentence economy", Tier: SkillTierDomain},
	{Name: "prose precision", Tier: SkillTierDomain},
	{Name: "emotional compression", Tier: SkillTierDomain},
	{Name: "dialogue intelligence", Tier: SkillTierDomain},
	{Name: "worldbuilding economy", Tier: SkillTierDomain},
	{Name: "word choice", Tier: SkillTierDomain},
	{Name: "sentence variety", Tier: SkillTierDomain},
	{Name: "sentence complexity", Tier: SkillTierDomain},
	{Name: "paragraph control", Tier: SkillTierDomain},
	{Name: "narrative sequencing", Tier: SkillTierDomain},
	{Name: "descriptive precision", Tier: SkillTierDomain},
	{Name: "dialogue basics", Tier: SkillTierDomain},
	{Name: "insight density", Tier: SkillTierDomain},
	{Name: "authority and voice", Tier: SkillTierDomain},
	{Name: "tone calibration", Tier: SkillTierDomain},
	{Name: "accuracy", Tier: SkillTierDomain},
	{Name: "analysis depth", Tier: SkillTierDomain},
	{Name: "assignment alignment", Tier: SkillTierDomain},
	{Name: "example quality", Tier: SkillTierDomain},
	{Name: "grammar control", Tier: SkillTierDomain},
	{Name: "objection handling", Tier: SkillTierDomain},
	{Name: "professional format", Tier: SkillTierDomain},
	{Name: "reasoning quality", Tier: SkillTierDomain},
	{Name: "reflection depth", Tier: SkillTierDomain},
	{Name: "revision habits", Tier: SkillTierDomain},
	{Name: "rhetorical force", Tier: SkillTierDomain},
	{Name: "source handling", Tier: SkillTierDomain},
	{Name: "spelling and mechanics", Tier: SkillTierDomain},
	{Name: "story development", Tier: SkillTierDomain},
	{Name: "structure and pacing", Tier: SkillTierDomain},
	{Name: "technical precision", Tier: SkillTierDomain},
	{Name: "thesis clarity", Tier: SkillTierDomain},
	{Name: "user goal alignment", Tier: SkillTierDomain},
	{Name: "image freshness", Tier: SkillTierDomain},
	{Name: "lineation", Tier: SkillTierDomain},
	{Name: "sonic patterning", Tier: SkillTierDomain},
	{Name: "image logic", Tier: SkillTierDomain},
	{Name: "stanza movement", Tier: SkillTierDomain},
	{Name: "visual exposition", Tier: SkillTierDomain},
	{Name: "beat design", Tier: SkillTierDomain},
	{Name: "act structure", Tier: SkillTierDomain},
	{Name: "oral cadence", Tier: SkillTierDomain},
	{Name: "rhetorical repetition", Tier: SkillTierDomain},
	{Name: "audience energy", Tier: SkillTierDomain},
	{Name: "microcopy clarity", Tier: SkillTierDomain},
	{Name: "error-state guidance", Tier: SkillTierDomain},
	{Name: "information scent", Tier: SkillTierDomain},

	{Name: "symbolic control", Tier: SkillTierSpecialty},
	{Name: "mythic tone", Tier: SkillTierSpecialty},
	{Name: "tragic inevitability", Tier: SkillTierSpecialty},
}

var SupportedSkills = supportedSkillNames()

func supportedSkillNames() []string {
	out := make([]string, 0, len(SkillDefinitions))
	for _, skill := range SkillDefinitions {
		out = append(out, skill.Name)
	}
	return out
}

func IsSupportedSkill(skill string) bool {
	_, ok := SkillByName(skill)
	return ok
}

func SkillByName(skill string) (SkillDefinition, bool) {
	for _, value := range SkillDefinitions {
		if value.Name == skill {
			return value, true
		}
	}
	return SkillDefinition{}, false
}

func SkillTierForName(skill string) SkillTier {
	def, ok := SkillByName(skill)
	if !ok {
		return ""
	}
	return def.Tier
}
