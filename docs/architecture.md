# Architecture

## Product Shape

`writing-coach` is a Go application with a Next.js web client. The web app is the only published browser entrypoint. It proxies API and Kratos browser traffic internally across the Docker network.

The product loop is:

1. choose or generate a practice path
2. generate an assignment
3. submit a draft
4. review the draft against the current active skills
5. revise or move on
6. use that history to shape the next assignment

The current product is no longer fiction-only. It supports multiple writing domains through built-in and generated skill trees, while still preserving the same assignment and review loop.

## Core Systems

### Practice paths and skills

The coaching model is built around `Topical Guide Objectives (TGOs)`, surfaced to users as skills.

Rules:

- each enrollment progresses through one active tree at a time
- exactly 3 TGOs are active at any given time
- assignments and reviews are scoped to those 3 active skills
- completed skills remain visible for maintenance and regression checks
- unlocking is prerequisite-based rather than strictly linear

The app ships with a built-in tree catalog and also supports generated personalized trees created from onboarding answers.

### Assignment generation

Assignment generation combines:

- enrollment-scoped curriculum state
- active TGOs
- onboarding profile
- recurring weaknesses and analyzer findings
- optional model-backed generation

If an LLM is available, it generates assignments and revision briefs under a structured schema. If not, the app falls back to deterministic generation.

### Review generation

Reviews are layered:

1. deterministic analyzers inspect the submission
2. active TGOs narrow which findings matter most
3. an optional model-backed reviewer turns that context into coaching language
4. deterministic fallback remains available when no model is enabled

The deterministic analyzer stack currently includes:

- built-in heuristics
- Vale
- LanguageTool
- spaCy plus TextDescriptives

Vale, heuristics, the NLP sidecar, and deterministic fallback review language are domain-aware. The coaching pipeline is also now writing-language-aware, though English is the only shipped coaching language today.

Coverage and ownership standards for deterministic analyzers are defined in [analyzer-coverage-model.md](/home/tomasino/Sites/personal/writing-coach/docs/analyzer-coverage-model.md).

### Curriculum updates

The curriculum service updates enrollment-scoped progress after each review. It tracks:

- active and completed TGOs
- recurring weaknesses
- recurring analyzer findings
- review history
- next-focus recommendations

### Localization and writing language

The app separates:

- UI locale
- writing language

That distinction matters because a user might browse the UI in one language while practicing writing in another. UI localization is prepared through `next-intl`. Coaching language support is prepared separately through onboarding profile data, analyzer gating, and explicit LLM request fields.

## Package Boundaries

```text
cmd/writing-coach/
internal/app/
internal/api/
internal/cli/
internal/config/
internal/curriculum/
internal/db/
internal/domain/
internal/prompt/
internal/review/
migrations/
web/
docker/
```

Boundary rules:

- `internal/domain` holds durable business types and tree/profile logic
- `internal/db` owns persistence, queries, and migrations
- `internal/prompt`, `internal/review`, and `internal/curriculum` provide the main coaching services
- `internal/app` wires concrete runtime dependencies
- `internal/api` exposes JSON endpoints for the web app and admin flows
- `internal/cli` exposes local/operator commands
- `web/` contains the Next.js client
- `docker/` contains sidecar services such as LanguageTool and the NLP analyzer

## Persistence Model

SQLite is the system of record.

Key persisted entities include:

- users and authenticated sessions through Kratos integration
- tree definitions and tree versions
- user tree enrollments
- active and completed TGOs per enrollment
- onboarding profiles
- exercises
- submissions
- reviews
- review artifacts
- skill scores
- review jobs
- AI provider settings and activity events

The schema is designed to keep assignment, review, and curriculum history inspectable over time rather than only storing current state.

## Deployment Posture

The intended deployment is Docker Compose behind host `nginx`.

Current production posture:

- host `nginx` terminates TLS
- host `nginx` can serve a static maintenance page when the upstream is unavailable
- the published web upstream stays bound to localhost
- the web app proxies API and Kratos browser routes internally
- Ory Kratos handles identity, recovery, verification, and sessions
- SQLite is stored on a persistent volume
- LanguageTool runs as a sidecar service
- spaCy plus TextDescriptives run as a sidecar service
- Vale is bundled into the app image

The intended topology is:

`host nginx -> localhost-bound web upstream -> internal Docker services`

## Current Implementation Notes

- the web app includes onboarding, current assignment, review, comparison, progress, archive, timeline, AI settings, admin, about, and auth flows
- review and revision generation run through background jobs and queue-state polling where needed
- assignment archive pages treat old assignments as historical records rather than active workflow stages
- admin currently focuses on account access and AI provider activity, not user provisioning
- personal AI provider settings are stored encrypted at rest when configured

## Design Direction

The app is structured to keep three properties stable:

- analyzer-first review rather than model-only review
- skill-scoped coaching rather than generic feedback
- contributor-extensible paths for new trees, locales, and future coaching languages
