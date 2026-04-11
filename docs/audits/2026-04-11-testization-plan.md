# Audit Testization Plan (Post Full Audit)

- Date: 2026-04-11
- Inputs:
  - `docs/audits/2026-04-11-pedagogy-implementation-audit.md`
  - `docs/audits/2026-04-11-failures-and-regressions.md`
- Goal: convert stable audit checks into repeatable automated gates, while keeping judgment-heavy checks as periodic manual review.

## Current CI Baseline

- Backend: `./scripts/test-backend.sh` -> `go test ./cmd/... ./internal/...`
- Frontend: lint + build only (`npm run lint`, `npm run build`)
- E2E is available (`npm run test:e2e`) but not currently in CI.

## Audit-To-Test Matrix

| Audit ID | Should be automated? | Test type | CI stage | Target location(s) | Notes |
|---|---|---|---|---|---|
| PED-001 | Yes | API integration | backend-required | `internal/api/server_test.go` | Keep loop-level route tests for prompt/submission/review/revision/archive timeline. |
| PED-002 | Yes | unit + API integration | backend-required | `internal/db/db_test.go`, `internal/api/server_test.go` | Enforce exactly-3 active TGO invariants at DB/API layers. |
| PED-003 | Yes | API integration | backend-required | `internal/api/server_test.go` | Reject duplicate/empty TGO selection payloads. |
| PED-004 | Yes | API integration | backend-required | `internal/api/server_test.go` | Verify locked/unlocked prerequisite selection behavior. |
| PED-005 | Yes | domain unit tests | backend-required | `internal/domain/tree_catalog_test.go`, `internal/domain/tree_validation_test.go` | Keep DAG/cycle/reachability/seed checks strict. |
| PED-006 | Yes | db unit/integration | backend-required | `internal/db/db_test.go` | Idempotent completion insert behavior. |
| PED-007 | Yes | curriculum unit tests | backend-required | `internal/curriculum/service_test.go` | Slipping completed skills should hold advancement. |
| PED-008 | Yes (critical) | review service unit + API integration | backend-required | `internal/review/service_test.go`, `internal/api/server_test.go` | Add explicit regression test for invalid active-TGO cardinality fallback behavior. |
| PED-009 | Yes | review service unit | backend-required | `internal/review/service_phase1_test.go` | Assert analyzer pass always precedes provider call path. |
| PED-010 | Yes | review service unit | backend-required | `internal/review/service_test.go` | Deterministic fallback path on provider disabled/error. |
| PED-011 | Yes | review service unit | backend-required | `internal/review/service_phase1_test.go` | Score provenance/tagging and non-authoritative provider stream assertions. |
| LANG-001 | Yes | domain unit | backend-required | `internal/domain/writing_language_test.go` | Normalize + support flags locked by tests. |
| LANG-002 | Yes | analyzer unit | backend-required | `internal/analyzer/*_test.go` | Unsupported language fail-closed warnings for each analyzer family. |
| LANG-003 | Yes | provider client unit | backend-required | `internal/openai/client_generation_test.go`, `internal/gemini/*`, `internal/anthropic/*` | Ensure writing language propagates in generation/review requests. |
| ANL-001 | Yes | analyzer unit | backend-required | `internal/analyzer/rule_registry_test.go`, `heuristic_test.go` | Heuristic applicability gating verification. |
| ANL-002 | Yes | static unit (schema/registry completeness) | backend-required | `internal/analyzer/rule_registry_test.go` | Assert registry fields non-empty and structurally valid. |
| ANL-003 | Yes (after fix) | static + integration | backend-required | new `internal/analyzer/ownership_test.go` | Ensure all deterministic findings map to owned registry entries across analyzer families. |
| ANL-004 | Yes (after fix) | static + integration | backend-required | new `internal/analyzer/dedup_test.go` | Duplicate-signal suppression/owner arbitration checks. |
| ANL-005 | Yes (after fix) | integration | backend-required | new `internal/analyzer/precedence_test.go` | Specialty/domain/global precedence resolution checks. |
| SCR-001 | Yes | scoring unit | backend-required | `internal/scoring/engine_test.go` | Domain rubric selection + general fallback. |
| SCR-002 | Yes | scoring unit | backend-required | `internal/scoring/engine_test.go` | Active TGOs influence candidate skills. |
| SCR-003 | Yes | sidecar contract + analyzer unit | backend-required | `docker/nlp-analyzer/test_app.py`, `internal/analyzer/nlp_test.go` | Keep phase-4 key contract stable. |
| SCR-004 | Yes (critical) | scoring unit | backend-required | `internal/scoring/engine_test.go` | Missing-optional-metric behavior contract test. |
| SCR-005 | Yes | scoring unit/log assertion | backend-required | `internal/scoring/engine_test.go` | Partial metric diagnostic coverage. |
| OPS-001 | Yes | existing suite | backend-required | `./scripts/test-backend.sh` | Already enforced. |
| OPS-002 | Yes | e2e | frontend-required | `web/tests/e2e/*` via CI job | Add dedicated e2e CI stage with browser install. |
| OPS-003 | Yes | e2e contract tests | frontend-required | `web/tests/e2e/*` | Prefer `data-testid` assertions over literal copy where possible. |
| DATA-001 | Yes | API integration | backend-required | `internal/api/server_test.go`, `internal/db/*_test.go` | History/timeline/review-artifact integrity. |
| UX-001 | Partially | snapshot/smoke + manual | frontend-advisory | new lightweight UI smoke tests + periodic manual review | Automate structural assertions, keep copy/trust review manual each release. |
| I18N-001 | Yes | frontend unit/config test | frontend-required | new `web/src/lib/i18n-config.test.ts` (or equivalent) | Fail if global `timeZone` not configured. |
| SEC-001 | Yes | unit + integration | backend-required | `internal/secrets/crypto_test.go`, `internal/api/server_test.go` | Secret-required storage, decrypt error behavior, encrypted persistence. |

## What Should Stay Manual (Periodic Audit)

- Pedagogical quality judgment of feedback usefulness/tone beyond deterministic criteria.
- UX truthfulness spot-checks against product messaging (`about` and onboarding narratives), at release milestones.
- Governance intent reviews for analyzer ownership/precedence design changes.

## New CI Stages To Add

1. `backend-required` (existing): no change to trigger policy.
2. `frontend-required` (existing + expanded): keep lint/build and add timezone-config guard.
3. `e2e-required` (new, initially non-blocking for 1-2 PRs, then required):
   - install Playwright browser (`npx playwright install chromium`)
   - run `npm run test:e2e`
   - retain traces/artifacts on failure.
4. `load-advisory` (new nightly/scheduled):
   - concurrency scenario focused on AI jobs + provider-event writes + dashboard reads against SQLite.
   - objective: catch `SQLITE_BUSY` regressions and verify retry/backoff behavior.

## Sequenced Implementation Plan

1. Stabilize and fix failing e2e contracts (`F-004`) and timezone config (`F-006`).
2. Add missing high-risk backend regression tests (`PED-008`, `SCR-004`).
3. Add e2e job to CI as non-blocking for a short burn-in period, then mark required.
4. Implement analyzer governance enforcement (`ANL-003/004/005`) with tests.
5. Add scheduled load/reliability checks for SQLite contention (`F-005`).

## Ownership Suggestion

- Backend platform: `PED-*`, `LANG-*`, `ANL-*`, `SCR-*`, `DATA-*`, `SEC-*`
- Frontend platform: `OPS-002/003`, `I18N-001`, UX smoke assertions
- QA/release owner: manual UX/pedagogy spot-check cadence and sign-off

