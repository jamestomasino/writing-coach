'use client'

import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Callout } from '@/components/callout'
import { Eyebrow } from '@/components/eyebrow'
import { Subheading } from '@/components/heading'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { acceptAssignment, createAssignment, getAIJob, getDashboard } from '@/lib/api'
import type { Dashboard, Exercise } from '@/lib/types'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import { useTranslations } from 'next-intl'
import { useRouter } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function NewAssignmentView() {
  const t = useTranslations('newAssignmentView')
  const router = useRouter()
  const { session, loading: sessionLoading, error: sessionError } = useRequiredAppSession('/new-assignment')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [preview, setPreview] = useState<Exercise | null>(null)
  const [generating, setGenerating] = useState(false)
  const [accepting, setAccepting] = useState(false)
  const [setupFlow, setSetupFlow] = useState(false)

  async function waitForExercise(jobId: number) {
    for (let attempt = 0; attempt < 120; attempt += 1) {
      const job = await getAIJob(jobId)
      if (job.status === 'completed') {
        if (job.result?.exercise) {
          return job.result.exercise
        }
        throw new Error(t('generateError'))
      }
      if (job.status === 'failed') {
        throw new Error(job.last_error || t('generateError'))
      }
      await new Promise((resolve) => window.setTimeout(resolve, 1500))
    }
    throw new Error(t('generateError'))
  }

  useEffect(() => {
    let cancelled = false
    async function load() {
      if (!session) {
        return
      }
      try {
        setSetupFlow(session.setup_step === 'needs_first_assignment')
        const state = await getDashboard()
        if (!cancelled) {
          setDashboard(state)
          setSelected(state.active_tgos.map((item) => item.code))
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : t('loadError'))
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
  }, [router, session, t])

  const selectable = useMemo(() => {
    if (!dashboard) {
      return []
    }
    const byDisplayKey = new Map<string, Dashboard['active_tgos'][number]>()
    for (const tgo of dashboard.active_tgos) {
      const key = `${tgo.title.trim().toLowerCase()}::${tgo.description.trim().toLowerCase()}`
      byDisplayKey.set(key, tgo)
    }
    for (const tgo of dashboard.upcoming_tgos) {
      const key = `${tgo.title.trim().toLowerCase()}::${tgo.description.trim().toLowerCase()}`
      if (!byDisplayKey.has(key)) {
        byDisplayKey.set(key, tgo)
      }
    }
    return [...byDisplayKey.values()]
  }, [dashboard])

  function toggle(code: string) {
    setSelected((current) => {
      if (current.includes(code)) {
        if (current.length === 1) {
          return current
        }
        return current.filter((item) => item !== code)
      }
      if (current.length >= 3) {
        return current
      }
      return [...current, code]
    })
  }

  async function generate() {
    try {
      setGenerating(true)
      setError(null)
      setPreview(null)
      const job = await createAssignment(selected)
      const exercise = await waitForExercise(job.id)
      setPreview(exercise)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('generateError'))
    } finally {
      setGenerating(false)
    }
  }

  async function acceptPreview() {
    if (!preview) {
      return
    }
    try {
      setAccepting(true)
      setError(null)
      await acceptAssignment(preview)
      router.push('/')
    } catch (err) {
      setError(err instanceof Error ? err.message : t('acceptError'))
    } finally {
      setAccepting(false)
    }
  }

  if (sessionLoading || loading) {
    return <LoadingState label={t('loading')} />
  }
  if ((sessionError || error) && !dashboard) {
    return <AppErrorState title={t('flowErrorTitle')} error={sessionError ?? error ?? t('flowErrorTitle')} />
  }
  if (!dashboard) {
    return <LoadingState />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={setupFlow ? t('setupStepEyebrow') : t('setupEyebrow')}
        title={t('title')}
        intro={
          setupFlow
            ? t('setupIntro')
            : t('intro')
        }
      />

      {setupFlow ? (
        <Callout
          tone="active"
          eyebrow={t('calloutEyebrow')}
          title={t('calloutTitle')}
          body={t('calloutBody')}
        >
          <ul className="space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            <li>{t('calloutBullet1')}</li>
            <li>{t('calloutBullet2')}</li>
            <li>{t('calloutBullet3')}</li>
          </ul>
        </Callout>
      ) : null}

      {error ? <EmptyState title={t('issueTitle')} body={error} /> : null}

      <WorkspaceCard>
        <CardHeader
          eyebrow={t('focusEyebrow')}
          title={t('focusTitle')}
          description={t('focusDescription')}
        />
        <div className="mt-5 grid gap-4 lg:grid-cols-2">
          {selectable.map((tgo) => {
            const active = selected.includes(tgo.code)
            return (
              <button
                key={tgo.code}
                type="button"
                data-testid={`skill-option-${tgo.code}`}
                data-skill-code={tgo.code}
                onClick={() => toggle(tgo.code)}
                disabled={generating}
                className={`rounded-2xl border p-4 text-left transition ${
                  active
                    ? 'border-stone-800 bg-stone-900 text-white'
                    : 'border-stone-200 bg-stone-50 text-zinc-900 hover:border-stone-400 dark:border-white/10 dark:bg-white/5 dark:text-white'
                } ${generating ? 'cursor-not-allowed opacity-60' : ''}`}
              >
                <div className="flex items-center justify-between gap-4">
                  <span className="font-semibold">{tgo.title}</span>
                  <Badge color={active ? 'amber' : 'zinc'}>{tgo.stage}</Badge>
                </div>
                <p className={`mt-2 text-sm ${active ? 'text-stone-200' : 'text-zinc-600 dark:text-zinc-300'}`}>
                  {tgo.description}
                </p>
              </button>
            )
          })}
        </div>
        <div className="mt-5 flex items-center justify-between gap-3">
          <Text>{t('selectedCount', { count: selected.length })}</Text>
          <Button
            color="dark/zinc"
            onClick={generate}
            disabled={selected.length !== 3 || generating}
            data-testid="generate-assignment-button"
          >
            {generating ? t('generating') : t('generate')}
          </Button>
        </div>
      </WorkspaceCard>

      {generating ? (
        <WorkspaceCard className="border-cyan-200 bg-cyan-50 dark:border-cyan-500/20 dark:bg-cyan-500/10">
          <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
            <div className="flex items-start gap-4">
              <div
                className="mt-0.5 inline-flex size-10 shrink-0 items-center justify-center rounded-full border border-cyan-300 bg-white/80 dark:border-cyan-400/20 dark:bg-black/10"
                aria-hidden="true"
              >
                <span className="size-5 animate-spin rounded-full border-2 border-cyan-700/25 border-t-cyan-700 dark:border-cyan-200/25 dark:border-t-cyan-200" />
              </div>
              <div>
                <Eyebrow tone="cyan">{t('generationEyebrow')}</Eyebrow>
                <Subheading>{t('generationTitle')}</Subheading>
                <Text className="mt-2">{t('generationBody')}</Text>
              </div>
            </div>
            <div className="rounded-2xl border border-cyan-300/70 bg-white/70 px-4 py-4 lg:w-80 dark:border-cyan-400/20 dark:bg-black/10">
              <Eyebrow tone="cyan">Working</Eyebrow>
              <div
                className="mt-3 space-y-2"
                role="status"
                aria-live="polite"
                aria-label={t('progressLabel')}
              >
                <div className="h-2 w-full animate-pulse rounded-full bg-cyan-200/80 dark:bg-cyan-200/15" />
                <div className="h-2 w-5/6 animate-pulse rounded-full bg-cyan-200/70 [animation-delay:120ms] dark:bg-cyan-200/12" />
                <div className="h-2 w-2/3 animate-pulse rounded-full bg-cyan-200/60 [animation-delay:240ms] dark:bg-cyan-200/10" />
              </div>
            </div>
          </div>
        </WorkspaceCard>
      ) : null}

      {preview ? (
        <WorkspaceCard>
          <CardHeader
            eyebrow={t('draftEyebrow')}
            title={preview.title}
            actions={
              <div className="flex gap-2">
                <Button plain onClick={generate} disabled={generating}>
                  {generating ? t('tryingAnother') : t('tryAnother')}
                </Button>
                <Button
                  onClick={acceptPreview}
                  color="dark/zinc"
                  disabled={generating || accepting}
                  data-testid="accept-assignment-button"
                >
                  {accepting ? t('opening') : t('useAssignment')}
                </Button>
              </div>
            }
          />
          <Text className="mt-3">{preview.brief}</Text>
          <div className="mt-5 grid gap-6 lg:grid-cols-2">
            <div>
              <p className="text-sm font-semibold text-zinc-900 dark:text-white">{t('constraints')}</p>
              <ul className="mt-2 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                {preview.constraints.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
            <div>
              <p className="text-sm font-semibold text-zinc-900 dark:text-white">{t('successCriteria')}</p>
              <ul className="mt-2 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                {preview.success_criteria.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
          </div>
        </WorkspaceCard>
      ) : null}
    </div>
  )
}
