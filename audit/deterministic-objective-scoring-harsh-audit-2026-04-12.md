# Deterministic Objective Scoring Harsh Audit (2026-04-12)

## Scope

This audit evaluates whether deterministic scoring is truly objective-specific once writing type/track/path context is applied.

Reviewed surfaces:

- `internal/review/objective_scores.go`
- `internal/scoring/engine.go`
- `internal/scoring/rubrics/*.json`
- `internal/scoring/*test.go`
- `internal/review/objective_scores*_test.go`
- `docs/objective-level-assessment-migration-plan.md`

## Harsh Grading Rubric

Each category is scored `0..5` where:

- `5`: production-grade, adversarially robust, objective-complete
- `3`: acceptable baseline, but material blind spots remain
- `1`: mostly nominal coverage
- `0`: missing capability

Weights are intentionally strict.

| Category | Weight | Score | Weighted |
|---|---:|---:|---:|
| Objective specificity (rule truly bound to objective code) | 25 | 1 | 5.0 |
| Context specificity (domain/track/path shifts materially change objective judgment) | 20 | 2 | 8.0 |
| Evidence quality (traceability from score to objective-specific signals) | 15 | 1 | 3.0 |
| Test depth (positive/negative/adversarial/permutation) | 25 | 2 | 10.0 |
| Anti-gaming robustness | 15 | 2 | 6.0 |
| **Total** | **100** |  | **32.0 / 100** |

## Key Findings (Harsh)

### 1) Objective scoring is still a skill-bridge, not objective-native

Evidence:

- `BuildObjectiveScores` assigns objective score from mapped deterministic **skill** score when present (`basis = deterministic_skill_bridge`).
  - `internal/review/objective_scores.go:55`
  - `internal/review/objective_scores.go:57`
- Evidence kind is explicitly `objective_deterministic_bridge`, not objective rule execution.
  - `internal/review/objective_scores.go:64`

Impact:

- Many objectives sharing one skill inherit the same score logic.
- Objective guide nuance (`assessmentFocus`, etc.) is not an authoritative scoring surface.

Severity: High

---

### 2) Rule granularity is skill-level and often default-template, not objective-level

Evidence:

- Scoring engine computes skill scores via rubric skill config; missing skill config falls back to `DefaultSkillConfig`.
  - `internal/scoring/engine.go:115`
  - `internal/scoring/engine.go:117`
- Candidate scoring set is skill list capped at 8.
  - `internal/scoring/engine.go:252`
  - `internal/scoring/engine.go:294`

Quantified compression (from repo data):

- Public objectives: `884`
- Unique mapped skills: `36`
- Average objectives per skill: `24.56`
- Top overloaded skill families:
  - `revision habits=83`
  - `structural signposting=54`
  - `clarity and coherence=53`

Rubric explicitness coverage (tree-skill pairs):

- Explicit skill config: `55 / 181` (`30.4%`)
- Default-only scoring path: `126 / 181` (`69.6%`)

Impact:

- Most objective judgments collapse onto shared defaults instead of objective semantics.

Severity: High

---

### 3) Coverage test guarantees existence, not validity

Evidence:

- `TestPublicTGOsHaveDeterministicScoreCoverage` uses one synthetic report, then only asserts mapped skill has a deterministic score and rubric id.
  - `internal/scoring/tgo_deterministic_coverage_test.go:18`
  - `internal/scoring/tgo_deterministic_coverage_test.go:81`

What is not tested:

- Objective-specific pass/fail examples.
- Contradictory cases where two objectives sharing a skill should diverge.
- Sensitivity to track/path variants for the same objective concept.

Severity: High

---

### 4) Evidence payload is not objective-actionable enough

Evidence:

- Objective score evidence stores: `tgo_code`, `mapped_skill`, `assessment_status`, `basis` only.
  - `internal/review/objective_scores.go:63`
- It does not include objective-specific triggered rule IDs, threshold boundaries, or span-level trace to source text.

Impact:

- Hard to justify harsh grading decisions for a specific objective.
- Weak forensic/debug value for false positives/negatives.

Severity: Medium-High

---

### 5) Missing optional metrics are no-op (good for resiliency, risky for harsh gating)

Evidence:

- Optional gate metrics are ignored when absent.
  - `internal/scoring/engine_test.go:371`

Impact:

- Improves stability under partial analyzer availability.
- Under harsh grading, this can over-credit submissions when high-discriminative metrics are absent.

Severity: Medium

## Where We Are Sufficient

- Deterministic path exists for every public objective (coverage as existence).
- Domain-level rubric families are active and tested.
- Top-score gate behavior and calibration spread have broad fixture checks.

This is adequate for baseline consistency, not adequate for objective-level strictness.

## Deficiency Summary

1. Objective-level semantics are not first-class in deterministic scoring.
2. Most score paths are shared defaults, not objective-specific rule packs.
3. Test suite is broad but shallow for objective discrimination.
4. Evidence is insufficient for strict adjudication and appeals/debugging.

## Remediation Strategy

## Phase A (Immediate, 1-2 weeks): Integrity Guards

1. Add objective-native rule IDs to deterministic outputs.
2. Require each active objective score to contain:
   - `objective_rule_ids[]`
   - `metric_snapshot` (objective-relevant only)
   - `trigger_summary`
3. CI gate: fail if any public objective resolves with empty objective rule IDs.

Tooling needed:

- `cmd/objective-rule-lint`:
  - verifies every public TGO has at least one objective rule binding
  - verifies evidence schema completeness

## Phase B (Near-term, 2-4 weeks): Objective Rule Manifest

1. Introduce `internal/scoring/objective_rules/*.json` keyed by `tgo_code`.
2. Keep domain rubrics as priors/defaults, but objective rules become primary.
3. Support inheritance model:
   - domain baseline
   - track override
   - objective overlay (authoritative)

Tooling needed:

- rule compiler to merge domain/track/objective into deterministic executable plan
- static validator for conflicting/duplicate thresholds

## Phase C (Near-term, 3-5 weeks): Hard Testization

1. Build an objective test corpus per TGO:
   - `gold_pass.md`
   - `gold_fail.md`
   - `adversarial.md`
2. Add metamorphic tests:
   - adding objective evidence should never lower objective score
   - removing required objective evidence should never raise score
3. Add pairwise discrimination tests:
   - objectives sharing a skill must diverge when objective-specific criteria diverge

Tooling needed:

- `cmd/objective-score-harness` to run corpus + metamorphic checks and emit regression diffs

## Phase D (Medium-term, 4-8 weeks): Anti-gaming + Observability

1. Add score confidence and risk flags (`metric_missing`, `rule_sparse`, `rule_conflict`).
2. Add leaderboard for weak objectives:
   - highest false-positive/false-negative drift
   - largest disagreement vs manual benchmark set
3. Add policy: no objective may ship with only default fallback logic.

Tooling needed:

- objective drift dashboard using calibration snapshots
- mutation testing for rubric/objective rules

## Proposed Release Gates (Harsh)

A release fails if any condition is true:

1. Any public objective lacks objective-native rule IDs.
2. Any objective has <2 deterministic test fixtures (pass/fail).
3. Pairwise discrimination suite reports unresolved collisions.
4. >5% of objective scores in staging have `metric_missing` risk flag.
5. Objective evidence payload missing threshold/rule trace fields.

## Recommended Next Actions

1. Build `objective-rule-lint` and wire it into backend CI.
2. Define objective rule manifest schema and migrate one domain first (academic).
3. Create a seed corpus for top overloaded skills (`revision habits`, `structural signposting`, `clarity and coherence`) and run discrimination tests.
4. Add evidence schema v2 with objective rule trace fields before further rubric tuning.
