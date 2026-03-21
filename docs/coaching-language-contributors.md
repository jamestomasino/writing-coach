# Coaching Language Contributor Guide

This app now separates:

- UI locale
- writing language

Those are different settings.

A user can browse the app in one language and still practice writing in another.

## Current State

The app currently ships one coaching language:

- `en` (`English`)

That means:

- onboarding stores a `writing_language`
- the onboarding UI exposes that choice even though English is the only shipped option today
- LLM requests receive that language explicitly
- deterministic analyzers run only when that language is supported
- unsupported languages fail closed with warnings instead of silently applying English rules

## Where Language Lives

### Profile storage

The writing language is stored on the onboarding profile:

- [onboarding.go](/home/tomasino/writing-coach/internal/domain/onboarding.go)
- [profile_settings_store.go](/home/tomasino/writing-coach/internal/db/profile_settings_store.go)
- [onboarding_handlers.go](/home/tomasino/writing-coach/internal/api/onboarding_handlers.go)

### Shared language rules

Normalization and support checks live in:

- [writing_language.go](/home/tomasino/writing-coach/internal/domain/writing_language.go)

### Analyzer gating

Deterministic analyzer language gating lives in:

- [language.go](/home/tomasino/writing-coach/internal/analyzer/language.go)

Current behavior:

- `LanguageTool` only runs when the writing language maps to a supported LanguageTool code
- `Vale` only runs when a supported language pack exists
- the spaCy/TextDescriptives sidecar only runs when the current writing language is supported
- heuristic analysis also fails closed for unsupported languages
- deterministic fallback review language also admits unsupported-language limits instead of pretending full coverage

### LLM requests

Prompt and review requests explicitly include the writing language:

- [types.go](/home/tomasino/writing-coach/internal/llm/types.go)
- [client.go](/home/tomasino/writing-coach/internal/openai/client.go)
- [client.go](/home/tomasino/writing-coach/internal/anthropic/client.go)
- [client.go](/home/tomasino/writing-coach/internal/gemini/client.go)

## How To Add A New Coaching Language

### 1. Register the language

Update:

- [writing_language.go](/home/tomasino/writing-coach/internal/domain/writing_language.go)

Add:

- normalization aliases
- display label
- support flag
- LanguageTool mapping if available

### 2. Expose it in onboarding

Add the option in:

- [onboarding.go](/home/tomasino/writing-coach/internal/domain/onboarding.go)

If you want the language to be selectable in the browser UI, also confirm the onboarding form copy still explains the support level clearly.

### 3. Decide deterministic support analyzer by analyzer

Do not assume one analyzer implies another.

For each language, decide:

- `LanguageTool`: supported language code available?
- `Vale`: real language-specific styles available?
- `spaCy/TextDescriptives`: model and metrics reliable enough?
- `heuristic`: still valid for sentence and paragraph heuristics?

If a tool is not reliable for that language, keep it disabled.

### 4. Add language-specific Vale packs

Do not reuse the English Vale rules for another language.

Add a parallel style family under `styles/`, then update:

- [vale.go](/home/tomasino/writing-coach/internal/analyzer/vale.go)

### 5. Add NLP sidecar support

If you add a new spaCy pipeline or another NLP backend, update:

- [app.py](/home/tomasino/writing-coach/docker/nlp-analyzer/app.py)

Be explicit about:

- model choice
- supported metrics
- unsupported metrics that should be skipped

### 6. Review the deterministic fallback copy

If you want no-LLM coaching to work well in that language, update:

- [service.go](/home/tomasino/writing-coach/internal/prompt/service.go)
- [service.go](/home/tomasino/writing-coach/internal/review/service.go)

Right now those fallbacks are still English-oriented.

### 7. Verify the LLM path

LLM prompts already receive `writing_language`, but contributors should test:

- assignment generation in the target language
- revision brief generation in the target language
- review generation in the target language
- annotation quote handling
- behavior when no model is configured, so deterministic fallback messaging stays honest

## Contributor Rules

- Never run English deterministic rules on another language unless you can justify the quality.
- Prefer explicit unsupported warnings over low-confidence feedback.
- Keep UI locale and writing language separate.
- Add tests when you add a language mapping or analyzer behavior.

## Suggested Wiki Link

If the project wiki gets a contributor section, link this page as:

- `Adding a coaching language`
