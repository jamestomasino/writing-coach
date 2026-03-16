package analyzer

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
)

type Vale struct {
	Binary     string
	ConfigPath string
	WorkingDir string
}

func (v Vale) Name() string {
	return "vale"
}

func (v Vale) Analyze(ctx context.Context, text string) (Report, error) {
	if v.Binary == "" {
		return Report{Warnings: []string{"vale not configured"}}, nil
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
	if v.ConfigPath != "" {
		args = append(args, "--config", v.ConfigPath)
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
