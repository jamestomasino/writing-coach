# Pedagogy Implementation Gap Tracker

Date: 2026-04-11  
Scope: Whitepaper claims in `web/src/app/(app)/about/pedagogy/page.tsx` versus current application behavior.

## Status scale
- `Full`: implemented and evidenced in current code paths.
- `Partial`: implemented in part, but with missing enforcement, observability, or guardrails.
- `Gap`: claim currently not enforced/verified sufficiently.

## Claim-by-claim audit

| ID | Whitepaper claim (paraphrased) | Current implementation evidence | Status | What is missing | Needed additions (tooling/data/intelligence/orchestration/checks) |
|---|---|---|---|---|---|
| C1 | System uses TLO/ELO-style objective decomposition with prerequisite progression. | Tree model includes prerequisites and validation in `internal/domain/tree_validation.go`; unlock logic in `internal/domain/tgo.go` (`NextUnlockedFromDefinition`), selection enforcement in `internal/api/workspace_handlers.go` (`prereqsMetForSelection`). | Full | None critical. | Keep invariant tests for no cycles/missing prereqs and unlock correctness. |
| C2 | Each assignment is constrained to 3 active objectives. | Enforced in API and store: `setActiveTGOsForSelection` and `SetActiveTGOs` both require exactly 3. | Full | None critical. | Keep request-validation and store-level tests. |
| C3 | Workflow is a connected assignment chain, not isolated prompts. | Revision flow links exercises/submissions and assignment chain APIs summarize chain continuity (`internal/api/server.go` assignment chain handlers). | Full | None critical. | Keep chain continuity tests (root/current/revision counts, ordering). |
| C4 | Reviews are deterministic evidence-first before provider language output. | Analyzer runs first; deterministic scoring generated before provider call in `internal/review/service.go`. | Full | None critical. | Keep phase-order unit tests plus review-kind fallback coverage. |
| C5 | Provider scoring is non-authoritative and tagged as such. | Provider scores tagged `llm:<provider>:non_authoritative` via `tagProviderScores`; deterministic scores retained. | Full | None critical. | Keep tests ensuring deterministic score survives conflicts and API output prioritizes deterministic rows. |
| C6 | Prior strengths are rechecked; slipping should pause advancement pressure until stable. | Slipping detection exists (`deterministicCompletedChecks`), and curriculum recommendation downgrades difficulty/returns hold rationale in `internal/curriculum/service.go`. | Partial | Slipping does not currently alter unlock eligibility or progression state directly; completed TGOs remain completed and prerequisite unlock graph is unchanged. This is “advice pressure,” not a hard progression pause. | Add explicit progression-state gate: when slipping exists, set enrollment hold flag and prevent new unlock selections until hold clears. Add integration tests proving blocked advancement during hold. |
| C7 | Revision brief prioritizes highest-leverage fixes. | Review has `NextFocus`, strengths/weaknesses, findings, and optional comparison artifacts. | Partial | No explicit priority model with ranked interventions and confidence/value estimates; prioritization is implicit and non-standardized. | Add deterministic intervention prioritizer (ranked actions with reason codes, impact score, effort score). Persist in artifacts. Add tests for stable ranking under fixed inputs. |
| C8 | Next review checks whether specific deficits were corrected. | Comparison API exists (`review.CompareSubmissions`), with weakness carryover and skill deltas. | Partial | No mandatory closure loop requiring checks against prior targeted deficits as a first-class "intervention outcome" object. | Add intervention tracking schema (target deficit -> expected signal -> observed delta). Add regression tests for closure accounting (resolved/persisting/new). |
| C9 | Each decision should be auditable against preserved evidence. | Review artifacts persist analyzer report, recommendation, comparison, annotations in `review_artifacts`; score rows persist rule evidence and versions. | Partial | Missing unified “decision ledger” linking each consequential decision (recommendation chosen, advancement/hold, TGO completion) to explicit rule version + evidence refs + actor/source. | Add `decision_events` table + emit events in review/save/sync paths. Add integrity checks: each review must yield required event set with non-empty evidence pointers. |
| C10 | Learning progression is coordinated across stages with objective unlocks (not random drift). | Unlock mechanics and objective sets are deterministic; active/completed TGOs are stateful. | Full | None critical. | Keep unlock-path and seed progression tests across all trees. |
| C11 | Evidence-to-design mapping is explicit and defensible. | Whitepaper text maps evidence -> product behavior; code supports many mapped elements. | Partial | Mapping is mostly narrative, not runtime-verifiable. No dashboard proving “this mechanism produced this learning effect.” | Add pedagogy telemetry model and reporting endpoint (objective trajectory, intervention efficacy, hold frequency/resolution, time-to-mastery). |
| C12 | Coaching contract should make next decision clearer each step. | NextFocus and recommendation rationale exist. | Partial | No explicit decision-confidence metric and no user-visible reason codes tied to deterministic evidence. | Add reason-coded recommendation payload and UI display of top evidence lines per decision. Add acceptance tests on reason-code presence. |

## High-priority gaps (order of implementation)

1. **Progression hold is advisory, not enforced (C6)**
- Risk: core pedagogy claim can be bypassed by selecting unlocked objectives while slipped skills persist.
- Required: hard gating semantics in progression state.

2. **No unified decision ledger (C9)**
- Risk: “auditable decisions” claim is only partially true; evidence exists but decision provenance is fragmented.
- Required: evented, versioned decision trace.

3. **Intervention prioritization and closure loop not formalized (C7, C8)**
- Risk: recommendations may be useful but are not consistently measurable as high-leverage interventions.
- Required: deterministic prioritizer + intervention outcome tracking.

4. **Pedagogy telemetry not outcome-linked enough (C11, C12)**
- Risk: we cannot rigorously demonstrate claim-level efficacy over time.
- Required: observability metrics + checks + dashboards.

## Proposed implementation blueprint

### A) Progression hold enforcement
- Add enrollment progression state fields:
  - `hold_active` (bool)
  - `hold_reason_code` (enum, e.g. `completed_tgo_slipping`)
  - `hold_trigger_review_id` (fk)
  - `hold_cleared_review_id` (fk nullable)
- In curriculum sync:
  - if any completed-TGO slipping -> activate hold
  - require N consecutive holding/mastered checks to clear (configurable)
- In active objective selection endpoint:
  - block selection of non-current objectives while hold is active (except explicit override path for admins)
- Tests:
  - slipping triggers hold
  - hold blocks advancement selection
  - hold clears only after threshold evidence

### B) Decision ledger
- Add `decision_events` table:
  - `id`, `user_id`, `tree_id`, `enrollment_id`, `review_id`, `submission_id`, `event_type`, `decision_payload_json`, `rule_version`, `evidence_refs_json`, `created_at`
- Event types:
  - `review_scored`
  - `recommendation_issued`
  - `progression_hold_activated`
  - `progression_hold_cleared`
  - `tgo_completed`
  - `tgo_advancement_blocked`
- Emit events in:
  - review completion pipeline
  - curriculum sync
  - active TGO selection validation
- Checks:
  - CI invariant: every completed review must have `review_scored` + `recommendation_issued` events.
  - CI invariant: slipping check + no hold event => failure.

### C) Intervention prioritizer + closure tracking
- Add deterministic prioritizer:
  - input: analyzer findings, score deltas, previous unresolved deficits
  - output: ordered intervention list with `impact`, `effort`, `confidence`, `reason_codes`
- Persist as artifacts + decision event payload.
- Add intervention lifecycle fields:
  - `introduced_review_id`, `target_signal`, `resolved_review_id`, `status`
- Checks:
  - each review must either resolve or carry forward prior interventions explicitly.
  - regression tests on ranking determinism with fixed fixtures.

### D) Pedagogy observability and governance
- Add rollup metrics:
  - objective mastery velocity
  - hold incidence and mean time-to-clear
  - intervention resolution rate
  - recurrence rate for previously resolved deficits
- Add calibration/governance checks:
  - drift alarms for deterministic distribution shifts
  - conflict-rate thresholds for hybrid calibration
- Add admin endpoint/report:
  - “pedagogy integrity report” with pass/fail checks and trendlines.

## Suggested CI audit gates

1. `pedagogy_invariants_test`
- Exactly 3 active objectives invariant.
- Unlocks only when prerequisites met.
- Slipping completed objective activates hold.
- Hold blocks advancement.

2. `decision_trace_test`
- For each synthetic review completion, required decision events are present with evidence refs and rule versions.

3. `intervention_closure_test`
- Targeted deficits from prior review are either resolved or explicitly persisted with rationale.

4. `rubric_evidence_completeness_test`
- Deterministic score rows must include non-empty evidence JSON and score version.

5. `artifact_contract_test`
- Stored artifacts must include analyzer report, recommendation, comparison, and annotations for non-legacy reviews.

## Immediate next slice (smallest high-value implementation)

1. Implement progression hold enforcement + tests.
2. Introduce minimal decision_events schema and emit `review_scored` + `recommendation_issued`.
3. Add CI gate validating those two events for every review in integration harness.

