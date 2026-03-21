package analyzer

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Vale struct {
	Binary        string
	ConfigPath    string
	WorkingDir    string
	StylesPath    string
	MinAlertLevel string
	BaseStyles    []string
}

func (v Vale) Name() string {
	return "vale"
}

func (v Vale) Analyze(ctx context.Context, text string) (Report, error) {
	return v.AnalyzeWithContext(ctx, text, ContextOptions{})
}

func (v Vale) AnalyzeWithContext(ctx context.Context, text string, options ContextOptions) (Report, error) {
	if v.Binary == "" {
		return Report{Warnings: []string{"vale not configured"}}, nil
	}
	if !deterministicLanguageSupported(options.WritingLanguage) {
		return Report{Warnings: []string{unsupportedLanguageWarning(v.Name(), options.WritingLanguage)}}, nil
	}

	tempDir := ""
	if v.WorkingDir != "" {
		tempDir = v.WorkingDir
	}
	file, err := os.CreateTemp(tempDir, "writing-coach-vale-*.md")
	if err != nil {
		return Report{}, err
	}
	defer os.Remove(file.Name())

	if _, err := file.WriteString(text); err != nil {
		file.Close()
		return Report{}, err
	}
	if err := file.Close(); err != nil {
		return Report{}, err
	}

	args := []string{"--output=JSON", "--no-exit"}
	configPath := v.ConfigPath
	tempConfigPath := ""
	if dynamicConfig, ok := v.dynamicConfig(options); ok {
		configFile, err := os.CreateTemp(tempDir, "writing-coach-vale-*.ini")
		if err != nil {
			return Report{}, err
		}
		if _, err := configFile.WriteString(dynamicConfig); err != nil {
			configFile.Close()
			_ = os.Remove(configFile.Name())
			return Report{}, err
		}
		if err := configFile.Close(); err != nil {
			_ = os.Remove(configFile.Name())
			return Report{}, err
		}
		tempConfigPath = configFile.Name()
		defer os.Remove(tempConfigPath)
		configPath = tempConfigPath
	}
	if configPath != "" {
		args = append(args, "--config", configPath)
	}
	target := file.Name()
	if v.WorkingDir != "" {
		if rel, err := filepath.Rel(v.WorkingDir, file.Name()); err == nil {
			target = rel
		}
	}
	args = append(args, target)
	cmd := exec.CommandContext(ctx, v.Binary, args...)
	if v.WorkingDir != "" {
		cmd.Dir = v.WorkingDir
	}
	output, err := cmd.Output()
	if err != nil {
		return Report{}, err
	}

	var raw map[string][]struct {
		Check    string `json:"Check"`
		Message  string `json:"Message"`
		Severity string `json:"Severity"`
	}
	if err := json.Unmarshal(output, &raw); err != nil {
		return Report{}, err
	}

	report := Report{Metrics: map[string]int{}}
	for _, entries := range raw {
		for _, entry := range entries {
			report.Findings = append(report.Findings, Finding{
				Analyzer: "vale",
				Category: entry.Check,
				Severity: normalizeSeverity(entry.Severity),
				Message:  entry.Message,
			})
		}
	}
	return report, nil
}

func (v Vale) dynamicConfig(options ContextOptions) (string, bool) {
	stylesPath := strings.TrimSpace(v.StylesPath)
	if stylesPath == "" {
		return "", false
	}
	styles := append([]string{}, v.BaseStyles...)
	styles = append(styles, valeStylesForContext(options)...)
	styles = uniqueStrings(styles)
	if len(styles) == 0 {
		return "", false
	}
	minAlertLevel := strings.TrimSpace(v.MinAlertLevel)
	if minAlertLevel == "" {
		minAlertLevel = "suggestion"
	}
	var builder strings.Builder
	builder.WriteString("StylesPath = ")
	builder.WriteString(stylesPath)
	builder.WriteString("\n")
	builder.WriteString("MinAlertLevel = ")
	builder.WriteString(minAlertLevel)
	builder.WriteString("\n\n[*.{md,txt}]\nBasedOnStyles = ")
	builder.WriteString(strings.Join(styles, ", "))
	builder.WriteString("\n\n[*.go]\nBasedOnStyles =\n")
	return builder.String(), true
}

func valeStylesForContext(options ContextOptions) []string {
	var styles []string
	switch DomainForContext(options) {
	case DomainFantasy:
		styles = append(styles, "WritingCoachFiction", "WritingCoachFantasy")
	case DomainFiction:
		styles = append(styles, "WritingCoachFiction")
	case DomainTechnical:
		styles = append(styles, "WritingCoachTechnical")
	case DomainAcademic:
		styles = append(styles, "WritingCoachAcademic")
	case DomainProfessional:
		styles = append(styles, "WritingCoachProfessional")
	case DomainThoughtLeadership:
		styles = append(styles, "WritingCoachThoughtLeadership")
	case DomainMarketing:
		styles = append(styles, "WritingCoachMarketing")
	}
	return styles
}

func containsAny(values []string, needles ...string) bool {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		for _, needle := range needles {
			if strings.Contains(value, needle) {
				return true
			}
		}
	}
	return false
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		out = append(out, value)
	}
	return out
}
