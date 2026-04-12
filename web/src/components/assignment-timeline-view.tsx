'use client'

import { ArrowPathIcon } from '@heroicons/react/16/solid'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { Callout } from '@/components/callout'
import { CardHeader } from '@/components/card-header'
import { PageHeader } from '@/components/page-header'
import { Strong, Text } from '@/components/text'
import { formatLocalDateTime } from '@/lib/datetime'
import { useAssignmentTimeline } from '@/lib/use-assignment-timeline'
import type { AssignmentTimelineStep } from '@/lib/types'
import { ProviderProvenance } from './provider-provenance'
import { AppErrorState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

function railTone(step: AssignmentTimelineStep) {
  if (step.kind === 'review') {
    return 'green'
  }
  if (step.kind === 'submission') {
    return 'amber'
  }
  return 'zinc'
}

function TimelineRail({
  steps,
  selectedStepID,
  onSelect,
}: {
  steps: AssignmentTimelineStep[]
  selectedStepID: string
  onSelect: (stepID: string) => void
}) {
  return (
    <div className="timeline-scrollbar overflow-x-auto pb-2">
      <div className="flex min-w-max items-center gap-3">
        {steps.map((step, index) => (
          <div key={step.id} className="flex items-center gap-3">
            <button
              type="button"
              onClick={() => onSelect(step.id)}
              className={`rounded-2xl border px-4 py-3 text-left transition ${
                selectedStepID === step.id
                  ? 'border-stone-900 bg-stone-900 text-white dark:border-white dark:bg-white dark:text-zinc-950'
                  : 'border-stone-200 bg-stone-50 text-zinc-900 hover:border-stone-400 dark:border-white/10 dark:bg-white/5 dark:text-white dark:hover:border-white/30'
              }`}
            >
              <div className="text-xs font-semibold uppercase tracking-[0.16em] opacity-70">{step.label}</div>
              <div className="mt-1 text-sm font-semibold">{step.title}</div>
              <div className="mt-1 text-xs opacity-70">{formatLocalDateTime(step.created_at) ?? step.created_at}</div>
            </button>
            {index < steps.length - 1 ? <div className="h-px w-8 bg-stone-300 dark:bg-white/15" /> : null}
          </div>
        ))}
      </div>
    </div>
  )
}

function ExerciseStepSection({ step, t }: { step: AssignmentTimelineStep; t: ReturnType<typeof useTranslations<'assignmentTimelineView'>> }) {
  if (!step.exercise) {
    return null
  }
  return (
    <WorkspaceCard>
      <CardHeader
        eyebrow={step.label}
        title={step.exercise.title}
        description={step.exercise.brief}
      />
      <div className="mt-5 grid gap-6 xl:grid-cols-2">
        <div>
          <ProviderProvenance providerNote={step.exercise.provider_note} />
        </div>
        <div>
          <Strong>{t('constraints')}</Strong>
          <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
            {step.exercise.constraints.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </div>
        <div>
          <Strong>{t('successCriteria')}</Strong>
          <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
            {step.exercise.success_criteria.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </div>
      </div>
    </WorkspaceCard>
  )
}

function SubmissionStepSection({
  step,
  t,
  onTryInPlayground,
}: {
  step: AssignmentTimelineStep
  t: ReturnType<typeof useTranslations<'assignmentTimelineView'>>
  onTryInPlayground: (content: string, stepId: string) => void
}) {
  if (!step.submission) {
    return null
  }
  return (
    <WorkspaceCard>
      <CardHeader
        eyebrow={step.label}
        title={t('wordCountTitle', { count: step.submission.word_count })}
        description={t('savedAt', { datetime: formatLocalDateTime(step.created_at) ?? step.created_at })}
        actions={
          <Button outline onClick={() => onTryInPlayground(step.submission?.content ?? '', step.id)}>
            {t('tryInPlayground')}
          </Button>
        }
      />
      <div className="mt-5 rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
        <pre className="whitespace-pre-wrap break-words font-sans text-sm leading-7 text-zinc-700 dark:text-zinc-200">{step.submission.content}</pre>
      </div>
    </WorkspaceCard>
  )
}

function ReviewStepSection({ step, t }: { step: AssignmentTimelineStep; t: ReturnType<typeof useTranslations<'assignmentTimelineView'>> }) {
  if (!step.review) {
    return null
  }
  return (
    <WorkspaceCard>
      <CardHeader
        eyebrow={step.label}
        title={step.review.summary}
        description={t('feedbackCompletedAt', { datetime: formatLocalDateTime(step.created_at) ?? step.created_at })}
        actions={
          <div className="flex gap-2">
            <Button href={`/reviews/${step.review.id}`} outline>
              {t('openFeedback')}
            </Button>
            {step.review.artifacts?.comparison && step.submission_id ? (
              <Button href={`/compare/${step.submission_id}`} plain>
                <ArrowPathIcon />
                {t('compareDrafts')}
              </Button>
            ) : null}
          </div>
        }
      />
      <div className="mt-5 grid gap-6 xl:grid-cols-2">
        <div className="space-y-4">
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('howCreated')}</div>
            <div className="mt-3">
              <ProviderProvenance providerNote={step.review.provider_note} kind="feedback" />
            </div>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('whatWorked')}</div>
            <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
              {step.review.strengths.map((item) => (
                <li key={item}>• {item}</li>
              ))}
            </ul>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('whatToImprove')}</div>
            <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
              {step.review.weaknesses.map((item) => (
                <li key={item}>• {item}</li>
              ))}
            </ul>
          </div>
        </div>
        <div className="space-y-4">
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('nextFocus')}</div>
            <Text className="mt-3">{step.review.next_focus}</Text>
          </div>
          {step.review.artifacts?.comparison ? (
            <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('whatChangedAcrossDrafts')}</div>
              <Text className="mt-3">{step.review.artifacts.comparison.summary}</Text>
            </div>
          ) : null}
        </div>
      </div>
    </WorkspaceCard>
  )
}

function StepSection({
  step,
  selected,
  t,
  onTryInPlayground,
}: {
  step: AssignmentTimelineStep
  selected: boolean
  t: ReturnType<typeof useTranslations<'assignmentTimelineView'>>
  onTryInPlayground: (content: string, stepId: string) => void
}) {
  return (
    <section id={step.id} className="scroll-mt-28">
      <div className="mb-3 flex items-center gap-3">
        <Badge color={railTone(step)}>{step.label}</Badge>
        <Text>{formatLocalDateTime(step.created_at) ?? step.created_at}</Text>
      </div>
      <div className={selected ? 'rounded-[2rem] bg-stone-100/70 p-3 dark:bg-white/5' : ''}>
        {step.kind === 'exercise' ? <ExerciseStepSection step={step} t={t} /> : null}
        {step.kind === 'submission' ? <SubmissionStepSection step={step} t={t} onTryInPlayground={onTryInPlayground} /> : null}
        {step.kind === 'review' ? <ReviewStepSection step={step} t={t} /> : null}
      </div>
    </section>
  )
}

export function AssignmentTimelineView({
  exerciseId,
  showCompletionState = false,
}: {
  exerciseId: number
  showCompletionState?: boolean
}) {
  const t = useTranslations('assignmentTimelineView')
  const router = useRouter()
  const { sessionLoading, sessionError, loading, error, assignment, selectedStepID, selectStep } =
    useAssignmentTimeline(exerciseId)

  function handleContinueInPlayground() {
    if (!assignment) {
      router.push('/playground')
      return
    }
    const latestSubmission = [...assignment.steps]
      .reverse()
      .find((step) => step.kind === 'submission' && step.submission?.content)?.submission
    if (latestSubmission?.content) {
      window.sessionStorage.setItem('playground-seed-content', latestSubmission.content)
      window.sessionStorage.setItem('playground-seed-source', `assignment:${assignment.root_exercise_id}`)
    }
    router.push('/playground?seed=assignment')
  }

  function handleTrySubmissionInPlayground(content: string, stepId: string) {
    if (content.trim().length === 0) {
      router.push('/playground')
      return
    }
    window.sessionStorage.setItem('playground-seed-content', content)
    window.sessionStorage.setItem('playground-seed-source', `assignment-step:${stepId}`)
    router.push('/playground?seed=assignment')
  }

  if (sessionLoading || loading) {
    return <LoadingState label={t('loading')} />
  }
  if (sessionError || error || !assignment) {
    return <AppErrorState title={t('unavailableTitle')} error={sessionError ?? error ?? t('unavailableBody')} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={t('eyebrow')}
        title={assignment.title}
        intro={t('intro')}
        actions={
          <>
            <Button href="/assignments" plain>
              {t('pastAssignments')}
            </Button>
            <Button href="/" outline>
              {t('currentAssignment')}
            </Button>
          </>
        }
      />

      {showCompletionState ? (
        <Callout
          tone="success"
          eyebrow={t('completeEyebrow')}
          title={t('completeTitle')}
          body={t('completeBody')}
          actions={
            <div className="flex flex-wrap gap-2">
              <Button href="/new-assignment" color="dark/zinc">
                {t('startNextAssignment')}
              </Button>
              <Button outline onClick={handleContinueInPlayground}>
                {t('continueInPlayground')}
              </Button>
            </div>
          }
        />
      ) : null}

      <WorkspaceCard>
        <CardHeader
          eyebrow={t('navigationEyebrow')}
          title={t('navigationTitle')}
          description={t('navigationDescription')}
        />
        <div className="mt-5">
          <TimelineRail steps={assignment.steps} selectedStepID={selectedStepID} onSelect={selectStep} />
        </div>
      </WorkspaceCard>

      <div className="space-y-8">
        {assignment.steps.map((step) => (
          <StepSection
            key={step.id}
            step={step}
            selected={selectedStepID === step.id}
            t={t}
            onTryInPlayground={handleTrySubmissionInPlayground}
          />
        ))}
      </div>
    </div>
  )
}
