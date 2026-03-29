# Scoring Pipeline Roadmap

## Purpose
Define a deterministic-first scoring system that materially raises review quality, keeps scoring explainable, and uses LLM scoring as a secondary layer constrained by deterministic evidence.

This document is the working plan for implementation and rollout.

## Product Direction
- Deterministic scoring leads.
- LLM scoring is secondary and must be influenced by deterministic outputs.
- Rubrics are organized by writing path and domain so scoring reflects real coaching intent.
- Trend analytics must remain stable and comparable over time.

## Current Baseline (as implemented)
- Active analyzers: Heuristic, Vale, LanguageTool, NLP sidecar.
- Analyzer orchestration: `internal/app/app.go` wires all analyzers into `analyzer.NewService`.
- Deterministic fallback scoring exists, but is intentionally simple:
  - score dimensions currently derive from word count, average sentence length, and total finding count.
  - logic lives in `internal/review/service.go` (`defaultScoresForActiveTGOs`, `scoreFrom*`).
- When model review is enabled, skill scores are model-generated and persisted in the same table as deterministic scores.
- Score trends read from `submission_skill_scores` with no source separation.

## Design Principles
1. Deterministic-first evidence
- Every score must be traceable to explicit signals (metrics/findings/rules).

2. Path-aware rubrics
- Scoring dimensions and thresholds should vary by domain/path (fiction, technical, academic, etc.).

3. Stable progression math
- Trends and curriculum decisions must avoid accidental jumps due to scorer changes.

4. Versioned scoring
- We must be able to answer: which scorer and rubric version produced this score?

5. LLM constrained role
- LLM can refine language and optionally calibrate edge cases, but cannot ignore deterministic evidence.

## Domain and Path Structure for Rubrics
Use existing domain normalization in `internal/analyzer/context_domain.go` as the canonical first-level rubric router:
- `fiction`
- `fantasy`
- `technical`
- `academic`
- `professional`
- `thought_leadership`
- `marketing`
- `general`

Then specialize by track/path slug (examples from `internal/domain/tree_catalog.go`):
- `story-craft-track`, `fantasy-fiction-track`, `science-fiction-track`, `memoir-personal-narrative-track`
- `technical-writing-track`, `academic-essay-track`, `professional-writing-track`
- `thought-leadership-track`, `journalism-reporting-track`, `persuasive-writing-track`
- `marketing-writing-track`, `content-marketing-track`, `grant-writing-track`, `educational-writing-track`

Routing rule for Phase 1:
- Primary: domain-level rubric family.
- Secondary override: track-level threshold/profile adjustments where needed.

## Target Architecture (end state)
1. Analyzer layer
- Produces normalized findings + metrics + deterministic evidence spans.

2. Deterministic scoring engine
- Applies rubric definitions (weights, thresholds, caps, floors) to produce per-skill scores.
- Emits explainability payload per skill.

3. Optional LLM calibration layer
- Receives deterministic summary + evidence and may propose bounded adjustments.
- Cannot create scores for unsupported skills or exceed configured adjustment bounds.

4. Persistence + reporting
- Store score source and score version.
- Separate trend views by source and blended mode.

## Phased Plan
- Phase 1: Deterministic scoring foundation (this document details implementation).
- Phase 2: Draft-to-draft delta scoring and evidence-linked annotations.
- Phase 3: LLM constrained calibration + conflict detection.
- Phase 4: Expanded deterministic tools (claim-evidence structure, cohesion/referent clarity, repetition/drift).

### Future Deterministic Tooling Notes
- Planned addition: `alex` (inclusive language checks) as a deterministic signal source.
- Evaluate for fit: `fastText` language identification as a pre-check for deterministic analyzer language gating.
- Constraints for both:
  - keep findings bounded and low-noise for coaching contexts,
  - preserve deterministic evidence contracts (`score_evidence_json`),
  - validate false-positive rate before promoting to default scoring inputs.

### Phase 2 Calibration Focus (added after Phase 1 validation)
- Tighten top-end score separation to avoid premature ceiling effects across domains.
- Add stronger rubric distinctions between high-performing and exceptional submissions (especially around 4->5 transitions).
- Validate with paired track calibration runs (at least one narrative track and one technical/informational track) before promoting rubric version updates.

---

## Phase 1: Deterministic Foundation Plan

### Goals
- Replace simplistic fallback scoring with rubric-driven deterministic scoring.
- Preserve current API behavior while improving scoring quality.
- Make score provenance explicit in the data model.
- Keep LLM review functional, but subordinate score handling to deterministic data contracts.

### Scope
In scope:
- New deterministic scoring engine module.
- Rubric schema + initial rubric packs by domain.
- Data model additions for score provenance/versioning.
- Deterministic explanation payload for each score.
- Reporting updates to filter/group by score source.

Out of scope (Phase 2+):
- Full delta scorer.
- LLM-based score adjustment logic.
- New heavy analyzers requiring additional sidecars.

### Workstreams

#### 1) Data model and migrations
Add fields to `submission_skill_scores` (or equivalent new table if preferred):
- `score_source TEXT NOT NULL` (`deterministic`, `llm`, `hybrid`)
- `score_version TEXT NOT NULL` (e.g. `det-v1`)
- `score_evidence_json TEXT NOT NULL DEFAULT '{}'`

Add index suggestions:
- `(submission_id, skill_name)`
- `(score_source, score_version)`
- `(skill_name, score_source)`

Compatibility:
- Backfill existing rows with:
  - `score_source = 'llm_or_legacy'` (or `legacy`) to avoid false certainty.
  - `score_version = 'legacy-unknown'`.

#### 2) Deterministic scorer engine
Create a dedicated package, e.g. `internal/scoring`:
- Input:
  - `analyzer.Report`
  - `AnalyzerContext` (domain/path)
  - active TGOs/completed TGOs
- Output:
  - `[]domain.SkillScore`
  - explanation objects per score (evidence/features used)

Behavior:
- Skill score = weighted aggregation of deterministic features.
- Clamp to `[1..5]`.
- Only score skills relevant to active TGOs + priority path skills (bounded list).
- Deterministic scoring runs for all reviews (not only fallback mode).

#### 3) Rubric definitions
Add rubric config files (YAML/JSON) under e.g. `internal/scoring/rubrics/`:
- `domain-fiction.yaml`
- `domain-technical.yaml`
- `domain-academic.yaml`
- `domain-professional.yaml`
- `domain-thought-leadership.yaml`
- `domain-marketing.yaml`
- `domain-general.yaml`

Each rubric defines:
- skill list
- feature mappings from analyzer metrics/findings
- per-feature thresholds and weights
- penalties/caps
- optional track-level override hooks

#### 4) Review pipeline integration
Update `internal/review/service.go`:
- deterministic scores computed first from analyzer report.
- if LLM returns scores, do not blindly replace deterministic scores in Phase 1.
- keep LLM review narrative fields as-is.

Phase 1 policy:
- Persist deterministic scores as authoritative scoring stream.
- Optionally persist model scores separately only if source-tagged and non-authoritative.

#### 5) Reporting and analytics updates
Update reporting queries:
- support `score_source` filters.
- default progress/averages to deterministic stream.
- allow comparison dashboards later (`deterministic` vs `llm`).

#### 6) Observability and QA
Add scoring diagnostics:
- count of skills scored per submission.
- missing rubric warnings.
- per-domain score distribution sanity checks.

Add tests:
- rubric load/validation tests.
- deterministic score regression tests by domain fixtures.
- migration tests for backfill and query compatibility.

### Barrier Considerations and Mitigations

1. Mixed-source score contamination
- Barrier: existing trend logic combines all rows.
- Mitigation: source/version fields + deterministic-default reporting.

2. Threshold calibration drift across writing types
- Barrier: one-size thresholds will mis-score domains.
- Mitigation: domain-first rubric families + track overrides.

3. Explainability trust gap
- Barrier: opaque scores reduce coach confidence.
- Mitigation: require evidence payload per score and expose in artifacts.

4. Language support limitations
- Barrier: deterministic analyzers are English-gated today.
- Mitigation: fail closed for unsupported languages with explicit warning; do not fake deterministic scores.

5. Backward compatibility risk
- Barrier: existing UI/API expects score arrays.
- Mitigation: keep shape stable; enrich internals first; expose optional metadata fields behind additive API changes.

6. Migration rollback constraints
- Barrier: migration runner is forward-only; no down migration framework exists.
- Mitigation: treat migration `0028` as additive/forward-only and require:
  - database snapshot before deploy,
  - staged rollout (apply migration before app rollout),
  - emergency fallback via app rollback only (schema remains compatible because fields are additive),
  - documented reindex/drop SQL runbook only if absolutely required.

### Phase 1 Acceptance Criteria
- Deterministic scorer produces path-aware per-skill scores for all supported English submissions.
- Every persisted score has `score_source` and `score_version`.
- Reporting defaults to deterministic stream and remains query-compatible.
- Lint/tests pass and migrations are reversible.
- At least one fixture suite per domain family validates scoring behavior.

### Phase 1 Operational Rollout Notes
- Migration `0028_add_skill_score_provenance.sql` is additive and safe to deploy before application update.
- Because migrations are forward-only in this repo, rollback means:
  - roll back app binaries first,
  - keep additive columns in place,
  - restore database from snapshot only if data-level rollback is required.

### Delivery Sequence
1. Migration + model updates (`score_source`, `score_version`, evidence payload).
2. Scoring engine scaffold + rubric loader + validation tests.
3. Domain rubric v1 implementation using existing analyzer signals.
4. Review service integration (deterministic-first score persistence).
5. Reporting query updates.
6. Fixture calibration pass and documentation updates.

### Decisions Locked for Phase 1
- Deterministic scores are primary.
- LLM score outputs are non-authoritative unless explicitly promoted in a later phase.
- Rubric routing uses `DomainForContext` first, then optional track overrides.

### Open Questions
- Whether to store model scores in a separate table vs same table with strict `score_source`.
- Whether curriculum progression should immediately ignore non-deterministic scores or support blended mode behind a flag.
- Which UI surfaces should show score evidence in Phase 1 vs Phase 2.
