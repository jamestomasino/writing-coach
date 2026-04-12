import { buildObjectiveExamples } from './objective-example-library'
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

function lowerFirst(value: string) {
  if (!value) return value
  return value[0].toLowerCase() + value.slice(1)
}

function stripTrailingPunctuation(value: string) {
  return value.trim().replace(/[.!?]+$/, '')
}

function firstSentence(value: string) {
  const trimmed = value.trim()
  if (!trimmed) return ''
  const match = trimmed.match(/^(.+?[.!?])(?:\s|$)/)
  return (match?.[1] ?? trimmed).trim()
}

function definitionFromOverview(value: string, title: string) {
  const base = stripTrailingPunctuation(firstSentence(value))
  if (!base) {
    return sentence(`${title} means applying this objective clearly and consistently in the draft`)
  }
  const normalized = base
    .replace(/^this objective teaches\s+/i, `${title} means `)
    .replace(/^this objective is about\s+/i, `${title} means `)
    .replace(/^this objective means\s+/i, `${title} means `)
    .replace(/^this skill teaches\s+/i, `${title} means `)
    .replace(/^this skill is about\s+/i, `${title} means `)
    .replace(/^this skill means\s+/i, `${title} means `)
    .replace(/^this objective\s+/i, `${title} `)
    .replace(/^this skill\s+/i, `${title} `)
  if (/^[-a-z0-9 ,'"()]+means\b/i.test(normalized)) {
    return sentence(normalized)
  }
  return sentence(`${title} means ${lowerFirst(normalized)}`)
}

const WHY_OVERRIDE_BY_CODE: Record<string, string> = {
  'academic-prompt-reading':
    'This matters because a misread assignment produces polished work that still misses the task. Accurate prompt parsing turns every later paragraph into targeted response instead of drift.',
  'objective-clarity':
    'This matters because readers cannot evaluate progress when the target is vague. A sharp target lets each section prove movement toward a visible outcome.',
  'context-first':
    'This matters because readers need orientation before they can interpret claims or scenes. Early context prevents confusion and reduces correction overhead later.',
  'scaffolding-load':
    'This matters because too much setup at once overwhelms working memory. Staged scaffolding keeps readers moving while still preserving depth.',
  'pronoun-clarity':
    'This matters because unclear references force rereading at exactly the moment attention should move forward. Clear referents keep meaning stable and pace intact.',
  'spec-clarity':
    'This matters because ambiguous requirements create rework and conflicting interpretations. Specific constraints let readers execute without guessing hidden rules.',
  'assumption-marking':
    'This matters because hidden assumptions make arguments brittle under scrutiny. Marking assumptions lets readers test scope, limits, and transfer conditions.',
  'academic-hedging-control':
    'This matters because over-qualification weakens claims while under-qualification overstates certainty. Calibrated qualification protects credibility and keeps the argument usable.',
  'informational-writing-basics':
    'This matters because informational prose fails when readers cannot find the core point quickly. Strong informational framing lets readers absorb and act in one pass.',
  'persuasive-plainstyle':
    'This matters because ornate language obscures decisions and inflates cognitive load. Plain phrasing preserves force by making claims easy to process and challenge.',
  'tradeoff-language':
    'This matters because decisions always carry costs and constraints. Explicit tradeoff language helps readers compare options without false certainty.',
  'academic-abstract-noun-control':
    'This matters because abstraction without grounding sounds sophisticated but explains little. Concrete anchors let readers see what each concept means in practice.',
  'summary-basics':
    'This matters because weak summary either omits stakes or copies surface detail. Strong summary preserves the argument spine while compressing length.',
  'alignment-language':
    'This matters because small wording mismatches between question, claim, and evidence can derail trust. Alignment language keeps each part pointing at the same problem.',
  'focus-under-length':
    'This matters because short formats magnify every irrelevant line. Focus under constraint preserves signal when space is limited.',
  'story-wound-pattern':
    'This matters because unresolved emotional patterns should shape present choices, not sit as backstory notes. Pattern visibility makes character behavior feel motivated rather than arbitrary.',
  'memoir-shame-handling':
    'This matters because shame-heavy material can collapse into self-protection or melodrama. Controlled handling keeps honesty, dignity, and reader trust in balance.',
  'memoir-selective-detail':
    'This matters because memoir impact depends on choosing details that carry memory and meaning. Selective detail prevents clutter and sharpens emotional precision.',
  'story-description-focus':
    'This matters because description should support movement, not pause it. Focused description keeps scenes vivid while preserving narrative momentum.',
  'memoir-metaphor-restraint':
    'This matters because heavy figurative language can sentimentalize lived material. Restraint keeps figurative lines earned and emotionally credible.',
  'story-atmosphere':
    'This matters because mood without control either muddies action or feels generic. Deliberate atmosphere supports tension, orientation, and tone coherence.',
  'story-image-freshness':
    'This matters because stock imagery fades on arrival and weakens voice. Fresh images create memorability and sharpen scene identity.',
  'memoir-body-awareness':
    'This matters because embodied cues carry truth that abstract reflection cannot. Body-level signals anchor memory in concrete lived experience.',
  'memoir-time-awareness':
    'This matters because memoir often moves across time and perspective. Clear temporal handling keeps readers oriented through shifts and returns.',
  'memoir-status-awareness':
    'This matters because social status dynamics quietly shape conflict and interpretation. Status awareness makes power movement legible without overexplaining it.',
  'story-backstory-restraint':
    'This matters because backstory can easily overtake present action. Controlled insertion keeps context useful while pressure remains active.',
  'memoir-memory-fragment':
    'This matters because fragment structure can either illuminate memory truth or confuse sequence. Controlled fragmentation preserves meaning while honoring discontinuity.',
  'story-exposition-control':
    'This matters because explanation that overtakes action flattens urgency and makes scenes feel static. Controlled narrative context keeps readers oriented while preserving momentum and emotional pressure.',
  'memoir-structure-linearity':
    'This matters because linear flow helps readers track causation and development across lived events. Strong linear handling prevents drift and retrospective blur.',
  'memoir-structure-braiding':
    'This matters because braided structure only works when threads earn convergence. Clear braid control keeps thematic payoff from feeling accidental.',
  'memoir-summary-compression':
    'This matters because compressed time can preserve pace or erase significance. Good compression keeps what changed visible while cutting low-value steps.',
  'opening-focus':
    'This matters because the opening sets interpretive expectations for everything after it. A focused start establishes stakes and direction before reader attention decays.',
  'persuasive-objection-forecast':
    'This matters because unaddressed resistance hardens before your strongest evidence appears. Early forecast lets the reader feel understood before being persuaded.',
  'persuasive-concession':
    'This matters because fair concession increases trust and strengthens the remaining claim. Strategic concession shows judgment rather than weakness.',
  'persuasive-ethical-pressure':
    'This matters because persuasive force without ethical boundaries becomes coercive. Ethical pressure keeps urgency high while respecting reader agency.',
  'letter-structure':
    'This matters because letters and emails are judged first by navigability. Clear opening, body, and close reduce misunderstanding and response delay.',
  'memoir-reflection-basics':
    'This matters because reflection should interpret experience, not repeat it. Basic reflective control turns event recall into insight.',
  'memoir-self-myth-avoidance':
    'This matters because self-myth narrows complexity and undermines credibility. Avoiding myth keeps the voice honest, layered, and trustworthy.',
  'memoir-question-holding':
    'This matters because premature closure flattens inquiry and reduces tension. Holding the right question keeps reflection alive across sections.',
  'memoir-pattern-linking':
    'This matters because isolated moments rarely produce insight on their own. Pattern linking shows how repeated signals accumulate into meaning.',
  'memoir-theme-pressure':
    'This matters because theme should emerge under narrative pressure, not as detached statement. Pressure-tested themes feel discovered rather than declared.',
  'memoir-loyalty-pressure':
    'This matters because loyalty conflicts drive some of the hardest moral decisions in memoir. Making that pressure visible clarifies stakes and consequence.',
  'memoir-ending-resonance':
    'This matters because endings are judged on aftereffect, not only closure. Resonant endings preserve thematic echo while honoring what changed.',
  'story-character-goal':
    'This matters because goals organize behavior and conflict at scene level. Clear goals make decisions legible and outcomes consequential.',
  'story-character-need':
    'This matters because deeper need explains why surface goals alone do not resolve tension. Need visibility adds psychological depth to choices.',
  'story-contradiction':
    'This matters because contradiction creates believable internal conflict. Managed contradiction prevents characters from reading as flat or purely strategic.',
  'character-goal-basics':
    'This matters because beginner character work fails when motivation stays implied. Basic goal clarity gives scenes a readable engine.',
  'problem-and-solution':
    'This matters because readers evaluate solutions against a clearly defined problem frame. Tight pairing prevents mismatch between diagnosis and response.',
  'story-relationship-pressure':
    'This matters because relationships change stakes, options, and cost of action. Pressure mapping keeps interpersonal conflict active and specific.',
  'showing-action':
    'This matters because action reveals intent faster than explanation. Showing through behavior keeps pacing strong and subtext available.',
  'emotion-through-action':
    'This matters because declared feeling can sound generic without behavioral evidence. Action-based emotion makes affect credible and scene-bound.',
  'story-subplot-use':
    'This matters because subplots should compound the main line, not distract from it. Purposeful subplot use increases depth while preserving coherence.',
  'story-motif':
    'This matters because motif repetition can either enrich theme or feel forced. Controlled recurrence creates cumulative meaning across scenes.',
  'story-object-handling':
    'This matters because objects can carry memory, power, and consequence when tracked well. Clear object handling prevents symbolic confusion and continuity breaks.',
  'story-worldbuilding-economy':
    'This matters because setting detail competes with plot for reader attention. Economic worldbuilding keeps context vivid without stalling the story.',
  'story-social-structure':
    'This matters because social rules shape risk, permission, and conflict options. Making those rules legible strengthens plausibility and stakes.',
  'story-history-pressure':
    'This matters because historical force should actively constrain present choices. History pressure turns background into live conflict.'
}

function whyOverride(node: SkillGraphNode) {
  return WHY_OVERRIDE_BY_CODE[node.code.toLowerCase()] ?? null
}

function whySpecificTail(node: SkillGraphNode) {
  const move = sentence(
    sanitizeProse(node.description || 'the core move stays visible in the draft')
      .replace(/\breally\b/gi, '')
      .replace(/\bvery\b/gi, '')
      .replace(/\bkind of\b/gi, '')
  )
  const rawMarker = sentence(
    sanitizeProse(node.mastery_hint || 'readers can point to concrete improvement in this area')
      .replace(/\breally\b/gi, '')
      .replace(/\bvery\b/gi, '')
      .replace(/\bkind of\b/gi, '')
  )
  const marker = rawMarker.replace(/^because\s+/i, '').trim()
  return `${sentence(`In this objective, the practice focus is: ${lowerFirst(move)}`)} ${sentence(
    `Application signal: ${lowerFirst(marker)}`
  )}`.trim()
}

function successMoveCue(node: SkillGraphNode) {
  const move = sentence(
    sanitizeProse(node.description || 'the target move is visible in the draft')
      .replace(/\breally\b/gi, '')
      .replace(/\bvery\b/gi, '')
      .replace(/\bkind of\b/gi, '')
  )
  return move.replace(/[.!?]+$/, '').trim()
}

function successMarkerCue(node: SkillGraphNode) {
  const marker = sentence(
    sanitizeProse(node.mastery_hint || 'readers can point to reliable control in this area')
      .replace(/\breally\b/gi, '')
      .replace(/\bvery\b/gi, '')
      .replace(/\bkind of\b/gi, '')
  )
  return marker.replace(/^because\s+/i, '').replace(/[.!?]+$/, '').trim()
}

function enrichSuccessLooksLike(node: SkillGraphNode, base: string[]) {
  if (base.length !== 3) return base
  const move = successMoveCue(node)
  const marker = successMarkerCue(node)
  const out = [...base]
  out[1] = `${out[1]} Objective cue: ${move}.`
  out[2] = `${out[2]} Missing-marker check: ${marker}.`
  return out
}

const SKILL_OVERVIEW_OVERRIDE_BY_CODE: Record<string, string> = {
  'academic-prompt-reading':
    'This objective teaches assignment decoding before drafting begins. It helps the learner separate required deliverables, optional context, and hidden constraints so the draft answers the task directly.',
  'objective-clarity':
    'This objective teaches target-setting language that can be checked in the draft. It gives the learner a clear outcome statement so every paragraph can be evaluated against one purpose.',
  'context-first':
    'This objective teaches front-loading the minimum context readers need to interpret the section. It helps the learner place frame and stakes before detail so readers do not misread early claims.',
  'scaffolding-load':
    'This objective teaches pacing of setup information across a section. It helps the learner stage complexity in workable steps so understanding grows without overload.',
  'pronoun-clarity':
    'This objective teaches explicit referent control at sentence level. It helps the learner keep who-or-what each pronoun points to unambiguous under dense reasoning.',
  'spec-clarity':
    'This objective teaches requirement-writing precision. It helps the learner express constraints, conditions, and expected outputs so another person can execute without inference.',
  'assumption-marking':
    'This objective teaches explicit premise control. It helps the learner name hidden assumptions so claims can be tested for scope, risk, and transfer.',
  'academic-hedging-control':
    'This objective teaches calibrated certainty in academic argument. It helps the learner qualify claims where evidence is partial while still making a usable, defensible position.',
  'informational-writing-basics':
    'This objective teaches high-clarity informational prose structure. It helps the learner present point, support, and implication in an order that supports fast comprehension.',
  'persuasive-plainstyle':
    'This objective teaches persuasive clarity without rhetorical inflation. It helps the learner make arguments easier to evaluate by using direct language and low-friction syntax.',
  'tradeoff-language':
    'This objective teaches explicit tradeoff framing in decisions. It helps the learner state what improves, what worsens, and what constraints govern the recommendation.',
  'academic-abstract-noun-control':
    'This objective teaches conversion of abstraction into concrete meaning. It helps the learner anchor conceptual language in observable examples, mechanisms, or implications.',
  'summary-basics':
    'This objective teaches compression without loss of argumentative backbone. It helps the learner retain claim, key evidence, and stakes while reducing length.',
  'alignment-language':
    'This objective teaches lexical and conceptual consistency across prompt, claim, and support. It helps the learner keep wording aligned so readers can track one coherent line.',
  'focus-under-length':
    'This objective teaches prioritization under strict space limits. It helps the learner keep only high-value lines that move the reader toward decision or understanding.',
  'story-wound-pattern':
    'This objective teaches how recurring emotional injury patterns shape present behavior. It helps the learner connect past pressure to current choice so character action feels motivated.',
  'memoir-shame-handling':
    'This objective teaches handling vulnerable material with control and precision. It helps the learner write difficult moments honestly without collapsing into self-erasure or spectacle.',
  'memoir-selective-detail':
    'This objective teaches choosing details that carry memory weight and thematic relevance. It helps the learner avoid clutter by keeping only details that change interpretation.',
  'story-description-focus':
    'This objective teaches description that serves movement rather than pause. It helps the learner use concrete detail to support conflict, orientation, and stakes.',
  'memoir-metaphor-restraint':
    'This objective teaches disciplined figurative language in personal narrative. It helps the learner use metaphor where it sharpens truth and cut it where it softens hard reality.',
  'story-atmosphere':
    'This objective teaches deliberate mood-building through setting, rhythm, and selection. It helps the learner shape emotional field while keeping events legible.',
  'story-image-freshness':
    'This objective teaches image originality with narrative purpose. It helps the learner replace stock phrasing with specific imagery that carries identity and force.',
  'memoir-body-awareness':
    'This objective teaches embodiment as evidence in reflective writing. It helps the learner use physical sensation and reaction to ground interpretation in lived experience.',
  'memoir-time-awareness':
    'This objective teaches temporal control across memory movement. It helps the learner mark shifts clearly so readers can follow sequence and significance.',
  'memoir-status-awareness':
    'This objective teaches reading and writing social rank movement inside scenes. It helps the learner show how power and deference shape speech, risk, and choice.',
  'story-backstory-restraint':
    'This objective teaches selective insertion of past information into active scenes. It helps the learner keep backstory functional by tying it to present pressure.',
  'memoir-memory-fragment':
    'This objective teaches structured use of fragment form for discontinuous memory. It helps the learner preserve coherence while honoring broken recall patterns.',
  'story-exposition-control':
    'This objective teaches balancing explanation with forward scene motion. It helps the learner deliver needed context in small, high-impact placements rather than blocks.',
  'memoir-structure-linearity':
    'This objective teaches chronological arrangement for clarity and cumulative effect. It helps the learner maintain continuity so growth and causation remain visible.',
  'memoir-structure-braiding':
    'This objective teaches weaving multiple threads toward thematic convergence. It helps the learner control return points and transitions so the braid lands intentionally.',
  'memoir-summary-compression':
    'This objective teaches skipping low-value time while preserving change. It helps the learner condense without erasing turning points or emotional consequence.',
  'opening-focus':
    'This objective teaches opening architecture that immediately establishes orientation and stakes. It helps the learner begin with a usable frame instead of diffuse setup.',
  'persuasive-objection-forecast':
    'This objective teaches anticipating resistance before it hardens. It helps the learner build persuasive momentum by acknowledging likely objections early and fairly.',
  'persuasive-concession':
    'This objective teaches strategic concession that strengthens overall argument integrity. It helps the learner concede valid limits while preserving core claim force.',
  'persuasive-ethical-pressure':
    'This objective teaches urgent persuasion without manipulation. It helps the learner apply moral force with transparent reasoning and respect for reader agency.',
  'letter-structure':
    'This objective teaches communicative architecture for letters and emails. It helps the learner organize opening purpose, supporting body, and clear close for fast response.',
  'memoir-reflection-basics':
    'This objective teaches interpretation layered on event, not event repetition. It helps the learner convert recalled moments into insight and implication.',
  'memoir-self-myth-avoidance':
    'This objective teaches resisting simplifying self-narratives. It helps the learner maintain complexity so voice remains credible and ethically grounded.',
  'memoir-question-holding':
    'This objective teaches sustaining inquiry across a piece instead of closing too early. It helps the learner keep reflective tension active until the text earns resolution.',
  'memoir-pattern-linking':
    'This objective teaches linking recurring moments into coherent insight. It helps the learner move from isolated anecdote toward patterned interpretation.',
  'memoir-theme-pressure':
    'This objective teaches deriving theme from conflict and consequence. It helps the learner keep thematic language grounded in scene-level evidence.',
  'memoir-loyalty-pressure':
    'This objective teaches writing competing loyalties as active decision pressure. It helps the learner make ethical stakes visible without flattening relationships.',
  'memoir-ending-resonance':
    'This objective teaches closing with lasting thematic aftereffect. It helps the learner design endings that echo earlier pressure while honoring change.',
  'story-character-goal':
    'This objective teaches explicit short-horizon objectives for characters in motion. It helps the learner make behavior legible through goal-driven choices.',
  'story-character-need':
    'This objective teaches deeper motivational deficits beneath surface wants. It helps the learner build emotional and structural depth beyond immediate plot tasks.',
  'story-contradiction':
    'This objective teaches productive internal contradiction in characterization. It helps the learner create believable tension between values, wants, and behavior.',
  'character-goal-basics':
    'This objective teaches foundational motivation clarity for early narrative drafting. It helps the learner keep scenes active by giving each main actor a concrete pursuit.',
  'problem-and-solution':
    'This objective teaches coherent pairing of diagnosis and response. It helps the learner ensure proposed action actually addresses the defined problem conditions.',
  'story-relationship-pressure':
    'This objective teaches using relational stakes as conflict engine. It helps the learner show how attachment, obligation, and fear reshape decisions in scene.',
  'showing-action':
    'This objective teaches behavior-led revelation of intent and conflict. It helps the learner replace explanatory telling with visible action sequences.',
  'emotion-through-action':
    'This objective teaches expressing emotion through decisions, movement, and interaction. It helps the learner make feeling credible by embedding it in behavior.',
  'story-subplot-use':
    'This objective teaches subordinate thread design in support of the main arc. It helps the learner add depth without sacrificing narrative cohesion.',
  'story-motif':
    'This objective teaches controlled recurrence of symbolic or thematic elements. It helps the learner build cumulative meaning through patterned return.',
  'story-object-handling':
    'This objective teaches continuity and significance management for physical objects. It helps the learner use objects as anchors of memory, leverage, and consequence.',
  'story-worldbuilding-economy':
    'This objective teaches high-yield setting detail selection. It helps the learner reveal rules and texture without stalling active conflict.',
  'story-social-structure':
    'This objective teaches integrating social norms and hierarchy into conflict logic. It helps the learner show how institutions and expectations constrain action.',
  'story-history-pressure':
    'This objective teaches using historical force as present-tense constraint. It helps the learner convert background history into active stakes and decision cost.',
}

function skillOverviewOverride(node: SkillGraphNode) {
  return SKILL_OVERVIEW_OVERRIDE_BY_CODE[node.code.toLowerCase()] ?? null
}

const OBJECTIVE_GOAL_OVERRIDE_BY_CODE: Record<string, string> = {
  'academic-independent-voice':
    'Keep source integration subordinate to your own reasoning line so readers can hear your judgment clearly. Make your stance audible in interpretation, transitions, and conclusions rather than citation volume.',
  'persuasive-positioning':
    'Frame the issue using terms, stakes, and constraints that the target reader already recognizes so relevance lands early. Choose angle and language that makes uptake more likely.',
  'story-beta-awareness':
    'Anticipate where readers need orientation, reassurance, or proof so momentum does not collapse. Add cues before predictable friction points.',
  'persuasive-reframing':
    'Change how the reader interprets the issue by shifting lens, stakes, or causal model so downstream judgments change. Use evidence to justify the new frame rather than asserting it.',
  'persuasive-ethos':
    'Establish trust through fair representation, specific support, and bounded claims so readers credit your judgment under disagreement. Signal judgment quality in how you handle limits and objections.',
  'technical-opening-summary':
    'Open with a compact summary of purpose, recommendation, and expected outcome so readers can place later detail quickly. Give readers a stable map before full development.',
  'technical-screenshot-judgment':
    'Use screenshots only when text cannot communicate state, layout, or error context clearly enough so visuals add real signal. Pair each image with a precise interpretation cue.',
  'academic-pattern-recognition':
    'Move from isolated examples to repeatable structures by naming the pattern, boundary, and implication. Make the rule explicit enough that a reader can apply it to a new case.',
  'independent-voice':
    'Keep source integration subordinate to your own reasoning line so readers can hear your judgment clearly. Make your stance audible in interpretation, transitions, and conclusions rather than citation volume.',
  'audience-alignment':
    'Define who the draft serves and what they need to do or understand next. Shape framing, detail level, and sequencing to match that reader context.',
  positioning:
    'Frame the issue using terms, stakes, and constraints that the target reader already recognizes so relevance lands early. Choose angle and language that makes uptake more likely.',
  'reader-pain-mapping':
    'Name the reader friction that blocks action, trust, or understanding. Tie recommendations directly to that pain so relevance is obvious.',
  'audience-segmentation':
    'Split messaging when reader groups have different context, authority, or goals. Adjust framing and evidence emphasis per segment without changing core logic.',
  'reader-awareness':
    'Anticipate where readers need orientation, reassurance, or proof so momentum does not collapse. Add cues before predictable friction points.',
  'academic-active-voice':
    'Choose active or passive voice by rhetorical function, not habit, so agency stays legible when it matters. Use active voice where agency matters and passive voice where process or object should lead.',
  'tense-consistency':
    'Maintain stable tense within a reasoning unit unless timeline movement is intentional. Mark temporal shifts clearly so readers do not lose sequence.',
  'portable-takeaway':
    'End sections with language readers can reuse outside this draft. Distill the core move into a memorable, transferable phrasing.',
  'story-point-of-view-control':
    'Hold perspective consistently within scene beats unless a deliberate shift is signaled so readers stay anchored in one consciousness. Keep sensory access, inference scope, and diction tied to one viewpoint at a time.',
  'beginning-middle-end':
    'Build a complete movement arc with setup, pressure development, and changed outcome so the section cannot end where it began. Ensure each stage alters what becomes possible next.',
  'revision-consistency-pass':
    'Run a dedicated pass for terminology, naming, and formatting alignment only so repeated references stay stable. Eliminate drift so repeated references mean exactly the same thing everywhere.',
  'revision-format-pass':
    'Run a dedicated pass for headings, list shape, spacing, and visual hierarchy so readers can navigate under time pressure. Make the document navigable under quick scanning conditions.',
  reframing:
    'Change how the reader interprets the issue by shifting lens, stakes, or causal model so downstream judgments change. Use evidence to justify the new frame rather than asserting it.',
  ethos:
    'Establish trust through fair representation, specific support, and bounded claims so readers credit your judgment under disagreement. Signal judgment quality in how you handle limits and objections.',
  'reader-load-control':
    'Reduce processing burden by chunking, ordering, and simplifying non-essential complexity. Keep reader effort focused on the key decision or insight.',
  'front-loaded-summary':
    'Place the essential decision, recommendation, or finding at the top of long documents. Let readers get oriented before entering details.',
  'academic-opening-summary':
    'Open with a compact summary of question, position, and significance so readers can place later detail quickly. Give readers a stable map before full development.',
  'summary-and-detail-balance':
    'Balance compression and depth so readers get both orientation and proof. Alternate overview with targeted detail where it advances the argument.',
  'reference-layout':
    'Design reference material for lookup speed rather than narrative flow so users can find answers at point of need. Use labels, grouping, and indexing cues that support point-of-need retrieval.',
  'email-thread-discipline':
    'Keep long threads coherent by restating ask, owner, and next step at each turn so decisions do not disappear in reply chains. Prevent decision context from getting buried across replies.',
  scannability:
    'Structure content for rapid extraction of key decisions and deadlines so busy readers can act without full linear reading. Make priority information discoverable in seconds.',
  'screenshot-judgment':
    'Use screenshots only when text cannot communicate state, layout, or error context clearly enough so visuals add real signal. Pair each image with a precise interpretation cue.',
  'visual-grouping':
    'Group related items by function so visual structure matches reasoning structure. Use proximity and labeling to reduce misassociation.',
  'table-use':
    'Use tables when comparison, threshold checking, or compact reference is central so readers can evaluate alternatives quickly. Keep columns purposeful and row semantics consistent.',
  'voice-active-voice':
    'Prefer active phrasing when accountability, motion, or decision ownership should be explicit so responsibility stays clear. Use passive constructions selectively for emphasis control.',
  'active-voice-control':
    'Prefer active phrasing when accountability, motion, or decision ownership should be explicit so responsibility stays clear. Use passive constructions selectively for emphasis control.',
  'voice-document-openings':
    'Set tone, purpose, and implied relationship in opening lines so readers know how to interpret stance immediately. Align the opening voice with audience expectations and message stakes.',
  'document-openings':
    'Set tone, purpose, and implied relationship in opening lines so readers know how to interpret stance immediately. Align the opening voice with audience expectations and message stakes.',
  'voice-confidence-without-overclaiming':
    'State conclusions with conviction proportional to evidence strength so confidence reads as earned rather than inflated. Avoid both hedged vagueness and certainty inflation.',
  'confidence-without-overclaiming':
    'State conclusions with conviction proportional to evidence strength so confidence reads as earned rather than inflated. Avoid both hedged vagueness and certainty inflation.',
  'voice-persuasion-with-restraint':
    'Persuade through clear reasoning and concrete support rather than pressure tactics so reader agency remains intact. Keep urgency present without exaggeration.',
  'persuasion-with-restraint':
    'Persuade through clear reasoning and concrete support rather than pressure tactics so reader agency remains intact. Keep urgency present without exaggeration.',
  'memoir-voice-presence':
    'Keep narrative sensibility active through selection, cadence, and interpretation choices so the perspective feels authored and distinct. Let perspective feel intentional rather than generic.',
  'memoir-humor-control':
    'Use humor to sharpen truth or pressure, not to evade vulnerability. Keep tonal shifts aligned with scene stakes.',
  'memoir-repetition-control':
    'Repeat words, motifs, or images only when recurrence adds cumulative meaning. Cut accidental repetition that dilutes force.',
}

function objectiveGoalOverride(node: SkillGraphNode) {
  return OBJECTIVE_GOAL_OVERRIDE_BY_CODE[node.code.toLowerCase()] ?? null
}

function objectiveGoal(node: SkillGraphNode) {
  const override = objectiveGoalOverride(node)
  if (override) {
    return sentence(override)
  }
  const rawDescription = node.description || `Practice ${node.title.toLowerCase()} in a visible, repeatable way`
  const description = sentence(sanitizeProse(rawDescription))
  switch (patternKey(node)) {
    case 'imagery':
      return `${description} ${sentence(
        'Select sensory and figurative details that deepen meaning or pressure, and cut decorative language that stalls movement'
      )}`.trim()
    case 'narrative':
      return `${description} ${sentence(
        'Keep context, pressure, and consequence tied to the present moment so scenes continue to move while explanation is delivered'
      )}`.trim()
    case 'clarity':
      return `${description} ${sentence(
        'State the controlling point early and keep references explicit so readers do not have to reconstruct intent'
      )}`.trim()
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
  if (
    /(exposition|backstory|worldbuilding|history pressure|social structure|object handling|motif|wound pattern|shame handling|body awareness|time awareness|status awareness|fragment handling|reflection|self-myth|question holding|pattern linking|theme through pressure|loyalty pressure|ending resonance|character goal|character need|character contradiction|relationship pressure|subplot|showing action|emotion through action|problem-and-solution|linear structure|braided structure|summary compression)/.test(
      bag
    )
  )
    return 'narrative'
  if (
    /(prompt reading|objective clarity|context first|scaffolding load|pronoun clarity|spec clarity|assumption marking|informational writing basics|plainstyle|summary basics|alignment language|focus under length|tradeoff language)/.test(
      bag
    )
  )
    return 'clarity'
  if (/(abstract noun|selective detail|description focus|metaphor|atmosphere|image freshness)/.test(bag)) return 'imagery'
  if (/(objection|concession|counterargument)/.test(bag)) return 'analysis'
  if (/(hedging|ethical pressure)/.test(bag)) return 'voice'
  if (/(opening focus|letter structure)/.test(bag)) return 'structure'
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
  const specific = whyOverride(node)
  if (specific) {
    return `${sentence(specific)} ${whySpecificTail(node)}`.trim()
  }
  switch (patternKey(node)) {
    case 'imagery':
      return `${sentence(
        'This matters because overgeneral or decorative language weakens emotional precision and dilutes scene force'
      )} ${sentence(
        'Strong image control gives readers concrete anchors while keeping prose aligned with meaning and momentum'
      )} ${whySpecificTail(node)}`.trim()
    case 'narrative':
      return `${sentence(
        'This matters because explanation that overtakes action flattens urgency and makes scenes feel static'
      )} ${sentence(
        'Controlled narrative context keeps readers oriented while preserving momentum and emotional pressure'
      )} ${whySpecificTail(node)}`.trim()
    case 'clarity':
      return `${sentence(
        'This matters because readers decide quickly whether writing is usable, and unclear intent causes immediate friction'
      )} ${sentence(
        'Strong clarity keeps purpose, scope, and references legible without forcing readers to infer missing links'
      )} ${whySpecificTail(node)}`.trim()
    case 'causal':
      return `${sentence(
        'This matters because readers disengage when outcomes feel unearned or disconnected from prior choices'
      )} ${sentence(
        'Strong causal control makes tension feel inevitable and keeps momentum moving forward instead of sideways'
      )} ${whySpecificTail(node)}`.trim()
    case 'scene':
      return `${sentence(
        'This matters because a scene without clear turns or spatial logic feels static even when characters are speaking'
      )} ${sentence(
        'Strong scene structure creates pressure, change, and orientation at the same time'
      )} ${whySpecificTail(node)}`.trim()
    case 'dialogue':
      return `${sentence(
        'This matters because dialogue should move power, intent, or conflict, not simply relay information'
      )} ${sentence(
        'When dialogue carries dramatic work, scenes become sharper and characters feel distinct'
      )} ${whySpecificTail(node)}`.trim()
    case 'pacing':
      return `${sentence(
        'This matters because pacing controls whether readers feel rising pressure or repetitive drag'
      )} ${sentence(
        'Strong pacing keeps important beats visible and prevents emotional flatlines between turning points'
      )} ${whySpecificTail(node)}`.trim()
    case 'thesis':
      return `${sentence(
        'This matters because readers cannot evaluate or trust an argument until the central position is specific and bounded'
      )} ${sentence(
        'Strong thesis control keeps every section aligned to one defensible line of reasoning'
      )} ${whySpecificTail(node)}`.trim()
    case 'evidence':
      return `${sentence(
        'This matters because claims are persuasive only when support is relevant, proportional, and interpreted'
      )} ${sentence(
        'Strong evidence handling turns facts into argument rather than leaving them as disconnected data'
      )} ${whySpecificTail(node)}`.trim()
    case 'source':
      return `${sentence(
        'This matters because source quality and integration determine whether your argument feels credible or borrowed'
      )} ${sentence(
        'Strong source control shows what each source contributes and why your interpretation remains primary'
      )} ${whySpecificTail(node)}`.trim()
    case 'voice':
      return `${sentence(
        'This matters because tone and voice determine whether readers trust the message before they assess the evidence'
      )} ${sentence(
        'Strong audience and voice control keeps writing credible, direct, and context-appropriate'
      )} ${whySpecificTail(node)}`.trim()
    case 'action':
      return `${sentence(
        'This matters because writing fails operationally when readers cannot tell what action is required, by whom, and by when'
      )} ${sentence(
        'Strong action clarity reduces execution errors and decision delays'
      )} ${whySpecificTail(node)}`.trim()
    case 'scan':
      return `${sentence(
        'This matters because readers often scan first and decide in seconds whether a document is usable'
      )} ${sentence(
        'Strong scannability puts critical information where it can be found quickly under time pressure'
      )} ${whySpecificTail(node)}`.trim()
    case 'structure':
      return `${sentence(
        'This matters because structure determines whether readers can follow progression without reconstructing links on their own'
      )} ${sentence(
        'Strong structural control makes complex material readable in one pass'
      )} ${whySpecificTail(node)}`.trim()
    case 'analysis':
      return `${sentence(
        'This matters because analysis quality is the difference between summary and insight'
      )} ${sentence(
        'Strong analytical control shows how evidence leads to interpretation and where limits still apply'
      )} ${whySpecificTail(node)}`.trim()
    case 'mechanics':
      return `${sentence(
        'This matters because sentence-level errors and vague wording break trust and slow comprehension'
      )} ${sentence(
        'Strong mechanics and precision keep attention on ideas instead of avoidable friction'
      )} ${whySpecificTail(node)}`.trim()
    case 'revision':
      return `${sentence(
        'This matters because revision quality determines whether drafts improve in substance or only in surface polish'
      )} ${sentence(
        'Strong revision habits focus effort on high-leverage issues first and produce visible gains'
      )} ${whySpecificTail(node)}`.trim()
    case 'technical':
      return `${sentence(
        'This matters because minor factual or procedural errors can make writing unusable even when the prose reads smoothly'
      )} ${sentence(
        'Strong technical accuracy reduces rework, protects credibility, and prevents downstream execution mistakes'
      )} ${whySpecificTail(node)}`.trim()
    default:
      return `${sentence(
        'This matters because unclear execution forces readers to infer missing logic, sequence, or significance'
      )} ${sentence(
        'Reliable execution makes purpose, consequence, and next action visible on first read'
      )} ${whySpecificTail(node)}`.trim()
  }
}

function successLooksLike(node: SkillGraphNode) {
  const base = (() => {
    switch (patternKey(node)) {
    case 'imagery':
      return [
        'Reader test: A reader can point to specific details that sharpen mood, memory, or pressure rather than merely decorate prose.',
        'Text signal: Sensory and figurative lines are concrete, selective, and tied to narrative or argumentative function.',
        'Failure signal: Description drifts into stock phrasing, abstraction, or image overload that slows the section.',
      ]
    case 'narrative':
      return [
        'Reader test: A reader can explain what is happening now, what context was added, and why it matters to current pressure.',
        'Text signal: Explanatory material appears in short, purposeful inserts tied to immediate conflict or decision.',
        'Failure signal: Context blocks overtake active beats and the section reads as summary instead of unfolding action.',
      ]
    case 'clarity':
      return [
        'Reader test: A reader can state the main point and required interpretation after one pass.',
        'Text signal: Core terms, referents, and scope remain explicit from opening through close.',
        'Failure signal: Readers can repeat the topic but still ask what exactly is meant, requested, or bounded.',
      ]
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
  })()
  return enrichSuccessLooksLike(node, base)
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
  const override = skillOverviewOverride(node)
  if (override) {
    return definitionFromOverview(override, node.title)
  }
  const description = sentence(sanitizeProse(node.description || 'apply this objective clearly and consistently in the draft'))
  return definitionFromOverview(description, node.title)
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
