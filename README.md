# Writing Coach

`writing-coach` is a Docker-deployed writing coaching system with a Go API backend and a Next.js web client. It generates exercises, reviews submissions, and adapts future practice based on accumulated feedback.

The initial implementation targets a narrow but complete loop:

1. Compose brings up the API, auth, analyzer, and persistence stack.
2. The API generates the next exercise prompt.
3. The web client submits a story draft.
4. The API reviews the draft with deterministic analysis and model synthesis.
5. Results are persisted so later prompts can adapt to prior work.

The app now supports per-user AI provider settings in the browser. Users can connect their own `Anthropic`, `Gemini`, `OpenAI`, `Groq`, or `xAI` key under `Settings > AI provider`.

If `OPENAI_API_KEY` is set in the environment, the server also offers a shared OpenAI fallback for users who have not configured a personal provider yet. If neither a personal provider nor the shared fallback is available, generation is blocked until AI setup is completed. If a model call fails during prompt or review generation, the app falls back to deterministic local logic where supported.

## Architecture

High-level architecture and implementation phases live in [docs/architecture.md](/home/tomasino/writing-coach/docs/architecture.md).
The planned web stack, Tailwind Plus kit mapping, and screen strategy live in [docs/web-foundation-plan.md](/home/tomasino/writing-coach/docs/web-foundation-plan.md).

## Runtime Model

The supported deployment path is:

- host `nginx` terminates TLS
- host `nginx` reverse-proxies localhost-bound Docker ports
- the app, Kratos, LanguageTool, and storage stay on the internal Docker network

The Go binary still starts via `writing-coach serve` inside the app container, but the CLI workflow is no longer treated as a primary user interface.

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

Core endpoints:

- `GET /api/users`
- `POST /api/users`
- `GET /api/users/{slug}`
- `GET /api/trees`
- `POST /api/trees`
- `GET /api/trees/{slug}`
- `GET /api/trees/{slug}/versions`
- `GET /api/trees/{slug}/versions/{version}`
- `GET /api/trees/{slug}/diff?from=<n>&to=<n>`
- `POST /api/trees/{slug}/versions/{version}/restore`
- `GET /api/enrollments`
- `POST /api/enrollments`
- `GET /api/enrollments/{id}/board`
- `GET /api/health`
- `GET /api/ready`
- `GET /api/auth/session`
- `GET /api/ai/settings`
- `PUT /api/ai/settings`
- `DELETE /api/ai/settings`
- `POST /api/ai/settings/validate`
- `GET /api/onboarding`
- `POST /api/onboarding`
- `GET /api/admins`
- `POST /api/admins`
- `DELETE /api/admins/{email}`
- `GET /api/context`
- `GET /api/dashboard`
- `GET /api/exercises`
- `GET /api/exercises/{id}`
- `POST /api/prompts/next`
- `POST /api/prompts/revise`
- `GET /api/submissions`
- `POST /api/submissions`
- `GET /api/submissions/{id}`
- `GET /api/reviews`
- `POST /api/reviews`
- `GET /api/reviews/{id}`
- `GET /api/compare?submission_id=<id>[&against=<id>]`

Optional per-request context:

- query params: `user`, `tree`, `user_name`
- headers: `X-Writing-Coach-User`, `X-Writing-Coach-Tree`

Optional API auth:

- set `WRITING_COACH_API_TOKEN`
- send either `Authorization: Bearer <token>` or `X-API-Token: <token>`
- `GET /api/health` remains public for container health checks

Ory Kratos integration:

- the API will validate browser/session authentication through Kratos `GET /sessions/whoami`
- when Kratos auth is enabled, each authenticated identity maps deterministically to its own internal writer profile
- this avoids storing password hashes in `writing-coach` itself
- `GET /api/auth/session` returns the resolved auth mode, Kratos identity, onboarding state, AI readiness, and effective user/tree context for the browser client

Examples:

```bash
curl http://localhost:11234/api/dashboard
curl -X POST http://localhost:11234/api/prompts/next
curl -X POST http://localhost:11234/api/submissions \
  -H 'Content-Type: application/json' \
  -d '{"exercise_id":1,"content":"A draft goes here."}'
curl -X POST http://localhost:11234/api/reviews \
  -H 'Content-Type: application/json' \
  -d '{"submission_id":1}'
```

## Configuration

The app stores non-secret configuration in `.writing-coach/config.json`.

Environment variables:

- `OPENAI_API_KEY` optional shared OpenAI fallback
- `OPENAI_BASE_URL`
- `WRITING_COACH_AI_KEY_SECRET` required if users will save personal provider keys
- `WRITING_COACH_PROMPT_MODEL`
- `WRITING_COACH_REVIEW_MODEL`
- `WRITING_COACH_WRITER_NAME`
- `WRITING_COACH_DEFAULT_USER_SLUG`
- `WRITING_COACH_DEFAULT_TREE_SLUG`
- `WRITING_COACH_API_TOKEN`
- `WRITING_COACH_ADMIN_EMAILS`
- `COACH_PUBLIC_URL`
- `WEB_PORT_BIND`
- `KRATOS_SMTP_CONNECTION_URI`

## Deterministic Analysis

Every review now runs a built-in heuristic analyzer. In the supported Compose deployment, the stack also includes:

- Vale bundled into the app image
- LanguageTool running as an internal Docker service

These findings are passed into the review pipeline and persisted as review artifacts for later reporting and UI use. If an external analyzer is unavailable, the app continues with heuristic analysis only.

For production email delivery, set `KRATOS_SMTP_CONNECTION_URI` to your Mailgun SMTP URI. Example:

```env
KRATOS_SMTP_CONNECTION_URI=smtp://postmaster@mg.example.com:MAILGUN_SMTP_PASSWORD@smtp.mailgun.org:587/?skip_ssl_verify=false
```

The initial Vale rules live under [styles/WritingCoach](/home/tomasino/writing-coach/styles/WritingCoach) and focus on:

- stock fantasy cliches
- over-explicit symbolic explanation
- filter verbs and hedging
- abstract emotional load-bearing words
- direct emotion labels and weak modifiers

## Current Status

The repository currently contains:

- a documented architecture plan
- a Go API server behind the browser client
- a Next.js web app built from the Catalyst component kit
- a questionnaire-driven onboarding flow that generates a user-specific TGO tree
- per-user AI provider settings for Anthropic, Gemini, OpenAI, Groq, and xAI
- structured review annotations tied to quoted text and TGOs for browser-side markup
- SQLite bootstrap and schema migration support
- model-backed prompt/review services with deterministic fallback behavior
- persisted review artifacts for analyzer output, recommendation state, and revision comparisons
- DB-backed tree definitions so new curricula can be created over the API instead of only in code

## Deployment

Start the full contained stack with Docker Compose:

```bash
cp .env.example .env
$EDITOR .env
docker compose up -d --build
```

That stack includes:

- `writing-coach-web` Next.js frontend
- `writing-coach` API
- `Vale` bundled into the app image
- `LanguageTool` as a dedicated Java container
- `Ory Kratos` for account management and password handling
- `Ory Kratos` for account management and password handling
- `Mailslurper` for local verification/recovery email testing

Everything required to run the stack now lives inside `docker-compose.yml` and `.env`. No host-side LanguageTool, Vale, database, or auth service is required. The app stores its SQLite/config state under `/app/.writing-coach` in the `writing-coach-data` volume, and Kratos stores its identity SQLite DB in the `kratos-data` volume.

Production deployment for `coach.tomasino.org` should:

- copy `.env.example` to `.env`
- set the Kratos secrets to long random values
- set `WRITING_COACH_AI_KEY_SECRET` to a long random value and keep it stable
- set `WRITING_COACH_ADMIN_EMAILS` to the Kratos email addresses allowed to create or edit curricula
- decide whether `OPENAI_API_KEY` should remain set as a shared system fallback or be left empty to require BYO provider setup
- keep the single published upstream bound to `127.0.0.1`
- let host nginx terminate TLS and proxy all traffic to the web container on `WEB_PORT_BIND`

The main browser-facing setting is:

- `COACH_PUBLIC_URL`

Kratos browser URLs default from `COACH_PUBLIC_URL` and do not normally need to be set explicitly.

Deployment examples live at:

- [deploy/docker-compose.example.yml](/home/tomasino/writing-coach/deploy/docker-compose.example.yml)
- [deploy/nginx.example.conf](/home/tomasino/writing-coach/deploy/nginx.example.conf)

Kratos configuration lives at:

- [deploy/kratos/kratos.yml.tmpl](/home/tomasino/writing-coach/deploy/kratos/kratos.yml.tmpl)
- [deploy/kratos/render-config.sh](/home/tomasino/writing-coach/deploy/kratos/render-config.sh)
- [deploy/kratos/identity.schema.json](/home/tomasino/writing-coach/deploy/kratos/identity.schema.json)

Default localhost binding from `.env.example`:

- `11234` writing-coach web entrypoint

The web container proxies `/api` and `/.ory/kratos/public` internally over the Docker network, so host nginx only needs a single upstream:

```nginx
server {
    listen 80;
    server_name coach.tomasino.org;

    location / {
        proxy_pass http://127.0.0.1:11234;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Next Milestones

- replace the stopgap admin provisioning screen with a true invite flow
- replace admin-side user provisioning with a true invite flow
- add richer tree and curriculum browsing screens in the web app
