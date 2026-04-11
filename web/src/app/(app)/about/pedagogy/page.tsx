import { Eyebrow } from '@/components/eyebrow'
import { Heading } from '@/components/heading'
import { Text, TextLink } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Pedagogy',
}

export default function HybridMethodologyPage() {
  return (
    <div className="space-y-8">
      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Eyebrow>Pedagogy</Eyebrow>
        <Heading className="mt-3">A Hybrid Objective-Driven Writing Coaching Model</Heading>
        <Text className="mt-3">Version 1.0 • April 11, 2026</Text>

        <Text className="mt-6">
          Writing Coach uses a hybrid pedagogical model: objective decomposition for structure and
          evidence-based writing research for learning efficacy. In this paper, we use
          <strong> Terminal Learning Objective (TLO)</strong> and
          <strong> Enabling Learning Objective (ELO)</strong> in their standard instructional
          design sense, where the TLO defines the end performance target and ELOs define the
          supporting sub-objectives required to reach it [6,7].
        </Text>
        <Text className="mt-4">
          The model applies that objective structure to writing development by constraining each
          assignment to three active skill objectives, because a bounded target set keeps feedback
          legible and revision decisions actionable while still leaving enough surface area for
          meaningful improvement in a single draft cycle. That choice is an implementation design
          constraint, not a universal law of writing instruction, and it is used here to keep the
          coaching loop operationally stable as objectives advance through prerequisite gates.
        </Text>
        <Text className="mt-4">
          The instructional rationale is supported by writing research showing reliable gains from
          explicit strategy instruction, process-oriented writing approaches, and structured
          feedback loops [1,3], while formative assessment findings support iterative
          criterion-referenced revision cycles in which learners can inspect evidence, revise, and
          recheck performance over time [2]. This is also consistent with deliberate-practice
          principles, where improvement depends on repeated attempts against a clear target plus
          immediate, specific feedback that can be acted on in the next pass [4].
        </Text>
        <Text className="mt-4">
          Learning progression research also supports coordinating instruction and assessment across
          stages of development instead of treating assignments as isolated tasks [5]. In this
          system, that principle is implemented through a connected assignment chain that preserves
          state between attempts, so each new assignment is informed by prior performance patterns
          instead of being generated as an unrelated prompt.
        </Text>
        <Text className="mt-4">
          A representative workflow looks like this: a learner selects three active objectives,
          writes and submits a draft, receives deterministic evidence and rubric-aligned scoring,
          reviews a revision brief that prioritizes the highest-leverage fixes, and then either
          revises the same assignment chain or advances to the next unlocked objective set. This
          sequence makes the coaching contract explicit: each step should make the next decision
          clearer, and each decision should be auditable against preserved evidence.
        </Text>
        <Text className="mt-4">
          Evidence-to-design mapping in this implementation follows a simple rule: formative
          assessment evidence [2] is operationalized as revision briefs with explicit next-focus
          priorities; deliberate practice findings [4] are operationalized as repeated objective
          cycles against the same assignment chain; and learning progression principles [5] are
          operationalized as prerequisite-based objective unlocks instead of random prompt drift.
        </Text>
        <Text className="mt-4">
          Example: a learner working on claim clarity, evidence integration, and sentence economy
          submits a draft that shows weak support for two major claims. The revision brief then
          prioritizes adding concrete evidence to those claims before stylistic polishing, and the
          next review checks whether those specific deficits were corrected.
        </Text>
        <Text className="mt-4">
          Before/after micro-example: before revision, a paragraph states a conclusion without
          source support; after revision, the same paragraph adds concrete evidence, names the
          warrant connecting evidence to claim, and trims one distracting stylistic flourish. The
          follow-up review then scores evidence integration first and only secondarily comments on
          style.
        </Text>
        <Text className="mt-4">
          At the technical layer, the app enforces an evidence-first review sequence. Deterministic
          analyzers run before model-generated language output. Model-derived scoring is tracked as
          non-authoritative provenance when present, meaning provider output is stored with explicit
          source labels as supporting context and not as the governing scoring authority.
          Previously strengthened skills are still checked so that backsliding can pause
          advancement pressure until fundamentals are stable again. This keeps progression coupled
          to observed performance instead of one-off output quality.
        </Text>
        <Text className="mt-4">
          The scope of this model is applied writing coaching in this product context. It is not
          presented as institutional doctrine; it is a practical synthesis that combines
          objective-driven instruction design concepts [6,7] with modern writing-instruction
          evidence [1-5].
        </Text>

        <Text className="mt-8 text-sm font-semibold tracking-wide text-zinc-800 uppercase dark:text-zinc-200">
          References
        </Text>
        <ol className="mt-3 list-decimal space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>
            Graham S, Perin D. A meta-analysis of writing instruction for adolescent students.
            <em> Journal of Educational Psychology</em>. 2007;99(3):445-476.{' '}
            <TextLink href="https://doi.org/10.1037/0022-0663.99.3.445">
              https://doi.org/10.1037/0022-0663.99.3.445
            </TextLink>
          </li>
          <li>
            Graham S, Hebert M, Harris KR. Formative assessment and writing: A meta-analysis.
            <em> The Elementary School Journal</em>. 2015;115(4):523-547.{' '}
            <TextLink href="https://doi.org/10.1086/681947">https://doi.org/10.1086/681947</TextLink>
          </li>
          <li>
            Graham S, et al. Effective writing instruction for students in grades 6 to 12: a best
            evidence meta-analysis. <em>Reading and Writing</em>. 2024.{' '}
            <TextLink href="https://doi.org/10.1007/s11145-024-10539-2">
              https://doi.org/10.1007/s11145-024-10539-2
            </TextLink>
          </li>
          <li>
            Whiteford AP, Rusciolelli CV. Training Advanced Writing Skills: The Case for Deliberate
            Practice. <em>Educational Psychologist</em>. 2009;44(4):250-266.{' '}
            <TextLink href="https://doi.org/10.1080/00461520903213600">
              https://doi.org/10.1080/00461520903213600
            </TextLink>
          </li>
          <li>
            Lehrer R, et al. Improving Learning: Using a Learning Progression to Coordinate
            Instruction and Assessment. <em>Frontiers in Education</em>. 2021;6.{' '}
            <TextLink href="https://doi.org/10.3389/feduc.2021.654212">
              https://doi.org/10.3389/feduc.2021.654212
            </TextLink>
          </li>
          <li>
            Marine Corps Systems Approach to Training (SAT) Manual. TLO/ELO construction and
            subordinate objective rules.{' '}
            <TextLink href="https://www.trngcmd.marines.mil/Portals/207/Docs/FLW/EEIC/SAT_Manual.pdf">
              https://www.trngcmd.marines.mil/Portals/207/Docs/FLW/EEIC/SAT_Manual.pdf
            </TextLink>
          </li>
          <li>
            NAVMC 1553.2 Marine Corps Formal School Management Policy Guidance. Definitions and
            policy language for Terminal and Enabling Learning Objectives (TLO/ELO).{' '}
            <TextLink href="https://www.trngcmd.marines.mil/Portals/207/Docs/FLW/EEIC/NAVMC%201553.2%20Marine%20Corps%20Formal%20School%20Management%20Policy%20Guidance.pdf">
              trngcmd.marines.mil/.../NAVMC%201553.2...Policy%20Guidance.pdf
            </TextLink>
          </li>
        </ol>
      </WorkspaceCard>
    </div>
  )
}
