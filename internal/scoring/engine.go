package scoring

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/domain"
)

//go:embed rubrics/*.json
var rubricFS embed.FS

const deterministicScoreSource = "deterministic"

var (
	loadRubricsOnce sync.Once
	loadedRubrics   map[string]Rubric
	loadErr         error
)

type Engine struct {
	rubrics map[string]Rubric
}

type Rubric struct {
	ID                 string                   `json:"id"`
	Version            string                   `json:"version"`
	Domain             string                   `json:"domain"`
	DefaultSkills      []string                 `json:"default_skills"`
	DefaultSkillConfig SkillConfig              `json:"default_skill_config"`
	Skills             map[string]SkillConfig   `json:"skills"`
	TrackOverrides     map[string]TrackOverride `json:"track_overrides"`
}

type TrackOverride struct {
	AddSkills        []string       `json:"add_skills"`
	RemoveSkills     []string       `json:"remove_skills"`
	SkillAdjustments map[string]int `json:"skill_adjustments"`
}

type SkillConfig struct {
	BaseScore         int            `json:"base_score"`
	FindingPenalty    int            `json:"finding_penalty"`
	MaxFindingPenalty int            `json:"max_finding_penalty"`
	NoFindingBonus    int            `json:"no_finding_bonus"`
	RangeRules        []RangeRule    `json:"range_rules"`
	CategoryPenalties map[string]int `json:"category_penalties"`
}

type RangeRule struct {
	Metric string `json:"metric"`
	Min    *int   `json:"min"`
	Max    *int   `json:"max"`
	Delta  int    `json:"delta"`
	Reason string `json:"reason"`
}

type ScoreEvidence struct {
	RubricID          string         `json:"rubric_id"`
	ScoreVersion      string         `json:"score_version"`
	Domain            string         `json:"domain"`
	TreeSlug          string         `json:"tree_slug"`
	Skill             string         `json:"skill"`
	BaseScore         int            `json:"base_score"`
	FinalScore        int            `json:"final_score"`
	FindingCount      int            `json:"finding_count"`
	MetricSnapshot    map[string]int `json:"metric_snapshot"`
	CategoryHistogram map[string]int `json:"category_histogram"`
	AppliedRules      []string       `json:"applied_rules"`
}

func NewEngine() (Engine, error) {
	loadRubricsOnce.Do(func() {
		loadedRubrics, loadErr = loadEmbeddedRubrics()
	})
	if loadErr != nil {
		return Engine{}, loadErr
	}
	return Engine{rubrics: loadedRubrics}, nil
}

func (e Engine) ScoreSubmission(sub domain.Submission, report analyzer.Report, options analyzer.ContextOptions, activeTGOs []domain.TGO) ([]domain.SkillScore, error) {
	domainName := analyzer.DomainForContext(options)
	rubric, ok := e.rubrics[domainName]
	if !ok {
		rubric = e.rubrics[analyzer.DomainGeneral]
	}
	if rubric.ID == "" {
		return nil, fmt.Errorf("no rubric available for domain %q", domainName)
	}

	candidateSkills := candidateSkills(rubric, options.TreeSlug, activeTGOs)
	if len(candidateSkills) == 0 {
		candidateSkills = append(candidateSkills, "clarity and coherence", "sentence economy")
	}

	categoryHistogram := findingHistogram(report)
	scores := make([]domain.SkillScore, 0, len(candidateSkills))
	for _, skill := range candidateSkills {
		config, ok := rubric.Skills[skill]
		if !ok {
			config = rubric.DefaultSkillConfig
		}
		score, evidence := scoreSkill(skill, config, rubric, options, report, categoryHistogram)
		evidenceJSON, err := json.Marshal(evidence)
		if err != nil {
			return nil, err
		}
		scores = append(scores, domain.SkillScore{
			SubmissionID:      sub.ID,
			Skill:             skill,
			Score:             score,
			ScoreSource:       deterministicScoreSource,
			ScoreVersion:      rubric.Version,
			ScoreEvidenceJSON: string(evidenceJSON),
		})
	}

	if override, ok := rubric.TrackOverrides[strings.ToLower(strings.TrimSpace(options.TreeSlug))]; ok {
		for i := range scores {
			if delta, has := override.SkillAdjustments[scores[i].Skill]; has {
				adjusted := clampScore(scores[i].Score + delta)
				scores[i].Score = adjusted
				if strings.TrimSpace(scores[i].ScoreEvidenceJSON) == "" || scores[i].ScoreEvidenceJSON == "{}" {
					continue
				}
				var evidence ScoreEvidence
				if err := json.Unmarshal([]byte(scores[i].ScoreEvidenceJSON), &evidence); err == nil {
					evidence.FinalScore = adjusted
					evidence.AppliedRules = append(evidence.AppliedRules, fmt.Sprintf("track override %s: %+d", options.TreeSlug, delta))
					if encoded, err := json.Marshal(evidence); err == nil {
						scores[i].ScoreEvidenceJSON = string(encoded)
					}
				}
			}
		}
	}

	sort.Slice(scores, func(i, j int) bool {
		if scores[i].Score == scores[j].Score {
			return scores[i].Skill < scores[j].Skill
		}
		return scores[i].Score > scores[j].Score
	})

	return scores, nil
}

func loadEmbeddedRubrics() (map[string]Rubric, error) {
	entries, err := rubricFS.ReadDir("rubrics")
	if err != nil {
		return nil, err
	}
	result := make(map[string]Rubric, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		content, err := rubricFS.ReadFile("rubrics/" + entry.Name())
		if err != nil {
			return nil, err
		}
		var rubric Rubric
		if err := json.Unmarshal(content, &rubric); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		domainName := strings.TrimSpace(rubric.Domain)
		if domainName == "" {
			return nil, fmt.Errorf("rubric %s missing domain", entry.Name())
		}
		if strings.TrimSpace(rubric.Version) == "" {
			rubric.Version = "det-v1"
		}
		if rubric.Skills == nil {
			rubric.Skills = map[string]SkillConfig{}
		}
		if rubric.TrackOverrides == nil {
			rubric.TrackOverrides = map[string]TrackOverride{}
		}
		result[domainName] = rubric
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("no rubrics loaded")
	}
	return result, nil
}

func candidateSkills(rubric Rubric, treeSlug string, activeTGOs []domain.TGO) []string {
	seen := map[string]bool{}
	add := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			return
		}
		seen[value] = true
	}
	for _, skill := range rubric.DefaultSkills {
		add(skill)
	}
	for _, tgo := range activeTGOs {
		if skill := strings.TrimSpace(domain.TGOCodeToSkill[tgo.Code]); skill != "" {
			add(skill)
		}
	}
	trackKey := strings.ToLower(strings.TrimSpace(treeSlug))
	if override, ok := rubric.TrackOverrides[trackKey]; ok {
		for _, skill := range override.AddSkills {
			add(skill)
		}
		if len(override.RemoveSkills) > 0 {
			remove := map[string]bool{}
			for _, skill := range override.RemoveSkills {
				remove[strings.TrimSpace(skill)] = true
			}
			for skill := range seen {
				if remove[skill] {
					delete(seen, skill)
				}
			}
		}
	}
	if len(seen) == 0 {
		return nil
	}
	out := make([]string, 0, len(seen))
	for skill := range seen {
		out = append(out, skill)
	}
	sort.Strings(out)
	if len(out) > 8 {
		out = out[:8]
	}
	return out
}

func scoreSkill(skill string, config SkillConfig, rubric Rubric, options analyzer.ContextOptions, report analyzer.Report, categoryHistogram map[string]int) (int, ScoreEvidence) {
	score := clampScore(config.BaseScore)
	evidence := ScoreEvidence{
		RubricID:          rubric.ID,
		ScoreVersion:      rubric.Version,
		Domain:            analyzer.DomainForContext(options),
		TreeSlug:          strings.TrimSpace(options.TreeSlug),
		Skill:             skill,
		BaseScore:         score,
		FinalScore:        score,
		FindingCount:      len(report.Findings),
		MetricSnapshot:    map[string]int{},
		CategoryHistogram: categoryHistogram,
		AppliedRules:      []string{},
	}

	for _, rule := range config.RangeRules {
		metric := strings.TrimSpace(rule.Metric)
		if metric == "" {
			continue
		}
		value, ok := report.Metrics[metric]
		if !ok {
			continue
		}
		evidence.MetricSnapshot[metric] = value
		if rule.Min != nil && value < *rule.Min {
			continue
		}
		if rule.Max != nil && value > *rule.Max {
			continue
		}
		score = clampScore(score + rule.Delta)
		if rule.Reason != "" {
			evidence.AppliedRules = append(evidence.AppliedRules, rule.Reason)
		}
	}

	if len(report.Findings) == 0 && config.NoFindingBonus > 0 {
		score = clampScore(score + config.NoFindingBonus)
		evidence.AppliedRules = append(evidence.AppliedRules, "no findings bonus")
	}
	if config.FindingPenalty > 0 && len(report.Findings) > 0 {
		penalty := len(report.Findings) / config.FindingPenalty
		if penalty > config.MaxFindingPenalty {
			penalty = config.MaxFindingPenalty
		}
		if penalty > 0 {
			score = clampScore(score - penalty)
			evidence.AppliedRules = append(evidence.AppliedRules, fmt.Sprintf("finding count penalty: -%d", penalty))
		}
	}

	for categoryKey, penalty := range config.CategoryPenalties {
		if penalty <= 0 {
			continue
		}
		hits := categoryHistogram[strings.ToLower(strings.TrimSpace(categoryKey))]
		if hits == 0 {
			continue
		}
		totalPenalty := hits * penalty
		score = clampScore(score - totalPenalty)
		evidence.AppliedRules = append(evidence.AppliedRules, fmt.Sprintf("category penalty %s: -%d", categoryKey, totalPenalty))
	}

	evidence.FinalScore = score
	return score, evidence
}

func findingHistogram(report analyzer.Report) map[string]int {
	hist := map[string]int{}
	for _, finding := range report.Findings {
		category := strings.ToLower(strings.TrimSpace(finding.Category))
		if category == "" {
			continue
		}
		hist[category]++
	}
	return hist
}

func clampScore(value int) int {
	if value < 1 {
		return 1
	}
	if value > 5 {
		return 5
	}
	return value
}
