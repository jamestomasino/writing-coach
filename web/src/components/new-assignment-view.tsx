'use client'

import { useEffect, useMemo, useState } from 'react'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { createAssignment, getDashboard, getSession } from '@/lib/api'
import type { Dashboard, Exercise } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function NewAssignmentView() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [needsOnboarding, setNeedsOnboarding] = useState(false)
  const [selected, setSelected] = useState<string[]>([])
  const [preview, setPreview] = useState<Exercise | null>(null)
  const [generating, setGenerating] = useState(false)

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
      const exercise = await createAssignment(selected)
      setPreview(exercise)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not generate assignment')
    } finally {
      setGenerating(false)
    }
  }

  if (loading) {
    return <LoadingState label="Loading skill selection…" />
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
          Choose exactly three unlocked skills. The prompt generator will build the next assignment around this set and your skill map’s coaching goals.
        </Text>
      </header>

      {error ? <EmptyState title="Prompt generation issue" body={error} /> : null}

      <WorkspaceCard>
        <Subheading>Choose 3 skills</Subheading>
        <Text className="mt-2">Mastered skills stay in the maintenance layer. This selection defines the primary review rubric for the next assignment.</Text>
        <div className="mt-5 grid gap-4 lg:grid-cols-2">
          {selectable.map((tgo) => {
            const active = selected.includes(tgo.code)
            return (
              <button
                key={tgo.code}
                type="button"
                onClick={() => toggle(tgo.code)}
                className={`rounded-2xl border p-4 text-left transition ${
                  active
                    ? 'border-stone-800 bg-stone-900 text-white'
                    : 'border-stone-200 bg-stone-50 text-zinc-900 hover:border-stone-400 dark:border-white/10 dark:bg-white/5 dark:text-white'
                }`}
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

      {preview ? (
        <WorkspaceCard>
          <div className="flex items-center justify-between gap-4">
            <Subheading>{preview.title}</Subheading>
            <div className="flex gap-2">
              <Button plain onClick={generate} disabled={generating}>
                Refresh prompt
              </Button>
              <Button href="/" color="dark/zinc">
                Accept and continue
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
