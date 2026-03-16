package domain

import (
	"fmt"
	"regexp"
	"strings"
)

type OnboardingProfile struct {
	UserID              int64
	WritingType         string
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
	return strings.TrimSpace(p.WritingType) != "" &&
		strings.TrimSpace(p.ExperienceLevel) != "" &&
		strings.TrimSpace(p.DesiredTone) != "" &&
		len(p.BiggestWeaknesses) > 0 &&
		len(p.DesiredOutcomes) > 0 &&
		strings.TrimSpace(p.DifficultyIntensity) != "" &&
		strings.TrimSpace(p.WritingGoals) != ""
}

func GenerateTreeDefinition(userSlug, userName string, profile OnboardingProfile) TGOTreeDefinition {
	templateKey := selectTemplate(profile)
	var def TGOTreeDefinition
	switch templateKey {
	case "youth-foundations":
		def = youthFoundationsTree
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
		return "story-craft"
	default:
		if experience == "beginner" {
			return "youth-foundations"
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

var storyCraftTree = TGOTreeDefinition{
	Slug:           "story-craft-track",
	Title:          "Story Craft Track",
	Description:    "Flexible fiction track for writers building scene, character pressure, and narrative clarity.",
	SeedCodes:      []string{"causal-clarity", "scene-architecture", "prose-precision"},
	PrioritySkills: []string{"narrative clarity", "scene architecture", "prose precision", "emotional compression", "dialogue intelligence", "image freshness", "worldbuilding economy"},
	TGOs: []TGO{
		{Code: "causal-clarity", Title: "Causal Clarity", Description: "Make action and consequence legible beat by beat.", Stage: "core", StageOrder: 1, MasteryHint: "Readers can follow each decision and its immediate consequence without backtracking."},
		{Code: "scene-architecture", Title: "Scene Architecture", Description: "Stage turns, entrances, exits, and power shifts cleanly.", Stage: "core", StageOrder: 2, MasteryHint: "Scenes remain spatially and dramatically legible even under pressure."},
		{Code: "prose-precision", Title: "Prose Precision", Description: "Replace soft modifiers with exact nouns and verbs.", Stage: "core", StageOrder: 3, MasteryHint: "Line-level revision sharpens verbs, nouns, and cadence without ornament."},
		{Code: "emotional-compression", Title: "Emotional Compression", Description: "Condense feeling into image, gesture, and consequence.", Stage: "story", StageOrder: 4, Prerequisites: []string{"causal-clarity"}, MasteryHint: "Feeling arrives through action and image rather than named emotion."},
		{Code: "dialogue-under-strain", Title: "Dialogue Under Strain", Description: "Use speech to reveal conflict and hierarchy under pressure.", Stage: "story", StageOrder: 5, Prerequisites: []string{"scene-architecture"}, MasteryHint: "Dialogue carries conflict instead of stopping the scene."},
		{Code: "descriptive-specificity", Title: "Descriptive Specificity", Description: "Use concrete details that sharpen the reader’s picture.", Stage: "story", StageOrder: 6, Prerequisites: []string{"prose-precision"}, MasteryHint: "Description relies on a few exact details rather than generic labels."},
		{Code: "worldbuilding-economy", Title: "Worldbuilding Economy", Description: "Imply context through pressure instead of explanation.", Stage: "story", StageOrder: 7, Prerequisites: []string{"scene-architecture"}, MasteryHint: "Background arrives through conflict-bearing details."},
		{Code: "image-freshness", Title: "Image Freshness", Description: "Prefer singular imagery over stock language.", Stage: "genre", StageOrder: 8, Prerequisites: []string{"descriptive-specificity"}, MasteryHint: "Images feel specific to the story at hand."},
	},
}

var thoughtLeadershipTree = TGOTreeDefinition{
	Slug:           "thought-leadership-track",
	Title:          "Thought Leadership Track",
	Description:    "Nonfiction track for sharper claims, stronger insight, and more authoritative structure.",
	SeedCodes:      []string{"claim-clarity", "audience-alignment", "sentence-economy"},
	PrioritySkills: []string{"claim clarity", "audience alignment", "sentence economy", "structural signposting", "insight density", "evidence integration", "authority and voice", "clarity and coherence"},
	TGOs: []TGO{
		{Code: "claim-clarity", Title: "Claim Clarity", Description: "State the central argument early and plainly.", Stage: "core", StageOrder: 1, MasteryHint: "The controlling claim is visible without excavation."},
		{Code: "audience-alignment", Title: "Audience Alignment", Description: "Aim the piece at a clearly understood reader and need.", Stage: "core", StageOrder: 2, MasteryHint: "The piece feels written for a specific reader, not the void."},
		{Code: "sentence-economy", Title: "Sentence Economy", Description: "Cut throat-clearing and reduce bloated syntax.", Stage: "core", StageOrder: 3, MasteryHint: "Most sentences justify their length and remain easy to track."},
		{Code: "structural-signposting", Title: "Structural Signposting", Description: "Guide the reader through the argument with clear sectional moves.", Stage: "argument", StageOrder: 4, Prerequisites: []string{"claim-clarity"}, MasteryHint: "Major turns in the argument are easy to follow."},
		{Code: "insight-density", Title: "Insight Density", Description: "Favor original value over generic summary.", Stage: "argument", StageOrder: 5, Prerequisites: []string{"claim-clarity", "audience-alignment"}, MasteryHint: "Paragraphs deliver actual insight, not only competent explanation."},
		{Code: "evidence-integration", Title: "Evidence Integration", Description: "Use examples and evidence to sharpen the claim instead of decorating it.", Stage: "argument", StageOrder: 6, Prerequisites: []string{"structural-signposting"}, MasteryHint: "Evidence advances the reasoning instead of merely appearing beside it."},
		{Code: "authority-and-voice", Title: "Authority and Voice", Description: "Sound decisive without becoming inflated or vague.", Stage: "voice", StageOrder: 7, Prerequisites: []string{"sentence-economy"}, MasteryHint: "The voice feels credible, direct, and earned."},
		{Code: "openings-that-frame", Title: "Openings That Frame", Description: "Open with the right tension, problem, or stake.", Stage: "voice", StageOrder: 8, Prerequisites: []string{"audience-alignment"}, MasteryHint: "The opening earns attention and frames the essay’s purpose."},
		{Code: "endings-that-land", Title: "Endings That Land", Description: "Close with consequence rather than summary mush.", Stage: "voice", StageOrder: 9, Prerequisites: []string{"insight-density"}, MasteryHint: "The ending leaves the reader with a sharpened conclusion or next move."},
	},
}

var professionalWritingTree = TGOTreeDefinition{
	Slug:           "professional-writing-track",
	Title:          "Professional Writing Track",
	Description:    "Practical writing track for clarity, structure, tone, and actionability in workplace communication.",
	SeedCodes:      []string{"objective-clarity", "audience-alignment", "sentence-economy"},
	PrioritySkills: []string{"clarity and coherence", "audience alignment", "sentence economy", "structural signposting", "tone calibration", "actionability", "scannability", "evidence integration"},
	TGOs: []TGO{
		{Code: "objective-clarity", Title: "Objective Clarity", Description: "Make the document’s purpose unmistakable.", Stage: "core", StageOrder: 1, MasteryHint: "Readers know what the document is trying to do within the first lines."},
		{Code: "audience-alignment", Title: "Audience Alignment", Description: "Match detail, tone, and framing to the intended reader.", Stage: "core", StageOrder: 2, MasteryHint: "The level of context and framing matches the audience’s needs."},
		{Code: "sentence-economy", Title: "Sentence Economy", Description: "Reduce clutter and keep sentences easy to scan.", Stage: "core", StageOrder: 3, MasteryHint: "Sentences are concise without becoming abrupt or vague."},
		{Code: "structural-signposting", Title: "Structural Signposting", Description: "Use headings, transitions, and ordering that reduce reader effort.", Stage: "document", StageOrder: 4, Prerequisites: []string{"objective-clarity"}, MasteryHint: "The document’s structure is obvious and useful."},
		{Code: "tone-calibration", Title: "Tone Calibration", Description: "Sound professional without becoming stiff, passive, or evasive.", Stage: "document", StageOrder: 5, Prerequisites: []string{"audience-alignment"}, MasteryHint: "Tone supports trust and clarity."},
		{Code: "actionability", Title: "Actionability", Description: "Make decisions, asks, and next steps explicit.", Stage: "document", StageOrder: 6, Prerequisites: []string{"objective-clarity", "structural-signposting"}, MasteryHint: "Readers can act without having to infer the ask."},
		{Code: "evidence-integration", Title: "Evidence Integration", Description: "Use relevant data, examples, or constraints to support decisions.", Stage: "document", StageOrder: 7, Prerequisites: []string{"actionability"}, MasteryHint: "Support strengthens the recommendation instead of cluttering it."},
		{Code: "scannability", Title: "Scannability", Description: "Shape prose and sections so busy readers can extract the key message quickly.", Stage: "document", StageOrder: 8, Prerequisites: []string{"sentence-economy", "structural-signposting"}, MasteryHint: "A fast pass still captures the essential points."},
	},
}
