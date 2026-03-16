'use client'

import { useEffect, useState } from 'react'
import { Badge } from '@/components/badge'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { getDashboard } from '@/lib/api'
import type { Dashboard } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function ProgressView() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const state = await getDashboard()
        if (!cancelled) {
          setDashboard(state)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load progress')
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

  if (loading) {
    return <LoadingState label="Loading progress board…" />
  }
  if (error || !dashboard) {
    return <EmptyState title="Progress unavailable" body={error ?? 'Could not load progress board.'} actionHref="/" actionLabel="Back to assignment" />
  }

  return (
    <div className="space-y-8">
      <header>
        <Heading>Progress board</Heading>
        <Text className="mt-2 max-w-3xl">
          Active TGOs remain the primary measure. Completed TGOs and recurring slips help maintain skills already earned.
        </Text>
      </header>

      <div className="grid gap-8 lg:grid-cols-3">
        <WorkspaceCard>
          <Subheading>Active TGOs</Subheading>
          <div className="mt-4 flex flex-wrap gap-2">
            {dashboard.active_tgos.map((tgo) => (
              <Badge key={tgo.code} color="blue">
                {tgo.title}
              </Badge>
            ))}
          </div>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>Completed TGOs</Subheading>
          <div className="mt-4 flex flex-wrap gap-2">
            {dashboard.completed_tgos.length === 0 ? <Text>No completed TGOs yet.</Text> : null}
            {dashboard.completed_tgos.map((tgo) => (
              <Badge key={tgo.code} color="green">
                {tgo.title}
              </Badge>
            ))}
          </div>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>Unlocked next</Subheading>
          <div className="mt-4 flex flex-wrap gap-2">
            {dashboard.upcoming_tgos.length === 0 ? <Text>No new TGO unlocks yet.</Text> : null}
            {dashboard.upcoming_tgos.map((tgo) => (
              <Badge key={tgo.code} color="amber">
                {tgo.title}
              </Badge>
            ))}
          </div>
        </WorkspaceCard>
      </div>

      <div className="grid gap-8 xl:grid-cols-2">
        <WorkspaceCard>
          <Subheading>Strongest skills</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {dashboard.strongest_skills.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>Weakest skills</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {dashboard.weakest_skills.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>

      <div className="grid gap-8 xl:grid-cols-2">
        <WorkspaceCard>
          <Subheading>Recurring weaknesses</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {dashboard.recurring_weaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>Completed TGO slips</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {dashboard.recurring_completed_slips.length === 0 ? <li>No completed-skill regressions have been flagged recently.</li> : null}
            {dashboard.recurring_completed_slips.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>
    </div>
  )
}
