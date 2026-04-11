# Progression Hold Enforcement Plan

Date: 2026-04-11
Owner: pedagogy audit remediation

## Objective
Implement rigorous progression-hold behavior for slipping completed objectives without breaking assignment flow, multi-track enrollment behavior, or performance.

## Phase structure

### Phase 1: Hold state foundation (shadow mode)
- Add enrollment-scoped hold fields to curriculum state.
- Set/clear hold state from review outcomes (`CompletedTGOChecks` slipping detection).
- Expose hold state in API curriculum payloads.
- Do not block advancement yet.
- Add tests for:
  - hold activation on slipping completed objective
  - hold clearing when slipping is absent
  - enrollment isolation (multi-path users)

Exit criteria:
- All backend tests pass.
- Hold state persists and is queryable without affecting existing flows.

### Phase 2: Hard advancement gate
- Enforce hold in active-objective selection path:
  - while hold is active, disallow active-TGO changes that advance beyond current set.
- Return deterministic reason code for blocked requests.
- Keep submit/review/revision paths available.
- Add tests for:
  - blocked advancement while hold active
  - allowed no-op selection (same active set)
  - allowed advancement after hold clears

Exit criteria:
- Gate semantics are deterministic and test-covered.
- No regression in assignment/revision flows.

### Phase 3: Decision trace and integrity checks
- Add decision events for review and hold transitions.
- Add CI invariants that enforce event presence and evidence references.
- Add metrics counters for hold activation/clear/block events.

Exit criteria:
- Decision provenance for progression is machine-auditable.
- Integrity checks fail fast on missing pedagogical traces.

## Design constraints
- Enrollment-scoped only: no global user hold.
- O(1) request-path checks: read materialized hold state, do not scan history.
- Additive migrations only, with safe defaults for existing data.
- Deterministic reason codes over prose for stable automation.
