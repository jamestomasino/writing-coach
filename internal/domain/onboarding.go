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
