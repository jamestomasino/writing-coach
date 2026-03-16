package domain

var SupportedSkills = []string{
	"prose precision",
	"image freshness",
	"symbolic control",
	"emotional compression",
	"scene architecture",
	"dialogue intelligence",
	"worldbuilding economy",
	"mythic tone",
	"tragic inevitability",
	"narrative clarity",
	"word choice",
	"sentence variety",
	"sentence complexity",
	"clarity and coherence",
	"paragraph control",
	"narrative sequencing",
	"descriptive precision",
	"dialogue basics",
	"claim clarity",
	"audience alignment",
	"sentence economy",
	"structural signposting",
	"insight density",
	"evidence integration",
	"authority and voice",
	"tone calibration",
	"actionability",
	"scannability",
}

func IsSupportedSkill(skill string) bool {
	for _, value := range SupportedSkills {
		if value == skill {
			return true
		}
	}
	return false
}
