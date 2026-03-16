# Sample Tree Library

The application now ships with a deeper reference catalog of writing curricula. These are meant to do two jobs:

1. act as immediately usable built-in tracks
2. serve as source material for future in-app tree generation and authoring

The runtime model is now shifting toward a single global skill graph. These reference trees remain important, but they now function primarily as authored regions inside that larger graph.

## Current reference trees

- `mythic-tragedy-apprenticeship`
  - advanced mythopoeic tragic fiction
  - 55 TGOs
  - emphasis: tragic inevitability, symbolic control, mythic tone, scene architecture

- `story-craft-track`
  - general fiction/storytelling progression
  - 52 TGOs
  - emphasis: scene construction, character pressure, worldbuilding economy, revision flow

- `youth-writing-foundations`
  - younger writers and early foundations
  - 52 TGOs
  - emphasis: word choice, sentence clarity, paragraph control, story sequencing, revision habits

- `thought-leadership-track`
  - essays, analysis, and public-facing idea writing
  - 52 TGOs
  - emphasis: claim clarity, insight density, evidence integration, authority and voice

- `professional-writing-track`
  - workplace writing, memos, updates, proposals, and internal communication
  - 52 TGOs
  - emphasis: objective clarity, structural signposting, tone calibration, actionability, scannability

- `academic-essay-track`
  - analytical essays, research papers, and close reading
  - 52 TGOs
  - emphasis: thesis clarity, evidence handling, analysis depth, source integration, revision

- `technical-writing-track`
  - documentation, references, tutorials, and support content
  - 52 TGOs
  - emphasis: user goal alignment, step clarity, scannability, accuracy, example quality

- `persuasive-writing-track`
  - argument, advocacy, opinion, and rhetorical persuasion
  - 52 TGOs
  - emphasis: claim clarity, audience alignment, reasoning quality, objection handling, rhetorical force

- `memoir-personal-narrative-track`
  - memoir, personal essays, and reflective narrative
  - 52 TGOs
  - emphasis: scene grounding, voice presence, reflection depth, emotional compression, memory handling

## Design constraints

- Each tree keeps exactly 3 seed TGOs.
- Each tree now has 50+ TGOs.
- TGO codes are globally unique across the built-in catalog.
- Prerequisites stay inside the tree they belong to.
- The current catalog is deliberately broad enough to support future generated hybrids and branch selection.

## Intended next use

These trees are the baseline library for:

- onboarding-driven starter recommendations into the global graph
- future generated personalized graph routes
- eventual in-app tree authoring
- admin-side curriculum editing and branching
