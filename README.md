# Writing Coach

`writing-coach` is a local-first Go CLI for generating fiction exercises, reviewing submissions, and adapting future practice based on accumulated feedback.

The initial implementation targets a narrow but complete loop:

1. Initialize local project state.
2. Generate the next exercise prompt.
3. Submit a story draft.
4. Review the draft with deterministic analysis.
5. Persist results so later prompts can adapt to prior work.

If `OPENAI_API_KEY` is set in the environment, prompt generation and review generation will use the OpenAI Responses API with structured JSON outputs. If credentials are missing or the API call fails, the app falls back to deterministic local logic.

## Architecture

High-level architecture and implementation phases live in [docs/architecture.md](/home/tomasino/writing-coach/docs/architecture.md).

## Planned Commands

- `writing-coach init`
- `writing-coach prompt next`
- `writing-coach submit --exercise <id> --file <path>`
- `writing-coach review --submission <id>`
- `writing-coach history`

## Make Targets

- `make init`
- `make build`
- `make prompt`
- `make submit EXERCISE=<id> FILE=<path>`
- `make review SUBMISSION=<id>`
- `make coach-review SUBMISSION=<id>`
- `make history`
- `make progress`
- `make vale-install`
- `make languagetool-start`
- `make languagetool-stop`
- `make languagetool-status`

The LanguageTool targets default to `/opt/languagetool` and port `8081`. Override with `LT_HOME=/path/to/install` or `LT_PORT=8090`.
`make coach-review` will start LanguageTool if needed before running a review.

## Configuration

The app stores non-secret configuration in `.writing-coach/config.json`.

Environment variables:

- `OPENAI_API_KEY`
- `OPENAI_BASE_URL`
- `WRITING_COACH_PROMPT_MODEL`
- `WRITING_COACH_REVIEW_MODEL`
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
- a Go CLI scaffold
- SQLite bootstrap and schema migration support
- model-backed prompt/review services with deterministic fallback behavior

## Next Milestones

- integrate the OpenAI API with structured JSON outputs
- add Vale and LanguageTool analyzers
- strengthen curriculum updates from review results
- add richer progress reporting and export
