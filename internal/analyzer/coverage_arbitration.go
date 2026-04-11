package analyzer

import "strings"

type CategoryOwnershipSpec struct {
	Category       string
	Layer          CoverageLayer
	Owner          RuleOwner
	FallbackPolicy FallbackPolicy
	AppliesWhen    AppliesWhen
}

type analyzerReportResult struct {
	analyzer string
	report   Report
}

func DeterministicCategoryOwnershipSpecs() []CategoryOwnershipSpec {
	return []CategoryOwnershipSpec{
		{
			Category:       "clarity",
			Layer:          LayerGlobal,
			Owner:          OwnerNLP,
			FallbackPolicy: FallbackWhenOwner,
		},
		{
			Category:       "structure",
			Layer:          LayerGlobal,
			Owner:          OwnerNLP,
			FallbackPolicy: FallbackWhenOwner,
		},
		{
			Category:       "readability",
			Layer:          LayerGlobal,
			Owner:          OwnerNLP,
			FallbackPolicy: FallbackWhenOwner,
		},
		{
			Category:       "mechanics",
			Layer:          LayerGlobal,
			Owner:          OwnerLanguageTool,
			FallbackPolicy: FallbackNever,
		},
		{
			Category:       "style policy",
			Layer:          LayerGlobal,
			Owner:          OwnerVale,
			FallbackPolicy: FallbackNever,
		},
		{
			Category:       "narrative progression",
			Layer:          LayerDomain,
			Owner:          OwnerNLP,
			FallbackPolicy: FallbackWhenOwner,
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainFiction, DomainFantasy},
			},
		},
		{
			Category:       "instructional completeness",
			Layer:          LayerDomain,
			Owner:          OwnerNLP,
			FallbackPolicy: FallbackWhenOwner,
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainTechnical},
			},
		},
		{
			Category:       "argument support",
			Layer:          LayerDomain,
			Owner:          OwnerNLP,
			FallbackPolicy: FallbackWhenOwner,
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainAcademic, DomainThoughtLeadership},
			},
		},
		{
			Category:       "actionability",
			Layer:          LayerDomain,
			Owner:          OwnerHeuristic,
			FallbackPolicy: FallbackNever,
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainProfessional},
			},
		},
		{
			Category:       "message hierarchy",
			Layer:          LayerDomain,
			Owner:          OwnerNLP,
			FallbackPolicy: FallbackWhenOwner,
			AppliesWhen: AppliesWhen{
				Domains: []string{DomainMarketing},
			},
		},
		{
			Category:       "memo execution",
			Layer:          LayerSpecialty,
			Owner:          OwnerHeuristic,
			FallbackPolicy: FallbackNever,
			AppliesWhen: AppliesWhen{
				Specialties: []string{"memo"},
			},
		},
		{
			Category:       "cta architecture",
			Layer:          LayerSpecialty,
			Owner:          OwnerHeuristic,
			FallbackPolicy: FallbackNever,
			AppliesWhen: AppliesWhen{
				Specialties: []string{"landing_page"},
			},
		},
		{
			Category:       "grant compliance framing",
			Layer:          LayerSpecialty,
			Owner:          OwnerHeuristic,
			FallbackPolicy: FallbackNever,
			AppliesWhen: AppliesWhen{
				Specialties: []string{"grant"},
			},
		},
		{
			Category:       "poetic craft proxies",
			Layer:          LayerSpecialty,
			Owner:          OwnerHeuristic,
			FallbackPolicy: FallbackNever,
			AppliesWhen: AppliesWhen{
				Specialties: []string{"poetry"},
			},
		},
		// Specialty override of global structure ownership for poetry-focused work.
		{
			Category:       "structure",
			Layer:          LayerSpecialty,
			Owner:          OwnerHeuristic,
			FallbackPolicy: FallbackNever,
			AppliesWhen: AppliesWhen{
				Specialties: []string{"poetry"},
			},
		},
	}
}

func mergeWithCoverageArbitration(options ContextOptions, reports ...analyzerReportResult) Report {
	merged := Report{Metrics: map[string]int{}}
	findings := make([]Finding, 0, 32)
	for _, result := range reports {
		merged.Warnings = append(merged.Warnings, result.report.Warnings...)
		for key, value := range result.report.Metrics {
			merged.Metrics[key] = value
		}
		for _, finding := range result.report.Findings {
			normalized := finding
			if strings.TrimSpace(normalized.Analyzer) == "" {
				normalized.Analyzer = result.analyzer
			}
			findings = append(findings, normalized)
		}
	}
	merged.Findings = arbitrateDeterministicFindings(findings, options)
	return merged
}

func arbitrateDeterministicFindings(findings []Finding, options ContextOptions) []Finding {
	if len(findings) == 0 {
		return nil
	}

	type categorizedFinding struct {
		finding          Finding
		category         string
		spec             *CategoryOwnershipSpec
		analyzerOwner    RuleOwner
		normalizedLayer  CoverageLayer
		keptByOwnerMatch bool
	}

	specs := DeterministicCategoryOwnershipSpecs()
	categorized := make([]categorizedFinding, 0, len(findings))
	for _, finding := range findings {
		category := canonicalCategoryForFinding(finding)
		spec := resolveOwnershipSpec(specs, category, options)
		layer := LayerGlobal
		if spec != nil {
			layer = spec.Layer
		}
		categorized = append(categorized, categorizedFinding{
			finding:         finding,
			category:        category,
			spec:            spec,
			analyzerOwner:   ownerForAnalyzer(finding.Analyzer),
			normalizedLayer: layer,
		})
	}

	grouped := make(map[string][]int)
	for idx, item := range categorized {
		if item.spec == nil {
			continue
		}
		key := string(item.normalizedLayer) + "|" + item.category
		grouped[key] = append(grouped[key], idx)
	}

	keep := make([]bool, len(categorized))
	for idx := range keep {
		keep[idx] = true
	}

	for _, indexes := range grouped {
		spec := categorized[indexes[0]].spec
		if spec == nil {
			continue
		}
		ownerPresent := false
		for _, idx := range indexes {
			if categorized[idx].analyzerOwner == spec.Owner {
				ownerPresent = true
				categorized[idx].keptByOwnerMatch = true
			}
		}
		if ownerPresent {
			for _, idx := range indexes {
				keep[idx] = categorized[idx].analyzerOwner == spec.Owner
			}
			continue
		}
		if spec.FallbackPolicy == FallbackWhenOwner {
			for _, idx := range indexes {
				keep[idx] = categorized[idx].analyzerOwner == OwnerHeuristic
			}
		}
	}

	out := make([]Finding, 0, len(findings))
	for idx, item := range categorized {
		if !keep[idx] {
			continue
		}
		out = append(out, item.finding)
	}
	return out
}

func ownerForAnalyzer(analyzerName string) RuleOwner {
	switch normalizedValue(analyzerName) {
	case string(OwnerHeuristic):
		return OwnerHeuristic
	case string(OwnerVale):
		return OwnerVale
	case string(OwnerLanguageTool):
		return OwnerLanguageTool
	case string(OwnerNLP):
		return OwnerNLP
	default:
		return ""
	}
}

func canonicalCategoryForFinding(finding Finding) string {
	owner := ownerForAnalyzer(finding.Analyzer)
	switch owner {
	case OwnerVale:
		return "style policy"
	case OwnerLanguageTool:
		return "mechanics"
	default:
		category := normalizedValue(finding.Category)
		if category == "" {
			return "uncategorized"
		}
		return category
	}
}

func resolveOwnershipSpec(specs []CategoryOwnershipSpec, category string, options ContextOptions) *CategoryOwnershipSpec {
	domain := DomainForContext(options)
	specialties := contextSpecialties(options)
	assignmentFormat := normalizedValue(options.AssignmentFormat)

	var best *CategoryOwnershipSpec
	bestRank := -1
	for idx := range specs {
		spec := &specs[idx]
		if normalizedValue(spec.Category) != category {
			continue
		}
		if !matchAllowedDomain(spec.AppliesWhen.Domains, domain) {
			continue
		}
		if len(spec.AppliesWhen.Specialties) > 0 && !hasIntersection(specialties, normalizeSlice(spec.AppliesWhen.Specialties)) {
			continue
		}
		if len(spec.AppliesWhen.AssignmentFormats) > 0 && !containsString(normalizeSlice(spec.AppliesWhen.AssignmentFormats), assignmentFormat) {
			continue
		}
		rank := coverageLayerRank(spec.Layer)
		if rank > bestRank {
			best = spec
			bestRank = rank
		}
	}
	return best
}

func coverageLayerRank(layer CoverageLayer) int {
	switch layer {
	case LayerSpecialty:
		return 3
	case LayerDomain:
		return 2
	default:
		return 1
	}
}
