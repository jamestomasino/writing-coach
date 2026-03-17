'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { ArrowPathIcon, ArrowUpTrayIcon, SparklesIcon, ExclamationTriangleIcon } from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { Strong, Text } from '@/components/text'
import { Textarea } from '@/components/textarea'
import { createRevisionAssignment, getDashboard, getExercise, getExercises, getReviewJob, getReviews, getSession, getSubmissions, reviewSubmission, submitDraft } from '@/lib/api'
import type { Dashboard, Exercise, Review, ReviewJob, Submission } from '@/lib/types'
import { EmptyState, LoadingState, TaskProgressState } from './status-state'
import { WorkspaceCard } from './workspace-card'

type WorkspaceState = {
  dashboard: Dashboard
  exercise?: Exercise
  submission?: Submission
  review?: Review
}

function countWords(value: string) {
  return value.trim() === '' ? 0 : value.trim().split(/\s+/).length
}

function renderMasteryState(tgo: Dashboard['active_tgos'][number]) {
  if (tgo.progress_mode === 'percent') {
    const percent = tgo.mastery_percent ?? 0
    return (
      <div className="mt-3">
        <div className="flex items-center justify-between gap-4 text-xs font-medium uppercase tracking-[0.16em] text-zinc-500 dark:text-zinc-400">
          <span>Mastery progress</span>
          <span>{percent}%</span>
        </div>
        <div className="mt-2 h-2 rounded-full bg-stone-200 dark:bg-white/10">
          <div className="h-2 rounded-full bg-stone-800 dark:bg-stone-200" style={{ width: `${percent}%` }} />
        </div>
      </div>
    )
  }
  return (
    <div className="mt-3 text-sm text-zinc-600 dark:text-zinc-300">
      <span className="font-semibold text-zinc-950 dark:text-white">Mastery state:</span>{' '}
      {tgo.mastery_stage ?? 'emerging'}
      {tgo.mastery_evidence_count ? ` across ${tgo.mastery_evidence_count} review${tgo.mastery_evidence_count === 1 ? '' : 's'}` : ''}
    </div>
  )
}

export function CurrentAssignmentView() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sessionRequired, setSessionRequired] = useState(false)
  const [workspace, setWorkspace] = useState<WorkspaceState | null>(null)
  const [draft, setDraft] = useState('')
  const [reviewing, setReviewing] = useState(false)
  const [preparingRevision, setPreparingRevision] = useState(false)
  const [reviewJob, setReviewJob] = useState<ReviewJob | null>(null)

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        setLoading(true)
        const session = await getSession()
        if (!session.authenticated) {
          if (!cancelled) {
            setSessionRequired(true)
          }
          return
        }
        if (!session.onboarding_complete) {
          router.replace('/onboarding')
          return
        }

        const dashboard = await getDashboard()
        const revisionExerciseID = Number(searchParams.get('revisionExercise') ?? 0)
        let exercise: Exercise | undefined
        if (revisionExerciseID > 0) {
          exercise = await getExercise(revisionExerciseID)
        } else {
          const exercises = await getExercises(1)
          exercise = exercises[0]
        }
        let submission: Submission | undefined
        let review: Review | undefined
        let pendingJob: ReviewJob | null = null
        if (exercise) {
          const submissions = await getSubmissions(exercise.id, 1)
          submission = submissions[0]
          if (submission) {
            const reviews = await getReviews(submission.id, 1)
            review = reviews[0]
            if (!review) {
              try {
                pendingJob = await getReviewJob(submission.id)
              } catch {}
            }
          }
        }

        if (!cancelled) {
          setWorkspace({ dashboard, exercise, submission, review })
          setDraft(submission?.content ?? '')
          setReviewJob(pendingJob)
          setSessionRequired(false)
          setError(null)
        }
      } catch (err) {
        if (!cancelled) {
          const message = err instanceof Error ? err.message : 'Failed to load current assignment'
          if (message.toLowerCase().includes('unauthorized')) {
            setSessionRequired(true)
          } else {
            setError(message)
          }
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
  }, [router, searchParams])

  useEffect(() => {
    if (!reviewJob || !workspace?.submission) {
      return
    }
    if (reviewJob.status !== 'queued' && reviewJob.status !== 'running') {
      return
    }

    let cancelled = false
    const timer = window.setInterval(async () => {
      try {
        const job = await getReviewJob(workspace.submission!.id)
        if (cancelled) {
          return
        }
        setReviewJob(job)
        if (job.status === 'completed' && job.review_id) {
          router.push(`/reviews/${job.review_id}`)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not refresh review status')
        }
      }
    }, 2000)

    return () => {
      cancelled = true
      window.clearInterval(timer)
    }
  }, [reviewJob, router, workspace?.submission])

  const wordCount = useMemo(() => countWords(draft), [draft])

  async function handleFile(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (!file) {
      return
    }
    setDraft(await file.text())
  }

  async function handleReview() {
    if (!workspace?.exercise || draft.trim() === '') {
      return
    }
    try {
      setReviewing(true)
      setError(null)
      const submission = await submitDraft({
        exerciseId: workspace.exercise.id,
        content: draft,
        parentSubmissionId: workspace.submission?.id,
      })
      const job = await reviewSubmission(submission.id)
      setWorkspace({ ...workspace, submission, review: undefined })
      setReviewJob(job)
      if (job.status === 'completed' && job.review_id) {
        router.push(`/reviews/${job.review_id}`)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Review failed')
    } finally {
      setReviewing(false)
    }
  }

  async function handleRetryReview() {
    if (!workspace?.submission) {
      return
    }
    try {
      setReviewing(true)
      setError(null)
      const job = await reviewSubmission(workspace.submission.id)
      setReviewJob(job)
      if (job.status === 'completed' && job.review_id) {
        router.push(`/reviews/${job.review_id}`)
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not retry review')
    } finally {
      setReviewing(false)
    }
  }

  async function handleRevisionPrompt() {
    if (!workspace?.submission) {
      return
    }
    try {
      setPreparingRevision(true)
      setError(null)
      const exercise = await createRevisionAssignment(workspace.submission.id)
      router.push(`/?revisionExercise=${exercise.id}`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create revision prompt')
    } finally {
      setPreparingRevision(false)
    }
  }

  if (loading) {
    return <LoadingState />
  }
  if (sessionRequired) {
    return (
      <EmptyState
        title="Sign in to continue"
        body="Authentication runs through the Kratos browser flow. Once you sign in, this page becomes your current assignment workspace."
        actionHref="/login"
        actionLabel="Open sign in"
      />
    )
  }
  if (error) {
    return <EmptyState title="Workspace unavailable" body={error} actionHref="/" actionLabel="Try again" />
  }
  if (!workspace) {
    return <LoadingState />
  }

  const { dashboard, exercise, review } = workspace
  const isRevisionBrief = searchParams.get('revisionExercise') !== null
  const reviewPending = reviewJob?.status === 'queued' || reviewJob?.status === 'running'
  const reviewFailed = reviewJob?.status === 'failed'
  const busy = reviewing || preparingRevision || reviewPending

  if (!exercise) {
    return (
      <div className="space-y-8">
        <header>
          <Eyebrow>Assignment workspace</Eyebrow>
          <Heading>Current assignment</Heading>
          <Text className="mt-2">You do not have an active assignment yet. Choose 3 unlocked skills and generate the next prompt.</Text>
        </header>
        <WorkspaceCard>
          <Eyebrow>Current focus</Eyebrow>
          <Subheading>Active skills</Subheading>
          <div className="mt-4 flex flex-wrap gap-2">
            {dashboard.active_tgos.map((tgo) => (
              <Badge key={tgo.code} color="blue">
                {tgo.title}
              </Badge>
            ))}
          </div>
          <div className="mt-6">
            <Button href="/new-assignment" color="dark/zinc">
              Start new assignment
            </Button>
          </div>
        </WorkspaceCard>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <Eyebrow>Assignment workspace</Eyebrow>
          <Heading>{exercise.title}</Heading>
          <Text className="mt-2 max-w-3xl">{exercise.brief}</Text>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button href="/new-assignment" outline>
            New assignment
          </Button>
          {review ? (
            <Button onClick={handleRevisionPrompt} color="dark/zinc" disabled={busy}>
              {preparingRevision ? 'Preparing revision brief…' : 'Revise from latest review'}
            </Button>
          ) : null}
        </div>
      </header>

      {reviewPending ? (
        <TaskProgressState
          title="Review in progress"
          body="Your draft is saved. The app is reviewing it in the background against the active skills and will open the coaching pass when it is ready."
          steps={[
            'Save the latest draft snapshot.',
            'Score the draft against the active skill rubric.',
            'Assemble the coaching summary, annotations, and next focus.',
          ]}
        />
      ) : null}

      {preparingRevision ? (
        <TaskProgressState
          title="Revision brief in progress"
          body="The app is building a revision brief from your latest review and active skills."
          steps={[
            'Load the latest reviewed draft.',
            'Extract the highest-priority revision targets.',
            'Generate a focused revision brief for the next pass.',
          ]}
        />
      ) : null}

      {reviewFailed ? (
        <WorkspaceCard>
          <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <Eyebrow>Review status</Eyebrow>
              <div className="flex items-center gap-2 text-sm font-semibold text-amber-900 dark:text-amber-200">
                <ExclamationTriangleIcon className="size-4" />
                Review failed
              </div>
              <Text className="mt-2">
                The draft saved, but the background review did not finish. You can retry without losing your submission.
              </Text>
              {reviewJob?.last_error ? <Text className="mt-2 text-sm">{reviewJob.last_error}</Text> : null}
            </div>
            <Button onClick={handleRetryReview} color="dark/zinc" disabled={reviewing}>
              {reviewing ? 'Retrying…' : 'Retry review'}
            </Button>
          </div>
        </WorkspaceCard>
      ) : null}

      {isRevisionBrief ? (
        <WorkspaceCard>
          <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <Eyebrow>Revision mode</Eyebrow>
              <Subheading>Revision brief</Subheading>
              <Text className="mt-2">
                This assignment was generated from the latest review. Keep the same core material, but revise explicitly against the active skills and the coaching notes below.
              </Text>
            </div>
            <Button href={`/compare/${workspace.submission?.id ?? 0}`} outline>
              <ArrowPathIcon />
              Open revision compare
            </Button>
          </div>
        </WorkspaceCard>
      ) : null}

      <div className="grid gap-8 xl:grid-cols-[2fr_1fr]">
        <WorkspaceCard>
          <div className="flex items-center justify-between gap-4">
            <div>
              <Eyebrow>Assignment brief</Eyebrow>
              <Subheading>Prompt</Subheading>
            </div>
            <Badge color="zinc">{exercise.generation_kind}</Badge>
          </div>
          <div className="mt-4 space-y-5">
            <div>
              <Strong>Constraints</Strong>
              <ul className="mt-2 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                {exercise.constraints.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
            <div>
              <Strong>Success criteria</Strong>
              <ul className="mt-2 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                {exercise.success_criteria.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
          </div>
        </WorkspaceCard>

        <WorkspaceCard>
          <Eyebrow>Current focus</Eyebrow>
          <Subheading>Active skills</Subheading>
          <Text className="mt-2">These are the three skills the review will measure most heavily on this assignment.</Text>
          <div className="mt-4 space-y-3">
            {dashboard.active_tgos.map((tgo) => (
              <div key={tgo.code} className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                <div className="flex items-center justify-between gap-4">
                  <Strong>{tgo.title}</Strong>
                  <Badge color="cyan">{tgo.stage}</Badge>
                </div>
                <Text className="mt-2">{tgo.description}</Text>
                {renderMasteryState(tgo)}
              </div>
            ))}
          </div>
        </WorkspaceCard>
      </div>

      <WorkspaceCard>
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <Eyebrow>Write and submit</Eyebrow>
            <Subheading>Draft submission</Subheading>
            <Text className="mt-2">Paste plain text or markdown, or load a local draft file before requesting a review.</Text>
          </div>
          <div className="flex items-center gap-3">
            <label className="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-stone-300 px-3 py-2 text-sm font-medium text-zinc-700 hover:bg-stone-50 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5">
              <ArrowUpTrayIcon className="size-4" />
              Upload draft
              <input type="file" accept=".md,.txt,text/plain,text/markdown" className="hidden" onChange={handleFile} disabled={busy} />
            </label>
            <Badge color="zinc">{wordCount} words</Badge>
          </div>
        </div>
        <div className="mt-5">
          <Textarea value={draft} onChange={(event) => setDraft(event.target.value)} rows={18} placeholder="Paste your draft here." disabled={busy} />
        </div>
          <div className="mt-5 flex flex-wrap items-center justify-between gap-3">
          <Text>
            {workspace.submission
              ? reviewPending
                ? `Latest saved draft: #${workspace.submission.draft_number}. Review queued ${reviewJob?.updated_at ?? 'just now'}.`
                : `Latest saved draft: #${workspace.submission.draft_number}`
              : 'No draft submitted yet for this assignment.'}
          </Text>
          <Button onClick={handleReview} color="dark/zinc" disabled={busy || draft.trim() === ''}>
            <SparklesIcon />
            {reviewing || reviewPending ? 'Review queued…' : 'Submit for review'}
          </Button>
        </div>
      </WorkspaceCard>

      {review ? (
        <WorkspaceCard>
          <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <Eyebrow>Latest feedback</Eyebrow>
              <Subheading>Latest coaching pass</Subheading>
              <Text className="mt-2">{review.summary}</Text>
            </div>
            <div className="flex gap-2">
              <Button href={`/reviews/${review.id}`} outline>
                Open review
              </Button>
              <Button href={`/compare/${workspace.submission?.id ?? 0}`} plain>
                Compare drafts
              </Button>
            </div>
          </div>
          {review.artifacts?.comparison ? (
            <div className="mt-5 grid gap-4 lg:grid-cols-3">
              <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">Revision summary</div>
                <Text className="mt-2 text-sm">{review.artifacts.comparison.summary}</Text>
              </div>
              <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">Addressed</div>
                <Text className="mt-2 text-sm">{review.artifacts.comparison.addressed_weaknesses.length}</Text>
              </div>
              <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">Persisting</div>
                <Text className="mt-2 text-sm">{review.artifacts.comparison.persisting_weaknesses.length}</Text>
              </div>
            </div>
          ) : null}
        </WorkspaceCard>
      ) : null}
    </div>
  )
}
