'use client'

import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { createRevisionAssignment, getAssignmentTimeline, getExercise, getReview, getSubmission } from '@/lib/api'
import type { Exercise, Review, Submission } from '@/lib/types'
import { ArrowPathIcon } from '@heroicons/react/16/solid'
import { useEffect, useState } from 'react'
import { ProviderProvenance } from './provider-provenance'
import { SkillScoreMeter } from './skill-score-meter'
import { AppErrorState, LoadingState, TaskProgressState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function ReviewView({ reviewId }: { reviewId: number }) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [review, setReview] = useState<Review | null>(null)
  const [submission, setSubmission] = useState<Submission | null>(null)
  const [exercise, setExercise] = useState<Exercise | null>(null)
  const [preparingRevision, setPreparingRevision] = useState(false)
  const [canActOnReview, setCanActOnReview] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const reviewData = await getReview(reviewId)
        const submissionData = await getSubmission(reviewData.submission_id)
        const exerciseData = await getExercise(submissionData.exercise_id)
        let reviewIsActionable = false
        try {
          const timeline = await getAssignmentTimeline(exerciseData.id)
          reviewIsActionable = timeline.is_current === true && timeline.latest_step_id === `review-${reviewData.id}`
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
    return <AppErrorState title="Review unavailable" error={error ?? 'The requested review could not be loaded.'} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Review"
        title={exercise.title}
        intro={review.summary}
        actions={
          <>
            <Button href={`/assignments/${exercise.id}`} plain>
              View timeline
            </Button>
            {canActOnReview ? (
              <Button onClick={handleRevisionPrompt} color="dark/zinc" disabled={preparingRevision}>
                {preparingRevision ? 'Preparing revision brief…' : 'Revise this draft'}
              </Button>
            ) : null}
            {canActOnReview ? (
              <Button href="/" outline>
                Accept and move on
              </Button>
            ) : null}
          </>
        }
      />

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

      {!canActOnReview ? (
        <WorkspaceCard>
          <Text>
            This review is part of your assignment history. To keep working, return to the active track&apos;s current
            assignment workspace.
          </Text>
        </WorkspaceCard>
      ) : null}

      <WorkspaceCard>
        <CardHeader eyebrow="Generation" title="Provider details" />
        <div className="mt-4 space-y-4">
          <ProviderProvenance providerNote={exercise.provider_note} />
          <ProviderProvenance providerNote={review.provider_note} />
        </div>
      </WorkspaceCard>

      <div className="grid gap-8 xl:grid-cols-[2fr_1fr]">
        <WorkspaceCard>
          <CardHeader eyebrow="Rubric" title="Active skill assessments" />
          <div className="mt-4 space-y-4">
            {review.tgo_assessments.map((assessment) => (
              <div key={assessment.tgo_code} className="rounded-2xl border border-stone-200 p-4 dark:border-white/10">
                <div className="flex items-center justify-between gap-4">
                  <span className="font-semibold text-zinc-950 dark:text-white">
                    {assessment.tgo_title ?? assessment.tgo_code}
                  </span>
                  <Badge
                    color={
                      assessment.status === 'mastered' ? 'green' : assessment.status === 'developing' ? 'amber' : 'zinc'
                    }
                  >
                    {assessment.status}
                  </Badge>
                </div>
                <Text className="mt-2">{assessment.evidence}</Text>
              </div>
            ))}
          </div>
        </WorkspaceCard>

        <WorkspaceCard>
          <CardHeader
            eyebrow="Maintenance"
            title="Mastered skill maintenance"
            description="These checks are lighter than the active rubric, but they help catch regression on already established skills."
          />
          <div className="mt-4 space-y-3">
            {review.completed_tgo_checks.length === 0 ? (
              <Text>No mastered-skill slips were flagged on this pass.</Text>
            ) : (
              review.completed_tgo_checks.map((assessment) => (
                <div
                  key={assessment.tgo_code}
                  className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5"
                >
                  <div className="flex items-center justify-between gap-4">
                    <span className="font-semibold text-zinc-950 dark:text-white">
                      {assessment.tgo_title ?? assessment.tgo_code}
                    </span>
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
          <CardHeader
            eyebrow="Scores"
            title="Skill scores"
            description="These scores summarize the latest draft across the tracked skill dimensions."
          />
          <div className="mt-4 space-y-3">
            {review.skill_scores.map((item) => (
              <div
                key={item.skill}
                className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5"
              >
                <SkillScoreMeter score={item} />
              </div>
            ))}
          </div>
        </WorkspaceCard>
        <WorkspaceCard>
          <CardHeader eyebrow="Strengths" title="Strengths" />
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {review.strengths.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
        <WorkspaceCard>
          <CardHeader eyebrow="Weaknesses" title="Weaknesses" />
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {review.weaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>

      {review.artifacts?.comparison ? (
        <WorkspaceCard>
          <CardHeader
            eyebrow="Revision"
            title="Revision trajectory"
            description={review.artifacts.comparison.summary}
            actions={
              <Button href={`/compare/${submission.id}`} outline>
                <ArrowPathIcon />
                Open full compare
              </Button>
            }
          />
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            <div className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">Addressed weaknesses</div>
              <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
                {review.artifacts.comparison.addressed_weaknesses.length === 0 ? (
                  <li>No earlier weaknesses were clearly resolved yet.</li>
                ) : null}
                {review.artifacts.comparison.addressed_weaknesses.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
            <div className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">Persisting weaknesses</div>
              <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
                {review.artifacts.comparison.persisting_weaknesses.length === 0 ? (
                  <li>No prior weaknesses are carrying forward.</li>
                ) : null}
                {review.artifacts.comparison.persisting_weaknesses.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
          </div>
        </WorkspaceCard>
      ) : null}

      <WorkspaceCard>
        <CardHeader
          eyebrow="Annotations"
          title="Inline coaching markup"
          description="These annotations tie short quoted passages to the active rubric, so revision decisions stay anchored to concrete lines instead of drifting into abstraction."
        />
        <div className="mt-4 space-y-4">
          {review.annotations.length === 0 ? (
            <Text>No line-level annotations were returned for this review.</Text>
          ) : (
            review.annotations.map((item, index) => (
              <div
                key={`${item.quote}-${index}`}
                className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5"
              >
                <div className="flex flex-wrap items-center gap-2">
                  <Badge color={item.severity === 'high' ? 'rose' : item.severity === 'medium' ? 'amber' : 'zinc'}>
                    {item.severity}
                  </Badge>
                  <Badge color="blue">{item.tgo_title ?? item.tgo_code}</Badge>
                  <Badge color="cyan">{item.category}</Badge>
                </div>
                <blockquote className="mt-3 border-l-2 border-stone-300 pl-4 text-sm text-zinc-700 italic dark:border-white/15 dark:text-zinc-200">
                  “{item.quote}”
                </blockquote>
                <Text className="mt-3">{item.comment}</Text>
              </div>
            ))
          )}
        </div>
        <div className="mt-6">
          <CardHeader eyebrow="Analyzer" title="Analyzer findings" />
          <div className="mt-3 space-y-3">
            {review.analyzer_findings.map((item) => (
              <div
                key={item}
                className="rounded-xl border border-stone-200 bg-white p-4 text-sm text-zinc-700 dark:border-white/10 dark:bg-zinc-950 dark:text-zinc-300"
              >
                {item}
              </div>
            ))}
          </div>
        </div>
      </WorkspaceCard>
    </div>
  )
}
