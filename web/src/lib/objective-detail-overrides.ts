export type ObjectiveDetailOverride = {
  skillOverview?: string
  objectiveGoal?: string
  whyThisObjective?: string
  successLooksLike?: string[]
  goodExample?: string
  badExample?: string
  revisionMoves?: string[]
  assessmentFocus?: string[]
  exampleSources?: Array<{ label: string; url: string }>
  exampleStrategy?: string
}

const ACADEMIC_SOURCES = {
  purdueThesis: {
    label: 'Purdue OWL: Developing Strong Thesis Statements',
    url: 'https://owl.purdue.edu/owl/general_writing/academic_writing/establishing_arguments/index.html',
  },
  purdueParagraphs: {
    label: 'Purdue OWL: Paragraphing',
    url: 'https://owl.purdue.edu/owl/general_writing/academic_writing/paragraphs_and_paragraphing/index.html',
  },
  purdueOutline: {
    label: 'Purdue OWL: Developing an Outline',
    url: 'https://owl.purdue.edu/owl/general_writing/the_writing_process/developing_an_outline/index.html',
  },
  purdueProofreading: {
    label: 'Purdue OWL: Proofreading for Errors',
    url: 'https://owl.purdue.edu/owl/general_writing/the_writing_process/proofreading/proofreading_for_errors.html',
  },
  purdueActivePassive: {
    label: 'Purdue OWL: Active and Passive Voice',
    url: 'https://owl.purdue.edu/owl/general_writing/academic_writing/active_and_passive_voice/index.html',
  },
  purdueQuoting: {
    label: 'Purdue OWL: Quoting, Paraphrasing, and Summarizing',
    url: 'https://owl.purdue.edu/owl/research_and_citation/using_research/quoting_paraphrasing_and_summarizing/index.html',
  },
  uncThesis: {
    label: 'UNC Writing Center: Thesis Statements',
    url: 'https://writingcenter.unc.edu/tips-and-tools/thesis-statements/',
  },
  uncIntroductions: {
    label: 'UNC Writing Center: Introductions',
    url: 'https://writingcenter.unc.edu/tips-and-tools/introductions/',
  },
  uncConclusions: {
    label: 'UNC Writing Center: Conclusions',
    url: 'https://writingcenter.unc.edu/tips-and-tools/conclusions/',
  },
  uncLiteratureReview: {
    label: 'UNC Writing Center: Literature Reviews',
    url: 'https://writingcenter.unc.edu/tips-and-tools/literature-reviews/',
  },
  utorontoCriticalReading: {
    label: 'University of Toronto Writing Advice: Critical Reading',
    url: 'https://advice.writing.utoronto.ca/researching/critical-reading/',
  },
  utorontoLiteratureReview: {
    label: 'University of Toronto Writing Advice: Literature Review',
    url: 'https://advice.writing.utoronto.ca/types-of-writing/literature-review/',
  },
  manchesterHedging: {
    label: 'University of Manchester Phrasebank: Using Cautious Language',
    url: 'https://www.phrasebank.manchester.ac.uk/using-cautious-language/',
  },
}

export const OBJECTIVE_DETAIL_OVERRIDES: Record<string, ObjectiveDetailOverride> = {
  'causal-clarity': {
    skillOverview:
      'This skill helps readers track why each event happens and what it changes. It turns a chain of moments into a chain of consequences.',
    objectiveGoal:
      'Make each major beat readable as choice -> consequence -> next pressure. Keep cause-and-effect links close enough that readers do not have to infer missing steps.',
    whyThisObjective:
      'This matters because story momentum collapses when outcomes feel random or disconnected. Strong control here makes stakes feel earned and keeps readers emotionally invested.',
    successLooksLike: [
      'Readers can explain what decision triggered each major turn.',
      'Big outcomes are prepared by visible earlier actions, not surprise author intervention.',
      'Chronology and causation stay clear even when the narrative moves quickly.',
    ],
    goodExample:
      'Good: Mara burns the ledger to hide her theft; the smoke triggers the alarm; the guard seals the south gate; now her only exit is through the flooded tunnel she cannot swim.',
    badExample:
      'Needs work: Mara is suddenly trapped at the chapter end, but no prior action explains the lockdown, and the danger feels imposed rather than caused by her choices.',
    revisionMoves: [
      'Mark each major turn and write the exact action that causes it.',
      'Add one bridge sentence where a consequence currently appears without setup.',
      'Cut or rewrite outcomes that cannot be traced to earlier decisions.',
    ],
    assessmentFocus: [
      'Every major turn should have a visible trigger inside the draft.',
      'Consequences should escalate pressure rather than reset the story.',
      'Reader orientation should hold without reconstructing hidden steps.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueParagraphs],
    exampleStrategy: 'Teach event logic as an explicit cause-and-effect chain.',
  },
  'scene-architecture': {
    skillOverview:
      'This skill shapes each scene so it has structure, pressure, and movement. It controls how entrances, exits, turns, and power shifts are staged on the page.',
    objectiveGoal:
      'Design scenes with a clear starting pressure, a contested middle, and a changed endpoint. Keep spatial orientation and power movement legible throughout the scene.',
    whyThisObjective:
      'This matters because scenes feel flat when characters speak without goals, resistance, or change. Strong structure makes scenes dramatic, readable, and necessary to the larger story.',
    successLooksLike: [
      'Each scene begins with a clear immediate objective under pressure.',
      'At least one turn changes leverage, information, or options before the scene ends.',
      'Entrances, exits, and object positions remain trackable while tension rises.',
    ],
    goodExample:
      'Good: Eli enters to secure funding, learns the donor already promised support to his rival, then leaves with a forged letter that flips his leverage for the next confrontation.',
    badExample:
      'Needs work: Two characters discuss plans for three pages, but no one pushes against the other, nothing changes by the exit, and the scene could be removed without consequence.',
    revisionMoves: [
      'Write the scene goal, obstacle, and turn in one line each before revision.',
      'Add one action beat where power visibly shifts between characters.',
      'Re-stage movement cues so readers can track who is where during conflict.',
    ],
    assessmentFocus: [
      'Scenes should produce change, not only conversation.',
      'Spatial and dramatic orientation should stay clear under pressure.',
      'Scene endings should alter the next decision landscape.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueParagraphs],
    exampleStrategy: 'Teach scene design as pressure + turn + changed state.',
  },
  'academic-abstract-noun-control': {
    skillOverview:
      'Academic readers need concrete claims, not vague abstraction piles. This skill keeps your nouns precise enough to test and discuss.',
    objectiveGoal: 'Use abstract nouns only when they name a specific, defined concept. Pair each abstraction with concrete evidence or mechanism.',
    whyThisObjective: 'Abstract-heavy prose sounds sophisticated but often hides weak reasoning. Concrete phrasing makes your argument falsifiable and easier to grade fairly.',
    successLooksLike: [
      'Most key nouns can be paraphrased into a concrete process, actor, or effect.',
      'Abstract terms are defined once and then used consistently.',
      'Readers can point to evidence under each major conceptual claim.',
    ],
    goodExample:
      'Good: Institutional trust declined after the new attendance rule added a two-hour verification delay, and students reported missing scholarship deadlines as a result.',
    badExample:
      'Needs work: Institutional trust deteriorated due to policy inefficiencies and procedural disengagement dynamics, creating systemic disaffection across stakeholder communities.',
    revisionMoves: [
      'Underline every abstract noun and ask what measurable event it names.',
      'Replace one vague concept per paragraph with a concrete actor-action-result line.',
      'Add one sentence of evidence after each key conceptual claim.',
    ],
    assessmentFocus: [
      'Abstract terms should be defined and stable, not ornamental.',
      'Claims should stay interpretable without insider jargon.',
      'Concrete examples should carry argumentative load, not decorative filler.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueParagraphs, ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.utorontoCriticalReading],
    exampleStrategy: 'Convert abstract claim language into observable actors, actions, and effects.',
  },
  'academic-abstract-writing': {
    skillOverview:
      'Abstract writing is useful when ideas are difficult, but the structure must still guide the reader. This skill teaches conceptual clarity without flattening nuance.',
    objectiveGoal: 'Present conceptual arguments in ordered steps with explicit transitions. Keep each paragraph tied to one conceptual job.',
    whyThisObjective: 'Dense ideas are not the problem; invisible structure is. Readers stay engaged when they can track conceptual progression.',
    successLooksLike: [
      'Conceptual sections open with a clear function statement.',
      'Transitions explain why the next concept follows the previous one.',
      'The paper can be outlined from topic sentences alone.',
    ],
    goodExample:
      'Good: This section defines legitimacy as consent plus procedural fairness. The next section tests that definition against two policy cases.',
    badExample:
      'Needs work: The paragraph shifts from legitimacy to epistemology to institutional memory without signaling purpose or argumentative sequence.',
    revisionMoves: [
      'Write a one-line function label for each paragraph before revising.',
      'Add bridge sentences that name the logical handoff between sections.',
      'Reorder paragraphs until each one answers the question raised by the prior one.',
    ],
    assessmentFocus: [
      'Section order should feel necessary, not accidental.',
      'Conceptual terms should not change meaning mid-paper.',
      'Readers should predict where the argument is heading after each transition.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueParagraphs, ACADEMIC_SOURCES.uncIntroductions, ACADEMIC_SOURCES.uncConclusions],
    exampleStrategy: 'Structure conceptual writing as a visible sequence of defined moves.',
  },
  'academic-active-voice': {
    skillOverview:
      'Active voice is not a style fad; it is a control tool for clarity and responsibility. This skill helps you choose voice intentionally.',
    objectiveGoal: 'Prefer active constructions when actor and responsibility matter. Use passive only when focus should stay on the process or result.',
    whyThisObjective: 'Unnecessary passive voice hides agency and weakens accountability in academic claims. Intentional voice choice improves precision.',
    successLooksLike: [
      'Most key claims name who did what.',
      'Passive voice appears only where actor identity is irrelevant or unknown.',
      'Methods and findings sections use voice choices consistently.',
    ],
    goodExample:
      'Good: We compared three coding protocols and found protocol B reduced labeling error by 14%.',
    badExample:
      'Needs work: A comparison of coding protocols was conducted, and a reduction in labeling error was observed.',
    revisionMoves: [
      'Circle passive verbs and decide whether the missing actor matters.',
      'Rewrite two high-stakes claims with explicit actor-action structure.',
      'Keep passive voice only when process emphasis is the intended focus.',
    ],
    assessmentFocus: [
      'Voice should match rhetorical intent, not habit.',
      'Agency should be explicit in evaluative and causal claims.',
      'Sentence revisions should increase clarity without distorting meaning.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueActivePassive, ACADEMIC_SOURCES.purdueProofreading],
    exampleStrategy: 'Use active voice for accountability; reserve passive for justified emphasis shifts.',
  },
  'academic-analysis-basics': {
    skillOverview:
      'Analysis is interpretation plus reasoning, not summary. This skill trains you to explain what evidence means and why it matters.',
    objectiveGoal: 'Move from observation to claim through explicit reasoning steps. Distinguish description, interpretation, and implication.',
    whyThisObjective: 'Many drafts summarize sources well but stop before analysis. Graders award depth when inference is visible and defensible.',
    successLooksLike: [
      'Paragraphs include a clear interpretive claim, not only reported facts.',
      'Reasoning links between evidence and claim are explicit.',
      'Alternative interpretations are acknowledged or tested.',
    ],
    goodExample:
      'Good: The attendance increase matters because it appears only after reminder timing changed, suggesting message cadence, not policy severity, drove compliance.',
    badExample:
      'Needs work: Attendance increased by 12%, so the writer concludes the policy worked, without testing timing effects, baseline shifts, or alternate causes.',
    revisionMoves: [
      'Add one because-clause after each evidence sentence.',
      'Separate summary sentences from inference sentences in each paragraph.',
      'Test one plausible alternative explanation and respond to it directly.',
    ],
    assessmentFocus: [
      'Evidence-to-claim logic should be explicit.',
      'Interpretive statements should stay proportionate to evidence.',
      'Counterreadings should be handled fairly, not ignored.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.utorontoCriticalReading, ACADEMIC_SOURCES.uncThesis],
    exampleStrategy: 'Show reasoning as visible steps from observation to interpretation and implication.',
  },
  'academic-analysis-pass': {
    skillOverview:
      'A dedicated analysis pass prevents surface edits from hiding weak logic. This skill gives you a repeatable revision order for depth.',
    objectiveGoal: 'Run a revision pass focused only on analytical strength. Improve inference quality before style polish.',
    whyThisObjective: 'Without a targeted pass, drafts often become smoother but not smarter. Separate passes improve quality faster.',
    successLooksLike: [
      'The draft has explicit edits to reasoning, not only wording.',
      'Claims are narrowed or qualified where evidence is thin.',
      'Revised paragraphs show clearer warrant language.',
    ],
    goodExample:
      'Good: On the analysis pass, the writer cuts two summary-heavy paragraphs and replaces them with one comparison explaining why case B weakens the initial claim.',
    badExample:
      'Needs work: The writer edits transitions and commas but leaves unsupported interpretive jumps unchanged.',
    revisionMoves: [
      'Mark every claim sentence and verify supporting evidence appears within three lines.',
      'Rewrite one paragraph per section to foreground inference over summary.',
      'Add qualifiers where certainty exceeds the available support.',
    ],
    assessmentFocus: [
      'Revision should improve reasoning architecture, not just readability.',
      'Claim scope should tighten where evidence is limited.',
      'Post-revision analysis should be more defensible under questioning.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueProofreading, ACADEMIC_SOURCES.utorontoCriticalReading],
    exampleStrategy: 'Isolate analysis in its own pass and measure changes in reasoning quality.',
  },
  'academic-citation-discipline': {
    skillOverview:
      'Citation discipline is about traceability and intellectual honesty. This skill ensures readers can verify claims and distinguish your ideas from sources.',
    objectiveGoal: 'Cite consistently and position sources precisely. Make attribution accurate at sentence level, not just paragraph level.',
    whyThisObjective: 'Inconsistent citation weakens trust and can trigger integrity concerns. Clear attribution strengthens authority and fairness.',
    successLooksLike: [
      'Every borrowed idea is attributed at the point of use.',
      'Citation format is consistent across the document.',
      'Signal phrases clarify source role before evidence appears.',
    ],
    goodExample:
      'Good: As Nguyen (2023) shows, delayed feedback increases dropout risk in first-year labs; this supports the paper\'s early-alert recommendation.',
    badExample:
      'Needs work: Studies show feedback timing matters. (Citation appears at paragraph end with no source framing.)',
    revisionMoves: [
      'Add source attribution at the exact sentence where outside reasoning appears.',
      'Standardize citation format with one style guide checklist.',
      'Introduce each source with a signal phrase that states relevance.',
    ],
    assessmentFocus: [
      'Attribution should be accurate and local to the borrowed idea.',
      'Citation style should be uniform and complete.',
      'Source framing should clarify why each citation belongs.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueQuoting, ACADEMIC_SOURCES.purdueProofreading],
    exampleStrategy: 'Treat citation as argument infrastructure, not formatting cleanup.',
  },
  'academic-claim-ladder': {
    skillOverview:
      'A claim ladder organizes argument depth from local observations to broader implications. This skill prevents leapfrogging from data to overclaim.',
    objectiveGoal: 'Build claims in increasing scope with explicit support at each rung. Keep global conclusions anchored to local evidence.',
    whyThisObjective: 'Readers reject arguments that jump levels too quickly. Laddered claims feel rigorous and trustworthy.',
    successLooksLike: [
      'Local findings appear before broad conclusions.',
      'Each claim level has evidence and a warrant line.',
      'Top-level claims include scope limits or qualifiers.',
    ],
    goodExample:
      'Good: Classroom interviews show confusion about criteria; therefore this cohort needs rubric pre-briefing, which suggests onboarding changes for similar first-year courses.',
    badExample:
      'Needs work: Two interviews were confusing, so higher education assessment systems are fundamentally broken.',
    revisionMoves: [
      'Label claim scope as local, mid-level, or general in the margin.',
      'Add one warrant sentence between each claim level jump.',
      'Qualify top-level claims with explicit context boundaries.',
    ],
    assessmentFocus: [
      'Scope progression should be gradual and justified.',
      'General claims should not outrun sample size or method.',
      'Warrant language should connect each ladder rung clearly.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.uncThesis, ACADEMIC_SOURCES.utorontoCriticalReading],
    exampleStrategy: 'Stage claims by scope and require evidence at every step upward.',
  },
  'academic-clarity-pass': {
    skillOverview:
      'A clarity pass focuses on reader comprehension before elegance. This skill ensures your argument can be followed in one read.',
    objectiveGoal: 'Revise for ambiguity, reference drift, and unclear sequence. Prioritize reader orientation in every section.',
    whyThisObjective: 'Complex ideas fail when wording obscures the path. Clarity revision increases both fairness and persuasive force.',
    successLooksLike: [
      'Pronoun and term references are unambiguous.',
      'Section transitions reveal direction and purpose.',
      'Readers can summarize each section accurately after one pass.',
    ],
    goodExample:
      'Good: The revised paragraph names the policy, actor, and outcome in the first sentence, then explains mechanism in two short follow-up lines.',
    badExample:
      'Needs work: The revised paragraph still uses "this" and "it" for three different antecedents, forcing repeated re-reading.',
    revisionMoves: [
      'Replace ambiguous pronouns with explicit nouns in key claims.',
      'Add one transition sentence at every section boundary.',
      'Read each paragraph aloud and rewrite lines that require backtracking.',
    ],
    assessmentFocus: [
      'Reference chains should be clear without guessing.',
      'Readers should not need to reconstruct chronology or logic.',
      'Clarity edits should preserve conceptual precision.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueProofreading, ACADEMIC_SOURCES.purdueParagraphs, ACADEMIC_SOURCES.uncIntroductions],
    exampleStrategy: 'Run a dedicated comprehension pass focused on orientation and reference clarity.',
  },
  'academic-close-reading': {
    skillOverview:
      'Close reading treats textual details as evidence for interpretation. This skill trains precision in quoting, observing, and inferring.',
    objectiveGoal: 'Ground interpretation in specific textual features. Explain how diction, structure, or tone supports your claim.',
    whyThisObjective: 'General commentary feels shallow without textual evidence. Close reading demonstrates analytical control.',
    successLooksLike: [
      'Interpretive claims cite precise language or structure.',
      'Quoted details are followed by analytical explanation.',
      'Alternative readings are considered where ambiguity matters.',
    ],
    goodExample:
      'Good: The repeated modal "might" weakens certainty, so the narrator\'s authority appears performative rather than stable.',
    badExample:
      'Needs work: The narrator seems unsure, but the paragraph gives no quoted language, no feature analysis, and no warrant linking text to interpretation.',
    revisionMoves: [
      'Select one short quote per analytical paragraph and annotate key words.',
      'Add a sentence that explains how each quoted feature supports the claim.',
      'Test one competing reading before finalizing interpretation.',
    ],
    assessmentFocus: [
      'Textual evidence should be specific and relevant.',
      'Interpretation should emerge from features, not assumptions.',
      'Analysis should handle ambiguity with discipline.',
    ],
    exampleSources: [ACADEMIC_SOURCES.utorontoCriticalReading, ACADEMIC_SOURCES.purdueQuoting],
    exampleStrategy: 'Anchor every interpretive move in visible textual detail and warrant.',
  },
  'academic-concept-definition': {
    skillOverview:
      'Concept definition establishes the terms your argument depends on. This skill prevents semantic drift and false agreement.',
    objectiveGoal: 'Define key terms in operational language early. Reuse definitions consistently across sections.',
    whyThisObjective: 'Undefined concepts let writers and readers assume different meanings. Stable definitions strengthen argumentative precision.',
    successLooksLike: [
      'Core terms are defined before heavy use.',
      'Definitions include boundaries or exclusions.',
      'Later sections apply terms consistently.',
    ],
    goodExample:
      'Good: In this paper, "institutional trust" means confidence in procedural fairness, not personal approval of leaders.',
    badExample:
      'Needs work: The paper uses "trust" to mean satisfaction, compliance, and legitimacy interchangeably, so readers cannot tell which claim is being evaluated.',
    revisionMoves: [
      'Define each core term in one sentence with scope boundaries.',
      'Add one example showing what the term includes and excludes.',
      'Replace synonym drift with the chosen defined term.',
    ],
    assessmentFocus: [
      'Definitions should be specific enough to test.',
      'Term usage should remain stable across the full paper.',
      'Concept boundaries should reduce, not increase, ambiguity.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.purdueParagraphs],
    exampleStrategy: 'Lock definitions early and enforce consistent usage through revision.',
  },
  'academic-conclusion-function': {
    skillOverview:
      'A conclusion should do analytical work, not repeat the introduction. This skill helps endings deliver consequence and next-step thinking.',
    objectiveGoal: 'Synthesize findings and state implications without introducing unrelated claims. End with a justified final position or question.',
    whyThisObjective: 'Weak conclusions feel abrupt or repetitive. Strong conclusions leave readers with clear intellectual payoff.',
    successLooksLike: [
      'Conclusion synthesizes major lines of reasoning succinctly.',
      'Implications follow from evidence, not speculation.',
      'The final lines extend significance without overclaiming.',
    ],
    goodExample:
      'Good: The findings support targeted mentoring in first-year labs; future work should test whether this holds in remote sections.',
    badExample:
      'Needs work: In conclusion, this topic is important and has many aspects. Thank you for reading.',
    revisionMoves: [
      'Rewrite the first conclusion sentence as a synthesis, not a restatement.',
      'Add one implication that is directly supported by earlier evidence.',
      'Cut any new claims that were not developed in the body.',
    ],
    assessmentFocus: [
      'Conclusion should feel earned by the body argument.',
      'Implications should remain proportionate to evidence.',
      'Ending tone should signal closure plus forward significance.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncConclusions, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'End with synthesis plus bounded implication, not summary filler.',
  },
  'academic-context-use': {
    skillOverview:
      'Context frames why your argument matters and how sources relate. This skill prevents evidence dumps and disconnected background sections.',
    objectiveGoal: 'Introduce background only when it clarifies your claim, method, or stakes. Keep context tightly tied to current argumentative purpose.',
    whyThisObjective: 'Too little context confuses readers; too much buries the argument. Effective context calibrates reader orientation.',
    successLooksLike: [
      'Context appears at points of need, not in one large dump.',
      'Background details connect directly to claim or method.',
      'Readers can see why each source is included.',
    ],
    goodExample:
      'Good: Before analyzing results, the writer briefly explains the district\'s policy timeline so trend changes are interpretable.',
    badExample:
      'Needs work: The paper opens with two pages of historical background that never returns in the argument.',
    revisionMoves: [
      'Cut background lines that do not change interpretation or stakes.',
      'Move essential context next to the claim it enables.',
      'Add one sentence explaining why each context block is necessary.',
    ],
    assessmentFocus: [
      'Context should reduce confusion without slowing argument flow.',
      'Background relevance should be explicit, not implied.',
      'Source framing should connect context to thesis progression.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncLiteratureReview, ACADEMIC_SOURCES.utorontoLiteratureReview],
    exampleStrategy: 'Use context as just-in-time scaffolding for argument comprehension.',
  },
  'academic-counterreading': {
    skillOverview:
      'Counterreading tests your interpretation against credible alternatives. This skill strengthens intellectual fairness and argumentative resilience.',
    objectiveGoal: 'Represent a strong alternative interpretation accurately, then respond with evidence. Avoid straw-man paraphrases.',
    whyThisObjective: 'Arguments become more credible when they engage serious opposition. Ignoring alternatives signals weak analysis.',
    successLooksLike: [
      'Alternative interpretations are stated in good faith.',
      'Response lines address strongest evidence, not weakest phrasing.',
      'Final position shows what was learned from the challenge.',
    ],
    goodExample:
      'Good: A policy-efficiency reading explains the cost trend, but equity data shows the burden shifted disproportionately to commuters.',
    badExample:
      'Needs work: Some people disagree for emotional reasons, so the writer dismisses opposition without restating the strongest alternative claim or evidence.',
    revisionMoves: [
      'Write the strongest opposing claim in neutral language first.',
      'Add one evidence-based response to that exact claim.',
      'Revise your thesis to reflect any valid limit exposed by the counterreading.',
    ],
    assessmentFocus: [
      'Opposition should be represented accurately and respectfully.',
      'Responses should use evidence, not tone, to persuade.',
      'Thesis should be refined when counterevidence reveals limits.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.utorontoCriticalReading],
    exampleStrategy: 'Use counterreading to pressure-test and sharpen your own argument.',
  },
  'academic-draft-outline': {
    skillOverview:
      'A draft outline reveals structural weaknesses before line editing. This skill gives you a fast diagnostic for argument order and balance.',
    objectiveGoal: 'Create a reverse outline from the draft and compare it to intended argument flow. Repair missing or redundant sections.',
    whyThisObjective: 'Writers often cannot see structural drift while drafting. Outlining exposes gaps and repetition quickly.',
    successLooksLike: [
      'Each paragraph has a clear functional label in the outline.',
      'Section order supports cumulative reasoning.',
      'Redundant paragraphs are merged or cut.',
    ],
    goodExample:
      'Good: The reverse outline shows two evidence sections making the same point; the writer merges them and adds a missing counterargument section.',
    badExample:
      'Needs work: The writer line-edits every paragraph without checking whether the structure still supports the thesis.',
    revisionMoves: [
      'Write a one-line purpose statement for each paragraph after drafting.',
      'Reorder sections to match claim progression from setup to implication.',
      'Cut or merge paragraphs that repeat function without adding value.',
    ],
    assessmentFocus: [
      'Outline should reflect actual paragraph jobs, not intended ones.',
      'Structural changes should improve argumentative flow measurably.',
      'Final draft should have fewer redundancy signals than the initial outline.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueOutline, ACADEMIC_SOURCES.purdueParagraphs],
    exampleStrategy: 'Diagnose structure with reverse outlining before sentence-level polish.',
  },
  'academic-evidence-basics': {
    skillOverview:
      'Evidence basics means matching support to claim type and explaining relevance. This skill prevents unsupported assertion and citation dumping.',
    objectiveGoal: 'Pair each major claim with relevant, sufficient support. Explain how evidence advances the argument.',
    whyThisObjective: 'Claims without support are unpersuasive. Support without interpretation is equally weak.',
    successLooksLike: [
      'Major claims are followed by concrete evidence quickly.',
      'Evidence type matches claim type (data, text, case, or theory).',
      'Interpretive sentences connect evidence to thesis.',
    ],
    goodExample:
      'Good: Survey data shows a 19% reduction in withdrawal requests after advising changes, which supports the claim that onboarding, not grading policy, drove retention gains.',
    badExample:
      'Needs work: The policy worked because students liked it, but no dataset, case evidence, or source is provided to support that inference.',
    revisionMoves: [
      'Mark unsupported claims and attach one relevant evidence item to each.',
      'Add one sentence explaining why the evidence is probative.',
      'Remove evidence blocks that do not support the active claim.',
    ],
    assessmentFocus: [
      'Support should be timely, relevant, and sufficient.',
      'Interpretation should accompany evidence in the same paragraph.',
      'Evidence quality should match argument stakes.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.purdueQuoting, ACADEMIC_SOURCES.uncThesis],
    exampleStrategy: 'Treat evidence as reasoning material, not proof decoration.',
  },
  'academic-evidence-proportion': {
    skillOverview:
      'Evidence proportion controls how much support each claim receives. This skill keeps papers balanced between assertion, support, and analysis.',
    objectiveGoal: 'Allocate evidence in proportion to claim importance and controversy. Avoid under-supporting central claims and overloading minor points.',
    whyThisObjective: 'Imbalanced support makes arguments feel arbitrary. Proportional evidence signals judgment and control.',
    successLooksLike: [
      'Core claims have the densest and strongest support.',
      'Minor points stay concise and do not crowd central analysis.',
      'Evidence volume aligns with argumentative weight.',
    ],
    goodExample:
      'Good: The central causation claim receives two datasets and one counterexample test, while a background claim gets one concise citation.',
    badExample:
      'Needs work: A minor definitional note gets a full page of sources while the main claim gets one unsupported sentence.',
    revisionMoves: [
      'Rank claims by importance and compare evidence depth per rank.',
      'Shift one evidence block from a minor point to a central claim.',
      'Condense over-supported side points into brief contextual citations.',
    ],
    assessmentFocus: [
      'Support distribution should reflect argument hierarchy.',
      'Central claims should not be evidence-thin.',
      'Evidence density should not obscure narrative flow.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.utorontoLiteratureReview],
    exampleStrategy: 'Match support density to claim priority and controversy level.',
  },
  'academic-final-resonance': {
    skillOverview:
      'Final resonance is the aftereffect of your conclusion. This skill helps you leave readers with a clear, durable intellectual takeaway.',
    objectiveGoal: 'Close by synthesizing findings and stating a bounded broader implication. Make the final lines memorable without exaggeration.',
    whyThisObjective: 'Strong analysis can be forgotten if the ending is weak. Resonant endings improve retention and perceived coherence.',
    successLooksLike: [
      'Final paragraph links evidence to a meaningful implication.',
      'Last sentence reflects argument identity, not generic closure.',
      'Tone stays measured and avoids rhetorical inflation.',
    ],
    goodExample:
      'Good: If attendance policy design changes who can comply, then compliance rates alone are a misleading measure of policy success.',
    badExample:
      'Needs work: More research is needed and this topic is very important in today\'s world.',
    revisionMoves: [
      'Rewrite the closing sentence to express one precise implication.',
      'Cut generic ending phrases that could fit any paper.',
      'Test whether the final line still makes sense without the body and revise for specificity.',
    ],
    assessmentFocus: [
      'Ending should feel specific to this argument, not reusable boilerplate.',
      'Implication should be supported by prior analysis.',
      'Resonance should come from precision, not overstatement.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncConclusions, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'Land the conclusion with precise implication rather than generic significance claims.',
  },
  'academic-footnote-judgment': {
    skillOverview:
      'Footnote judgment is deciding what belongs in the main line versus notes. This skill protects flow while preserving scholarly precision.',
    objectiveGoal: 'Use footnotes for supportive context, methodological caveats, or archival detail that would disrupt main argument flow.',
    whyThisObjective: 'Overloaded main text slows readers; overloaded footnotes hide key claims. Good judgment keeps argument and apparatus balanced.',
    successLooksLike: [
      'Main argument remains readable without consulting every note.',
      'Notes add value rather than dumping unresolved thoughts.',
      'Critical evidence remains in body text, not buried in notes.',
    ],
    goodExample:
      'Good: The body states the core finding; the footnote gives dataset construction detail for interested specialist readers.',
    badExample:
      'Needs work: The thesis caveat that limits the central claim appears only in a footnote, so body claims sound stronger than warranted.',
    revisionMoves: [
      'Move claim-limiting caveats from notes into body paragraphs.',
      'Shift one technical aside to a note if it breaks paragraph flow.',
      'Cut notes that repeat information already stated in the text.',
    ],
    assessmentFocus: [
      'Reader comprehension should not depend on hidden critical notes.',
      'Notes should support, not replace, argument structure.',
      'Body and notes should each have clear rhetorical roles.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueQuoting, ACADEMIC_SOURCES.purdueParagraphs],
    exampleStrategy: 'Keep the argument in the body; reserve notes for nonessential but useful depth.',
  },
  'academic-hedging-control': {
    skillOverview:
      'Hedging control calibrates certainty in academic claims. This skill helps you avoid both overclaiming and unnecessary vagueness.',
    objectiveGoal: 'Use cautious language where evidence is partial, and decisive language where evidence is strong. Signal uncertainty precisely.',
    whyThisObjective: 'Overstated certainty invites rebuttal. Excessive hedging weakens credibility and clarity.',
    successLooksLike: [
      'Claim strength matches evidence strength consistently.',
      'Qualifiers are purposeful, not habitual filler.',
      'Generalizations include scope limits when needed.',
    ],
    goodExample:
      'Good: These findings suggest mentorship timing is a likely contributor to retention gains in first-year sections.',
    badExample:
      'Needs work: This proves mentorship always determines retention outcomes across all institutions, even though the study only sampled one local program.',
    revisionMoves: [
      'Tag high-certainty verbs and verify evidence supports that strength.',
      'Replace vague hedges with explicit scope limits when possible.',
      'Write one sentence per section that states what remains uncertain.',
    ],
    assessmentFocus: [
      'Certainty markers should track evidence quality and scope.',
      'Hedges should clarify boundaries, not blur meaning.',
      'Strong claims should remain defensible under scrutiny.',
    ],
    exampleSources: [ACADEMIC_SOURCES.manchesterHedging, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'Calibrate certainty deliberately using evidence strength and scope boundaries.',
  },
  'academic-independent-voice': {
    skillOverview:
      'Independent voice means owning your reasoning while engaging sources responsibly. This skill prevents patchwork citation writing.',
    objectiveGoal: 'Maintain a clear argumentative line led by your analysis. Use sources to support or challenge your claim, not replace it.',
    whyThisObjective: 'Academic authority comes from judgment, not citation volume. Readers need to hear your reasoning decisions.',
    successLooksLike: [
      'Paragraphs are driven by your claims, not source order.',
      'Source material is framed and interpreted in your own argumentative language.',
      'The paper\'s stance remains clear through synthesis sections.',
    ],
    goodExample:
      'Good: While Patel emphasizes staffing ratios, this paper argues scheduling design is the stronger lever because variance drops after cadence changes.',
    badExample:
      'Needs work: Smith says X. Lee says Y. Patel says Z. (No authorial position is developed.)',
    revisionMoves: [
      'Start each body paragraph with your claim before citing sources.',
      'Add interpretation sentences after every quotation or paraphrase.',
      'Rewrite one source-heavy paragraph to foreground your line of reasoning.',
    ],
    assessmentFocus: [
      'Your argumentative voice should remain primary throughout.',
      'Source integration should support, not fragment, your thesis line.',
      'Synthesis should show judgment, not compilation.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueQuoting, ACADEMIC_SOURCES.uncLiteratureReview, ACADEMIC_SOURCES.utorontoCriticalReading],
    exampleStrategy: 'Lead with your claim, then position sources as evidence and dialogue partners.',
  },
  'academic-introduction-function': {
    skillOverview:
      'Introductions set problem, stakes, and argumentative direction. This skill ensures readers know what question the paper answers and why it matters.',
    objectiveGoal: 'Open by defining context, problem, and thesis in usable order. Prepare readers for the structure that follows.',
    whyThisObjective: 'Weak introductions lose readers before analysis begins. Strong openings reduce confusion and increase persuasive momentum.',
    successLooksLike: [
      'Introduction presents a specific problem and stakes.',
      'Thesis appears early enough to guide reading.',
      'Preview language aligns with body section order.',
    ],
    goodExample:
      'Good: First-year retention dropped most in commuter cohorts; this paper argues advising cadence, not course rigor, is the primary driver and tests three policy options.',
    badExample:
      'Needs work: Education has always been important throughout human history, and many factors influence student success.',
    revisionMoves: [
      'Draft a three-sentence opener: problem, stakes, thesis.',
      'Move the thesis into the first third of the introduction.',
      'Add one roadmap line that matches actual section order.',
    ],
    assessmentFocus: [
      'Opening should establish urgency without broad cliches.',
      'Thesis placement should support reader orientation.',
      'Roadmap promises should match delivered structure.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncIntroductions, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'Front-load problem, stakes, and thesis to orient readers immediately.',
  },
  'academic-line-editing': {
    skillOverview:
      'Line editing refines sentence-level clarity after structural issues are solved. This skill removes friction while preserving analytical meaning.',
    objectiveGoal: 'Revise sentence rhythm, precision, and redundancy without changing argument logic. Make each line carry its full function.',
    whyThisObjective: 'Good ideas can be buried by noisy prose. Clean lines reduce cognitive load and improve credibility.',
    successLooksLike: [
      'Sentences are concise without becoming abrupt or vague.',
      'Redundant phrasing is removed systematically.',
      'Terminology and tense remain consistent within sections.',
    ],
    goodExample:
      'Good: Because attendance logs were incomplete, we estimated baseline variance using registrar snapshots rather than weekly dashboards.',
    badExample:
      'Needs work: Due to the fact that attendance logs were somewhat incomplete in certain ways, it was necessary to make an estimation attempt.',
    revisionMoves: [
      'Cut filler phrases that do not change sentence meaning.',
      'Replace weak verb-noun pairs with direct verbs.',
      'Read each paragraph aloud and tighten one long sentence per pass.',
    ],
    assessmentFocus: [
      'Line edits should improve clarity without altering claims.',
      'Sentence economy should preserve necessary nuance.',
      'Post-edit prose should be easier to parse on first read.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueProofreading, ACADEMIC_SOURCES.purdueActivePassive],
    exampleStrategy: 'Treat line editing as precision tightening after argument structure is stable.',
  },
  'academic-literature-positioning': {
    skillOverview:
      'Literature positioning places your argument inside existing scholarship. This skill shows what you inherit, contest, and add.',
    objectiveGoal: 'Map key debates and locate your thesis within them. Make your contribution relative to prior work explicit.',
    whyThisObjective: 'Without positioning, papers read as isolated opinion. Clear placement signals scholarly awareness and contribution.',
    successLooksLike: [
      'Core schools or positions are summarized accurately and briefly.',
      'Your stance relative to those positions is explicit.',
      'Contribution language states what changes because of your argument.',
    ],
    goodExample:
      'Good: Prior studies prioritize staffing ratios; this paper extends that frame by testing timing effects within identical staffing levels.',
    badExample:
      'Needs work: Many scholars have written about this issue, but this paper will discuss it further.',
    revisionMoves: [
      'Identify two dominant positions and summarize each in one sentence.',
      'State where your thesis aligns and where it diverges.',
      'Add one contribution sentence that names the paper\'s distinct value.',
    ],
    assessmentFocus: [
      'Literature framing should be accurate, concise, and relevant.',
      'Positioning should clarify novelty without overstating originality.',
      'Contribution claims should be tied to actual analysis.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncLiteratureReview, ACADEMIC_SOURCES.utorontoLiteratureReview],
    exampleStrategy: 'Position the paper by naming prior frames, then stating your distinct intervention.',
  },
  'academic-method-signaling': {
    skillOverview:
      'Method signaling tells readers how evidence was produced and how to interpret confidence. This skill improves transparency and trust.',
    objectiveGoal: 'State method choices, scope limits, and evidence type at the point of interpretation. Make inferential boundaries visible.',
    whyThisObjective: 'Readers cannot evaluate claims without method context. Clear signaling prevents accidental overreach.',
    successLooksLike: [
      'Methods are named in concise, reader-usable terms.',
      'Interpretive claims reference method limits where relevant.',
      'Transitions from method to finding are explicit.',
    ],
    goodExample:
      'Good: Because this analysis uses a single-site case study, findings indicate plausible mechanisms rather than universal prevalence.',
    badExample:
      'Needs work: The data proves this pattern everywhere, but the paragraph never states sample limits, design constraints, or what the method cannot establish.',
    revisionMoves: [
      'Add one method-scope sentence before each major interpretation block.',
      'Tag universal language and replace it with method-appropriate scope.',
      'Link each key finding to the evidence type that supports it.',
    ],
    assessmentFocus: [
      'Method language should calibrate certainty and scope.',
      'Findings should not outrun the design limitations.',
      'Readers should understand what the method can and cannot support.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.purdueQuoting, ACADEMIC_SOURCES.utorontoCriticalReading],
    exampleStrategy: 'Signal method and scope where claims are interpreted, not only in a methods section.',
  },
  'academic-oral-defense': {
    skillOverview:
      'Oral defense readiness means you can explain and defend your choices under live questioning. This skill translates written argument into spoken clarity.',
    objectiveGoal: 'Prepare concise oral explanations of thesis, method, evidence, and limitations. Anticipate likely questions and practice precise responses.',
    whyThisObjective: 'Strong writing can underperform in defense when answers are disorganized. Oral readiness protects your argument under pressure.',
    successLooksLike: [
      'You can state thesis, method, and contribution in under two minutes.',
      'Likely committee objections have prepared evidence-based responses.',
      'You can explain one major limitation without collapsing confidence.',
    ],
    goodExample:
      'Good: When asked about sample bias, the student names the limitation, explains why the design was still fit for purpose, and proposes a next-study correction.',
    badExample:
      'Needs work: The student restates the abstract repeatedly and cannot connect questions to specific evidence decisions.',
    revisionMoves: [
      'Write a two-minute defense script covering thesis, method, and contribution.',
      'List five hard questions and draft evidence-based responses for each.',
      'Practice one limitation answer that includes both boundary and rationale.',
    ],
    assessmentFocus: [
      'Oral responses should stay structured, specific, and evidence-linked.',
      'Defensive tone should be replaced by analytical clarity.',
      'Limitations should be acknowledged without surrendering core claims.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.uncIntroductions, ACADEMIC_SOURCES.utorontoCriticalReading],
    exampleStrategy: 'Convert written argument into concise spoken modules for high-pressure questioning.',
  },
  'academic-paragraph-focus': {
    skillOverview: 'Paragraph focus keeps each paragraph on one argumentative job. This skill prevents drift and hidden topic changes.',
    objectiveGoal: 'Build paragraphs around one claim function at a time. Keep support and analysis tightly scoped to that function.',
    whyThisObjective: 'Focused paragraphs reduce reader load and improve grading clarity for argument quality.',
    successLooksLike: [
      'Topic sentence and support lines serve the same argumentative purpose.',
      'Side ideas are either cut or moved to their own paragraph.',
      'Paragraph endings transition cleanly to the next logical step.',
    ],
    goodExample:
      'Good: This paragraph argues that advising cadence matters, then uses one dataset and one interpretation line to support only that claim.',
    badExample:
      'Needs work: The paragraph starts on advising cadence, shifts to budget ethics, and ends with a methods caveat that belongs elsewhere.',
    revisionMoves: [
      'Write a one-line paragraph purpose statement in the margin.',
      'Cut sentences that do not advance the paragraph purpose.',
      'Move secondary ideas into separate paragraphs with clear topic lines.',
    ],
    assessmentFocus: [
      'Each paragraph should be functionally coherent.',
      'Support should map directly to the topic sentence claim.',
      'Paragraph sequence should build argument progression, not topic hopping.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueParagraphs, ACADEMIC_SOURCES.purdueOutline],
    exampleStrategy: 'Treat paragraphs as single-purpose argument units.',
  },
  'academic-paraphrase-control': {
    skillOverview: 'Paraphrase control means you restate source ideas accurately in your own argumentative language.',
    objectiveGoal: 'Paraphrase with fidelity and compression while preserving key meaning and attribution.',
    whyThisObjective: 'Weak paraphrase risks distortion or patchwriting and undermines source credibility.',
    successLooksLike: [
      'Paraphrases preserve source claim logic without close copying.',
      'Attribution appears where source reasoning is used.',
      'Paraphrased lines are integrated into your argument flow.',
    ],
    goodExample:
      'Good: Rather than quoting the full section, the writer condenses Nguyen\'s mechanism claim and links it directly to the paper\'s policy test.',
    badExample:
      'Needs work: The paragraph swaps a few words from the source sentence but keeps structure and phrasing almost unchanged.',
    revisionMoves: [
      'Summarize source claims from memory before checking wording.',
      'Compare paraphrase against source and restore any lost meaning.',
      'Add attribution and relevance framing in the same sentence block.',
    ],
    assessmentFocus: [
      'Paraphrase should be accurate, original in phrasing, and contextualized.',
      'Source ownership should remain visible.',
      'Paraphrased content should advance your argument, not float independently.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueQuoting, ACADEMIC_SOURCES.purdueProofreading],
    exampleStrategy: 'Reconstruct ideas, then verify fidelity and attribution.',
  },
  'academic-pattern-recognition': {
    skillOverview: 'Pattern recognition in analysis means identifying repeat structures across cases, not isolated anecdotes.',
    objectiveGoal: 'Detect and name recurring relationships in evidence. Distinguish signal from one-off noise.',
    whyThisObjective: 'Argument depth improves when patterns, exceptions, and thresholds are explicitly mapped.',
    successLooksLike: [
      'Multiple data points are compared to reveal a shared trend.',
      'Exceptions are analyzed rather than ignored.',
      'Pattern claims include scope conditions.',
    ],
    goodExample:
      'Good: Across three cohorts, retention gains appear only where reminder timing shifted, indicating timing as the likely common mechanism.',
    badExample:
      'Needs work: One cohort improved, so the writer assumes the same pattern applies universally without cross-case comparison.',
    revisionMoves: [
      'Group evidence into comparable cases before interpreting.',
      'Write one sentence naming the pattern and one sentence naming its boundary.',
      'Test one exception case and explain whether it weakens or refines the pattern.',
    ],
    assessmentFocus: [
      'Pattern claims should rest on repeated evidence, not singular examples.',
      'Boundary conditions should be explicit.',
      'Exception handling should strengthen interpretation discipline.',
    ],
    exampleSources: [ACADEMIC_SOURCES.utorontoCriticalReading, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'Infer patterns by comparing repeated cases and explaining exceptions.',
  },
  'academic-pivot-control': {
    skillOverview: 'Pivot control governs how you shift from one argument segment to another without losing momentum.',
    objectiveGoal: 'Use pivots that explain why the argument is turning and what question the next section answers.',
    whyThisObjective: 'Unsignaled pivots make strong analysis feel disjointed and hard to evaluate.',
    successLooksLike: [
      'Pivot sentences explicitly name the logic of transition.',
      'New sections answer a question raised by prior sections.',
      'Readers can track argument direction without re-reading.',
    ],
    goodExample:
      'Good: Having shown the policy improves attendance, the paper now examines whether gains are equitable across commuter and residential students.',
    badExample:
      'Needs work: The paper jumps from attendance outcomes to citation style discussion with no stated argumentative reason.',
    revisionMoves: [
      'Add one pivot sentence at each major section boundary.',
      'State the next section\'s question in the first paragraph.',
      'Cut digressive pivots that do not serve thesis progression.',
    ],
    assessmentFocus: [
      'Pivots should carry logical function, not filler phrasing.',
      'Section shifts should feel motivated by prior findings.',
      'Transition language should preserve argumentative continuity.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueParagraphs, ACADEMIC_SOURCES.uncConclusions],
    exampleStrategy: 'Frame pivots as explicit argumentative handoffs.',
  },
  'academic-precision-diction': {
    skillOverview: 'Precision diction uses exact terms that reduce ambiguity in analytical writing.',
    objectiveGoal: 'Choose terms with clear referents and methodological fit. Remove inflated or vague wording.',
    whyThisObjective: 'Imprecise diction weakens inference quality and invites misreading.',
    successLooksLike: [
      'Key terms are specific enough to test or operationalize.',
      'Vague intensifiers are replaced by measurable descriptors.',
      'Terminology remains stable across sections.',
    ],
    goodExample:
      'Good: Instead of \"better outcomes,\" the writer specifies \"a 14% drop in withdrawal requests during weeks 3-6.\"',
    badExample:
      'Needs work: The intervention produced major improvements and meaningful engagement across many dimensions, but no metric, mechanism, or timeframe is named.',
    revisionMoves: [
      'Replace one vague adjective per paragraph with a concrete measure or mechanism.',
      'Standardize synonyms so each technical concept has one term.',
      'Flag high-importance nouns and verify each has a clear referent.',
    ],
    assessmentFocus: [
      'Diction should increase interpretive precision.',
      'Word choice should match evidence granularity.',
      'Terminology consistency should support cumulative reasoning.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueProofreading, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'Trade broad labels for operationally precise language.',
  },
  'academic-prompt-reading': {
    skillOverview: 'Prompt reading converts assignment language into a working execution plan.',
    objectiveGoal: 'Parse verbs, scope limits, and deliverables from the prompt before drafting.',
    whyThisObjective: 'Even strong writing underperforms when it misses explicit task requirements.',
    successLooksLike: [
      'Prompt verbs are translated into a revision checklist.',
      'All required components are visible in final structure.',
      'Scope limits are respected throughout the paper.',
    ],
    goodExample:
      'Good: The writer identifies that the prompt requires comparison and evaluation, then builds sections that do both explicitly.',
    badExample:
      'Needs work: The paper analyzes one case deeply but ignores the prompt requirement to compare two competing approaches.',
    revisionMoves: [
      'Highlight directive verbs and required outputs in the prompt.',
      'Map each prompt requirement to a draft section.',
      'Add a final pass that checks prompt coverage before submission.',
    ],
    assessmentFocus: [
      'Draft structure should align with explicit prompt demands.',
      'Prompt constraints should be visible in claim scope.',
      'Final submission should answer all required tasks directly.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueOutline, ACADEMIC_SOURCES.uncThesis],
    exampleStrategy: 'Turn prompt language into traceable drafting and revision criteria.',
  },
  'academic-proofreading': {
    skillOverview: 'Proofreading is a final correctness pass that protects credibility after argument revision is complete.',
    objectiveGoal: 'Run targeted checks for mechanics, formatting, and consistency without re-architecting content.',
    whyThisObjective: 'Mechanical errors can distract readers and reduce trust in otherwise strong analysis.',
    successLooksLike: [
      'Common error patterns are checked with a repeatable checklist.',
      'Grammar and punctuation errors decrease from draft to final.',
      'Formatting and style remain consistent across the document.',
    ],
    goodExample:
      'Good: The final pass catches tense drift, citation punctuation mismatches, and repeated comma splice patterns before submission.',
    badExample:
      'Needs work: The writer submits immediately after major revisions and leaves unresolved agreement errors in key claim sentences.',
    revisionMoves: [
      'Proofread once for one error type at a time.',
      'Read aloud slowly to catch hidden syntax and punctuation issues.',
      'Use a final checklist for citation, formatting, and mechanics consistency.',
    ],
    assessmentFocus: [
      'Proofreading should target known recurring error patterns.',
      'Mechanical control should support uninterrupted reading.',
      'Final polish should not introduce new content changes.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueProofreading, ACADEMIC_SOURCES.purdueParagraphs],
    exampleStrategy: 'Use focused, repeatable proofreading passes by error category.',
  },
  'academic-quotation-restraint': {
    skillOverview: 'Quotation restraint keeps your argument central by using only the source language you need.',
    objectiveGoal: 'Quote selectively and interpret extensively. Keep quoted material proportionate to analytical purpose.',
    whyThisObjective: 'Overquoting shifts authority away from your reasoning and slows argumentative momentum.',
    successLooksLike: [
      'Quotes are short and chosen for analytical necessity.',
      'Most paragraph space is your interpretation.',
      'Quoted material is framed before and after use.',
    ],
    goodExample:
      'Good: The writer quotes one critical clause, then spends the paragraph unpacking its implication for the thesis.',
    badExample:
      'Needs work: Half the section is block quotes, while interpretation is limited to one generic sentence per quote.',
    revisionMoves: [
      'Cut quoted lines that can be paraphrased without loss.',
      'Add two interpretation sentences after each retained quote.',
      'Keep only quotes that carry unique wording you must analyze directly.',
    ],
    assessmentFocus: [
      'Quotation volume should be proportionate to analysis needs.',
      'Your voice should remain primary in argumentative sections.',
      'Quote framing should clarify relevance and scope.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueQuoting, ACADEMIC_SOURCES.uncLiteratureReview],
    exampleStrategy: 'Minimize direct quotation and maximize analytical interpretation.',
  },
  'academic-quote-integration': {
    skillOverview: 'Quote integration places source language smoothly inside your argument structure.',
    objectiveGoal: 'Introduce, embed, and interpret quotes as part of sentence-level reasoning rather than standalone inserts.',
    whyThisObjective: 'Dropped quotes interrupt flow and obscure argumentative ownership.',
    successLooksLike: [
      'Quotes are grammatically integrated with signal phrases.',
      'Context and relevance are clear before quote text appears.',
      'Interpretation follows immediately after quotation.',
    ],
    goodExample:
      'Good: As Ahmed argues, policy \"visibility without usability\" increases nominal compliance while reducing actual completion rates.',
    badExample:
      'Needs work: \"Visibility without usability increases nominal compliance.\" (No source framing or analysis follows.)',
    revisionMoves: [
      'Add a signal phrase before each quote with source role and relevance.',
      'Check quote grammar against your sentence frame.',
      'Write one interpretation sentence that links quote language to your claim.',
    ],
    assessmentFocus: [
      'Quotes should be syntactically and logically integrated.',
      'Signal phrases should clarify source authority and purpose.',
      'Post-quote analysis should carry argumentative work.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueQuoting, ACADEMIC_SOURCES.purdueParagraphs],
    exampleStrategy: 'Integrate quotes as evidence units with framing and interpretation.',
  },
  'academic-revision-triage': {
    skillOverview: 'Revision triage prioritizes high-impact issues before low-impact polish.',
    objectiveGoal: 'Sequence revision from thesis and structure to evidence and line-level style.',
    whyThisObjective: 'Unprioritized revision wastes effort and leaves major weaknesses unresolved.',
    successLooksLike: [
      'Revision plan lists top three high-leverage issues first.',
      'Major claim and structure issues are fixed before copyediting.',
      'Later passes target precision and rhythm after logic is stable.',
    ],
    goodExample:
      'Good: The writer first rewrites thesis scope, then repairs section order, then tightens sentence-level clarity in a final pass.',
    badExample:
      'Needs work: The writer spends an hour adjusting commas while the central claim remains ambiguous and unsupported.',
    revisionMoves: [
      'Rank draft problems by impact on argument comprehension.',
      'Complete one structural pass before any line-level polish.',
      'Document pass goals and verify each pass improved its target issue.',
    ],
    assessmentFocus: [
      'Revision order should reflect pedagogical leverage.',
      'Core argument quality should improve measurably after triage.',
      'Final polish should not substitute for unresolved structure issues.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueProofreading, ACADEMIC_SOURCES.purdueOutline],
    exampleStrategy: 'Use staged revision order to maximize improvement per pass.',
  },
  'academic-rhythm': {
    skillOverview: 'Academic rhythm controls sentence pacing for readability without reducing rigor.',
    objectiveGoal: 'Vary sentence length and cadence to maintain reader attention while preserving precision.',
    whyThisObjective: 'Monotone pacing can obscure strong ideas by increasing processing fatigue.',
    successLooksLike: [
      'Sentence length varies strategically across dense passages.',
      'Key claims are placed in high-emphasis positions.',
      'Cadence supports clarity rather than ornamental complexity.',
    ],
    goodExample:
      'Good: The section alternates compact claim sentences with longer explanatory lines, making the reasoning both clear and sustained.',
    badExample:
      'Needs work: Twelve consecutive long nominalized sentences flatten emphasis and hide the section\'s key claim.',
    revisionMoves: [
      'Break one overloaded sentence in each paragraph into two purposeful lines.',
      'Move key claims to sentence starts or ends for emphasis.',
      'Read dense sections aloud and smooth abrupt or monotonous cadence.',
    ],
    assessmentFocus: [
      'Rhythm should improve accessibility without oversimplification.',
      'Cadence variation should align with argumentative emphasis.',
      'Sentence flow should reduce rereading in dense sections.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueProofreading, ACADEMIC_SOURCES.purdueActivePassive],
    exampleStrategy: 'Shape cadence to support comprehension and emphasis in analytical prose.',
  },
  'academic-scope-control': {
    skillOverview: 'Scope control keeps thesis claims within defensible boundaries across evidence and method limits.',
    objectiveGoal: 'Align claim breadth with evidence strength, method limits, and assignment boundaries.',
    whyThisObjective: 'Overbroad claims are easy to challenge and weaken trust in valid findings.',
    successLooksLike: [
      'Thesis statements include explicit context boundaries.',
      'Generalizations are qualified where evidence is narrow.',
      'Section claims remain consistent with study scope.',
    ],
    goodExample:
      'Good: These results support advising-cadence changes in first-year commuter cohorts at this institution, not all undergraduate populations.',
    badExample:
      'Needs work: The paper draws universal policy conclusions from one small local case study with limited sampling.',
    revisionMoves: [
      'Annotate each major claim with its evidence boundary.',
      'Add qualifiers where claim scope exceeds available support.',
      'Cut or reframe universal language unsupported by method design.',
    ],
    assessmentFocus: [
      'Claim scope should remain method-appropriate.',
      'Boundary language should be explicit and consistent.',
      'Revised thesis should be harder to falsify through scope criticism.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.uncThesis],
    exampleStrategy: 'Constrain thesis breadth to what evidence can actually support.',
  },
  'academic-section-order': {
    skillOverview: 'Section order determines whether readers experience your argument as cumulative or scattered.',
    objectiveGoal: 'Arrange sections so each one answers a question opened by the previous section.',
    whyThisObjective: 'Strong points in weak order feel weaker than they are.',
    successLooksLike: [
      'Section sequence mirrors argumentative dependency.',
      'Earlier sections supply context required by later claims.',
      'Reordering reduces repetition and forward-reference confusion.',
    ],
    goodExample:
      'Good: The paper moves from problem definition to method to findings to implications, and each transition states why this order matters.',
    badExample:
      'Needs work: Implications appear before findings, then methods are introduced late, forcing readers to reinterpret prior claims.',
    revisionMoves: [
      'Create a section map with dependency arrows between claims.',
      'Move sections so evidence appears before high-level implication.',
      'Write transition lines that justify the chosen order explicitly.',
    ],
    assessmentFocus: [
      'Order should reduce inferential gaps and redundancy.',
      'Section dependencies should be visible to the reader.',
      'Revised structure should improve thesis progression clarity.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueOutline, ACADEMIC_SOURCES.purdueParagraphs],
    exampleStrategy: 'Use dependency logic, not drafting chronology, to set section order.',
  },
  'academic-sentence-control': {
    skillOverview: 'Sentence control ensures each sentence is grammatically stable, clear, and purposeful.',
    objectiveGoal: 'Write sentences that name actors, actions, and claims with minimal ambiguity.',
    whyThisObjective: 'Sentence instability compounds quickly in dense academic prose.',
    successLooksLike: [
      'Complex sentences remain syntactically clear and correctly punctuated.',
      'Pronoun and modifier reference stays unambiguous.',
      'Sentence forms support, rather than obscure, argument structure.',
    ],
    goodExample:
      'Good: Because baseline variance was high, the model reports confidence intervals alongside point estimates for each cohort comparison.',
    badExample:
      'Needs work: Because baseline variance was high and model estimates that maybe changed in ways, interpretation became unclear for readers.',
    revisionMoves: [
      'Rewrite one unclear complex sentence per paragraph for explicit actor-action structure.',
      'Check modifier placement near the words they qualify.',
      'Run a punctuation pass focused on clause boundaries and coordination.',
    ],
    assessmentFocus: [
      'Sentence clarity should hold under complexity.',
      'Grammar and punctuation should support interpretive precision.',
      'Edits should reduce ambiguity without flattening nuance.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueProofreading, ACADEMIC_SOURCES.purdueActivePassive],
    exampleStrategy: 'Stabilize complex syntax with explicit structure and boundary control.',
  },
  'academic-significance-lines': {
    skillOverview: 'Significance lines explain why a finding or interpretation matters beyond local detail.',
    objectiveGoal: 'Add concise implication statements that connect evidence to stakes, theory, or decision impact.',
    whyThisObjective: 'Without significance lines, analysis reads as technically correct but directionless.',
    successLooksLike: [
      'Key sections include explicit why-this-matters statements.',
      'Implications follow directly from evidence, not speculation.',
      'Significance language is specific to the paper\'s argument.',
    ],
    goodExample:
      'Good: This matters because retention gains concentrated in commuter cohorts, indicating policy design should target schedule constraints rather than generic motivation.',
    badExample:
      'Needs work: This finding is very important and should be considered by everyone everywhere moving forward.',
    revisionMoves: [
      'Write one implication sentence after each major finding section.',
      'Replace generic importance claims with concrete consequence language.',
      'Check that each significance line is directly evidence-supported.',
    ],
    assessmentFocus: [
      'Significance claims should be precise and warranted.',
      'Implication lines should clarify stakes for specific readers or contexts.',
      'Importance language should avoid broad empty phrasing.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncConclusions, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'Translate findings into bounded implications with explicit stakes.',
  },
  'academic-source-evaluation': {
    skillOverview: 'Source evaluation judges credibility, relevance, and methodological fit before citation use.',
    objectiveGoal: 'Select sources based on authority, recency, evidence quality, and fit with your claim.',
    whyThisObjective: 'Weak sources can contaminate strong reasoning and reduce academic trust.',
    successLooksLike: [
      'Sources are chosen with clear credibility criteria.',
      'Method and context of sources are considered before use.',
      'Low-quality or mismatched sources are excluded or qualified.',
    ],
    goodExample:
      'Good: The writer prioritizes peer-reviewed cohort studies over opinion pieces when making causal claims about retention interventions.',
    badExample:
      'Needs work: A viral blog post is used as primary support for a high-stakes methodological recommendation.',
    revisionMoves: [
      'Rate candidate sources on authority, method quality, and relevance.',
      'Replace one weak source per section with stronger evidence.',
      'Add brief caveat language when using limited or contested sources.',
    ],
    assessmentFocus: [
      'Source quality should match claim stakes.',
      'Credibility judgments should be explicit and consistent.',
      'Evaluation should include methodological fit, not only topical relevance.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncLiteratureReview, ACADEMIC_SOURCES.utorontoLiteratureReview],
    exampleStrategy: 'Apply explicit credibility criteria before integrating sources.',
  },
  'academic-source-pass': {
    skillOverview: 'A source pass is a revision round dedicated to evidence quality and integration fidelity.',
    objectiveGoal: 'Audit source relevance, attribution accuracy, and synthesis quality in one focused pass.',
    whyThisObjective: 'Source issues often persist when revision is unfocused.',
    successLooksLike: [
      'Weak or redundant sources are replaced or removed.',
      'Attribution and citation formatting are corrected consistently.',
      'Source-to-claim alignment improves after the pass.',
    ],
    goodExample:
      'Good: During the source pass, three outdated references are replaced with current peer-reviewed studies and citation placement is corrected sentence by sentence.',
    badExample:
      'Needs work: Sources remain unchanged across revisions even when they are misaligned with revised claims.',
    revisionMoves: [
      'Run a claim-to-source alignment audit for each body section.',
      'Replace outdated or weakly relevant citations with stronger fits.',
      'Correct attribution and citation placement at sentence level.',
    ],
    assessmentFocus: [
      'Source pass should produce measurable evidence-quality gains.',
      'Citation accuracy and alignment should improve simultaneously.',
      'Revised source set should better support thesis scope.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueQuoting, ACADEMIC_SOURCES.purdueProofreading],
    exampleStrategy: 'Use a dedicated revision pass to strengthen source integrity and fit.',
  },
  'academic-source-selection': {
    skillOverview: 'Source selection determines the evidentiary ceiling of your paper and its argumentative credibility.',
    objectiveGoal: 'Choose sources that collectively cover the claim from multiple credible angles.',
    whyThisObjective: 'A narrow or biased source set limits argument reliability and nuance.',
    successLooksLike: [
      'Source set includes high-quality and relevant perspectives.',
      'Selection reflects both supporting and challenging evidence.',
      'Sources are current enough for the claim domain.',
    ],
    goodExample:
      'Good: The writer combines longitudinal data, policy analysis, and qualitative interviews to support a multi-dimensional retention claim.',
    badExample:
      'Needs work: All sources are from one advocacy site with aligned assumptions and no methodological diversity.',
    revisionMoves: [
      'Define source-selection criteria before collecting references.',
      'Add at least one credible source that challenges your preferred interpretation.',
      'Remove sources that duplicate perspective without adding evidence value.',
    ],
    assessmentFocus: [
      'Selection should balance relevance, quality, and perspective range.',
      'Source diversity should improve argumentative robustness.',
      'Chosen sources should map to major claim components.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncLiteratureReview, ACADEMIC_SOURCES.utorontoLiteratureReview, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'Build a deliberately varied, high-credibility source portfolio.',
  },
  'academic-stakes-articulation': {
    skillOverview: 'Stakes articulation explains what changes if your claim is accepted, rejected, or ignored.',
    objectiveGoal: 'State practical, theoretical, or ethical consequences of the argument in concrete terms.',
    whyThisObjective: 'Without clear stakes, even strong analysis can feel academically detached.',
    successLooksLike: [
      'The paper names who is affected and how.',
      'Consequences are specific, not generic significance claims.',
      'Stake statements are evidence-linked and proportionate.',
    ],
    goodExample:
      'Good: If advising cadence remains unchanged, commuter students are likely to bear the largest withdrawal risk, widening equity gaps in first-year persistence.',
    badExample:
      'Needs work: This issue is very important for society and should be studied more in future contexts.',
    revisionMoves: [
      'Write one stakes sentence naming population, impact, and timescale.',
      'Attach each stakes claim to the evidence that supports it.',
      'Cut generic significance statements that do not specify consequence.',
    ],
    assessmentFocus: [
      'Stakes should be concrete, specific, and tied to argument evidence.',
      'Impact language should name affected groups explicitly.',
      'Stake magnitude should remain proportionate to support.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncIntroductions, ACADEMIC_SOURCES.uncConclusions, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'Translate analysis into explicit consequence statements with named impact.',
  },
  'academic-structure-basics': {
    skillOverview: 'Essay structure basics provide the backbone for claim development and reader orientation.',
    objectiveGoal: 'Organize introduction, body, and conclusion so each section has a clear argumentative role.',
    whyThisObjective: 'Foundational structure quality strongly predicts whether readers can follow your reasoning.',
    successLooksLike: [
      'Introduction frames problem and thesis clearly.',
      'Body sections develop claims in logical sequence.',
      'Conclusion synthesizes findings and implications.',
    ],
    goodExample:
      'Good: The essay opens with problem and thesis, develops three evidence-backed claims, and closes with a bounded implication tied to findings.',
    badExample:
      'Needs work: The essay alternates between background and conclusions without a stable body structure or claim progression.',
    revisionMoves: [
      'Outline section roles before drafting full paragraphs.',
      'Check whether each body section advances one distinct claim.',
      'Revise conclusion to synthesize rather than repeat.',
    ],
    assessmentFocus: [
      'Core essay components should be complete and functionally distinct.',
      'Section flow should support progressive reasoning.',
      'Structural clarity should reduce reader backtracking.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueOutline, ACADEMIC_SOURCES.purdueParagraphs, ACADEMIC_SOURCES.uncIntroductions],
    exampleStrategy: 'Use clear section roles to scaffold full argument development.',
  },
  'academic-style-consistency': {
    skillOverview: 'Style consistency maintains a stable scholarly voice and formatting pattern throughout the paper.',
    objectiveGoal: 'Keep tone, citation style, terminology, and formatting conventions consistent across sections.',
    whyThisObjective: 'Style inconsistency distracts readers and weakens perceived control.',
    successLooksLike: [
      'Citation, capitalization, and punctuation patterns are consistent.',
      'Voice and register remain stable across sections.',
      'Terminology and formatting do not drift across revisions.',
    ],
    goodExample:
      'Good: The draft applies one citation style consistently and maintains the same term set for key concepts from introduction through conclusion.',
    badExample:
      'Needs work: The paper alternates citation formats, shifts between informal and formal tone, and renames key concepts by section.',
    revisionMoves: [
      'Run a consistency pass for citation, terminology, and formatting.',
      'Create a style sheet for key terms and apply it globally.',
      'Standardize tone markers so register remains stable throughout.',
    ],
    assessmentFocus: [
      'Consistency should support reader trust and fluency.',
      'Style choices should be deliberate and repeatable.',
      'Terminology drift should be corrected across all sections.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueProofreading, ACADEMIC_SOURCES.purdueQuoting],
    exampleStrategy: 'Use a deliberate style pass to enforce global consistency decisions.',
  },
  'academic-summary-restraint': {
    skillOverview: 'Summary restraint ensures source summary supports analysis instead of replacing it.',
    objectiveGoal: 'Limit summary to context-setting and spend most space on interpretation and argument.',
    whyThisObjective: 'Over-summary signals low analytical depth and reduces originality.',
    successLooksLike: [
      'Summary portions are concise and purpose-driven.',
      'Interpretive lines outnumber descriptive recap lines.',
      'Each summary segment leads directly into analysis.',
    ],
    goodExample:
      'Good: After two summary lines on the study design, the paragraph pivots to evaluating why its sampling choices matter for this thesis.',
    badExample:
      'Needs work: The section retells source content for a page but never states how it changes the paper\'s argument.',
    revisionMoves: [
      'Cap summary to two or three lines per source use.',
      'Add interpretation lines immediately after each summary block.',
      'Cut recap sentences that do not feed a claim or implication.',
    ],
    assessmentFocus: [
      'Summary should serve analysis, not dominate it.',
      'Argument ownership should remain visible after source discussion.',
      'Paragraphs should move from recap to inference efficiently.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueQuoting, ACADEMIC_SOURCES.utorontoCriticalReading],
    exampleStrategy: 'Constrain summary and prioritize interpretive contribution.',
  },
  'academic-synthesis': {
    skillOverview: 'Source synthesis combines multiple sources into one line of reasoning.',
    objectiveGoal: 'Connect agreements, tensions, and gaps across sources to build your own argument.',
    whyThisObjective: 'Source-by-source reporting feels fragmented; synthesis creates scholarly contribution.',
    successLooksLike: [
      'Paragraphs compare sources rather than listing them sequentially.',
      'Synthesis identifies both convergence and disagreement.',
      'Your thesis position is sharpened by cross-source integration.',
    ],
    goodExample:
      'Good: Source A and B agree on retention gains, but C shows gains disappear under delayed advising, which refines the paper\'s mechanism claim.',
    badExample:
      'Needs work: The section summarizes each source separately and ends without comparing findings or stating a synthesized conclusion.',
    revisionMoves: [
      'Group sources by claim relationship, not publication order.',
      'Write one synthesis sentence per cluster naming agreement and tension.',
      'Use synthesis outcomes to revise thesis scope or mechanism claims.',
    ],
    assessmentFocus: [
      'Cross-source reasoning should be explicit and accurate.',
      'Synthesis should produce new argumentative value.',
      'Tension handling should improve thesis precision.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncLiteratureReview, ACADEMIC_SOURCES.utorontoLiteratureReview, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'Synthesize by relationship patterns to generate thesis-level insight.',
  },
  'academic-thesis-clarity': {
    skillOverview: 'Thesis clarity gives readers a specific, arguable central claim to evaluate.',
    objectiveGoal: 'State one precise thesis with clear scope and stakes early in the paper.',
    whyThisObjective: 'A blurred thesis weakens every downstream section, even when evidence is strong.',
    successLooksLike: [
      'Thesis is specific, contestable, and early.',
      'Body sections clearly support thesis components.',
      'Thesis scope matches available evidence and method.',
    ],
    goodExample:
      'Good: This paper argues that advising cadence, not grading policy, explains retention variation in first-year commuter cohorts.',
    badExample:
      'Needs work: This paper explores many issues related to student success and possible improvements.',
    revisionMoves: [
      'Rewrite thesis in one sentence with actor, mechanism, and scope.',
      'Check each body section against thesis support function.',
      'Remove thesis words that overclaim beyond evidence boundaries.',
    ],
    assessmentFocus: [
      'Thesis should be debatable and operationally clear.',
      'Scope should be explicit and defensible.',
      'Section alignment should reinforce thesis intelligibility.',
    ],
    exampleSources: [ACADEMIC_SOURCES.purdueThesis, ACADEMIC_SOURCES.uncThesis],
    exampleStrategy: 'Define a single testable thesis and align structure to it.',
  },
  'academic-title-framing': {
    skillOverview: 'Title framing sets reader expectations and signals argumentative focus before the first paragraph.',
    objectiveGoal: 'Write titles that accurately represent claim scope, method signal, and stakes.',
    whyThisObjective: 'Vague or inflated titles mislead readers and weaken trust in the paper\'s framing discipline.',
    successLooksLike: [
      'Title names the core variable, tension, or claim focus.',
      'Scope in title matches scope in thesis.',
      'Title is specific without becoming jargon-heavy.',
    ],
    goodExample:
      'Good: Advising Cadence and First-Year Retention: Evidence from Commuter Cohorts in a Single-Site Case Study.',
    badExample:
      'Needs work: Thoughts on Education and the Future of Student Experience in Modern Contexts.',
    revisionMoves: [
      'Draft three title options with different emphasis: claim, method, and stakes.',
      'Check title wording against thesis scope and evidence limits.',
      'Cut abstract filler words and keep concrete conceptual anchors.',
    ],
    assessmentFocus: [
      'Title should forecast the paper\'s real argument accurately.',
      'Scope cues should prevent reader overexpectation.',
      'Framing should attract the right audience for the argument.',
    ],
    exampleSources: [ACADEMIC_SOURCES.uncIntroductions, ACADEMIC_SOURCES.purdueThesis],
    exampleStrategy: 'Frame titles as precise promises about argument, scope, and evidence.',
  },
}
