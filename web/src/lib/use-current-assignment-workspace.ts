'use client'

import {
  createRevisionAssignment,
  getAssignments,
  getDashboard,
  getExercise,
  getReviewJob,
  getReviews,
  getSession,
  getSubmission,
  getSubmissions,
  reviewSubmission,
  submitDraft,
} from '@/lib/api'
import { requiredSetupPath } from '@/lib/onboarding-funnel'
import { hasUnsavedTrackDraft } from '@/lib/track-switch-guard'
import type { Dashboard, Exercise, Review, ReviewJob, Submission } from '@/lib/types'
import { useRouter } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'

declare global {
  interface Window {
    __writingCoachHasUnsavedDraft?: boolean
  }
}

type WorkspaceState = {
  dashboard: Dashboard
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

export function useCurrentAssignmentWorkspace(revisionExerciseID: number) {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [sessionRequired, setSessionRequired] = useState(false)
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
        const nextPath = requiredSetupPath(session, '/')
        if (nextPath) {
          router.replace(nextPath)
          return
        }

        const dashboard = await getDashboard()
        const inRevisionMode = revisionExerciseID > 0
        let exercise: Exercise | undefined
        if (inRevisionMode) {
          exercise = await withRetry(() => getExercise(revisionExerciseID))
        } else {
          const assignments = await getAssignments()
          const currentAssignment = assignments.find((item) => item.is_current && item.is_closed !== true)
          if (currentAssignment) {
            exercise = await getExercise(currentAssignment.current_exercise_id)
          }
        }
        let submission: Submission | undefined
        let review: Review | undefined
        let sourceSubmission: Submission | undefined
        let sourceReview: Review | undefined
        let pendingJob: ReviewJob | null = null

        if (exercise) {
          const submissions = inRevisionMode
            ? await withRetry(() => getSubmissions(exercise.id, 1))
            : await getSubmissions(exercise.id, 1)
          submission = submissions[0]
          if (submission) {
            const currentSubmissionID = submission.id
            const reviews = inRevisionMode
              ? await withRetry(() => getReviews(currentSubmissionID, 1))
              : await getReviews(currentSubmissionID, 1)
            review = reviews[0]
            if (!review) {
              try {
                pendingJob = await getReviewJob(currentSubmissionID)
              } catch {}
            }
          }
          if (exercise.source_submission_id) {
            const sourceSubmissionID = exercise.source_submission_id
            const loadedSourceSubmission = inRevisionMode
              ? await withRetry(() => getSubmission(sourceSubmissionID))
              : await getSubmission(sourceSubmissionID)
            sourceSubmission = loadedSourceSubmission
            const sourceReviews = inRevisionMode
              ? await withRetry(() => getReviews(loadedSourceSubmission.id, 1))
              : await getReviews(loadedSourceSubmission.id, 1)
            sourceReview = sourceReviews[0]
          }
        }

        if (!cancelled) {
          setWorkspace({ dashboard, exercise, submission, review, sourceSubmission, sourceReview })
          setDraft(submission?.content ?? sourceSubmission?.content ?? '')
          setReviewJob(pendingJob)
          setSessionRequired(false)
          setError(null)
          setRevisionPanel(sourceReview ? 'feedback' : 'brief')
        }
      } catch (err) {
        if (!cancelled) {
          const message = err instanceof Error ? err.message : 'Failed to load current assignment'
          if (message.toLowerCase().includes('unauthorized')) {
            setSessionRequired(true)
          } else if (revisionExerciseID > 0 && isTransientError(err)) {
            setError(
              'Could not load the revision workspace on the first attempt. Please try again. The revision brief may already be ready.'
            )
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
  }, [revisionExerciseID, router])

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
  const draftBaseline = workspace?.submission?.content ?? workspace?.sourceSubmission?.content ?? ''
  const hasUnsavedDraft = hasUnsavedTrackDraft({
    hasExercise: workspace?.exercise !== undefined,
    draft,
    baseline: draftBaseline,
  })

  useEffect(() => {
    window.__writingCoachHasUnsavedDraft = hasUnsavedDraft
    return () => {
      window.__writingCoachHasUnsavedDraft = false
    }
  }, [hasUnsavedDraft])

  useEffect(() => {
    if (!hasUnsavedDraft) {
      return
    }
    function handleBeforeUnload(event: BeforeUnloadEvent) {
      event.preventDefault()
      event.returnValue = ''
    }
    window.addEventListener('beforeunload', handleBeforeUnload)
    return () => {
      window.removeEventListener('beforeunload', handleBeforeUnload)
    }
  }, [hasUnsavedDraft])

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

  return {
    loading,
    error,
    sessionRequired,
    workspace,
    draft,
    setDraft,
    reviewing,
    preparingRevision,
    reviewJob,
    revisionPanel,
    setRevisionPanel,
    wordCount,
    hasUnsavedDraft,
    handleReview,
    handleRetryReview,
    handleRevisionPrompt,
  }
}
