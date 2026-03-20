'use client'

import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Callout } from '@/components/callout'
import { Eyebrow } from '@/components/eyebrow'
import { Subheading } from '@/components/heading'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { acceptAssignment, createAssignment, getDashboard, getSession } from '@/lib/api'
import { requiredSetupPath } from '@/lib/onboarding-funnel'
import type { Dashboard, Exercise } from '@/lib/types'
import { useRouter } from 'next/navigation'
import { useEffect, useMemo, useState } from 'react'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function NewAssignmentView() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [selected, setSelected] = useState<string[]>([])
  const [preview, setPreview] = useState<Exercise | null>(null)
  const [generating, setGenerating] = useState(false)
  const [accepting, setAccepting] = useState(false)
  const [setupFlow, setSetupFlow] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const session = await getSession()
        if (!session.authenticated) {
          router.replace('/about')
          return
        }
        const nextPath = requiredSetupPath(session, '/new-assignment')
        if (nextPath) {
          router.replace(nextPath)
          return
        }
        setSetupFlow(session.setup_step === 'needs_first_assignment')
        const state = await getDashboard()
        if (!cancelled) {
          setDashboard(state)
          setSelected(state.active_tgos.map((item) => item.code))
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load skill selection')
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
  }, [])

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
      const exercise = await createAssignment(selected)
      setPreview(exercise)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not generate assignment')
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
      setError(err instanceof Error ? err.message : 'Could not accept assignment')
    } finally {
      setAccepting(false)
    }
  }

  if (loading) {
    return <LoadingState label="Loading assignment setup…" />
  }
  if (error && !dashboard) {
    return <AppErrorState title="Could not load new assignment flow" error={error} />
  }
  if (!dashboard) {
    return <LoadingState />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={setupFlow ? 'Step 3 of 3 · First assignment' : 'Assignment setup'}
        title="New assignment"
        intro={
          setupFlow
            ? 'Choose the three skills you want your first assignment to emphasize. The prompt will come from your new track, and the review will score this selection most closely.'
            : 'Choose exactly three skills for the next review. The assignment prompt itself comes from the active track details, while this selection sets what the review will measure most closely.'
        }
      />

      {setupFlow ? (
        <Callout
          tone="active"
          eyebrow="Onboarding"
          title="Finish by generating your first assignment"
          body="Pick three skills, generate a prompt, and accept it. Once you do, the full writing workspace becomes your normal home."
        >
          <ul className="space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            <li>Keep the selection to exactly three skills.</li>
            <li>Generate the prompt, then refresh it if the first version misses the mark.</li>
            <li>Accept the prompt to enter the current assignment workspace.</li>
          </ul>
        </Callout>
      ) : null}

      {error ? <EmptyState title="Prompt generation issue" body={error} /> : null}

      <WorkspaceCard>
        <CardHeader
          eyebrow="Review focus"
          title="Choose 3 review skills"
          description="Mastered skills stay in the maintenance layer. This selection defines the primary review rubric for the next assignment, not the prompt premise."
        />
        <div className="mt-5 grid gap-4 lg:grid-cols-2">
          {selectable.map((tgo) => {
            const active = selected.includes(tgo.code)
            return (
              <button
                key={tgo.code}
                type="button"
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
          <Text>{selected.length} of 3 skills selected.</Text>
          <Button color="dark/zinc" onClick={generate} disabled={selected.length !== 3 || generating}>
            {generating ? 'Generating…' : 'Generate prompt'}
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
                <Eyebrow tone="cyan">Assignment generation</Eyebrow>
                <Subheading>Generating assignment</Subheading>
                <Text className="mt-2">
                  Building a new prompt from the active track details and current coaching context. This usually takes a
                  few seconds.
                </Text>
              </div>
            </div>
            <div className="rounded-2xl border border-cyan-300/70 bg-white/70 px-4 py-4 lg:w-80 dark:border-cyan-400/20 dark:bg-black/10">
              <Eyebrow tone="cyan">Working</Eyebrow>
              <div
                className="mt-3 space-y-2"
                role="status"
                aria-live="polite"
                aria-label="Assignment generation in progress"
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
            eyebrow="Generated prompt"
            title={preview.title}
            actions={
              <div className="flex gap-2">
                <Button plain onClick={generate} disabled={generating}>
                  {generating ? 'Refreshing…' : 'Refresh prompt'}
                </Button>
                <Button onClick={acceptPreview} color="dark/zinc" disabled={generating || accepting}>
                  {accepting ? 'Accepting…' : 'Accept and continue'}
                </Button>
              </div>
            }
          />
          <Text className="mt-3">{preview.brief}</Text>
          <div className="mt-5 grid gap-6 lg:grid-cols-2">
            <div>
              <p className="text-sm font-semibold text-zinc-900 dark:text-white">Constraints</p>
              <ul className="mt-2 space-y-2 text-sm text-zinc-600 dark:text-zinc-300">
                {preview.constraints.map((item) => (
                  <li key={item}>• {item}</li>
                ))}
              </ul>
            </div>
            <div>
              <p className="text-sm font-semibold text-zinc-900 dark:text-white">Success criteria</p>
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
