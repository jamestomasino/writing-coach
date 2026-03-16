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
}

func IsSupportedSkill(skill string) bool {
	for _, value := range SupportedSkills {
		if value == skill {
			return true
		}
	}
	return false
}
