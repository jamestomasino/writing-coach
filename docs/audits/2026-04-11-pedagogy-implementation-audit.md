# Writing Coach Pedagogy + Implementation Compliance Audit Worksheet

- Audit date: 2026-04-11
- Auditor: Codex (passes 1-3)
- Scope: Pedagogy goals + technical implementation compliance against `README.md`, `docs/architecture.md`, `docs/analyzer-coverage-model.md`, `docs/scoring-phase4-metrics.md`, `docs/tree-library.md`
- Status scale: `PASS` | `FAIL` | `PARTIAL` | `NOT STARTED`

## Worksheet

| ID | Area | Compliance Check | Evidence (code/tests/docs) | Status | Notes |
|---|---|---|---|---|---|
| PED-001 | Core loop | Assignment/review loop preserved end-to-end (path -> assignment -> submission -> review -> revise/move on -> history-informed next step) | `README.md`, `docs/architecture.md`, `internal/api/workspace_handlers.go`, `internal/api/server_test.go` (`TestExerciseSubmissionAndReviewEndpoints`, `TestAssignmentTimelineEndpoint`, `TestClosingAssignmentMarksChainClosedAndBlocksRevision`, `TestArchiveTrackRemovesItAndSwitchesActiveContext`) | PASS | API integration tests cover main and branch flows (review, revision chain, close, archive/context switch). |
| PED-002 | Skill model | Exactly 3 active TGOs enforced at selection/update boundaries | `internal/db/db.go` (`SetActiveTGOs`), `internal/api/workspace_handlers.go` (`setActiveTGOsForSelection`), `internal/domain/tree_validation.go` | PASS | DB + API enforce exactly 3; tree defs require 3 seeds. |
| PED-003 | Skill model | Duplicate/empty TGO selection blocked | `internal/api/workspace_handlers.go` | PASS | Duplicate and empty codes explicitly rejected. |
| PED-004 | Skill model | Selection respects unlock/prerequisite gates | `internal/api/workspace_handlers.go` (`prereqsMetForSelection`, selectable map) | PASS | Only unlocked/non-completed TGOs selectable. |
| PED-005 | Skill model | Built-in trees satisfy DAG integrity and seed constraints | `internal/domain/tree_validation.go`, `internal/domain/tree_catalog_test.go` | PASS | Cycle checks, reachability, unique codes, seed rules covered. |
| PED-006 | Progression | Completed TGOs can be marked and tracked without duplication | `internal/db/db.go` (`MarkTGOCompleted`) | PASS | Idempotent insert with `NOT EXISTS`. |
| PED-007 | Progression | Slipping completed TGOs pause advancement pressure | `internal/curriculum/service.go` | PASS | Explicit slipping check returns hold recommendation. |
| PED-008 | Reviews | Reviews scoped to current active TGOs without silent domain mismatch fallback | `internal/review/service.go` (`reviewTGOs`) | PASS | Remediated in commit `410b1be`: removed hardcoded story fallback; tests updated. |
| PED-009 | Reviews | Deterministic analyzers run before model review | `internal/review/service.go` | PASS | Analyzer report computed before any provider call. |
| PED-010 | Reviews | Deterministic fallback available when provider disabled/fails | `internal/review/service.go` | PASS | Explicit deterministic and deterministic-fallback paths. |
| PED-011 | Reviews | Provider scores are non-authoritative and provenance-tagged | `internal/review/service.go` (`tagProviderScores`, calibration path) | PASS | Non-authoritative provider stream persisted with explicit source tags. |
| LANG-001 | Writing language | Writing language normalized and support-gated | `internal/domain/writing_language.go` | PASS | Normalization + support checks centralized. |
| LANG-002 | Writing language | Deterministic analyzers fail closed for unsupported languages | `internal/analyzer/heuristic.go`, `internal/analyzer/vale.go`, `internal/analyzer/languagetool.go`, `internal/analyzer/nlp.go` | PASS | Unsupported language returns warnings and skips analysis. |
| LANG-003 | Writing language | LLM requests carry writing language field | `internal/llm/types.go`, `internal/openai/client.go`, `internal/anthropic/client.go`, `internal/gemini/client.go` | PASS | Writing language included in request payloads/prompts. |
| ANL-001 | Coverage model | Heuristic applicability gates implemented | `internal/analyzer/rule_registry.go`, `internal/analyzer/heuristic.go` | PASS | Heuristic rules checked through `shouldEvaluateHeuristicRule`. |
| ANL-002 | Coverage model | Typed deterministic registry exists with ownership/layer metadata | `internal/analyzer/rule_registry.go` | PASS | Typed `RuleSpec` registry with owner/layer/category/applicability metadata is present. |
| ANL-003 | Coverage model | Registry is source of truth across deterministic stack (`heuristic`, `vale`, `languagetool`, `nlp`) | `docs/analyzer-coverage-model.md`, `internal/analyzer/rule_registry.go`, `internal/analyzer/coverage_arbitration.go` | PASS | Remediated in commits `b7e324e` + `af365d1`: cross-family ownership metadata added and arbitration now derives from registry. |
| ANL-004 | Coverage model | Ownership non-duplication guards across analyzer families | `internal/analyzer/service.go`, `internal/analyzer/coverage_arbitration.go`, `internal/analyzer/coverage_arbitration_test.go` | PASS | Centralized arbitration enforces owner preference and heuristic fallback policy before merged findings are returned. |
| ANL-005 | Coverage model | Specialty/domain/global precedence conflicts explicitly resolved | `internal/analyzer/coverage_arbitration.go`, `internal/analyzer/coverage_arbitration_test.go` | PASS | Layer precedence (`specialty > domain > global`) is resolved uniformly during deterministic finding arbitration. |
| SCR-001 | Scoring | Domain rubric selection with general fallback | `internal/scoring/engine.go` | PASS | Uses domain rubric, falls back to general. |
| SCR-002 | Scoring | Active TGOs influence candidate scoring skills | `internal/scoring/engine.go` (`candidateSkills`) | PASS | Active TGO skill mappings included. |
| SCR-003 | Metric contract | Phase 4 NLP metric keys emitted by sidecar | `docker/nlp-analyzer/app.py` | PASS | Contract keys populated in `metrics_payload`. |
| SCR-004 | Metric contract | Missing optional metrics treated as no-op in rubric logic | `docs/scoring-phase4-metrics.md`, `internal/scoring/engine.go` (`topScoreGateFailure`) | PASS | Remediated in commit `410b1be`: missing gate metrics no longer force top-score downgrade. |
| SCR-005 | Metric contract | Missing Phase 4 metrics surfaced diagnostically | `internal/scoring/engine.go` (`logMissingPhase4Metrics`) | PASS | Logs partial-metric warnings. |
| OPS-001 | Quality gate | Backend test suite green at audit start | `./scripts/test-backend.sh` run on 2026-04-11 | PASS | All backend packages passed (cached). |
| OPS-002 | Quality gate | Frontend lint/tests/e2e healthy | `./scripts/test-frontend.sh`, `npm run test:e2e`, `web/tests/e2e/*` | PASS | Remediated with commits `7b15159` + `7b51e38`: e2e contract drift fixed and SQLite busy contention mitigations added. |
| OPS-003 | Quality gate | E2E suite is aligned with current UX copy/navigation contracts | `web/tests/e2e/multi-track.spec.ts`, `web/tests/e2e/playground.spec.ts`, `web/src/messages/en.json` | PASS | Updated assertions/selectors to current UX contracts and strengthened helper behavior. |
| DATA-001 | Persistence | Review and score history remain inspectable over time | `internal/api/workspace_handlers.go`, `internal/db/db.go`, `internal/db/reporting_store.go`, `internal/api/server_test.go` | PASS | Timeline/history/review/artifact endpoints and tests confirm assignment-chain and archive/history retrieval behavior. |
| UX-001 | Transparency | User-facing copy still matches analyzer-first architecture | `web/src/messages/en.json`, runtime captures: `docs/audits/screenshots-about-pass3.png`, `docs/audits/screenshots-about-full-pass3.png`, `docs/audits/screenshots-onboarding-pass3.png`, `docs/audits/screenshots-new-assignment-pass3.png` | PASS | Runtime UI copy aligns with documented analyzer-first messaging and onboarding/assignment framing. |
| I18N-001 | Localization robustness | Global `timeZone` configured for `next-intl` to avoid environment fallback mismatches | `web/src/components/app-providers.tsx`, `web/scripts/check-i18n-timezone.mjs`, CI/frontend scripts | PASS | Remediated in commit `7b15159`: explicit timezone configured and guarded in local/CI checks. |
| SEC-001 | Secrets | Provider key encryption + key-management requirements enforced | `internal/api/ai_provider_api.go`, `internal/secrets/crypto.go`, `internal/db/profile_settings_store.go`, `README.md` | PASS | Personal provider storage requires secret, keys are AES-GCM encrypted at rest, and decryption failures are surfaced instead of silently bypassed. |

## Audit Summary (Passes 1-3)

- Checks logged: 31
- `PASS`: 31
- `PARTIAL`: 0
- `FAIL`: 0
- `NOT STARTED`: 0

## Linked Failure/Regression Log

- See `docs/audits/2026-04-11-failures-and-regressions.md`
