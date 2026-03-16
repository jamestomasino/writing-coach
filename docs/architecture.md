# Architecture Plan

## Goal

Build a local-first writing coach that supports repeated fiction exercises, stores all outputs and reviews, and adapts the next exercise based on what the writer most needs to practice.

The target style is epic tragedy in a mythopoeic mode with fantasy influences, while preserving the intellectual and emotional rigor associated with Frank Herbert and Ursula K. Le Guin.

## Product Shape

The first implementation is a Go application with SQLite persistence and two adapters:

- a CLI for direct local use
- a JSON API for a future web interface

The web layer should remain an adapter over the same prompt, review, curriculum, and persistence services rather than a separate system.

## Core Engines

### Prompt Engine

Produces the next exercise from:

- per-user curriculum state
- active TGOs for the enrolled tree
- recent exercises and reviews
- current training pressure

It should balance:

- one new challenge at a time
- reinforcement of one weak skill
- variety in prompt form
- continuity with the writer's stated aesthetic goals

### Review Engine

Combines deterministic analysis with model-based critique.

Deterministic analysis should cover:

- grammar and spelling
- sentence-length variance
- word count and paragraph shape
- repetition and overused constructions
- dialogue vs narration balance

Model-based review should eventually cover:

- thematic depth
- tragic inevitability
- mythic tone
- symbolic control
- clarity vs ornamentation
- specific next-step advice

### Learning Engine

Updates the enrollment-scoped curriculum state after every submission and review.

The learning engine should track:

- skill strengths
- recurring weaknesses
- current difficulty level
- recent prompt patterns to avoid repetition
- the next exercise strategy

## TGO Curriculum

The coaching model is built around `Topical Guide Objectives (TGOs)`.

Rules:

- exactly 3 TGOs are active at any given time
- every assignment must declare those 3 TGOs explicitly
- review prioritizes those 3 TGOs over general craft commentary
- stable mastery moves a TGO to the completed list
- a newly unlocked TGO replaces the completed one

Advancement is structured but not strictly linear. TGOs carry prerequisites so later genre-specific work only unlocks after enough command of the core storytelling foundations.

The architecture now treats a `TGO tree` as the reusable pedagogical unit. Each user progresses through a chosen tree independently.

This enables:

- one advanced mythic-tragedy tree for your own fiction work
- a youth foundations tree for your son
- later trees for essays, scene craft, fantasy foundations, or other writing goals

## Persistence Model

SQLite is the system of record. The schema should support direct reporting and preserve enough history to re-evaluate earlier work if rubrics change.

Primary tables:

- `users`
- `tgo_trees`
- `tree_tgos`
- `user_tree_enrollments`
- `user_curriculum_state`
- `enrollment_active_tgos`
- `enrollment_completed_tgos`
- `exercises`
- `submissions`
- `reviews`
- `skill_dimensions`
- `submission_skill_scores`
- `curriculum_state`
- `events`

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
```

### Boundary Rules

- `internal/domain` holds durable business types.
- `internal/db` owns persistence and migrations.
- `internal/prompt`, `internal/review`, and `internal/curriculum` expose narrow service interfaces.
- `internal/app` wires concrete implementations.
- `internal/api` exposes JSON endpoints for a browser-based UI or remote deployment.
- `internal/cli` handles command parsing and user-facing output.

## Deployment Posture

The API can run:

- locally with no auth during development
- behind nginx on a server
- in Docker with SQLite mounted from the host

An optional API token gate protects all endpoints except health checks. That is intentionally minimal and should be treated as the first deployment layer, not the final public auth model.

## Interfaces

The service layer should converge on interfaces shaped like:

```go
type PromptGenerator interface {
    NextExercise(ctx context.Context, state CurriculumState) (Exercise, error)
}

type Reviewer interface {
    ReviewSubmission(ctx context.Context, sub Submission, ctxData ReviewContext) (Review, error)
}

type SkillUpdater interface {
    Update(ctx context.Context, sub Submission, review Review) (CurriculumState, error)
}
```

## Implementation Phases

### Phase 1

Deliver one complete local loop:

- initialize config and database
- generate a prompt
- accept a submission
- generate a deterministic review
- update curriculum state

### Phase 2

Integrate OpenAI with structured JSON outputs for prompt generation and review synthesis.

Implementation notes:

- call the Responses API directly from Go over `net/http`
- require schema-conformant JSON for both exercises and reviews
- preserve deterministic local services as the fallback path
- keep provider choice transparent in CLI output and stored exercise/review metadata

### Phase 3

Add open source analyzers:

- Vale for style and custom prose rules
- LanguageTool for grammar and style suggestions

Implementation status:

- built-in heuristic analyzer is active
- Vale integration is available via CLI invocation when configured
- LanguageTool integration is available via local HTTP server when configured
- repo-local Vale rules encode project-specific style guidance for mythic tragic fantasy

### Phase 4

Expand reporting, tree management, and deployment hardening while keeping the API stable for the first browser UI layer.

## Initial Technical Decisions

- Use the Go standard library for the CLI and HTTP API instead of heavier frameworks.
- Use `modernc.org/sqlite` for a pure-Go SQLite driver.
- Keep placeholder prompt/review logic deterministic until the OpenAI layer is added.
- Store raw and normalized review artifacts separately once model integration begins.
- Keep deployment simple: one API process, SQLite on disk, Docker packaging, nginx reverse proxy in front.

## Phase 1 Deliverable

At the end of the first implementation slice, the repo should be able to:

1. create local state with `init`
2. generate and store an exercise with `prompt next`
3. store a draft with `submit`
4. review the draft with `review`
5. show progress and recent work with `history`
