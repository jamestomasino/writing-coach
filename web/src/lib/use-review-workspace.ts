'use client'

import {
  closeAssignment,
  createRevisionAssignment,
  getAIJob,
  getAssignmentTimeline,
  getExercise,
  getReview,
  getSubmission,
} from '@/lib/api'
import type { Exercise, Review, Submission } from '@/lib/types'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import { useEffect, useState } from 'react'

export function useReviewWorkspace(reviewId: number) {
  const { session, loading: sessionLoading, error: sessionError } = useRequiredAppSession(`/reviews/${reviewId}`)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [review, setReview] = useState<Review | null>(null)
  const [submission, setSubmission] = useState<Submission | null>(null)
  const [exercise, setExercise] = useState<Exercise | null>(null)
  const [preparingRevision, setPreparingRevision] = useState(false)
  const [closingAssignment, setClosingAssignment] = useState(false)
  const [canActOnReview, setCanActOnReview] = useState(false)

  useEffect(() => {
    let cancelled = false

    async function load() {
      if (!session) {
        return
      }
      try {
        const reviewData = await getReview(reviewId)
        const submissionData = await getSubmission(reviewData.submission_id)
        const exerciseData = await getExercise(submissionData.exercise_id)
        let reviewIsActionable = false
        try {
          const timeline = await getAssignmentTimeline(exerciseData.id)
          reviewIsActionable = timeline.is_current === true && timeline.is_closed !== true && timeline.latest_step_id === `review-${reviewData.id}`
        } catch {
          reviewIsActionable = false
        }
        if (!cancelled) {
          setReview(reviewData)
          setSubmission(submissionData)
          setExercise(exerciseData)
          setCanActOnReview(reviewIsActionable)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load review')
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
  }, [reviewId, session])

  async function prepareRevisionPrompt() {
    if (!submission) {
      return null
    }
    try {
      setPreparingRevision(true)
      setError(null)
      const job = await createRevisionAssignment(submission.id)
      for (let attempt = 0; attempt < 120; attempt += 1) {
        const nextJob = attempt === 0 ? job : await getAIJob(job.id)
        if (nextJob.status === 'completed') {
          return nextJob.result?.exercise ?? null
        }
        if (nextJob.status === 'failed') {
          throw new Error(nextJob.last_error || 'Could not create revision prompt')
        }
        await new Promise((resolve) => window.setTimeout(resolve, 1500))
      }
      throw new Error('Could not create revision prompt')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create revision prompt')
      return null
    } finally {
      setPreparingRevision(false)
    }
  }

  async function acceptAndCloseAssignment() {
    if (!exercise) {
      return null
    }
    try {
      setClosingAssignment(true)
      setError(null)
      await closeAssignment(exercise.id)
      return exercise.id
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not finish assignment')
      return null
    } finally {
      setClosingAssignment(false)
    }
  }

  return {
    sessionLoading,
    sessionError,
    loading,
    error,
    review,
    submission,
    exercise,
    preparingRevision,
    closingAssignment,
    canActOnReview,
    prepareRevisionPrompt,
    acceptAndCloseAssignment,
  }
}
