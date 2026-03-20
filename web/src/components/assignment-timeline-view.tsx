'use client'

import { useEffect, useState } from 'react'
import { ArrowPathIcon } from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { PageHeader } from '@/components/page-header'
import { Strong, Text } from '@/components/text'
import { getAssignmentTimeline, getSession } from '@/lib/api'
import { requiredSetupPath } from '@/lib/onboarding-funnel'
import type { AssignmentTimeline, AssignmentTimelineStep } from '@/lib/types'
import { useRouter } from 'next/navigation'
import { ProviderProvenance } from './provider-provenance'
import { SkillScoreMeter } from './skill-score-meter'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

function formatTimelineTimestamp(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

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
    <div className="overflow-x-auto pb-2">
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
              <div className="mt-1 text-xs opacity-70">{formatTimelineTimestamp(step.created_at)}</div>
            </button>
            {index < steps.length - 1 ? <div className="h-px w-8 bg-stone-300 dark:bg-white/15" /> : null}
          </div>
        ))}
      </div>
    </div>
  )
}

function ExerciseStepSection({ step }: { step: AssignmentTimelineStep }) {
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
          <Strong>Constraints</Strong>
          <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
            {step.exercise.constraints.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </div>
        <div>
          <Strong>Success criteria</Strong>
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

function SubmissionStepSection({ step }: { step: AssignmentTimelineStep }) {
  if (!step.submission) {
    return null
  }
  return (
    <WorkspaceCard>
      <CardHeader
        eyebrow={step.label}
        title={`${step.submission.word_count} words`}
        description={`Saved ${formatTimelineTimestamp(step.created_at)}.`}
      />
      <div className="mt-5 rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
        <pre className="whitespace-pre-wrap break-words font-sans text-sm leading-7 text-zinc-700 dark:text-zinc-200">{step.submission.content}</pre>
      </div>
    </WorkspaceCard>
  )
}

function ReviewStepSection({ step }: { step: AssignmentTimelineStep }) {
  if (!step.review) {
    return null
  }
  return (
    <WorkspaceCard>
      <CardHeader
        eyebrow={step.label}
        title={step.review.summary}
        description={`Review completed ${formatTimelineTimestamp(step.created_at)}.`}
        actions={
          <div className="flex gap-2">
            <Button href={`/reviews/${step.review.id}`} outline>
              Open review
            </Button>
            {step.review.artifacts?.comparison && step.submission_id ? (
              <Button href={`/compare/${step.submission_id}`} plain>
                <ArrowPathIcon />
                Compare drafts
              </Button>
            ) : null}
          </div>
        }
      />
      <div className="mt-5 grid gap-6 xl:grid-cols-2">
        <div className="space-y-4">
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">Provider details</div>
            <div className="mt-3">
              <ProviderProvenance providerNote={step.review.provider_note} />
            </div>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">Strengths</div>
            <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
              {step.review.strengths.map((item) => (
                <li key={item}>• {item}</li>
              ))}
            </ul>
          </div>
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">Weaknesses</div>
            <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
              {step.review.weaknesses.map((item) => (
                <li key={item}>• {item}</li>
              ))}
            </ul>
          </div>
        </div>
        <div className="space-y-4">
          <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
            <div className="text-sm font-semibold text-zinc-950 dark:text-white">Next focus</div>
            <Text className="mt-3">{step.review.next_focus}</Text>
          </div>
          {step.review.skill_scores.length ? (
            <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">Skill scores</div>
              <div className="mt-3 space-y-3">
                {step.review.skill_scores.map((item) => (
                  <SkillScoreMeter key={item.skill} score={item} compact />
                ))}
              </div>
            </div>
          ) : null}
          {step.review.artifacts?.comparison ? (
            <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">Revision trajectory</div>
              <Text className="mt-3">{step.review.artifacts.comparison.summary}</Text>
            </div>
          ) : null}
        </div>
      </div>
    </WorkspaceCard>
  )
}

function StepSection({ step, selected }: { step: AssignmentTimelineStep; selected: boolean }) {
  return (
    <section id={step.id} className="scroll-mt-28">
      <div className="mb-3 flex items-center gap-3">
        <Badge color={railTone(step)}>{step.label}</Badge>
        <Text>{formatTimelineTimestamp(step.created_at)}</Text>
      </div>
      <div className={selected ? 'rounded-[2rem] bg-stone-100/70 p-3 dark:bg-white/5' : ''}>
        {step.kind === 'exercise' ? <ExerciseStepSection step={step} /> : null}
        {step.kind === 'submission' ? <SubmissionStepSection step={step} /> : null}
        {step.kind === 'review' ? <ReviewStepSection step={step} /> : null}
      </div>
    </section>
  )
}

export function AssignmentTimelineView({ exerciseId }: { exerciseId: number }) {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [assignment, setAssignment] = useState<AssignmentTimeline | null>(null)
  const [selectedStepID, setSelectedStepID] = useState('')

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const session = await getSession()
        if (!session.authenticated) {
          router.replace('/about')
          return
        }
        const nextPath = requiredSetupPath(session, `/assignments/${exerciseId}`)
        if (nextPath) {
          router.replace(nextPath)
          return
        }
        const data = await getAssignmentTimeline(exerciseId)
        if (cancelled) {
          return
        }
        setAssignment(data)
        setSelectedStepID(data.latest_step_id ?? data.steps[0]?.id ?? '')
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load assignment timeline')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [exerciseId, router])

  function handleSelect(stepID: string) {
    setSelectedStepID(stepID)
    document.getElementById(stepID)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  if (loading) {
    return <LoadingState label="Loading assignment timeline…" />
  }
  if (error || !assignment) {
    return <AppErrorState title="Assignment unavailable" error={error ?? 'Could not load the requested assignment.'} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Assignment timeline"
        title={assignment.title}
        intro="Review the full history of this assignment in one place."
        actions={
          <>
            <Button href="/assignments" plain>
              Past assignments
            </Button>
            <Button href="/" outline>
              Current assignment
            </Button>
          </>
        }
      />

      <WorkspaceCard>
        <CardHeader
          eyebrow="Navigation"
          title="Assignment flow"
          description="Jump to any step in this assignment."
        />
        <div className="mt-5">
          <TimelineRail steps={assignment.steps} selectedStepID={selectedStepID} onSelect={handleSelect} />
        </div>
      </WorkspaceCard>

      <div className="space-y-8">
        {assignment.steps.map((step) => (
          <StepSection key={step.id} step={step} selected={selectedStepID === step.id} />
        ))}
      </div>
    </div>
  )
}
