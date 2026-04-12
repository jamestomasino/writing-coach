import type { SkillGraphNode } from './types'

export type ObjectiveExampleSource = {
  label: string
  url: string
}

export type ObjectiveExampleSet = {
  good: string
  needsWork: string
  sources: ObjectiveExampleSource[]
  strategy: string
}

type ExampleBundle = {
  strategy: string
  good: string[]
  needsWork: string[]
  sources: ObjectiveExampleSource[]
}

const SOURCE_CATALOG = {
  purdueParagraphs: {
    label: 'Purdue OWL: Paragraphing',
    url: 'https://owl.purdue.edu/owl/general_writing/academic_writing/paragraphs_and_paragraphing/index.html',
  },
  purdueThesis: {
    label: 'Purdue OWL: Developing Strong Thesis Statements',
    url: 'https://owl.purdue.edu/owl/general_writing/academic_writing/establishing_arguments/index.html',
  },
  purdueProofreading: {
    label: 'Purdue OWL: Proofreading for Errors',
    url: 'https://owl.purdue.edu/owl/general_writing/the_writing_process/proofreading/proofreading_for_errors.html',
  },
  purdueActivePassive: {
    label: 'Purdue OWL: Active and Passive Voice',
    url: 'https://owl.purdue.edu/owl/general_writing/academic_writing/active_and_passive_voice/index.html',
  },
  purdueBusinessTone: {
    label: 'Purdue OWL: Tone in Business Writing',
    url: 'https://owl.purdue.edu/owl/subject_specific_writing/professional_technical_writing/tone_in_business_writing.html',
  },
  purdueApaStylistics: {
    label: 'Purdue OWL: APA Stylistics Basics',
    url: 'https://owl.purdue.edu/owl/research_and_citation/apa_style/apa_formatting_and_style_guide/apa_stylistics_basics.html',
  },
  uncTransitions: {
    label: 'UNC Writing Center: Transitions',
    url: 'https://writingcenter.unc.edu/tips-and-tools/transitions/',
  },
  uncThesis: {
    label: 'UNC Writing Center: Thesis Statements',
    url: 'https://writingcenter.unc.edu/tips-and-tools/thesis-statements/',
  },
  digitalGovPlainLanguage: {
    label: 'Digital.gov: Plain Language Guide',
    url: 'https://digital.gov/guides/plain-language/',
  },
  govUkWriting: {
    label: 'GOV.UK: Writing for GOV.UK',
    url: 'https://www.gov.uk/guidance/content-design/writing-for-gov-uk',
  },
  microsoftStyle: {
    label: 'Microsoft Writing Style Guide',
    url: 'https://learn.microsoft.com/en-us/style-guide/welcome/',
  },
  googleDevStyleVoice: {
    label: 'Google Developer Style Guide: Active Voice',
    url: 'https://developers.google.com/style/voice',
  },
  projectGutenberg: {
    label: 'Project Gutenberg (public-domain literature corpus)',
    url: 'https://www.gutenberg.org/',
  },
} satisfies Record<string, ObjectiveExampleSource>

const BUNDLES: Record<string, ExampleBundle> = {
  clarity: {
    strategy: 'plain-language objective framing with one concrete reader outcome',
    good: [
      'In the first paragraph, the writer states the decision: shift Saturday support hours to 8 a.m.-2 p.m. because ticket volume peaks before noon.',
      'Each paragraph does one job: context, recommendation, evidence, and next step. A reader can summarize the message after one pass.',
      'Key terms stay stable: the draft uses "pilot" for the four-week trial and never swaps in "launch" or "rollout" midstream.',
    ],
    needsWork: [
      'The draft opens with background, then jumps to recommendations, then circles back to context. The reader has to reconstruct the point.',
      'The same term changes meaning by section: "migration" refers to data transfer in one paragraph and team staffing in the next.',
      'The message implies a decision but never states it directly, so readers leave unsure what exactly was approved.',
    ],
    sources: [SOURCE_CATALOG.digitalGovPlainLanguage, SOURCE_CATALOG.govUkWriting, SOURCE_CATALOG.purdueParagraphs],
  },
  claim: {
    strategy: 'debatable thesis style from writing-center guidance',
    good: [
      'The argument is specific and contestable: remote onboarding should include a paid mentor hour each week because early social support reduces month-one attrition.',
      'The thesis appears early and guides the structure: claim first, then two reasons, then evidence for each reason.',
      'The writer states what changes if the claim is adopted: fewer handoff errors and faster time-to-productivity for new analysts.',
    ],
    needsWork: [
      'The central point stays broad: "onboarding matters." No one can disagree with it, so there is no real argument to test.',
      'The claim appears only in the conclusion, so earlier sections read like unrelated notes instead of one coherent case.',
      'The piece lists opinions but does not commit to a position, leaving readers without a clear stance to evaluate.',
    ],
    sources: [SOURCE_CATALOG.purdueThesis, SOURCE_CATALOG.uncThesis],
  },
  audience: {
    strategy: 'reader-need mapping and channel fit',
    good: [
      'For executives, the writer leads with risk, cost, and decision date; implementation detail is moved to an appendix for operators.',
      'The document anticipates stakeholder objections and answers them before asking for approval.',
      'Version A for engineers includes exact constraints; version B for clients explains impact in plain language with no internal jargon.',
    ],
    needsWork: [
      'The same draft tries to serve experts and first-time readers at once, so it is too thin for one group and too dense for the other.',
      'Likely objections are ignored, so the reader hits friction and disengages at the first unanswered concern.',
      'The piece uses team shorthand without explanation, forcing outside readers to decode terms before they can follow the argument.',
    ],
    sources: [SOURCE_CATALOG.govUkWriting, SOURCE_CATALOG.digitalGovPlainLanguage, SOURCE_CATALOG.purdueBusinessTone],
  },
  structure: {
    strategy: 'signposted progression with explicit transitions',
    good: [
      'Section headers mirror the reasoning path: Problem -> Options -> Recommendation -> Risks -> Owner and Date.',
      'Bridge lines explain the handoff: "Now that cost is bounded, the next question is rollout risk."',
      'Each section ends by setting up the next one, so the reader can track why the sequence is intentional.',
    ],
    needsWork: [
      'Headings are generic ("Overview," "More Notes," "Other Thoughts") and do not signal argumentative function.',
      'Paragraphs shift topics without transition, so readers cannot tell whether the writer is adding support or changing direction.',
      'The structure repeats similar points across sections instead of compounding toward a decision.',
    ],
    sources: [SOURCE_CATALOG.uncTransitions, SOURCE_CATALOG.purdueParagraphs, SOURCE_CATALOG.govUkWriting],
  },
  evidence: {
    strategy: 'claim-support pairing with interpretation',
    good: [
      'The writer cites a support-ticket sample, then explains why the pattern supports the recommendation rather than assuming the data speaks for itself.',
      'Each claim is paired with one focused example that matches scope and context.',
      'Quoted material is short and framed; most of the paragraph is the writer interpreting relevance and limits.',
    ],
    needsWork: [
      'Evidence appears as a list of numbers with no explanation of what decision the numbers support.',
      'The example is real but off-target; it describes a different user segment than the claim being made.',
      'Long quotes replace analysis, so the reader sees source text but not the writer\'s reasoning.',
    ],
    sources: [SOURCE_CATALOG.purdueThesis, SOURCE_CATALOG.purdueApaStylistics, SOURCE_CATALOG.uncThesis],
  },
  reasoning: {
    strategy: 'cause-effect chain and objection-aware logic',
    good: [
      'The chain is explicit: delayed bug triage increases reopen rate, which extends cycle time, which pushes release risk into the next sprint.',
      'The writer compares two plausible interpretations and shows why one better fits the observed evidence.',
      'A strong counterargument is stated fairly, then narrowed with a clear boundary condition.',
    ],
    needsWork: [
      'Conclusions arrive without intermediate reasoning, so readers must guess how the writer moved from data to claim.',
      'The draft treats correlation as proof of causation without testing alternate explanations.',
      'Potential objections are dismissed instead of addressed, which weakens credibility under scrutiny.',
    ],
    sources: [SOURCE_CATALOG.purdueThesis, SOURCE_CATALOG.uncTransitions],
  },
  actionability: {
    strategy: 'owner + action + deadline formatting',
    good: [
      'Action is explicit: "Priya submits the vendor shortlist by Tuesday 2:00 p.m.; Marco signs off by end of day."',
      'The closing states the next move and success criteria: run a seven-day pilot and report defect rate and completion time.',
      'Error guidance tells users exactly what to do: retry with a valid project ID or contact support with the shown reference code.',
    ],
    needsWork: [
      'The ask is implied with soft language ("it might be good to follow up soon") and no owner or due date.',
      'The document ends with "let me know your thoughts" even though a decision is required this week.',
      'Error text names the failure but gives no recovery path, leaving users blocked.',
    ],
    sources: [SOURCE_CATALOG.digitalGovPlainLanguage, SOURCE_CATALOG.govUkWriting, SOURCE_CATALOG.microsoftStyle],
  },
  scannability: {
    strategy: 'front-loaded key line with parallel lists',
    good: [
      'The first two lines provide the decision, deadline, and owner; details are grouped under short, descriptive subheads.',
      'Bullets are parallel and specific: each line starts with a verb and ends with one measurable outcome.',
      'Long sections are chunked into short paragraphs with one purpose each, so busy readers can locate answers fast.',
    ],
    needsWork: [
      'Key details are buried in dense paragraphs, forcing readers to search for dates and responsibilities.',
      'Bullet points mix nouns, questions, and full paragraphs, so scanning does not reduce effort.',
      'The document repeats context in every section instead of summarizing once and linking forward.',
    ],
    sources: [SOURCE_CATALOG.govUkWriting, SOURCE_CATALOG.digitalGovPlainLanguage, SOURCE_CATALOG.microsoftStyle],
  },
  sentenceControl: {
    strategy: 'active voice and economical syntax',
    good: [
      'The sentence names actor and action directly: "The audit team flagged three unresolved controls in Q2."',
      'A long sentence is split into two clean steps, preserving meaning while reducing reader load.',
      'Modifiers are trimmed so the verb carries the point instead of filler phrases.',
    ],
    needsWork: [
      'The sentence hides agency in passive chains ("it was determined") so responsibility is unclear.',
      'Multiple clauses stack without clear hierarchy, forcing rereads to recover the main idea.',
      'Filler phrases and hedges dilute the point until the sentence says little despite its length.',
    ],
    sources: [SOURCE_CATALOG.purdueActivePassive, SOURCE_CATALOG.googleDevStyleVoice, SOURCE_CATALOG.microsoftStyle],
  },
  paragraph: {
    strategy: 'one-job-per-paragraph with visible topic line',
    good: [
      'The paragraph opens with its claim, develops one line of support, and ends by linking to the next move.',
      'Support details all serve the same purpose, so the paragraph feels developed rather than scattered.',
      'Paragraph length matches function: short for status updates, fuller where reasoning needs development.',
    ],
    needsWork: [
      'One paragraph mixes definition, anecdote, and recommendation with no controlling topic sentence.',
      'Support details drift into side issues, so the original point gets buried by the end.',
      'Paragraph boundaries follow visual convenience rather than idea boundaries, making logic hard to track.',
    ],
    sources: [SOURCE_CATALOG.purdueParagraphs, SOURCE_CATALOG.uncTransitions],
  },
  narrative: {
    strategy: 'beat-level cause and effect anchored in scene time',
    good: [
      'When Lena pockets the key, the alarm stays armed; that one choice triggers the power cut and traps both characters in the archive.',
      'The timeline is easy to follow: sunrise briefing, noon confrontation, night evacuation, each with clear temporal markers.',
      'Backstory appears only where it changes present pressure, then the scene returns to immediate action.',
    ],
    needsWork: [
      'Major outcomes arrive without triggering choices, so events feel authored rather than caused.',
      'The story jumps across time with no anchor words, forcing readers to rebuild chronology by inference.',
      'Backstory interrupts the scene for a full page and drains urgency from the active conflict.',
    ],
    sources: [SOURCE_CATALOG.projectGutenberg, SOURCE_CATALOG.purdueParagraphs, SOURCE_CATALOG.uncTransitions],
  },
  scene: {
    strategy: 'goal-conflict-turn structure with spatial orientation',
    good: [
      'The scene starts with a concrete goal, escalates through resistance, and ends in a reversal that changes the next decision.',
      'Movement and object placement stay legible: readers can track who is near the exit, who holds the file, and why that matters.',
      'Entrances and exits change power immediately, so boundaries between scenes carry dramatic weight.',
    ],
    needsWork: [
      'Characters talk at length without pursuing conflicting goals, so the scene has motion but little pressure.',
      'Spatial cues are missing; props appear when needed without earlier placement, breaking scene credibility.',
      'The scene ends on information repetition rather than a turn, so momentum stalls into the next section.',
    ],
    sources: [SOURCE_CATALOG.projectGutenberg, SOURCE_CATALOG.uncTransitions],
  },
  dialogue: {
    strategy: 'spoken lines reveal intent and power shift',
    good: [
      'Each line does work: one character asks for access, the other deflects, and status shifts with each exchange.',
      'Subtext carries pressure; the words stay polite while the implied threat becomes unmistakable.',
      'Speaker intent stays distinct, so readers can follow conflict without relying on heavy dialogue tags.',
    ],
    needsWork: [
      'Dialogue repeats known facts and does not alter stakes, so the scene reads like exposition in quotation marks.',
      'All speakers share the same cadence and diction, making character voices hard to distinguish.',
      'Lines state emotions directly instead of letting gesture, silence, or contradiction reveal them.',
    ],
    sources: [SOURCE_CATALOG.projectGutenberg, SOURCE_CATALOG.purdueParagraphs],
  },
  precision: {
    strategy: 'specific nouns and verbs over abstract modifiers',
    good: [
      'Instead of "things improved," the writer names the shift: onboarding time dropped from 14 days to 9 days after checklist rollout.',
      'The draft replaces soft qualifiers with concrete detail, so readers can picture action and measure impact.',
      'Technical terms are accurate and consistent, reducing ambiguity in high-stakes instructions.',
    ],
    needsWork: [
      'The prose leans on broad words ("nice," "issues," "stuff") that hide what actually happened.',
      'Measurements are omitted, so improvement claims cannot be evaluated or compared across drafts.',
      'Terminology drifts between synonyms, creating avoidable uncertainty about process and scope.',
    ],
    sources: [SOURCE_CATALOG.digitalGovPlainLanguage, SOURCE_CATALOG.googleDevStyleVoice, SOURCE_CATALOG.purdueProofreading],
  },
  voiceTone: {
    strategy: 'credible, direct tone calibrated to audience and stakes',
    good: [
      'The voice is confident without inflation: it states limits, then makes a clear recommendation grounded in available evidence.',
      'Tone stays respectful under disagreement, preserving trust while still drawing a firm line on risk.',
      'Cadence varies for emphasis: short sentences land decisions, longer ones unpack reasoning.',
    ],
    needsWork: [
      'The prose overclaims certainty and reads as performative authority instead of earned judgment.',
      'Tone swings from casual to legalistic across sections, making stance and audience unclear.',
      'The writer softens every recommendation with hedges, so important decisions feel optional.',
    ],
    sources: [SOURCE_CATALOG.purdueBusinessTone, SOURCE_CATALOG.microsoftStyle, SOURCE_CATALOG.govUkWriting],
  },
  grammarMechanics: {
    strategy: 'error pattern control through targeted proofreading passes',
    good: [
      'Subject-verb agreement holds in complex sentences, and punctuation clarifies clause boundaries instead of obscuring them.',
      'A final mechanics pass catches tense drift and comma splices before submission.',
      'Dialogue punctuation and capitalization are consistent, so form no longer distracts from meaning.',
    ],
    needsWork: [
      'Frequent agreement and tense errors force the reader to pause and reinterpret sentence meaning.',
      'Comma placement is inconsistent, creating accidental run-ons and fragments in key passages.',
      'Mechanical errors recur across revisions, suggesting no deliberate proofreading strategy.',
    ],
    sources: [SOURCE_CATALOG.purdueProofreading, SOURCE_CATALOG.purdueActivePassive],
  },
  revision: {
    strategy: 'ordered revision passes: clarity -> support -> style',
    good: [
      'The writer triages revision in sequence: first fixes argument gaps, then evidence alignment, then sentence polish.',
      'A focused clarity pass removes ambiguous references and repairs causal gaps before stylistic edits.',
      'Later drafts are measurably better on the targeted objective, not just different in wording.',
    ],
    needsWork: [
      'Revision effort goes into surface edits while core structure and reasoning problems remain unchanged.',
      'Each new draft rewrites randomly without a priority order, so recurring weaknesses persist.',
      'Changes are broad but untracked, making it hard to tell whether the target objective actually improved.',
    ],
    sources: [SOURCE_CATALOG.purdueProofreading, SOURCE_CATALOG.uncTransitions, SOURCE_CATALOG.digitalGovPlainLanguage],
  },
  sourceQuality: {
    strategy: 'source reliability, framing, and synthesis',
    good: [
      'Sources are current, relevant, and introduced with context about why each source is credible for this claim.',
      'The writer synthesizes multiple sources to build a single line of reasoning instead of stacking isolated quotes.',
      'Citation handling is consistent and transparent, so readers can verify evidence quickly.',
    ],
    needsWork: [
      'The draft leans on unsourced assertions or low-credibility references for high-stakes claims.',
      'Sources are dropped in without framing, leaving readers unsure how each source supports the argument.',
      'Citation format and attribution are inconsistent, which weakens trust and traceability.',
    ],
    sources: [SOURCE_CATALOG.purdueApaStylistics, SOURCE_CATALOG.purdueThesis, SOURCE_CATALOG.microsoftStyle],
  },
}

const CLUSTER_BY_SKILL: Record<string, string> = {
  'clarity and coherence': 'clarity',
  'claim clarity': 'claim',
  'audience alignment': 'audience',
  'structural signposting': 'structure',
  'evidence integration': 'evidence',
  'narrative clarity': 'narrative',
  'scene architecture': 'scene',
  actionability: 'actionability',
  scannability: 'scannability',
  'voice presence': 'voiceTone',
  'sentence economy': 'sentenceControl',
  'prose precision': 'precision',
  'emotional compression': 'narrative',
  'dialogue intelligence': 'dialogue',
  'worldbuilding economy': 'narrative',
  'word choice': 'precision',
  'sentence variety': 'sentenceControl',
  'sentence complexity': 'sentenceControl',
  'paragraph control': 'paragraph',
  'narrative sequencing': 'narrative',
  'descriptive precision': 'precision',
  'dialogue basics': 'dialogue',
  'insight density': 'reasoning',
  'authority and voice': 'voiceTone',
  'tone calibration': 'voiceTone',
  accuracy: 'precision',
  'analysis depth': 'reasoning',
  'assignment alignment': 'audience',
  'example quality': 'evidence',
  'grammar control': 'grammarMechanics',
  'objection handling': 'reasoning',
  'professional format': 'scannability',
  'reasoning quality': 'reasoning',
  'reflection depth': 'reasoning',
  'revision habits': 'revision',
  'rhetorical force': 'voiceTone',
  'source handling': 'sourceQuality',
  'spelling and mechanics': 'grammarMechanics',
  'story development': 'narrative',
  'structure and pacing': 'scene',
  'technical precision': 'precision',
  'thesis clarity': 'claim',
  'user goal alignment': 'actionability',
  'image freshness': 'precision',
  'symbolic control': 'narrative',
}

function hash(input: string) {
  let h = 2166136261
  for (let index = 0; index < input.length; index += 1) {
    h ^= input.charCodeAt(index)
    h = Math.imul(h, 16777619)
  }
  return h >>> 0
}

function pick<T>(items: T[], key: string) {
  if (items.length === 0) {
    throw new Error('Cannot pick from empty list')
  }
  return items[hash(key) % items.length]
}

function inferCluster(node: SkillGraphNode) {
  const skill = (node.skill_name ?? '').trim().toLowerCase()
  if (skill && CLUSTER_BY_SKILL[skill]) {
    return CLUSTER_BY_SKILL[skill]
  }

  const bag = `${node.code} ${node.title} ${node.description}`.toLowerCase()
  if (/(claim|thesis|argument)/.test(bag)) {
    return 'claim'
  }
  if (/(dialogue|subtext|speaker|voice difference)/.test(bag)) {
    return 'dialogue'
  }
  if (/(scene|tension|reversal|entrance|exit|pacing|spatial)/.test(bag)) {
    return 'scene'
  }
  if (/(source|citation|quote|evidence)/.test(bag)) {
    return 'sourceQuality'
  }
  if (/(grammar|spelling|punctuation|tense|comma|agreement)/.test(bag)) {
    return 'grammarMechanics'
  }
  if (/(revision|pass|triage|proofread)/.test(bag)) {
    return 'revision'
  }
  if (/(ask|owner|deadline|next step|action|instruction)/.test(bag)) {
    return 'actionability'
  }
  if (/(causal|timeline|sequence|backstory|point of view|narrative)/.test(bag)) {
    return 'narrative'
  }
  return 'clarity'
}

const CLUSTER_SIGNAL: Record<string, string> = {
  clarity: 'The controlling idea and terminology stay stable for the reader.',
  claim: 'The central claim is explicit, specific, and testable.',
  audience: 'Framing and depth match the intended reader and channel.',
  structure: 'Section order and transitions make reasoning easy to follow.',
  evidence: 'Evidence is chosen precisely and interpreted for relevance.',
  reasoning: 'Logic is explicit, sequenced, and resilient under objection.',
  actionability: 'Next actions are concrete, owned, and time-bounded.',
  scannability: 'Key points are front-loaded and easy to find under time pressure.',
  sentenceControl: 'Sentence-level control keeps meaning fast and unambiguous.',
  paragraph: 'Each paragraph has one clear function and visible development.',
  narrative: 'Story causality and temporal orientation remain legible.',
  scene: 'Scene goals, friction, and turns produce momentum.',
  dialogue: 'Dialogue reveals intent, pressure, and change.',
  precision: 'Diction and detail are specific enough to verify and visualize.',
  voiceTone: 'Voice stays credible while tone fits audience and stakes.',
  grammarMechanics: 'Mechanical control supports fluency and trust.',
  revision: 'Revision is prioritized in a deliberate, high-leverage order.',
  sourceQuality: 'Sources are reliable, framed, and synthesized into argument.',
}

const CLUSTER_GAP_SIGNAL: Record<string, string> = {
  clarity: 'The controlling idea or terminology still drifts across sections.',
  claim: 'The claim remains too broad, late, or non-testable.',
  audience: 'Reader needs and channel constraints are not consistently addressed.',
  structure: 'Section order or transitions force the reader to reconstruct logic.',
  evidence: 'Evidence appears but is weakly matched or under-interpreted.',
  reasoning: 'Reasoning steps are skipped, brittle, or objection-blind.',
  actionability: 'Next steps are still vague on owner, timing, or completion criteria.',
  scannability: 'Important details are buried and hard to locate quickly.',
  sentenceControl: 'Sentence shape still adds avoidable processing friction.',
  paragraph: 'Paragraph boundaries or focus remain unstable.',
  narrative: 'Causality or chronology still requires reader guesswork.',
  scene: 'Scene pressure flattens because goals or turns are underbuilt.',
  dialogue: 'Dialogue still carries exposition more than dramatic movement.',
  precision: 'Word choice remains broad where specificity is needed.',
  voiceTone: 'Tone or authority signal is inconsistent with audience expectations.',
  grammarMechanics: 'Mechanical errors continue to compete with meaning.',
  revision: 'Revision effort is still spread thin instead of prioritized.',
  sourceQuality: 'Source credibility, framing, or synthesis remains uneven.',
}

function objectiveSignal(node: SkillGraphNode, cluster: string) {
  void node
  return CLUSTER_SIGNAL[cluster] ?? `The draft shows observable control of ${node.title.toLowerCase()}.`
}

function objectiveGapSignal(node: SkillGraphNode, cluster: string) {
  void node
  return CLUSTER_GAP_SIGNAL[cluster] ?? `Control of ${node.title.toLowerCase()} is still inconsistent across the draft.`
}

export function buildObjectiveExamples(node: SkillGraphNode): ObjectiveExampleSet {
  const cluster = inferCluster(node)
  const bundle = BUNDLES[cluster] ?? BUNDLES.clarity
  const key = `${node.code}:${node.title}:${node.stage}:${node.stage_order}`

  return {
    good: `Good: ${pick(bundle.good, `${key}:good`)} ${objectiveSignal(node, cluster)}`,
    needsWork: `Needs work: ${pick(bundle.needsWork, `${key}:bad`)} ${objectiveGapSignal(node, cluster)}`,
    sources: bundle.sources,
    strategy: bundle.strategy,
  }
}

export function objectiveExampleCoverageBySkill() {
  return {
    skillToCluster: { ...CLUSTER_BY_SKILL },
    clusterCount: Object.keys(BUNDLES).length,
  }
}
