'use client'

import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { useReviewWorkspace } from '@/lib/use-review-workspace'
import { ArrowPathIcon } from '@heroicons/react/16/solid'
import { useTranslations } from 'next-intl'
import { ProviderProvenance } from './provider-provenance'
import { SkillScoreMeter } from './skill-score-meter'
import { AppErrorState, LoadingState, TaskProgressState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function ReviewView({ reviewId }: { reviewId: number }) {
  const t = useTranslations('reviewView')
  const {
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
  } = useReviewWorkspace(reviewId)

  async function handleRevisionPrompt() {
    const revisionExercise = await prepareRevisionPrompt()
    if (revisionExercise) {
      window.location.href = `/?revisionExercise=${revisionExercise.id}`
    }
  }

  async function handleAcceptAndMoveOn() {
    const exerciseId = await acceptAndCloseAssignment()
    if (exerciseId) {
      window.location.href = `/assignments/${exerciseId}?completed=1`
    }
  }

  if (sessionLoading || loading) {
    return <LoadingState label={t('loading')} />
  }
  if (sessionError || error || !review || !submission || !exercise) {
    return <AppErrorState title={t('unavailableTitle')} error={sessionError ?? error ?? t('unavailableBody')} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={t('eyebrow')}
        title={exercise.title}
        intro={review.summary}
        actions={
          <>
            <Button href={`/assignments/${exercise.id}`} plain>
              View timeline
            </Button>
            {canActOnReview ? (
              <Button onClick={handleRevisionPrompt} color="dark/zinc" disabled={preparingRevision}>
                {preparingRevision ? t('preparingRevisionBrief') : t('createRevisionBrief')}
              </Button>
            ) : null}
            {canActOnReview ? (
              <Button onClick={handleAcceptAndMoveOn} outline disabled={closingAssignment}>
                {closingAssignment ? t('finishingAssignment') : t('finishAndMoveOn')}
              </Button>
            ) : null}
          </>
        }
      />

      {preparingRevision ? (
        <TaskProgressState
          title={t('revisionProgressTitle')}
          body={t('revisionProgressBody')}
          steps={[t('revisionStep1'), t('revisionStep2'), t('revisionStep3')]}
        />
      ) : null}

      {!canActOnReview ? (
        <WorkspaceCard>
          <Text>{t('olderAssignmentBody')}</Text>
        </WorkspaceCard>
      ) : null}

      <WorkspaceCard>
        <CardHeader eyebrow={t('aiDetailsEyebrow')} title={t('aiDetailsTitle')} />
        <div className="mt-4 space-y-4">
          <ProviderProvenance providerNote={exercise.provider_note} />
          <ProviderProvenance providerNote={review.provider_note} kind="feedback" />
        </div>
      </WorkspaceCard>

      <div className="grid gap-8 xl:grid-cols-[2fr_1fr]">
        <WorkspaceCard>
          <CardHeader eyebrow={t('currentSkillsEyebrow')} title={t('currentSkillsTitle')} />
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
            eyebrow={t('olderSkillsEyebrow')}
            title={t('olderSkillsTitle')}
            description={t('olderSkillsDescription')}
          />
          <div className="mt-4 space-y-3">
            {review.completed_tgo_checks.length === 0 ? (
              <Text>{t('noOlderSkillsSlipped')}</Text>
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
            eyebrow={t('ratingsEyebrow')}
            title={t('ratingsTitle')}
            description={t('ratingsDescription')}
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
          <CardHeader eyebrow={t('workedEyebrow')} title={t('workedTitle')} />
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {review.strengths.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
        <WorkspaceCard>
          <CardHeader eyebrow={t('improveEyebrow')} title={t('improveTitle')} />
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
            eyebrow={t('revisionEyebrow')}
            title={t('revisionTitle')}
            description={review.artifacts.comparison.summary}
            actions={
              <Button href={`/compare/${submission.id}`} outline>
                <ArrowPathIcon />
                {t('openFullCompare')}
              </Button>
            }
          />
          <div className="mt-5 grid gap-4 lg:grid-cols-2">
            <div className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('whatImproved')}</div>
              <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
                {review.artifacts.comparison.addressed_weaknesses.length === 0 ? (
                  <li>{t('noResolvedWeaknesses')}</li>
                ) : null}
                {review.artifacts.comparison.addressed_weaknesses.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
            <div className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('whatStillNeedsWork')}</div>
              <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
                {review.artifacts.comparison.persisting_weaknesses.length === 0 ? (
                  <li>{t('noPersistingWeaknesses')}</li>
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
          eyebrow={t('lineNotesEyebrow')}
          title={t('lineNotesTitle')}
          description={t('lineNotesDescription')}
        />
        <div className="mt-4 space-y-4">
          {review.annotations.length === 0 ? (
            <Text>{t('noLineNotes')}</Text>
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
          <CardHeader eyebrow={t('signalsEyebrow')} title={t('signalsTitle')} />
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
