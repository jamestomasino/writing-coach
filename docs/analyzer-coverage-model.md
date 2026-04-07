# Analyzer Coverage Model

This document defines app-wide deterministic analyzer coverage rules. It applies to both assignment reviews and playground reviews.

Goals:

- provide baseline coverage across all writing styles in the app
- avoid duplicate checks across heuristics, Vale, LanguageTool, and NLP
- enforce applicability gates so misaligned checks do not run
- keep every deterministic finding measurable, valid, and explainable

## Taxonomy Layers

Use three layers of categories:

- `global`: checks that apply broadly across prose writing
- `domain`: checks that apply to writing families (fiction, technical, academic, professional, marketing, thought leadership)
- `specialty`: checks that apply to narrow formats (memo, landing page, tutorial, grant narrative, poem, etc.)

Precedence:

1. specialty
2. domain
3. global

Specialty gates can disable domain/global checks when they conflict with writing goals.

## Ownership and Non-Duplication

Each category and rule has a single primary owner:

- `heuristic`: app-specific deterministic rules not well-covered elsewhere
- `vale`: style-guide and wording policy checks
- `languagetool`: grammar, spelling, punctuation, agreement
- `nlp`: discourse and semantic pattern checks

Rule:

- Do not add a heuristic if a third-party tool already provides stable signal for the same goal.
- If overlap is unavoidable, mark heuristic as fallback-only and suppress it when owner signal is available.

### Ownership Baseline

Global categories:

- `clarity`: heuristic + nlp
- `structure`: heuristic + nlp
- `readability`: heuristic + nlp
- `mechanics`: languagetool
- `style policy`: vale

Domain categories:

- `narrative progression` (fiction): heuristic + nlp
- `instructional completeness` (technical): heuristic + nlp
- `argument support` (academic/thought leadership): nlp + heuristic fallback
- `actionability` (professional): heuristic
- `message hierarchy` (marketing): heuristic + nlp

Specialty categories:

- `memo execution` (memo): heuristic
- `cta architecture` (landing page): heuristic
- `grant compliance framing` (grant): heuristic
- `poetic craft proxies` (poetry): heuristic specialty rules only

## Applicability Contract

Every deterministic rule must define:

- metric source keys used as evidence
- writing domains where it applies
- specialties/formats where it applies
- explicit exclusions
- minimum content requirements (for example min words)
- threshold mapping (note/warning/error)

Rules must support hard skip conditions. Example:

- dialogue checks must skip poetry and non-narrative formats
- scannability checks for heading/list structure should skip short narrative scenes

## Rule Registry Requirements

The analyzer layer should maintain a typed registry with:

- `rule_id`
- `layer`: global/domain/specialty
- `owner`: heuristic/vale/languagetool/nlp
- `category`
- `purpose` (validity target)
- `metric_keys`
- `applies_when`
- `skip_when`
- `thresholds`
- `fallback_policy`

The registry is the source of truth for:

- check eligibility
- UI labeling and provenance
- scoring/rubric wiring
- regression tests for applicability

## Initial Coverage Standard

Minimum global coverage across all writing styles:

- sentence length distribution
- paragraph segmentation
- length adequacy by format
- grammar/spelling/punctuation signal
- style policy violations
- discourse-level coherence proxy (NLP)

Minimum domain overlays:

- fiction/fantasy: causality, scene progression, dialogue balance
- technical/educational: procedural completeness, scannability, prerequisite/result framing
- academic: claim-support density and reasoning continuity
- professional: owner/action/deadline clarity and decision framing
- marketing/content marketing: value proposition, support, CTA clarity
- thought leadership/journalism/persuasive: thesis continuity, evidence cadence, attribution framing

Minimum specialty overlays:

- memo, landing page, grant, tutorial, poem

## Validation Standard

A deterministic rule is acceptable only if it is:

- measurable: deterministic scalar or count
- valid: directly maps to writing goal
- aligned: applicable to selected context
- interpretable: can explain evidence in plain language

## Test Strategy

Add three classes of tests:

- registry completeness: expected categories exist per domain and global baseline
- applicability guards: misaligned contexts are skipped (for example dialogue on poetry)
- ownership guards: heuristics do not duplicate owner categories unless marked fallback-only

## Rollout Plan

Phase 1:

- define and ship registry metadata for current heuristic rules
- annotate ownership and applicability

Phase 2:

- route heuristic execution through registry gates
- remove or fallback-gate overlapping checks

Phase 3:

- expand domain/specialty rule coverage to baseline standard
- add per-domain validation fixtures

Phase 4:

- expose category provenance in UI for debug transparency
- tune thresholds using stored review artifact analysis
