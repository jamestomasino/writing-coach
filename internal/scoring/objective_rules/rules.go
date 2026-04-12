package objective_rules

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"

	"github.com/tomasino/writing-coach/internal/analyzer"
)

//go:embed *.json
var manifestFS embed.FS

type MetricRule struct {
	ID     string `json:"id"`
	Metric string `json:"metric"`
	Min    *int   `json:"min"`
	Max    *int   `json:"max"`
	Delta  int    `json:"delta"`
	Reason string `json:"reason"`
}

type CategoryRule struct {
	ID       string `json:"id"`
	Category string `json:"category"`
	MinHits  *int   `json:"min_hits"`
	MaxHits  *int   `json:"max_hits"`
	Delta    int    `json:"delta"`
	Reason   string `json:"reason"`
}

type TopScoreGate struct {
	RequireRuleTriggerCount int            `json:"require_rule_trigger_count"`
	MinMetrics              map[string]int `json:"min_metrics"`
	MaxMetrics              map[string]int `json:"max_metrics"`
}

type RuleSet struct {
	RuleID         string         `json:"rule_id"`
	Domain         string         `json:"domain"`
	TrackSlugs     []string       `json:"track_slugs"`
	TGOCodes       []string       `json:"tgo_codes"`
	MetricRules    []MetricRule   `json:"metric_rules"`
	CategoryRules  []CategoryRule `json:"category_rules"`
	TopScoreGate   TopScoreGate   `json:"top_score_gate"`
	EvidenceLabels []string       `json:"evidence_labels"`
}

type Manifest struct {
	Version  string    `json:"version"`
	RuleSets []RuleSet `json:"rule_sets"`
}

var (
	loadOnce sync.Once
	loadErr  error
	index    map[string][]RuleSet
)

func Resolve(tgoCode string, options analyzer.ContextOptions) (RuleSet, bool) {
	sets := ResolveAll(tgoCode, options)
	if len(sets) == 0 {
		return RuleSet{}, false
	}
	return sets[0], true
}

func ResolveAll(tgoCode string, options analyzer.ContextOptions) []RuleSet {
	if err := ensureLoaded(); err != nil {
		return nil
	}
	code := strings.TrimSpace(strings.ToLower(tgoCode))
	if code == "" {
		return nil
	}
	candidates := index[code]
	if len(candidates) == 0 {
		return nil
	}
	domainName := strings.TrimSpace(strings.ToLower(analyzer.DomainForContext(options)))
	track := strings.TrimSpace(strings.ToLower(options.TreeSlug))
	var matched []RuleSet
	for _, set := range candidates {
		if !ruleSetMatches(set, domainName, track) {
			continue
		}
		matched = append(matched, set)
	}
	sort.SliceStable(matched, func(i, j int) bool {
		iSpecific := len(matched[i].TGOCodes) == 1
		jSpecific := len(matched[j].TGOCodes) == 1
		if iSpecific != jSpecific {
			return iSpecific
		}
		return matched[i].RuleID < matched[j].RuleID
	})
	return matched
}

func HasAnyForCodeDomain(tgoCode, domainName string) bool {
	if err := ensureLoaded(); err != nil {
		return false
	}
	code := strings.TrimSpace(strings.ToLower(tgoCode))
	domainName = strings.TrimSpace(strings.ToLower(domainName))
	if code == "" || domainName == "" {
		return false
	}
	for _, set := range index[code] {
		if strings.TrimSpace(strings.ToLower(set.Domain)) == domainName {
			return true
		}
	}
	return false
}

func ensureLoaded() error {
	loadOnce.Do(func() {
		loaded, err := loadManifests()
		if err != nil {
			loadErr = err
			return
		}
		index = loaded
	})
	return loadErr
}

func ruleSetMatches(set RuleSet, domainName, track string) bool {
	if strings.TrimSpace(strings.ToLower(set.Domain)) != domainName {
		return false
	}
	if len(set.TrackSlugs) == 0 {
		return true
	}
	for _, slug := range set.TrackSlugs {
		if strings.TrimSpace(strings.ToLower(slug)) == track {
			return true
		}
	}
	return false
}

func loadManifests() (map[string][]RuleSet, error) {
	entries, err := manifestFS.ReadDir(".")
	if err != nil {
		return nil, err
	}
	result := map[string][]RuleSet{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		raw, err := manifestFS.ReadFile(entry.Name())
		if err != nil {
			return nil, err
		}
		var manifest Manifest
		if err := json.Unmarshal(raw, &manifest); err != nil {
			return nil, fmt.Errorf("parse %s: %w", entry.Name(), err)
		}
		for _, set := range manifest.RuleSets {
			if strings.TrimSpace(set.RuleID) == "" {
				return nil, fmt.Errorf("manifest %s contains ruleset with empty rule_id", entry.Name())
			}
			if strings.TrimSpace(set.Domain) == "" {
				return nil, fmt.Errorf("manifest %s ruleset %s missing domain", entry.Name(), set.RuleID)
			}
			for _, code := range set.TGOCodes {
				normalized := strings.TrimSpace(strings.ToLower(code))
				if normalized == "" {
					continue
				}
				result[normalized] = append(result[normalized], set)
			}
		}
	}
	for code := range result {
		sort.Slice(result[code], func(i, j int) bool {
			return result[code][i].RuleID < result[code][j].RuleID
		})
	}
	return result, nil
}
