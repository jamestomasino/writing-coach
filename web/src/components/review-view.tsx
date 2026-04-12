'use client'

import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { Link } from '@/components/link'
import { objectiveConceptKey } from '@/lib/objective-concepts'
import type { ReviewAnnotation } from '@/lib/types'
import { skillLevelUpState } from '@/lib/skill-level-up'
import { useReviewWorkspace } from '@/lib/use-review-workspace'
import { ArrowPathIcon, InformationCircleIcon } from '@heroicons/react/16/solid'
import { useTranslations } from 'next-intl'
import { useEffect, useMemo, useState } from 'react'
import { ProviderProvenance } from './provider-provenance'
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
    dashboard,
    comparison,
    preparingRevision,
    closingAssignment,
    reopeningAssignment,
    canActOnReview,
    assignmentClosed,
    prepareRevisionPrompt,
    acceptAndCloseAssignment,
    reopenClosedAssignment,
  } = useReviewWorkspace(reviewId)
  const [dismissedNotes, setDismissedNotes] = useState<Set<string>>(new Set())
  const [showDismissedNotes, setShowDismissedNotes] = useState(false)

  const dismissStorageKey = useMemo(() => `review-intentional-notes-${reviewId}`, [reviewId])

  useEffect(() => {
    try {
      const stored = window.localStorage.getItem(dismissStorageKey)
      if (!stored) {
        setDismissedNotes(new Set())
        return
      }
      const parsed = JSON.parse(stored)
      if (Array.isArray(parsed)) {
        setDismissedNotes(new Set(parsed.filter((item): item is string => typeof item === 'string')))
      }
    } catch {
      setDismissedNotes(new Set())
    }
  }, [dismissStorageKey])

  function noteKey(item: ReviewAnnotation, index: number) {
    return `${item.tgo_code}:${item.quote}:${index}`
  }

  function markNoteIntentional(item: ReviewAnnotation, index: number) {
    const key = noteKey(item, index)
    const next = new Set(dismissedNotes)
    next.add(key)
    setDismissedNotes(next)
    window.localStorage.setItem(dismissStorageKey, JSON.stringify(Array.from(next)))
  }

  function restoreNote(item: ReviewAnnotation, index: number) {
    const key = noteKey(item, index)
    const next = new Set(dismissedNotes)
    next.delete(key)
    setDismissedNotes(next)
    window.localStorage.setItem(dismissStorageKey, JSON.stringify(Array.from(next)))
  }

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

  async function handleReopenAssignment() {
    const reopened = await reopenClosedAssignment()
    if (reopened) {
      window.location.reload()
    }
  }

  if (sessionLoading || loading) {
    return <LoadingState label={t('loading')} />
  }
  if (sessionError || error || !review || !submission || !exercise) {
    return <AppErrorState title={t('unavailableTitle')} error={sessionError ?? error ?? t('unavailableBody')} />
  }

  const annotationEntries = review.annotations.map((item, index) => ({ item, index }))
  const visibleAnnotations = annotationEntries.filter(({ item, index }) => !dismissedNotes.has(noteKey(item, index)))
  const hiddenAnnotations = annotationEntries.filter(({ item, index }) => dismissedNotes.has(noteKey(item, index)))
  const developingCount = review.tgo_assessments.filter((item) => item.status !== 'mastered').length
  const masteryCount = review.tgo_assessments.length - developingCount

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
          <Text>{assignmentClosed ? t('closedAssignmentBody') : t('olderAssignmentBody')}</Text>
          <div className="mt-4 flex flex-wrap gap-3">
            {assignmentClosed ? (
              <Button onClick={handleReopenAssignment} color="dark/zinc" disabled={reopeningAssignment}>
                {reopeningAssignment ? t('reopeningAssignment') : t('reopenAssignment')}
              </Button>
            ) : null}
            <Button href="/playground" outline>
              {t('experimentInPlayground')}
            </Button>
          </div>
        </WorkspaceCard>
      ) : null}

      <WorkspaceCard>
        <CardHeader eyebrow={t('guideEyebrow')} title={t('guideTitle')} description={t('guideBody')} />
        <ul className="mt-4 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
          <li>• {t('guideBullet1')}</li>
          <li>• {t('guideBullet2')}</li>
          <li>• {t('guideBullet3')}</li>
        </ul>
      </WorkspaceCard>

      <WorkspaceCard>
        <CardHeader
          eyebrow={t('nextMoveEyebrow')}
          title={developingCount > 0 ? t('nextMoveReviseTitle') : t('nextMoveAdvanceTitle')}
          description={
            developingCount > 0
              ? t('nextMoveReviseBody', { count: developingCount })
              : t('nextMoveAdvanceBody', { count: masteryCount })
          }
        />
        <ul className="mt-4 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
          {developingCount > 0 ? (
            <>
              <li>• {t('nextMoveReviseStep1')}</li>
              <li>• {t('nextMoveReviseStep2')}</li>
              <li>• {t('nextMoveReviseStep3')}</li>
            </>
          ) : (
            <>
              <li>• {t('nextMoveAdvanceStep1')}</li>
              <li>• {t('nextMoveAdvanceStep2')}</li>
              <li>• {t('nextMoveAdvanceStep3')}</li>
            </>
          )}
        </ul>
      </WorkspaceCard>

      {dashboard?.active_tgos?.length ? (
        <WorkspaceCard>
          <CardHeader eyebrow={t('unlockEyebrow')} title={t('unlockTitle')} description={t('unlockBody')} />
          <div className="mt-4 space-y-3">
            {dashboard.active_tgos.slice(0, 3).map((tgo) => {
              const state = skillLevelUpState(tgo)
              const showSkillInfo = tgo.code.trim().length > 0
              return (
                <div
                  key={tgo.code}
                  className="rounded-xl border border-stone-200 bg-stone-50 p-3 text-sm dark:border-white/10 dark:bg-white/5"
                >
                  <div className="flex items-center gap-2">
                    <div className="font-semibold text-zinc-950 dark:text-white">{tgo.title}</div>
                    {showSkillInfo ? (
                      <Link
                        href={`/skills/${encodeURIComponent(objectiveConceptKey(tgo.title))}`}
                        aria-label={`Open ${tgo.title} details`}
                        className="inline-flex items-center justify-center rounded-full border border-stone-300 bg-white p-0.5 text-zinc-500 data-hover:text-zinc-900 dark:border-white/15 dark:bg-black/10 dark:text-zinc-400 dark:data-hover:text-zinc-100"
                      >
                        <InformationCircleIcon className="size-4" aria-hidden="true" />
                      </Link>
                    ) : null}
                  </div>
                  <div className="mt-1 text-xs text-zinc-600 dark:text-zinc-400">
                    {t('unlockEvidenceLine', {
                      current: Math.max(0, tgo.mastery_evidence_count ?? 0),
                      target: 3,
                    })}
                  </div>
                  <div className="mt-1 text-zinc-700 dark:text-zinc-300">
                    {state.mode === 'ready'
                      ? t('unlockReady')
                      : state.mode === 'building_history'
                        ? t('unlockBuildingHistory', { count: state.remainingHistory })
                        : t('unlockConsolidating')}
                  </div>
                </div>
              )
            })}
          </div>
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
                {(() => {
                  const relatedAnnotations = review.annotations
                    .filter((item) => item.tgo_code === assessment.tgo_code)
                    .slice(0, 2)

                  if (relatedAnnotations.length === 0) {
                    return <Text className="mt-3 text-sm text-zinc-500 dark:text-zinc-400">{t('evidenceFromDraftEmpty')}</Text>
                  }

                  return (
                    <div className="mt-3 rounded-xl border border-stone-200 bg-stone-50/80 p-3 dark:border-white/10 dark:bg-white/5">
                      <div className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-600 dark:text-zinc-300">
                        {t('evidenceFromDraftTitle')}
                      </div>
                      <div className="mt-2 text-xs text-zinc-600 dark:text-zinc-400">{t('evidenceFromDraftBody')}</div>
                      <div className="mt-3 space-y-2">
                        {relatedAnnotations.map((item, index) => (
                          <div
                            key={`${assessment.tgo_code}-${index}-${item.quote}`}
                            className="rounded-lg border border-stone-200 bg-white px-3 py-2 dark:border-white/10 dark:bg-black/10"
                          >
                            <blockquote className="text-xs italic text-zinc-700 dark:text-zinc-200">“{item.quote}”</blockquote>
                            <div className="mt-1 text-xs text-zinc-600 dark:text-zinc-300">{item.comment}</div>
                          </div>
                        ))}
                      </div>
                    </div>
                  )
                })()}
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
        {comparison?.skill_deltas && comparison.skill_deltas.length > 0 ? (
          <WorkspaceCard>
            <CardHeader
              eyebrow={t('scoreMovementEyebrow')}
              title={t('scoreMovementTitle')}
              description={t('scoreMovementDescription')}
            />
            <div className="mt-4 space-y-3">
              {comparison.skill_deltas.map((item) => (
                <div
                  key={item.skill}
                  className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5"
                >
                  <div className="flex items-center justify-between gap-3">
                    <div className="text-sm font-semibold capitalize text-zinc-900 dark:text-white">
                      {t('scoreSignalLabel', { skill: item.skill })}
                    </div>
                    <Badge color={item.direction === 'up' ? 'green' : item.direction === 'down' ? 'amber' : 'zinc'}>
                      {item.direction === 'up'
                        ? t('scoreMovementUp')
                        : item.direction === 'down'
                          ? t('scoreMovementDown')
                          : t('scoreMovementFlat')}
                    </Badge>
                  </div>
                  <Text className="mt-2 text-sm">
                    {t('scoreMovementLine', { baseline: item.baseline_score, current: item.current_score })}
                  </Text>
                  {item.evidence_quotes && item.evidence_quotes.length > 0 ? (
                    <ul className="mt-2 space-y-1 text-sm text-zinc-700 dark:text-zinc-300">
                      {item.evidence_quotes.slice(0, 2).map((quote) => (
                        <li key={quote}>• “{quote}”</li>
                      ))}
                    </ul>
                  ) : (
                    <Text className="mt-2 text-sm text-zinc-500 dark:text-zinc-400">{t('scoreMovementNoEvidence')}</Text>
                  )}
                </div>
              ))}
            </div>
          </WorkspaceCard>
        ) : null}
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
            <>
              <div className="flex items-center justify-between gap-3 text-xs text-zinc-600 dark:text-zinc-400">
                <span>{t('lineNotesIntentionalHint')}</span>
                {hiddenAnnotations.length > 0 ? (
                  <button
                    type="button"
                    onClick={() => setShowDismissedNotes((value) => !value)}
                    className="rounded-md border border-stone-300 px-2 py-1 font-medium text-zinc-700 hover:bg-stone-50 dark:border-white/15 dark:text-zinc-200 dark:hover:bg-white/5"
                  >
                    {showDismissedNotes
                      ? t('hideIntentionalNotes')
                      : t('showIntentionalNotes', { count: hiddenAnnotations.length })}
                  </button>
                ) : null}
              </div>
              {visibleAnnotations.map(({ item, index }) => (
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
                <div className="mt-3">
                  <button
                    type="button"
                    onClick={() => markNoteIntentional(item, index)}
                    className="rounded-md border border-stone-300 px-2.5 py-1.5 text-xs font-medium text-zinc-700 hover:bg-stone-100 dark:border-white/15 dark:text-zinc-200 dark:hover:bg-white/10"
                  >
                    {t('markAsIntentional')}
                  </button>
                </div>
              </div>
              ))}
              {showDismissedNotes
                ? hiddenAnnotations.map(({ item, index }) => (
                    <div
                      key={`dismissed-${item.quote}-${index}`}
                      className="rounded-xl border border-stone-200/80 bg-white/70 p-4 opacity-75 dark:border-white/10 dark:bg-white/5"
                    >
                      <div className="flex flex-wrap items-center gap-2">
                        <Badge color="zinc">{t('intentionalChoice')}</Badge>
                        <Badge color="blue">{item.tgo_title ?? item.tgo_code}</Badge>
                      </div>
                      <blockquote className="mt-3 border-l-2 border-stone-300 pl-4 text-sm text-zinc-700 italic dark:border-white/15 dark:text-zinc-200">
                        “{item.quote}”
                      </blockquote>
                      <Text className="mt-3">{item.comment}</Text>
                      <div className="mt-3">
                        <button
                          type="button"
                          onClick={() => restoreNote(item, index)}
                          className="rounded-md border border-stone-300 px-2.5 py-1.5 text-xs font-medium text-zinc-700 hover:bg-stone-100 dark:border-white/15 dark:text-zinc-200 dark:hover:bg-white/10"
                        >
                          {t('restoreToActiveNotes')}
                        </button>
                      </div>
                    </div>
                  ))
                : null}
            </>
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
