'use client'

import { ArrowPathIcon } from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Heading, Subheading } from '@/components/heading'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { useCompareWorkspace } from '@/lib/use-compare-workspace'
import { ProviderProvenance } from './provider-provenance'
import { SkillScoreMeter } from './skill-score-meter'
import { AppErrorState, EmptyState, LoadingState, TaskProgressState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function CompareView({ submissionId }: { submissionId: number }) {
  const { sessionLoading, sessionError, loading, error, comparison, submission, review, preparingRevision, canActOnComparison, prepareRevisionPrompt } =
    useCompareWorkspace(submissionId)

  async function handleRevisionPrompt() {
    const revisionExercise = await prepareRevisionPrompt()
    if (revisionExercise) {
      window.location.href = `/?revisionExercise=${revisionExercise.id}`
    }
  }

  if (sessionLoading || loading) {
    return <LoadingState label="Loading comparison…" />
  }
  if (sessionError || error || !comparison || !submission) {
    return <AppErrorState title="Comparison unavailable" error={sessionError ?? error ?? 'The requested comparison is not available yet.'} />
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
            {canActOnComparison ? (
              <Button onClick={handleRevisionPrompt} color="dark/zinc" disabled={preparingRevision}>
                <ArrowPathIcon />
                {preparingRevision ? 'Preparing revision brief…' : 'Revise again'}
              </Button>
            ) : null}
          </>
        }
      />

      {preparingRevision ? (
        <TaskProgressState
          title="Revision brief in progress"
          body="Turning this comparison into the next revision brief."
          steps={[
            'Load the latest revision history.',
            'Pull forward the weaknesses that still matter most.',
            'Build a new focused revision assignment.',
          ]}
        />
      ) : null}

      {!canActOnComparison ? (
        <WorkspaceCard>
          <Text>This comparison belongs to a completed assignment chain. You can review it here, but the next step is to start a new assignment.</Text>
        </WorkspaceCard>
      ) : null}

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
            <div className="mt-4">
              <ProviderProvenance providerNote={review.provider_note} />
            </div>
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
                <SkillScoreMeter score={item} compact />
              </div>
            ))}
          </div>
        </WorkspaceCard>
      ) : null}
    </div>
  )
}
