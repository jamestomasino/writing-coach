import { buildObjectiveExamples } from './objective-example-library'
import { getSkillDetailByName } from './skill-details'
import type { SkillGraphNode } from './types'

export type GeneratedObjectiveDetail = {
  skillOverview: string
  objectiveGoal: string
  whyThisObjective: string
  successLooksLike: string[]
  goodExample: string
  badExample: string
  revisionMoves: string[]
  assessmentFocus: string[]
  exampleSources: Array<{ label: string; url: string }>
  exampleStrategy: string
  studentReadinessCheck: string
}

function sentence(value: string) {
  const trimmed = value.trim()
  if (!trimmed) {
    return ''
  }
  return /[.!?]$/.test(trimmed) ? trimmed : `${trimmed}.`
}

function sanitizeProse(value: string) {
  return value.replace(/\breally\b/gi, '').replace(/\bvery\b/gi, '').replace(/\s{2,}/g, ' ').trim()
}

function stageLabel(stage: string) {
  const normalized = (stage || '').toLowerCase()
  if (normalized.includes('foundation') || normalized.includes('core')) {
    return 'foundational control'
  }
  if (normalized.includes('revision')) {
    return 'revision durability'
  }
  if (normalized.includes('advanced')) {
    return 'advanced transfer'
  }
  return `${stage} stage control`
}

function clipWords(value: string, maxWords: number) {
  const words = value.trim().split(/\s+/).filter(Boolean)
  if (words.length <= maxWords) {
    return value.trim()
  }
  return `${words.slice(0, maxWords).join(' ')}`
}

function objectiveGoal(node: SkillGraphNode) {
  const rawDescription = node.description || `Practice ${node.title.toLowerCase()} in a visible, repeatable way`
  const description = sentence(sanitizeProse(rawDescription))
  const mastery = node.mastery_hint ? sentence(`Target mastery marker: ${node.mastery_hint}`) : ''
  return `${description}${mastery ? ` ${mastery}` : ''}`.trim()
}

function whyThisObjective(node: SkillGraphNode) {
  const stage = stageLabel(node.stage)
  const base = sentence(`This objective builds ${stage} for ${node.title.toLowerCase()}`)
  const mastery = node.mastery_hint
    ? sentence(`When this holds: ${clipWords(node.mastery_hint, 14)}`)
    : sentence('When this holds, readers can follow the draft without reconstructing intent')
  return `${base} ${mastery}`.trim()
}

function successLooksLike(node: SkillGraphNode) {
  const title = node.title.toLowerCase()
  const first = `The draft shows clear, repeated evidence of ${title} in multiple sections.`
  const second = node.mastery_hint
    ? sentence(`Observable marker: ${node.mastery_hint}`)
    : 'Observable marker: readers do not need to guess what changed or why it matters.'
  const third = `Control of ${title} stays visible in revision, not only in a single pass.`
  return [first, second, third]
}

function revisionMoves(node: SkillGraphNode) {
  const title = node.title.toLowerCase()
  const description = sanitizeProse(node.description.toLowerCase())
  return [
    `Mark one paragraph where ${title} is weakest and annotate the exact breakdown in clarity or logic.`,
    `Rewrite that section so it satisfies this objective description: ${description}.`,
    `Add one verification line using the mastery marker before submitting the next revision.`,
  ]
}

function assessmentFocus(node: SkillGraphNode) {
  const title = node.title.toLowerCase()
  return [
    `Evidence for ${title} should be observable in the draft, not only implied by intent.`,
    'Reader orientation should improve after revision rather than stay flat.',
    'Objective control should hold under pressure across sections and subsequent drafts.',
  ]
}

function skillOverview(node: SkillGraphNode) {
  const skill = (node.skill_name ?? '').trim()
  const detail = skill ? getSkillDetailByName(skill) : undefined
  if (!detail) {
    return sentence(`This objective belongs to ${skill || 'an unmapped skill family'} and trains visible control in real assignments`)
  }
  return sentence(`${detail.oneLine} ${detail.whatItMeans} ${detail.whyItMatters}`)
}

function studentReadinessCheck(node: SkillGraphNode) {
  return sentence(
    `Student check: can you explain how you will apply ${node.title.toLowerCase()} in one paragraph of your assignment, and name one revision move you will run first`
  )
}

export function buildGeneratedObjectiveDetail(node: SkillGraphNode): GeneratedObjectiveDetail {
  const examples = buildObjectiveExamples(node)
  return {
    skillOverview: skillOverview(node),
    objectiveGoal: objectiveGoal(node),
    whyThisObjective: whyThisObjective(node),
    successLooksLike: successLooksLike(node),
    goodExample: examples.good,
    badExample: examples.needsWork,
    revisionMoves: revisionMoves(node),
    assessmentFocus: assessmentFocus(node),
    exampleSources: examples.sources,
    exampleStrategy: examples.strategy,
    studentReadinessCheck: studentReadinessCheck(node),
  }
}
