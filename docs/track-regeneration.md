# Track Regeneration Plan

This project now validates built-in trees as DAGs with valid seeds and reachable nodes.

Planarity is now enforced across the shipped built-in catalog. This document remains as a record of the regeneration strategy that got the trees there and the taxonomy issues that drove the refactor.

## Current Audit

From the pre-regeneration built-in catalog:

- Non-planar trees:
  - `thought-leadership-track`
  - `professional-writing-track`
  - `technical-writing-track`
  - `marketing-writing-track`
  - `journalism-reporting-track`
  - `grant-writing-track`
  - `global-writing-skill-graph`

- Planar trees:
  - `youth-writing-foundations`
  - `story-craft-track`
  - `academic-essay-track`
  - `persuasive-writing-track`
  - `memoir-personal-narrative-track`
  - `fantasy-fiction-track`
  - `science-fiction-track`
  - `romance-fiction-track`
  - `literary-fiction-track`
  - `mystery-thriller-track`
  - `content-marketing-track`
  - `educational-writing-track`

- Skills used by the catalog but not represented in `SupportedSkills`:
  - `accuracy`
  - `analysis depth`
  - `assignment alignment`
  - `example quality`
  - `grammar control`
  - `objection handling`
  - `professional format`
  - `reasoning quality`
  - `reflection depth`
  - `revision habits`
  - `rhetorical force`
  - `source handling`
  - `spelling and mechanics`
  - `story development`
  - `structure and pacing`
  - `technical precision`
  - `thesis clarity`
  - `user goal alignment`
  - `voice presence`

## What The Audit Means

The problems are not all the same.

- `thought-leadership-track`, `professional-writing-track`, and `technical-writing-track` are base nonfiction graphs whose internal dependency pattern is already non-planar.
- `marketing-writing-track`, `journalism-reporting-track`, and `grant-writing-track` inherit non-planarity from those base trees because they are cloned variants rather than independently shaped curricula.
- Some planar tracks still have taxonomy issues because their priority skills are more specific than the current shared skill list.

This suggests two different jobs:

1. reshape graphs so prerequisite structure stays planar
2. redesign the skill taxonomy so track missions are expressed cleanly instead of being squeezed into an outdated list

## Regeneration Rules

If we want future tracks to remain both DAGs and planar, generation needs harder structural rules.

- Keep exactly 3 seed nodes.
- Keep every node to at most 2 prerequisites unless there is a strong reason otherwise.
- Prefer local progression:
  - a node should usually depend on one direct predecessor in its strand
  - optional second prerequisite can come from one neighboring strand
- Avoid repeated braid patterns where several late nodes each depend on different pairs from multiple strands.
- Treat the curriculum as 3 to 5 parallel strands that occasionally merge, rather than a free-form mesh.
- Keep branch merges sparse and intentional near major stage transitions.
- Add validation at generation time for:
  - DAG
  - seed reachability
  - planarity
  - maximum prerequisite count

In practice, this means generating a layered curriculum skeleton first, then assigning skills and node content to that skeleton, instead of writing node prerequisites ad hoc.

## Skill Taxonomy Direction

The current catalog already implies a better taxonomy than `SupportedSkills`.

We should move toward a canonical skill inventory grouped roughly like this:

- Core clarity:
  - `clarity and coherence`
  - `sentence economy`
  - `sentence variety`
  - `paragraph control`
  - `scannability`
  - `structural signposting`

- Evidence and reasoning:
  - `claim clarity`
  - `evidence integration`
  - `reasoning quality`
  - `analysis depth`
  - `source handling`
  - `thesis clarity`

- Audience and rhetoric:
  - `audience alignment`
  - `tone calibration`
  - `authority and voice`
  - `rhetorical force`
  - `objection handling`
  - `actionability`

- Narrative and fiction craft:
  - `narrative clarity`
  - `scene architecture`
  - `story development`
  - `emotional compression`
  - `dialogue intelligence`
  - `image freshness`
  - `worldbuilding economy`
  - `structure and pacing`
  - `voice presence`
  - `reflection depth`

- Technical and professional:
  - `user goal alignment`
  - `technical precision`
  - `accuracy`
  - `example quality`
  - `professional format`

- Revision and control:
  - `revision habits`
  - `grammar control`
  - `spelling and mechanics`
  - `assignment alignment`

This gives us a more honest skill surface for onboarding, reporting, and future generated tracks.

## Recommended Refactor Order

1. Replace `SupportedSkills` with a canonical inventory derived from the real catalog.
2. Freeze clone-based expansion for non-planar base trees.
3. Regenerate bespoke planar versions of:
   - `thought-leadership-track`
   - `professional-writing-track`
   - `technical-writing-track`
4. Rebuild dependent clone tracks on top of those new planar bases:
   - `marketing-writing-track`
   - `journalism-reporting-track`
   - `grant-writing-track`
5. Decide whether any retired internal templates should be regenerated or removed.
6. Decide whether the global graph should remain a union graph that may be non-planar, or become a curated planar meta-track.

## Practical Next Step

Do not start by hand-editing a few edges in the existing non-planar trees.

Start by choosing one base track, preferably `thought-leadership-track`, and regenerate it from a planar skeleton with:

- 3 seeds
- 3 or 4 strands
- <= 2 prerequisites per node
- explicit mission skills at the strand level

Once that base works, use the same skeleton strategy for the other nonfiction bases and only then rebuild their specialized variants.
