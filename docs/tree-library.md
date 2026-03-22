# Tree Library

This project ships with a built-in catalog of writing-skill trees. These trees do two jobs:

1. provide immediately usable practice paths
2. provide source material and templates for generated personalized trees

Users can also create generated practice paths from onboarding answers. Those generated trees are persisted like any other tree and enrolled like any other path.

## Built-in Trees

Current public catalog:

- `story-craft-track`
  - general fiction and story craft
  - emphasis: scene construction, character pressure, worldbuilding economy, revision flow

- `fantasy-fiction-track`
  - fantasy fiction
  - emphasis: scene control, image freshness, world pressure, character stakes

- `science-fiction-track`
  - science fiction
  - emphasis: speculative clarity, consequence, world pressure, narrative control

- `romance-fiction-track`
  - romance fiction
  - emphasis: emotional tension, scene escalation, relational clarity, payoff control

- `literary-fiction-track`
  - literary fiction
  - emphasis: scene precision, image control, emotional compression, voice

- `mystery-thriller-track`
  - mystery and thriller writing
  - emphasis: suspense control, clue handling, scene pressure, pacing

- `thought-leadership-track`
  - essays, analysis, and public-facing idea writing
  - emphasis: claim clarity, insight density, evidence integration, authority and voice

- `professional-writing-track`
  - workplace writing, memos, updates, proposals, and internal communication
  - emphasis: objective clarity, structural signposting, tone calibration, actionability, scannability

- `marketing-writing-track`
  - marketing copy and campaign writing
  - emphasis: value clarity, audience alignment, message force, conversion readiness

- `content-marketing-track`
  - editorial and content marketing
  - emphasis: reader value, structure, authority, retention, useful specificity

- `journalism-reporting-track`
  - journalism and reporting
  - emphasis: factual clarity, source handling, structure, attribution, narrative control

- `educational-writing-track`
  - educational and explanatory writing
  - emphasis: concept sequencing, clarity, scaffolding, example quality, instructional flow

- `grant-writing-track`
  - grant and proposal writing
  - emphasis: need framing, evidence, feasibility, structure, persuasion

- `academic-essay-track`
  - analytical essays and research writing
  - emphasis: thesis clarity, evidence handling, analysis depth, source integration, revision

- `technical-writing-track`
  - documentation, references, tutorials, and support content
  - emphasis: user goal alignment, step clarity, scannability, accuracy, example quality

- `persuasive-writing-track`
  - argument, advocacy, and opinion writing
  - emphasis: claim clarity, audience alignment, reasoning quality, objection handling, rhetorical force

- `memoir-personal-narrative-track`
  - memoir, personal essays, and reflective narrative
  - emphasis: scene grounding, voice presence, reflection depth, emotional compression, memory handling

## Internal Templates

These are still used as internal curriculum templates, but they are not part of the public built-in track catalog exposed by the writing-type dropdown:

- `youth-writing-foundations`

## Shared Constraints

Across the built-in catalog:

- each tree starts from exactly 3 recommended starter TGOs
- TGO codes are globally unique
- prerequisites stay explicit and tree-aware
- trees are broad enough to support generated hybrids and onboarding-based selection
- tree definitions are persisted and versioned for auditability

## Visual UML

The built-in paths also have generated UML source and rendered color diagrams for visual review:

- [Skill Path UML](./skill-paths.md)

That visual catalog is generated directly from `internal/domain/tree_catalog.go`, so the diagrams track the live tree definitions rather than a hand-maintained sketch.

## How Trees Are Used

Built-in trees currently power:

- direct enrollment in known practice paths
- onboarding-driven recommendations
- generated personalized practice paths
- reporting and progress views
- archive and assignment history grouped by enrolled path

## Current Direction

The tree system is meant to stay reusable and extensible.

That means new trees should:

- map cleanly to a writing domain or training goal
- preserve the 3-active-skill coaching model
- expose a coherent progression rather than a loose tag collection
- be safe to analyze deterministically and review against over time
