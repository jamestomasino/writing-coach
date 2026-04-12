import type { SkillGraphNode } from './types'

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

function objectiveGoalForCode(code: string, title: string, description: string) {
  const normalized = code.toLowerCase()
  if (normalized.includes('causal') || normalized.includes('thread') || normalized.includes('sequence')) {
    return sentence(
      `Make action and consequence legible beat by beat. Show the choice and result close together so readers can track the chain.`,
      `Build reliable control of ${title.toLowerCase()}.`
    )
  }
  if (normalized.includes('point-of-view') || normalized.includes('perspective')) {
    return sentence(
      `Keep perspective stable so readers always know whose eyes they are in. Signal intentional shifts early and clearly.`,
      `Build reliable control of ${title.toLowerCase()}.`
    )
  }
  if (normalized.includes('time') || normalized.includes('linearity') || normalized.includes('braid')) {
    return sentence(
      `Make time movement easy to follow. Place clear time anchors where the draft jumps or folds the timeline.`,
      `Build reliable control of ${title.toLowerCase()}.`
    )
  }
  if (normalized.includes('backstory') || normalized.includes('exposition') || normalized.includes('summary')) {
    return sentence(
      `Use context only when it changes what is happening now. Keep backstory short and tied to present pressure.`,
      `Build reliable control of ${title.toLowerCase()}.`
    )
  }
  if (normalized.includes('dialogue')) {
    return sentence(
      `Write dialogue that reveals tension, motive, or change. Cut filler lines that do not move the moment.`,
      `Build reliable control of ${title.toLowerCase()}.`
    )
  }
  return sentence(
    description,
    `Practice ${title.toLowerCase()} in clear, repeatable moves so the pattern is visible in the draft.`
  )
}

function whyNowForCode(code: string, title: string, masteryHint: string) {
  const normalized = code.toLowerCase()
  if (normalized.includes('causal') || normalized.includes('thread') || normalized.includes('sequence')) {
    return sentence(
      `Readers can follow each decision and its immediate consequence without backtracking. Clear causality keeps attention on meaning, not confusion.`,
      `When ${title.toLowerCase()} is steady, readers can follow your writing with less effort.`
    )
  }
  if (normalized.includes('point-of-view') || normalized.includes('perspective')) {
    return sentence(
      `Stable perspective keeps emotion and stakes clear in real time. Readers trust the narrative when viewpoint drift is controlled.`,
      `When ${title.toLowerCase()} is steady, readers can follow your writing with less effort.`
    )
  }
  if (normalized.includes('time') || normalized.includes('linearity') || normalized.includes('braid')) {
    return sentence(
      `Clear timing reduces reader load and keeps momentum high. Readers can focus on significance instead of reconstructing order.`,
      `When ${title.toLowerCase()} is steady, readers can follow your writing with less effort.`
    )
  }
  if (normalized.includes('backstory') || normalized.includes('exposition') || normalized.includes('summary')) {
    return sentence(
      `Tight context keeps scenes active and emotionally present. Readers stay with the current decision instead of drifting into summary.`,
      `When ${title.toLowerCase()} is steady, readers can follow your writing with less effort.`
    )
  }
  if (normalized.includes('dialogue')) {
    return sentence(
      `Focused dialogue creates pace and reveals character under pressure. Readers hear intent, conflict, and change in each exchange.`,
      `When ${title.toLowerCase()} is steady, readers can follow your writing with less effort.`
    )
  }
  return sentence(
    masteryHint,
    `This objective improves readability now and makes later objectives easier to transfer across assignments.`
  )
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

function revisionMovesForObjective(node: SkillGraphNode) {
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
  if (moves.length < 3) {
    moves.push('Submit another revision and check whether this objective evidence gets stronger.')
  }
  return moves.slice(0, 3)
}

export function buildObjectiveDetail(node: SkillGraphNode): ObjectiveDetail {
  const skillFamily = (node.skill_name ?? '').trim() || 'unmapped objective cluster'
  const objectiveGoal = objectiveGoalForCode(node.code, node.title, node.description)
  const whyThisObjective = whyNowForCode(node.code, node.title, node.mastery_hint ?? '')

  const successLooksLike = [
    sentence(`The draft shows clear evidence of ${node.title.toLowerCase()}`, 'The draft shows clear objective evidence.'),
    sentence(`The pattern repeats in more than one part of the draft`, 'The pattern is repeatable, not isolated.'),
    sentence(`The same control holds in later drafts`, 'Control holds across revisions.'),
  ]

  const goodExample = `Good: The writing clearly demonstrates ${node.title.toLowerCase()} in a way the reader can track without pausing to decode intent.`
  const badExample = `Needs work: The writing gestures at ${node.title.toLowerCase()}, but the reader has to guess what changed, when it changed, or why it matters.`

  return {
    code: node.code,
    title: node.title,
    skillFamily,
    objectiveGoal,
    whyThisObjective,
    successLooksLike,
    goodExample,
    badExample,
    revisionMoves: revisionMovesForObjective(node),
    assessmentFocus: assessmentFocusForCode(node.code),
  }
}
