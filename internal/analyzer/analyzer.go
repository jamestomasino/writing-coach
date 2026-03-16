package analyzer

import "context"

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

type Analyzer interface {
	Name() string
	Analyze(ctx context.Context, text string) (Report, error)
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
