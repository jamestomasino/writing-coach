# Writing Coach

`writing-coach` is a local-first Go application for generating fiction exercises, reviewing submissions, and adapting future practice based on accumulated feedback. It now exposes the same coaching loop through both a CLI and a JSON API so a web interface can sit on top of the same core services.

The initial implementation targets a narrow but complete loop:

1. Initialize local project state.
2. Generate the next exercise prompt.
3. Submit a story draft.
4. Review the draft with deterministic analysis.
5. Persist results so later prompts can adapt to prior work.

If `OPENAI_API_KEY` is set in the environment, prompt generation and review generation will use the OpenAI Responses API with structured JSON outputs. If credentials are missing or the API call fails, the app falls back to deterministic local logic.

## Architecture

High-level architecture and implementation phases live in [docs/architecture.md](/home/tomasino/writing-coach/docs/architecture.md).
The planned web stack, Tailwind Plus kit mapping, and screen strategy live in [docs/web-foundation-plan.md](/home/tomasino/writing-coach/docs/web-foundation-plan.md).

## Planned Commands

- `writing-coach init`
- `writing-coach serve`
- `writing-coach prompt next`
- `writing-coach submit --exercise <id> --file <path>`
- `writing-coach review --submission <id>`
- `writing-coach history`

## Make Targets

- `make init`
- `make build`
- `make serve`
- `make prompt`
- `make prompt-revise SUBMISSION=<id>`
- `make submit EXERCISE=<id> FILE=<path> [REVISE_FROM=<submission-id>]`
- `make review SUBMISSION=<id>`
- `make coach-review SUBMISSION=<id>`
- `make compare SUBMISSION=<id> [AGAINST=<submission-id>]`
- `make history`
- `make progress`
- `make vale-install`
- `make languagetool-start`
- `make languagetool-stop`
- `make languagetool-status`
- `make docker-build`
- `make docker-run`

The LanguageTool targets default to `/opt/languagetool` and port `8081`. Override with `LT_HOME=/path/to/install` or `LT_PORT=8090`.
`make coach-review` will start LanguageTool if needed before running a review.
Use `REVISE_FROM=<submission-id>` on `make submit` to record a later draft as a revision of an earlier one.
Use `make compare SUBMISSION=<id>` to compare a reviewed draft against its previous reviewed draft.
Use `make prompt-revise SUBMISSION=<id>` to generate a rewrite brief from the last review of a submission.

## TGO Model

The curriculum now runs on exactly 3 active `TGOs` (Topical Guide Objectives) at a time.

- Every assignment is shaped around those 3 active TGOs.
- Reviews assess those same 3 TGOs first.
- A TGO moves to the completed list only after stable mastery.
- When one is completed, a new unlocked TGO replaces it.
- Completed TGOs remain part of the coaching context so regressions can still be noticed.

The current track begins with:

- `causal-clarity`
- `scene-architecture`
- `prose-precision`

Later TGOs unlock through prerequisites rather than a rigid straight line, so the advancement path can branch while still staying structured.

## Multi-User And Tree Model

The app is no longer conceptually single-author. The stable coaching unit is now:

- a `user`
- a `TGO tree`
- that user's enrollment and progress through that tree

This means you can keep your current advanced mythic-tragedy track while later adding a separate youth writing foundations tree for your son, or a different tree for essays, reports, or genre-specific fiction.

Current implementation status:

- user and tree records are persisted
- active and completed TGOs are enrollment-scoped
- exercises, submissions, and reviews are user- and tree-scoped
- curriculum state is enrollment-scoped

The default config still points at one user and one tree, but the API already accepts alternate `user` and `tree` slugs per request.

## JSON API

Start the backend locally with:

```bash
make serve
```

By default it listens on `:8080`. Override with:

```bash
WRITING_COACH_HTTP_ADDR=:8090 make serve
```

Core endpoints:

- `GET /api/users`
- `GET /api/users/{slug}`
- `GET /api/trees`
- `GET /api/trees/{slug}`
- `GET /api/enrollments`
- `POST /api/enrollments`
- `GET /api/enrollments/{id}/board`
- `GET /api/health`
- `GET /api/context`
- `GET /api/dashboard`
- `POST /api/prompts/next`
- `POST /api/prompts/revise`
- `POST /api/submissions`
- `GET /api/submissions/{id}`
- `POST /api/reviews`
- `GET /api/compare?submission_id=<id>[&against=<id>]`

Optional per-request context:

- query params: `user`, `tree`, `user_name`
- headers: `X-Writing-Coach-User`, `X-Writing-Coach-Tree`

Optional API auth:

- set `WRITING_COACH_API_TOKEN`
- send either `Authorization: Bearer <token>` or `X-API-Token: <token>`
- `GET /api/health` remains public for container health checks

Examples:

```bash
curl http://localhost:8080/api/dashboard
curl -X POST http://localhost:8080/api/prompts/next
curl -X POST http://localhost:8080/api/submissions \
  -H 'Content-Type: application/json' \
  -d '{"exercise_id":1,"content":"A draft goes here."}'
curl -X POST http://localhost:8080/api/reviews \
  -H 'Content-Type: application/json' \
  -d '{"submission_id":1}'
```

## Configuration

The app stores non-secret configuration in `.writing-coach/config.json`.

Environment variables:

- `OPENAI_API_KEY`
- `OPENAI_BASE_URL`
- `WRITING_COACH_PROMPT_MODEL`
- `WRITING_COACH_REVIEW_MODEL`
- `WRITING_COACH_HTTP_ADDR`
- `WRITING_COACH_API_TOKEN`
- `VALE_BINARY`
- `LANGUAGETOOL_URL`

## Deterministic Analysis

Every review now runs a built-in heuristic analyzer. Optional external analyzers can be enabled:

- Vale: install `vale`; the app will auto-detect it and use the in-repo [.vale.ini](/home/tomasino/writing-coach/.vale.ini). `make vale-install` installs a repo-local binary at `.writing-coach/bin/vale`. Set `VALE_BINARY` to override the executable path.
- LanguageTool: set `LANGUAGETOOL_URL` to a running server such as `http://localhost:8010`

These findings are passed into the review pipeline and surfaced in CLI output. If external analyzers are unavailable, the app continues with heuristic analysis only.

The initial Vale rules live under [styles/WritingCoach](/home/tomasino/writing-coach/styles/WritingCoach) and focus on:

- stock fantasy cliches
- over-explicit symbolic explanation
- filter verbs and hedging
- abstract emotional load-bearing words
- direct emotion labels and weak modifiers

## Current Status

The repository currently contains:

- a documented architecture plan
- a Go CLI plus JSON API server
- SQLite bootstrap and schema migration support
- model-backed prompt/review services with deterministic fallback behavior

## Deployment

Build and run the container locally with:

```bash
make docker-build
make docker-run API_TOKEN=change-me
```

The container serves the API on port `8080` and stores its SQLite/config state under `/app/.writing-coach`, which `make docker-run` mounts from the repo. If `API_TOKEN` is set, the API requires that bearer token for every endpoint except `/api/health`.

Deployment examples live at:

- [deploy/docker-compose.example.yml](/home/tomasino/writing-coach/deploy/docker-compose.example.yml)
- [deploy/nginx.example.conf](/home/tomasino/writing-coach/deploy/nginx.example.conf)

An nginx reverse proxy can sit in front of it with a simple upstream:

```nginx
server {
    listen 80;
    server_name writing.example.com;

    location / {
        proxy_pass http://127.0.0.1:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Next Milestones

- add custom tree authoring and editing, not just built-in trees
- make prompt and review generation adapt more strongly to each tree's pedagogy
- store richer review artifacts for later audit and calibration
- build the first web UI against the stabilized API
