'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { ArrowPathIcon, ArrowUpTrayIcon, SparklesIcon, ExclamationTriangleIcon } from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Callout } from '@/components/callout'
import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { PageHeader } from '@/components/page-header'
import { Strong, Text } from '@/components/text'
import { Textarea } from '@/components/textarea'
import { createRevisionAssignment, getAISettings, getDashboard, getExercise, getExercises, getReviewJob, getReviews, getSession, getSubmission, getSubmissions, reviewSubmission, submitDraft } from '@/lib/api'
import type { AIProviderSettings, Dashboard, Exercise, Review, ReviewJob, Submission } from '@/lib/types'
import { MasteryProgress } from './mastery-progress'
import { AppErrorState, EmptyState, LoadingState, TaskProgressState } from './status-state'
import { WorkspaceCard } from './workspace-card'

type WorkspaceState = {
  dashboard: Dashboard
  aiSettings?: AIProviderSettings
  exercise?: Exercise
  submission?: Submission
  review?: Review
  sourceSubmission?: Submission
  sourceReview?: Review
}

function countWords(value: string) {
  return value.trim() === '' ? 0 : value.trim().split(/\s+/).length
}

function isTransientError(err: unknown) {
  if (!(err instanceof Error)) {
    return false
  }
  const message = err.message.toLowerCase()
  return (
    message.includes('failed to fetch') ||
    message.includes('networkerror') ||
    message.includes('network request failed') ||
    message.includes('timeout') ||
    message.includes('500') ||
    message.includes('502') ||
    message.includes('503') ||
    message.includes('504')
  )
}

async function withRetry<T>(load: () => Promise<T>, attempts = 3, delayMs = 250): Promise<T> {
  let lastError: unknown
  for (let attempt = 1; attempt <= attempts; attempt += 1) {
    try {
      return await load()
    } catch (err) {
      lastError = err
      if (attempt === attempts || !isTransientError(err)) {
        throw err
      }
      await new Promise((resolve) => window.setTimeout(resolve, delayMs * attempt))
    }
  }
  throw lastError instanceof Error ? lastError : new Error('Request failed')
}

export function CurrentAssignmentView() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sessionRequired, setSessionRequired] = useState(false)
  const [needsAISetup, setNeedsAISetup] = useState(false)
  const [workspace, setWorkspace] = useState<WorkspaceState | null>(null)
  const [draft, setDraft] = useState('')
  const [reviewing, setReviewing] = useState(false)
  const [preparingRevision, setPreparingRevision] = useState(false)
  const [reviewJob, setReviewJob] = useState<ReviewJob | null>(null)
  const [revisionPanel, setRevisionPanel] = useState<'brief' | 'feedback'>('brief')

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
        if (!session.ai_provider_ready) {
          if (!cancelled) {
            setNeedsAISetup(true)
          }
          return
        }

        const [dashboard, aiSettings] = await Promise.all([getDashboard(), getAISettings()])
        const revisionExerciseID = Number(searchParams.get('revisionExercise') ?? 0)
        const inRevisionMode = revisionExerciseID > 0
        let exercise: Exercise | undefined
        if (inRevisionMode) {
          exercise = await withRetry(() => getExercise(revisionExerciseID))
        } else {
          const exercises = await getExercises(1)
          exercise = exercises[0]
        }
        let submission: Submission | undefined
        let review: Review | undefined
        let sourceSubmission: Submission | undefined
        let sourceReview: Review | undefined
        let pendingJob: ReviewJob | null = null
        if (exercise) {
          const submissions = inRevisionMode ? await withRetry(() => getSubmissions(exercise.id, 1)) : await getSubmissions(exercise.id, 1)
          submission = submissions[0]
          if (submission) {
            const currentSubmissionID = submission.id
            const reviews = inRevisionMode ? await withRetry(() => getReviews(currentSubmissionID, 1)) : await getReviews(currentSubmissionID, 1)
            review = reviews[0]
            if (!review) {
              try {
                pendingJob = await getReviewJob(currentSubmissionID)
              } catch {}
            }
          }
          if (exercise.source_submission_id) {
            sourceSubmission = inRevisionMode ? await withRetry(() => getSubmission(exercise.source_submission_id!)) : await getSubmission(exercise.source_submission_id)
            const sourceSubmissionID = sourceSubmission.id
            const sourceReviews = inRevisionMode ? await withRetry(() => getReviews(sourceSubmissionID, 1)) : await getReviews(sourceSubmissionID, 1)
            sourceReview = sourceReviews[0]
          }
        }

        if (!cancelled) {
          setWorkspace({ dashboard, aiSettings, exercise, submission, review, sourceSubmission, sourceReview })
          setDraft(submission?.content ?? sourceSubmission?.content ?? '')
          setReviewJob(pendingJob)
          setSessionRequired(false)
          setNeedsAISetup(false)
          setError(null)
          setRevisionPanel(sourceReview ? 'feedback' : 'brief')
        }
      } catch (err) {
        if (!cancelled) {
          const message = err instanceof Error ? err.message : 'Failed to load current assignment'
          if (message.toLowerCase().includes('unauthorized')) {
            setSessionRequired(true)
          } else {
            const revisionExerciseID = Number(searchParams.get('revisionExercise') ?? 0)
            if (revisionExerciseID > 0 && isTransientError(err)) {
              setError('Could not load the revision workspace on the first attempt. Please try again. The revision brief may already be ready.')
            } else {
              setError(message)
            }
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
  if (needsAISetup) {
    return (
      <EmptyState
        title="AI setup required"
        body="Add an AI provider before generating assignments, feedback, or revision briefs."
        actionHref="/ai-settings?required=1&next=/"
        actionLabel="Set up AI provider"
      />
    )
  }
  if (error) {
    return <AppErrorState error={error} title="Workspace unavailable" />
  }
  if (!workspace) {
    return <LoadingState />
  }

  const { dashboard, exercise, review, sourceSubmission, sourceReview } = workspace
  const isRevisionBrief = searchParams.get('revisionExercise') !== null
  const reviewPending = reviewJob?.status === 'queued' || reviewJob?.status === 'running'
  const reviewFailed = reviewJob?.status === 'failed'
  const busy = reviewing || preparingRevision || reviewPending
  const compareSubmissionID = workspace.submission?.id ?? 0

  if (!exercise) {
    return (
      <div className="space-y-8">
        <PageHeader
          eyebrow="Assignment workspace"
          title="Current assignment"
          intro="You do not have an active assignment yet. Choose 3 unlocked skills and generate the next prompt."
        />
        <WorkspaceCard>
          <CardHeader eyebrow="Current focus" title="Active skills" />
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
      <PageHeader
        eyebrow="Assignment workspace"
        title={exercise.title}
        intro={exercise.brief}
        actions={
          <>
            <Button href={`/assignments/${exercise.id}`} plain>
              View timeline
            </Button>
            <Button href="/new-assignment" outline>
              New assignment
            </Button>
            {review ? (
              <Button onClick={handleRevisionPrompt} color="dark/zinc" disabled={busy}>
                {preparingRevision ? 'Preparing revision brief…' : 'Revise from latest review'}
              </Button>
            ) : null}
          </>
        }
      />

      {workspace.aiSettings?.system_fallback && !workspace.aiSettings.has_key ? (
        <Callout
          title="Using the shared system provider"
          body="You can keep working as usual, or connect your own provider key in AI provider settings."
          actions={
            <Button href="/ai-settings" outline>
              AI provider settings
            </Button>
          }
        />
      ) : null}

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
        <Callout
          tone="warning"
          eyebrow="Review status"
          title="Review failed"
          body="The draft saved, but the background review did not finish. You can retry without losing your submission."
          actions={
            <Button onClick={handleRetryReview} color="dark/zinc" disabled={reviewing}>
              {reviewing ? 'Retrying…' : 'Retry review'}
            </Button>
          }
        >
          <div className="flex items-center gap-2 text-sm font-semibold text-amber-900 dark:text-amber-200">
            <ExclamationTriangleIcon className="size-4" />
            Background review interrupted
          </div>
          {reviewJob?.last_error ? <Text className="mt-2 text-sm">{reviewJob.last_error}</Text> : null}
        </Callout>
      ) : null}

      {isRevisionBrief ? (
        <WorkspaceCard>
          <CardHeader
            eyebrow="Revision mode"
            title="Revision brief"
            description="This assignment was generated from the latest review. Keep the same core material, but revise explicitly against the active skills and the coaching notes below."
            actions={
              review ? (
                <Button href={`/compare/${compareSubmissionID}`} outline>
                  <ArrowPathIcon />
                  Open revision compare
                </Button>
              ) : sourceReview ? (
                <Button href={`/reviews/${sourceReview.id}`} outline>
                  Open prior review
                </Button>
              ) : null
            }
          />
          {sourceReview ? (
            <div className="mt-5">
              <div className="inline-flex rounded-xl border border-stone-200 bg-stone-50 p-1 dark:border-white/10 dark:bg-white/5">
                <button
                  type="button"
                  onClick={() => setRevisionPanel('brief')}
                  className={`rounded-lg px-3 py-2 text-sm font-medium transition ${revisionPanel === 'brief' ? 'bg-white text-zinc-950 shadow-sm dark:bg-white/10 dark:text-white' : 'text-zinc-600 hover:text-zinc-950 dark:text-zinc-300 dark:hover:text-white'}`}
                >
                  Revision brief
                </button>
                <button
                  type="button"
                  onClick={() => setRevisionPanel('feedback')}
                  className={`rounded-lg px-3 py-2 text-sm font-medium transition ${revisionPanel === 'feedback' ? 'bg-white text-zinc-950 shadow-sm dark:bg-white/10 dark:text-white' : 'text-zinc-600 hover:text-zinc-950 dark:text-zinc-300 dark:hover:text-white'}`}
                >
                  Prior feedback
                </button>
              </div>
              {revisionPanel === 'brief' ? (
                <div className="mt-4 rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
                  <div className="flex flex-wrap items-center gap-3">
                    <Badge color="zinc">Source draft #{sourceSubmission?.draft_number ?? 1}</Badge>
                    <Badge color="cyan">{exercise.generation_kind}</Badge>
                  </div>
                  <Text className="mt-3">
                    The editor is preloaded with your latest reviewed draft so you can revise directly instead of pasting it back in.
                  </Text>
                </div>
              ) : (
                <div className="mt-4 space-y-4">
                  <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
                    <div className="text-sm font-semibold uppercase tracking-[0.16em] text-zinc-500 dark:text-zinc-400">Prior coaching summary</div>
                    <Text className="mt-3">{sourceReview.summary}</Text>
                  </div>
                  <div className="grid gap-4 lg:grid-cols-2">
                    <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
                      <div className="text-sm font-semibold text-zinc-950 dark:text-white">Strengths to preserve</div>
                      <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                        {sourceReview.strengths.map((item) => (
                          <li key={item}>• {item}</li>
                        ))}
                      </ul>
                    </div>
                    <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
                      <div className="text-sm font-semibold text-zinc-950 dark:text-white">Revision targets</div>
                      <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                        {sourceReview.weaknesses.map((item) => (
                          <li key={item}>• {item}</li>
                        ))}
                      </ul>
                    </div>
                  </div>
                  <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
                    <div className="text-sm font-semibold text-zinc-950 dark:text-white">Next focus</div>
                    <Text className="mt-3">{sourceReview.next_focus}</Text>
                  </div>
                </div>
              )}
            </div>
          ) : null}
        </WorkspaceCard>
      ) : null}

      <div className="grid gap-8 xl:grid-cols-[2fr_1fr]">
        <WorkspaceCard>
          <CardHeader eyebrow="Assignment brief" title="Prompt" actions={<Badge color="zinc">{exercise.generation_kind}</Badge>} />
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
          <CardHeader
            eyebrow="Current focus"
            title="Active skills"
            description="These are the three skills the review will measure most heavily on this assignment."
          />
          <div className="mt-4 space-y-3">
            {dashboard.active_tgos.map((tgo) => (
              <div key={tgo.code} className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                <div className="flex items-center justify-between gap-4">
                  <Strong>{tgo.title}</Strong>
                  <Badge color="cyan">{tgo.stage}</Badge>
                </div>
                <Text className="mt-2">{tgo.description}</Text>
                <MasteryProgress tgo={tgo} />
              </div>
            ))}
          </div>
        </WorkspaceCard>
      </div>

      <WorkspaceCard>
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <CardHeader
              eyebrow="Write and submit"
              title="Draft submission"
              description="Paste plain text or markdown, or load a local draft file before requesting a review."
            />
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
          <CardHeader
            eyebrow="Latest feedback"
            title="Latest coaching pass"
            description={review.summary}
            actions={
              <div className="flex gap-2">
                <Button href={`/assignments/${exercise.id}`} plain>
                  View timeline
                </Button>
                <Button href={`/reviews/${review.id}`} outline>
                  Open review
                </Button>
                <Button href={`/compare/${compareSubmissionID}`} plain>
                  Compare drafts
                </Button>
              </div>
            }
          />
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
