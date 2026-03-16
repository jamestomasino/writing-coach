'use client'

import { useEffect, useState } from 'react'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { createRevisionAssignment, getExercise, getReview, getSubmission } from '@/lib/api'
import type { Exercise, Review, Submission } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function ReviewView({ reviewId }: { reviewId: number }) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [review, setReview] = useState<Review | null>(null)
  const [submission, setSubmission] = useState<Submission | null>(null)
  const [exercise, setExercise] = useState<Exercise | null>(null)
  const [revisionCreated, setRevisionCreated] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const reviewData = await getReview(reviewId)
        const submissionData = await getSubmission(reviewData.submission_id)
        const exerciseData = await getExercise(submissionData.exercise_id)
        if (!cancelled) {
          setReview(reviewData)
          setSubmission(submissionData)
          setExercise(exerciseData)
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
  }, [reviewId])

  async function handleRevisionPrompt() {
    if (!submission) {
      return
    }
    try {
      await createRevisionAssignment(submission.id)
      setRevisionCreated(true)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create revision prompt')
    }
  }

  if (loading) {
    return <LoadingState label="Loading review…" />
  }
  if (error || !review || !submission || !exercise) {
    return <EmptyState title="Review unavailable" body={error ?? 'The requested review could not be loaded.'} actionHref="/" actionLabel="Back to assignment" />
  }

  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <Heading>{exercise.title}</Heading>
          <Text className="mt-2 max-w-3xl">{review.summary}</Text>
        </div>
        <div className="flex gap-2">
          <Button onClick={handleRevisionPrompt} color="dark/zinc">
            Revise this draft
          </Button>
          <Button href="/" outline>
            Accept and move on
          </Button>
        </div>
      </header>

      {revisionCreated ? (
        <WorkspaceCard>
          <Subheading>Revision brief created</Subheading>
          <Text className="mt-2">A revision assignment has been generated. Return to the assignment workspace to continue drafting against it.</Text>
        </WorkspaceCard>
      ) : null}

      <div className="grid gap-8 xl:grid-cols-[2fr_1fr]">
        <WorkspaceCard>
          <Subheading>Active TGO assessments</Subheading>
          <div className="mt-4 space-y-4">
            {review.tgo_assessments.map((assessment) => (
              <div key={assessment.tgo_code} className="rounded-2xl border border-stone-200 p-4 dark:border-white/10">
                <div className="flex items-center justify-between gap-4">
                  <span className="font-semibold text-zinc-950 dark:text-white">{assessment.tgo_code}</span>
                  <Badge color={assessment.status === 'mastered' ? 'green' : assessment.status === 'developing' ? 'amber' : 'zinc'}>
                    {assessment.status}
                  </Badge>
                </div>
                <Text className="mt-2">{assessment.evidence}</Text>
              </div>
            ))}
          </div>
        </WorkspaceCard>

        <WorkspaceCard>
          <Subheading>Completed TGO maintenance</Subheading>
          <Text className="mt-2">These checks are lighter than the active rubric, but they help catch regression on already established skills.</Text>
          <div className="mt-4 space-y-3">
            {review.completed_tgo_checks.length === 0 ? (
              <Text>No completed TGO slips were flagged on this pass.</Text>
            ) : (
              review.completed_tgo_checks.map((assessment) => (
                <div key={assessment.tgo_code} className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                  <div className="flex items-center justify-between gap-4">
                    <span className="font-semibold text-zinc-950 dark:text-white">{assessment.tgo_code}</span>
                    <Badge color="amber">{assessment.status}</Badge>
                  </div>
                  <Text className="mt-2">{assessment.evidence}</Text>
                </div>
              ))
            )}
          </div>
        </WorkspaceCard>
      </div>

      <div className="grid gap-8 xl:grid-cols-2">
        <WorkspaceCard>
          <Subheading>Strengths</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {review.strengths.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>Weaknesses</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {review.weaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>

      <WorkspaceCard>
        <Subheading>Coach notes</Subheading>
        <Text className="mt-2">The current API does not yet emit full inline annotation spans, so this first slice surfaces evidence notes and analyzer findings as the browser-facing markup layer.</Text>
        <div className="mt-4 space-y-3">
          {review.analyzer_findings.map((item) => (
            <div key={item} className="rounded-xl border border-stone-200 bg-stone-50 p-4 text-sm text-zinc-700 dark:border-white/10 dark:bg-white/5 dark:text-zinc-300">
              {item}
            </div>
          ))}
        </div>
      </WorkspaceCard>
    </div>
  )
}
