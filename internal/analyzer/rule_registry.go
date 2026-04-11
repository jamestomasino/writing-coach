package analyzer

import "strings"

// CoverageLayer identifies whether a deterministic rule is global, domain-level, or specialty-level.
type CoverageLayer string

const (
	LayerGlobal    CoverageLayer = "global"
	LayerDomain    CoverageLayer = "domain"
	LayerSpecialty CoverageLayer = "specialty"
)

// RuleOwner identifies which analyzer family should primarily own a signal.
type RuleOwner string

const (
	OwnerHeuristic    RuleOwner = "heuristic"
	OwnerVale         RuleOwner = "vale"
	OwnerLanguageTool RuleOwner = "languagetool"
	OwnerNLP          RuleOwner = "nlp"
)

// FallbackPolicy determines whether a rule should still run when owner-adjacent signals exist.
type FallbackPolicy string

const (
	FallbackNever     FallbackPolicy = "never"
	FallbackWhenOwner FallbackPolicy = "when_owner_unavailable"
)

type RuleThreshold struct {
	NoteAt    int
	WarningAt int
	ErrorAt   int
}

type AppliesWhen struct {
	Domains           []string
	Specialties       []string
	AssignmentFormats []string
	MinWords          int
}

type SkipWhen struct {
	Domains           []string
	Specialties       []string
	AssignmentFormats []string
}

// RuleSpec is metadata-only for deterministic coverage planning and guardrails.
// This registry can be used to gate execution and enforce ownership over time.
type RuleSpec struct {
	ID             string
	Title          string
	Layer          CoverageLayer
	Owner          RuleOwner
	Category       string
	Purpose        string
	MetricKeys     []string
	AppliesWhen    AppliesWhen
	SkipWhen       SkipWhen
	Thresholds     map[string]RuleThreshold // keyed by domain or "default"
	FallbackPolicy FallbackPolicy
}

// CurrentHeuristicRuleSpecs documents the existing heuristic rule surface.
// It is intentionally metadata-only so the existing runtime behavior remains unchanged.
func CurrentHeuristicRuleSpecs() []RuleSpec {
	return []RuleSpec{
		{
			ID:         "heuristic.avg_sentence_length_high",
			Title:      "Average Sentence Length High",
			Layer:      LayerGlobal,
			Owner:      OwnerHeuristic,
			Category:   "clarity",
			Purpose:    "Detect sentence complexity that can reduce readability.",
			MetricKeys: []string{"avg_sentence_length"},
			AppliesWhen: AppliesWhen{
				Domains:  []string{DomainGeneral, DomainFiction, DomainFantasy, DomainTechnical, DomainAcademic, DomainProfessional, DomainThoughtLeadership, DomainMarketing},
				MinWords: 1,
			},
			Thresholds: map[string]RuleThreshold{
				"default": {WarningAt: 24},
			},
			FallbackPolicy: FallbackNever,
		},
		{
			ID:         "heuristic.avg_sentence_length_low",
			Title:      "Average Sentence Length Low",
			Layer:      LayerGlobal,
			Owner:      OwnerHeuristic,
			Category:   "rhythm",
			Purpose:    "Detect overly clipped sentence rhythm that can reduce flow.",
			MetricKeys: []string{"avg_sentence_length"},
			AppliesWhen: AppliesWhen{
				Domains:  []string{DomainGeneral, DomainFiction, DomainFantasy, DomainTechnical, DomainAcademic, DomainProfessional, DomainThoughtLeadership, DomainMarketing},
				MinWords: 1,
			},
			Thresholds: map[string]RuleThreshold{
				"default": {WarningAt: 8},
			},
			FallbackPolicy: FallbackNever,
		},
		{
			ID:         "heuristic.adverb_density",
			Title:      "Adverb Density",
			Layer:      LayerGlobal,
			Owner:      OwnerHeuristic,
			Category:   "prose precision",
			Purpose:    "Detect excess modifier usage that can weaken specificity.",
			MetricKeys: []string{"adverb_count", "word_count"},
			AppliesWhen: AppliesWhen{
				Domains:  []string{DomainGeneral, DomainFiction, DomainFantasy, DomainTechnical, DomainAcademic, DomainProfessional, DomainThoughtLeadership, DomainMarketing},
				MinWords: 1,
			},
			FallbackPolicy: FallbackNever,
		},
		{
			ID:         "heuristic.comparison_as_density",
			Title:      "As-Comparison Density",
			Layer:      LayerDomain,
			Owner:      OwnerHeuristic,
			Category:   "image freshness",
			Purpose:    "Detect overuse of comparative phrasing in narrative and idea-driven prose.",
			MetricKeys: []string{"comparison_as"},
			AppliesWhen: AppliesWhen{
				Domains:  []string{DomainGeneral, DomainFiction, DomainFantasy, DomainThoughtLeadership},
				MinWords: 1,
			},
			FallbackPolicy: FallbackNever,
		},
		{
			ID:         "heuristic.long_single_paragraph",
			Title:      "Long Single Paragraph",
			Layer:      LayerGlobal,
			Owner:      OwnerHeuristic,
			Category:   "structure",
			Purpose:    "Detect missing chunking when long content is written in one paragraph.",
			MetricKeys: []string{"paragraph_count", "word_count"},
			AppliesWhen: AppliesWhen{
				Domains:  []string{DomainGeneral, DomainFiction, DomainFantasy, DomainTechnical, DomainAcademic, DomainProfessional, DomainThoughtLeadership, DomainMarketing},
				MinWords: 250,
			},
			FallbackPolicy: FallbackNever,
		},
		{
			ID:         "heuristic.dialogue_absence_long_scene",
			Title:      "Dialogue Absence in Long Scene",
			Layer:      LayerDomain,
			Owner:      OwnerHeuristic,
			Category:   "dialogue intelligence",
			Purpose:    "Detect long narrative scenes without quoted dialogue when dialogue may be expected.",
			MetricKeys: []string{"dialogue_marks", "word_count"},
			AppliesWhen: AppliesWhen{
				Domains:  []string{DomainFiction, DomainFantasy},
				MinWords: 700,
			},
			SkipWhen: SkipWhen{
				Specialties: []string{"poetry"},
			},
			FallbackPolicy: FallbackNever,
		},
		{
			ID:         "heuristic.brief_draft_by_domain",
			Title:      "Brief Draft by Domain",
			Layer:      LayerGlobal,
			Owner:      OwnerHeuristic,
			Category:   "development",
			Purpose:    "Detect draft-length insufficiency for the selected writing goals.",
			MetricKeys: []string{"word_count"},
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainGeneral, DomainFiction, DomainFantasy, DomainTechnical, DomainAcademic, DomainProfessional, DomainThoughtLeadership, DomainMarketing},
			},
			FallbackPolicy: FallbackNever,
		},
	}
}

// CurrentDeterministicRuleSpecs is the shared metadata registry for deterministic
// analyzer ownership and applicability across heuristic + third-party analyzers.
func CurrentDeterministicRuleSpecs() []RuleSpec {
	specs := append([]RuleSpec{}, CurrentHeuristicRuleSpecs()...)
	specs = append(specs,
		RuleSpec{
			ID:             "nlp.global.clarity",
			Title:          "Global Clarity Ownership",
			Layer:          LayerGlobal,
			Owner:          OwnerNLP,
			Category:       "clarity",
			Purpose:        "Primary owner for discourse-level clarity findings.",
			FallbackPolicy: FallbackWhenOwner,
		},
		RuleSpec{
			ID:             "nlp.global.structure",
			Title:          "Global Structure Ownership",
			Layer:          LayerGlobal,
			Owner:          OwnerNLP,
			Category:       "structure",
			Purpose:        "Primary owner for discourse-level structural findings.",
			FallbackPolicy: FallbackWhenOwner,
		},
		RuleSpec{
			ID:             "nlp.global.readability",
			Title:          "Global Readability Ownership",
			Layer:          LayerGlobal,
			Owner:          OwnerNLP,
			Category:       "readability",
			Purpose:        "Primary owner for readability and flow findings.",
			FallbackPolicy: FallbackWhenOwner,
		},
		RuleSpec{
			ID:             "languagetool.global.mechanics",
			Title:          "Global Mechanics Ownership",
			Layer:          LayerGlobal,
			Owner:          OwnerLanguageTool,
			Category:       "mechanics",
			Purpose:        "Primary owner for grammar/spelling/punctuation findings.",
			FallbackPolicy: FallbackNever,
		},
		RuleSpec{
			ID:             "vale.global.style_policy",
			Title:          "Global Style Policy Ownership",
			Layer:          LayerGlobal,
			Owner:          OwnerVale,
			Category:       "style policy",
			Purpose:        "Primary owner for style-guide policy findings.",
			FallbackPolicy: FallbackNever,
		},
		RuleSpec{
			ID:       "nlp.domain.narrative_progression",
			Title:    "Domain Narrative Progression Ownership",
			Layer:    LayerDomain,
			Owner:    OwnerNLP,
			Category: "narrative progression",
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainFiction, DomainFantasy},
			},
			Purpose:        "Primary owner for fiction/fantasy progression findings.",
			FallbackPolicy: FallbackWhenOwner,
		},
		RuleSpec{
			ID:       "nlp.domain.instructional_completeness",
			Title:    "Domain Instructional Completeness Ownership",
			Layer:    LayerDomain,
			Owner:    OwnerNLP,
			Category: "instructional completeness",
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainTechnical},
			},
			Purpose:        "Primary owner for technical instructional completeness findings.",
			FallbackPolicy: FallbackWhenOwner,
		},
		RuleSpec{
			ID:       "nlp.domain.argument_support",
			Title:    "Domain Argument Support Ownership",
			Layer:    LayerDomain,
			Owner:    OwnerNLP,
			Category: "argument support",
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainAcademic, DomainThoughtLeadership},
			},
			Purpose:        "Primary owner for claim/evidence support findings.",
			FallbackPolicy: FallbackWhenOwner,
		},
		RuleSpec{
			ID:       "heuristic.domain.actionability",
			Title:    "Domain Actionability Ownership",
			Layer:    LayerDomain,
			Owner:    OwnerHeuristic,
			Category: "actionability",
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainProfessional},
			},
			Purpose:        "Primary owner for professional writing actionability findings.",
			FallbackPolicy: FallbackNever,
		},
		RuleSpec{
			ID:       "nlp.domain.message_hierarchy",
			Title:    "Domain Message Hierarchy Ownership",
			Layer:    LayerDomain,
			Owner:    OwnerNLP,
			Category: "message hierarchy",
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainMarketing},
			},
			Purpose:        "Primary owner for message hierarchy findings in marketing contexts.",
			FallbackPolicy: FallbackWhenOwner,
		},
		RuleSpec{
			ID:       "heuristic.specialty.memo_execution",
			Title:    "Specialty Memo Execution Ownership",
			Layer:    LayerSpecialty,
			Owner:    OwnerHeuristic,
			Category: "memo execution",
			AppliesWhen: AppliesWhen{
				Specialties: []string{"memo"},
			},
			Purpose:        "Primary owner for memo-specific execution findings.",
			FallbackPolicy: FallbackNever,
		},
		RuleSpec{
			ID:       "heuristic.specialty.cta_architecture",
			Title:    "Specialty CTA Architecture Ownership",
			Layer:    LayerSpecialty,
			Owner:    OwnerHeuristic,
			Category: "cta architecture",
			AppliesWhen: AppliesWhen{
				Specialties: []string{"landing_page"},
			},
			Purpose:        "Primary owner for landing-page CTA architecture findings.",
			FallbackPolicy: FallbackNever,
		},
		RuleSpec{
			ID:       "heuristic.specialty.grant_compliance_framing",
			Title:    "Specialty Grant Compliance Framing Ownership",
			Layer:    LayerSpecialty,
			Owner:    OwnerHeuristic,
			Category: "grant compliance framing",
			AppliesWhen: AppliesWhen{
				Specialties: []string{"grant"},
			},
			Purpose:        "Primary owner for grant framing findings.",
			FallbackPolicy: FallbackNever,
		},
		RuleSpec{
			ID:       "heuristic.specialty.poetic_craft_proxies",
			Title:    "Specialty Poetic Craft Proxy Ownership",
			Layer:    LayerSpecialty,
			Owner:    OwnerHeuristic,
			Category: "poetic craft proxies",
			AppliesWhen: AppliesWhen{
				Specialties: []string{"poetry"},
			},
			Purpose:        "Primary owner for poetry specialty deterministic proxies.",
			FallbackPolicy: FallbackNever,
		},
		RuleSpec{
			ID:       "heuristic.specialty.structure_poetry_override",
			Title:    "Specialty Structure Ownership Override (Poetry)",
			Layer:    LayerSpecialty,
			Owner:    OwnerHeuristic,
			Category: "structure",
			AppliesWhen: AppliesWhen{
				Specialties: []string{"poetry"},
			},
			Purpose:        "Poetry specialty overrides global structure ownership.",
			FallbackPolicy: FallbackNever,
		},
	)
	return specs
}

func heuristicRuleIndex() map[string]RuleSpec {
	index := make(map[string]RuleSpec)
	for _, spec := range CurrentHeuristicRuleSpecs() {
		index[spec.ID] = spec
	}
	return index
}

var currentHeuristicRuleIndex = heuristicRuleIndex()

func shouldEvaluateHeuristicRule(ruleID string, options ContextOptions, wordCount int) bool {
	spec, ok := currentHeuristicRuleIndex[ruleID]
	if !ok {
		return true
	}

	domain := DomainForContext(options)
	specialties := contextSpecialties(options)
	assignmentFormat := normalizedValue(options.AssignmentFormat)

	if !matchAllowedDomain(spec.AppliesWhen.Domains, domain) {
		return false
	}
	if spec.AppliesWhen.MinWords > 0 && wordCount < spec.AppliesWhen.MinWords {
		return false
	}
	if len(spec.AppliesWhen.Specialties) > 0 && !hasIntersection(specialties, normalizeSlice(spec.AppliesWhen.Specialties)) {
		return false
	}
	if len(spec.AppliesWhen.AssignmentFormats) > 0 && !containsString(normalizeSlice(spec.AppliesWhen.AssignmentFormats), assignmentFormat) {
		return false
	}

	if containsString(normalizeSlice(spec.SkipWhen.Domains), domain) {
		return false
	}
	if hasIntersection(specialties, normalizeSlice(spec.SkipWhen.Specialties)) {
		return false
	}
	if containsString(normalizeSlice(spec.SkipWhen.AssignmentFormats), assignmentFormat) {
		return false
	}
	return true
}

func contextSpecialties(options ContextOptions) []string {
	combined := normalizedValue(strings.Join([]string{
		options.TemplateKey,
		options.TreeSlug,
		options.WritingType,
		options.AssignmentFormat,
	}, " "))

	out := make([]string, 0, 8)
	if containsAny([]string{combined}, "poetry", "poem", "haiku", "sonnet") {
		out = append(out, "poetry")
	}
	if containsAny([]string{combined}, "memo") {
		out = append(out, "memo")
	}
	if containsAny([]string{combined}, "landing page") {
		out = append(out, "landing_page")
	}
	if containsAny([]string{combined}, "grant") {
		out = append(out, "grant")
	}
	if containsAny([]string{combined}, "tutorial", "how-to", "guide") {
		out = append(out, "tutorial")
	}
	if containsAny([]string{combined}, "scene") {
		out = append(out, "scene")
	}
	return out
}

func matchAllowedDomain(allowed []string, domain string) bool {
	if len(allowed) == 0 {
		return true
	}
	return containsString(normalizeSlice(allowed), domain)
}

func normalizeSlice(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if normalized := normalizedValue(value); normalized != "" {
			out = append(out, normalized)
		}
	}
	return out
}

func normalizedValue(value string) string {
	return strings.ToLower(strings.TrimSpace(value))
}

func containsString(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}

func hasIntersection(left, right []string) bool {
	if len(left) == 0 || len(right) == 0 {
		return false
	}
	rightSet := make(map[string]struct{}, len(right))
	for _, value := range right {
		rightSet[value] = struct{}{}
	}
	for _, value := range left {
		if _, ok := rightSet[value]; ok {
			return true
		}
	}
	return false
}
