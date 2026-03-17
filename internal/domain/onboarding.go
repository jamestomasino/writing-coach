package domain

import (
	"fmt"
	"regexp"
	"strings"
)

type OnboardingProfile struct {
	UserID              int64
	WritingType         string
	AssignmentFormat    string
	TargetAudience      string
	SubjectMatter       string
	ExperienceLevel     string
	DesiredTone         string
	BiggestWeaknesses   []string
	DesiredOutcomes     []string
	DifficultyIntensity string
	WritingGoals        string
	GeneratedTreeSlug   string
	TemplateKey         string
}

func (p OnboardingProfile) Complete() bool {
	return len(p.MissingFields()) == 0
}

func (p OnboardingProfile) MissingFields() []string {
	var missing []string
	if strings.TrimSpace(p.WritingType) == "" {
		missing = append(missing, "primary writing domain")
	}
	if strings.TrimSpace(p.AssignmentFormat) == "" {
		missing = append(missing, "common assignment format")
	}
	if strings.TrimSpace(p.TargetAudience) == "" {
		missing = append(missing, "target audience")
	}
	if strings.TrimSpace(p.SubjectMatter) == "" {
		missing = append(missing, "typical subject matter")
	}
	if strings.TrimSpace(p.ExperienceLevel) == "" {
		missing = append(missing, "experience level")
	}
	if strings.TrimSpace(p.DesiredTone) == "" {
		missing = append(missing, "tone target")
	}
	if len(p.BiggestWeaknesses) == 0 {
		missing = append(missing, "biggest weaknesses")
	}
	if len(p.DesiredOutcomes) == 0 {
		missing = append(missing, "desired outcomes")
	}
	if strings.TrimSpace(p.DifficultyIntensity) == "" {
		missing = append(missing, "difficulty and intensity")
	}
	if strings.TrimSpace(p.WritingGoals) == "" {
		missing = append(missing, "writing goals")
	}
	return missing
}

func GenerateTreeDefinition(userSlug, userName string, profile OnboardingProfile) TGOTreeDefinition {
	templateKey := selectTemplate(profile)
	var def TGOTreeDefinition
	switch templateKey {
	case "youth-foundations":
		def = youthFoundationsTree
	case "academic-essay":
		def = academicEssayTree
	case "technical-writing":
		def = technicalWritingTree
	case "persuasive-writing":
		def = persuasiveWritingTree
	case "memoir-personal-narrative":
		def = memoirNarrativeTree
	case "thought-leadership":
		def = thoughtLeadershipTree
	case "professional-writing":
		def = professionalWritingTree
	case "story-craft":
		def = storyCraftTree
	default:
		def = mythicTragedyTree
		templateKey = "mythic-tragedy"
	}

	def.Slug = generatedTreeSlug(userSlug)
	def.Title = generatedTreeTitle(userName, profile, def.Title)
	def.Description = generatedTreeDescription(profile, def.Description)
	return def
}

func TemplateKeyForProfile(profile OnboardingProfile) string {
	return selectTemplate(profile)
}

func CoachingBrief(profile OnboardingProfile) string {
	parts := []string{}
	if value := strings.TrimSpace(profile.WritingType); value != "" {
		parts = append(parts, "writing type: "+value)
	}
	if value := strings.TrimSpace(profile.AssignmentFormat); value != "" {
		parts = append(parts, "format: "+value)
	}
	if value := strings.TrimSpace(profile.TargetAudience); value != "" {
		parts = append(parts, "audience: "+value)
	}
	if value := strings.TrimSpace(profile.SubjectMatter); value != "" {
		parts = append(parts, "subject matter: "+value)
	}
	if value := PromptToneGuidance(profile); value != "" {
		parts = append(parts, "tone guidance: "+value)
	}
	if len(profile.DesiredOutcomes) > 0 && strings.TrimSpace(profile.DesiredOutcomes[0]) != "" {
		parts = append(parts, "primary goal: "+strings.TrimSpace(profile.DesiredOutcomes[0]))
	}
	if len(profile.BiggestWeaknesses) > 0 && strings.TrimSpace(profile.BiggestWeaknesses[0]) != "" {
		parts = append(parts, "watch for: "+strings.TrimSpace(profile.BiggestWeaknesses[0]))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "; ")
}

func PromptToneGuidance(profile OnboardingProfile) string {
	raw := strings.TrimSpace(profile.DesiredTone)
	if raw == "" {
		return ""
	}

	normalized := strings.ToLower(raw)
	replacer := strings.NewReplacer("/", ",", ";", ",", "&", ",", " and ", ",")
	parts := strings.Split(replacer.Replace(normalized), ",")

	seen := map[string]bool{}
	var guidance []string
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
		guidance = append(guidance, value)
	}

	for _, part := range parts {
		part = strings.TrimSpace(part)
		switch part {
		case "serious":
			add("keep the mood weighty and controlled")
		case "emotional":
			add("keep the emotional stakes close to the surface")
		case "clear":
			add("make each sentence easy to follow")
		case "persuasive":
			add("use direct claims and concrete stakes")
		case "analytical":
			add("favor clean reasoning over ornament")
		case "decisive":
			add("sound confident and specific")
		case "practical":
			add("prefer useful specifics over abstraction")
		case "concise":
			add("keep the language lean")
		case "grave":
			add("avoid breezy or playful phrasing")
		case "mythic":
			add("use elevated but readable language with a sense of larger forces")
		case "funny", "humorous", "comic":
			add("allow brief moments of wit without undercutting the main pressure")
		case "warm":
			add("let the voice feel humane and open")
		case "formal":
			add("keep the diction disciplined and composed")
		}
	}

	if len(guidance) == 0 {
		return raw
	}
	if len(guidance) > 3 {
		guidance = guidance[:3]
	}
	return strings.Join(guidance, "; ")
}

func PromptGoal(profile OnboardingProfile) string {
	if value := strings.TrimSpace(profile.WritingGoals); value != "" {
		return value
	}
	if len(profile.DesiredOutcomes) > 0 {
		return strings.TrimSpace(profile.DesiredOutcomes[0])
	}
	return ""
}

func PromptProfileLines(profile OnboardingProfile) []string {
	lines := []string{}
	if value := strings.TrimSpace(profile.WritingType); value != "" {
		lines = append(lines, "writing domain: "+value)
	}
	if value := strings.TrimSpace(profile.AssignmentFormat); value != "" {
		lines = append(lines, "assignment format: "+value)
	}
	if value := strings.TrimSpace(profile.TargetAudience); value != "" {
		lines = append(lines, "target audience: "+value)
	}
	if value := strings.TrimSpace(profile.SubjectMatter); value != "" {
		lines = append(lines, "prompt material: draw from "+value)
	}
	if value := PromptToneGuidance(profile); value != "" {
		lines = append(lines, "tone guidance: "+value)
	}
	if value := PromptGoal(profile); value != "" {
		lines = append(lines, "goal: "+value)
	}
	if len(profile.DesiredOutcomes) > 0 {
		lines = append(lines, "desired outcomes: "+strings.Join(profile.DesiredOutcomes, ", "))
	}
	return lines
}

var slugCleaner = regexp.MustCompile(`[^a-z0-9]+`)

func generatedTreeSlug(userSlug string) string {
	cleaned := strings.ToLower(strings.TrimSpace(userSlug))
	cleaned = slugCleaner.ReplaceAllString(cleaned, "-")
	cleaned = strings.Trim(cleaned, "-")
	if cleaned == "" {
		cleaned = "writer"
	}
	return fmt.Sprintf("%s-track", cleaned)
}

func generatedTreeTitle(userName string, profile OnboardingProfile, fallback string) string {
	name := strings.TrimSpace(userName)
	if name == "" {
		name = "Writer"
	}
	switch selectTemplate(profile) {
	case "youth-foundations":
		return fmt.Sprintf("%s's Foundations Track", name)
	case "academic-essay":
		return fmt.Sprintf("%s's Academic Essay Track", name)
	case "technical-writing":
		return fmt.Sprintf("%s's Technical Writing Track", name)
	case "persuasive-writing":
		return fmt.Sprintf("%s's Persuasive Writing Track", name)
	case "memoir-personal-narrative":
		return fmt.Sprintf("%s's Memoir Track", name)
	case "thought-leadership":
		return fmt.Sprintf("%s's Thought Leadership Track", name)
	case "professional-writing":
		return fmt.Sprintf("%s's Professional Writing Track", name)
	case "story-craft":
		return fmt.Sprintf("%s's Story Craft Track", name)
	default:
		return fmt.Sprintf("%s's Mythic Fiction Track", name)
	}
}

func generatedTreeDescription(profile OnboardingProfile, fallback string) string {
	parts := []string{fallback}
	if tone := strings.TrimSpace(profile.DesiredTone); tone != "" {
		parts = append(parts, "Tone target: "+tone+".")
	}
	if goals := strings.TrimSpace(profile.WritingGoals); goals != "" {
		parts = append(parts, "Goals: "+goals)
	}
	return strings.Join(parts, " ")
}

func selectTemplate(profile OnboardingProfile) string {
	writingType := strings.ToLower(strings.TrimSpace(profile.WritingType))
	experience := strings.ToLower(strings.TrimSpace(profile.ExperienceLevel))
	toneAndGoals := strings.ToLower(strings.Join([]string{profile.DesiredTone, profile.WritingGoals}, " "))

	switch writingType {
	case "academic", "academic writing", "essay", "essay writing", "research", "research writing":
		return "academic-essay"
	case "technical", "technical writing", "documentation", "docs":
		return "technical-writing"
	case "persuasive", "persuasive writing", "argument", "argumentative writing", "advocacy":
		return "persuasive-writing"
	case "memoir", "personal narrative", "nonfiction narrative":
		return "memoir-personal-narrative"
	case "thought leadership", "thought-leadership":
		return "thought-leadership"
	case "professional", "professional writing", "professional-writing":
		return "professional-writing"
	case "fiction", "story", "stories":
		if experience == "beginner" {
			return "youth-foundations"
		}
		if strings.Contains(toneAndGoals, "myth") || strings.Contains(toneAndGoals, "fantasy") || strings.Contains(toneAndGoals, "tragic") {
			return "mythic-tragedy"
		}
		if strings.Contains(toneAndGoals, "memoir") || strings.Contains(toneAndGoals, "personal narrative") {
			return "memoir-personal-narrative"
		}
		return "story-craft"
	default:
		if experience == "beginner" {
			return "youth-foundations"
		}
		if strings.Contains(toneAndGoals, "academic") || strings.Contains(toneAndGoals, "essay") || strings.Contains(toneAndGoals, "research") {
			return "academic-essay"
		}
		if strings.Contains(toneAndGoals, "technical writing") || strings.Contains(toneAndGoals, "documentation") || strings.Contains(toneAndGoals, "docs") {
			return "technical-writing"
		}
		if strings.Contains(toneAndGoals, "persuasive") || strings.Contains(toneAndGoals, "argument") || strings.Contains(toneAndGoals, "advocacy") {
			return "persuasive-writing"
		}
		if strings.Contains(toneAndGoals, "memoir") || strings.Contains(toneAndGoals, "personal narrative") {
			return "memoir-personal-narrative"
		}
		if strings.Contains(toneAndGoals, "thought leadership") {
			return "thought-leadership"
		}
		if strings.Contains(toneAndGoals, "professional") || strings.Contains(toneAndGoals, "business") {
			return "professional-writing"
		}
	}
	return "story-craft"
}
