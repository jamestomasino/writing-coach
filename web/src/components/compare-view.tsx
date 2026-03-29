'use client'

import { ArrowPathIcon } from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Heading, Subheading } from '@/components/heading'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { useCompareWorkspace } from '@/lib/use-compare-workspace'
import { useTranslations } from 'next-intl'
import { ProviderProvenance } from './provider-provenance'
import { SkillScoreMeter } from './skill-score-meter'
import { AppErrorState, EmptyState, LoadingState, TaskProgressState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function CompareView({ submissionId }: { submissionId: number }) {
  const t = useTranslations('compareView')
  const { sessionLoading, sessionError, loading, error, comparison, submission, review, preparingRevision, canActOnComparison, prepareRevisionPrompt } =
    useCompareWorkspace(submissionId)

  async function handleRevisionPrompt() {
    const revisionExercise = await prepareRevisionPrompt()
    if (revisionExercise) {
      window.location.href = `/?revisionExercise=${revisionExercise.id}`
    }
  }

  if (sessionLoading || loading) {
    return <LoadingState label={t('loading')} />
  }
  if (sessionError || error || !comparison || !submission) {
    return <AppErrorState title={t('unavailableTitle')} error={sessionError ?? error ?? t('unavailableBody')} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={t('eyebrow')}
        title={t('title')}
        intro={comparison.summary}
        actions={
          <>
            <Button href={`/assignments/${submission.exercise_id}`} plain>
              View timeline
            </Button>
            {canActOnComparison ? (
              <Button onClick={handleRevisionPrompt} color="dark/zinc" disabled={preparingRevision}>
                <ArrowPathIcon />
                {preparingRevision ? t('preparingRevisionBrief') : t('reviseAgain')}
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

      {!canActOnComparison ? (
        <WorkspaceCard>
          <Text>{t('olderAssignmentBody')}</Text>
        </WorkspaceCard>
      ) : null}

      <div className="grid gap-8 lg:grid-cols-3">
        <WorkspaceCard>
          <CardHeader eyebrow={t('currentDraftEyebrow')} title={t('draftTitle')} />
          <Text className="mt-2">{t('revisionDraft', { draftNumber: submission.draft_number })}</Text>
        </WorkspaceCard>
        <WorkspaceCard>
          <CardHeader eyebrow={t('lengthChangeEyebrow')} title={t('lengthChangeTitle')} />
          <div className="mt-3">
            <Badge color={comparison.word_delta >= 0 ? 'green' : 'amber'}>
              {comparison.word_delta >= 0 ? `+${comparison.word_delta}` : comparison.word_delta}
            </Badge>
          </div>
        </WorkspaceCard>
        <WorkspaceCard>
          <CardHeader eyebrow={t('stillUnresolvedEyebrow')} title={t('stillUnresolvedTitle')} />
          <Text className="mt-2">{t('persistingIssues', { count: comparison.persisting_weaknesses.length })}</Text>
        </WorkspaceCard>
      </div>

      <div className="grid gap-8 xl:grid-cols-2">
        <WorkspaceCard>
          <CardHeader eyebrow={t('improvedEyebrow')} title={t('improvedTitle')} />
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {comparison.addressed_weaknesses.length === 0 ? <li>{t('noResolved')}</li> : null}
            {comparison.addressed_weaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
        <WorkspaceCard>
          <CardHeader eyebrow={t('stillActiveEyebrow')} title={t('stillActiveTitle')} />
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {comparison.persisting_weaknesses.length === 0 ? <li>{t('noPersisting')}</li> : null}
            {comparison.persisting_weaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>

      {review ? (
        <div className="grid gap-8 xl:grid-cols-[2fr_1fr]">
          <WorkspaceCard>
            <CardHeader eyebrow={t('currentSkillsEyebrow')} title={t('currentSkillsTitle')} />
            <div className="mt-4">
              <ProviderProvenance providerNote={review.provider_note} kind="feedback" />
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
              eyebrow={t('olderSkillsEyebrow')}
              title={t('olderSkillsTitle')}
              description={t('olderSkillsDescription')}
            />
            <div className="mt-4 space-y-3">
              {review.completed_tgo_checks.length === 0 ? <Text>{t('noOlderSkillsSlipped')}</Text> : null}
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
          <CardHeader eyebrow={t('ratingsEyebrow')} title={t('ratingsTitle')} />
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
