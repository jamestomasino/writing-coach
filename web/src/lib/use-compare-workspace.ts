'use client'

import { createRevisionAssignment, getAssignmentTimeline, getComparison, getReviews, getSubmission } from '@/lib/api'
import type { Comparison, Review, Submission } from '@/lib/types'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import { useEffect, useState } from 'react'

export function useCompareWorkspace(submissionId: number) {
  const { session, loading: sessionLoading, error: sessionError } = useRequiredAppSession(`/compare/${submissionId}`)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [comparison, setComparison] = useState<Comparison | null>(null)
  const [submission, setSubmission] = useState<Submission | null>(null)
  const [review, setReview] = useState<Review | null>(null)
  const [preparingRevision, setPreparingRevision] = useState(false)
  const [canActOnComparison, setCanActOnComparison] = useState(false)

  useEffect(() => {
    let cancelled = false

    async function load() {
      if (!session) {
        return
      }
      try {
        const submissionData = await getSubmission(submissionId)
        const [comparisonData, reviews] = await Promise.all([getComparison(submissionId), getReviews(submissionId, 1)])
        const latestReview = reviews[0] ?? null
        let comparisonIsActionable = false
        if (latestReview) {
          try {
            const timeline = await getAssignmentTimeline(submissionData.exercise_id)
            comparisonIsActionable = timeline.is_current === true && timeline.is_closed !== true && timeline.latest_step_id === `review-${latestReview.id}`
          } catch {}
        }
        if (!cancelled) {
          setComparison(comparisonData)
          setSubmission(submissionData)
          setReview(latestReview)
          setCanActOnComparison(comparisonIsActionable)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load comparison')
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
  }, [session, submissionId])

  async function prepareRevisionPrompt() {
    if (!submission) {
      return null
    }
    try {
      setPreparingRevision(true)
      return await createRevisionAssignment(submission.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create revision prompt')
      return null
    } finally {
      setPreparingRevision(false)
    }
  }

  return {
    sessionLoading,
    sessionError,
    loading,
    error,
    comparison,
    submission,
    review,
    preparingRevision,
    canActOnComparison,
    prepareRevisionPrompt,
  }
}
