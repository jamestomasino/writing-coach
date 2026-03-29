package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/tomasino/writing-coach/internal/analyzer"
	"github.com/tomasino/writing-coach/internal/config"
	"github.com/tomasino/writing-coach/internal/domain"
	"github.com/tomasino/writing-coach/internal/scoring"
)

type submissionSnapshot struct {
	SubmissionID int64
	TreeSlug     string
	Content      string
	WordCount    int
	Before       map[string]domain.SkillScore
	Evidence     map[string]scoring.ScoreEvidence
}

type trackStats struct {
	TrackSlug    string
	Domain       string
	Compared     int
	BeforeFive   int
	AfterFive    int
	BeforeSum    int
	AfterSum     int
	BeforeMaxSum int
	AfterMaxSum  int
}

type falseNegativeCandidate struct {
	SubmissionID int64
	TrackSlug    string
	Domain       string
	BeforeMax    int
	AfterMax     int
	BeforeAvg    float64
	AfterAvg     float64
	AvgDrop      float64
	Dropped      []string
	Reasons      []string
}

func main() {
	ctx := context.Background()

	projectRoot, err := os.Getwd()
	if err != nil {
		die("resolve cwd", err)
	}
	cfg, err := config.Load(projectRoot)
	if err != nil {
		die("load config", err)
	}

	dbPath := flag.String("db", cfg.DatabaseURL, "path to sqlite database")
	limitPerTrack := flag.Int("limit-per-track", 80, "max submissions per track (newest first)")
	minSamples := flag.Int("min-samples", 20, "minimum submissions required for stable calibration conclusions")
	outPath := flag.String("out", "", "optional report output path")
	flag.Parse()

	if *limitPerTrack <= 0 {
		dieMsg("--limit-per-track must be > 0")
	}
	if *minSamples <= 0 {
		dieMsg("--min-samples must be > 0")
	}

	resolvedDBPath := resolveDBPath(projectRoot, *dbPath)
	snaps, err := loadSnapshots(ctx, resolvedDBPath, *limitPerTrack, cfg.DefaultTreeSlug)
	if err != nil {
		die("load snapshots", err)
	}
	if len(snaps) == 0 {
		dieMsg("no deterministic submission scores found")
	}

	engine, err := scoring.NewEngine()
	if err != nil {
		die("init scoring engine", err)
	}

	trackOrder := make([]string, 0)
	trackSeen := map[string]bool{}
	trackAgg := map[string]*trackStats{}
	var candidates []falseNegativeCandidate

	for _, snap := range snaps {
		opts := analyzer.ContextOptions{
			TreeSlug:    snap.TreeSlug,
			WritingType: writingTypeForTree(snap.TreeSlug),
		}
		report := reconstructReport(snap)
		afterScores, err := engine.ScoreSubmission(
			domain.Submission{ID: snap.SubmissionID, Content: snap.Content, WordCount: snap.WordCount},
			report,
			opts,
			nil,
		)
		if err != nil {
			continue
		}

		afterBySkill := map[string]domain.SkillScore{}
		for _, s := range afterScores {
			afterBySkill[s.Skill] = s
		}

		domainName := analyzer.DomainForContext(opts)
		if !trackSeen[snap.TreeSlug] {
			trackOrder = append(trackOrder, snap.TreeSlug)
			trackSeen[snap.TreeSlug] = true
		}
		agg := trackAgg[snap.TreeSlug]
		if agg == nil {
			agg = &trackStats{TrackSlug: snap.TreeSlug, Domain: domainName}
			trackAgg[snap.TreeSlug] = agg
		}

		beforeVals := make([]domain.SkillScore, 0, len(snap.Before))
		afterVals := make([]domain.SkillScore, 0, len(snap.Before))
		for skill, before := range snap.Before {
			after, ok := afterBySkill[skill]
			if !ok {
				continue
			}
			agg.Compared++
			agg.BeforeSum += before.Score
			agg.AfterSum += after.Score
			if before.Score == 5 {
				agg.BeforeFive++
			}
			if after.Score == 5 {
				agg.AfterFive++
			}
			beforeVals = append(beforeVals, before)
			afterVals = append(afterVals, after)
		}
		if len(beforeVals) == 0 {
			continue
		}

		beforeMax := maxScore(beforeVals)
		afterMax := maxScore(afterVals)
		agg.BeforeMaxSum += beforeMax
		agg.AfterMaxSum += afterMax

		beforeAvg := avgScore(beforeVals)
		afterAvg := avgScore(afterVals)
		if beforeMax == 5 && afterMax <= 4 {
			dropped, reasons := explainDrops(snap.Before, afterBySkill)
			candidates = append(candidates, falseNegativeCandidate{
				SubmissionID: snap.SubmissionID,
				TrackSlug:    snap.TreeSlug,
				Domain:       domainName,
				BeforeMax:    beforeMax,
				AfterMax:     afterMax,
				BeforeAvg:    beforeAvg,
				AfterAvg:     afterAvg,
				AvgDrop:      beforeAvg - afterAvg,
				Dropped:      dropped,
				Reasons:      reasons,
			})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].AvgDrop == candidates[j].AvgDrop {
			return candidates[i].BeforeAvg > candidates[j].BeforeAvg
		}
		return candidates[i].AvgDrop > candidates[j].AvgDrop
	})

	report := renderReport(trackOrder, trackAgg, candidates, len(snaps), *limitPerTrack, *minSamples, resolvedDBPath)
	fmt.Print(report)
	if strings.TrimSpace(*outPath) != "" {
		if err := os.MkdirAll(filepath.Dir(*outPath), 0o755); err != nil {
			die("mkdir output dir", err)
		}
		if err := os.WriteFile(*outPath, []byte(report), 0o644); err != nil {
			die("write report", err)
		}
	}
}

func loadSnapshots(ctx context.Context, dbPath string, limitPerTrack int, defaultTreeSlug string) ([]submissionSnapshot, error) {
	sqlDB, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}
	defer sqlDB.Close()

	sqlDB.SetMaxOpenConns(1)
	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, err
	}

	scoreSourceExists, err := hasColumn(ctx, sqlDB, "submission_skill_scores", "score_source")
	if err != nil {
		return nil, err
	}
	scoreEvidenceExists, err := hasColumn(ctx, sqlDB, "submission_skill_scores", "score_evidence_json")
	if err != nil {
		return nil, err
	}

	evidenceSelect := "'{}' AS score_evidence_json"
	if scoreEvidenceExists {
		evidenceSelect = "COALESCE(score_evidence_json, '{}') AS score_evidence_json"
	}
	scoreSourceWhere := ""
	if scoreSourceExists {
		scoreSourceWhere = "WHERE score_source = 'deterministic'"
	}

	query := fmt.Sprintf(`
		WITH ranked_scores AS (
			SELECT
				submission_id,
				skill_name,
				score,
				%s,
				ROW_NUMBER() OVER (
					PARTITION BY submission_id, skill_name
					ORDER BY id DESC
				) AS rn
			FROM submission_skill_scores
			%s
		),
		track_ranked AS (
			SELECT
				s.id AS submission_id,
				COALESCE(NULLIF(t.slug, ''), NULLIF(u.active_tree_slug, ''), ?) AS tree_slug,
				s.content,
				s.word_count,
				rs.skill_name,
				rs.score,
				rs.score_evidence_json,
				ROW_NUMBER() OVER (
					PARTITION BY COALESCE(NULLIF(t.slug, ''), NULLIF(u.active_tree_slug, ''), ?)
					ORDER BY s.id DESC
				) AS track_rank
			FROM submissions s
			LEFT JOIN users u ON u.id = s.user_id
			LEFT JOIN tgo_trees t ON t.id = s.tree_id
			JOIN ranked_scores rs ON rs.submission_id = s.id
			WHERE rs.rn = 1
		)
		SELECT
			submission_id,
			tree_slug,
			content,
			word_count,
			skill_name,
			score,
			score_evidence_json
		FROM track_ranked
		WHERE track_rank <= ?
		ORDER BY submission_id DESC
	`, evidenceSelect, scoreSourceWhere)

	rows, err := sqlDB.QueryContext(ctx, query, defaultTreeSlug, defaultTreeSlug, limitPerTrack)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bySubmission := map[int64]*submissionSnapshot{}
	order := make([]int64, 0)
	for rows.Next() {
		var submissionID int64
		var treeSlug, content, skill, evidenceRaw string
		var wordCount, score int
		if err := rows.Scan(&submissionID, &treeSlug, &content, &wordCount, &skill, &score, &evidenceRaw); err != nil {
			return nil, err
		}
		snap := bySubmission[submissionID]
		if snap == nil {
			snap = &submissionSnapshot{
				SubmissionID: submissionID,
				TreeSlug:     treeSlug,
				Content:      content,
				WordCount:    wordCount,
				Before:       map[string]domain.SkillScore{},
				Evidence:     map[string]scoring.ScoreEvidence{},
			}
			bySubmission[submissionID] = snap
			order = append(order, submissionID)
		}
		snap.Before[skill] = domain.SkillScore{Skill: skill, Score: score, ScoreEvidenceJSON: evidenceRaw}
		if strings.TrimSpace(evidenceRaw) != "" && evidenceRaw != "{}" {
			var evidence scoring.ScoreEvidence
			if err := json.Unmarshal([]byte(evidenceRaw), &evidence); err == nil {
				snap.Evidence[skill] = evidence
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	snaps := make([]submissionSnapshot, 0, len(order))
	for _, id := range order {
		snaps = append(snaps, *bySubmission[id])
	}
	return snaps, nil
}

func resolveDBPath(projectRoot, configuredPath string) string {
	configuredPath = strings.TrimSpace(configuredPath)
	if configuredPath != "" {
		if _, err := os.Stat(configuredPath); err == nil {
			return configuredPath
		}
	}
	fallback := filepath.Join(projectRoot, ".writing-coach", "writing-coach.db")
	if _, err := os.Stat(fallback); err == nil {
		return fallback
	}
	return configuredPath
}

func hasColumn(ctx context.Context, db *sql.DB, tableName, columnName string) (bool, error) {
	rows, err := db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return false, err
	}
	defer rows.Close()

	for rows.Next() {
		var (
			cid      int
			name     string
			ctype    string
			notNull  int
			defaultV sql.NullString
			primary  int
		)
		if err := rows.Scan(&cid, &name, &ctype, &notNull, &defaultV, &primary); err != nil {
			return false, err
		}
		if strings.EqualFold(strings.TrimSpace(name), strings.TrimSpace(columnName)) {
			return true, nil
		}
	}
	if err := rows.Err(); err != nil {
		return false, err
	}
	return false, nil
}

func reconstructReport(snap submissionSnapshot) analyzer.Report {
	heuristic := analyzer.Heuristic{}
	opts := analyzer.ContextOptions{TreeSlug: snap.TreeSlug, WritingType: writingTypeForTree(snap.TreeSlug)}
	report, err := heuristic.AnalyzeWithContext(context.Background(), snap.Content, opts)
	if err != nil {
		report = analyzer.Report{Metrics: map[string]int{}}
	}
	if report.Metrics == nil {
		report.Metrics = map[string]int{}
	}
	if snap.WordCount > 0 {
		report.Metrics["word_count"] = snap.WordCount
	}

	categoryHistogram := map[string]int{}
	maxFindingCount := 0
	for _, evidence := range snap.Evidence {
		if evidence.FindingCount > maxFindingCount {
			maxFindingCount = evidence.FindingCount
		}
		for category, count := range evidence.CategoryHistogram {
			category = strings.ToLower(strings.TrimSpace(category))
			if category == "" {
				continue
			}
			if count > categoryHistogram[category] {
				categoryHistogram[category] = count
			}
		}
		for metric, value := range evidence.MetricSnapshot {
			report.Metrics[metric] = value
		}
	}

	findings := make([]analyzer.Finding, 0)
	for category, count := range categoryHistogram {
		for i := 0; i < count; i++ {
			findings = append(findings, analyzer.Finding{Category: category, Severity: "warning"})
		}
	}
	if len(findings) < maxFindingCount {
		for i := len(findings); i < maxFindingCount; i++ {
			findings = append(findings, analyzer.Finding{Category: "uncategorized", Severity: "warning"})
		}
	}
	report.Findings = findings
	return report
}

func writingTypeForTree(treeSlug string) string {
	if tree, ok := domain.BuiltInTreeBySlug(treeSlug); ok {
		return tree.Title
	}
	return treeSlug
}

func avgScore(scores []domain.SkillScore) float64 {
	if len(scores) == 0 {
		return 0
	}
	total := 0
	for _, score := range scores {
		total += score.Score
	}
	return float64(total) / float64(len(scores))
}

func maxScore(scores []domain.SkillScore) int {
	max := 0
	for _, score := range scores {
		if score.Score > max {
			max = score.Score
		}
	}
	return max
}

func explainDrops(before map[string]domain.SkillScore, after map[string]domain.SkillScore) ([]string, []string) {
	type drop struct {
		skill  string
		delta  int
		reason string
	}
	drops := make([]drop, 0)
	for skill, b := range before {
		a, ok := after[skill]
		if !ok || a.Score >= b.Score {
			continue
		}
		reason := ""
		if strings.TrimSpace(a.ScoreEvidenceJSON) != "" && a.ScoreEvidenceJSON != "{}" {
			var evidence scoring.ScoreEvidence
			if err := json.Unmarshal([]byte(a.ScoreEvidenceJSON), &evidence); err == nil {
				for i := len(evidence.AppliedRules) - 1; i >= 0; i-- {
					rule := strings.TrimSpace(evidence.AppliedRules[i])
					if strings.Contains(rule, "top score gate") {
						reason = rule
						break
					}
				}
				if reason == "" && len(evidence.AppliedRules) > 0 {
					reason = evidence.AppliedRules[len(evidence.AppliedRules)-1]
				}
			}
		}
		drops = append(drops, drop{skill: skill, delta: b.Score - a.Score, reason: reason})
	}
	sort.Slice(drops, func(i, j int) bool {
		if drops[i].delta == drops[j].delta {
			return drops[i].skill < drops[j].skill
		}
		return drops[i].delta > drops[j].delta
	})

	skills := make([]string, 0, min(3, len(drops)))
	reasons := make([]string, 0, min(3, len(drops)))
	for i := 0; i < len(drops) && i < 3; i++ {
		skills = append(skills, fmt.Sprintf("%s (-%d)", drops[i].skill, drops[i].delta))
		if strings.TrimSpace(drops[i].reason) != "" {
			reasons = append(reasons, fmt.Sprintf("%s: %s", drops[i].skill, drops[i].reason))
		}
	}
	return skills, reasons
}

func renderReport(trackOrder []string, agg map[string]*trackStats, candidates []falseNegativeCandidate, submissionCount, limitPerTrack, minSamples int, dbPath string) string {
	var b strings.Builder
	now := time.Now().UTC().Format(time.RFC3339)
	fmt.Fprintf(&b, "# Scoring Backtest Report\n\n")
	fmt.Fprintf(&b, "Generated: %s\n", now)
	fmt.Fprintf(&b, "Database: `%s`\n", dbPath)
	fmt.Fprintf(&b, "Submissions sampled: %d (limit per track: %d)\n\n", submissionCount, limitPerTrack)
	if submissionCount < minSamples {
		fmt.Fprintf(&b, "## Data Sufficiency Warning\n\n")
		fmt.Fprintf(&b, "Only %d submissions were sampled. Minimum recommended for calibration decisions is %d.\n\n", submissionCount, minSamples)
	}

	type domainAgg struct {
		Compared   int
		BeforeFive int
		AfterFive  int
		BeforeSum  int
		AfterSum   int
	}
	domainRollup := map[string]*domainAgg{}

	fmt.Fprintf(&b, "## Track Summary\n\n")
	fmt.Fprintf(&b, "| track | domain | compared skills | 5%% before | 5%% after | avg before | avg after | avg shift |\n")
	fmt.Fprintf(&b, "|---|---:|---:|---:|---:|---:|---:|---:|\n")
	for _, track := range trackOrder {
		st := agg[track]
		if st == nil || st.Compared == 0 {
			continue
		}
		beforeFivePct := pct(st.BeforeFive, st.Compared)
		afterFivePct := pct(st.AfterFive, st.Compared)
		beforeAvg := float64(st.BeforeSum) / float64(st.Compared)
		afterAvg := float64(st.AfterSum) / float64(st.Compared)
		fmt.Fprintf(&b, "| %s | %s | %d | %.1f%% | %.1f%% | %.2f | %.2f | %+0.2f |\n", track, st.Domain, st.Compared, beforeFivePct, afterFivePct, beforeAvg, afterAvg, afterAvg-beforeAvg)

		d := domainRollup[st.Domain]
		if d == nil {
			d = &domainAgg{}
			domainRollup[st.Domain] = d
		}
		d.Compared += st.Compared
		d.BeforeFive += st.BeforeFive
		d.AfterFive += st.AfterFive
		d.BeforeSum += st.BeforeSum
		d.AfterSum += st.AfterSum
	}

	domains := make([]string, 0, len(domainRollup))
	for domainName := range domainRollup {
		domains = append(domains, domainName)
	}
	sort.Strings(domains)

	fmt.Fprintf(&b, "\n## Domain Summary\n\n")
	fmt.Fprintf(&b, "| domain | compared skills | 5%% before | 5%% after | avg before | avg after | avg shift |\n")
	fmt.Fprintf(&b, "|---|---:|---:|---:|---:|---:|---:|\n")
	for _, domainName := range domains {
		d := domainRollup[domainName]
		if d.Compared == 0 {
			continue
		}
		beforeFivePct := pct(d.BeforeFive, d.Compared)
		afterFivePct := pct(d.AfterFive, d.Compared)
		beforeAvg := float64(d.BeforeSum) / float64(d.Compared)
		afterAvg := float64(d.AfterSum) / float64(d.Compared)
		fmt.Fprintf(&b, "| %s | %d | %.1f%% | %.1f%% | %.2f | %.2f | %+0.2f |\n", domainName, d.Compared, beforeFivePct, afterFivePct, beforeAvg, afterAvg, afterAvg-beforeAvg)
	}

	fmt.Fprintf(&b, "\n## Potential False-Negative Candidates (5 -> 4 cap)\n\n")
	if len(candidates) == 0 {
		fmt.Fprintf(&b, "none\n")
		return b.String()
	}
	limit := min(20, len(candidates))
	fmt.Fprintf(&b, "| submission | track | domain | before max/avg | after max/avg | avg drop | dropped skills |\n")
	fmt.Fprintf(&b, "|---:|---|---|---|---|---:|---|\n")
	for i := 0; i < limit; i++ {
		c := candidates[i]
		fmt.Fprintf(&b, "| %d | %s | %s | %d / %.2f | %d / %.2f | %.2f | %s |\n", c.SubmissionID, c.TrackSlug, c.Domain, c.BeforeMax, c.BeforeAvg, c.AfterMax, c.AfterAvg, c.AvgDrop, strings.Join(c.Dropped, ", "))
	}
	fmt.Fprintf(&b, "\n### Candidate Reasons\n\n")
	for i := 0; i < limit; i++ {
		c := candidates[i]
		fmt.Fprintf(&b, "- submission %d (%s): ", c.SubmissionID, c.TrackSlug)
		if len(c.Reasons) == 0 {
			fmt.Fprintf(&b, "no explicit gate reason captured\n")
			continue
		}
		fmt.Fprintf(&b, "%s\n", strings.Join(c.Reasons, " | "))
	}

	return b.String()
}

func pct(numerator, denominator int) float64 {
	if denominator == 0 {
		return 0
	}
	return (float64(numerator) / float64(denominator)) * 100
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func die(stage string, err error) {
	fmt.Fprintf(os.Stderr, "error: %s: %v\n", stage, err)
	os.Exit(1)
}

func dieMsg(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}
