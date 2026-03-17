package domain

import "sort"

const GlobalSkillGraphSlug = "global-writing-skill-graph"
const GlobalSkillGraphTitle = "Writing Skill Graph"

type SkillGraphRegion struct {
	Slug           string
	Title          string
	Description    string
	SeedCodes      []string
	PrioritySkills []string
	NodeCodes      []string
}

type SkillGraphNode struct {
	TGO
	SourceTreeSlug  string
	SourceTreeTitle string
	Unlocks         []string
}

type SkillGraph struct {
	Slug        string
	Title       string
	Description string
	Regions     []SkillGraphRegion
	Nodes       []SkillGraphNode
}

func buildBuiltInCatalog() (TGOTreeDefinition, TGOTreeDefinition, TGOTreeDefinition, TGOTreeDefinition, TGOTreeDefinition, TGOTreeDefinition, TGOTreeDefinition, TGOTreeDefinition, TGOTreeDefinition, []TGOTreeDefinition, map[string]string) {
	skillMap := map[string]string{}
	node := func(code, title, skill, description, stage string, order int, mastery string, prerequisites ...string) TGO {
		skillMap[code] = skill
		return TGO{
			Code:          code,
			Title:         title,
			Description:   description,
			Stage:         stage,
			StageOrder:    order,
			Prerequisites: append([]string(nil), prerequisites...),
			MasteryHint:   mastery,
			ProgressMode:  InferProgressMode(skill),
		}
	}

	mythic := TGOTreeDefinition{
		Slug:           "mythic-tragedy-apprenticeship",
		Title:          WriterTrackName,
		Description:    "Advanced mythopoeic tragic fiction track.",
		SeedCodes:      []string{"causal-clarity", "scene-architecture", "prose-precision"},
		PrioritySkills: []string{"tragic inevitability", "symbolic control", "mythic tone", "emotional compression", "scene architecture", "narrative clarity", "prose precision", "worldbuilding economy", "dialogue intelligence", "image freshness"},
		TGOs: []TGO{
			node("causal-clarity", "Causal Clarity", "narrative clarity", "Make action and consequence legible beat by beat.", "core", 1, "Readers can follow each decision and its immediate consequence without backtracking."),
			node("scene-architecture", "Scene Architecture", "scene architecture", "Stage turns, entrances, exits, and power shifts cleanly.", "core", 2, "Scenes remain spatially and dramatically legible even under pressure."),
			node("prose-precision", "Prose Precision", "prose precision", "Replace soft modifiers with exact nouns and verbs.", "core", 3, "Line-level revision consistently sharpens verbs, nouns, and cadence without puffing the diction."),
			node("emotional-compression", "Emotional Compression", "emotional compression", "Condense feeling into image, gesture, and consequence.", "core", 4, "Feeling arrives through action and image rather than named emotion.", "causal-clarity"),
			node("tension-escalation", "Tension Escalation", "scene architecture", "Increase pressure within a scene rather than restating stakes.", "core", 5, "Each beat leaves the scene tighter than before.", "scene-architecture"),
			node("point-of-view-discipline", "Point of View Discipline", "narrative clarity", "Hold a stable lens so emotional and causal information lands cleanly.", "core", 6, "The narrative lens stays coherent and the reader always knows whose pressure is primary.", "causal-clarity"),
			node("gesture-as-subtext", "Gesture as Subtext", "emotional compression", "Use movement and physical response to carry what characters cannot say.", "core", 7, "Gesture consistently reveals the pressure beneath speech.", "emotional-compression"),
			node("motivation-pressure", "Motivation Pressure", "tragic inevitability", "Make motives force decisions rather than merely explain them.", "core", 8, "Choices feel driven from within the character's need structure.", "causal-clarity"),
			node("consequence-memory", "Consequence Memory", "narrative clarity", "Carry forward the effects of prior scenes instead of resetting the emotional field.", "core", 9, "Earlier consequences continue to shape later action.", "causal-clarity"),
			node("line-level-rhythm", "Line-Level Rhythm", "prose precision", "Control cadence so the prose can tighten or broaden intentionally.", "core", 10, "Sentence movement supports tension rather than floating at one tempo.", "prose-precision"),
			node("transition-pressure", "Transition Pressure", "scene architecture", "Move between scenes without losing momentum or causal force.", "core", 11, "Transitions carry forward pressure instead of draining it.", "scene-architecture"),
			node("conflict-triangulation", "Conflict Triangulation", "scene architecture", "Layer at least two competing pressures into a scene so conflict deepens instead of repeating.", "scene", 12, "Scenes carry intersecting pressures rather than a single flat disagreement.", "tension-escalation", "motivation-pressure"),
			node("entrance-exit-weight", "Entrance and Exit Weight", "scene architecture", "Make arrivals and departures alter the emotional and power balance.", "scene", 13, "Beginnings and endings of scenes create a felt shift in force.", "scene-architecture"),
			node("spatial-legibility", "Spatial Legibility", "scene architecture", "Keep bodies, objects, and thresholds clear enough that ritual and conflict remain trackable.", "scene", 14, "The reader never loses where power is staged in the room.", "scene-architecture"),
			node("status-turns", "Status Turns", "dialogue intelligence", "Track shifts in rank or leverage as scenes progress.", "scene", 15, "Power changes become visible through action and speech rather than explanation.", "tension-escalation"),
			node("withheld-information-control", "Withheld Information Control", "narrative clarity", "Delay information deliberately without confusing the reader.", "scene", 16, "Mystery creates pressure without muddying comprehension.", "point-of-view-discipline"),
			node("reversal-construction", "Reversal Construction", "scene architecture", "Build reversals that arise from choice and pressure instead of arbitrary surprise.", "scene", 17, "Turns feel both surprising and inevitable in retrospect.", "conflict-triangulation"),
			node("dialogue-under-strain", "Dialogue Under Strain", "dialogue intelligence", "Use speech to reveal rank, fracture, and motive under pressure.", "scene", 18, "Dialogue carries conflict and hierarchy without flattening into exposition.", "scene-architecture", "prose-precision"),
			node("silence-deployment", "Silence Deployment", "dialogue intelligence", "Use silence, omission, and interruption as active dramatic tools.", "scene", 19, "Gaps in speech do visible work in the conflict.", "dialogue-under-strain"),
			node("ritualized-action", "Ritualized Action", "mythic tone", "Let repeated actions build sacred or historical weight.", "scene", 20, "Ritual action gathers pressure instead of reading like decorative repetition.", "spatial-legibility", "gesture-as-subtext"),
			node("desire-contradiction", "Desire Contradiction", "tragic inevitability", "Give characters conflicting wants that force self-wounding choices.", "character", 21, "The strongest choices satisfy one value by betraying another.", "motivation-pressure"),
			node("flaw-leverage", "Flaw Leverage", "tragic inevitability", "Turn weakness into a force that changes events rather than a personality label.", "character", 22, "The flaw actively alters what the character does under stress.", "desire-contradiction"),
			node("mask-and-reveal", "Mask and Reveal", "dialogue intelligence", "Let self-presentation crack under pressure in visible increments.", "character", 23, "The reader can feel the gap between persona and need tightening.", "dialogue-under-strain"),
			node("loyalty-pressure", "Loyalty Pressure", "tragic inevitability", "Use bonds and duties to intensify difficult choices.", "character", 24, "Relationships complicate decisions rather than merely softening them.", "desire-contradiction"),
			node("inner-logic-coherence", "Inner Logic Coherence", "narrative clarity", "Make even destructive choices understandable from the character's value system.", "character", 25, "The reader may disagree with the choice but understands why it happened.", "motivation-pressure"),
			node("wound-recurrence", "Wound Recurrence", "emotional compression", "Let recurring injury patterns shape reaction without overexplaining backstory.", "character", 26, "Past wounds manifest through pattern rather than lecture.", "emotional-compression"),
			node("sacrificial-costing", "Sacrificial Costing", "tragic inevitability", "Make sacrifice expensive, specific, and irreversible.", "character", 27, "The loss feels concrete and cannot be hand-waved away later.", "loyalty-pressure", "consequence-memory"),
			node("descriptive-hierarchy", "Descriptive Hierarchy", "image freshness", "Choose details according to dramatic importance rather than equal decorative weight.", "world", 28, "Description highlights the details that carry pressure in the scene.", "prose-precision"),
			node("worldbuilding-economy", "Worldbuilding Economy", "worldbuilding economy", "Imply history through pressure, not exposition.", "world", 29, "Setting, history, and politics arrive through conflict-bearing details.", "scene-architecture"),
			node("institutional-pressure", "Institutional Pressure", "worldbuilding economy", "Let laws, rites, and structures push on the characters.", "world", 30, "The world acts on the scene instead of sitting inert behind it.", "worldbuilding-economy"),
			node("historical-shadow", "Historical Shadow", "worldbuilding economy", "Make prior events haunt the present action without summarizing the timeline.", "world", 31, "History is felt as pressure, taboo, or memory in the present scene.", "worldbuilding-economy"),
			node("cosmology-restraint", "Cosmology Restraint", "mythic tone", "Treat metaphysical material with gravity and limitation instead of lore-sprawl.", "world", 32, "The sacred or cosmic dimension deepens the narrative without flooding it.", "worldbuilding-economy", "mythic-register"),
			node("setting-as-fate", "Setting as Fate", "tragic inevitability", "Use place to narrow choices and reinforce doom.", "world", 33, "Place actively limits what characters can do or become.", "institutional-pressure"),
			node("object-recurrence", "Object Recurrence", "symbolic control", "Repeat charged objects with increasing pressure and altered context.", "symbol", 34, "Objects gain force through recurrence rather than explanation.", "prose-precision", "emotional-compression"),
			node("symbolic-discipline", "Symbolic Discipline", "symbolic control", "Let objects carry fate without direct explanation.", "symbol", 35, "Symbols recur with pressure and coherence without being glossed for the reader.", "prose-precision", "emotional-compression"),
			node("metaphor-restraint", "Metaphor Restraint", "prose precision", "Keep figurative language selective enough that symbolic pressure stays readable.", "symbol", 36, "Metaphor sharpens the scene instead of fogging it.", "prose-precision"),
			node("mythic-register", "Mythic Register", "mythic tone", "Sustain gravity without inflated diction or pastiche.", "symbol", 37, "The prose feels elevated and disciplined rather than imitative or overwritten.", "symbolic-discipline", "prose-precision"),
			node("prophetic-ambiguity", "Prophetic Ambiguity", "symbolic control", "Handle prophecy or omen so it intensifies choice instead of dictating plot mechanically.", "symbol", 38, "Omen deepens dread without making character agency irrelevant.", "symbolic-discipline", "withheld-information-control"),
			node("image-freshness", "Image Freshness", "image freshness", "Prefer singular earned imagery over fantasy stock language.", "symbol", 39, "Images feel specific to the scene rather than drawn from generic fantasy vocabulary.", "mythic-register"),
			node("motif-weaving", "Motif Weaving", "symbolic control", "Thread recurring images across scenes so they accumulate meaning.", "symbol", 40, "Motifs echo across the draft without becoming heavy-handed.", "object-recurrence", "image-freshness"),
			node("fate-vs-choice-balance", "Fate and Choice Balance", "tragic inevitability", "Let destiny intensify decisions instead of replacing them.", "tragedy", 41, "The characters remain morally and causally responsible inside a fated frame.", "prophetic-ambiguity", "motivation-pressure"),
			node("tragic-inevitability", "Tragic Inevitability", "tragic inevitability", "Make choices close futures rather than open them.", "tragedy", 42, "The ending feels sealed by prior choices, not imposed by authorial summary.", "causal-clarity", "scene-architecture"),
			node("doom-acceleration", "Doom Acceleration", "tragic inevitability", "Make later choices shrink options faster than earlier ones.", "tragedy", 43, "The field of possible rescue narrows visibly over time.", "tragic-inevitability"),
			node("mercy-denial", "Mercy Denial", "tragic inevitability", "Withhold easy relief unless it costs something else essential.", "tragedy", 44, "Moments that might save the character carry real counter-costs.", "tragic-inevitability"),
			node("consequence-irrevocability", "Consequence Irrevocability", "tragic inevitability", "Prevent late-stage reversals from undoing earned damage.", "tragedy", 45, "The final costs remain binding.", "sacrificial-costing", "doom-acceleration"),
			node("catastrophic-reversal", "Catastrophic Reversal", "tragic inevitability", "Build the final turn so it arises from the story's own pressure system.", "tragedy", 46, "The collapse feels prepared, not sprung.", "reversal-construction", "consequence-irrevocability"),
			node("lament-with-restraint", "Lament with Restraint", "emotional compression", "Allow grief to become audible without turning the prose slack or melodramatic.", "tragedy", 47, "Grief lands hard without overstatement.", "emotional-compression", "catastrophic-reversal"),
			node("chapter-pressure", "Chapter Pressure", "scene architecture", "End larger units on consequential instability rather than generic cliffhangers.", "structure", 48, "Chapter breaks intensify onward movement.", "transition-pressure"),
			node("parallel-scene-design", "Parallel Scene Design", "scene architecture", "Echo and invert scene structures so the arc gains formal force.", "structure", 49, "Repeated scene shapes create visible development rather than repetition.", "chapter-pressure"),
			node("subplot-convergence", "Subplot Convergence", "narrative clarity", "Bring secondary threads into the main fatal pressure instead of leaving them ornamental.", "structure", 50, "Subplots resolve by feeding the central conflict.", "consequence-memory", "chapter-pressure"),
			node("ensemble-balance", "Ensemble Balance", "scene architecture", "Distribute attention across multiple important figures without dissolving focus.", "structure", 51, "Secondary figures matter without stealing the story's center.", "status-turns", "point-of-view-discipline"),
			node("revision-triage", "Revision Triage", "prose precision", "Prioritize the revisions that improve force rather than endlessly polishing surfaces.", "revision", 52, "Later drafts improve structural force before cosmetic flourish.", "line-level-rhythm", "chapter-pressure"),
			node("density-with-breath", "Density with Breath", "mythic tone", "Control compression so the prose feels rich rather than clogged.", "revision", 53, "Dense passages remain readable because release valves are placed intentionally.", "line-level-rhythm", "mythic-register"),
			node("anti-pastiche-control", "Anti-Pastiche Control", "mythic tone", "Keep influence visible only as depth of craft, not imitation of voice or phrase.", "revision", 54, "The prose feels owned rather than borrowed.", "mythic-register", "image-freshness"),
			node("ending-resonance", "Ending Resonance", "symbolic control", "Close on an image or action that carries the whole tragic field without explaining it.", "revision", 55, "The ending resonates outward through prior motifs and consequences.", "motif-weaving", "lament-with-restraint"),
		},
	}

	youth := TGOTreeDefinition{
		Slug:           "youth-writing-foundations",
		Title:          "Youth Writing Foundations",
		Description:    "Foundational writing track for younger writers building sentence, paragraph, and story control.",
		SeedCodes:      []string{"word-choice", "sentence-variety", "sentence-clarity"},
		PrioritySkills: []string{"word choice", "sentence variety", "clarity and coherence", "paragraph control", "narrative sequencing", "descriptive precision", "dialogue basics"},
		TGOs: []TGO{
			node("word-choice", "Word Choice", "word choice", "Choose specific words instead of vague filler.", "foundation", 1, "The draft prefers concrete, accurate words over generic ones like nice, bad, thing, and stuff."),
			node("sentence-variety", "Sentence Variety", "sentence variety", "Mix short, medium, and longer sentences so the prose does not drag or stutter.", "foundation", 2, "Sentence openings and lengths vary enough to keep the writing moving naturally."),
			node("sentence-clarity", "Sentence Clarity", "clarity and coherence", "Make each sentence easy to follow the first time.", "foundation", 3, "Most sentences read cleanly without confusion about who did what."),
			node("capitalization-control", "Capitalization Control", "spelling and mechanics", "Use capitals correctly at sentence starts and for names.", "foundation", 4, "Capital letters appear where the reader expects them.", "sentence-clarity"),
			node("punctuation-basics", "Punctuation Basics", "spelling and mechanics", "Use periods, question marks, and commas in the right places.", "foundation", 5, "Basic punctuation supports meaning instead of getting in the way.", "sentence-clarity"),
			node("subject-verb-agreement", "Subject-Verb Agreement", "grammar control", "Match singular and plural subjects with the right verb form.", "foundation", 6, "Subject and verb stay in agreement even in longer sentences.", "sentence-clarity"),
			node("complete-sentences", "Complete Sentences", "clarity and coherence", "Avoid fragments unless used deliberately for effect.", "foundation", 7, "Most sentences feel finished and complete.", "sentence-clarity"),
			node("simple-transitions", "Simple Transitions", "narrative sequencing", "Use words like then, next, after, and because to guide the reader.", "foundation", 8, "The reader can move from sentence to sentence without getting lost.", "sentence-clarity"),
			node("detail-selection", "Detail Selection", "descriptive precision", "Choose a few helpful details instead of trying to say everything.", "foundation", 9, "Details help the reader picture the idea quickly.", "word-choice"),
			node("paragraph-control", "Paragraph Control", "paragraph control", "Group related ideas into readable paragraphs.", "foundation", 10, "Paragraphs stay focused on one small unit of action or thought.", "sentence-clarity"),
			node("topic-sentence-basics", "Topic Sentence Basics", "paragraph control", "Start information paragraphs with a clear main idea.", "paragraph", 11, "The first sentence signals what the paragraph will explain.", "paragraph-control"),
			node("supporting-detail-balance", "Supporting Detail Balance", "paragraph control", "Add enough support without wandering away from the point.", "paragraph", 12, "Paragraphs feel developed but not messy.", "paragraph-control"),
			node("paragraph-order", "Paragraph Order", "clarity and coherence", "Arrange paragraphs in a sensible sequence.", "paragraph", 13, "The reader sees why one paragraph follows the previous one.", "paragraph-control", "simple-transitions"),
			node("sentence-complexity", "Sentence Complexity", "sentence complexity", "Control longer sentences without losing clarity.", "paragraph", 14, "Longer sentences still stay grammatically stable and easy to follow.", "sentence-variety", "sentence-clarity"),
			node("pronoun-clarity", "Pronoun Clarity", "clarity and coherence", "Make sure words like he, she, they, and it clearly point to the right noun.", "paragraph", 15, "The reader rarely has to guess what a pronoun refers to.", "sentence-clarity"),
			node("list-control", "List Control", "scannability", "Use lists or grouped items so related ideas are easier to read.", "paragraph", 16, "Grouped information is easier to scan and remember.", "paragraph-order"),
			node("opening-focus", "Opening Focus", "narrative sequencing", "Start a piece with a clear situation or topic.", "paragraph", 17, "The reader knows what the piece is about right away.", "topic-sentence-basics"),
			node("closing-sense", "Closing Sense", "narrative sequencing", "End with a sentence that feels finished instead of suddenly stopping.", "paragraph", 18, "The ending gives the reader a clear sense of closure.", "paragraph-order"),
			node("narrative-sequencing", "Narrative Sequencing", "narrative sequencing", "Put events in a clear order with useful transitions.", "story", 19, "The reader can retell the story events in the right order without guessing.", "sentence-clarity", "paragraph-control"),
			node("beginning-middle-end", "Beginning, Middle, End", "narrative sequencing", "Give stories a stable shape instead of a pile of events.", "story", 20, "Stories feel complete rather than abruptly cut off.", "narrative-sequencing"),
			node("character-goal-basics", "Character Goal Basics", "story development", "Give the main character a want the reader can understand.", "story", 21, "The reader knows what the character is trying to do.", "narrative-sequencing"),
			node("problem-and-solution", "Problem and Solution", "story development", "Build stories around a problem that needs an answer.", "story", 22, "The plot moves because something needs to be solved.", "character-goal-basics"),
			node("cause-and-effect", "Cause and Effect", "narrative sequencing", "Show why one event leads to the next.", "story", 23, "Events feel connected rather than random.", "narrative-sequencing"),
			node("scene-setting-basics", "Scene Setting Basics", "descriptive precision", "Help the reader know where and when the scene is happening.", "story", 24, "The scene is anchored with a few clear cues.", "detail-selection"),
			node("character-description", "Character Description", "descriptive precision", "Choose details that help the reader remember who matters.", "story", 25, "Characters are identifiable without long catalogs of features.", "detail-selection"),
			node("descriptive-specificity", "Descriptive Specificity", "descriptive precision", "Use concrete details that help the reader picture the scene.", "story", 26, "Description relies on a few strong details instead of broad labels.", "word-choice"),
			node("showing-action", "Showing Action", "story development", "Write actions clearly enough that the reader can picture them happening.", "story", 27, "Action scenes unfold in a readable order.", "cause-and-effect"),
			node("emotion-through-action", "Emotion Through Action", "story development", "Show feelings through what characters do, say, or notice.", "story", 28, "Emotion appears in action instead of only labels like sad or mad.", "showing-action"),
			node("dialogue-basics", "Dialogue Basics", "dialogue basics", "Use dialogue to show speaker intent and keep the reader oriented.", "dialogue", 29, "Dialogue is easy to track and does more than fill space.", "sentence-clarity"),
			node("speaker-tracking", "Speaker Tracking", "dialogue basics", "Make it clear who is talking.", "dialogue", 30, "Readers can follow the conversation without confusion.", "dialogue-basics"),
			node("dialogue-punctuation", "Dialogue Punctuation", "spelling and mechanics", "Use quotation marks and end punctuation correctly in dialogue.", "dialogue", 31, "Dialogue mechanics stop distracting the reader.", "dialogue-basics", "punctuation-basics"),
			node("dialogue-purpose", "Dialogue Purpose", "dialogue basics", "Let dialogue move the story or reveal character instead of repeating what the reader already knows.", "dialogue", 32, "Conversations change something.", "dialogue-basics"),
			node("voice-difference", "Voice Difference", "dialogue basics", "Help different characters sound slightly different from each other.", "dialogue", 33, "Speakers feel distinct without needing labels every line.", "dialogue-purpose"),
			node("question-use", "Question Use", "dialogue basics", "Use questions to create movement in dialogue.", "dialogue", 34, "Questions open the next beat instead of stalling the scene.", "dialogue-purpose"),
			node("informational-writing-basics", "Informational Writing Basics", "clarity and coherence", "Explain a topic clearly with grouped facts and examples.", "forms", 35, "Information writing stays organized and readable.", "topic-sentence-basics", "supporting-detail-balance"),
			node("opinion-writing-basics", "Opinion Writing Basics", "claim clarity", "State an opinion and support it with simple reasons.", "forms", 36, "The reader knows the opinion and sees at least a few reasons for it.", "topic-sentence-basics"),
			node("letter-structure", "Letter Structure", "professional format", "Use a greeting, body, and closing in a simple letter or email.", "forms", 37, "The format fits the writing situation.", "paragraph-control"),
			node("instructions-clarity", "Instructions Clarity", "actionability", "Write steps so another person could actually follow them.", "forms", 38, "A reader can do the task without guessing missing steps.", "paragraph-order", "list-control"),
			node("summary-basics", "Summary Basics", "clarity and coherence", "Retell the main points without including every small detail.", "forms", 39, "The summary captures what matters most.", "opening-focus", "closing-sense"),
			node("revision-rereading", "Revision Rereading", "revision habits", "Reread the whole piece once for sense before fixing details.", "revision", 40, "The writer notices missing words, jumps, and confusions through rereading.", "sentence-clarity"),
			node("revision-cutting", "Revision Cutting", "revision habits", "Remove extra words or repeated ideas.", "revision", 41, "The piece gets shorter or sharper where it needs to.", "revision-rereading"),
			node("revision-adding", "Revision Adding", "revision habits", "Add the detail, explanation, or transition the reader needs.", "revision", 42, "The second draft answers obvious reader questions better than the first.", "revision-rereading"),
			node("proofreading-basics", "Proofreading Basics", "spelling and mechanics", "Do a final pass for capitalization, punctuation, and spelling.", "revision", 43, "Simple mechanical errors decrease across drafts.", "capitalization-control", "punctuation-basics"),
			node("spelling-patterns", "Spelling Patterns", "spelling and mechanics", "Use common spelling patterns more reliably.", "revision", 44, "Frequently used words are spelled correctly most of the time.", "proofreading-basics"),
			node("tense-consistency", "Tense Consistency", "grammar control", "Keep verb tense stable unless the timeline really changes.", "revision", 45, "The reader does not get bounced between past and present by accident.", "subject-verb-agreement", "narrative-sequencing"),
			node("comma-control", "Comma Control", "spelling and mechanics", "Use commas where they help separate ideas and items.", "revision", 46, "Commas begin helping the reader instead of feeling random.", "punctuation-basics", "sentence-complexity"),
			node("audience-awareness", "Audience Awareness", "audience alignment", "Think about what the reader needs to know first.", "revision", 47, "The draft explains enough for the reader it is meant for.", "opening-focus"),
			node("tone-awareness", "Tone Awareness", "tone calibration", "Match the tone to the assignment or audience.", "revision", 48, "The writing sounds appropriate to the purpose.", "audience-awareness"),
			node("focus-under-length", "Focus Under Length", "clarity and coherence", "Stay on topic even when the piece gets longer.", "revision", 49, "Longer assignments do not drift as much.", "supporting-detail-balance", "paragraph-order"),
			node("evidence-basics", "Evidence Basics", "evidence integration", "Use a fact, example, or quote to support a point.", "forms", 50, "Support actually helps prove the point being made.", "opinion-writing-basics", "informational-writing-basics"),
			node("creative-risk-taking", "Creative Risk Taking", "story development", "Try a less obvious detail, image, or event choice when the piece needs more life.", "revision", 51, "The writer begins making braver, more specific choices.", "descriptive-specificity", "voice-difference"),
			node("independent-self-check", "Independent Self-Check", "revision habits", "Use a repeatable checklist before turning work in.", "revision", 52, "The writer can catch more issues without direct prompting.", "revision-rereading", "proofreading-basics"),
		},
	}

	story := TGOTreeDefinition{
		Slug:           "story-craft-track",
		Title:          "Story Craft Track",
		Description:    "Flexible fiction track for writers building scene, character pressure, and narrative clarity.",
		SeedCodes:      []string{"story-causal-clarity", "story-scene-architecture", "story-prose-precision"},
		PrioritySkills: []string{"narrative clarity", "scene architecture", "prose precision", "emotional compression", "dialogue intelligence", "image freshness", "worldbuilding economy"},
		TGOs: []TGO{
			node("story-causal-clarity", "Causal Clarity", "narrative clarity", "Make action and consequence legible beat by beat.", "core", 1, "Readers can follow each decision and its immediate consequence without backtracking."),
			node("story-scene-architecture", "Scene Architecture", "scene architecture", "Stage turns, entrances, exits, and power shifts cleanly.", "core", 2, "Scenes remain spatially and dramatically legible even under pressure."),
			node("story-prose-precision", "Prose Precision", "prose precision", "Replace soft modifiers with exact nouns and verbs.", "core", 3, "Line-level revision sharpens verbs, nouns, and cadence without ornament."),
			node("story-point-of-view", "Point of View Control", "narrative clarity", "Hold a stable perspective so readers know whose experience they are in.", "core", 4, "The narrative lens stays coherent throughout scenes.", "story-causal-clarity"),
			node("story-tension-growth", "Tension Growth", "scene architecture", "Increase dramatic pressure instead of circling the same beat.", "core", 5, "Scenes ratchet upward rather than repeating the same pressure.", "story-scene-architecture"),
			node("story-transition-flow", "Transition Flow", "scene architecture", "Move between scenes cleanly without flattening momentum.", "core", 6, "Transitions preserve story pressure.", "story-scene-architecture"),
			node("story-subtext-basics", "Subtext Basics", "dialogue intelligence", "Let some meaning stay beneath the surface of speech and action.", "core", 7, "The scene suggests more than characters directly say.", "story-prose-precision"),
			node("story-emotional-compression", "Emotional Compression", "emotional compression", "Condense feeling into image, gesture, and consequence.", "core", 8, "Feeling arrives through action and image rather than named emotion.", "story-causal-clarity"),
			node("story-description-focus", "Description Focus", "image freshness", "Use detail selectively so description supports story movement.", "core", 9, "Description sharpens scenes instead of delaying them.", "story-prose-precision"),
			node("story-conflict-design", "Conflict Design", "scene architecture", "Build scenes around competing wants rather than static conversation.", "scene", 10, "Most scenes have clear friction.", "story-tension-growth"),
			node("story-reversal", "Reversal Construction", "scene architecture", "Turn scenes in a way that feels surprising but earned.", "scene", 11, "Scene endings change the dramatic position in a meaningful way.", "story-conflict-design"),
			node("story-spatial-legibility", "Spatial Legibility", "scene architecture", "Keep movement, position, and important objects trackable.", "scene", 12, "Readers stay oriented inside the scene.", "story-scene-architecture"),
			node("story-status-dynamics", "Status Dynamics", "dialogue intelligence", "Track who holds power and how it shifts in a scene.", "scene", 13, "Power changes become visible through speech and behavior.", "story-conflict-design"),
			node("story-silence", "Silence and Omission", "dialogue intelligence", "Use what is unsaid as a source of pressure.", "scene", 14, "Silence does active dramatic work.", "story-subtext-basics"),
			node("story-entrance-exit", "Entrance and Exit Design", "scene architecture", "Make arrivals and departures change the scene's energy.", "scene", 15, "Scene boundaries matter dramatically.", "story-reversal"),
			node("story-object-handling", "Object Handling", "worldbuilding economy", "Use props and objects in ways that alter action or meaning.", "scene", 16, "Objects matter because they affect what happens.", "story-spatial-legibility"),
			node("story-character-goal", "Character Goal", "story development", "Give the protagonist a clear near-term want inside the scene.", "character", 17, "The reader knows what the character is trying to get right now.", "story-causal-clarity"),
			node("story-character-need", "Character Need", "story development", "Shape deeper need beneath the visible goal.", "character", 18, "The story shows pressure beneath the obvious want.", "story-character-goal"),
			node("story-contradiction", "Character Contradiction", "story development", "Let characters want things that do not sit comfortably together.", "character", 19, "Character choices reveal tension inside the self.", "story-character-need"),
			node("story-voice-difference", "Voice Difference", "dialogue intelligence", "Help major characters sound and perceive differently.", "character", 20, "Characters feel distinct in speech and outlook.", "story-subtext-basics"),
			node("story-backstory-restraint", "Backstory Restraint", "narrative clarity", "Bring in backstory only where it changes the present pressure.", "character", 21, "Backstory arrives at useful points instead of stalling scenes.", "story-point-of-view"),
			node("story-relationship-pressure", "Relationship Pressure", "story development", "Use relationships to complicate goals and choices.", "character", 22, "Bonds and loyalties make scenes harder, not easier.", "story-character-need"),
			node("story-wound-pattern", "Wound Pattern", "emotional compression", "Show recurring emotional injury through reaction patterns.", "character", 23, "The draft implies inner damage through repeated behavior.", "story-emotional-compression"),
			node("story-change-arc", "Change Arc", "story development", "Track how the protagonist's responses evolve over time.", "character", 24, "Character movement is visible across the story.", "story-contradiction", "story-relationship-pressure"),
			node("story-setting-pressure", "Setting Pressure", "worldbuilding economy", "Use setting as a constraint on behavior and possibility.", "world", 25, "Place shapes what can happen.", "story-description-focus"),
			node("story-worldbuilding-economy", "Worldbuilding Economy", "worldbuilding economy", "Imply background through conflict-bearing details.", "world", 26, "Context arrives through action rather than exposition.", "story-setting-pressure"),
			node("story-social-structure", "Social Structure", "worldbuilding economy", "Make rank, rules, or institutions legible in the story world.", "world", 27, "Social order is visible without lecture.", "story-worldbuilding-economy"),
			node("story-history-pressure", "History Pressure", "worldbuilding economy", "Let prior events haunt the present situation.", "world", 28, "The world has a felt past that affects the present scene.", "story-worldbuilding-economy"),
			node("story-sensory-balance", "Sensory Balance", "descriptive precision", "Use more than one sensory channel without overloading the prose.", "world", 29, "Sensory detail enriches the scene without clutter.", "story-description-focus"),
			node("story-atmosphere", "Atmosphere Control", "image freshness", "Shape a scene's mood through concrete choices in detail and rhythm.", "world", 30, "Atmosphere feels intentional and sustained.", "story-sensory-balance"),
			node("story-image-freshness", "Image Freshness", "image freshness", "Prefer singular imagery over stock language.", "style", 31, "Images feel specific to the story at hand.", "story-description-focus"),
			node("story-metaphor-restraint", "Metaphor Restraint", "prose precision", "Use figurative language selectively so it sharpens instead of fogging.", "style", 32, "Figurative language earns its place.", "story-prose-precision"),
			node("story-rhythm-control", "Rhythm Control", "prose precision", "Vary sentence movement to support dramatic need.", "style", 33, "Cadence shifts intentionally with the scene.", "story-prose-precision"),
			node("story-paragraph-momentum", "Paragraph Momentum", "prose precision", "Shape paragraphs to carry momentum instead of flattening it.", "style", 34, "Paragraphing supports the reading pace.", "story-rhythm-control"),
			node("story-exposition-control", "Exposition Control", "narrative clarity", "Deliver necessary explanation in measured doses.", "style", 35, "The story informs without stalling.", "story-backstory-restraint"),
			node("story-dialogue-under-strain", "Dialogue Under Strain", "dialogue intelligence", "Use dialogue to carry conflict and motive under pressure.", "style", 36, "Dialogue scenes keep moving dramatically.", "story-subtext-basics", "story-status-dynamics"),
			node("story-motif", "Motif Weaving", "symbolic control", "Repeat small images or objects so they accumulate meaning.", "style", 37, "Motifs echo productively across the draft.", "story-image-freshness"),
			node("story-theme-pressure", "Theme Through Pressure", "story development", "Let theme emerge from conflict and consequence instead of explanation.", "structure", 38, "The story says something through what happens.", "story-change-arc", "story-worldbuilding-economy"),
			node("story-plot-escalation", "Plot Escalation", "story development", "Make later events cost more than earlier ones.", "structure", 39, "The story gets harder in meaningful ways.", "story-change-arc", "story-reversal"),
			node("story-subplot-use", "Subplot Use", "story development", "Make side threads feed the main pressure instead of distracting from it.", "structure", 40, "Secondary threads matter to the whole story.", "story-plot-escalation"),
			node("story-middle-management", "Middle Management", "structure and pacing", "Keep the middle of the story developing instead of sagging.", "structure", 41, "The middle continues to turn and deepen the story.", "story-plot-escalation"),
			node("story-climax-construction", "Climax Construction", "structure and pacing", "Build the final confrontation from prior conflicts and choices.", "structure", 42, "The climax feels earned and inevitable in context.", "story-middle-management", "story-theme-pressure"),
			node("story-ending-payoff", "Ending Payoff", "structure and pacing", "Close on a result that answers the story's main pressure.", "structure", 43, "The ending feels consequential, not simply finished.", "story-climax-construction"),
			node("story-revision-triage", "Revision Triage", "revision habits", "Prioritize structural fixes before polishing surfaces.", "revision", 44, "Later drafts improve the right problems first.", "story-climax-construction"),
			node("story-scene-cutting", "Scene Cutting", "revision habits", "Remove scenes or beats that do not move story, character, or pressure.", "revision", 45, "The story gets leaner without losing force.", "story-revision-triage"),
			node("story-clarity-pass", "Clarity Pass", "revision habits", "Do a revision pass focused only on confusion and causal gaps.", "revision", 46, "Reader disorientation decreases from draft to draft.", "story-revision-triage"),
			node("story-language-pass", "Language Pass", "revision habits", "Do a separate pass focused on diction, rhythm, and repeated phrasing.", "revision", 47, "Line quality improves without disturbing major structure.", "story-rhythm-control", "story-revision-triage"),
			node("story-beta-awareness", "Reader Awareness", "audience alignment", "Anticipate where readers will need orientation, payoff, or release.", "revision", 48, "Drafts answer likely reader questions more effectively.", "story-clarity-pass"),
			node("story-opening-grip", "Opening Grip", "structure and pacing", "Open with enough pressure or curiosity to earn continuation.", "revision", 49, "The opening pulls the reader into the story quickly.", "story-character-goal", "story-setting-pressure"),
			node("story-scene-sequencing", "Scene Sequencing", "structure and pacing", "Arrange scenes in the order that best compounds pressure.", "revision", 50, "Scene order feels intentional rather than merely chronological.", "story-transition-flow", "story-middle-management"),
			node("story-emotional-arc", "Emotional Arc", "emotional compression", "Track the emotional movement of the whole story, not just individual scenes.", "revision", 51, "The story has an emotional shape that accumulates.", "story-emotional-compression", "story-change-arc"),
			node("story-final-resonance", "Final Resonance", "image freshness", "Land the ending with an image, action, or line that continues to echo.", "revision", 52, "The story closes with durable aftereffect.", "story-ending-payoff", "story-motif"),
		},
	}

	thought := TGOTreeDefinition{
		Slug:           "thought-leadership-track",
		Title:          "Thought Leadership Track",
		Description:    "Nonfiction track for sharper claims, stronger insight, and more authoritative structure.",
		SeedCodes:      []string{"claim-clarity", "audience-alignment", "sentence-economy"},
		PrioritySkills: []string{"claim clarity", "audience alignment", "sentence economy", "structural signposting", "insight density", "evidence integration", "authority and voice", "clarity and coherence"},
		TGOs: []TGO{
			node("claim-clarity", "Claim Clarity", "claim clarity", "State the central argument early and plainly.", "core", 1, "The controlling claim is visible without excavation."),
			node("audience-alignment", "Audience Alignment", "audience alignment", "Aim the piece at a clearly understood reader and need.", "core", 2, "The piece feels written for a specific reader, not the void."),
			node("sentence-economy", "Sentence Economy", "sentence economy", "Cut throat-clearing and reduce bloated syntax.", "core", 3, "Most sentences justify their length and remain easy to track."),
			node("problem-framing", "Problem Framing", "claim clarity", "Define the problem or tension the piece is addressing.", "core", 4, "Readers understand why the argument matters.", "claim-clarity"),
			node("stakes-articulation", "Stakes Articulation", "claim clarity", "Make clear what changes if the reader accepts or ignores the argument.", "core", 5, "The consequences of the claim are visible.", "problem-framing"),
			node("reader-pain-mapping", "Reader Pain Mapping", "audience alignment", "Name the friction or uncertainty your audience actually feels.", "core", 6, "The piece meets a real reader need instead of an imagined generic one.", "audience-alignment"),
			node("terminology-discipline", "Terminology Discipline", "clarity and coherence", "Use key terms consistently and define them when needed.", "core", 7, "Important terms keep the same meaning across the piece.", "claim-clarity"),
			node("paragraph-focus", "Paragraph Focus", "clarity and coherence", "Make each paragraph serve one argumentative function.", "core", 8, "Paragraphs do one job well instead of several badly.", "sentence-economy"),
			node("structural-signposting", "Structural Signposting", "structural signposting", "Guide the reader through the argument with clear sectional moves.", "structure", 9, "Major turns in the argument are easy to follow.", "claim-clarity"),
			node("opening-hook", "Opening Hook", "authority and voice", "Open with tension, surprise, or consequence rather than generic introduction.", "structure", 10, "The opening earns attention and frames the piece's purpose.", "audience-alignment"),
			node("thesis-placement", "Thesis Placement", "claim clarity", "Place the main claim where the reader can use it.", "structure", 11, "The argument's center is discoverable early enough to guide reading.", "claim-clarity"),
			node("section-ordering", "Section Ordering", "structural signposting", "Arrange sections so the reasoning compounds instead of repeating.", "structure", 12, "The order of sections feels necessary rather than arbitrary.", "structural-signposting"),
			node("bridge-sentences", "Bridge Sentences", "structural signposting", "Use transitions that connect the logic between sections.", "structure", 13, "The reader sees why the next section follows the last.", "section-ordering"),
			node("scaffolding-load", "Scaffolding Load", "clarity and coherence", "Provide enough setup for the reader without overexplaining the obvious.", "structure", 14, "Supportive framing helps rather than slows the piece.", "reader-pain-mapping"),
			node("ending-that-lands", "Ending That Lands", "authority and voice", "Close with consequence, implication, or next move rather than vague summary.", "structure", 15, "The conclusion sharpens the piece's aftereffect.", "stakes-articulation"),
			node("insight-density", "Insight Density", "insight density", "Favor original value over generic summary.", "insight", 16, "Paragraphs deliver actual insight, not only competent explanation.", "claim-clarity", "audience-alignment"),
			node("novel-distinction", "Novel Distinction", "insight density", "State what your view adds, changes, or sees differently.", "insight", 17, "The piece distinguishes itself from common takes.", "insight-density"),
			node("conceptual-compression", "Conceptual Compression", "insight density", "Express complicated reasoning with as little waste as possible.", "insight", 18, "The argument becomes easier to remember and share.", "sentence-economy", "insight-density"),
			node("counterargument-handling", "Counterargument Handling", "claim clarity", "Represent the strongest alternative view fairly before responding.", "insight", 19, "The piece grows more credible because it meets resistance honestly.", "claim-clarity", "section-ordering"),
			node("analogy-selection", "Analogy Selection", "evidence integration", "Use analogies that actually clarify the reasoning.", "insight", 20, "Analogies simplify without distorting the argument.", "insight-density"),
			node("framework-construction", "Framework Construction", "insight density", "Offer a model, lens, or structure the reader can reuse.", "insight", 21, "The piece gives the reader a durable way to think.", "novel-distinction", "conceptual-compression"),
			node("implication-expansion", "Implication Expansion", "insight density", "Follow the argument outward into meaningful consequences.", "insight", 22, "The reader sees why the insight changes decisions or interpretation.", "stakes-articulation", "framework-construction"),
			node("evidence-integration", "Evidence Integration", "evidence integration", "Use examples and evidence to sharpen the claim instead of decorating it.", "evidence", 23, "Evidence advances the reasoning instead of merely appearing beside it.", "structural-signposting"),
			node("example-selection", "Example Selection", "evidence integration", "Choose examples that illuminate the exact claim being made.", "evidence", 24, "Examples feel precise, not merely available.", "evidence-integration"),
			node("quote-restraint", "Quote Restraint", "evidence integration", "Quote only what the reader needs and do the interpretive work yourself.", "evidence", 25, "Quoted material supports the author's reasoning instead of replacing it.", "evidence-integration"),
			node("data-interpretation", "Data Interpretation", "evidence integration", "Explain what data means rather than assuming the number speaks for itself.", "evidence", 26, "Evidence arrives with interpretation and relevance.", "evidence-integration"),
			node("case-study-use", "Case Study Use", "evidence integration", "Use specific cases to deepen understanding without losing the general claim.", "evidence", 27, "Case studies serve the argument rather than becoming detours.", "example-selection"),
			node("evidence-proportion", "Evidence Proportion", "evidence integration", "Balance support so the piece is neither under-argued nor overburdened with proof.", "evidence", 28, "Support feels sufficient and proportionate to the claim.", "data-interpretation", "quote-restraint"),
			node("source-framing", "Source Framing", "authority and voice", "Introduce outside sources in a way that clarifies why they matter.", "evidence", 29, "References feel selected and interpreted, not pasted in.", "evidence-integration"),
			node("authority-and-voice", "Authority and Voice", "authority and voice", "Sound decisive without becoming inflated or vague.", "voice", 30, "The voice feels credible, direct, and earned.", "sentence-economy"),
			node("tone-calibration-thought", "Tone Calibration", "tone calibration", "Match tone to audience and platform without flattening your point of view.", "voice", 31, "The tone earns trust instead of distance or overfamiliarity.", "audience-alignment"),
			node("sentence-pressure", "Sentence Pressure", "sentence economy", "Keep individual sentences moving toward a point instead of idling.", "voice", 32, "Sentences feel driven by thought instead of filler.", "sentence-economy"),
			node("cadence-variety", "Cadence Variety", "authority and voice", "Vary sentence rhythm so the prose sounds deliberate rather than monotonous.", "voice", 33, "The prose has audible shape and emphasis.", "sentence-pressure"),
			node("verb-forward-style", "Verb-Forward Style", "sentence economy", "Favor active, decision-carrying verbs over inert abstractions.", "voice", 34, "The prose sounds more direct and forceful.", "sentence-economy"),
			node("opening-that-frames", "Opening That Frames", "authority and voice", "Frame the piece with the right tension, question, or contradiction.", "voice", 35, "The beginning establishes the intellectual field quickly.", "opening-hook", "problem-framing"),
			node("ending-that-opens", "Ending That Opens", "authority and voice", "End in a way that points the reader toward consequence or application.", "voice", 36, "The ending feels like an earned next step.", "ending-that-lands", "implication-expansion"),
			node("platform-fit", "Platform Fit", "audience alignment", "Shape density, format, and framing to the publishing context.", "audience", 37, "The piece feels natural to its channel rather than transplanted.", "tone-calibration-thought"),
			node("reader-objection-forecasting", "Reader Objection Forecasting", "audience alignment", "Anticipate likely resistance and answer it before the reader disengages.", "audience", 38, "The piece stays ahead of the reader's skepticism.", "counterargument-handling", "reader-pain-mapping"),
			node("trust-building-specificity", "Trust-Building Specificity", "authority and voice", "Use specifics that build credibility without showboating expertise.", "audience", 39, "Specificity increases trust rather than clutter.", "authority-and-voice"),
			node("actionable-insight", "Actionable Insight", "actionability", "Translate abstract reasoning into useful next moves.", "audience", 40, "Readers can apply the argument after reading.", "implication-expansion"),
			node("audience-segmentation", "Audience Segmentation", "audience alignment", "Recognize when one audience needs a different framing from another.", "audience", 41, "The writer can adjust argument framing by reader group.", "platform-fit", "reader-pain-mapping"),
			node("series-thinking", "Series Thinking", "structural signposting", "Write one piece as part of a larger body of thought, not an isolated post.", "audience", 42, "Pieces begin to reinforce each other across publication.", "framework-construction", "platform-fit"),
			node("argument-arc", "Argument Arc", "structural signposting", "Shape the full piece so the argument deepens in deliberate stages.", "advanced", 43, "The reader feels guided through a progression rather than a stack of points.", "section-ordering", "insight-density"),
			node("strategic-repetition", "Strategic Repetition", "clarity and coherence", "Repeat key ideas with variation so they stick without becoming redundant.", "advanced", 44, "Key concepts recur memorably instead of monotonously.", "argument-arc"),
			node("thought-contrast", "Thought Contrast", "insight density", "Use juxtaposition to sharpen distinctions and reveal stakes.", "advanced", 45, "Comparisons deepen rather than merely decorate the argument.", "novel-distinction"),
			node("concept-bridging", "Concept Bridging", "insight density", "Connect ideas from different domains without making the bridge feel forced.", "advanced", 46, "Cross-domain connections feel earned and illuminating.", "framework-construction"),
			node("revision-triage-thought", "Revision Triage", "revision habits", "Fix claim, structure, and evidence before polishing voice.", "revision", 47, "Revisions improve the argument in the right order.", "argument-arc"),
			node("clarity-pass-thought", "Clarity Pass", "revision habits", "Revise once only for reader confusion, ambiguity, and hidden assumptions.", "revision", 48, "The piece becomes easier to follow without losing depth.", "revision-triage-thought"),
			node("evidence-pass-thought", "Evidence Pass", "revision habits", "Revise once only for support quality and proportionality.", "revision", 49, "Support becomes better matched to claims.", "revision-triage-thought"),
			node("voice-pass-thought", "Voice Pass", "revision habits", "Do a separate revision pass for tone, cadence, and confidence.", "revision", 50, "The prose sounds more owned and deliberate.", "revision-triage-thought"),
			node("headline-framing", "Headline Framing", "claim clarity", "Title the piece so it frames the real promise or provocation.", "revision", 51, "The title attracts the right reader with the right promise.", "opening-that-frames"),
			node("portable-takeaway", "Portable Takeaway", "insight density", "Leave the reader with a phrasing or model they can carry away and repeat.", "revision", 52, "A memorable formulation survives after the piece ends.", "ending-that-opens", "framework-construction"),
		},
	}

	professional := TGOTreeDefinition{
		Slug:           "professional-writing-track",
		Title:          "Professional Writing Track",
		Description:    "Practical writing track for clarity, structure, tone, and actionability in workplace communication.",
		SeedCodes:      []string{"objective-clarity", "professional-audience-alignment", "professional-sentence-economy"},
		PrioritySkills: []string{"clarity and coherence", "audience alignment", "sentence economy", "structural signposting", "tone calibration", "actionability", "scannability", "evidence integration"},
		TGOs: []TGO{
			node("objective-clarity", "Objective Clarity", "clarity and coherence", "Make the document's purpose unmistakable.", "core", 1, "Readers know what the document is trying to do within the first lines."),
			node("professional-audience-alignment", "Audience Alignment", "audience alignment", "Match detail, tone, and framing to the intended reader.", "core", 2, "The level of context and framing matches the audience's needs."),
			node("professional-sentence-economy", "Sentence Economy", "sentence economy", "Reduce clutter and keep sentences easy to scan.", "core", 3, "Sentences are concise without becoming abrupt or vague."),
			node("context-first", "Context First", "clarity and coherence", "Provide the background the reader needs before diving into detail.", "core", 4, "Readers understand the situation before being asked to decide or act.", "objective-clarity"),
			node("ask-visibility", "Ask Visibility", "actionability", "State the request, decision, or needed action plainly.", "core", 5, "The reader does not have to infer what is being asked.", "objective-clarity"),
			node("deadline-clarity", "Deadline Clarity", "actionability", "Make dates, timing, and urgency explicit.", "core", 6, "Time-sensitive writing leaves little room for ambiguity.", "ask-visibility"),
			node("ownership-clarity", "Ownership Clarity", "actionability", "Clarify who is responsible for what.", "core", 7, "The reader can tell who owns the next step.", "ask-visibility"),
			node("terminology-consistency", "Terminology Consistency", "clarity and coherence", "Use project or domain terms consistently.", "core", 8, "Terms keep the same meaning across the document.", "objective-clarity"),
			node("reader-load-control", "Reader Load Control", "scannability", "Reduce unnecessary cognitive load in high-context documents.", "core", 9, "The reader can process the message quickly.", "professional-sentence-economy"),
			node("professional-structural-signposting", "Structural Signposting", "structural signposting", "Use headings, transitions, and ordering that reduce reader effort.", "structure", 10, "The document's structure is obvious and useful.", "objective-clarity"),
			node("front-loaded-summary", "Front-Loaded Summary", "scannability", "Open longer documents with the essential point and recommendation.", "structure", 11, "Busy readers can grasp the message from the top.", "context-first", "professional-structural-signposting"),
			node("bullet-discipline", "Bullet Discipline", "scannability", "Use bullets for true lists and keep them parallel and clear.", "structure", 12, "Bullets increase clarity instead of becoming visual clutter.", "professional-structural-signposting"),
			node("section-order", "Section Order", "structural signposting", "Arrange sections in the order the reader needs them.", "structure", 13, "Readers encounter information at the moment it is useful.", "professional-structural-signposting"),
			node("bridge-lines", "Bridge Lines", "structural signposting", "Use short connecting lines that explain why the next section matters.", "structure", 14, "Transitions reduce re-reading and confusion.", "section-order"),
			node("summary-vs-detail-balance", "Summary and Detail Balance", "scannability", "Balance overview and detail so neither overwhelms the other.", "structure", 15, "Documents feel complete without feeling overfull.", "front-loaded-summary", "reader-load-control"),
			node("document-openings", "Document Openings", "tone calibration", "Open with the right mix of context, purpose, and tone.", "structure", 16, "The opening sets the right professional frame immediately.", "context-first", "professional-audience-alignment"),
			node("document-closings", "Document Closings", "actionability", "Close with a clear next step or explicit status.", "structure", 17, "The document ends with directional clarity.", "ask-visibility", "ownership-clarity"),
			node("professional-tone-calibration", "Tone Calibration", "tone calibration", "Sound professional without becoming stiff, passive, or evasive.", "tone", 18, "Tone supports trust and clarity.", "professional-audience-alignment"),
			node("directness-with-respect", "Directness with Respect", "tone calibration", "Be clear and firm without becoming sharp or dismissive.", "tone", 19, "Messages stay humane while still being decisive.", "professional-tone-calibration"),
			node("confidence-without-overclaiming", "Confidence Without Overclaiming", "tone calibration", "State recommendations confidently while respecting uncertainty.", "tone", 20, "The prose sounds trustworthy because certainty is calibrated.", "professional-tone-calibration"),
			node("bad-news-delivery", "Bad News Delivery", "tone calibration", "Handle negative updates honestly without confusion or unnecessary harshness.", "tone", 21, "The message is clear, respectful, and hard to misread.", "directness-with-respect"),
			node("stakeholder-sensitivity", "Stakeholder Sensitivity", "audience alignment", "Adapt wording to different stakeholder concerns without losing substance.", "tone", 22, "The same core message can be tuned for different readers.", "professional-audience-alignment", "professional-tone-calibration"),
			node("decision-language", "Decision Language", "actionability", "Use wording that makes decisions and recommendations unmistakable.", "tone", 23, "Readers can tell what is proposed, decided, or still open.", "confidence-without-overclaiming", "ask-visibility"),
			node("active-voice-control", "Active Voice Control", "sentence economy", "Prefer active constructions when responsibility and action matter.", "tone", 24, "The writing more often names who did what.", "professional-sentence-economy"),
			node("meeting-note-clarity", "Meeting Note Clarity", "structural signposting", "Capture what happened, what was decided, and what remains open.", "forms", 25, "Meeting notes work as a dependable reference.", "front-loaded-summary", "ownership-clarity"),
			node("status-update-control", "Status Update Control", "clarity and coherence", "Report progress, blockers, and next steps succinctly.", "forms", 26, "Status updates are easy to scan and act on.", "objective-clarity", "document-closings"),
			node("proposal-structure", "Proposal Structure", "structural signposting", "Organize proposals so recommendation, rationale, and cost are easy to evaluate.", "forms", 27, "Proposals guide the reader toward a decision.", "front-loaded-summary", "decision-language"),
			node("email-thread-discipline", "Email Thread Discipline", "scannability", "Keep long email threads readable and avoid burying the ask.", "forms", 28, "Even mid-thread, readers can spot what matters.", "ask-visibility", "bullet-discipline"),
			node("spec-clarity", "Spec Clarity", "clarity and coherence", "Write requirements so they reduce interpretation drift.", "forms", 29, "Requirements are concrete enough to align execution.", "terminology-consistency", "objective-clarity"),
			node("supporting-rationale", "Supporting Rationale", "evidence integration", "Pair recommendations with the constraints or facts that justify them.", "analysis", 30, "The reader sees why the recommendation exists.", "ask-visibility"),
			node("professional-evidence-integration", "Evidence Integration", "evidence integration", "Use relevant data, examples, or constraints to support decisions.", "analysis", 31, "Support strengthens the recommendation instead of cluttering it.", "supporting-rationale"),
			node("risk-articulation", "Risk Articulation", "clarity and coherence", "Name tradeoffs, risks, and uncertainty directly.", "analysis", 32, "The writing makes risk legible rather than implicit.", "confidence-without-overclaiming"),
			node("assumption-marking", "Assumption Marking", "clarity and coherence", "Make assumptions visible so readers can challenge or confirm them.", "analysis", 33, "Hidden assumptions become inspectable.", "risk-articulation"),
			node("decision-framing", "Decision Framing", "actionability", "Present options and a recommendation in a way that supports choice.", "analysis", 34, "Decision-makers can compare options quickly and fairly.", "supporting-rationale", "risk-articulation"),
			node("tradeoff-language", "Tradeoff Language", "clarity and coherence", "Explain what is gained and lost with each option.", "analysis", 35, "Tradeoffs read as concrete rather than fuzzy.", "decision-framing"),
			node("quantification-basics", "Quantification Basics", "evidence integration", "Use numbers when they improve precision and decision quality.", "analysis", 36, "Quantitative detail sharpens claims without overwhelming the document.", "professional-evidence-integration"),
			node("professional-scannability", "Scannability", "scannability", "Shape prose and sections so busy readers can extract the key message quickly.", "scanning", 37, "A fast pass still captures the essential points.", "professional-sentence-economy", "professional-structural-signposting"),
			node("heading-discipline", "Heading Discipline", "scannability", "Use headings that communicate meaning, not just labels.", "scanning", 38, "A skimming reader can follow the document from headings alone.", "professional-scannability"),
			node("visual-grouping", "Visual Grouping", "scannability", "Group related content so the page supports the argument.", "scanning", 39, "Layout choices help readers interpret structure.", "professional-scannability"),
			node("table-use", "Table Use", "scannability", "Use tables when comparison or compact lookup matters.", "scanning", 40, "Tables clarify instead of complicating the message.", "visual-grouping", "tradeoff-language"),
			node("annotation-restraint", "Annotation Restraint", "scannability", "Use emphasis, bolding, and notes selectively so signal remains visible.", "scanning", 41, "Visual emphasis points to what matters most.", "visual-grouping"),
			node("executive-briefing-style", "Executive Briefing Style", "tone calibration", "Adapt writing for senior readers who need outcome, risk, and ask first.", "advanced", 42, "Senior-facing writing becomes shorter, clearer, and more decision-ready.", "front-loaded-summary", "decision-language"),
			node("cross-functional-translation", "Cross-Functional Translation", "audience alignment", "Translate specialized context for readers outside your function.", "advanced", 43, "Specialized content becomes usable across teams.", "stakeholder-sensitivity", "terminology-consistency"),
			node("persuasion-with-restraint", "Persuasion with Restraint", "tone calibration", "Persuade through clarity and support rather than hype.", "advanced", 44, "Recommendations feel convincing because they are well reasoned.", "decision-framing", "professional-evidence-integration"),
			node("conflict-de-escalation", "Conflict De-Escalation", "tone calibration", "Write disagreement or correction in a way that preserves working relationships.", "advanced", 45, "Difficult messages stay clear without becoming inflammatory.", "directness-with-respect", "bad-news-delivery"),
			node("alignment-language", "Alignment Language", "clarity and coherence", "State what is agreed, what is pending, and what changed.", "advanced", 46, "The document reduces coordination ambiguity.", "meeting-note-clarity", "status-update-control"),
			node("revision-triage-professional", "Revision Triage", "revision habits", "Fix purpose, ask, and structure before polishing wording.", "revision", 47, "Revisions improve the highest-value issues first.", "decision-framing", "document-openings"),
			node("clarity-pass-professional", "Clarity Pass", "revision habits", "Revise once only for ambiguity, missing context, and reader effort.", "revision", 48, "Documents become easier to process quickly.", "revision-triage-professional"),
			node("tone-pass-professional", "Tone Pass", "revision habits", "Revise once only for tone, directness, and professionalism.", "revision", 49, "Tone becomes more reliable across different message types.", "revision-triage-professional"),
			node("format-pass-professional", "Format Pass", "revision habits", "Revise once only for headings, bullets, layout, and scan behavior.", "revision", 50, "Formatting better supports the document's purpose.", "revision-triage-professional", "professional-scannability"),
			node("proof-of-action", "Proof of Action", "actionability", "Check that every important document leaves the reader knowing what to do next.", "revision", 51, "Action items and decisions survive the final pass.", "format-pass-professional", "document-closings"),
			node("reusable-template-thinking", "Reusable Template Thinking", "structural signposting", "Recognize recurring document patterns and standardize the best ones.", "revision", 52, "The writer begins building dependable internal writing systems.", "proposal-structure", "status-update-control"),
		},
	}

	academic := TGOTreeDefinition{
		Slug:           "academic-essay-track",
		Title:          "Academic Essay Track",
		Description:    "Structured analytical and research writing track for essays, papers, and close reading.",
		SeedCodes:      []string{"academic-thesis-clarity", "academic-structure-basics", "academic-evidence-basics"},
		PrioritySkills: []string{"thesis clarity", "evidence integration", "structural signposting", "analysis depth", "source handling", "clarity and coherence", "revision habits"},
		TGOs: []TGO{
			node("academic-thesis-clarity", "Thesis Clarity", "thesis clarity", "State a debatable claim early and clearly.", "core", 1, "The reader can point to the paper's central claim without guesswork."),
			node("academic-structure-basics", "Essay Structure Basics", "structural signposting", "Shape the essay with a readable introduction, body, and conclusion.", "core", 2, "The overall form supports the argument instead of obscuring it."),
			node("academic-evidence-basics", "Evidence Basics", "evidence integration", "Use quotations, details, or examples to support claims.", "core", 3, "Claims are usually supported by concrete textual or factual evidence."),
			node("academic-prompt-reading", "Prompt Reading", "assignment alignment", "Answer the actual assignment rather than a nearby topic.", "core", 4, "The essay stays aligned with the prompt's verbs and scope.", "academic-thesis-clarity"),
			node("academic-scope-control", "Scope Control", "thesis clarity", "Keep the claim narrow enough to be argued well.", "core", 5, "The paper argues one manageable claim instead of several diffuse ones.", "academic-thesis-clarity"),
			node("academic-paragraph-focus", "Paragraph Focus", "clarity and coherence", "Make each paragraph serve one argumentative function.", "core", 6, "Paragraphs do one job instead of collecting loosely related thoughts.", "academic-structure-basics"),
			node("academic-topic-sentences", "Topic Sentences", "structural signposting", "Open paragraphs with lines that orient the reader to the paragraph's role.", "core", 7, "The reader can follow the progression from topic sentences alone.", "academic-paragraph-focus"),
			node("academic-quote-integration", "Quote Integration", "source handling", "Blend quotations into your own syntax and reasoning.", "core", 8, "Quoted material reads as part of the argument, not pasted beside it.", "academic-evidence-basics"),
			node("academic-analysis-basics", "Analysis Basics", "analysis depth", "Explain how evidence supports the claim instead of leaving the connection implied.", "core", 9, "The essay interprets evidence instead of merely presenting it.", "academic-evidence-basics"),
			node("academic-summary-restraint", "Summary Restraint", "analysis depth", "Reduce plot or source summary when analysis should lead.", "core", 10, "Summary stays subordinate to argument.", "academic-analysis-basics"),
			node("academic-introduction-function", "Introduction Function", "structural signposting", "Use the introduction to frame stakes, context, and claim.", "structure", 11, "The introduction prepares the argument instead of circling it.", "academic-structure-basics"),
			node("academic-conclusion-function", "Conclusion Function", "structural signposting", "Use the conclusion to sharpen consequence rather than repeat points mechanically.", "structure", 12, "The ending extends the argument's significance.", "academic-structure-basics"),
			node("academic-section-order", "Section Order", "structural signposting", "Arrange paragraphs so the reasoning compounds in a useful order.", "structure", 13, "The sequence of ideas feels deliberate and cumulative.", "academic-topic-sentences"),
			node("academic-transitions", "Transitions", "structural signposting", "Use transitions to show the relationship between ideas and sections.", "structure", 14, "The reader can see why one point follows another.", "academic-section-order"),
			node("academic-claim-ladder", "Claim Ladder", "analysis depth", "Build from smaller claims toward the larger thesis.", "structure", 15, "Each section clearly supports the broader argument.", "academic-scope-control", "academic-analysis-basics"),
			node("academic-counterreading", "Counterreading", "analysis depth", "Acknowledge plausible alternative interpretations before refining your own.", "structure", 16, "The essay appears more rigorous because it meets resistance directly.", "academic-claim-ladder"),
			node("academic-close-reading", "Close Reading", "analysis depth", "Attend carefully to language, structure, or detail in a specific passage.", "analysis", 17, "The essay extracts meaning from precise textual choices.", "academic-analysis-basics"),
			node("academic-pattern-recognition", "Pattern Recognition", "analysis depth", "Move from one example to a repeatable pattern or structural insight.", "analysis", 18, "Interpretation scales beyond isolated moments.", "academic-close-reading"),
			node("academic-concept-definition", "Concept Definition", "clarity and coherence", "Define key terms with enough precision for the paper's argument.", "analysis", 19, "Important words have stable, usable meanings in the essay.", "academic-thesis-clarity"),
			node("academic-stakes-articulation", "Stakes Articulation", "analysis depth", "Explain why the argument matters beyond being technically correct.", "analysis", 20, "The essay clarifies what changes when the claim is accepted.", "academic-introduction-function"),
			node("academic-context-use", "Context Use", "source handling", "Provide historical, theoretical, or textual context where it improves understanding.", "analysis", 21, "Context sharpens interpretation rather than drowning it.", "academic-concept-definition"),
			node("academic-precision-diction", "Precision Diction", "clarity and coherence", "Prefer exact academic language over vague inflation.", "analysis", 22, "The prose sounds precise without becoming padded.", "academic-summary-restraint"),
			node("academic-sentence-control", "Sentence Control", "clarity and coherence", "Write sentences that stay clear even when handling complex ideas.", "analysis", 23, "Complex thought remains readable at the sentence level.", "academic-precision-diction"),
			node("academic-quotation-restraint", "Quotation Restraint", "source handling", "Quote only what the reader needs and analyze the rest in your own words.", "analysis", 24, "The essay relies more on your reasoning than on long quotations.", "academic-quote-integration"),
			node("academic-source-selection", "Source Selection", "source handling", "Choose sources that genuinely support the paper's question and level.", "research", 25, "Sources fit the argument instead of appearing because they were easy to find.", "academic-evidence-basics"),
			node("academic-source-evaluation", "Source Evaluation", "source handling", "Judge the relevance, credibility, and use of a source before relying on it.", "research", 26, "Sources are chosen more critically and strategically.", "academic-source-selection"),
			node("academic-synthesis", "Source Synthesis", "source handling", "Put sources in conversation instead of summarizing them one at a time.", "research", 27, "The paper thinks across sources rather than stacking them.", "academic-source-evaluation"),
			node("academic-citation-discipline", "Citation Discipline", "source handling", "Cite consistently and clearly according to the assignment's expectations.", "research", 28, "Citations stop distracting from the argument.", "academic-source-selection"),
			node("academic-paraphrase-control", "Paraphrase Control", "source handling", "Paraphrase accurately without drifting into patchwriting.", "research", 29, "Paraphrases are both faithful and genuinely your own prose.", "academic-source-evaluation"),
			node("academic-literature-positioning", "Literature Positioning", "analysis depth", "Position your claim within an existing scholarly conversation.", "research", 30, "The essay makes clear how it relates to prior interpretations or arguments.", "academic-synthesis"),
			node("academic-method-signaling", "Method Signaling", "structural signposting", "Tell the reader what kind of reading or reasoning the essay is doing.", "research", 31, "The paper's method is legible early enough to guide interpretation.", "academic-introduction-function"),
			node("academic-evidence-proportion", "Evidence Proportion", "evidence integration", "Balance the amount of support so analysis still leads.", "research", 32, "Evidence feels proportionate to the claim's needs.", "academic-close-reading", "academic-quotation-restraint"),
			node("academic-footnote-judgment", "Footnote Judgment", "source handling", "Move secondary detail out of the main line of argument when appropriate.", "research", 33, "The body stays readable because supporting material is placed wisely.", "academic-context-use"),
			node("academic-style-consistency", "Style Consistency", "clarity and coherence", "Keep register, tense, and terminology consistent across the essay.", "style", 34, "The paper feels controlled rather than patched together.", "academic-sentence-control"),
			node("academic-hedging-control", "Hedging Control", "clarity and coherence", "Use qualification where needed without weakening the claim into mush.", "style", 35, "The essay sounds careful and confident at once.", "academic-thesis-clarity"),
			node("academic-active-voice", "Active Voice Judgment", "clarity and coherence", "Choose active or passive constructions deliberately rather than by habit.", "style", 36, "Voice choices support clarity and emphasis.", "academic-sentence-control"),
			node("academic-abstract-noun-control", "Abstract Noun Control", "clarity and coherence", "Reduce abstract overload by anchoring ideas in concrete language or examples.", "style", 37, "The prose handles concepts without becoming weightless.", "academic-precision-diction"),
			node("academic-rhythm", "Academic Rhythm", "clarity and coherence", "Vary sentence rhythm so the prose stays readable under intellectual load.", "style", 38, "The essay avoids sounding mechanically flat.", "academic-style-consistency"),
			node("academic-significance-lines", "Significance Lines", "analysis depth", "Write sentences that explicitly state why the evidence matters.", "style", 39, "Readers do not have to infer the argumentative payoff alone.", "academic-stakes-articulation"),
			node("academic-pivot-control", "Pivot Control", "structural signposting", "Make analytical turns visible when the paper changes scale, lens, or emphasis.", "style", 40, "Important shifts never feel abrupt or hidden.", "academic-transitions"),
			node("academic-draft-outline", "Draft Outline", "revision habits", "Outline the draft after writing to inspect actual structure.", "revision", 41, "The writer can see what the paper is truly doing, not just what it meant to do.", "academic-section-order"),
			node("academic-revision-triage", "Revision Triage", "revision habits", "Fix thesis, structure, and evidence before polishing sentences.", "revision", 42, "Revision energy goes to the highest-value problems first.", "academic-draft-outline"),
			node("academic-clarity-pass", "Clarity Pass", "revision habits", "Revise once only for reader confusion, missing context, and sentence-level fog.", "revision", 43, "The paper becomes easier to follow without losing complexity.", "academic-revision-triage"),
			node("academic-analysis-pass", "Analysis Pass", "revision habits", "Revise once only for interpretive depth and argumentative payoff.", "revision", 44, "Body paragraphs say more than they did in the previous draft.", "academic-revision-triage"),
			node("academic-source-pass", "Source Pass", "revision habits", "Revise once only for quotation, citation, and source balance.", "revision", 45, "Source handling becomes cleaner and more controlled.", "academic-revision-triage"),
			node("academic-line-editing", "Line Editing", "revision habits", "Do a late pass for diction, repetition, and awkward syntax.", "revision", 46, "Surface prose improves without derailing structural work.", "academic-style-consistency", "academic-revision-triage"),
			node("academic-proofreading", "Proofreading", "revision habits", "Check formatting, citations, punctuation, and small mechanical errors at the end.", "revision", 47, "Mechanical noise decreases in final drafts.", "academic-citation-discipline"),
			node("academic-title-framing", "Title Framing", "thesis clarity", "Give the essay a title that signals the real question or claim.", "revision", 48, "The title frames the argument instead of merely naming the topic.", "academic-introduction-function"),
			node("academic-abstract-writing", "Abstract Writing", "structural signposting", "Compress the essay's argument into a short, accurate summary when needed.", "revision", 49, "The writer can state the paper's logic compactly and clearly.", "academic-thesis-clarity", "academic-section-order"),
			node("academic-oral-defense", "Oral Defense Readiness", "analysis depth", "Write the paper so its reasoning can be defended aloud under questioning.", "revision", 50, "The essay's logic holds when challenged directly.", "academic-counterreading", "academic-stakes-articulation"),
			node("academic-final-resonance", "Final Resonance", "analysis depth", "End with a conclusion that extends the essay's significance beyond summary.", "revision", 51, "The paper leaves the reader with a sharpened consequence.", "academic-conclusion-function", "academic-stakes-articulation"),
			node("academic-independent-voice", "Independent Voice", "analysis depth", "Let the paper sound like a thinking writer rather than a collage of source language.", "revision", 52, "The essay feels authored and intellectually owned.", "academic-synthesis", "academic-line-editing"),
		},
	}

	technical := TGOTreeDefinition{
		Slug:           "technical-writing-track",
		Title:          "Technical Writing Track",
		Description:    "Documentation and technical communication track for guides, specs, references, and system explanations.",
		SeedCodes:      []string{"technical-user-goal", "technical-structure-basics", "technical-step-clarity"},
		PrioritySkills: []string{"user goal alignment", "structural signposting", "actionability", "scannability", "accuracy", "example quality", "technical precision"},
		TGOs: []TGO{
			node("technical-user-goal", "User Goal Alignment", "user goal alignment", "Write from the reader's task, not the writer's internal system map.", "core", 1, "The document clearly serves a user goal."),
			node("technical-structure-basics", "Documentation Structure Basics", "structural signposting", "Organize documents into predictable sections that support fast lookup.", "core", 2, "The document shape reduces reader hunting."),
			node("technical-step-clarity", "Step Clarity", "actionability", "Write steps that can be followed without guessing missing actions.", "core", 3, "A reader can execute the task directly from the instructions."),
			node("technical-prereq-clarity", "Prerequisite Clarity", "actionability", "State requirements and assumptions before the task begins.", "core", 4, "Readers know what they need before starting.", "technical-user-goal"),
			node("technical-term-control", "Terminology Control", "technical precision", "Use consistent terms for the same concept throughout the document.", "core", 5, "Readers are not forced to translate between synonyms.", "technical-user-goal"),
			node("technical-scope-control", "Scope Control", "user goal alignment", "Keep each document scoped to a clear task or reference need.", "core", 6, "The document does one job well instead of several badly.", "technical-user-goal"),
			node("technical-sequence-control", "Sequence Control", "actionability", "Order steps in the sequence the user must actually perform them.", "core", 7, "Instruction order matches execution order.", "technical-step-clarity"),
			node("technical-result-visibility", "Result Visibility", "actionability", "Tell the reader what should happen after a step succeeds.", "core", 8, "Users can tell whether they are still on track.", "technical-step-clarity"),
			node("technical-error-prevention", "Error Prevention", "accuracy", "Warn the reader before common mistakes or dangerous actions.", "core", 9, "The document helps users avoid predictable failure.", "technical-prereq-clarity"),
			node("technical-reader-context", "Reader Context", "user goal alignment", "Provide just enough system context for the task to make sense.", "core", 10, "Readers understand why the task exists without drowning in architecture.", "technical-scope-control"),
			node("technical-opening-summary", "Opening Summary", "scannability", "Open longer docs with a compact summary of purpose and outcome.", "structure", 11, "Readers can decide quickly whether the doc is relevant.", "technical-structure-basics"),
			node("technical-section-order", "Section Order", "structural signposting", "Arrange sections in the order a user needs them.", "structure", 12, "The document's order matches reader workflow.", "technical-structure-basics"),
			node("technical-heading-discipline", "Heading Discipline", "scannability", "Use headings that communicate meaning, not vague labels.", "structure", 13, "A skimming reader can navigate by headings alone.", "technical-section-order"),
			node("technical-paragraph-economy", "Paragraph Economy", "scannability", "Keep paragraphs short enough for high-information documents.", "structure", 14, "Dense material remains readable because chunking is controlled.", "technical-structure-basics"),
			node("technical-list-design", "List Design", "scannability", "Use ordered and unordered lists deliberately to improve scanning.", "structure", 15, "Lists clarify instead of cluttering.", "technical-heading-discipline"),
			node("technical-cross-links", "Cross-Link Judgment", "structural signposting", "Link to adjacent docs when detail belongs elsewhere.", "structure", 16, "The document participates in a usable documentation system.", "technical-scope-control"),
			node("technical-glossary-handling", "Glossary Handling", "technical precision", "Define or link unfamiliar terms at the right moment.", "structure", 17, "Terminology becomes usable without overexplaining basics.", "technical-term-control"),
			node("technical-reference-layout", "Reference Layout", "scannability", "Shape reference docs for lookup rather than linear reading.", "structure", 18, "Reference material becomes faster to retrieve.", "technical-heading-discipline"),
			node("technical-task-vs-reference", "Task vs Reference Distinction", "user goal alignment", "Separate how-to content from conceptual or reference material.", "structure", 19, "Different document types stop fighting each other.", "technical-scope-control", "technical-reference-layout"),
			node("technical-example-basics", "Example Basics", "example quality", "Use examples that match the user's real task.", "examples", 20, "Examples clarify the document instead of decorating it.", "technical-user-goal"),
			node("technical-example-minimality", "Minimal Reproducible Examples", "example quality", "Strip examples down to the smallest useful case.", "examples", 21, "Examples are small enough to understand and reuse quickly.", "technical-example-basics"),
			node("technical-example-realism", "Example Realism", "example quality", "Keep examples close enough to real use that transfer is easy.", "examples", 22, "Examples feel plausible and portable.", "technical-example-basics"),
			node("technical-input-output", "Input and Output Clarity", "example quality", "Show what goes in and what comes out when examples matter.", "examples", 23, "The reader can compare expected and actual behavior.", "technical-result-visibility"),
			node("technical-code-comment-restraint", "Code Comment Restraint", "technical precision", "Comment examples only where the comment adds meaning.", "examples", 24, "Comments support comprehension instead of duplicating the obvious.", "technical-example-minimality"),
			node("technical-edge-case-signaling", "Edge Case Signaling", "accuracy", "Call out unusual conditions without overwhelming the main path.", "examples", 25, "Edge cases are visible but do not drown the primary task.", "technical-error-prevention"),
			node("technical-api-surface-clarity", "API Surface Clarity", "technical precision", "Describe parameters, returns, and behavior with precision.", "reference", 26, "Readers can use the interface without inference gaps.", "technical-reference-layout"),
			node("technical-parameter-docs", "Parameter Documentation", "technical precision", "Explain inputs in a way that reduces misuse.", "reference", 27, "Parameter docs clarify expectations and constraints.", "technical-api-surface-clarity"),
			node("technical-state-models", "State Model Explanation", "technical precision", "Explain stateful behavior or lifecycle changes clearly.", "reference", 28, "Readers can reason about changing system state.", "technical-reader-context"),
			node("technical-constraint-language", "Constraint Language", "accuracy", "Write limits, defaults, and caveats in unambiguous terms.", "reference", 29, "Important constraints are hard to miss or misread.", "technical-api-surface-clarity"),
			node("technical-versioning-signals", "Versioning Signals", "accuracy", "Mark version-specific behavior clearly.", "reference", 30, "Readers can tell what applies to their environment.", "technical-constraint-language"),
			node("technical-deprecation-notes", "Deprecation Notes", "accuracy", "Handle outdated behavior with clear alternatives and timing.", "reference", 31, "Deprecated material is still usable but not misleading.", "technical-versioning-signals"),
			node("technical-troubleshooting-basics", "Troubleshooting Basics", "accuracy", "Write error and troubleshooting sections around actual failure modes.", "support", 32, "Readers can recover from common failure with the doc alone.", "technical-error-prevention"),
			node("technical-symptom-cause", "Symptom to Cause Mapping", "accuracy", "Map user-visible failure to likely causes in a readable way.", "support", 33, "Troubleshooting becomes diagnostic, not random.", "technical-troubleshooting-basics"),
			node("technical-remediation-order", "Remediation Order", "actionability", "Present recovery steps from safest and simplest to most disruptive.", "support", 34, "The support flow lowers risk while increasing effectiveness.", "technical-troubleshooting-basics"),
			node("technical-logging-guidance", "Logging Guidance", "technical precision", "Tell the reader what information to gather when diagnosing problems.", "support", 35, "Support instructions produce more useful bug reports.", "technical-symptom-cause"),
			node("technical-compatibility-notes", "Compatibility Notes", "accuracy", "Document platform, environment, or dependency differences clearly.", "support", 36, "Users can detect whether the doc applies to their setup.", "technical-versioning-signals"),
			node("technical-screenshot-judgment", "Screenshot Judgment", "scannability", "Use screenshots only when they add real orientation value.", "support", 37, "Images help users instead of aging into clutter.", "technical-heading-discipline"),
			node("technical-tone-control", "Technical Tone Control", "clarity and coherence", "Sound direct and helpful without overexplaining or condescending.", "style", 38, "The document feels professional and reader-centered.", "technical-user-goal"),
			node("technical-sentence-economy", "Sentence Economy", "technical precision", "Keep technical prose concise without dropping needed precision.", "style", 39, "The document stays lean and exact.", "technical-term-control"),
			node("technical-active-voice", "Active Voice Judgment", "technical precision", "Choose active or passive constructions deliberately based on task clarity.", "style", 40, "Voice choice supports comprehension.", "technical-sentence-economy"),
			node("technical-ambiguity-hunting", "Ambiguity Hunting", "accuracy", "Remove phrases that could plausibly be interpreted in more than one way.", "style", 41, "The final prose leaves fewer interpretation traps.", "technical-constraint-language"),
			node("technical-consistency-pass", "Consistency Pass", "revision habits", "Revise once only for term, heading, and formatting consistency.", "revision", 42, "The doc feels internally coherent.", "technical-term-control", "technical-heading-discipline"),
			node("technical-accuracy-pass", "Accuracy Pass", "revision habits", "Revise once only to verify facts, commands, and outputs.", "revision", 43, "The doc becomes more trustworthy because details are checked.", "technical-versioning-signals"),
			node("technical-user-walkthrough", "User Walkthrough", "revision habits", "Run the doc as if you were a new user following it literally.", "revision", 44, "Gaps in actionability become visible through walkthrough.", "technical-step-clarity"),
			node("technical-triage", "Revision Triage", "revision habits", "Fix goal alignment, structure, and accuracy before polishing phrasing.", "revision", 45, "The highest-impact doc problems are solved first.", "technical-user-walkthrough"),
			node("technical-shortening-pass", "Shortening Pass", "revision habits", "Cut repetition and unnecessary explanation late in the process.", "revision", 46, "The doc becomes faster to scan without losing essential meaning.", "technical-triage"),
			node("technical-example-pass", "Example Pass", "revision habits", "Revise examples separately for clarity, realism, and correctness.", "revision", 47, "Examples become more useful than the surrounding prose alone.", "technical-example-realism", "technical-triage"),
			node("technical-support-pass", "Support Pass", "revision habits", "Revise support and troubleshooting sections against real user failure cases.", "revision", 48, "Recovery guidance improves with each draft.", "technical-troubleshooting-basics", "technical-triage"),
			node("technical-navigation-pass", "Navigation Pass", "revision habits", "Revise headings, links, and chunking so the doc works under skim conditions.", "revision", 49, "Readers find information faster.", "technical-cross-links", "technical-consistency-pass"),
			node("technical-release-note-discipline", "Release Note Discipline", "accuracy", "Summarize changes in a way users can act on quickly.", "revision", 50, "Change communication becomes clearer and more useful.", "technical-versioning-signals", "technical-opening-summary"),
			node("technical-ownership-signals", "Ownership Signals", "user goal alignment", "Make it clear where readers should go for deeper help or changes.", "revision", 51, "Readers know where this document sits in the larger support system.", "technical-cross-links", "technical-troubleshooting-basics"),
			node("technical-final-polish", "Final Polish", "revision habits", "Do a last pass for friction points that slow reading without adding value.", "revision", 52, "The final document feels clean, deliberate, and trustworthy.", "technical-shortening-pass", "technical-ambiguity-hunting"),
		},
	}

	persuasive := TGOTreeDefinition{
		Slug:           "persuasive-writing-track",
		Title:          "Persuasive Writing Track",
		Description:    "Argument and persuasion track for opinion pieces, advocacy, proposals, and rhetorical writing.",
		SeedCodes:      []string{"persuasive-claim", "persuasive-audience", "persuasive-reasoning"},
		PrioritySkills: []string{"claim clarity", "audience alignment", "reasoning quality", "evidence integration", "rhetorical force", "objection handling", "actionability"},
		TGOs: []TGO{
			node("persuasive-claim", "Claim Clarity", "claim clarity", "State what you want the reader to believe or do.", "core", 1, "The reader can state the piece's central claim quickly."),
			node("persuasive-audience", "Audience Alignment", "audience alignment", "Write for a specific audience's values, fears, and incentives.", "core", 2, "The rhetoric feels aimed, not generic."),
			node("persuasive-reasoning", "Reasoning Quality", "reasoning quality", "Support the claim with reasons that actually connect to it.", "core", 3, "The argument has visible logical support."),
			node("persuasive-stakes", "Stakes", "claim clarity", "Make clear why the reader should care now.", "core", 4, "The urgency or importance of the issue is visible.", "persuasive-claim"),
			node("persuasive-positioning", "Positioning", "audience alignment", "Frame the issue in terms the audience already recognizes as meaningful.", "core", 5, "The opening meets the reader in their world.", "persuasive-audience"),
			node("persuasive-thesis-narrowing", "Thesis Narrowing", "claim clarity", "Keep the persuasive goal specific enough to win.", "core", 6, "The claim is actionable and arguable, not vague.", "persuasive-claim"),
			node("persuasive-reason-order", "Reason Order", "reasoning quality", "Sequence reasons from strongest or most useful to weakest.", "core", 7, "The argument gains rather than loses force as it unfolds.", "persuasive-reasoning"),
			node("persuasive-evidence-basics", "Evidence Basics", "evidence integration", "Use evidence that strengthens persuasion rather than merely filling space.", "core", 8, "Support makes the case more convincing.", "persuasive-reasoning"),
			node("persuasive-tone-basics", "Tone Basics", "rhetorical force", "Sound forceful without sounding hysterical or smug.", "core", 9, "Tone helps persuasion rather than fighting it.", "persuasive-audience"),
			node("persuasive-action-target", "Action Target", "actionability", "Name the concrete decision, change, or response you want.", "core", 10, "The reader knows what the writing is trying to move.", "persuasive-stakes"),
			node("persuasive-opening-hook", "Opening Hook", "rhetorical force", "Open with a tension, cost, example, or contradiction that earns attention.", "structure", 11, "The opening makes the reader care enough to continue.", "persuasive-stakes"),
			node("persuasive-structure", "Argument Structure", "structural signposting", "Shape the piece so claim, reasons, evidence, and ask are easy to follow.", "structure", 12, "The persuasive line is legible from start to finish.", "persuasive-reason-order"),
			node("persuasive-transitions", "Transitions", "structural signposting", "Use transitions that show how each reason advances the case.", "structure", 13, "The piece feels cumulative rather than assembled.", "persuasive-structure"),
			node("persuasive-paragraph-unity", "Paragraph Unity", "reasoning quality", "Keep each paragraph focused on one persuasive move.", "structure", 14, "Paragraphs advance the case cleanly.", "persuasive-structure"),
			node("persuasive-conclusion", "Conclusion", "actionability", "End with consequence and next move rather than vague restatement.", "structure", 15, "The ending leaves the reader with a sharpened direction.", "persuasive-action-target"),
			node("persuasive-reframing", "Reframing", "rhetorical force", "Shift how the reader sees the issue, not just what facts they know.", "structure", 16, "The piece changes the interpretive frame, not only the information set.", "persuasive-positioning"),
			node("persuasive-ethos", "Ethos", "rhetorical force", "Build credibility through judgment, fairness, and useful specifics.", "appeal", 17, "The reader trusts the writer more as the piece goes on.", "persuasive-tone-basics"),
			node("persuasive-logos", "Logos", "reasoning quality", "Make the logic of the case visible and hard to evade.", "appeal", 18, "The argument's reasoning is easy to follow and hard to dismiss casually.", "persuasive-reasoning"),
			node("persuasive-pathos", "Pathos", "rhetorical force", "Use emotion to intensify stakes without replacing reasoning.", "appeal", 19, "Emotion strengthens the case rather than distorting it.", "persuasive-stakes"),
			node("persuasive-example-selection", "Example Selection", "evidence integration", "Choose examples that carry real persuasive leverage.", "appeal", 20, "Examples feel strategically chosen.", "persuasive-evidence-basics"),
			node("persuasive-analogy", "Analogy", "reasoning quality", "Use analogy where it clarifies the argument instead of oversimplifying it.", "appeal", 21, "Analogies sharpen rather than blur the case.", "persuasive-logos"),
			node("persuasive-story-use", "Story Use", "rhetorical force", "Use narrative examples to humanize stakes without drifting into sentimentality.", "appeal", 22, "Story deepens urgency and clarity together.", "persuasive-pathos"),
			node("persuasive-sourced-authority", "Sourced Authority", "evidence integration", "Use outside authorities to support the case without outsourcing the argument.", "appeal", 23, "Authorities strengthen the case but do not replace the writer's reasoning.", "persuasive-ethos", "persuasive-evidence-basics"),
			node("persuasive-objection-forecast", "Objection Forecasting", "objection handling", "Anticipate the strongest resistance before the reader forms it fully.", "objection", 24, "The argument stays ahead of likely pushback.", "persuasive-audience", "persuasive-logos"),
			node("persuasive-counterargument", "Counterargument Handling", "objection handling", "Represent the opposing case fairly before answering it.", "objection", 25, "The piece gains credibility by meeting opposition honestly.", "persuasive-objection-forecast"),
			node("persuasive-concession", "Concession", "objection handling", "Concede what is true or fair without surrendering the case.", "objection", 26, "The writer sounds more trustworthy and less brittle.", "persuasive-counterargument"),
			node("persuasive-tradeoffs", "Tradeoffs", "reasoning quality", "Acknowledge costs and limitations rather than pretending the proposal is free.", "objection", 27, "The argument feels more real and therefore more convincing.", "persuasive-concession"),
			node("persuasive-fallacy-avoidance", "Fallacy Avoidance", "reasoning quality", "Reduce strawman, false choice, and slippery logic errors.", "objection", 28, "The argument becomes cleaner under scrutiny.", "persuasive-counterargument"),
			node("persuasive-moral-language", "Moral Language Control", "rhetorical force", "Use moral framing carefully so it clarifies values instead of inflaming noise.", "objection", 29, "Value language sharpens the issue without becoming shrill.", "persuasive-pathos"),
			node("persuasive-data-use", "Data Use", "evidence integration", "Use numbers where they sharpen credibility and consequence.", "evidence", 30, "Quantitative support improves the case's force.", "persuasive-evidence-basics"),
			node("persuasive-quote-restraint", "Quote Restraint", "evidence integration", "Quote only what helps the reader trust or grasp the case.", "evidence", 31, "Quoted material remains subordinate to persuasion.", "persuasive-sourced-authority"),
			node("persuasive-case-study", "Case Study", "evidence integration", "Use case studies to make abstract claims concrete.", "evidence", 32, "Cases illuminate the argument's real-world stakes.", "persuasive-example-selection"),
			node("persuasive-comparison", "Comparison", "reasoning quality", "Compare options or outcomes in a way that clarifies the reader's choice.", "evidence", 33, "Comparison sharpens the decision frame.", "persuasive-tradeoffs"),
			node("persuasive-consequence-modeling", "Consequence Modeling", "reasoning quality", "Show what likely follows if the reader accepts or rejects the claim.", "evidence", 34, "Consequences become easier to imagine and weigh.", "persuasive-stakes", "persuasive-comparison"),
			node("persuasive-plainstyle", "Plainstyle", "clarity and coherence", "Prefer direct language over inflated or slogan-heavy prose.", "style", 35, "The writing sounds clearer and more trustworthy.", "persuasive-tone-basics"),
			node("persuasive-cadence", "Cadence", "rhetorical force", "Use sentence rhythm to intensify emphasis and release.", "style", 36, "The prose sounds more deliberate and memorable.", "persuasive-plainstyle"),
			node("persuasive-repetition", "Strategic Repetition", "rhetorical force", "Repeat key language with variation so it sticks.", "style", 37, "The argument gains memorability without sounding lazy.", "persuasive-cadence"),
			node("persuasive-slogan-restraint", "Slogan Restraint", "clarity and coherence", "Avoid reducing the piece to catchphrases when the issue needs reasoning.", "style", 38, "Memorable language does not crowd out thought.", "persuasive-repetition"),
			node("persuasive-verb-force", "Verb Force", "rhetorical force", "Choose verbs that carry decision and consequence.", "style", 39, "The prose becomes more active and persuasive at the line level.", "persuasive-plainstyle"),
			node("persuasive-pivot-lines", "Pivot Lines", "structural signposting", "Write lines that turn the reader from one phase of persuasion to the next.", "style", 40, "Important turns feel guided rather than abrupt.", "persuasive-transitions"),
			node("persuasive-audience-segmentation", "Audience Segmentation", "audience alignment", "Adjust the case for sub-audiences with different stakes or objections.", "advanced", 41, "The writer can adapt one issue across multiple reader groups.", "persuasive-audience", "persuasive-objection-forecast"),
			node("persuasive-medium-fit", "Medium Fit", "audience alignment", "Shape argument length, density, and tone to the medium.", "advanced", 42, "The persuasive strategy fits the channel.", "persuasive-positioning"),
			node("persuasive-campaign-thinking", "Campaign Thinking", "actionability", "Write one persuasive piece as part of a larger sequence, not a one-off burst.", "advanced", 43, "The piece knows what stage of persuasion it belongs to.", "persuasive-action-target", "persuasive-medium-fit"),
			node("persuasive-ethical-pressure", "Ethical Pressure", "objection handling", "Press the case ethically without sliding into manipulation.", "advanced", 44, "The rhetoric stays strong without becoming coercive or dishonest.", "persuasive-moral-language", "persuasive-concession"),
			node("persuasive-revision-triage", "Revision Triage", "revision habits", "Fix claim, audience, and structure before polishing rhetoric.", "revision", 45, "Revisions improve the case in the right order.", "persuasive-structure"),
			node("persuasive-logic-pass", "Logic Pass", "revision habits", "Revise once only for weak reasoning, unsupported leaps, and hidden assumptions.", "revision", 46, "The case becomes harder to dismantle.", "persuasive-revision-triage"),
			node("persuasive-objection-pass", "Objection Pass", "revision habits", "Revise once only for resistance, counterargument, and audience friction.", "revision", 47, "The argument gets better at surviving contact with disagreement.", "persuasive-revision-triage"),
			node("persuasive-rhetoric-pass", "Rhetoric Pass", "revision habits", "Revise once only for tone, cadence, and memorable phrasing.", "revision", 48, "The prose becomes more compelling without losing clarity.", "persuasive-revision-triage"),
			node("persuasive-evidence-pass", "Evidence Pass", "revision habits", "Revise once only for support quality, relevance, and proportion.", "revision", 49, "Support becomes more convincing and better placed.", "persuasive-revision-triage"),
			node("persuasive-headline", "Headline Framing", "claim clarity", "Title the piece so the claim or tension is immediately visible.", "revision", 50, "The title attracts the right reader with the right promise.", "persuasive-opening-hook"),
			node("persuasive-call-to-action", "Call to Action", "actionability", "End with a next step the reader can actually imagine taking.", "revision", 51, "The piece converts urgency into direction.", "persuasive-conclusion", "persuasive-action-target"),
			node("persuasive-memory-line", "Memory Line", "rhetorical force", "Leave the reader with one final line or image that carries the case forward.", "revision", 52, "The piece stays with the reader after it ends.", "persuasive-rhetoric-pass", "persuasive-call-to-action"),
		},
	}

	memoir := TGOTreeDefinition{
		Slug:           "memoir-personal-narrative-track",
		Title:          "Memoir and Personal Narrative Track",
		Description:    "Personal narrative track for lived experience, reflection, scene work, memory handling, and essayistic meaning.",
		SeedCodes:      []string{"memoir-scene-grounding", "memoir-voice-presence", "memoir-reflection-basics"},
		PrioritySkills: []string{"scene architecture", "voice presence", "reflection depth", "emotional compression", "narrative clarity", "image freshness", "revision habits"},
		TGOs: []TGO{
			node("memoir-scene-grounding", "Scene Grounding", "scene architecture", "Anchor scenes in time, place, and embodied detail.", "core", 1, "The reader can enter the memory without confusion."),
			node("memoir-voice-presence", "Voice Presence", "voice presence", "Let the narrator's sensibility feel present on the page.", "core", 2, "The narrator sounds like a person rather than a report."),
			node("memoir-reflection-basics", "Reflection Basics", "reflection depth", "Move beyond recounting to say what the remembered material means.", "core", 3, "The piece interprets experience instead of only relaying it."),
			node("memoir-causal-thread", "Causal Thread", "narrative clarity", "Make clear how one remembered moment leads into the next.", "core", 4, "The reader can follow the movement of the narrative without guessing.", "memoir-scene-grounding"),
			node("memoir-selective-detail", "Selective Detail", "image freshness", "Choose details that carry memory, mood, or significance.", "core", 5, "Details feel revealing, not merely exhaustive.", "memoir-scene-grounding"),
			node("memoir-emotional-compression", "Emotional Compression", "emotional compression", "Let feeling arrive through concrete image, gesture, and consequence.", "core", 6, "Emotion lands through lived detail rather than labels.", "memoir-scene-grounding"),
			node("memoir-time-awareness", "Time Awareness", "narrative clarity", "Keep the reader oriented when the piece moves across time.", "core", 7, "Time shifts stay legible.", "memoir-causal-thread"),
			node("memoir-memory-honesty", "Memory Honesty", "reflection depth", "Acknowledge uncertainty where memory is partial or unstable.", "core", 8, "The narrator sounds credible because certainty is calibrated.", "memoir-voice-presence"),
			node("memoir-scene-vs-summary", "Scene vs Summary", "scene architecture", "Balance dramatized scenes with compressive summary.", "core", 9, "The piece knows when to dwell and when to move.", "memoir-scene-grounding"),
			node("memoir-narrator-position", "Narrator Position", "voice presence", "Clarify who is telling the story now and how that self differs from the one in the scene.", "core", 10, "The distance between experiencing self and narrating self is usable.", "memoir-voice-presence"),
			node("memoir-scene-turn", "Scene Turn", "scene architecture", "Build scenes that change pressure, understanding, or relationship.", "scene", 11, "Scenes end somewhere different from where they began.", "memoir-scene-vs-summary"),
			node("memoir-dialogue-use", "Dialogue Use", "voice presence", "Use dialogue to sharpen scene and character rather than reconstruct everything said.", "scene", 12, "Dialogue feels selective and meaningful.", "memoir-scene-grounding"),
			node("memoir-gesture-subtext", "Gesture as Subtext", "emotional compression", "Use gesture and body detail to carry what cannot be said directly.", "scene", 13, "Physical detail reveals hidden feeling.", "memoir-emotional-compression"),
			node("memoir-setting-pressure", "Setting Pressure", "scene architecture", "Let place and environment push on the remembered action.", "scene", 14, "Setting matters because it shapes the scene's pressure.", "memoir-selective-detail"),
			node("memoir-status-awareness", "Status Awareness", "narrative clarity", "Notice family, social, or institutional power in scenes.", "scene", 15, "Power relationships become visible through scene detail.", "memoir-scene-turn"),
			node("memoir-scene-entry-exit", "Scene Entry and Exit", "scene architecture", "Enter and leave scenes at pressure points rather than dead air.", "scene", 16, "Scenes begin and end with purpose.", "memoir-scene-turn"),
			node("memoir-voice-texture", "Voice Texture", "voice presence", "Develop sentences that feel shaped by the narrator's mind and ear.", "voice", 17, "The prose sounds like this narrator and no one else.", "memoir-voice-presence"),
			node("memoir-diction-precision", "Diction Precision", "voice presence", "Choose words that feel exact to memory and sensibility.", "voice", 18, "Language feels lived and deliberate.", "memoir-voice-texture"),
			node("memoir-rhythm", "Rhythm", "voice presence", "Control cadence so the prose can widen, tighten, or pause intentionally.", "voice", 19, "Sentence movement supports feeling and thought.", "memoir-diction-precision"),
			node("memoir-humor-control", "Humor Control", "voice presence", "Use humor where it reveals narrator and pressure rather than deflecting everything.", "voice", 20, "Humor feels purposeful, not evasive.", "memoir-voice-texture"),
			node("memoir-vulnerability", "Vulnerability", "reflection depth", "Allow risk and self-implication onto the page.", "voice", 21, "The piece feels exposed enough to matter.", "memoir-memory-honesty"),
			node("memoir-self-myth-avoidance", "Self-Myth Avoidance", "reflection depth", "Resist flattening the self into hero, victim, or saint.", "voice", 22, "The narrator appears humanly mixed and therefore credible.", "memoir-vulnerability"),
			node("memoir-reflection-depth", "Reflection Depth", "reflection depth", "Draw insight from experience without converting the piece into lecture.", "reflection", 23, "Reflection adds meaning without stopping the narrative cold.", "memoir-reflection-basics"),
			node("memoir-question-holding", "Question Holding", "reflection depth", "Let some questions stay open long enough to deepen the piece.", "reflection", 24, "The essay can hold uncertainty without collapsing into vagueness.", "memoir-memory-honesty"),
			node("memoir-pattern-linking", "Pattern Linking", "reflection depth", "Show how separate memories speak to one another.", "reflection", 25, "The piece builds meaning across repeated moments.", "memoir-reflection-depth"),
			node("memoir-present-insight", "Present Insight", "reflection depth", "Make clear what the narrating self knows now that the past self could not.", "reflection", 26, "The reflective layer has genuine vantage.", "memoir-narrator-position"),
			node("memoir-theme-pressure", "Theme Through Pressure", "reflection depth", "Let theme emerge from scenes and reflection together.", "reflection", 27, "The piece says something without flattening into slogan.", "memoir-pattern-linking"),
			node("memoir-memory-fragment", "Fragment Handling", "narrative clarity", "Use fragments of memory intentionally rather than as accidental confusion.", "reflection", 28, "Fragmented structure still feels shaped.", "memoir-time-awareness", "memoir-memory-honesty"),
			node("memoir-metaphor-restraint", "Metaphor Restraint", "image freshness", "Use figurative language where it sharpens memory instead of sentimentalizing it.", "reflection", 29, "Metaphor deepens the piece without fogging it.", "memoir-selective-detail"),
			node("memoir-relationship-portraiture", "Relationship Portraiture", "scene architecture", "Render important relationships through pressure, habit, and scene.", "character", 30, "Relationships feel dynamic rather than summarized.", "memoir-scene-turn"),
			node("memoir-self-portraiture", "Self Portraiture", "voice presence", "Render the past self with enough specificity to feel embodied and limited.", "character", 31, "The remembered self feels like a real person in time.", "memoir-narrator-position"),
			node("memoir-character-complexity", "Character Complexity", "reflection depth", "Allow important figures to remain mixed, contradictory, or partially unknowable.", "character", 32, "People on the page resist flattening.", "memoir-self-myth-avoidance"),
			node("memoir-dialogic-tension", "Dialogic Tension", "voice presence", "Let spoken exchange reveal pressure beneath the words.", "character", 33, "Dialogue carries more than surface information.", "memoir-dialogue-use"),
			node("memoir-loyalty-pressure", "Loyalty Pressure", "reflection depth", "Write family and intimate bonds with real competing obligations.", "character", 34, "Relationships complicate truth-telling in the piece.", "memoir-relationship-portraiture"),
			node("memoir-shame-handling", "Shame Handling", "emotional compression", "Approach shame directly enough to matter without turning it into spectacle.", "character", 35, "The piece can hold shame without melodrama.", "memoir-vulnerability"),
			node("memoir-body-awareness", "Body Awareness", "image freshness", "Use bodily perception and sensation to ground memory and reflection.", "character", 36, "The piece remembers through the body, not just the intellect.", "memoir-gesture-subtext"),
			node("memoir-structure-linearity", "Linear Structure", "narrative clarity", "Use linear order well when the material benefits from it.", "structure", 37, "Chronology serves the piece instead of flattening it.", "memoir-time-awareness"),
			node("memoir-structure-braiding", "Braided Structure", "narrative clarity", "Braid multiple time strands or subjects without losing the reader.", "structure", 38, "Braided pieces feel shaped rather than scattered.", "memoir-pattern-linking", "memoir-time-awareness"),
			node("memoir-structure-framing", "Framing Device", "scene architecture", "Use present-day framing scenes or motifs to hold the piece together.", "structure", 39, "The frame helps organize meaning, not just decorate it.", "memoir-present-insight"),
			node("memoir-opening-pressure", "Opening Pressure", "voice presence", "Open with enough tension, image, or question to earn the reader's attention.", "structure", 40, "The piece begins with force rather than throat-clearing.", "memoir-scene-grounding"),
			node("memoir-ending-resonance", "Ending Resonance", "reflection depth", "End with an image, realization, or release that continues to echo.", "structure", 41, "The ending deepens the piece's meaning instead of merely stopping.", "memoir-theme-pressure"),
			node("memoir-chapter-pressure", "Chapter Pressure", "scene architecture", "If the piece is longer, end chapters on pressure, discovery, or altered relation.", "structure", 42, "Longer-form memoir continues to pull forward.", "memoir-scene-entry-exit"),
			node("memoir-image-freshness", "Image Freshness", "image freshness", "Prefer singular, remembered imagery over stock emotional shorthand.", "style", 43, "Images feel specific to this life and this scene.", "memoir-selective-detail"),
			node("memoir-summary-compression", "Summary Compression", "narrative clarity", "Compress long stretches of time without losing force or coherence.", "style", 44, "Summary passages remain vivid and useful.", "memoir-scene-vs-summary"),
			node("memoir-reflective-pivot", "Reflective Pivot", "reflection depth", "Turn cleanly from remembered scene into reflective meaning and back again.", "style", 45, "The piece moves between scene and thought without clunkiness.", "memoir-reflection-depth"),
			node("memoir-repetition-control", "Repetition Control", "voice presence", "Keep recurring phrases or image clusters intentional rather than accidental.", "style", 46, "Repetition becomes pattern instead of drag.", "memoir-voice-texture"),
			node("memoir-line-editing", "Line Editing", "revision habits", "Do a late pass for diction, rhythm, and repetition after structural work.", "revision", 47, "Language sharpens without destabilizing the piece.", "memoir-rhythm"),
			node("memoir-revision-triage", "Revision Triage", "revision habits", "Fix scene, structure, and reflection before polishing style.", "revision", 48, "Revision energy goes to the highest-value memoir problems first.", "memoir-structure-braiding", "memoir-reflection-depth"),
			node("memoir-clarity-pass", "Clarity Pass", "revision habits", "Revise once only for orientation in time, scene, and narrator position.", "revision", 49, "The reader gets lost less often in revision drafts.", "memoir-revision-triage"),
			node("memoir-reflection-pass", "Reflection Pass", "revision habits", "Revise once only for depth, honesty, and interpretive force.", "revision", 50, "The reflective layer gains weight and necessity.", "memoir-revision-triage"),
			node("memoir-ethical-pass", "Ethical Pass", "revision habits", "Revise with attention to fairness, privacy, and the costs of representation.", "revision", 51, "The piece gets ethically sharper without becoming timid.", "memoir-character-complexity", "memoir-revision-triage"),
			node("memoir-final-shape", "Final Shape", "revision habits", "Make the last structural decisions so the piece feels inevitable in its finished form.", "revision", 52, "The final draft feels fully composed rather than merely accumulated.", "memoir-ending-resonance", "memoir-line-editing"),
		},
	}

	trees := []TGOTreeDefinition{mythic, youth, story, thought, professional, academic, technical, persuasive, memoir}
	return mythic, youth, story, thought, professional, academic, technical, persuasive, memoir, trees, skillMap
}

var (
	mythicTragedyTree, youthFoundationsTree, storyCraftTree, thoughtLeadershipTree, professionalWritingTree, academicEssayTree, technicalWritingTree, persuasiveWritingTree, memoirNarrativeTree, BuiltInTrees, TGOCodeToSkill = buildBuiltInCatalog()
	fantasyFictionTree, scienceFictionTree, romanceFictionTree, literaryFictionTree, mysteryThrillerTree                                                                                                                       = buildExpandedFictionTrees()
)

func init() {
	BuiltInTrees = append(BuiltInTrees, fantasyFictionTree, scienceFictionTree, romanceFictionTree, literaryFictionTree, mysteryThrillerTree)
	registerTreeSkills(fantasyFictionTree, scienceFictionTree, romanceFictionTree, literaryFictionTree, mysteryThrillerTree)
}

func buildExpandedFictionTrees() (TGOTreeDefinition, TGOTreeDefinition, TGOTreeDefinition, TGOTreeDefinition, TGOTreeDefinition) {
	return cloneTreeDefinition(storyCraftTree, cloneTreeOptions{
			Slug:           "fantasy-fiction-track",
			Title:          "Fantasy Track",
			Description:    "Fiction track for fantasy writers building scene control, world pressure, character stakes, and durable narrative movement.",
			CodePrefix:     "fantasy",
			PrioritySkills: []string{"worldbuilding economy", "image freshness", "scene architecture", "narrative clarity", "dialogue intelligence", "emotional compression", "prose precision"},
		}),
		cloneTreeDefinition(storyCraftTree, cloneTreeOptions{
			Slug:           "science-fiction-track",
			Title:          "Science Fiction Track",
			Description:    "Fiction track for science fiction writers building idea pressure, scene clarity, world logic, and character consequence.",
			CodePrefix:     "scifi",
			PrioritySkills: []string{"worldbuilding economy", "narrative clarity", "scene architecture", "prose precision", "dialogue intelligence", "image freshness", "structure and pacing"},
		}),
		cloneTreeDefinition(storyCraftTree, cloneTreeOptions{
			Slug:           "romance-fiction-track",
			Title:          "Romance Track",
			Description:    "Fiction track for romance writers building relationship pressure, scene turns, emotional movement, and clear character stakes.",
			CodePrefix:     "romance",
			PrioritySkills: []string{"scene architecture", "emotional compression", "dialogue intelligence", "story development", "narrative clarity", "voice presence", "prose precision"},
		}),
		cloneTreeDefinition(storyCraftTree, cloneTreeOptions{
			Slug:           "literary-fiction-track",
			Title:          "Literary Fiction Track",
			Description:    "Fiction track for literary writers building scene control, image discipline, emotional pressure, and formal clarity.",
			CodePrefix:     "literary",
			PrioritySkills: []string{"image freshness", "emotional compression", "prose precision", "narrative clarity", "scene architecture", "dialogue intelligence", "story development"},
		}),
		cloneTreeDefinition(storyCraftTree, cloneTreeOptions{
			Slug:           "mystery-thriller-track",
			Title:          "Mystery and Thriller Track",
			Description:    "Fiction track for mystery and thriller writers building suspense, clue control, scene pressure, and reader orientation.",
			CodePrefix:     "thriller",
			PrioritySkills: []string{"narrative clarity", "scene architecture", "structure and pacing", "dialogue intelligence", "worldbuilding economy", "prose precision", "story development"},
		})
}

type cloneTreeOptions struct {
	Slug           string
	Title          string
	Description    string
	CodePrefix     string
	PrioritySkills []string
}

func cloneTreeDefinition(base TGOTreeDefinition, options cloneTreeOptions) TGOTreeDefinition {
	codeMap := make(map[string]string, len(base.TGOs))
	for _, tgo := range base.TGOs {
		codeMap[tgo.Code] = options.CodePrefix + "-" + tgo.Code
	}

	out := TGOTreeDefinition{
		Slug:           options.Slug,
		Title:          options.Title,
		Description:    options.Description,
		PrioritySkills: append([]string(nil), options.PrioritySkills...),
	}
	for _, code := range base.SeedCodes {
		if next, ok := codeMap[code]; ok {
			out.SeedCodes = append(out.SeedCodes, next)
		}
	}
	for _, tgo := range base.TGOs {
		clone := tgo
		clone.Code = codeMap[tgo.Code]
		clone.Prerequisites = remapCodes(tgo.Prerequisites, codeMap)
		out.TGOs = append(out.TGOs, clone)
	}
	return out
}

func remapCodes(values []string, codeMap map[string]string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if mapped, ok := codeMap[value]; ok {
			out = append(out, mapped)
			continue
		}
		out = append(out, value)
	}
	return out
}

func registerTreeSkills(trees ...TGOTreeDefinition) {
	for _, tree := range trees {
		for _, tgo := range tree.TGOs {
			if skill := TGOCodeToSkill[trimStoryClonePrefix(tgo.Code)]; skill != "" {
				TGOCodeToSkill[tgo.Code] = skill
			}
		}
	}
}

func trimStoryClonePrefix(code string) string {
	for _, prefix := range []string{"fantasy-", "scifi-", "romance-", "literary-", "thriller-"} {
		if len(code) > len(prefix) && code[:len(prefix)] == prefix {
			return code[len(prefix):]
		}
	}
	return code
}

func treeForTemplateKey(templateKey string) TGOTreeDefinition {
	switch templateKey {
	case "youth-foundations":
		return youthFoundationsTree
	case "academic-essay":
		return academicEssayTree
	case "technical-writing":
		return technicalWritingTree
	case "persuasive-writing":
		return persuasiveWritingTree
	case "memoir-personal-narrative":
		return memoirNarrativeTree
	case "thought-leadership":
		return thoughtLeadershipTree
	case "professional-writing":
		return professionalWritingTree
	case "fantasy-fiction":
		return fantasyFictionTree
	case "science-fiction":
		return scienceFictionTree
	case "romance-fiction":
		return romanceFictionTree
	case "literary-fiction":
		return literaryFictionTree
	case "mystery-thriller":
		return mysteryThrillerTree
	case "story-craft":
		return storyCraftTree
	default:
		return storyCraftTree
	}
}

func GlobalSkillGraphDefinition() TGOTreeDefinition {
	prioritySeen := map[string]bool{}
	var priority []string
	var tgos []TGO
	for _, tree := range BuiltInTrees {
		for _, skill := range tree.PrioritySkills {
			if prioritySeen[skill] {
				continue
			}
			prioritySeen[skill] = true
			priority = append(priority, skill)
		}
		tgos = append(tgos, tree.TGOs...)
	}
	return TGOTreeDefinition{
		Slug:           GlobalSkillGraphSlug,
		Title:          GlobalSkillGraphTitle,
		Description:    "A global writing skill graph built from all curated curriculum regions.",
		SeedCodes:      []string{"sentence-clarity", "story-causal-clarity", "claim-clarity"},
		PrioritySkills: priority,
		TGOs:           tgos,
	}
}

func SkillGraphFromBuiltIns() SkillGraph {
	unlocks := map[string][]string{}
	for _, tree := range BuiltInTrees {
		for _, tgo := range tree.TGOs {
			for _, prereq := range tgo.Prerequisites {
				unlocks[prereq] = append(unlocks[prereq], tgo.Code)
			}
		}
	}
	var nodes []SkillGraphNode
	for _, tree := range BuiltInTrees {
		for _, tgo := range tree.TGOs {
			next := append([]string(nil), unlocks[tgo.Code]...)
			sort.Strings(next)
			nodes = append(nodes, SkillGraphNode{
				TGO:             tgo,
				SourceTreeSlug:  tree.Slug,
				SourceTreeTitle: tree.Title,
				Unlocks:         next,
			})
		}
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].StageOrder == nodes[j].StageOrder {
			return nodes[i].Code < nodes[j].Code
		}
		return nodes[i].StageOrder < nodes[j].StageOrder
	})
	var regions []SkillGraphRegion
	for _, tree := range BuiltInTrees {
		nodeCodes := make([]string, 0, len(tree.TGOs))
		for _, tgo := range tree.TGOs {
			nodeCodes = append(nodeCodes, tgo.Code)
		}
		regions = append(regions, SkillGraphRegion{
			Slug:           tree.Slug,
			Title:          tree.Title,
			Description:    tree.Description,
			SeedCodes:      append([]string(nil), tree.SeedCodes...),
			PrioritySkills: append([]string(nil), tree.PrioritySkills...),
			NodeCodes:      nodeCodes,
		})
	}
	return SkillGraph{
		Slug:        GlobalSkillGraphSlug,
		Title:       GlobalSkillGraphTitle,
		Description: "A single unlockable writing skill graph spanning fiction, nonfiction, professional, and developmental writing.",
		Regions:     regions,
		Nodes:       nodes,
	}
}

func RecommendedStarterCodes(profile OnboardingProfile) []string {
	tree := treeForTemplateKey(TemplateKeyForProfile(profile))
	return append([]string(nil), tree.SeedCodes...)
}

func RecommendedRegionSlugs(profile OnboardingProfile) []string {
	primary := treeForTemplateKey(TemplateKeyForProfile(profile)).Slug
	regions := []string{primary}
	switch primary {
	case storyCraftTree.Slug:
		regions = append(regions, fantasyFictionTree.Slug, memoirNarrativeTree.Slug)
	case fantasyFictionTree.Slug:
		regions = append(regions, storyCraftTree.Slug, scienceFictionTree.Slug)
	case scienceFictionTree.Slug:
		regions = append(regions, fantasyFictionTree.Slug, mysteryThrillerTree.Slug)
	case romanceFictionTree.Slug:
		regions = append(regions, literaryFictionTree.Slug, memoirNarrativeTree.Slug)
	case literaryFictionTree.Slug:
		regions = append(regions, memoirNarrativeTree.Slug, storyCraftTree.Slug)
	case mysteryThrillerTree.Slug:
		regions = append(regions, scienceFictionTree.Slug, storyCraftTree.Slug)
	case youthFoundationsTree.Slug:
		regions = append(regions, storyCraftTree.Slug, academicEssayTree.Slug)
	case thoughtLeadershipTree.Slug:
		regions = append(regions, persuasiveWritingTree.Slug, professionalWritingTree.Slug)
	case professionalWritingTree.Slug:
		regions = append(regions, technicalWritingTree.Slug, persuasiveWritingTree.Slug)
	case academicEssayTree.Slug:
		regions = append(regions, thoughtLeadershipTree.Slug, persuasiveWritingTree.Slug)
	case technicalWritingTree.Slug:
		regions = append(regions, professionalWritingTree.Slug, academicEssayTree.Slug)
	case persuasiveWritingTree.Slug:
		regions = append(regions, thoughtLeadershipTree.Slug, academicEssayTree.Slug)
	case memoirNarrativeTree.Slug:
		regions = append(regions, storyCraftTree.Slug, thoughtLeadershipTree.Slug)
	}
	return regions
}
