'use client'

import { useEffect, useState } from 'react'
import { ArrowPathIcon } from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Heading, Subheading } from '@/components/heading'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { createRevisionAssignment, getComparison, getReviews, getSubmission } from '@/lib/api'
import type { Comparison, Review, Submission } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function CompareView({ submissionId }: { submissionId: number }) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [comparison, setComparison] = useState<Comparison | null>(null)
  const [submission, setSubmission] = useState<Submission | null>(null)
  const [review, setReview] = useState<Review | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
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
  }, [submissionId])

  async function handleRevisionPrompt() {
    if (!submission) {
      return
    }
    try {
      const revisionExercise = await createRevisionAssignment(submission.id)
      window.location.href = `/?revisionExercise=${revisionExercise.id}`
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not create revision prompt')
    }
  }

  if (loading) {
    return <LoadingState label="Loading comparison…" />
  }
  if (error || !comparison || !submission) {
    return <EmptyState title="Comparison unavailable" body={error ?? 'The requested comparison is not available yet.'} actionHref="/" actionLabel="Back to assignment" />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Compare"
        title="Revision compare"
        intro={comparison.summary}
        actions={
          <>
            <Button href={`/assignments/${submission.exercise_id}`} plain>
              View timeline
            </Button>
            <Button onClick={handleRevisionPrompt} color="dark/zinc">
              <ArrowPathIcon />
              Revise again
            </Button>
          </>
        }
      />

      <div className="grid gap-8 lg:grid-cols-3">
        <WorkspaceCard>
          <CardHeader eyebrow="Draft" title="Draft" />
          <Text className="mt-2">Revision draft #{submission.draft_number}</Text>
        </WorkspaceCard>
        <WorkspaceCard>
          <CardHeader eyebrow="Delta" title="Word delta" />
          <div className="mt-3">
            <Badge color={comparison.word_delta >= 0 ? 'green' : 'amber'}>
              {comparison.word_delta >= 0 ? `+${comparison.word_delta}` : comparison.word_delta}
            </Badge>
          </div>
        </WorkspaceCard>
        <WorkspaceCard>
          <CardHeader eyebrow="Persistence" title="Persistence check" />
          <Text className="mt-2">{comparison.persisting_weaknesses.length} weaknesses still carrying forward.</Text>
        </WorkspaceCard>
      </div>

      <div className="grid gap-8 xl:grid-cols-2">
        <WorkspaceCard>
          <CardHeader eyebrow="Resolved" title="Addressed weaknesses" />
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {comparison.addressed_weaknesses.length === 0 ? <li>No weaknesses were marked resolved yet.</li> : null}
            {comparison.addressed_weaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
        <WorkspaceCard>
          <CardHeader eyebrow="Still active" title="Persisting weaknesses" />
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {comparison.persisting_weaknesses.length === 0 ? <li>No persisting weaknesses were flagged.</li> : null}
            {comparison.persisting_weaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>

      {review ? (
        <div className="grid gap-8 xl:grid-cols-[2fr_1fr]">
          <WorkspaceCard>
            <CardHeader eyebrow="State" title="Current skill state" />
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
            <CardHeader
              eyebrow="Maintenance"
              title="Maintenance checks"
              description="Mastered skills are still checked lightly so previously earned skills do not decay unnoticed."
            />
            <div className="mt-4 space-y-3">
              {review.completed_tgo_checks.length === 0 ? <Text>No mastered-skill slips were flagged on this revision.</Text> : null}
              {review.completed_tgo_checks.map((assessment) => (
                <div key={assessment.tgo_code} className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                  <div className="flex items-center justify-between gap-4">
                    <span className="font-semibold text-zinc-950 dark:text-white">{assessment.tgo_title ?? assessment.tgo_code}</span>
                    <Badge color="amber">{assessment.status}</Badge>
                  </div>
                  <Text className="mt-2">{assessment.evidence}</Text>
                </div>
              ))}
            </div>
          </WorkspaceCard>
        </div>
      ) : null}

      {review?.skill_scores.length ? (
        <WorkspaceCard>
          <CardHeader eyebrow="Scores" title="Skill scores" />
          <div className="mt-4 grid gap-4 lg:grid-cols-3">
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
      ) : null}
    </div>
  )
}
