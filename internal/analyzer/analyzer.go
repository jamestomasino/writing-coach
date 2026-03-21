package analyzer

import (
	"context"

	"github.com/tomasino/writing-coach/internal/domain"
)

type Finding struct {
	Analyzer string
	Category string
	Severity string
	Message  string
}

type Report struct {
	Findings []Finding
	Metrics  map[string]int
	Warnings []string
}

type ContextOptions struct {
	TreeSlug         string
	WritingType      string
	AssignmentFormat string
	TemplateKey      string
}

type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, text string) (Report, error)
}

type ContextualAnalyzer interface {
	Analyzer
	AnalyzeWithContext(ctx context.Context, text string, options ContextOptions) (Report, error)
}

func ContextFromProfile(treeSlug string, profile domain.OnboardingProfile) ContextOptions {
	return ContextOptions{
		TreeSlug:         treeSlug,
		WritingType:      profile.WritingType,
		AssignmentFormat: profile.AssignmentFormat,
		TemplateKey:      profile.TemplateKey,
	}
}

func Merge(reports ...Report) Report {
	merged := Report{
		Metrics: make(map[string]int),
	}

	for _, report := range reports {
		merged.Findings = append(merged.Findings, report.Findings...)
		merged.Warnings = append(merged.Warnings, report.Warnings...)
		for key, value := range report.Metrics {
			merged.Metrics[key] = value
		}
	}

	return merged
}
