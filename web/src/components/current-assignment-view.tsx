'use client'

import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { Callout } from '@/components/callout'
import { CardHeader } from '@/components/card-header'
import { PageHeader } from '@/components/page-header'
import { Strong, Text } from '@/components/text'
import { Textarea } from '@/components/textarea'
import { formatLocalDateTime } from '@/lib/datetime'
import { useCurrentAssignmentWorkspace } from '@/lib/use-current-assignment-workspace'
import { ArrowPathIcon, ArrowUpTrayIcon, ExclamationTriangleIcon, SparklesIcon } from '@heroicons/react/16/solid'
import { useTranslations } from 'next-intl'
import { useRouter, useSearchParams } from 'next/navigation'
import { useEffect } from 'react'
import { MasteryProgress } from './mastery-progress'
import { ProviderProvenance } from './provider-provenance'
import { AppErrorState, EmptyState, LoadingState, TaskProgressState } from './status-state'
import { WorkspaceCard } from './workspace-card'

function tierLabel(tier?: string, fallback?: string) {
  return (tier ?? fallback ?? '').replace(/-/g, ' ')
}

export function CurrentAssignmentView() {
  const t = useTranslations('currentAssignmentView')
  const router = useRouter()
  const searchParams = useSearchParams()
  const revisionExerciseID = Number(searchParams.get('revisionExercise') ?? 0)
  const {
    loading,
    error,
    sessionRequired,
    workspace,
    draft,
    setDraft,
    reviewing,
    preparingRevision,
    reviewJob,
    revisionPanel,
    setRevisionPanel,
    wordCount,
    handleReview,
    handleRetryReview,
    handleRevisionPrompt,
  } = useCurrentAssignmentWorkspace(revisionExerciseID)

  async function handleFile(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0]
    if (!file) {
      return
    }
    setDraft(await file.text())
  }

  if (loading) {
    return <LoadingState />
  }
  if (sessionRequired) {
    return (
      <EmptyState
        title={t('signInTitle')}
        body={t('signInBody')}
        actionHref="/login"
        actionLabel={t('signInAction')}
      />
    )
  }
  if (error) {
    return <AppErrorState error={error} title={t('workspaceUnavailable')} />
  }
  if (!workspace) {
    return <LoadingState />
  }

  const { dashboard, exercise, review, sourceSubmission, sourceReview } = workspace
  const isRevisionBrief = searchParams.get('revisionExercise') !== null
  const reviewPending = reviewJob?.status === 'queued' || reviewJob?.status === 'running'
  const reviewFailed = reviewJob?.status === 'failed'
  const busy = reviewing || preparingRevision || reviewPending
  const compareSubmissionID = workspace.submission?.id ?? 0
  const queuedReviewTime = formatLocalDateTime(reviewJob?.updated_at)

  if (!exercise) {
    return (
      <div className="space-y-8">
        <PageHeader
          eyebrow={t('workspaceEyebrow')}
          title={t('workspaceTitle')}
          intro={t('noAssignmentIntro')}
        />
        <WorkspaceCard>
          <CardHeader eyebrow={t('skillsEyebrow')} title={t('skillsTitle')} />
          <div className="mt-4 flex flex-wrap gap-2">
            {dashboard.active_tgos.map((tgo) => (
              <Badge key={tgo.code} color="blue">
                {tgo.title}
              </Badge>
            ))}
          </div>
          <div className="mt-6">
            <Button href="/new-assignment" color="dark/zinc">
              {t('startNewAssignment')}
            </Button>
          </div>
        </WorkspaceCard>
      </div>
    )
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={t('workspaceEyebrow')}
        title={exercise.title}
        intro={exercise.brief}
        actions={
          <>
            <Button href={`/assignments/${exercise.id}`} plain>
              View timeline
            </Button>
            {review ? (
              <Button onClick={handleRevisionPrompt} color="dark/zinc" disabled={busy}>
                {preparingRevision ? t('preparingRevisionBrief') : t('createRevisionBrief')}
              </Button>
            ) : null}
          </>
        }
      />
      {reviewPending ? (
        <TaskProgressState
          title={t('reviewProgressTitle')}
          body={t('reviewProgressBody')}
          steps={[
            t('reviewStep1'),
            t('reviewStep2'),
            t('reviewStep3'),
          ]}
        />
      ) : null}

      {preparingRevision ? (
        <TaskProgressState
          title={t('revisionProgressTitle')}
          body={t('revisionProgressBody')}
          steps={[
            t('revisionStep1'),
            t('revisionStep2'),
            t('revisionStep3'),
          ]}
        />
      ) : null}

      {reviewFailed ? (
        <Callout
          tone="warning"
          eyebrow={t('reviewStatusEyebrow')}
          title={t('reviewFailedTitle')}
          body={t('reviewFailedBody')}
          actions={
            <Button onClick={handleRetryReview} color="dark/zinc" disabled={reviewing}>
              {reviewing ? t('retrying') : t('retryReview')}
            </Button>
          }
        >
          <div className="flex items-center gap-2 text-sm font-semibold text-amber-900 dark:text-amber-200">
            <ExclamationTriangleIcon className="size-4" />
            {t('backgroundReviewInterrupted')}
          </div>
          {reviewJob?.last_error ? <Text className="mt-2 text-sm">{reviewJob.last_error}</Text> : null}
        </Callout>
      ) : null}

      {isRevisionBrief ? (
        <WorkspaceCard>
          <CardHeader
            eyebrow={t('revisionModeEyebrow')}
            title={t('revisionModeTitle')}
            description={t('revisionModeDescription')}
            actions={
              review ? (
                <Button href={`/compare/${compareSubmissionID}`} outline>
                  <ArrowPathIcon />
                  Compare drafts
                </Button>
              ) : sourceReview ? (
                <Button href={`/reviews/${sourceReview.id}`} outline>
                  {t('openEarlierFeedback')}
                </Button>
              ) : null
            }
          />
          {sourceReview ? (
            <div className="mt-5">
              <div className="inline-flex rounded-xl border border-stone-200 bg-stone-50 p-1 dark:border-white/10 dark:bg-white/5">
                <button
                  type="button"
                  onClick={() => setRevisionPanel('brief')}
                  className={`rounded-lg px-3 py-2 text-sm font-medium transition ${revisionPanel === 'brief' ? 'bg-white text-zinc-950 shadow-sm dark:bg-white/10 dark:text-white' : 'text-zinc-600 hover:text-zinc-950 dark:text-zinc-300 dark:hover:text-white'}`}
                >
                  {t('revisionBriefTab')}
                </button>
                <button
                  type="button"
                  onClick={() => setRevisionPanel('feedback')}
                  className={`rounded-lg px-3 py-2 text-sm font-medium transition ${revisionPanel === 'feedback' ? 'bg-white text-zinc-950 shadow-sm dark:bg-white/10 dark:text-white' : 'text-zinc-600 hover:text-zinc-950 dark:text-zinc-300 dark:hover:text-white'}`}
                >
                  {t('earlierFeedbackTab')}
                </button>
              </div>
              {revisionPanel === 'brief' ? (
                <div className="mt-4 rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
                  <div className="flex flex-wrap items-center gap-3">
                    <Badge color="zinc">{t('sourceDraft', { draftNumber: sourceSubmission?.draft_number ?? 1 })}</Badge>
                    <Badge color="cyan">{exercise.generation_kind}</Badge>
                  </div>
                  <div className="mt-3">
                    <ProviderProvenance providerNote={exercise.provider_note} compact />
                  </div>
                  <Text className="mt-3">{t('lastReviewedDraftLoaded')}</Text>
                </div>
              ) : (
                <div className="mt-4 space-y-4">
                  <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
                    <div className="text-sm font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400">
                      {t('priorCoachingSummary')}
                    </div>
                    <Text className="mt-3">{sourceReview.summary}</Text>
                  </div>
                  <div className="grid gap-4 lg:grid-cols-2">
                    <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
                      <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('strengthsToPreserve')}</div>
                      <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                        {sourceReview.strengths.map((item) => (
                          <li key={item}>• {item}</li>
                        ))}
                      </ul>
                    </div>
                    <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
                      <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('revisionTargets')}</div>
                      <ul className="mt-3 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                        {sourceReview.weaknesses.map((item) => (
                          <li key={item}>• {item}</li>
                        ))}
                      </ul>
                    </div>
                  </div>
                  <div className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5">
                    <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('nextFocus')}</div>
                    <Text className="mt-3">{sourceReview.next_focus}</Text>
                  </div>
                </div>
              )}
            </div>
          ) : null}
        </WorkspaceCard>
      ) : null}

      <div className="grid gap-8 xl:grid-cols-[2fr_1fr]">
        <WorkspaceCard>
          <CardHeader
            eyebrow={t('briefEyebrow')}
            title={t('briefTitle')}
            actions={<Badge color="zinc">{exercise.generation_kind}</Badge>}
          />
          <div className="mt-4 space-y-5">
            <ProviderProvenance providerNote={exercise.provider_note} />
            <div>
              <Strong>Constraints</Strong>
              <ul className="mt-2 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                {exercise.constraints.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
            <div>
              <Strong>Success criteria</Strong>
              <ul className="mt-2 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                {exercise.success_criteria.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
          </div>
        </WorkspaceCard>

        <WorkspaceCard>
          <CardHeader
            eyebrow={t('skillsEyebrow')}
            title={t('skillsForAssignmentTitle')}
            description={t('skillsForAssignmentDescription')}
          />
          <div className="mt-4 space-y-3">
            {dashboard.active_tgos.map((tgo) => (
              <div
                key={tgo.code}
                className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5"
              >
                <div className="flex items-center justify-between gap-4">
                  <Strong>{tgo.title}</Strong>
                  <Badge color="cyan">{tierLabel(tgo.skill_tier, tgo.stage)}</Badge>
                </div>
                <Text className="mt-2">{tgo.description}</Text>
                <MasteryProgress tgo={tgo} />
              </div>
            ))}
          </div>
        </WorkspaceCard>
      </div>

      <WorkspaceCard>
        <div className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <CardHeader
              eyebrow={t('writeEyebrow')}
              title={t('writeTitle')}
              description={t('writeDescription')}
            />
          </div>
          <div className="flex items-center gap-3">
            <label className="inline-flex cursor-pointer items-center gap-2 rounded-lg border border-stone-300 px-3 py-2 text-sm font-medium text-zinc-700 hover:bg-stone-50 dark:border-white/10 dark:text-zinc-200 dark:hover:bg-white/5">
              <ArrowUpTrayIcon className="size-4" />
              {t('uploadDraft')}
              <input
                type="file"
                accept=".md,.txt,text/plain,text/markdown"
                className="hidden"
                onChange={handleFile}
                disabled={busy}
              />
            </label>
            <Badge color="zinc">{wordCount} words</Badge>
          </div>
        </div>
        <div className="mt-5">
          <Textarea
            value={draft}
            onChange={(event) => setDraft(event.target.value)}
            rows={18}
            placeholder={t('draftPlaceholder')}
            disabled={busy}
            data-testid="draft-textarea"
          />
        </div>
        <div className="mt-5 flex flex-wrap items-center justify-between gap-3">
          <Text>
            {workspace.submission
              ? reviewPending
                ? t('savedDraftQueued', {
                    draftNumber: workspace.submission.draft_number,
                    time: queuedReviewTime ?? t('justNow'),
                  })
                : t('savedDraft', { draftNumber: workspace.submission.draft_number })
              : t('noDraftYet')}
          </Text>
          <Button
            onClick={handleReview}
            color="dark/zinc"
            disabled={busy || draft.trim() === ''}
            data-testid="submit-review-button"
          >
            <SparklesIcon />
            {reviewing || reviewPending ? t('feedbackRequested') : t('getFeedback')}
          </Button>
        </div>
      </WorkspaceCard>

      {review ? (
        <WorkspaceCard>
          <CardHeader
            eyebrow={t('latestFeedbackEyebrow')}
            title={t('latestFeedbackTitle')}
            description={review.summary}
            actions={
              <div className="flex gap-2">
                <Button href={`/assignments/${exercise.id}`} plain>
                  {t('viewTimeline')}
                </Button>
                <Button href={`/reviews/${review.id}`} outline>
                  {t('openReview')}
                </Button>
                <Button href={`/compare/${compareSubmissionID}`} plain>
                  {t('compareDrafts')}
                </Button>
              </div>
            }
          />
          {review.artifacts?.comparison ? (
            <div className="mt-5 grid gap-4 lg:grid-cols-3">
              <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('revisionSummary')}</div>
                <Text className="mt-2 text-sm">{review.artifacts.comparison.summary}</Text>
              </div>
              <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('addressed')}</div>
                <Text className="mt-2 text-sm">{review.artifacts.comparison.addressed_weaknesses.length}</Text>
              </div>
              <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('persisting')}</div>
                <Text className="mt-2 text-sm">{review.artifacts.comparison.persisting_weaknesses.length}</Text>
              </div>
            </div>
          ) : null}
        </WorkspaceCard>
      ) : null}
    </div>
  )
}
