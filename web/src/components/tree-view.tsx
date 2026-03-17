'use client'

import { useEffect, useMemo, useState } from 'react'
import { ArrowRightIcon } from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { getDashboard, getOnboarding, getSession, getTree } from '@/lib/api'
import type { Dashboard, OnboardingState, Tree } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

type TreeNodeStatus = 'active' | 'completed' | 'unlocked' | 'locked'

type TreeNode = {
  code: string
  title: string
  description: string
  stage: string
  stage_order: number
  prerequisites?: string[]
  mastery_hint?: string
  status: TreeNodeStatus
  unlocks: string[]
}

function statusLabel(status: TreeNodeStatus) {
  switch (status) {
    case 'active':
      return 'Active'
    case 'completed':
      return 'Completed'
    case 'unlocked':
      return 'Unlocked'
    default:
      return 'Locked'
  }
}

function statusColor(status: TreeNodeStatus) {
  switch (status) {
    case 'active':
      return 'blue' as const
    case 'completed':
      return 'green' as const
    case 'unlocked':
      return 'amber' as const
    default:
      return 'zinc' as const
  }
}

function stageTitle(stage: string) {
  return stage.replace(/-/g, ' ')
}

export function TreeView() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tree, setTree] = useState<Tree | null>(null)
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [onboarding, setOnboarding] = useState<OnboardingState | null>(null)

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const session = await getSession()
        if (!session.authenticated) {
          throw new Error('Sign in to inspect your skill map')
        }
        if (!session.onboarding_complete || !session.active_tree_slug) {
          if (!cancelled) {
            setOnboarding({ onboarding_complete: false })
          }
          return
        }

        const [treeData, dashboardData, onboardingData] = await Promise.all([
          getTree(session.active_tree_slug),
          getDashboard(),
          getOnboarding(),
        ])

        if (!cancelled) {
          setTree(treeData)
          setDashboard(dashboardData)
          setOnboarding(onboardingData)
          setError(null)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load skill map')
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

  const stages = useMemo(() => {
    if (!tree || !dashboard) {
      return []
    }
    const treeTGOs = tree.tgos ?? []
    const activeTGOs = dashboard.active_tgos ?? []
    const completedTGOs = dashboard.completed_tgos ?? []
    const upcomingTGOs = dashboard.upcoming_tgos ?? []
    const active = new Set(activeTGOs.map((tgo) => tgo.code))
    const completed = new Set(completedTGOs.map((tgo) => tgo.code))
    const unlocked = new Set(upcomingTGOs.map((tgo) => tgo.code))
    const titleByCode = new Map(treeTGOs.map((tgo) => [tgo.code, tgo.title]))
    const unlocks = new Map<string, string[]>()
    for (const tgo of treeTGOs) {
      for (const prerequisite of tgo.prerequisites ?? []) {
        const next = unlocks.get(prerequisite) ?? []
        next.push(tgo.title)
        unlocks.set(prerequisite, next)
      }
    }
    const groups = new Map<string, { order: number; title: string; items: TreeNode[] }>()

    for (const tgo of treeTGOs) {
      const status: TreeNodeStatus = active.has(tgo.code)
        ? 'active'
        : completed.has(tgo.code)
          ? 'completed'
          : unlocked.has(tgo.code)
            ? 'unlocked'
            : 'locked'
      const key = tgo.stage
      if (!groups.has(key)) {
        groups.set(key, {
          order: tgo.stage_order,
          title: stageTitle(tgo.stage),
          items: [],
        })
      }
      const group = groups.get(key)
      if (!group) {
        continue
      }
      group.order = Math.min(group.order, tgo.stage_order)
      group.items.push({
        ...tgo,
        prerequisites: (tgo.prerequisites ?? []).map((prerequisite) => titleByCode.get(prerequisite) ?? prerequisite),
        status,
        unlocks: unlocks.get(tgo.code) ?? [],
      })
    }

    return [...groups.values()]
      .sort((a, b) => a.order - b.order)
      .map((group) => ({
        ...group,
        items: group.items.sort((a, b) => a.title.localeCompare(b.title)),
      }))
  }, [dashboard, tree])

  if (loading) {
    return <LoadingState label="Loading skill map…" />
  }
  if (!onboarding?.onboarding_complete) {
    return (
      <EmptyState
        title="Build your track first"
        body="You need to complete the starter path setup before the app can show your skill map."
        actionHref="/onboarding"
        actionLabel="Open starter path"
      />
    )
  }
  if (error || !tree || !dashboard) {
    return <EmptyState title="Tree unavailable" body={error ?? 'Could not load the current tree.'} actionHref="/" actionLabel="Back to assignment" />
  }

  const activeTGOs = dashboard.active_tgos ?? []
  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <Heading>Skill map</Heading>
          <Text className="mt-2 max-w-3xl">
            This view shows how the tree unlocks. Each card lists what must come first and what it opens next, so you can see the path through the track.
          </Text>
        </div>
        <div className="flex flex-wrap gap-2 text-sm text-zinc-600 dark:text-zinc-300">
          <Badge color="blue">{activeTGOs.length} active</Badge>
          <Badge color="green">{(dashboard.completed_tgos ?? []).length} mastered</Badge>
          <Badge color="amber">{(dashboard.upcoming_tgos ?? []).length} unlocked next</Badge>
        </div>
      </header>

      <WorkspaceCard>
        <div className="flex flex-wrap items-center gap-3">
          <Badge color="blue">Active</Badge>
          <Text className="text-sm">Being measured right now.</Text>
          <Badge color="green">Completed</Badge>
          <Text className="text-sm">Mastered and kept in lighter maintenance.</Text>
          <Badge color="amber">Unlocked</Badge>
          <Text className="text-sm">Ready to be activated next.</Text>
          <Badge color="zinc">Locked</Badge>
          <Text className="text-sm">Still waiting on prerequisites.</Text>
        </div>
      </WorkspaceCard>

      <div className="grid gap-6 xl:grid-cols-4">
        {stages.map((stage, index) => (
          <WorkspaceCard key={`${stage.order}-${stage.title}`} className="overflow-hidden">
            <div className="flex items-center justify-between gap-3">
              <div>
                <div className="text-xs font-semibold uppercase tracking-[0.18em] text-zinc-500 dark:text-zinc-400">
                  Stage {index + 1}
                </div>
                <Subheading className="mt-2">{stage.title}</Subheading>
              </div>
              <Badge color="zinc">{stage.items.length}</Badge>
            </div>
            <div className="mt-5 space-y-4">
              {stage.items.map((tgo) => (
                <article
                  key={tgo.code}
                  className="rounded-2xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5"
                >
                  <div className="flex items-start justify-between gap-4">
                    <div>
                      <div className="text-sm font-semibold text-zinc-950 dark:text-white">{tgo.title}</div>
                      <Text className="mt-2 text-sm">{tgo.description}</Text>
                    </div>
                    <Badge color={statusColor(tgo.status)}>{statusLabel(tgo.status)}</Badge>
                  </div>
                  {tgo.prerequisites && tgo.prerequisites.length > 0 ? (
                    <div className="mt-4">
                      <div className="text-[11px] font-semibold uppercase tracking-[0.16em] text-zinc-500 dark:text-zinc-400">
                        Requires
                      </div>
                      <div className="mt-2 flex flex-wrap gap-2">
                        {tgo.prerequisites.map((prerequisite) => (
                          <Badge key={prerequisite} color="zinc">
                            {prerequisite}
                          </Badge>
                        ))}
                      </div>
                    </div>
                  ) : null}
                  {tgo.unlocks.length > 0 ? (
                    <div className="mt-4">
                      <div className="flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.16em] text-zinc-500 dark:text-zinc-400">
                        <ArrowRightIcon className="size-3" />
                        Unlocks
                      </div>
                      <div className="mt-2 flex flex-wrap gap-2">
                        {tgo.unlocks.slice(0, 3).map((unlock) => (
                          <Badge key={unlock} color="amber">
                            {unlock}
                          </Badge>
                        ))}
                        {tgo.unlocks.length > 3 ? <Badge color="amber">+{tgo.unlocks.length - 3} more</Badge> : null}
                      </div>
                    </div>
                  ) : null}
                  {tgo.mastery_hint ? (
                    <div className="mt-4 rounded-xl bg-white/70 px-3 py-3 text-sm text-zinc-700 ring-1 ring-stone-200 dark:bg-black/10 dark:text-zinc-300 dark:ring-white/10">
                      <span className="font-semibold text-zinc-950 dark:text-white">Mastery marker:</span> {tgo.mastery_hint}
                    </div>
                  ) : null}
                </article>
              ))}
            </div>
          </WorkspaceCard>
        ))}
      </div>

      <WorkspaceCard>
        <div className="flex items-center justify-between gap-4">
          <Subheading>Current active branch</Subheading>
          <Text className="text-sm">These are the live skills at the front of the tree right now.</Text>
        </div>
        <div className="mt-5 grid gap-4 lg:grid-cols-3">
          {activeTGOs.map((tgo) => (
            <article key={tgo.code} className="rounded-2xl border border-blue-200 bg-blue-50 p-4 dark:border-blue-500/20 dark:bg-blue-500/10">
              <div className="flex items-center justify-between gap-3">
                <div className="text-sm font-semibold text-blue-950 dark:text-blue-100">{tgo.title}</div>
                <Badge color="blue">{tgo.stage}</Badge>
              </div>
              <Text className="mt-2 text-sm">{tgo.description}</Text>
            </article>
          ))}
        </div>
      </WorkspaceCard>
    </div>
  )
}
