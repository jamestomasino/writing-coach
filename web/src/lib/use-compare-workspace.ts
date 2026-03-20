'use client'

import { createRevisionAssignment, getComparison, getReviews, getSubmission } from '@/lib/api'
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

  useEffect(() => {
    let cancelled = false

    async function load() {
      if (!session) {
        return
      }
      try {
        const submissionData = await getSubmission(submissionId)
        const [comparisonData, reviews] = await Promise.all([getComparison(submissionId), getReviews(submissionId, 1)])
        if (!cancelled) {
          setComparison(comparisonData)
          setSubmission(submissionData)
          setReview(reviews[0] ?? null)
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
    prepareRevisionPrompt,
  }
}
