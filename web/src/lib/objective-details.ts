import type { SkillGraphNode } from './types'
import type { SkillDetail } from './skill-details'

export type ObjectiveDetail = {
  code: string
  title: string
  skillFamily: string
  objectiveGoal: string
  whyThisObjective: string
  successLooksLike: string[]
  goodExample: string
  badExample: string
  revisionMoves: string[]
  assessmentFocus: string[]
}

function sentence(value: string, fallback: string) {
  const trimmed = value.trim()
  if (!trimmed) {
    return fallback
  }
  return /[.!?]$/.test(trimmed) ? trimmed : `${trimmed}.`
}

function assessmentFocusForCode(code: string) {
  const normalized = code.toLowerCase()
  const out: string[] = []

  if (normalized.includes('causal') || normalized.includes('thread') || normalized.includes('sequence')) {
    out.push('Cause and effect should be easy to follow from one beat to the next.')
  }
  if (normalized.includes('point-of-view') || normalized.includes('perspective')) {
    out.push('Point of view should stay stable unless the shift is intentional and clearly signaled.')
  }
  if (normalized.includes('time') || normalized.includes('linearity') || normalized.includes('braid')) {
    out.push('Time movement should stay legible so readers know where they are.')
  }
  if (normalized.includes('backstory') || normalized.includes('exposition') || normalized.includes('summary')) {
    out.push('Context should support present pressure, not interrupt or stall it.')
  }
  if (normalized.includes('dialogue')) {
    out.push('Dialogue should reveal pressure, conflict, or change, not filler talk.')
  }
  if (out.length === 0) {
    out.push('The objective pattern should appear clearly in the draft, not just in intent.')
  }
  out.push('Control should hold across revisions and assignments, not only one pass.')
  return out
}

function revisionMovesForObjective(node: SkillGraphNode, family?: SkillDetail) {
  const normalized = node.code.toLowerCase()
  const moves: string[] = []
  if (normalized.includes('causal') || normalized.includes('thread')) {
    moves.push('Add one clear "because/so" link at each major turn.')
  }
  if (normalized.includes('point-of-view')) {
    moves.push('Mark each paragraph by who is perceiving the moment, and remove accidental drift.')
  }
  if (normalized.includes('time') || normalized.includes('linearity') || normalized.includes('braid')) {
    moves.push('Add time anchors where the draft jumps in time.')
  }
  if (normalized.includes('backstory') || normalized.includes('exposition')) {
    moves.push('Cut one backstory block and keep only details that change the present scene.')
  }
  if (moves.length < 2) {
    moves.push('Pick one paragraph where this objective is weakest and rewrite it with concrete action.')
  }
  if (moves.length < 3 && family?.revisionMoves?.length) {
    for (const item of family.revisionMoves) {
      if (moves.length >= 3) {
        break
      }
      if (!moves.includes(item)) {
        moves.push(item)
      }
    }
  }
  if (moves.length < 3) {
    moves.push('Submit another revision and check whether this objective evidence gets stronger.')
  }
  return moves.slice(0, 3)
}

export function buildObjectiveDetail(node: SkillGraphNode, family?: SkillDetail): ObjectiveDetail {
  const skillFamily = (node.skill_name ?? '').trim() || 'unmapped skill family'
  const objectiveGoal = sentence(node.description, `Build reliable control of ${node.title.toLowerCase()}.`)
  const whyThisObjective = sentence(
    node.mastery_hint ?? '',
    `When ${node.title.toLowerCase()} is steady, readers can follow your writing with less effort.`
  )

  const successLooksLike = [
    sentence(`The draft shows clear evidence of ${node.title.toLowerCase()}`, 'The draft shows clear objective evidence.'),
    sentence(`The pattern repeats in more than one part of the draft`, 'The pattern is repeatable, not isolated.'),
    sentence(`The same control holds in later drafts`, 'Control holds across revisions.'),
  ]

  const familyStrong = family?.strongExample?.replace(/^Strong:\s*/i, '').trim()
  const familyWeak = family?.weakExample?.replace(/^Weak:\s*/i, '').trim()

  const goodExample = `Good: ${familyStrong ?? `The writing clearly demonstrates ${node.title.toLowerCase()} in a way the reader can track.`}`
  const badExample = `Needs work: ${familyWeak ?? `The writing gestures at ${node.title.toLowerCase()}, but readers have to guess what changed and why.`}`

  return {
    code: node.code,
    title: node.title,
    skillFamily,
    objectiveGoal,
    whyThisObjective,
    successLooksLike,
    goodExample,
    badExample,
    revisionMoves: revisionMovesForObjective(node, family),
    assessmentFocus: assessmentFocusForCode(node.code),
  }
}
