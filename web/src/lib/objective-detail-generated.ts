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
  return description
}

function whyThisObjective(node: SkillGraphNode) {
  const skill = (node.skill_name ?? '').toLowerCase()
  const stage = stageLabel(node.stage)
  const stageHint = stage.includes('revision') ? 'during revision' : 'while drafting'

  if (skill.includes('narrative') || skill.includes('story') || skill.includes('scene')) {
    return `${sentence(
      `This matters now because narrative writing loses momentum when readers cannot track cause, pressure, and consequence ${stageHint}`
    )} ${sentence(
      'When this control is strong, scenes feel intentional, stakes escalate cleanly, and readers stay oriented inside the story instead of decoding what happened'
    )}`.trim()
  }
  if (skill.includes('claim') || skill.includes('thesis') || skill.includes('analysis') || skill.includes('reasoning')) {
    return `${sentence(
      `This matters now because analytical writing persuades only when the central position and supporting logic are explicit ${stageHint}`
    )} ${sentence(
      'When this control is weak, readers may understand your topic but still reject your conclusion because the reasoning path feels incomplete'
    )}`.trim()
  }
  if (skill.includes('evidence') || skill.includes('source')) {
    return `${sentence(
      `This matters now because readers trust evidence only when support is relevant, framed, and connected to a clear claim ${stageHint}`
    )} ${sentence(
      'Strong control here prevents source dumping, improves credibility, and makes it obvious why each citation belongs in the paragraph'
    )}`.trim()
  }
  if (skill.includes('tone') || skill.includes('voice') || skill.includes('audience')) {
    return `${sentence(
      `This matters now because tone and voice shape whether readers trust your intent and stay engaged ${stageHint}`
    )} ${sentence(
      'When audience fit is off, even accurate points can sound evasive, inflated, or dismissive, which weakens the writing outcome'
    )}`.trim()
  }
  if (skill.includes('actionability') || skill.includes('scannability') || skill.includes('professional format')) {
    return `${sentence(
      `This matters now because practical writing succeeds only when readers can find key information quickly and act on it ${stageHint}`
    )} ${sentence(
      'Without this control, readers spend time searching for owners, decisions, or next steps, and execution quality drops even when the content is correct'
    )}`.trim()
  }
  if (skill.includes('grammar') || skill.includes('spelling') || skill.includes('mechanics') || skill.includes('precision')) {
    return `${sentence(
      `This matters now because sentence-level control protects clarity, precision, and credibility ${stageHint}`
    )} ${sentence(
      'Mechanical noise and vague wording force readers to re-interpret basic meaning, which distracts from your argument and weakens trust'
    )}`.trim()
  }

  return `${sentence(
    `This matters now because missing control in this area makes writing harder to follow ${stageHint}`
  )} ${sentence(
    'When this objective is stable, readers spend cognitive effort on your ideas rather than on reconstructing sentence-level intent'
  )}`.trim()
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
