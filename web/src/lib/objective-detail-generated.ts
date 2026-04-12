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
  return value
    .replace(/\breally\b/gi, '')
    .replace(/\bvery\b/gi, '')
    .replace(/\bkind of\b/gi, '')
    .replace(/\s{2,}/g, ' ')
    .trim()
}

function objectiveGoal(node: SkillGraphNode) {
  const rawDescription = node.description || `Practice ${node.title.toLowerCase()} in a visible, repeatable way`
  const description = sentence(sanitizeProse(rawDescription))
  switch (patternKey(node)) {
    case 'causal':
      return `${description} ${sentence(
        'Make each major turn legible by placing the trigger, action, and consequence close enough for one-pass reading'
      )}`.trim()
    case 'scene':
      return `${description} ${sentence(
        'Build each scene around a concrete objective, resistance, and changed state so the section cannot end where it began'
      )}`.trim()
    case 'dialogue':
      return `${description} ${sentence(
        'Use exchanges to move pressure or leverage, not to recap information readers already have'
      )}`.trim()
    case 'pacing':
      return `${description} ${sentence(
        'Control beat size and transition speed so tension rises deliberately instead of flattening or spiking randomly'
      )}`.trim()
    case 'thesis':
      return `${description} ${sentence(
        'State a bounded, arguable position early and keep each section tied to that single line of claim'
      )}`.trim()
    case 'evidence':
      return `${description} ${sentence(
        'Select support that is proportionate to claim importance and explain the inference from evidence to conclusion'
      )}`.trim()
    case 'source':
      return `${description} ${sentence(
        'Frame each source for role and reliability, then synthesize across sources instead of stacking summaries'
      )}`.trim()
    case 'voice':
      return `${description} ${sentence(
        'Keep stance and register consistent with audience expectations while preserving directness and authority'
      )}`.trim()
    case 'action':
      return `${description} ${sentence(
        'Specify who does what by when, and make the requested outcome easy to execute without follow-up clarification'
      )}`.trim()
    case 'scan':
      return `${description} ${sentence(
        'Surface decisions, deadlines, and key evidence where a scanning reader can find them in seconds'
      )}`.trim()
    case 'structure':
      return `${description} ${sentence(
        'Organize sections so each paragraph has one job and each transition makes the next step predictable'
      )}`.trim()
    case 'analysis':
      return `${description} ${sentence(
        'Show the reasoning chain from observation to interpretation and acknowledge competing readings where relevant'
      )}`.trim()
    case 'mechanics':
      return `${description} ${sentence(
        'Keep syntax, punctuation, and word choice precise enough that readers do not need to stop and re-parse sentences'
      )}`.trim()
    case 'revision':
      return `${description} ${sentence(
        'Run revision passes in a deliberate order so structural issues are solved before sentence-level polish'
      )}`.trim()
    case 'technical':
      return `${description} ${sentence(
        'Preserve factual and procedural accuracy so a reader can execute or verify the content without hidden assumptions'
      )}`.trim()
    default:
      return `${description} ${sentence(
        'Translate this into observable text choices so another reader can point to where it works and where it breaks'
      )}`.trim()
  }
}

function patternKey(node: SkillGraphNode) {
  const bag = `${node.code} ${node.title} ${node.description} ${node.skill_name ?? ''}`.toLowerCase()
  if (/(scene architecture|entrance|exit|reversal|status dynamics|spatial|blocking|beat design|scene turn|scene objective|scene pressure)/.test(bag)) return 'scene'
  if (/(causal|cause-and-effect|consequence|trigger|domino|cause chain|effect chain)/.test(bag)) return 'causal'
  if (/(dialogue|subtext|speaker|silence|voice difference)/.test(bag)) return 'dialogue'
  if (/(tension|pacing|arc|escalation|middle management|ending payoff|opening grip)/.test(bag)) return 'pacing'
  if (/(thesis|claim clarity|thesis clarity|stakes articulation|scope control)/.test(bag)) return 'thesis'
  if (/(evidence integration|example quality|data interpretation|quantification|evidence proportion)/.test(bag)) return 'evidence'
  if (/(source handling|citation|quote|paraphrase|synthesis|literature|source selection|source evaluation)/.test(bag)) return 'source'
  if (/(tone|voice|audience|authority|rhetorical|cadence|platform fit)/.test(bag)) return 'voice'
  if (/(actionability|ask visibility|deadline|ownership|step|instruction|decision language|result visibility)/.test(bag)) return 'action'
  if (/(scannability|heading|list|table|annotation|layout|summary and detail|front-loaded)/.test(bag)) return 'scan'
  if (/(sentence|paragraph|transition|signposting|structure basics|section order|pivot)/.test(bag)) return 'structure'
  if (/(analysis|reasoning|counterreading|pattern recognition|concept bridging|insight)/.test(bag)) return 'analysis'
  if (/(grammar|spelling|mechanics|tense|comma|agreement|precision|word choice)/.test(bag)) return 'mechanics'
  if (
    /(technical|accuracy|correctness|factual|terminology|units|calculation|specification|method|procedure|constraint|compliance|safety|risk)/.test(
      bag
    )
  )
    return 'technical'
  if (/(revision|proofreading|pass|triage|self-check|line editing)/.test(bag)) return 'revision'
  return 'general'
}

function whyThisObjective(node: SkillGraphNode) {
  switch (patternKey(node)) {
    case 'causal':
      return `${sentence(
        'This matters because readers disengage when outcomes feel unearned or disconnected from prior choices'
      )} ${sentence(
        'Strong causal control makes tension feel inevitable and keeps momentum moving forward instead of sideways'
      )}`.trim()
    case 'scene':
      return `${sentence(
        'This matters because a scene without clear turns or spatial logic feels static even when characters are speaking'
      )} ${sentence(
        'Strong scene structure creates pressure, change, and orientation at the same time'
      )}`.trim()
    case 'dialogue':
      return `${sentence(
        'This matters because dialogue should move power, intent, or conflict, not simply relay information'
      )} ${sentence(
        'When dialogue carries dramatic work, scenes become sharper and characters feel distinct'
      )}`.trim()
    case 'pacing':
      return `${sentence(
        'This matters because pacing controls whether readers feel rising pressure or repetitive drag'
      )} ${sentence(
        'Strong pacing keeps important beats visible and prevents emotional flatlines between turning points'
      )}`.trim()
    case 'thesis':
      return `${sentence(
        'This matters because readers cannot evaluate or trust an argument until the central position is specific and bounded'
      )} ${sentence(
        'Strong thesis control keeps every section aligned to one defensible line of reasoning'
      )}`.trim()
    case 'evidence':
      return `${sentence(
        'This matters because claims are persuasive only when support is relevant, proportional, and interpreted'
      )} ${sentence(
        'Strong evidence handling turns facts into argument rather than leaving them as disconnected data'
      )}`.trim()
    case 'source':
      return `${sentence(
        'This matters because source quality and integration determine whether your argument feels credible or borrowed'
      )} ${sentence(
        'Strong source control shows what each source contributes and why your interpretation remains primary'
      )}`.trim()
    case 'voice':
      return `${sentence(
        'This matters because tone and voice determine whether readers trust the message before they assess the evidence'
      )} ${sentence(
        'Strong audience and voice control keeps writing credible, direct, and context-appropriate'
      )}`.trim()
    case 'action':
      return `${sentence(
        'This matters because writing fails operationally when readers cannot tell what action is required, by whom, and by when'
      )} ${sentence(
        'Strong action clarity reduces execution errors and decision delays'
      )}`.trim()
    case 'scan':
      return `${sentence(
        'This matters because readers often scan first and decide in seconds whether a document is usable'
      )} ${sentence(
        'Strong scannability puts critical information where it can be found quickly under time pressure'
      )}`.trim()
    case 'structure':
      return `${sentence(
        'This matters because structure determines whether readers can follow progression without reconstructing links on their own'
      )} ${sentence(
        'Strong structural control makes complex material readable in one pass'
      )}`.trim()
    case 'analysis':
      return `${sentence(
        'This matters because analysis quality is the difference between summary and insight'
      )} ${sentence(
        'Strong analytical control shows how evidence leads to interpretation and where limits still apply'
      )}`.trim()
    case 'mechanics':
      return `${sentence(
        'This matters because sentence-level errors and vague wording break trust and slow comprehension'
      )} ${sentence(
        'Strong mechanics and precision keep attention on ideas instead of avoidable friction'
      )}`.trim()
    case 'revision':
      return `${sentence(
        'This matters because revision quality determines whether drafts improve in substance or only in surface polish'
      )} ${sentence(
        'Strong revision habits focus effort on high-leverage issues first and produce visible gains'
      )}`.trim()
    case 'technical':
      return `${sentence(
        'This matters because minor factual or procedural errors can make writing unusable even when the prose reads smoothly'
      )} ${sentence(
        'Strong technical accuracy reduces rework, protects credibility, and prevents downstream execution mistakes'
      )}`.trim()
    default:
      return `${sentence(
        'This matters because unclear execution forces readers to infer missing logic, sequence, or significance'
      )} ${sentence(
        'Reliable execution makes purpose, consequence, and next action visible on first read'
      )}`.trim()
  }
}

function successLooksLike(node: SkillGraphNode) {
  switch (patternKey(node)) {
    case 'causal':
      return [
        'Reader test: A reader can trace each major outcome back to a specific prior decision or trigger.',
        'Text signal: Cause -> consequence links are explicit at scene turns, not implied after the fact.',
        'Failure signal: Readers describe key outcomes as sudden, unearned, or disconnected.',
      ]
    case 'scene':
      return [
        'Reader test: A reader can identify the scene goal, obstacle, and change by the end of the section.',
        'Text signal: Entrances, exits, and spatial cues remain clear while leverage shifts.',
        'Failure signal: The scene feels like conversation without turn, pressure, or changed state.',
      ]
    case 'dialogue':
      return [
        'Reader test: A reader can explain each speaker\'s intent without needing heavy dialogue tags.',
        'Text signal: Lines carry conflict, subtext, or power movement instead of exposition recap.',
        'Failure signal: Dialogue repeats known facts and leaves stakes unchanged.',
      ]
    case 'pacing':
      return [
        'Reader test: A reader feels steady escalation and can name where momentum increases.',
        'Text signal: Beat lengths vary by dramatic need and transitions preserve pressure.',
        'Failure signal: Sections stall, repeat emotional beats, or rush major turns without setup.',
      ]
    case 'thesis':
      return [
        'Reader test: A reader can state the central position and its scope in one sentence.',
        'Text signal: Thesis appears early and body sections map directly to its components.',
        'Failure signal: Readers find a topic but cannot identify the exact arguable claim.',
      ]
    case 'evidence':
      return [
        'Reader test: A reader can explain why each evidence item is present and what claim it supports.',
        'Text signal: Evidence is selected, interpreted, and proportioned to claim importance.',
        'Failure signal: Data appears without warrant, or minor points are over-supported while core claims stay thin.',
      ]
    case 'source':
      return [
        'Reader test: A reader can distinguish your argument from source summary at every major point.',
        'Text signal: Sources are framed for relevance, integrated cleanly, and synthesized across positions.',
        'Failure signal: Paragraphs read as citation stacks, patchwriting, or disconnected source notes.',
      ]
    case 'voice':
      return [
        'Reader test: Target readers describe the tone as credible, clear, and appropriately calibrated.',
        'Text signal: Register and stance stay stable while still matching audience context.',
        'Failure signal: The prose reads as inflated, evasive, harsh, or inconsistent by section.',
      ]
    case 'action':
      return [
        'Reader test: A reader can execute the next step without asking who acts or when.',
        'Text signal: Actions, owners, and deadlines are explicit and easy to locate.',
        'Failure signal: Readers ask follow-up questions about sequence, ownership, or timing.',
      ]
    case 'scan':
      return [
        'Reader test: A reader can find the main decision and key details in under a minute.',
        'Text signal: Headings, lists, and ordering surface high-priority information first.',
        'Failure signal: Important details are buried in dense blocks or inconsistent formatting.',
      ]
    case 'structure':
      return [
        'Reader test: A reader can outline the section flow and explain why each part follows the prior one.',
        'Text signal: Paragraphs and transitions perform one clear function at a time.',
        'Failure signal: Topic jumps force readers to reconstruct links between sections.',
      ]
    case 'analysis':
      return [
        'Reader test: A reader can distinguish observation, interpretation, and implication in each major paragraph.',
        'Text signal: Inference steps are explicit, and competing readings are handled fairly.',
        'Failure signal: The draft summarizes evidence but does not show how conclusions were derived.',
      ]
    case 'mechanics':
      return [
        'Reader test: A reader can process sentences on first pass without stopping for grammar or wording confusion.',
        'Text signal: Syntax, punctuation, and terminology consistently support meaning.',
        'Failure signal: Readers re-read lines for basic meaning because errors or vagueness interrupt flow.',
      ]
    case 'revision':
      return [
        'Reader test: A reader can see substantive improvement between drafts, not only cosmetic edits.',
        'Text signal: Revision passes target high-impact issues in a deliberate order.',
        'Failure signal: Surface polish improves while core argument or structure weaknesses persist.',
      ]
    case 'technical':
      return [
        'Reader test: A domain-aware reader can execute or verify the content without correcting facts, units, or procedural steps.',
        'Text signal: Terms are used consistently, constraints are explicit, and steps preserve operational accuracy.',
        'Failure signal: The draft sounds fluent but introduces factual drift, ambiguous instructions, or invalid assumptions.',
      ]
    default:
      return [
        'Reader test: A reader can state the section purpose, key support, and next implication after one pass.',
        'Text signal: Paragraph roles are distinct and links between ideas are explicit.',
        'Failure signal: Readers can name the topic but cannot explain progression or decision impact.',
      ]
  }
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
