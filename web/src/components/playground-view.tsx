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
import { createPlaygroundReview } from '@/lib/api'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import type { PlaygroundReviewInput, Review } from '@/lib/types'
import { useTranslations } from 'next-intl'
import { FormEvent, useState } from 'react'

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

export function PlaygroundView() {
  const t = useTranslations('playgroundView')
  const { loading: sessionLoading, error: sessionError } = useRequiredAppSession('/playground')
  const [form, setForm] = useState<PlaygroundReviewInput>(initialForm)
  const [review, setReview] = useState<Review | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (submitting) {
      return
    }
    setSubmitting(true)
    setError(null)
    try {
      const nextReview = await createPlaygroundReview(form)
      setReview(nextReview)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('reviewError'))
    } finally {
      setSubmitting(false)
    }
  }

  if (sessionLoading) {
    return <LoadingState label={t('loading')} />
  }
  if (sessionError) {
    return <AppErrorState title={t('unavailableTitle')} error={sessionError} />
  }

  const wordCount = localWordCount(form.content ?? '')

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={t('eyebrow')}
        title={t('title')}
        intro={t('intro')}
        actions={
          <Button href="/about" plain>
            {t('openAbout')}
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

      <div className="grid gap-8 xl:grid-cols-[1.15fr_0.85fr]">
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
                {form.content.trim() ? null : <span>{t('emptyHint')}</span>}
              </div>
              <div className="flex gap-3">
                <Button
                  type="button"
                  outline
                  onClick={() => {
                    setForm(initialForm)
                    setError(null)
                  }}
                  disabled={submitting}
                >
                  {t('reset')}
                </Button>
                <Button color="dark/zinc" type="submit" disabled={submitting || form.content.trim() === ''}>
                  {submitting ? t('reviewing') : t('review')}
                </Button>
              </div>
            </div>
          </form>
        </WorkspaceCard>

        <WorkspaceCard className="border-stone-200/80 bg-linear-to-br from-stone-50 via-white to-sky-50 dark:border-white/10 dark:from-zinc-900 dark:via-zinc-900 dark:to-sky-950/30">
          <CardHeader eyebrow={t('tipsEyebrow')} title={t('tipsTitle')} />
          <div className="mt-4 space-y-4">
            <Text>{t('tipsBody1')}</Text>
            <Text>{t('tipsBody2')}</Text>
            <Text>{t('tipsBody3')}</Text>
          </div>
        </WorkspaceCard>
      </div>

      {!review ? (
        <WorkspaceCard>
          <CardHeader eyebrow={t('resultsEyebrow')} title={t('resultsTitle')} description={t('resultsEmpty')} />
        </WorkspaceCard>
      ) : (
        <div className="space-y-8">
          <WorkspaceCard>
            <CardHeader eyebrow={t('aiDetailsEyebrow')} title={t('aiDetailsTitle')} description={review.summary} />
            <div className="mt-4 space-y-4">
              <ProviderProvenance providerNote={review.provider_note} />
            </div>
          </WorkspaceCard>

          <div className="grid gap-8 xl:grid-cols-2">
            <WorkspaceCard>
              <CardHeader eyebrow={t('workedEyebrow')} title={t('workedTitle')} />
              <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
                {review.strengths.length === 0 ? <li>{t('noStrengths')}</li> : null}
                {review.strengths.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </WorkspaceCard>
            <WorkspaceCard>
              <CardHeader eyebrow={t('improveEyebrow')} title={t('improveTitle')} />
              <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
                {review.weaknesses.length === 0 ? <li>{t('noWeaknesses')}</li> : null}
                {review.weaknesses.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </WorkspaceCard>
          </div>

          <div className="grid gap-8 xl:grid-cols-[1.3fr_0.7fr]">
            <WorkspaceCard>
              <CardHeader eyebrow={t('lineNotesEyebrow')} title={t('lineNotesTitle')} description={t('lineNotesDescription')} />
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
                        {item.tgo_title || item.tgo_code ? <Badge color="blue">{item.tgo_title ?? item.tgo_code}</Badge> : null}
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
            </WorkspaceCard>

            <div className="space-y-8">
              <WorkspaceCard>
                <CardHeader eyebrow={t('ratingsEyebrow')} title={t('ratingsTitle')} description={t('ratingsDescription')} />
                <div className="mt-4 space-y-3">
                  {review.skill_scores.length === 0 ? (
                    <Text>{t('noRatings')}</Text>
                  ) : (
                    review.skill_scores.map((item) => (
                      <div
                        key={item.skill}
                        className="rounded-xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5"
                      >
                        <SkillScoreMeter score={item} />
                      </div>
                    ))
                  )}
                </div>
              </WorkspaceCard>

              <WorkspaceCard>
                <CardHeader eyebrow={t('signalsEyebrow')} title={t('signalsTitle')} />
                <div className="mt-4 space-y-3">
                  {review.analyzer_findings.length === 0 ? (
                    <Text>{t('noSignals')}</Text>
                  ) : (
                    review.analyzer_findings.map((item) => (
                      <div
                        key={item}
                        className="rounded-xl border border-stone-200 bg-white p-4 text-sm text-zinc-700 dark:border-white/10 dark:bg-zinc-950 dark:text-zinc-300"
                      >
                        {item}
                      </div>
                    ))
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
