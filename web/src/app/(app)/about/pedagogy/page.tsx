import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { Text, TextLink } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Pedagogy',
}

export default function HybridMethodologyPage() {
  return (
    <div className="space-y-6">
      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Eyebrow>Pedagogy</Eyebrow>
        <Heading className="mt-3">A Hybrid Objective-Driven Writing Coaching Model</Heading>
        <Text className="mt-3">Version 1.0 • April 11, 2026</Text>
        <Text className="mt-3">
          This paper describes the instructional model behind Writing Coach: a hybrid of
          objective-driven training structure and evidence-based writing instruction research.
          We use an objective ladder inspired by TLO/ELO-style decomposition, then apply modern
          writing pedagogy for feedback, revision, and progression.
        </Text>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>1. Instructional premise</Subheading>
        <Text className="mt-3">
          Objective decomposition is used to make learning targets explicit, teachable, and
          testable. In practical terms, this means larger writing outcomes are broken into smaller
          skill objectives that can be practiced repeatedly and evaluated with clear criteria.
        </Text>
        <Text className="mt-3">
          In this application, each assignment activates exactly three skill objectives at a time.
          That bounded scope is intentional: it reduces cognitive overload and keeps revision
          decisions anchored to a short list of current goals.
        </Text>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>2. Research base for writing instruction</Subheading>
        <Text className="mt-3">
          The model is aligned with writing-instruction findings that show positive effects for
          strategy instruction, process writing, feedback, and peer/structured support in middle
          and secondary grades. Formative assessment evidence also supports iterative
          check-and-revise cycles as a driver of writing growth.
        </Text>
        <Text className="mt-3">
          We interpret those findings operationally as: focused objectives, frequent feedback,
          explicit revision targets, and longitudinal progress tracking. This is also compatible
          with deliberate-practice principles for advanced writing development.
        </Text>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>3. Hybrid model specification</Subheading>
        <ul className="mt-3 list-disc space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>
            Objective ladder: A staged skill graph defines prerequisite relationships and unlock
            order.
          </li>
          <li>
            Active objective cap: Exactly three active objectives per assignment to preserve
            signal quality and coachability.
          </li>
          <li>
            Evidence-first review: Deterministic analyzers run before model-generated language.
          </li>
          <li>
            Non-authoritative model scoring: Provider-derived scores are tagged for provenance and
            constrained against deterministic evidence.
          </li>
          <li>
            Progression guardrail: Previously strong skills are still sampled to detect backsliding;
            slipping fundamentals can pause advancement pressure.
          </li>
          <li>
            History continuity: Prompt, submission, review, and revision artifacts remain linked
            for longitudinal coaching.
          </li>
        </ul>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>4. Implementation in Writing Coach</Subheading>
        <Text className="mt-3">
          The current implementation maps this model into an analyzer-first pipeline, domain-aware
          scoring rubrics, and assignment-chain timelines. Ownership arbitration across analyzer
          families limits duplicate or conflicting deterministic findings. AI is used after
          evidence assembly to generate clearer assignment language and feedback phrasing.
        </Text>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>5. Scope and limitations</Subheading>
        <ul className="mt-3 list-disc space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>
            This is an applied pedagogical model, not a claim of formal adoption by any military
            training institution.
          </li>
          <li>
            Objective decomposition methods (for example TLO/ELO frameworks) are used here as
            instructional inspiration and engineering structure.
          </li>
          <li>
            The model currently prioritizes English deterministic analysis quality and uses
            fail-closed behavior for unsupported deterministic language contexts.
          </li>
        </ul>
      </WorkspaceCard>

      <WorkspaceCard className="border-stone-200/80 bg-white dark:border-white/10 dark:bg-zinc-900">
        <Subheading>References</Subheading>
        <ul className="mt-3 list-disc space-y-2 pl-5 text-sm/6 text-zinc-600 dark:text-zinc-300">
          <li>
            Graham, S., & Perin, D. (2007). A meta-analysis of writing instruction for adolescent
            students. <em>Journal of Educational Psychology</em>, 99(3), 445-476.{' '}
            <TextLink href="https://doi.org/10.1037/0022-0663.99.3.445">doi.org/10.1037/0022-0663.99.3.445</TextLink>
          </li>
          <li>
            Graham, S., Hebert, M., & Harris, K. R. (2015). Formative assessment and writing: A
            meta-analysis. <em>The Elementary School Journal</em>, 115(4), 523-547.{' '}
            <TextLink href="https://doi.org/10.1086/681947">doi.org/10.1086/681947</TextLink>
          </li>
          <li>
            Graham, S., et al. (2024). Effective writing instruction for students in grades 6 to
            12: a best evidence meta-analysis. <em>Reading and Writing</em>.{' '}
            <TextLink href="https://doi.org/10.1007/s11145-024-10539-2">doi.org/10.1007/s11145-024-10539-2</TextLink>
          </li>
          <li>
            Whiteford, A. P., & Rusciolelli, C. V. (2009). Training advanced writing skills: The
            case for deliberate practice. <em>Educational Psychologist</em>, 44(4), 250-266.{' '}
            <TextLink href="https://doi.org/10.1080/00461520903213600">doi.org/10.1080/00461520903213600</TextLink>
          </li>
          <li>
            Lehrer, R., et al. (2021). Using a learning progression to coordinate instruction and
            assessment. <em>Frontiers in Education</em>, 6.{' '}
            <TextLink href="https://doi.org/10.3389/feduc.2021.654212">
              doi.org/10.3389/feduc.2021.654212
            </TextLink>
          </li>
          <li>
            ASEE (2024). A hybrid pedagogy through Topical Guide Objective to enhance student
            learning in MIPS instruction set design.{' '}
            <TextLink href="https://doi.org/10.18260/1-2--46448">doi.org/10.18260/1-2--46448</TextLink>
          </li>
        </ul>
      </WorkspaceCard>
    </div>
  )
}
