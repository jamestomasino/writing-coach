# Writing Coach

`writing-coach` is a Docker-deployed writing practice app with:

- a Go API backend
- a Next.js web client
- Ory Kratos for authentication
- SQLite persistence
- deterministic analysis plus LLM-backed prompt and review generation

The app is built around a closed feedback loop:

1. set a writing track through onboarding
2. generate an assignment
3. submit a draft
4. receive feedback and scores
5. revise or move on
6. adapt future assignments based on accumulated history

## What It Does

- keeps exactly 3 active `TGOs` in focus at a time
- generates prompts and revision briefs
- reviews drafts against active skills
- preserves assignment history across prompt, draft, feedback, revision, and later passes
- supports per-user AI provider settings
- falls back to deterministic local logic where supported if a model call fails

Supported personal AI providers:

- Anthropic
- Gemini
- OpenAI
- Groq
- xAI

## Deployment Model

The intended deployment is a single Docker Compose stack behind host `nginx`:

- host `nginx` terminates TLS
- host `nginx` reverse-proxies a localhost-bound web port
- the web app, API, Kratos, LanguageTool, and storage stay on the internal Docker network

This repository contains mixed-license materials. The web UI includes Tailwind Plus-derived code. See:

- [LICENSE](/home/tomasino/writing-coach/LICENSE)
- [NOTICE.md](/home/tomasino/writing-coach/NOTICE.md)
- [docs/licensing.md](/home/tomasino/writing-coach/docs/licensing.md)
- [web/LICENSE.md](/home/tomasino/writing-coach/web/LICENSE.md)

## Quick Start

```bash
cp .env.example .env
$EDITOR .env
docker compose up -d --build
```

Then open the public URL you configured in `COACH_PUBLIC_URL`.

Default localhost binding from `.env.example`:

- `127.0.0.1:11234:3000`

The web container proxies `/api` and `/.ory/kratos/public` internally, so host `nginx` only needs one upstream:

```nginx
server {
    listen 80;
    server_name coach.example.com;

    location / {
        proxy_pass http://127.0.0.1:11234;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

## Required Setup

At minimum, set these in `.env` before production use:

- `COACH_PUBLIC_URL`
- `KRATOS_COOKIE_SECRET`
- `KRATOS_CIPHER_SECRET`
- `KRATOS_UI_COOKIE_SECRET`
- `KRATOS_UI_CSRF_SECRET`
- `WRITING_COACH_ADMIN_EMAILS`

If users will save their own provider keys, also set:

- `WRITING_COACH_AI_KEY_SECRET`

Generate secrets with:

```bash
openssl rand -base64 48
```

Keep `WRITING_COACH_AI_KEY_SECRET` stable after deployment. Changing it later will make previously saved user keys unreadable.

## AI Provider Modes

The app supports two operating modes:

### 1. Shared fallback enabled

Set:

```env
OPENAI_API_KEY=...
WRITING_COACH_AI_KEY_SECRET=...
```

Behavior:

- the app can use a shared system OpenAI key
- users can optionally save their own provider keys
- existing users can keep working without immediate setup

### 2. Bring-your-own-provider only

Set:

```env
OPENAI_API_KEY=
WRITING_COACH_AI_KEY_SECRET=...
```

Behavior:

- there is no shared system fallback
- users must save their own provider key to run model-backed generation

If `WRITING_COACH_AI_KEY_SECRET` is missing, personal provider storage is unavailable.

## Important Environment Variables

- `OPENAI_API_KEY`
  Optional shared OpenAI fallback.
- `OPENAI_BASE_URL`
  Optional custom base URL for the shared OpenAI-compatible provider.
- `WRITING_COACH_AI_KEY_SECRET`
  Required for storing users’ personal provider keys.
- `WRITING_COACH_PROMPT_MODEL`
  Default shared prompt model.
- `WRITING_COACH_REVIEW_MODEL`
  Default shared review model.
- `WRITING_COACH_AI_VALIDATE_LIMIT_PER_MINUTE`
  Per-user cap for provider validation attempts.
- `WRITING_COACH_AI_VALIDATE_GLOBAL_LIMIT_PER_MINUTE`
  App-wide cap for provider validation attempts.
- `WRITING_COACH_AI_PROVIDER_EVENT_RETENTION_DAYS`
  Retention window for admin-visible provider activity events.
- `WRITING_COACH_WRITER_NAME`
- `WRITING_COACH_DEFAULT_USER_SLUG`
- `WRITING_COACH_DEFAULT_TREE_SLUG`
- `WRITING_COACH_API_TOKEN`
- `WRITING_COACH_ADMIN_EMAILS`
- `COACH_PUBLIC_URL`
- `WEB_PORT_BIND`
- `KRATOS_SMTP_CONNECTION_URI`

## Production Notes

- keep the published upstream bound to localhost
- let host `nginx` terminate TLS
- keep Compose volumes persistent
- replace all default Kratos secrets
- decide whether `OPENAI_API_KEY` stays as a transition fallback or is removed

State is stored in Docker volumes:

- `writing-coach-data` for app data and SQLite
- `kratos-data` for Kratos identity storage

## AI Validation Hardening

- `Validate connection` and `Save provider` use the same validation budget
- the default per-user cap is `6` checks per minute
- the default global cap is `60` checks per minute
- repeated bad-key retries eventually return `429` without hitting the upstream provider
- provider activity events are retained for `30` days by default
- admin users can inspect provider activity in the admin workspace

## Deterministic Analysis

Every review runs built-in heuristic analysis. In the Compose deployment, the stack also includes:

- Vale bundled into the app image
- LanguageTool as an internal Docker service

These findings feed the review pipeline and are saved as review artifacts for later reporting and UI use.

If an external analyzer is unavailable, the app continues with heuristic analysis only.

The initial Vale rules live under [styles/WritingCoach](/home/tomasino/writing-coach/styles/WritingCoach).

## API Overview

Core browser-facing and app endpoints include:

- `GET /api/health`
- `GET /api/ready`
- `GET /api/auth/session`
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
- `GET /api/ai/settings`
- `PUT /api/ai/settings`
- `DELETE /api/ai/settings`
- `POST /api/ai/settings/validate`

There are also admin, tree, user, onboarding, and enrollment endpoints.

Optional per-request context:

- query params: `user`, `tree`, `user_name`
- headers: `X-Writing-Coach-User`, `X-Writing-Coach-Tree`

Optional API auth:

- set `WRITING_COACH_API_TOKEN`
- send `Authorization: Bearer <token>` or `X-API-Token: <token>`

## Architecture Notes

Useful docs:

- [docs/architecture.md](/home/tomasino/writing-coach/docs/architecture.md)
- [docs/ai-provider-rollout-plan.md](/home/tomasino/writing-coach/docs/ai-provider-rollout-plan.md)
- [docs/licensing.md](/home/tomasino/writing-coach/docs/licensing.md)
- [docs/web-foundation-plan.md](/home/tomasino/writing-coach/docs/web-foundation-plan.md)
- [docs/web-app-plan.md](/home/tomasino/writing-coach/docs/web-app-plan.md)
- [docs/tree-library.md](/home/tomasino/writing-coach/docs/tree-library.md)

## Current Status

The repository currently includes:

- a Go API server
- a Next.js web app
- questionnaire-driven onboarding
- assignment timeline and archive browsing
- per-user AI provider settings
- admin-visible AI provider activity reporting
- structured review annotations and comparison artifacts
- SQLite migrations and bootstrap support
- DB-backed trees and enrollment-scoped progress

## Deployment References

- [deploy/docker-compose.example.yml](/home/tomasino/writing-coach/deploy/docker-compose.example.yml)
- [deploy/nginx.example.conf](/home/tomasino/writing-coach/deploy/nginx.example.conf)
- [deploy/kratos/kratos.yml.tmpl](/home/tomasino/writing-coach/deploy/kratos/kratos.yml.tmpl)
- [deploy/kratos/render-config.sh](/home/tomasino/writing-coach/deploy/kratos/render-config.sh)
- [deploy/kratos/identity.schema.json](/home/tomasino/writing-coach/deploy/kratos/identity.schema.json)
