'use client'

import { useEffect, useState } from 'react'
import { ArrowPathIcon } from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { createRevisionAssignment, getExercise, getReview, getSubmission } from '@/lib/api'
import type { Exercise, Review, Submission } from '@/lib/types'
import { EmptyState, LoadingState, TaskProgressState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function ReviewView({ reviewId }: { reviewId: number }) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [review, setReview] = useState<Review | null>(null)
  const [submission, setSubmission] = useState<Submission | null>(null)
  const [exercise, setExercise] = useState<Exercise | null>(null)
  const [preparingRevision, setPreparingRevision] = useState(false)

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
      setPreparingRevision(true)
      setError(null)
      const revisionExercise = await createRevisionAssignment(submission.id)
      window.location.href = `/?revisionExercise=${revisionExercise.id}`
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create revision prompt')
    } finally {
      setPreparingRevision(false)
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
          <Button onClick={handleRevisionPrompt} color="dark/zinc" disabled={preparingRevision}>
            {preparingRevision ? 'Preparing revision brief…' : 'Revise this draft'}
          </Button>
          <Button href="/" outline>
            Accept and move on
          </Button>
        </div>
      </header>

      {preparingRevision ? (
        <TaskProgressState
          title="Revision brief in progress"
          body="The app is translating this review into a focused revision brief."
          steps={[
            'Load the reviewed draft and coaching artifacts.',
            'Select the most urgent revision targets.',
            'Build the next revision assignment around those targets.',
          ]}
        />
      ) : null}

      <div className="grid gap-8 xl:grid-cols-[2fr_1fr]">
        <WorkspaceCard>
          <Subheading>Active skill assessments</Subheading>
          <div className="mt-4 space-y-4">
            {review.tgo_assessments.map((assessment) => (
              <div key={assessment.tgo_code} className="rounded-2xl border border-stone-200 p-4 dark:border-white/10">
                <div className="flex items-center justify-between gap-4">
                  <span className="font-semibold text-zinc-950 dark:text-white">{assessment.tgo_title ?? assessment.tgo_code}</span>
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
          <Subheading>Mastered skill maintenance</Subheading>
          <Text className="mt-2">These checks are lighter than the active rubric, but they help catch regression on already established skills.</Text>
          <div className="mt-4 space-y-3">
            {review.completed_tgo_checks.length === 0 ? (
              <Text>No mastered-skill slips were flagged on this pass.</Text>
            ) : (
              review.completed_tgo_checks.map((assessment) => (
                <div key={assessment.tgo_code} className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                  <div className="flex items-center justify-between gap-4">
                    <span className="font-semibold text-zinc-950 dark:text-white">{assessment.tgo_title ?? assessment.tgo_code}</span>
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
          <Subheading>Skill scores</Subheading>
          <Text className="mt-2">These scores summarize the latest draft across the tracked skill dimensions.</Text>
          <div className="mt-4 space-y-3">
            {review.skill_scores.map((item) => (
              <div key={item.skill} className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                <div className="flex items-center justify-between gap-4">
                  <span className="font-semibold capitalize text-zinc-950 dark:text-white">{item.skill}</span>
                  <Badge color={item.score >= 4 ? 'green' : item.score === 3 ? 'amber' : 'rose'}>{item.score}/5</Badge>
                </div>
              </div>
            ))}
          </div>
        </WorkspaceCard>
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

      {review.artifacts?.comparison ? (
        <WorkspaceCard>
          <div className="flex flex-col gap-3 lg:flex-row lg:items-start lg:justify-between">
            <div>
              <Subheading>Revision trajectory</Subheading>
              <Text className="mt-2">{review.artifacts.comparison.summary}</Text>
            </div>
            <Button href={`/compare/${submission.id}`} outline>
              <ArrowPathIcon />
              Open full compare
            </Button>
          </div>
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            <div className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">Addressed weaknesses</div>
              <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
                {review.artifacts.comparison.addressed_weaknesses.length === 0 ? <li>No earlier weaknesses were clearly resolved yet.</li> : null}
                {review.artifacts.comparison.addressed_weaknesses.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
            <div className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">Persisting weaknesses</div>
              <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
                {review.artifacts.comparison.persisting_weaknesses.length === 0 ? <li>No prior weaknesses are carrying forward.</li> : null}
                {review.artifacts.comparison.persisting_weaknesses.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
          </div>
        </WorkspaceCard>
      ) : null}

      <WorkspaceCard>
        <Subheading>Inline coaching markup</Subheading>
        <Text className="mt-2">These annotations tie short quoted passages to the active rubric, so revision decisions stay anchored to concrete lines instead of drifting into abstraction.</Text>
        <div className="mt-4 space-y-4">
          {review.annotations.length === 0 ? (
            <Text>No line-level annotations were returned for this review.</Text>
          ) : (
            review.annotations.map((item, index) => (
              <div key={`${item.quote}-${index}`} className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                <div className="flex flex-wrap items-center gap-2">
                  <Badge color={item.severity === 'high' ? 'rose' : item.severity === 'medium' ? 'amber' : 'zinc'}>
                    {item.severity}
                  </Badge>
                  <Badge color="blue">{item.tgo_title ?? item.tgo_code}</Badge>
                  <Badge color="cyan">{item.category}</Badge>
                </div>
                <blockquote className="mt-3 border-l-2 border-stone-300 pl-4 text-sm italic text-zinc-700 dark:border-white/15 dark:text-zinc-200">
                  “{item.quote}”
                </blockquote>
                <Text className="mt-3">{item.comment}</Text>
              </div>
            ))
          )}
        </div>
        <div className="mt-6">
          <Subheading>Analyzer findings</Subheading>
          <div className="mt-3 space-y-3">
            {review.analyzer_findings.map((item) => (
              <div key={item} className="rounded-xl border border-stone-200 bg-white p-4 text-sm text-zinc-700 dark:border-white/10 dark:bg-zinc-950 dark:text-zinc-300">
                {item}
              </div>
            ))}
          </div>
        </div>
      </WorkspaceCard>
    </div>
  )
}
