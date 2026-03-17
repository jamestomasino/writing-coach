'use client'

import { useEffect, useMemo, useState } from 'react'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { acceptAssignment, createAssignment, getDashboard, getSession } from '@/lib/api'
import type { Dashboard, Exercise } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'
import { useRouter } from 'next/navigation'

export function NewAssignmentView() {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [needsOnboarding, setNeedsOnboarding] = useState(false)
  const [selected, setSelected] = useState<string[]>([])
  const [preview, setPreview] = useState<Exercise | null>(null)
  const [generating, setGenerating] = useState(false)
  const [accepting, setAccepting] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const session = await getSession()
        if (!session.onboarding_complete) {
          if (!cancelled) {
            setNeedsOnboarding(true)
            setDashboard(null)
          }
          return
        }
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
  if (needsOnboarding) {
    return (
      <EmptyState
        title="Build your starter path first"
        body="You need an active skill map before you can choose skills for a new assignment."
        actionHref="/onboarding"
        actionLabel="Set starter path"
      />
    )
  }
  if (error && !dashboard) {
    return <EmptyState title="Could not load new assignment flow" body={error} actionHref="/" actionLabel="Back to assignment" />
  }
  if (!dashboard) {
    return <LoadingState />
  }

  return (
    <div className="space-y-8">
      <header>
        <Heading>New assignment</Heading>
        <Text className="mt-2 max-w-3xl">
          Choose exactly three skills for the next review. The assignment prompt itself comes from your track details, while this selection sets what the review will measure most closely.
        </Text>
      </header>

      {error ? <EmptyState title="Prompt generation issue" body={error} /> : null}

      <WorkspaceCard>
        <Subheading>Choose 3 review skills</Subheading>
        <Text className="mt-2">Mastered skills stay in the maintenance layer. This selection defines the primary review rubric for the next assignment, not the prompt premise.</Text>
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
                <p className={`mt-2 text-sm ${active ? 'text-stone-200' : 'text-zinc-600 dark:text-zinc-300'}`}>{tgo.description}</p>
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
                <Subheading>Generating assignment</Subheading>
                <Text className="mt-2">Building a new prompt from your track details and current coaching context. This usually takes a few seconds.</Text>
              </div>
            </div>
            <div className="rounded-2xl border border-cyan-300/70 bg-white/70 px-4 py-4 dark:border-cyan-400/20 dark:bg-black/10 lg:w-80">
              <Eyebrow tone="cyan">Working</Eyebrow>
              <div className="mt-3 space-y-2" role="status" aria-live="polite" aria-label="Assignment generation in progress">
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
          <div className="flex items-center justify-between gap-4">
            <Subheading>{preview.title}</Subheading>
            <div className="flex gap-2">
              <Button plain onClick={generate} disabled={generating}>
                {generating ? 'Refreshing…' : 'Refresh prompt'}
              </Button>
              <Button onClick={acceptPreview} color="dark/zinc" disabled={generating || accepting}>
                {accepting ? 'Accepting…' : 'Accept and continue'}
              </Button>
            </div>
          </div>
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
