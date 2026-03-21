package domain

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

type OnboardingProfile struct {
	EnrollmentID        int64
	UserID              int64
	WritingLanguage     string
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

type OnboardingOption struct {
	Value string
	Label string
}

type OnboardingOptions struct {
	WritingLanguages  []OnboardingOption
	WritingDomains    []OnboardingOption
	AssignmentFormats []OnboardingOption
	ExperienceLevels  []OnboardingOption
	DifficultyLevels  []OnboardingOption
	Weaknesses        []OnboardingOption
	DesiredOutcomes   []OnboardingOption
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

func AvailableOnboardingOptions() OnboardingOptions {
	return OnboardingOptions{
		WritingLanguages: []OnboardingOption{
			{Value: "en", Label: "English"},
		},
		WritingDomains: []OnboardingOption{
			{Value: "fiction", Label: "Fiction"},
			{Value: "fantasy fiction", Label: "Fantasy fiction"},
			{Value: "science fiction", Label: "Science fiction"},
			{Value: "romance", Label: "Romance"},
			{Value: "literary fiction", Label: "Literary fiction"},
			{Value: "mystery", Label: "Mystery / thriller"},
			{Value: "thought leadership", Label: "Thought leadership"},
			{Value: "professional writing", Label: "Professional writing"},
			{Value: "marketing writing", Label: "Marketing writing"},
			{Value: "content marketing", Label: "Content marketing"},
			{Value: "journalism", Label: "Journalism / reporting"},
			{Value: "educational writing", Label: "Educational writing"},
			{Value: "grant writing", Label: "Grant writing"},
			{Value: "academic writing", Label: "Academic writing"},
			{Value: "technical writing", Label: "Technical writing"},
			{Value: "persuasive writing", Label: "Persuasive writing"},
			{Value: "memoir", Label: "Memoir / personal narrative"},
			{Value: "other", Label: "Other"},
		},
		AssignmentFormats: []OnboardingOption{
			{Value: "scene", Label: "scene"},
			{Value: "short story", Label: "short story"},
			{Value: "blog post", Label: "blog post"},
			{Value: "op-ed", Label: "op-ed"},
			{Value: "memo", Label: "memo"},
			{Value: "email", Label: "email"},
			{Value: "landing page", Label: "landing page"},
			{Value: "product announcement", Label: "product announcement"},
			{Value: "essay", Label: "essay"},
			{Value: "how-to guide", Label: "how-to guide"},
		},
		ExperienceLevels: []OnboardingOption{
			{Value: "beginner", Label: "Beginner"},
			{Value: "intermediate", Label: "Intermediate"},
			{Value: "advanced", Label: "Advanced"},
		},
		DifficultyLevels: []OnboardingOption{
			{Value: "steady", Label: "Steady"},
			{Value: "ambitious", Label: "Ambitious"},
			{Value: "gentle", Label: "Gentle"},
		},
		Weaknesses: []OnboardingOption{
			{Value: "word choice", Label: "word choice"},
			{Value: "sentence variety", Label: "sentence variety"},
			{Value: "sentence economy", Label: "sentence economy"},
			{Value: "paragraph control", Label: "paragraph control"},
			{Value: "narrative clarity", Label: "narrative clarity"},
			{Value: "scene architecture", Label: "scene architecture"},
			{Value: "symbolic control", Label: "symbolic control"},
			{Value: "tone calibration", Label: "tone calibration"},
			{Value: "evidence integration", Label: "evidence integration"},
		},
		DesiredOutcomes: []OnboardingOption{
			{Value: "publish stronger fiction", Label: "publish stronger fiction"},
			{Value: "write clearer essays", Label: "write clearer essays"},
			{Value: "improve professional communication", Label: "improve professional communication"},
			{Value: "develop a distinctive voice", Label: "develop a distinctive voice"},
			{Value: "build revision discipline", Label: "build revision discipline"},
			{Value: "write thought leadership with authority", Label: "write thought leadership with authority"},
		},
	}
}

func GenerateTreeDefinition(userSlug, userName string, profile OnboardingProfile) TGOTreeDefinition {
	templateKey := selectTemplate(profile)
	def := treeForTemplateKey(templateKey)

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
	if value := WritingLanguageLabel(profile.WritingLanguage); strings.TrimSpace(value) != "" {
		parts = append(parts, "writing language: "+value)
	}
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
	if value := WritingLanguageLabel(profile.WritingLanguage); strings.TrimSpace(value) != "" {
		lines = append(lines, "writing language: "+value)
	}
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
	if value := PromptScenarioGuidance(profile); value != "" {
		lines = append(lines, "assignment seed: "+value)
	}
	if len(profile.DesiredOutcomes) > 0 {
		lines = append(lines, "desired outcomes: "+strings.Join(profile.DesiredOutcomes, ", "))
	}
	return lines
}

func PromptScenarioGuidance(profile OnboardingProfile) string {
	format := strings.TrimSpace(profile.AssignmentFormat)
	subject := strings.TrimSpace(profile.SubjectMatter)
	audience := strings.TrimSpace(profile.TargetAudience)
	domainLabel := strings.TrimSpace(profile.WritingType)

	switch {
	case looksLikeFictionFormat(format) || looksLikeFictionDomain(domainLabel):
		parts := []string{"build the assignment around one concrete character situation or turning problem"}
		if subject != "" {
			parts = append(parts, "drawn from "+subject)
		}
		parts = append(parts, "give the writer a specific pressure point, choice, or conflict instead of a generic setup")
		return strings.Join(parts, "; ")
	default:
		parts := []string{"build the assignment around one concrete communication situation"}
		if subject != "" {
			parts = append(parts, "drawn from "+subject)
		}
		if audience != "" {
			parts = append(parts, "for "+audience)
		}
		parts = append(parts, "give the writer a real problem, decision, or message to handle instead of a generic topic")
		return strings.Join(parts, "; ")
	}
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
	return generatedTrackLabel(profile, fallback)
}

func generatedTreeDescription(profile OnboardingProfile, fallback string) string {
	parts := []string{generatedTrackSummary(profile)}
	if tone := strings.TrimSpace(profile.DesiredTone); tone != "" {
		parts = append(parts, "Tone target: "+tone+".")
	}
	if goals := strings.TrimSpace(profile.WritingGoals); goals != "" {
		parts = append(parts, "Goals: "+goals)
	}
	return strings.Join(parts, " ")
}

func GeneratedTreeDisplay(userName string, profile OnboardingProfile, fallbackTitle, fallbackDescription string) (string, string) {
	return generatedTreeTitle(userName, profile, fallbackTitle), generatedTreeDescription(profile, fallbackDescription)
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
	case "marketing", "marketing writing", "copywriting":
		return "marketing-writing"
	case "content marketing", "content strategy", "brand content":
		return "content-marketing"
	case "journalism", "reporting", "journalism and reporting":
		return "journalism-reporting"
	case "educational writing", "instructional writing", "explanatory writing":
		return "educational-writing"
	case "grant writing", "grant proposal", "grants":
		return "grant-writing"
	case "fantasy", "fantasy fiction", "epic fantasy", "urban fantasy":
		return "fantasy-fiction"
	case "science fiction", "sci-fi", "sci fi", "speculative fiction":
		return "science-fiction"
	case "romance", "romance fiction", "romantic fiction":
		return "romance-fiction"
	case "literary fiction", "literary":
		return "literary-fiction"
	case "mystery", "thriller", "mystery thriller", "crime fiction":
		return "mystery-thriller"
	case "fiction", "story", "stories":
		if experience == "beginner" {
			return "youth-foundations"
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
		if strings.Contains(toneAndGoals, "fantasy") {
			return "fantasy-fiction"
		}
		if strings.Contains(toneAndGoals, "science fiction") || strings.Contains(toneAndGoals, "sci-fi") || strings.Contains(toneAndGoals, "sci fi") {
			return "science-fiction"
		}
		if strings.Contains(toneAndGoals, "romance") {
			return "romance-fiction"
		}
		if strings.Contains(toneAndGoals, "literary fiction") {
			return "literary-fiction"
		}
		if strings.Contains(toneAndGoals, "mystery") || strings.Contains(toneAndGoals, "thriller") {
			return "mystery-thriller"
		}
		if strings.Contains(toneAndGoals, "thought leadership") {
			return "thought-leadership"
		}
		if strings.Contains(toneAndGoals, "marketing") || strings.Contains(toneAndGoals, "copywriting") {
			return "marketing-writing"
		}
		if strings.Contains(toneAndGoals, "content marketing") || strings.Contains(toneAndGoals, "content strategy") {
			return "content-marketing"
		}
		if strings.Contains(toneAndGoals, "journalism") || strings.Contains(toneAndGoals, "reporting") {
			return "journalism-reporting"
		}
		if strings.Contains(toneAndGoals, "educational writing") || strings.Contains(toneAndGoals, "instructional writing") || strings.Contains(toneAndGoals, "explanatory writing") {
			return "educational-writing"
		}
		if strings.Contains(toneAndGoals, "grant writing") || strings.Contains(toneAndGoals, "grant proposal") {
			return "grant-writing"
		}
		if strings.Contains(toneAndGoals, "professional") || strings.Contains(toneAndGoals, "business") {
			return "professional-writing"
		}
	}
	return "story-craft"
}

func looksLikeFictionFormat(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, token := range []string{"scene", "story", "chapter", "dialogue", "monologue"} {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func looksLikeFictionDomain(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	for _, token := range []string{"fiction", "fantasy", "novel", "short story", "memoir", "narrative"} {
		if strings.Contains(value, token) {
			return true
		}
	}
	return false
}

func generatedTrackLabel(profile OnboardingProfile, fallback string) string {
	if label := strings.TrimSpace(profile.WritingType); label != "" {
		label = displayCase(label)
		if !strings.Contains(strings.ToLower(label), "track") {
			label += " Track"
		}
		return label
	}

	switch selectTemplate(profile) {
	case "youth-foundations":
		return "Foundations Track"
	case "academic-essay":
		return "Academic Essay Track"
	case "technical-writing":
		return "Technical Writing Track"
	case "persuasive-writing":
		return "Persuasive Writing Track"
	case "memoir-personal-narrative":
		return "Memoir Track"
	case "thought-leadership":
		return "Thought Leadership Track"
	case "professional-writing":
		return "Professional Writing Track"
	case "marketing-writing":
		return "Marketing Writing Track"
	case "content-marketing":
		return "Content Marketing Track"
	case "journalism-reporting":
		return "Journalism and Reporting Track"
	case "educational-writing":
		return "Educational Writing Track"
	case "grant-writing":
		return "Grant Writing Track"
	case "fantasy-fiction":
		return "Fantasy Track"
	case "science-fiction":
		return "Science Fiction Track"
	case "romance-fiction":
		return "Romance Track"
	case "literary-fiction":
		return "Literary Fiction Track"
	case "mystery-thriller":
		return "Mystery and Thriller Track"
	case "story-craft":
		return "Story Craft Track"
	default:
		if strings.TrimSpace(fallback) != "" {
			return fallback
		}
		return "Writing Track"
	}
}

func generatedTrackSummary(profile OnboardingProfile) string {
	writingType := strings.TrimSpace(profile.WritingType)
	format := strings.TrimSpace(profile.AssignmentFormat)
	audience := strings.TrimSpace(profile.TargetAudience)

	switch {
	case writingType != "" && format != "" && audience != "":
		return fmt.Sprintf("Skill track for %s, with %s assignments for %s.", writingType, format, audience)
	case writingType != "" && format != "":
		return fmt.Sprintf("Skill track for %s, with a focus on %s assignments.", writingType, format)
	case writingType != "" && audience != "":
		return fmt.Sprintf("Skill track for %s, shaped for %s.", writingType, audience)
	case writingType != "":
		return fmt.Sprintf("Skill track for %s.", writingType)
	case format != "" && audience != "":
		return fmt.Sprintf("Skill track for %s assignments for %s.", format, audience)
	case format != "":
		return fmt.Sprintf("Skill track for %s assignments.", format)
	default:
		return "Skill track for writing practice."
	}
}

func displayCase(value string) string {
	words := strings.Fields(value)
	for i, word := range words {
		runes := []rune(strings.ToLower(word))
		if len(runes) == 0 {
			continue
		}
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}
