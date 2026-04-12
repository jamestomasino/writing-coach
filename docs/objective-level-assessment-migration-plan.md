# Objective-Level Assessment Migration Plan

- Status: Draft for implementation
- Date: 2026-04-12
- Owner: Product + Engineering

## 1) Purpose

Move deterministic assessment from **Skill Family** level to **Skill Objective** (TGO) level, while keeping the learner experience focused on:

- Skill Families (high-level capabilities)
- Skill Objectives (specific unlockable objectives in the tree)

The current Draft Signals UI will be removed. Historical scored data from past assignments/revisions must remain intact and readable.

## 2) Terminology Model

- Skill Family (internal analogy: TLO)
  - Broad capability (for example: `narrative clarity`, `claim clarity`).
- Skill Objective (internal analogy: ELO; current TGO)
  - Specific measurable objective node (for example: `story-causal-clarity`).
- Progression graph
  - Composed of Skill Objectives; prerequisite/unlock behavior remains objective-driven.

## 3) Target Product Behavior

1. Assignments still target 3 active Skill Objectives.
2. Review scoring and evidence are deterministic at Skill Objective level.
3. Level-up logic uses objective-level evidence/history only.
4. Skill Family pages remain as teaching abstractions and rollups.
5. Draft Signals UI is removed from learner flows.

## 4) Data Compatibility Requirements

Historical data will not be rewritten or discarded.

1. Keep `submission_skill_scores` as immutable historical records.
2. Introduce objective-level score storage for new and future reviews.
3. Provide compatibility reads so old reviews remain viewable in timeline/history.
4. Ensure compare/history pages work across mixed eras (legacy family-scored and new objective-scored reviews).

## 5) Technical Design

### 5.1 Storage

Add a new table for objective-level deterministic scoring (name indicative):

- `submission_objective_scores`
  - `submission_id`
  - `tgo_code`
  - `score` (1-5)
  - `score_source` (start with `deterministic`)
  - `score_version`
  - `score_evidence_json`
  - timestamps

Indexes:

- `(submission_id, tgo_code, score_source)`
- `(tgo_code, created_at)`

### 5.2 Rule System

Define deterministic checks per objective:

1. Every active/public Skill Objective must have at least 1 deterministic check path.
2. Family-level rubrics become fallback templates, not the final assessment unit.
3. Objective scoring should emit objective-specific evidence strings and metric snapshots.

### 5.3 Review Domain Model

Add objective-level scored fields to review payloads:

- `objective_scores[]` (primary scored list for progression era moving forward)

Retain existing fields for compatibility during migration window:

- `skill_scores[]` (legacy/family-scored; read-only for old reviews)

### 5.4 API Strategy

During migration:

1. Existing endpoints return legacy fields for old records.
2. New reviews return `objective_scores` as primary.
3. Frontend views use objective-first rendering where available.
4. No Draft Signals section is shown in learner-facing UI.

## 6) UI/UX Changes

### 6.1 Keep

- Skill Family library/detail pages
- Skill Objective library/detail pages
- Skill tree and objective unlock flow

### 6.2 Remove

- Draft Signals cards/meters/labels in review/compare/timeline views
- Draft-signal explanatory copy

### 6.3 Add/Emphasize

- Objective-level evidence and status clarity for each of the 3 active objectives
- Level-up guidance tied directly to objective evidence streaks and requirements

## 7) Migration Phases

## Phase 0: Objective Guide Deepening (Gate Before Objective Scoring)

Purpose:

Establish clear objective-specific pedagogy before implementing objective-level deterministic checks.

Scope:

1. Expand Skill Objective detail pages with:
   - objective-specific learning intent
   - objective-specific success criteria
   - good/needs-work examples
   - concrete revision moves
   - deterministic assessment focus notes
2. Keep Skill Family pages as crosswalk/context, not the primary measurement surface.
3. Add coverage guardrails so all public objectives can render complete guide content.

Pedagogy refresh checkpoint (required before Phase 1 deterministic objective checks):

1. Re-read the pedagogy paper and validate objective guide voice/claims align.
2. Confirm each objective’s guide language is plain, actionable, and non-contradictory with progression rules.
3. Confirm each objective guide defines at least one deterministic-check intent statement.

## Phase 1: Schema + Dual Write (No UI Removal Yet)

1. Add `submission_objective_scores` table + store methods.
2. Implement deterministic objective scoring writer.
3. Continue writing legacy family scores in parallel for safety.
4. Add integrity tests: every active objective gets deterministic score output.

Exit criteria:

- New reviews persist objective scores reliably.
- No regression in existing review/job flows.

## Phase 2: Objective-First Reads

1. Add `objective_scores` to API responses.
2. Update comparison logic to compare objective-level scores first.
3. Keep fallback behavior for legacy records lacking objective scores.

Exit criteria:

- Mixed historical data renders correctly.
- Compare/history pass for both legacy and new reviews.

## Phase 3: UI Cutover + Draft Signals Removal

1. Remove Draft Signals section from user-facing screens.
2. Render objective-level score/evidence blocks only.
3. Update copy and help text to reference Skill Objectives directly.

Exit criteria:

- No Draft Signals UI remains in learner flows.
- Users can interpret objective progress without ambiguity.

## Phase 4: Legacy Sunset (Optional, Deferred)

1. Stop writing legacy `submission_skill_scores` for new reviews.
2. Keep read support for old records.
3. Evaluate archival strategy later (no destructive migration required).

Exit criteria:

- Objective-level scoring is stable in production.
- Legacy records remain accessible.

## 8) Test Plan

### 8.1 Backend

1. Objective score persistence tests (store and API).
2. Deterministic objective coverage test:
   - Every public objective maps to at least one deterministic check path.
3. Compare tests across:
   - objective vs objective
   - legacy vs legacy
   - mixed legacy/objective eras

### 8.2 Frontend

1. Remove Draft Signals assertions from views.
2. Add assertions for objective-level evidence/status visibility.
3. Timeline/history tests for old reviews still rendering.

### 8.3 Integrity Gates

Keep objective coverage checks in backend CI as a hard gate.

## 9) Rollout + Safety

1. Release behind objective-read feature flag if needed.
2. Deploy Phase 1 first and monitor job success + API payload integrity.
3. Enable objective-first UI only after mixed-data verification.
4. Retain rollback path:
   - fallback to legacy reads without data loss.

## 10) Risks and Mitigations

1. Risk: Mixed-era compare confusion.
   - Mitigation: explicit compare adapter for objective/legacy combinations.
2. Risk: Objective rule gaps.
   - Mitigation: CI gate requiring deterministic path per public objective.
3. Risk: User trust drop during terminology shift.
   - Mitigation: tight copy updates and progressive rollout.

## 11) Acceptance Criteria

1. Deterministic assessment is objective-level for all new reviews.
2. Draft Signals UI is fully removed from learner views.
3. Historical reviews remain accessible with no score-data loss.
4. Progression/level-up decisions use objective evidence history.
5. CI fails if any public objective lacks deterministic scoring coverage.

## 12) Deferred Follow-Up

- Resolve unused skill-family governance (active vs reserved taxonomy) from backlog item `FB-003`.
