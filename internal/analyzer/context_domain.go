package analyzer

import "strings"

const (
	DomainGeneral            = "general"
	DomainFiction            = "fiction"
	DomainFantasy            = "fantasy"
	DomainTechnical          = "technical"
	DomainAcademic           = "academic"
	DomainProfessional       = "professional"
	DomainThoughtLeadership  = "thought_leadership"
	DomainMarketing          = "marketing"
)

func DomainForContext(options ContextOptions) string {
	templateKey := strings.ToLower(strings.TrimSpace(options.TemplateKey))
	writingType := strings.ToLower(strings.TrimSpace(options.WritingType))
	assignmentFormat := strings.ToLower(strings.TrimSpace(options.AssignmentFormat))
	treeSlug := strings.ToLower(strings.TrimSpace(options.TreeSlug))

	switch {
	case containsAny([]string{templateKey, treeSlug}, "fantasy-fiction", "mythic-tragedy"):
		return DomainFantasy
	case containsAny([]string{templateKey, treeSlug}, "science-fiction", "romance-fiction", "literary-fiction", "mystery-thriller", "story-craft", "memoir-personal-narrative", "youth-foundations"):
		return DomainFiction
	case containsAny([]string{templateKey, treeSlug}, "technical-writing", "educational-writing"):
		return DomainTechnical
	case containsAny([]string{templateKey, treeSlug}, "academic-essay"):
		return DomainAcademic
	case containsAny([]string{templateKey, treeSlug}, "professional-writing", "grant-writing"):
		return DomainProfessional
	case containsAny([]string{templateKey, treeSlug}, "thought-leadership", "journalism-reporting", "persuasive-writing"):
		return DomainThoughtLeadership
	case containsAny([]string{templateKey, treeSlug}, "marketing-writing", "content-marketing"):
		return DomainMarketing
	}

	combined := strings.Join([]string{writingType, assignmentFormat, treeSlug}, " ")
	switch {
	case strings.Contains(combined, "fantasy"):
		return DomainFantasy
	case strings.Contains(combined, "fiction") || strings.Contains(combined, "scene") || strings.Contains(combined, "short story") || strings.Contains(combined, "memoir"):
		return DomainFiction
	case strings.Contains(combined, "technical") || strings.Contains(combined, "documentation") || strings.Contains(combined, "how-to") || strings.Contains(combined, "guide") || strings.Contains(combined, "educational"):
		return DomainTechnical
	case strings.Contains(combined, "academic") || strings.Contains(combined, "essay") || strings.Contains(combined, "research"):
		return DomainAcademic
	case strings.Contains(combined, "marketing") || strings.Contains(combined, "landing page") || strings.Contains(combined, "product announcement"):
		return DomainMarketing
	case strings.Contains(combined, "thought leadership") || strings.Contains(combined, "journalism") || strings.Contains(combined, "reporting") || strings.Contains(combined, "persuasive"):
		return DomainThoughtLeadership
	case strings.Contains(combined, "professional") || strings.Contains(combined, "memo") || strings.Contains(combined, "email") || strings.Contains(combined, "grant"):
		return DomainProfessional
	default:
		return DomainGeneral
	}
}
