# Skill Detail Voice and Teaching Guidelines

Status: Adopted design guideline

Owner: Product + Pedagogy + UI

Last updated: 2026-04-12

## 1) Purpose

This document defines the teaching voice and content rules for all skill detail pages in the app.

Goals:
- Keep explanations clear at a 5th-grade reading level (or simpler).
- Teach grammar and writing with practical, real-world examples.
- Show what to do next, not just what is wrong.
- Preserve writer voice while improving craft.

## 2) Source Models We Are Adapting

### 2.1 Grammar Girl model (Mignon Fogarty / Quick and Dirty Tips)

What we are borrowing:
- Short, focused lessons (usually brief, often framed around one question).
- Plain-language explanations over technical jargon.
- Memorable examples and rewrites that show "before -> after."
- Practical editing moves (cut deadwood, front-load key info, use strong verbs).
- Reader/listener question framing to make lessons feel human and relevant.

Evidence:
- Grammar Girl page shows listener-question framing and short tip format.
- QDT site positions content as short-form tips (often around ten minutes or less).
- "How to Write Clear Sentences" demonstrates front-load, organize, cut deadwood, and rewrite examples.

### 2.2 Purdue OWL model (widely used public writing instruction)

What we are borrowing:
- Audience + purpose fit as a first principle.
- Stepwise process checklists (before, during, after proofreading/revision).
- Error pattern training (identify your repeat errors, then check deliberately).
- Direct examples of incorrect vs corrected forms.
- Clarity-first style guidance: concise, plain writing; active voice by default where appropriate.

Evidence:
- OWL proofreading pages use explicit step-by-step strategy and common-error checklists.
- OWL style pages emphasize plain, clear style and avoiding unnecessary wordiness.
- OWL tone pages stress confidence, courtesy, appropriate difficulty, and reader benefit.

## 3) Combined Method for Writing Coach Skill Pages

We use a hybrid teaching voice:
- Grammar Girl for tone and accessibility.
- Purdue OWL for structure and instructional rigor.

Working rule:
- Explain one concept at a time (small chunk).
- Show one concrete example pair (strong vs weak).
- Give 2-4 immediate revision moves.
- End with one clear "next action".

## 4) Content Contract for Every Skill Detail Page

Every skill page must include these fields (already supported by the structured data model):
- `oneLine`: one-sentence summary of the skill.
- `whatItMeans`: plain-language definition.
- `whyItMatters`: impact on reader understanding/trust.
- `lookFor[]`: observable checks a writer can use.
- `strongExample`: short positive example.
- `weakExample`: short negative example.
- `revisionMoves[]`: specific edit actions.
- `coachTip`: short encouragement + focus cue.

Optional extension fields (recommended for future):
- `quickTest`: one yes/no self-check.
- `commonTrap`: frequent mistake in this skill.
- `exceptionCase`: when a normal rule can be bent.
- `voiceSafeRewrite`: improvement that keeps author voice.

## 5) Voice and Language Rules (UI Copy)

### 5.1 Readability rules

- Target 5th-grade reading level.
- Prefer common words over abstract terms.
- Keep sentences short: usually 8-16 words.
- One idea per sentence.
- Use active voice by default.

### 5.2 Tone rules

- Friendly, direct, calm.
- No shaming language.
- No "gotcha" phrasing.
- Avoid robotic command tone.
- Preserve stylistic freedom: suggest options, not one "correct" voice.

### 5.3 Instruction rules

- Tell writers what to do next in concrete verbs: cut, add, name, reorder, clarify.
- Prioritize high-leverage edits before polish.
- Use sequence language: First, Next, Then.
- If scoring is shown, connect score to visible evidence.

## 6) Rule-Explanation Pattern (for Grammar/Mechanics Skills)

When explaining a grammar or usage rule, follow this pattern:
1. Name the rule in plain words.
2. Show a quick test (how to detect it).
3. Show one weak example.
4. Show one stronger rewrite.
5. Note one exception or context where variation is okay.

## 7) High-Value English Rule Areas to Cover Across Skill Pages

These are common pain points and should be covered where relevant:
- Sentence fragments.
- Run-on sentences and comma splices.
- Subject-verb agreement.
- Pronoun reference clarity.
- Parallel structure in lists/series.
- Apostrophes (possessives vs contractions).
- Active vs passive voice choice.
- Wordiness and filler removal.
- Register/tone fit for audience.
- Inclusive, non-discriminatory language.

## 8) UI Behavior Guidelines

- Skill names should link to skill detail pages from all major contexts.
- Content should scan fast:
  - short headings,
  - short paragraphs,
  - bullet lists for checks/moves,
  - clear example cards.
- Show "Stronger" and "Needs work" labels (not "good/bad").
- Keep examples realistic and everyday.
- Keep evidence and action close together in layout.

## 9) Quality Checklist for Skill Detail Content

Before publishing or updating a skill page, verify:
- The one-line summary is understandable to a middle-school reader.
- At least one example pair is concrete and specific.
- Revision moves are actionable and testable.
- Wording does not force a single writing voice.
- Guidance is consistent with our pedagogy (focus -> evidence -> revision -> progression).
- No unexplained jargon remains.

## 10) Authoring Workflow

When drafting or revising skill-page content:
1. Start from internal skill objective intent.
2. Draft in plain language using this guideline.
3. Add strong/weak examples.
4. Add 2-4 revision moves.
5. Run internal review for readability and actionability.
6. Publish and monitor learner confusion signals.

## 11) Implementation Notes for Current App

Current structured fields in `web/src/lib/skill-details.ts` are aligned with this guideline.

Recommended next enhancements:
- Add optional `quickTest`, `commonTrap`, `exceptionCase`, and `voiceSafeRewrite` fields.
- Add a lightweight content lint check for banned jargon and overlong sentences.
- Add regression checks to ensure all skills include at least one concrete example and one action verb in revision moves.

Implemented checks:
- `web/scripts/check-skill-detail-content.mjs`
  - Validates skill coverage against `internal/domain/skills.go`.
  - Validates required fields for every skill detail object.
  - Enforces baseline readability and actionability constraints.
  - Enforces strong/weak example labeling and banned-phrase guardrails.
- Included in `scripts/test-frontend.sh` via `npm run check:skill-details`.

## 12) Sources

Primary sources used to derive this guideline:

1. Quick and Dirty Tips homepage and host showcase:
   - https://www.quickanddirtytips.com/

2. Grammar Girl show page:
   - https://www.quickanddirtytips.com/grammar-girl/

3. Grammar Girl archive example (clear sentence instruction):
   - https://www.quickanddirtytips.com/qdtarchive/how-to-write-clear-sentences/

4. Purdue OWL, Beginning Proofreading:
   - https://owl.purdue.edu/owl/general_writing/the_writing_process/proofreading/index.html

5. Purdue OWL, Proofreading for Errors:
   - https://owl.purdue.edu/owl/general_writing/the_writing_process/proofreading/proofreading_for_errors.html

6. Purdue OWL, Graduate Writing Style:
   - https://owl.purdue.edu/owl/graduate_writing/graduate_writing_topics/graduate_writing_style_new.html

7. Purdue OWL, Tone in Business Writing:
   - https://owl.purdue.edu/owl/subject_specific_writing/professional_technical_writing/tone_in_business_writing.html
