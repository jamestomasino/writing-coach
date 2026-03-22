# Track Expansion Roadmap

This document turns the curriculum review into an implementation sequence.

The goal is not to create as many public tracks as possible. The goal is to make the catalog:

- pedagogically complete enough to feel credible
- structurally reusable across multiple enrolled tracks
- safe to regenerate as DAGs and, where possible, planar graphs

## Current Baseline

The current public catalog already has:

- 17 public tracks
- 52 TGOs per public track
- a shared skill lattice with strong overlap across tracks
- a hard DAG validator for built-in trees

The next step is not broad expansion first. It is disciplined expansion.

## Recommended Public Additions

Add these four missing public tracks first:

1. `poetry-track`
2. `scriptwriting-track`
3. `speechwriting-track`
4. `ux-writing-track`

These are the highest-value additions because each introduces constraints that are not adequately represented by the current prose-heavy catalog.

## Shared Skill Additions

The taxonomy now includes domain skills needed to support those tracks without inventing bespoke vocabularies later:

- poetry:
  - `lineation`
  - `sonic patterning`
  - `image logic`
  - `stanza movement`
- scriptwriting:
  - `visual exposition`
  - `beat design`
  - `act structure`
- speechwriting:
  - `oral cadence`
  - `rhetorical repetition`
  - `audience energy`
- UX writing:
  - `microcopy clarity`
  - `error-state guidance`
  - `information scent`

These should remain domain skills, not specialty skills. They are reusable across adjacent tracks and elective arcs.

## Track Shape Rules

Every new or regenerated public track should follow these rules:

- exactly 3 seeds
- exactly 2 core seeds and 1 domain seed
- 40-60 TGOs for the main path
- most TGOs mapped to shared core or domain skills
- no more than 0-3 true specialty skills per track
- prerequisites must remain a DAG
- prefer planar skeletons during regeneration rather than trying to patch planarity after the fact

## Specialty Arc Strategy

Not every curricular need should become a new top-level track.

Prefer specialty arcs for:

- investigative journalism
- longform essay
- conversion copy
- startup / executive communication
- nonprofit fundraising strategy
- personal essay

Recommended arc size:

- 8-15 TGOs
- built mostly from existing shared skills
- 1-2 new domain or specialty skills only when necessary

## First Build Sequence

Use this order:

1. Regenerate `thought-leadership-track`
2. Add `speechwriting-track`
3. Regenerate `technical-writing-track`
4. Add `ux-writing-track`
5. Add `poetry-track`
6. Add `scriptwriting-track`

Why this order:

- `thought-leadership-track` is already upstream of other nonfiction paths and should become a cleaner reusable base
- `speechwriting-track` can reuse thought-leadership foundations while introducing real oral-rhetoric differences
- `technical-writing-track` is a strong base for UX writing and currently needs structural cleanup anyway
- `ux-writing-track` can then become a narrow but distinct domain track with high skill transfer
- `poetry` and `scriptwriting` are important, but they are more isolated and can land after the nonfiction base is stronger

## Seed Recommendations

Recommended seed profiles for the first four additions/regenerations:

`thought-leadership-track`

- core: `claim clarity`
- core: `audience alignment`
- domain: `sentence economy`

`speechwriting-track`

- core: `claim clarity`
- core: `voice presence`
- domain: `oral cadence`

`technical-writing-track`

- core: `clarity and coherence`
- core: `scannability`
- domain: `user goal alignment`

`ux-writing-track`

- core: `clarity and coherence`
- core: `audience alignment`
- domain: `microcopy clarity`

## Mission Guidance By New Domain

`poetry-track`

- mission: compression, image, sound, and line-based movement
- likely shared base:
  - `clarity and coherence`
  - `voice presence`
  - `image freshness`
- likely domain emphasis:
  - `lineation`
  - `sonic patterning`
  - `image logic`
  - `stanza movement`

`scriptwriting-track`

- mission: dramatic movement through visible action, beat logic, and scene turns
- likely shared base:
  - `scene architecture`
  - `dialogue intelligence`
  - `narrative clarity`
- likely domain emphasis:
  - `visual exposition`
  - `beat design`
  - `act structure`

`speechwriting-track`

- mission: writing for listening, memory, repetition, and live audience movement
- likely shared base:
  - `claim clarity`
  - `audience alignment`
  - `voice presence`
- likely domain emphasis:
  - `oral cadence`
  - `rhetorical repetition`
  - `audience energy`

`ux-writing-track`

- mission: writing under interface constraints to help users complete tasks with low friction
- likely shared base:
  - `clarity and coherence`
  - `actionability`
  - `audience alignment`
- likely domain emphasis:
  - `microcopy clarity`
  - `error-state guidance`
  - `information scent`
  - `user goal alignment`

## Regeneration Standard For Existing Tracks

When rebuilding an existing non-planar track:

- do not preserve every legacy edge
- preserve mission, seeds, and the broad skill mix
- rebuild from a clean planar skeleton
- keep progression easier to explain than the current mesh
- prefer fewer prerequisites per node
- preserve seed migration safety for existing users

## Success Criteria

The next phase is working when:

- the shared taxonomy supports new domains without bespoke drift
- regenerated tracks remain clearly aligned to their writing mission
- users can finish a main path and still have adjacent options
- cross-track transfer feels meaningful
- the catalog grows by a few genuinely distinct domains rather than by fragmentation
