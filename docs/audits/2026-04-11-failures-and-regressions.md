# Failures and Regressions Log (For Remediation Planning)

- Audit date: 2026-04-11
- Source worksheet: `docs/audits/2026-04-11-pedagogy-implementation-audit.md`

## Failure Status

Status as of 2026-04-11 remediation pass:

- `Resolved`: `F-001`, `F-002`, `F-003`, `F-004`, `F-005`, `F-006`, `F-007`, `F-008`
- `Open`: none

## Detailed Items

| Failure ID | Worksheet ID | Severity | Type | Finding | Evidence | Impact | Status | Remediation Notes |
|---|---|---|---|---|---|---|---|
| F-001 | PED-008 | High | Compliance failure | Review scoping silently falls back to hardcoded story TGOs when active TGO count is not exactly 3. | `internal/review/service.go` (`reviewTGOs`) | Non-fiction tracks can receive off-domain scoping when data state is inconsistent; violates “reviews tied to current skill goals”. | Resolved | Fixed in commit `410b1be`: removed silent hardcoded fallback and added coverage in `internal/review/deterministic_helpers_test.go`. |
| F-002 | ANL-003 | High | Compliance gap | Analyzer coverage registry is not yet the source of truth across all deterministic analyzer families. | `docs/analyzer-coverage-model.md`, `internal/analyzer/rule_registry.go` comment + scope | Coverage ownership, gating, and duplication policy cannot be centrally enforced for Vale/LanguageTool/NLP; drift risk against model goals. | Resolved | Fixed in commit `af365d1`: added cross-family deterministic ownership metadata in `CurrentDeterministicRuleSpecs` and wired arbitration specs from registry metadata. |
| F-003 | SCR-004 | Medium | Contract mismatch | Missing metrics in top-score gate produce gate failure/downgrade, conflicting with no-op expectation for absent optional metrics. | `docs/scoring-phase4-metrics.md`, `internal/scoring/engine.go` (`topScoreGateFailure`) | Optional metric outages or partial analyzer availability can suppress top scores beyond intended no-op behavior. | Resolved | Fixed in commit `410b1be`: absent gate metrics now no-op; explicit unit coverage added in `internal/scoring/engine_test.go`. |
| F-004 | OPS-003 | Medium | Test regression | E2E specs are partially out of sync with current UX copy/navigation assumptions. | `web/tests/e2e/multi-track.spec.ts`, `web/tests/e2e/playground.spec.ts`, `web/src/messages/en.json` | Failing tests reduce release confidence and can mask real regressions. | Resolved | Fixed in commit `7b15159`: e2e copy/flow assertions updated and hardened against route/label drift. |
| F-005 | OPS-002 | High | Reliability risk | Intermittent `SQLITE_BUSY` during AI job/event writes appears under e2e load and correlates with API failures/timeouts. | E2E logs from `npm run test:e2e` (for example `ai job failed ... database is locked`, `api GET /api/dashboard -> 500`) | Under concurrent load, assignment/review workflows can fail transiently; risks user-facing flakiness in production-like contention. | Resolved | Mitigated in commit `7b51e38`: added SQLite busy retry/backoff in AI job + provider-event write paths (`internal/db/sqlite_retry.go`) with tests. |
| F-006 | I18N-001 | Medium | Configuration gap | `next-intl` emits `ENVIRONMENT_FALLBACK` because global `timeZone` is not configured. | `web/src/components/app-providers.tsx`, Playwright/web logs | Potential locale/date markup mismatch between environments; noisy logs and harder SSR determinism. | Resolved | Fixed in commit `7b15159`: set provider `timeZone`, added guard script, and wired checks into frontend script + CI. |
| F-007 | ANL-004 | High | Governance gap | Ownership non-duplication guardrails are not enforced across `vale`, `languagetool`, and `nlp` alongside heuristics. | `internal/analyzer/heuristic.go`, `internal/analyzer/vale.go`, `internal/analyzer/languagetool.go`, `internal/analyzer/nlp.go` | Duplicate or conflicting deterministic signals can drift over time and skew scoring/feedback emphasis. | Resolved | Fixed in commit `b7e324e` plus `af365d1`: centralized deterministic finding arbitration now enforces owner preference and heuristic fallback policy. |
| F-008 | ANL-005 | Medium | Governance gap | Specialty/domain/global precedence is only partially implemented (primarily heuristic gating). | `docs/analyzer-coverage-model.md`, `internal/analyzer/rule_registry.go`, analyzer implementations | Context-specific precedence conflicts may be resolved inconsistently across analyzer families. | Resolved | Fixed in commit `b7e324e` plus `af365d1`: arbitration resolves applicable ownership by layer precedence (`specialty > domain > global`) with tests. |

## Notes

- Backend tests passed in audit runs (`./scripts/test-backend.sh`).
- Frontend lint/build passed (`./scripts/test-frontend.sh`), but current e2e suite fails and needs remediation (`npm run test:e2e`).

## Deferred Feedback Backlog

These are not classified as regressions from the 2026-04-11 audit worksheet, but they are important pedagogy/product follow-ups requested for the next remediation cycle.

| Feedback ID | Area | Finding | Why It Matters | Status | Suggested Acceptance Criteria |
|---|---|---|---|---|---|
| FB-001 | Assignment generation variety | Subsequent assignment prompts can repeat the same narrative pattern families too often (example: repeated “choice X vs Y under time pressure” framing early in fiction progression). | Repetition can reduce engagement and narrow deliberate practice transfer. Prompt diversity is part of sustained skill growth. | Resolved | Implemented recency-aware assignment summaries in prompt input, added pattern repeat detection (including choice-under-time-pressure), and auto-retry with explicit variety guidance when repetition is detected. Covered by prompt and provider-input unit tests. |
| FB-002 | Learner progression transparency | Learners need clearer narrative guidance on what is required to level up current skills: whether to revise the same draft more, or continue demonstrating control across additional assignments. | Ambiguity in progression expectations can reduce trust and lead to misaligned effort. | Resolved | Added state-driven level-up guidance callouts in both New Assignment and Current Assignment flows. Guidance now explains hold vs revise vs consolidate pathways using progression-hold state and active-skill mastery signals, with e2e coverage asserting guidance is visible in onboarding/new-assignment and active assignment flows. |
