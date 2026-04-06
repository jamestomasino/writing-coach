# Architecture Refactor Plan (Performance, Security, Simplicity)

## Scope and Constraints
- Prioritize architectural structure first, with behavior-preserving refactors before policy changes.
- Keep runtime behavior stable unless explicitly called out.
- Maintain or improve test coverage relative to baseline.

## Baseline
- Test command set:
  - `go test ./internal/analyzer ./internal/api ./internal/config ./internal/curriculum ./internal/db ./internal/domain ./internal/openai ./internal/prompt ./internal/review ./internal/scoring ./internal/secrets -coverprofile=/tmp/writing_coach_cover_before.out`
- Baseline coverage:
  - `53.8%` total statements (`/tmp/writing_coach_cover_before.out`)

## Coverage Gate
For every stage:
1. Run full repo tests: `go test ./...`
2. Run coverage suite for internal tested packages:
   - `go test ./internal/analyzer ./internal/api ./internal/config ./internal/curriculum ./internal/db ./internal/domain ./internal/openai ./internal/prompt ./internal/review ./internal/scoring ./internal/secrets -coverprofile=/tmp/writing_coach_cover_after.out`
3. Verify total coverage is not below baseline (`53.8%`).

## Stage 1: API HTTP Structure Decomposition
Goal: Reduce `internal/api/server.go` responsibility without behavior change.

Changes:
- Extract HTTP middleware (`withCORS`, `withRecovery`, `withServerLogging`, status recorder) into `internal/api/middleware.go`.
- Extract JSON request/response helpers into `internal/api/http_json.go`.
- Extract route wiring into `internal/api/routes.go` with grouped registration helpers.

Acceptance:
- No endpoint behavior changes.
- `internal/api` tests pass unchanged.
- Coverage >= baseline.

## Stage 2: Shared LLM Prompt Input Builders
Goal: Remove provider request-shape duplication and centralize prompt input composition.

Changes:
- Add canonical user-input builder helpers in `internal/openai` for:
  - exercise generation
  - revision generation
  - submission review
- Switch `internal/gemini` and `internal/anthropic` to these shared builders.
- Keep provider transport logic separate.

Acceptance:
- Existing OpenAI/Gemini/Anthropic tests pass.
- No behavior drift in generated prompt text shape.
- Coverage >= baseline.

## Stage 3: App Composition Cleanup
Goal: Simplify startup wiring and reduce duplicate service construction.

Changes:
- In `internal/app`, construct prompt/review/curriculum services once and inject into API/CLI.
- Keep config and migration flow unchanged.
- Add/adjust focused tests if needed.

Acceptance:
- Startup behavior unchanged.
- CLI/API entrypoints remain compatible.
- Coverage >= baseline.

## Post-Stage Follow-Up (Not in this pass)
- Session-resolution write path split (read-first + explicit bootstrap path).
- HTTP server timeout hardening policy.
- SQLite concurrency and WAL tuning with benchmark-driven config.
