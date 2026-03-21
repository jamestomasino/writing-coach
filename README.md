# Writing Coach

`writing-coach` is a Docker-deployed writing practice app for structured skill building.

It combines:

- a Go API backend
- a Next.js web client
- Ory Kratos for authentication
- SQLite persistence
- deterministic analysis plus LLM-backed prompt and review generation
- assignment history, revision compare, and archive browsing

The core loop is simple:

1. choose a practice path
2. generate an assignment
3. submit a draft
4. get feedback tied to a few active skills
5. revise or move on
6. use that history to shape the next assignment

## What Users Get

- exactly 3 active skill goals at a time
- new assignments and revision briefs
- reviews tied to the current skill goals
- assignment history across prompt, draft, feedback, revision, and later passes
- per-user AI provider settings
- deterministic fallback behavior when model-backed generation is unavailable

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
- the web app, API, Kratos, analyzers, and storage stay on the internal Docker network

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

    proxy_intercept_errors on;
    error_page 502 503 504 /maintenance.html;

    location = /maintenance.html {
        root /var/www/writing-coach;
        internal;
    }

    location / {
        proxy_pass http://127.0.0.1:11234;
        proxy_set_header Host $host;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

For safer outages, copy [maintenance.html](/home/tomasino/writing-coach/deploy/maintenance.html) to a host path such as `/var/www/writing-coach/maintenance.html`. With `proxy_intercept_errors on`, `nginx` will serve that static page whenever the upstream returns `502`, `503`, or `504`.

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

## How Feedback Works

Reviews are intentionally layered.

1. Deterministic analyzers inspect the draft for concrete issues.
2. The active skill goals decide what matters most on this assignment.
3. The review and revision flow use those signals to shape feedback and the next step.

That means the app does not simply send a draft to a model and accept the result uncritically.

The deterministic layer includes:

- built-in heuristic checks
- `Vale` for style and custom prose rules
- `LanguageTool` for grammar and usage suggestions
- `spaCy` plus `TextDescriptives` for sentence and readability analysis

If a language model is enabled, it works on top of that structure. It helps write clearer assignments, coaching summaries, and revision briefs. It does not replace the deterministic analysis layer.

If no model is available, the app still produces deterministic prompts, reviews, and revision briefs.

The app also tracks a separate `writing language` on each practice path. That is distinct from the UI locale. English is the only shipped coaching language right now, but the analyzer and model pipeline are wired so contributors can add future language support without redesigning the app.

Contributor reference:

- [docs/coaching-language-contributors.md](/home/tomasino/writing-coach/docs/coaching-language-contributors.md)

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
- `WRITING_COACH_NLP_ANALYZER_URL`
  Optional internal spaCy/TextDescriptives analyzer service URL.
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
- let host `nginx` serve a static maintenance page for `502`, `503`, and `504`
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
- spaCy plus TextDescriptives as an internal Docker service

These findings are stored as review artifacts for later reporting and UI use.

If an external analyzer is unavailable, the app continues with the remaining analyzers.

The initial Vale rules live under [styles/WritingCoach](/home/tomasino/writing-coach/styles/WritingCoach).

## Web Experience

The browser UI is built around the assignment loop:

- current assignment workspace as the default home
- background review and revision queue states with progress loaders
- full assignment timelines showing prompt, draft, feedback, and revision steps
- archive browsing for older assignments
- an About page that explains the coaching loop in plain language

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
- [docs/release-process.md](/home/tomasino/writing-coach/docs/release-process.md)
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
