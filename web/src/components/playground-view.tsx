'use client'

import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Description, Field, FieldGroup, Fieldset, Label } from '@/components/fieldset'
import { Input } from '@/components/input'
import { PageHeader } from '@/components/page-header'
import { ProviderProvenance } from '@/components/provider-provenance'
import { AppErrorState, LoadingState, TaskProgressState } from '@/components/status-state'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/table'
import { Text } from '@/components/text'
import { Textarea } from '@/components/textarea'
import { WorkspaceCard } from '@/components/workspace-card'
import { formatLocalDateTime } from '@/lib/datetime'
import {
  createPlaygroundDraft,
  createPlaygroundSession,
  createPlaygroundSessionReview,
  getAIJob,
  getPlaygroundSession,
  getPlaygroundSessionDrafts,
  getPlaygroundSessionReviews,
  updatePlaygroundSession,
} from '@/lib/api'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import type { AIJob, PlaygroundDraft, PlaygroundReview, PlaygroundReviewInput, PlaygroundSession } from '@/lib/types'
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

type AnalyzerFindingView = {
  analyzer: string
  category: string
  severity: string
  message: string
}

type AnalyzerMetricView = {
  key: string
  value: number
}

type GroupedFindingView = AnalyzerFindingView & {
  count: number
  indexes: number[]
}

type ToolReportView = {
  id: string
  label: string
  findings: AnalyzerFindingView[]
  metrics: AnalyzerMetricView[]
  warnings: string[]
}

type AnalyzerReportPayload = {
  findings?: unknown
  Findings?: unknown
  metrics?: unknown
  Metrics?: unknown
  warnings?: unknown
  Warnings?: unknown
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

function normalizeToolID(value: string) {
  const normalized = value.trim().toLowerCase().replace(/[\s_-]+/g, '')
  if (normalized === 'languagetool') {
    return 'languagetool'
  }
  if (normalized === 'nlp') {
    return 'nlp'
  }
  if (normalized === 'vale') {
    return 'vale'
  }
  if (normalized === 'heuristic') {
    return 'heuristic'
  }
  return 'other'
}

function toolLabel(toolID: string) {
  switch (toolID) {
    case 'heuristic':
      return 'Heuristic'
    case 'vale':
      return 'Vale'
    case 'languagetool':
      return 'LanguageTool'
    case 'nlp':
      return 'NLP'
    default:
      return 'Other'
  }
}

function metricToolID(metricKey: string) {
  if (metricKey.startsWith('nlp_')) {
    return 'nlp'
  }
  if (metricKey.startsWith('languagetool_')) {
    return 'languagetool'
  }
  if (metricKey.startsWith('vale_')) {
    return 'vale'
  }
  if (metricKey.startsWith('heuristic_')) {
    return 'heuristic'
  }
  if (metricKey === 'word_count' || metricKey === 'avg_sentence_length' || metricKey === 'adverb_count') {
    return 'heuristic'
  }
  return 'other'
}

function warningToolID(warning: string) {
  const normalized = warning.trim().toLowerCase()
  for (const tool of ['heuristic', 'vale', 'languagetool', 'nlp']) {
    if (normalized.startsWith(tool + ':') || normalized.startsWith(tool + ' ')) {
      return tool
    }
  }
  return 'other'
}

function normalizeSeverity(value: string) {
  const normalized = value.trim().toLowerCase()
  if (normalized === 'error') {
    return 'error'
  }
  if (normalized === 'warning' || normalized === 'warn') {
    return 'warning'
  }
  return 'note'
}

function formatMetricLabel(metricKey: string) {
  return metricKey.replaceAll('_', ' ')
}

function parseAnalyzerReport(value: unknown): ToolReportView[] {
  if (!value || typeof value !== 'object') {
    return []
  }
  const payload = value as AnalyzerReportPayload
  const tools = new Map<string, ToolReportView>()
  const ensureTool = (id: string) => {
    const key = id || 'other'
    const existing = tools.get(key)
    if (existing) {
      return existing
    }
    const created: ToolReportView = {
      id: key,
      label: toolLabel(key),
      findings: [],
      metrics: [],
      warnings: [],
    }
    tools.set(key, created)
    return created
  }

  const findings = Array.isArray(payload.findings)
    ? payload.findings
    : Array.isArray(payload.Findings)
      ? payload.Findings
      : []
  for (const item of findings) {
    if (!item || typeof item !== 'object') {
      continue
    }
    const finding = item as {
      analyzer?: unknown
      Analyzer?: unknown
      category?: unknown
      Category?: unknown
      severity?: unknown
      Severity?: unknown
      message?: unknown
      Message?: unknown
    }
    const analyzerRaw = typeof finding.analyzer === 'string'
      ? finding.analyzer
      : typeof finding.Analyzer === 'string'
        ? finding.Analyzer
        : ''
    const categoryRaw = typeof finding.category === 'string'
      ? finding.category
      : typeof finding.Category === 'string'
        ? finding.Category
        : ''
    const severityRaw = typeof finding.severity === 'string'
      ? finding.severity
      : typeof finding.Severity === 'string'
        ? finding.Severity
        : ''
    const messageRaw = typeof finding.message === 'string'
      ? finding.message
      : typeof finding.Message === 'string'
        ? finding.Message
        : ''
    const analyzer = analyzerRaw.trim()
    const toolID = normalizeToolID(analyzer)
    ensureTool(toolID).findings.push({
      analyzer: analyzer || toolLabel(toolID),
      category: categoryRaw.trim() || 'general',
      severity: normalizeSeverity(severityRaw),
      message: messageRaw.trim(),
    })
  }

  const metricsRaw = payload.metrics && typeof payload.metrics === 'object'
    ? payload.metrics
    : payload.Metrics && typeof payload.Metrics === 'object'
      ? payload.Metrics
      : null
  if (metricsRaw) {
    for (const [key, raw] of Object.entries(metricsRaw as Record<string, unknown>)) {
      if (!Number.isFinite(raw)) {
        continue
      }
      const metric: AnalyzerMetricView = { key, value: Number(raw) }
      ensureTool(metricToolID(key)).metrics.push(metric)
    }
  }

  const warnings = Array.isArray(payload.warnings)
    ? payload.warnings
    : Array.isArray(payload.Warnings)
      ? payload.Warnings
      : []
  for (const warning of warnings) {
    if (typeof warning !== 'string' || warning.trim() === '') {
      continue
    }
    ensureTool(warningToolID(warning)).warnings.push(warning.trim())
  }

  return Array.from(tools.values())
    .filter((item) => item.findings.length > 0 || item.metrics.length > 0 || item.warnings.length > 0)
    .sort((a, b) => a.label.localeCompare(b.label))
}

function severityCount(findings: AnalyzerFindingView[], severity: string) {
  return findings.filter((item) => item.severity === severity).length
}

function severityTone(severity: string) {
  switch (severity) {
    case 'error':
      return 'bg-rose-500'
    case 'warning':
      return 'bg-amber-500'
    default:
      return 'bg-sky-500'
  }
}

function severityRank(severity: string) {
  switch (severity) {
    case 'error':
      return 0
    case 'warning':
      return 1
    default:
      return 2
  }
}

function groupFindings(findings: AnalyzerFindingView[]): GroupedFindingView[] {
  const grouped = new Map<string, GroupedFindingView>()
  for (const [index, finding] of findings.entries()) {
    const key = `${finding.severity}\u0000${finding.category}\u0000${finding.message}`
    const existing = grouped.get(key)
    if (existing) {
      existing.count += 1
      existing.indexes.push(index + 1)
      continue
    }
    grouped.set(key, {
      ...finding,
      count: 1,
      indexes: [index + 1],
    })
  }
  return Array.from(grouped.values()).sort((a, b) => {
    const severityDelta = severityRank(a.severity) - severityRank(b.severity)
    if (severityDelta !== 0) {
      return severityDelta
    }
    if (b.count !== a.count) {
      return b.count - a.count
    }
    return a.indexes[0] - b.indexes[0]
  })
}

export function PlaygroundView({ sessionId }: { sessionId?: number }) {
  const t = useTranslations('playgroundView')
  const router = useRouter()
  const { loading: sessionLoading, error: sessionError } = useRequiredAppSession(sessionId ? `/playground/${sessionId}` : '/playground')
  const [form, setForm] = useState<PlaygroundReviewInput>(initialForm)
  const [savedSession, setSavedSession] = useState<PlaygroundSession | null>(null)
  const [drafts, setDrafts] = useState<PlaygroundDraft[]>([])
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
        const [nextSession, nextDrafts, nextReviews] = await Promise.all([
          getPlaygroundSession(nextSessionID),
          getPlaygroundSessionDrafts(nextSessionID),
          getPlaygroundSessionReviews(nextSessionID),
        ])
        if (cancelled) {
          return
        }
        setSavedSession(nextSession)
        setForm(sessionToForm(nextSession))
        setDrafts(nextDrafts)
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
  const analyzerTools = useMemo(
    () => parseAnalyzerReport(currentReview?.review.artifacts?.analyzer_report),
    [currentReview],
  )
  const analyzerTotals = useMemo(
    () =>
      analyzerTools.reduce(
        (totals, tool) => ({
          findings: totals.findings + tool.findings.length,
          metrics: totals.metrics + tool.metrics.length,
          warnings: totals.warnings + tool.warnings.length,
        }),
        { findings: 0, metrics: 0, warnings: 0 },
      ),
    [analyzerTools],
  )

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
    const [nextSession, nextDrafts, nextReviews] = await Promise.all([
      getPlaygroundSession(nextSessionID),
      getPlaygroundSessionDrafts(nextSessionID),
      getPlaygroundSessionReviews(nextSessionID),
    ])
    setSavedSession(nextSession)
    setDrafts(nextDrafts)
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
      const nextSession = await persistSession()
      await createPlaygroundDraft(nextSession.id)
      await refreshSessionReviews(nextSession.id)
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
                  {t('savedDraftLabel')}: {savedSession.draft_count}
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
                eyebrow={t('draftsEyebrow')}
                title={t('draftsTitle')}
                description={drafts.length > 0 ? undefined : t('draftsEmpty')}
              />
              <div className="mt-4 space-y-3">
                {drafts.length === 0 ? <Text>{t('draftsEmpty')}</Text> : null}
                {drafts.map((item, index) => (
                  <div key={item.id} className="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                    <div className="flex items-center justify-between gap-3">
                      <div className="text-sm font-semibold">{t('draftNumber', { count: drafts.length - index })}</div>
                      <Badge color="zinc">{item.word_count}</Badge>
                    </div>
                    <Text className="mt-2 text-sm">{formatLocalDateTime(item.created_at) ?? item.created_at}</Text>
                    <Text className="mt-2 line-clamp-3 text-sm">{item.content}</Text>
                    <div className="mt-3">
                      <Button
                        type="button"
                        outline
                        onClick={() => setForm((current) => ({ ...current, content: item.content }))}
                      >
                        {t('loadDraft')}
                      </Button>
                    </div>
                  </div>
                ))}
              </div>
            </WorkspaceCard>

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
                <ProviderProvenance providerNote={currentReview.review.provider_note} kind="feedback" />
              </div>
            </WorkspaceCard>
          </div>

          <div className="grid gap-8 xl:grid-cols-2">
            {currentReview.review.artifacts?.comparison ? (
              <WorkspaceCard>
                <CardHeader eyebrow={t('compareEyebrow')} title={t('compareTitle')} description={currentReview.review.artifacts.comparison.summary} />
                <div className="mt-4 grid gap-4 md:grid-cols-3">
                  <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                    <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('wordDelta')}</div>
                    <Text className="mt-2 text-sm">{currentReview.review.artifacts.comparison.word_delta}</Text>
                  </div>
                  <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                    <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('addressed')}</div>
                    <Text className="mt-2 text-sm">{(currentReview.review.artifacts.comparison.addressed_weaknesses ?? []).length}</Text>
                  </div>
                  <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                    <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('persisting')}</div>
                    <Text className="mt-2 text-sm">{(currentReview.review.artifacts.comparison.persisting_weaknesses ?? []).length}</Text>
                  </div>
                </div>
              </WorkspaceCard>
            ) : null}
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

          <WorkspaceCard>
            <CardHeader eyebrow={t('technicalEyebrow')} title={t('technicalTitle')} description={t('technicalDescription')} />
            <details className="mt-4 rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
              <summary className="cursor-pointer text-sm font-semibold text-zinc-900 dark:text-white">{t('technicalSummary')}</summary>
              {analyzerTools.length === 0 ? (
                <Text className="mt-3 text-sm">{t('technicalEmpty')}</Text>
              ) : (
                <div className="mt-3 space-y-4">
                  <div className="grid gap-3 md:grid-cols-4">
                    <div className="rounded-lg border border-stone-200 bg-white px-3 py-2 text-xs dark:border-white/10 dark:bg-zinc-950">
                      <div className="font-semibold text-zinc-900 dark:text-white">{t('technicalTools')}</div>
                      <div className="mt-1 text-zinc-700 dark:text-zinc-300">{analyzerTools.length}</div>
                    </div>
                    <div className="rounded-lg border border-stone-200 bg-white px-3 py-2 text-xs dark:border-white/10 dark:bg-zinc-950">
                      <div className="font-semibold text-zinc-900 dark:text-white">{t('technicalFindings')}</div>
                      <div className="mt-1 text-zinc-700 dark:text-zinc-300">{analyzerTotals.findings}</div>
                    </div>
                    <div className="rounded-lg border border-stone-200 bg-white px-3 py-2 text-xs dark:border-white/10 dark:bg-zinc-950">
                      <div className="font-semibold text-zinc-900 dark:text-white">{t('technicalMetrics')}</div>
                      <div className="mt-1 text-zinc-700 dark:text-zinc-300">{analyzerTotals.metrics}</div>
                    </div>
                    <div className="rounded-lg border border-stone-200 bg-white px-3 py-2 text-xs dark:border-white/10 dark:bg-zinc-950">
                      <div className="font-semibold text-zinc-900 dark:text-white">{t('technicalWarnings')}</div>
                      <div className="mt-1 text-zinc-700 dark:text-zinc-300">{analyzerTotals.warnings}</div>
                    </div>
                  </div>

                  {analyzerTools.map((tool) => {
                    const errors = severityCount(tool.findings, 'error')
                    const warnings = severityCount(tool.findings, 'warning')
                    const notes = severityCount(tool.findings, 'note')
                    const maxSeverity = Math.max(errors, warnings, notes, 1)
                    const groupedFindings = groupFindings(tool.findings)

                    return (
                      <div key={tool.id} className="rounded-lg border border-stone-200 bg-white p-3 dark:border-white/10 dark:bg-zinc-950">
                        <div className="mb-3 flex flex-wrap items-center gap-2">
                          <div className="text-sm font-semibold text-zinc-900 dark:text-white">{tool.label}</div>
                          <Badge color="zinc">{t('technicalFindingsBadge', { count: tool.findings.length })}</Badge>
                          <Badge color="zinc">{t('technicalMetricsBadge', { count: tool.metrics.length })}</Badge>
                          <Badge color="zinc">{t('technicalWarningsBadge', { count: tool.warnings.length })}</Badge>
                        </div>

                        {tool.findings.length > 0 ? (
                          <div className="mb-4 space-y-2">
                            {[
                              { id: 'error', count: errors },
                              { id: 'warning', count: warnings },
                              { id: 'note', count: notes },
                            ].map((item) => (
                              <div key={item.id} className="grid grid-cols-[5.5rem_1fr_2.2rem] items-center gap-2 text-xs">
                                <div className="font-medium capitalize text-zinc-700 dark:text-zinc-300">{item.id}</div>
                                <div className="h-2 rounded-full bg-stone-200 dark:bg-white/10">
                                  <div
                                    className={`h-2 rounded-full ${severityTone(item.id)}`}
                                    style={{ width: `${(item.count / maxSeverity) * 100}%` }}
                                  />
                                </div>
                                <div className="text-right text-zinc-600 dark:text-zinc-400">{item.count}</div>
                              </div>
                            ))}
                          </div>
                        ) : null}

                        {tool.metrics.length > 0 ? (
                          <div className="mb-4">
                            <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-zinc-500 dark:text-zinc-400">{t('technicalMetricsTable')}</div>
                            <Table dense striped>
                              <TableHead>
                                <TableRow>
                                  <TableHeader>{t('technicalMetric')}</TableHeader>
                                  <TableHeader>{t('technicalValue')}</TableHeader>
                                </TableRow>
                              </TableHead>
                              <TableBody>
                                {tool.metrics
                                  .slice()
                                  .sort((a, b) => a.key.localeCompare(b.key))
                                  .map((metric) => (
                                    <TableRow key={metric.key}>
                                      <TableCell className="capitalize">{formatMetricLabel(metric.key)}</TableCell>
                                      <TableCell>{metric.value}</TableCell>
                                    </TableRow>
                                  ))}
                              </TableBody>
                            </Table>
                          </div>
                        ) : null}

                        {tool.findings.length > 0 ? (
                          <div className="mb-4">
                            <div className="mb-2 flex flex-wrap items-center justify-between gap-2">
                              <div className="text-xs font-semibold uppercase tracking-wide text-zinc-500 dark:text-zinc-400">{t('technicalFindingsTable')}</div>
                              <div className="text-xs text-zinc-600 dark:text-zinc-400">
                                {t('technicalCoverageHint', { raw: tool.findings.length, unique: groupedFindings.length })}
                              </div>
                            </div>
                            <Table dense striped>
                              <TableHead>
                                <TableRow>
                                  <TableHeader>{t('technicalSeverity')}</TableHeader>
                                  <TableHeader>{t('technicalCategory')}</TableHeader>
                                  <TableHeader>{t('technicalMessage')}</TableHeader>
                                  <TableHeader>{t('technicalCount')}</TableHeader>
                                  <TableHeader>{t('technicalIndexes')}</TableHeader>
                                </TableRow>
                              </TableHead>
                              <TableBody>
                                {groupedFindings.map((finding, index) => (
                                  <TableRow key={`${tool.id}-${index}-${finding.severity}-${finding.category}-${finding.message}`}>
                                    <TableCell className="capitalize">{finding.severity}</TableCell>
                                    <TableCell>{finding.category || 'general'}</TableCell>
                                    <TableCell className="whitespace-normal">{finding.message || '-'}</TableCell>
                                    <TableCell>{finding.count}</TableCell>
                                    <TableCell>{finding.indexes.join(', ')}</TableCell>
                                  </TableRow>
                                ))}
                              </TableBody>
                            </Table>
                          </div>
                        ) : null}

                        {tool.warnings.length > 0 ? (
                          <div>
                            <div className="mb-2 text-xs font-semibold uppercase tracking-wide text-zinc-500 dark:text-zinc-400">{t('technicalWarningsTable')}</div>
                            <ul className="space-y-1 text-xs text-zinc-700 dark:text-zinc-300">
                              {tool.warnings.map((warning) => (
                                <li key={warning}>• {warning}</li>
                              ))}
                            </ul>
                          </div>
                        ) : null}
                      </div>
                    )
                  })}
                </div>
              )}
            </details>
          </WorkspaceCard>
        </div>
      )}
    </div>
  )
}
