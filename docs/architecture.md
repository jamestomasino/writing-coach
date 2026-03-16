# Architecture Plan

## Goal

Build a local-first writing coach that supports repeated fiction exercises, stores all outputs and reviews, and adapts the next exercise based on what the writer most needs to practice.

The target style is epic tragedy in a mythopoeic mode with fantasy influences, while preserving the intellectual and emotional rigor associated with Frank Herbert and Ursula K. Le Guin.

## Product Shape

The first implementation is a CLI application written in Go with SQLite persistence. This keeps the workflow inspectable, portable, and easy to automate.

Later layers such as a TUI or small web interface should remain optional adapters over the same core services.

## Core Engines

### Prompt Engine

Produces the next exercise from:

- long-term writer profile
- curriculum state
- recent exercises and reviews
- current training focus

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

Updates the curriculum state after every submission and review.

The learning engine should track:

- skill strengths
- recurring weaknesses
- current difficulty level
- recent prompt patterns to avoid repetition
- the next exercise strategy

## Persistence Model

SQLite is the system of record. The schema should support direct reporting and preserve enough history to re-evaluate earlier work if rubrics change.

Primary tables:

- `writer_profile`
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
- `internal/cli` handles command parsing and user-facing output.

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

Expand reporting, curriculum sophistication, and optional UI layers.

## Initial Technical Decisions

- Use the Go standard library for the first CLI instead of a heavier command framework.
- Use `modernc.org/sqlite` for a pure-Go SQLite driver.
- Keep placeholder prompt/review logic deterministic until the OpenAI layer is added.
- Store raw and normalized review artifacts separately once model integration begins.

## Phase 1 Deliverable

At the end of the first implementation slice, the repo should be able to:

1. create local state with `init`
2. generate and store an exercise with `prompt next`
3. store a draft with `submit`
4. review the draft with `review`
5. show progress and recent work with `history`
