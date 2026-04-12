export type SkillTier = 'core' | 'domain' | 'specialty'

export type SkillDetail = {
  name: string
  slug: string
  tier: SkillTier
  oneLine: string
  whatItMeans: string
  whyItMatters: string
  lookFor: string[]
  strongExample: string
  weakExample: string
  revisionMoves: string[]
  coachTip: string
}

const SKILLS: Array<{ name: string; tier: SkillTier }> = [
  { name: 'clarity and coherence', tier: 'core' },
  { name: 'claim clarity', tier: 'core' },
  { name: 'audience alignment', tier: 'core' },
  { name: 'structural signposting', tier: 'core' },
  { name: 'evidence integration', tier: 'core' },
  { name: 'narrative clarity', tier: 'core' },
  { name: 'scene architecture', tier: 'core' },
  { name: 'actionability', tier: 'core' },
  { name: 'scannability', tier: 'core' },
  { name: 'voice presence', tier: 'core' },
  { name: 'sentence economy', tier: 'domain' },
  { name: 'prose precision', tier: 'domain' },
  { name: 'emotional compression', tier: 'domain' },
  { name: 'dialogue intelligence', tier: 'domain' },
  { name: 'worldbuilding economy', tier: 'domain' },
  { name: 'word choice', tier: 'domain' },
  { name: 'sentence variety', tier: 'domain' },
  { name: 'sentence complexity', tier: 'domain' },
  { name: 'paragraph control', tier: 'domain' },
  { name: 'narrative sequencing', tier: 'domain' },
  { name: 'descriptive precision', tier: 'domain' },
  { name: 'dialogue basics', tier: 'domain' },
  { name: 'insight density', tier: 'domain' },
  { name: 'authority and voice', tier: 'domain' },
  { name: 'tone calibration', tier: 'domain' },
  { name: 'accuracy', tier: 'domain' },
  { name: 'analysis depth', tier: 'domain' },
  { name: 'assignment alignment', tier: 'domain' },
  { name: 'example quality', tier: 'domain' },
  { name: 'grammar control', tier: 'domain' },
  { name: 'objection handling', tier: 'domain' },
  { name: 'professional format', tier: 'domain' },
  { name: 'reasoning quality', tier: 'domain' },
  { name: 'reflection depth', tier: 'domain' },
  { name: 'revision habits', tier: 'domain' },
  { name: 'rhetorical force', tier: 'domain' },
  { name: 'source handling', tier: 'domain' },
  { name: 'spelling and mechanics', tier: 'domain' },
  { name: 'story development', tier: 'domain' },
  { name: 'structure and pacing', tier: 'domain' },
  { name: 'technical precision', tier: 'domain' },
  { name: 'thesis clarity', tier: 'domain' },
  { name: 'user goal alignment', tier: 'domain' },
  { name: 'image freshness', tier: 'domain' },
  { name: 'lineation', tier: 'domain' },
  { name: 'sonic patterning', tier: 'domain' },
  { name: 'image logic', tier: 'domain' },
  { name: 'stanza movement', tier: 'domain' },
  { name: 'visual exposition', tier: 'domain' },
  { name: 'beat design', tier: 'domain' },
  { name: 'act structure', tier: 'domain' },
  { name: 'oral cadence', tier: 'domain' },
  { name: 'rhetorical repetition', tier: 'domain' },
  { name: 'audience energy', tier: 'domain' },
  { name: 'microcopy clarity', tier: 'domain' },
  { name: 'error-state guidance', tier: 'domain' },
  { name: 'information scent', tier: 'domain' },
  { name: 'symbolic control', tier: 'specialty' },
  { name: 'mythic tone', tier: 'specialty' },
  { name: 'tragic inevitability', tier: 'specialty' },
]

function slugify(name: string) {
  return name
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function titleCase(name: string) {
  return name.replace(/\b\w/g, (letter) => letter.toUpperCase())
}

function genericDetail(name: string, tier: SkillTier): SkillDetail {
  const title = titleCase(name)
  return {
    name,
    slug: slugify(name),
    tier,
    oneLine: `${title} helps readers understand your writing faster.`,
    whatItMeans: `In plain words: make ${name} clear and steady from start to end.`,
    whyItMatters: 'When this is weak, readers work too hard. When it is strong, readers stay with you.',
    lookFor: [
      'Can a reader explain your point after one read?',
      'Do your sentences and choices support the same goal?',
      'Is there one clear next step for the reader?',
    ],
    strongExample: 'Strong: "The bus was late, so Maya called her mom. She arrived safe."',
    weakExample: 'Weak: "Things happened and it was kind of bad and stuff changed."',
    revisionMoves: [
      'Cut one vague sentence and replace it with a concrete line.',
      'Name who did what in each key sentence.',
      'Read it out loud and fix any part that feels confusing.',
    ],
    coachTip: 'Aim for simple, clear lines. Clear is strong.',
  }
}

function applyPattern(detail: SkillDetail): SkillDetail {
  const name = detail.name
  const withPatch = (patch: Partial<SkillDetail>): SkillDetail => ({ ...detail, ...patch })
  const has = (value: string) => name.includes(value)

  if (has('evidence') || has('source')) {
    return withPatch({
      oneLine: 'This skill makes your points believable.',
      whatItMeans: 'Make a claim, then back it up with proof the reader can see.',
      whyItMatters: 'Without support, even good ideas feel weak.',
      strongExample: 'Strong: "The plan cut wait time by 30%, from 10 minutes to 7."',
      weakExample: 'Weak: "The plan is way better. Trust me."',
      revisionMoves: [
        'Underline each claim. Add one fact, quote, or example under it.',
        'Keep only support that directly matches the claim.',
        'Explain in one line how the evidence proves the point.',
      ],
      coachTip: 'No naked claims. Every big claim needs support.',
    })
  }

  if (has('clarity and coherence')) {
    return withPatch({
      oneLine: 'This skill keeps your main point easy to follow.',
      whatItMeans: 'Make each paragraph connect to one clear through-line.',
      whyItMatters: 'Readers trust writing they can track without re-reading.',
      strongExample: 'Strong: "Costs rose 18%, so we cut scope and shipped on time."',
      weakExample: 'Weak: "A lot changed, and the project ended up different somehow."',
      revisionMoves: [
        'Name the main point at the top of each section.',
        'Cut lines that do not support the section goal.',
        'Link each section to the next with one bridge sentence.',
      ],
      coachTip: 'If readers ask "wait, what is the point?", coherence is slipping.',
    })
  }

  if (has('assignment alignment')) {
    return withPatch({
      oneLine: 'This skill keeps the draft aligned to the assignment ask.',
      whatItMeans: 'Answer the prompt directly before adding extra context.',
      whyItMatters: 'Strong writing can still fail when it misses the task.',
      strongExample: 'Strong: "The brief asks for two risks, so this section ranks and explains both."',
      weakExample: 'Weak: "The writing is polished, but it answers a different question."',
      revisionMoves: [
        'Write the exact prompt asks into a checklist.',
        'Mark where each ask is answered in the draft.',
        'Cut strong paragraphs that do not serve the prompt.',
      ],
      coachTip: 'First satisfy the assignment contract, then optimize style.',
    })
  }

  if (has('dialogue')) {
    return withPatch({
      oneLine: 'This skill makes spoken lines sound real and useful.',
      whatItMeans: 'Use dialogue to reveal want, conflict, or mood, not just filler chat.',
      whyItMatters: 'Flat dialogue slows scenes. Sharp dialogue moves scenes.',
      strongExample: 'Strong: "Give me the key." "No. You lied last time."',
      weakExample: 'Weak: "Hello." "Hi." "How are you?" "Fine."',
      revisionMoves: [
        'Give each speaker a clear goal in the scene.',
        'Cut greetings and repeat lines that add no pressure.',
        'Tag only when the speaker is unclear.',
      ],
      coachTip: 'Each line should change something.',
    })
  }

  if (has('scene') || has('beat') || has('act structure') || has('structure and pacing')) {
    return withPatch({
      oneLine: 'This skill gives your story shape and momentum.',
      whatItMeans: 'Build clear steps: setup, turn, and result. Keep the pace moving.',
      whyItMatters: 'If shape is fuzzy, readers lose interest fast.',
      strongExample: 'Strong: "He enters late, sees the empty chair, and knows he lost his chance."',
      weakExample: 'Weak: "Stuff happens for a while, then it just ends."',
      revisionMoves: [
        'Mark scene start, turn, and end in the draft.',
        'Cut one slow paragraph before each key turn.',
        'Raise stakes in the middle of the scene.',
      ],
      coachTip: 'Every scene needs a turn.',
    })
  }

  if (has('narrative')) {
    return withPatch({
      oneLine: 'This skill keeps the story easy to follow.',
      whatItMeans: 'Show clear cause and effect from one moment to the next.',
      whyItMatters: 'Readers stay engaged when they can track what changed and why.',
      strongExample: 'Strong: "She missed the bus, so she ran and arrived out of breath."',
      weakExample: 'Weak: "She was late. Then she was at school. Then she cried."',
      revisionMoves: [
        'Add "because/so" logic between key beats.',
        'Name the choice that causes each big result.',
        'Reorder beats so time moves forward cleanly.',
      ],
      coachTip: 'If a reader asks "why did that happen?", add a causal line.',
    })
  }

  if (has('story development')) {
    return withPatch({
      oneLine: 'This skill builds meaningful change across the story.',
      whatItMeans: 'Give characters goals, pressure, and consequences that escalate.',
      whyItMatters: 'Stories feel flat when events do not force harder choices.',
      strongExample: 'Strong: "She hides the letter, loses her ally, and must face the hearing alone."',
      weakExample: 'Weak: "Scenes happen, but nothing changes the character path."',
      revisionMoves: [
        'Mark each scene with a goal, obstacle, and consequence.',
        'Raise stakes after each major decision point.',
        'Cut scenes that do not change character or plot pressure.',
      ],
      coachTip: 'If the same person leaves every scene unchanged, raise pressure.',
    })
  }

  if (has('worldbuilding economy')) {
    return withPatch({
      oneLine: 'This skill reveals world detail without slowing the story.',
      whatItMeans: 'Introduce setting facts only when they affect current action.',
      whyItMatters: 'Dense lore blocks momentum when readers need immediate stakes.',
      strongExample: 'Strong: "The gate tax matters now because it strips her last coin before escape."',
      weakExample: 'Weak: "Three pages of history appear before the first conflict starts."',
      revisionMoves: [
        'Move lore details next to the action they change.',
        'Cut world facts that have no scene consequence.',
        'Show one setting rule through character choice.',
      ],
      coachTip: 'World detail should create pressure, not pause it.',
    })
  }

  if (has('emotional compression')) {
    return withPatch({
      oneLine: 'This skill shows emotion through action and image.',
      whatItMeans: 'Replace emotion labels with gesture, diction, and consequence.',
      whyItMatters: 'Readers feel more when emotion is dramatized, not announced.',
      strongExample: 'Strong: "He folds the letter twice, then tears only his own signature."',
      weakExample: 'Weak: "He was very sad and deeply upset about everything."',
      revisionMoves: [
        'Replace one emotion label with a physical action.',
        'Add one concrete sensory detail to carry feeling.',
        'Cut adverbs that explain emotion the scene already shows.',
      ],
      coachTip: 'Name emotion less often; stage it more clearly.',
    })
  }

  if (has('word choice') || has('prose precision') || has('descriptive') || has('image')) {
    return withPatch({
      oneLine: 'This skill makes your writing vivid and exact.',
      whatItMeans: 'Pick specific words and details instead of broad or blurry ones.',
      whyItMatters: 'Specific words help readers see and feel the moment.',
      strongExample: 'Strong: "Rain tapped the metal roof like fast fingers."',
      weakExample: 'Weak: "It was very nice and kind of interesting."',
      revisionMoves: [
        'Replace one vague adjective with a concrete noun or verb.',
        'Cut filler words like very, really, kind of.',
        'Choose one detail that carries mood.',
      ],
      coachTip: 'Specific beats generic.',
    })
  }

  if (has('sentence') || has('paragraph') || has('signposting') || has('scannability') || has('microcopy') || has('information scent')) {
    return withPatch({
      oneLine: 'This skill helps readers move through your writing with less effort.',
      whatItMeans: 'Keep units short, ordered, and easy to scan.',
      whyItMatters: 'Good structure lowers reader effort and raises understanding.',
      strongExample: 'Strong: "Step 1: Save your file. Step 2: Restart the app."',
      weakExample: 'Weak: "You should maybe save first and also restart and maybe check settings."',
      revisionMoves: [
        'Split long blocks into short paragraphs or bullets.',
        'Start each section with its main point.',
        'Move extra detail below the key instruction.',
      ],
      coachTip: 'Put the most important line first.',
    })
  }

  if (has('claim') || has('thesis') || has('reasoning') || has('analysis') || has('insight') || has('objection')) {
    return withPatch({
      oneLine: 'This skill makes your argument clear and strong.',
      whatItMeans: 'State one clear point, explain your logic, and handle pushback.',
      whyItMatters: 'Readers trust writing that is clear, fair, and well reasoned.',
      strongExample: 'Strong: "School should start later because sleep improves focus and attendance."',
      weakExample: 'Weak: "School start times are just bad and everyone knows it."',
      revisionMoves: [
        'Write your main claim in one short sentence.',
        'Add one reason and one piece of support for each key claim.',
        'Answer one likely reader objection directly.',
      ],
      coachTip: 'Clear claim, clear reason, clear support.',
    })
  }

  if (has('example quality')) {
    return withPatch({
      oneLine: 'This skill picks examples that directly prove your point.',
      whatItMeans: 'Choose evidence that matches scope, audience, and claim.',
      whyItMatters: 'Weak examples make strong arguments look ungrounded.',
      strongExample: 'Strong: "Use Q2 support logs because the proposal changes support staffing."',
      weakExample: 'Weak: "Use any interesting anecdote, even if it is unrelated."',
      revisionMoves: [
        'Link each example to one specific claim.',
        'Replace broad anecdotes with scoped, relevant cases.',
        'Explain in one line why the example is representative.',
      ],
      coachTip: 'The best example is not the loudest one; it is the tightest fit.',
    })
  }

  if (has('reflection depth')) {
    return withPatch({
      oneLine: 'This skill turns reflection into specific learning.',
      whatItMeans: 'Describe what changed, why it changed, and what to repeat.',
      whyItMatters: 'Reflection without specifics does not improve future drafts.',
      strongExample: 'Strong: "Cutting the opener raised clarity, so next time I will front-load the ask."',
      weakExample: 'Weak: "I learned a lot and will keep improving somehow."',
      revisionMoves: [
        'Name one concrete change and its outcome.',
        'Link one mistake to a repeatable revision rule.',
        'Write one measurable goal for the next draft.',
      ],
      coachTip: 'Good reflection produces a next-draft rule, not a mood.',
    })
  }

  if (has('lineation')) {
    return withPatch({
      oneLine: 'This skill uses line breaks to control emphasis and pace.',
      whatItMeans: 'Break lines where sound, image, or syntax gains force.',
      whyItMatters: 'Random breaks make poems feel accidental instead of intentional.',
      strongExample: 'Strong: "I kept the key / not for escape / but for proof."',
      weakExample: 'Weak: "I kept / the key not / for escape but for / proof."',
      revisionMoves: [
        'Read each line break aloud for pressure and pause.',
        'Move weak breaks to words with stronger stress.',
        'Keep syntactic units clear unless fracture is intentional.',
      ],
      coachTip: 'A line break should add meaning, not just shape.',
    })
  }

  if (has('sonic patterning')) {
    return withPatch({
      oneLine: 'This skill shapes sound to reinforce meaning.',
      whatItMeans: 'Use rhythm, echo, and contrast without forcing rhyme.',
      whyItMatters: 'Sound pattern can carry emotion before logic catches up.',
      strongExample: 'Strong: "Cold coins clicked in the cup; the clock kept counting."',
      weakExample: 'Weak: "Words are random in sound and flatten the line."',
      revisionMoves: [
        'Read the draft aloud and mark dull stretches.',
        'Add light sound echoes on key words, not every line.',
        'Cut accidental tongue-twisters that block flow.',
      ],
      coachTip: 'Sound should support the moment, not distract from it.',
    })
  }

  if (has('stanza movement')) {
    return withPatch({
      oneLine: 'This skill makes each stanza advance thought or tension.',
      whatItMeans: 'Treat stanza breaks as turns, not decorative spacing.',
      whyItMatters: 'Without movement, stanzas feel like chopped paragraphs.',
      strongExample: 'Strong: "Stanza one names the wound; stanza two tests a response; stanza three reframes it."',
      weakExample: 'Weak: "Each stanza repeats the same idea with new line breaks."',
      revisionMoves: [
        'Name the function of each stanza in one phrase.',
        'Cut or merge stanzas that repeat without turning.',
        'Start each new stanza with a fresh pressure point.',
      ],
      coachTip: 'A stanza break should change direction, depth, or stakes.',
    })
  }

  if (has('visual exposition')) {
    return withPatch({
      oneLine: 'This skill explains visual information with clear narrative focus.',
      whatItMeans: 'Guide readers through what to look at and why it matters.',
      whyItMatters: 'Unfocused description makes visuals harder, not easier, to read.',
      strongExample: 'Strong: "The peak at week six marks the failed rollout, then recovery begins after patch 2."',
      weakExample: 'Weak: "The chart has many lines and looks complicated overall."',
      revisionMoves: [
        'Start with the one visual pattern that matters most.',
        'Mark the moment where the trend changes direction.',
        'Link the visual pattern to the decision it supports.',
      ],
      coachTip: 'Describe visuals in decision order, not scan order.',
    })
  }

  if (has('tone') || has('voice') || has('audience') || has('rhetorical') || has('oral cadence') || has('audience energy')) {
    return withPatch({
      oneLine: 'This skill helps your writing sound right for your reader.',
      whatItMeans: 'Match your words and tone to your audience and purpose.',
      whyItMatters: 'Good tone builds trust. Bad tone pushes readers away.',
      strongExample: 'Strong: "I know this delay is frustrating. Here is the fix and timeline."',
      weakExample: 'Weak: "Calm down. It is not a big deal."',
      revisionMoves: [
        'Name who the reader is before revising.',
        'Cut lines that sound vague, harsh, or fake.',
        'Use direct, respectful wording for hard points.',
      ],
      coachTip: 'Sound like a calm, helpful human.',
    })
  }

  if (has('actionability') || has('user goal') || has('error-state') || has('professional format')) {
    return withPatch({
      oneLine: 'This skill tells readers exactly what to do next.',
      whatItMeans: 'Give clear next steps, owners, and timing.',
      whyItMatters: 'Action beats confusion.',
      strongExample: 'Strong: "Click Save, then email the file to Ana by 3 PM."',
      weakExample: 'Weak: "Please handle this soon when you can."',
      revisionMoves: [
        'Add one clear action verb to each step.',
        'Name owner and due time where needed.',
        'Remove soft words like soon, maybe, kind of.',
      ],
      coachTip: 'If no one knows what to do next, actionability is weak.',
    })
  }

  if (has('grammar') || has('spelling') || has('mechanics') || has('technical precision') || has('accuracy')) {
    return withPatch({
      oneLine: 'This skill keeps your writing correct and dependable.',
      whatItMeans: 'Use correct form, facts, and terms so readers can trust the text.',
      whyItMatters: 'Small errors can break trust fast.',
      strongExample: 'Strong: "The update shipped on April 4 at 2:00 PM UTC."',
      weakExample: 'Weak: "Update maybe shipped last week around 2ish."',
      revisionMoves: [
        'Run one pass only for facts, names, and numbers.',
        'Run one pass only for spelling and punctuation.',
        'Standardize terms so the same thing has one name.',
      ],
      coachTip: 'Correct details build trust.',
    })
  }

  if (has('revision')) {
    return withPatch({
      oneLine: 'This skill helps you improve the draft in smart passes.',
      whatItMeans: 'Fix big issues first, then polish small issues.',
      whyItMatters: 'Order matters. Better order gives faster gains.',
      strongExample: 'Strong: "Pass 1 claim, pass 2 structure, pass 3 wording."',
      weakExample: 'Weak: "Spend an hour changing commas before fixing the main point."',
      revisionMoves: [
        'Do one pass for meaning, one for structure, one for polish.',
        'Keep a short checklist and use it every draft.',
        'Track one repeat issue and fix it on purpose.',
      ],
      coachTip: 'Do not polish a weak structure.',
    })
  }

  if (has('symbolic control') || has('mythic tone') || has('tragic inevitability')) {
    return withPatch({
      oneLine: 'This skill builds deeper theme and mood.',
      whatItMeans: 'Use repeated images and tone to hint at bigger meaning.',
      whyItMatters: 'Strong theme gives your work lasting weight.',
      strongExample: 'Strong: "The cracked clock appears each time he avoids the truth."',
      weakExample: 'Weak: "Random symbols appear but do not connect to the story."',
      revisionMoves: [
        'Pick one image and repeat it with purpose.',
        'Link tone choices to character change.',
        'Cut symbols that do not support the core conflict.',
      ],
      coachTip: 'Theme should grow from story action, not sit on top of it.',
    })
  }

  return detail
}

function buildSkillDetail(name: string, tier: SkillTier): SkillDetail {
  return applyPattern(genericDetail(name, tier))
}

const DETAILS = SKILLS.map((item) => buildSkillDetail(item.name, item.tier)).sort((a, b) => a.name.localeCompare(b.name))
const DETAIL_BY_NAME = new Map(DETAILS.map((item) => [item.name.toLowerCase(), item]))
const DETAIL_BY_SLUG = new Map(DETAILS.map((item) => [item.slug, item]))

export const allSkillDetails = DETAILS

export function skillSlug(name: string) {
  return slugify(name)
}

export function skillHref(name: string) {
  return `/skills/${skillSlug(name)}`
}

export function getSkillDetailByName(name: string) {
  return DETAIL_BY_NAME.get(name.trim().toLowerCase())
}

export function getSkillDetailBySlug(slug: string) {
  return DETAIL_BY_SLUG.get(slug.trim().toLowerCase())
}

export function hasSkillDetail(name: string) {
  return DETAIL_BY_NAME.has(name.trim().toLowerCase())
}
