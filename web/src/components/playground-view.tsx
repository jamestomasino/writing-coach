'use client'

import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Description, Field, FieldGroup, Fieldset, Label } from '@/components/fieldset'
import { Input } from '@/components/input'
import { PageHeader } from '@/components/page-header'
import { ProviderProvenance } from '@/components/provider-provenance'
import { SkillScoreMeter } from '@/components/skill-score-meter'
import { AppErrorState, LoadingState, TaskProgressState } from '@/components/status-state'
import { Text } from '@/components/text'
import { Textarea } from '@/components/textarea'
import { WorkspaceCard } from '@/components/workspace-card'
import { formatLocalDateTime } from '@/lib/datetime'
import {
  createPlaygroundSession,
  createPlaygroundSessionReview,
  getAIJob,
  getPlaygroundSession,
  getPlaygroundSessionReviews,
  updatePlaygroundSession,
} from '@/lib/api'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import type { AIJob, PlaygroundReview, PlaygroundReviewInput, PlaygroundSession } from '@/lib/types'
import { useTranslations } from 'next-intl'
import { FormEvent, useEffect, useMemo, useState } from 'react'
import { useRouter } from 'next/navigation'

const initialForm: PlaygroundReviewInput = {
  content: '',
  writing_language: 'en',
  writing_type: '',
  assignment_format: '',
  coaching_brief: '',
}

function localWordCount(value: string) {
  return value
    .trim()
    .split(/\s+/)
    .filter(Boolean).length
}

function sessionToForm(session: PlaygroundSession): PlaygroundReviewInput {
  return {
    content: session.content,
    writing_language: session.writing_language || 'en',
    writing_type: session.writing_type ?? '',
    assignment_format: session.assignment_format ?? '',
    coaching_brief: session.coaching_brief ?? '',
  }
}

export function PlaygroundView({ sessionId }: { sessionId?: number }) {
  const t = useTranslations('playgroundView')
  const router = useRouter()
  const { loading: sessionLoading, error: sessionError } = useRequiredAppSession(sessionId ? `/playground/${sessionId}` : '/playground')
  const [form, setForm] = useState<PlaygroundReviewInput>(initialForm)
  const [savedSession, setSavedSession] = useState<PlaygroundSession | null>(null)
  const [reviews, setReviews] = useState<PlaygroundReview[]>([])
  const [selectedReviewID, setSelectedReviewID] = useState<number | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [loadingSession, setLoadingSession] = useState(Boolean(sessionId))
  const [saving, setSaving] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    if (!sessionId) {
      setLoadingSession(false)
      return
    }
    const nextSessionID = sessionId
    let cancelled = false
    async function load() {
      setLoadingSession(true)
      try {
        const [nextSession, nextReviews] = await Promise.all([
          getPlaygroundSession(nextSessionID),
          getPlaygroundSessionReviews(nextSessionID),
        ])
        if (cancelled) {
          return
        }
        setSavedSession(nextSession)
        setForm(sessionToForm(nextSession))
        setReviews(nextReviews)
        setSelectedReviewID(nextReviews[0]?.id ?? null)
        setError(null)
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : t('loadError'))
        }
      } finally {
        if (!cancelled) {
          setLoadingSession(false)
        }
      }
    }
    void load()
    return () => {
      cancelled = true
    }
  }, [sessionId, t])

  const currentReview = useMemo(() => {
    if (reviews.length === 0) {
      return null
    }
    return reviews.find((item) => item.id === selectedReviewID) ?? reviews[0]
  }, [reviews, selectedReviewID])

  async function waitForReview(job: AIJob) {
    for (let attempt = 0; attempt < 120; attempt += 1) {
      const nextJob = attempt === 0 ? job : await getAIJob(job.id)
      if (nextJob.status === 'completed') {
        if (nextJob.result?.review) {
          return nextJob.result.review
        }
        throw new Error(t('reviewError'))
      }
      if (nextJob.status === 'failed') {
        throw new Error(nextJob.last_error || t('reviewError'))
      }
      await new Promise((resolve) => window.setTimeout(resolve, 1500))
    }
    throw new Error(t('reviewError'))
  }

  async function persistSession() {
    const payload = {
      content: form.content,
      writing_language: form.writing_language ?? 'en',
      writing_type: form.writing_type ?? '',
      assignment_format: form.assignment_format ?? '',
      coaching_brief: form.coaching_brief ?? '',
    }
    if (savedSession) {
      const next = await updatePlaygroundSession(savedSession.id, payload)
      setSavedSession(next)
      return next
    }
    const created = await createPlaygroundSession(payload)
    setSavedSession(created)
    router.replace(`/playground/${created.id}`)
    return created
  }

  async function refreshSessionReviews(nextSessionID: number) {
    const [nextSession, nextReviews] = await Promise.all([
      getPlaygroundSession(nextSessionID),
      getPlaygroundSessionReviews(nextSessionID),
    ])
    setSavedSession(nextSession)
    setReviews(nextReviews)
    setSelectedReviewID(nextReviews[0]?.id ?? null)
  }

  async function handleSave() {
    if (saving || submitting || form.content.trim() === '') {
      return
    }
    setSaving(true)
    setError(null)
    try {
      await persistSession()
    } catch (err) {
      setError(err instanceof Error ? err.message : t('saveError'))
    } finally {
      setSaving(false)
    }
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (submitting) {
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      const nextSession = await persistSession()
      const job = await createPlaygroundSessionReview(nextSession.id)
      await waitForReview(job)
      await refreshSessionReviews(nextSession.id)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('reviewError'))
    } finally {
      setSubmitting(false)
    }
  }

  function handleReset() {
    if (savedSession) {
      setForm(sessionToForm(savedSession))
      return
    }
    setForm(initialForm)
    setError(null)
  }

  if (sessionLoading || loadingSession) {
    return <LoadingState label={t('loading')} />
  }
  if (sessionError) {
    return <AppErrorState title={t('unavailableTitle')} error={sessionError} />
  }
  if (error && sessionId && !savedSession) {
    return <AppErrorState title={t('unavailableTitle')} error={error} />
  }

  const wordCount = localWordCount(form.content ?? '')
  const pageTitle = savedSession ? t('sessionTitle', { title: savedSession.title }) : t('title')

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={t('eyebrow')}
        title={pageTitle}
        intro={t('intro')}
        actions={
          <Button href="/playground/history" outline>
            {t('openHistory')}
          </Button>
        }
      />

      {submitting ? (
        <TaskProgressState
          title={t('reviewProgressTitle')}
          body={t('reviewProgressBody')}
          steps={[t('reviewStep1'), t('reviewStep2'), t('reviewStep3')]}
        />
      ) : null}

      <WorkspaceCard>
        <CardHeader
          eyebrow={t('inputEyebrow')}
          title={t('inputTitle')}
          description={t('inputDescription')}
        />
        <form className="mt-6 space-y-6" onSubmit={handleSubmit}>
          <Fieldset>
            <FieldGroup>
              <Field>
                <Label>{t('contentLabel')}</Label>
                <Description>{t('contentHelp')}</Description>
                <Textarea
                  name="content"
                  rows={18}
                  value={form.content}
                  placeholder={t('contentPlaceholder')}
                  onChange={(event) => setForm((current) => ({ ...current, content: event.target.value }))}
                />
              </Field>

              <div className="grid gap-4 md:grid-cols-2">
                <Field>
                  <Label>{t('writingTypeLabel')}</Label>
                  <Description>{t('writingTypeHelp')}</Description>
                  <Input
                    name="writing_type"
                    value={form.writing_type ?? ''}
                    placeholder={t('writingTypePlaceholder')}
                    onChange={(event) => setForm((current) => ({ ...current, writing_type: event.target.value }))}
                  />
                </Field>
                <Field>
                  <Label>{t('formatLabel')}</Label>
                  <Description>{t('formatHelp')}</Description>
                  <Input
                    name="assignment_format"
                    value={form.assignment_format ?? ''}
                    placeholder={t('formatPlaceholder')}
                    onChange={(event) => setForm((current) => ({ ...current, assignment_format: event.target.value }))}
                  />
                </Field>
              </div>

              <div className="grid gap-4 md:grid-cols-[0.35fr_1fr]">
                <Field>
                  <Label>{t('languageLabel')}</Label>
                  <Description>{t('languageHelp')}</Description>
                  <Input
                    name="writing_language"
                    value={form.writing_language ?? ''}
                    placeholder={t('languagePlaceholder')}
                    onChange={(event) => setForm((current) => ({ ...current, writing_language: event.target.value }))}
                  />
                </Field>
                <Field>
                  <Label>{t('briefLabel')}</Label>
                  <Description>{t('briefHelp')}</Description>
                  <Input
                    name="coaching_brief"
                    value={form.coaching_brief ?? ''}
                    placeholder={t('briefPlaceholder')}
                    onChange={(event) => setForm((current) => ({ ...current, coaching_brief: event.target.value }))}
                  />
                </Field>
              </div>
            </FieldGroup>
          </Fieldset>

          {error ? (
            <div className="rounded-2xl border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700 dark:border-red-500/20 dark:bg-red-500/10 dark:text-red-200">
              {error}
            </div>
          ) : null}

          <div className="flex flex-wrap items-center justify-between gap-3 border-t border-stone-200 pt-4 dark:border-white/10">
            <div className="flex flex-wrap items-center gap-2 text-sm text-zinc-500 dark:text-zinc-400">
              <Badge color="zinc">{t('wordCount', { count: wordCount })}</Badge>
              {savedSession ? <Badge color="green">{t('saved')}</Badge> : null}
              {savedSession ? (
                <span>
                  {t('savedDraftLabel')}: {formatLocalDateTime(savedSession.updated_at) ?? savedSession.updated_at}
                </span>
              ) : form.content.trim() ? null : (
                <span>{t('emptyHint')}</span>
              )}
            </div>
            <div className="flex gap-3">
              <Button type="button" outline onClick={handleReset} disabled={submitting || saving}>
                {t('reset')}
              </Button>
              <Button type="button" outline onClick={handleSave} disabled={submitting || saving || form.content.trim() === ''}>
                {saving ? t('saving') : t('save')}
              </Button>
              <Button color="dark/zinc" type="submit" disabled={submitting || saving || form.content.trim() === ''}>
                {submitting ? t('reviewing') : t('review')}
              </Button>
            </div>
          </div>
        </form>
      </WorkspaceCard>

      {!currentReview ? (
        <WorkspaceCard>
          <CardHeader eyebrow={t('resultsEyebrow')} title={t('resultsTitle')} description={t('resultsEmpty')} />
        </WorkspaceCard>
      ) : (
        <div className="space-y-8">
          <div className="grid gap-8 xl:grid-cols-[0.95fr_1.35fr]">
            <WorkspaceCard>
              <CardHeader
                eyebrow={t('previousReviewsEyebrow')}
                title={t('previousReviewsTitle')}
                description={reviews.length > 0 ? undefined : t('previousReviewsEmpty')}
              />
              <div className="mt-4 space-y-3">
                {reviews.length === 0 ? <Text>{t('previousReviewsEmpty')}</Text> : null}
                {reviews.map((item, index) => (
                  <button
                    key={item.id}
                    type="button"
                    onClick={() => setSelectedReviewID(item.id)}
                    className={`w-full rounded-2xl border px-4 py-3 text-left transition ${item.id === currentReview.id ? 'border-zinc-900 bg-zinc-950 text-white dark:border-white dark:bg-white/10 dark:text-white' : 'border-stone-200 bg-stone-50 hover:border-stone-300 dark:border-white/10 dark:bg-white/5 dark:text-zinc-100'}`}
                  >
                    <div className="flex items-center justify-between gap-3">
                      <div className="text-sm font-semibold">
                        {index === 0 ? t('latestReview') : t('reviewNumber', { count: reviews.length - index })}
                      </div>
                      <Badge color={item.id === currentReview.id ? 'green' : 'zinc'}>{item.review.skill_scores.length}</Badge>
                    </div>
                    <Text className={`mt-2 text-sm ${item.id === currentReview.id ? 'text-zinc-200 dark:text-zinc-200' : ''}`}>
                      {formatLocalDateTime(item.created_at) ?? item.created_at}
                    </Text>
                    <Text className={`mt-2 text-sm ${item.id === currentReview.id ? 'text-zinc-200 dark:text-zinc-200' : ''}`}>
                      {item.review.summary}
                    </Text>
                  </button>
                ))}
              </div>
            </WorkspaceCard>

            <WorkspaceCard>
              <CardHeader eyebrow={t('aiDetailsEyebrow')} title={t('aiDetailsTitle')} description={currentReview.review.summary} />
              <div className="mt-4 space-y-4">
                <ProviderProvenance providerNote={currentReview.review.provider_note} />
              </div>
            </WorkspaceCard>
          </div>

          <div className="grid gap-8 xl:grid-cols-2">
            <WorkspaceCard>
              <CardHeader eyebrow={t('workedEyebrow')} title={t('workedTitle')} />
              <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
                {currentReview.review.strengths.length === 0 ? <li>{t('noStrengths')}</li> : null}
                {currentReview.review.strengths.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </WorkspaceCard>
            <WorkspaceCard>
              <CardHeader eyebrow={t('improveEyebrow')} title={t('improveTitle')} />
              <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
                {currentReview.review.weaknesses.length === 0 ? <li>{t('noWeaknesses')}</li> : null}
                {currentReview.review.weaknesses.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </WorkspaceCard>
          </div>

          <div className="grid gap-8 xl:grid-cols-[1.3fr_0.7fr]">
            <WorkspaceCard>
              <CardHeader eyebrow={t('lineNotesEyebrow')} title={t('lineNotesTitle')} description={t('lineNotesDescription')} />
              <div className="mt-4 space-y-3">
                {currentReview.review.annotations.length === 0 ? (
                  <Text>{t('noLineNotes')}</Text>
                ) : (
                  currentReview.review.annotations.map((item, index) => (
                    <div key={`${item.quote}-${index}`} className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                      <div className="text-sm font-semibold text-zinc-900 dark:text-white">{item.quote}</div>
                      <Text className="mt-2 text-sm">{item.comment}</Text>
                    </div>
                  ))
                )}
              </div>
            </WorkspaceCard>

            <div className="space-y-8">
              <WorkspaceCard>
                <CardHeader eyebrow={t('ratingsEyebrow')} title={t('ratingsTitle')} description={t('ratingsDescription')} />
                <div className="mt-4 space-y-3">
                  {currentReview.review.skill_scores.length === 0 ? (
                    <Text>{t('noRatings')}</Text>
                  ) : (
                    currentReview.review.skill_scores.map((item) => (
                      <div key={item.skill} className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                        <SkillScoreMeter score={item} />
                      </div>
                    ))
                  )}
                </div>
              </WorkspaceCard>

              <WorkspaceCard>
                <CardHeader eyebrow={t('signalsEyebrow')} title={t('signalsTitle')} />
                <div className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
                  {currentReview.review.analyzer_findings.length === 0 ? (
                    <Text>{t('noSignals')}</Text>
                  ) : (
                    currentReview.review.analyzer_findings.map((item) => <div key={item}>• {item}</div>)
                  )}
                </div>
              </WorkspaceCard>
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
