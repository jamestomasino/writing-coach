'use client'

import { useEffect, useMemo, useState } from 'react'
import { ArrowPathIcon, CheckBadgeIcon, LockClosedIcon, SparklesIcon } from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
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
    const groups = new Map<string, { order: number; title: string; items: TreeNode[] }>()

    for (const tgo of treeTGOs) {
      const status: TreeNodeStatus = active.has(tgo.code)
        ? 'active'
        : completed.has(tgo.code)
          ? 'completed'
          : unlocked.has(tgo.code)
            ? 'unlocked'
            : 'locked'
      const key = `${tgo.stage_order}:${tgo.stage}`
      if (!groups.has(key)) {
        groups.set(key, {
          order: tgo.stage_order,
          title: stageTitle(tgo.stage),
          items: [],
        })
      }
      groups.get(key)?.items.push({
        ...tgo,
        status,
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

  const profile = onboarding.profile
  const activeTGOs = dashboard.active_tgos ?? []
  const completedTGOs = dashboard.completed_tgos ?? []
  const upcomingTGOs = dashboard.upcoming_tgos ?? []
  const prioritySkills = tree.priority_skills ?? []

  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <Heading>{tree.title}</Heading>
          <Text className="mt-2 max-w-3xl">{tree.description}</Text>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button href="/new-assignment" outline>
            New assignment
          </Button>
          <Button href="/onboarding" color="dark/zinc">
            <ArrowPathIcon />
            Refresh track
          </Button>
        </div>
      </header>

      <div className="grid gap-8 xl:grid-cols-[1.35fr_1fr]">
        <WorkspaceCard>
          <div className="flex items-center justify-between gap-4">
            <Subheading>Current track state</Subheading>
            <Badge color="zinc">{completedTGOs.length} mastered</Badge>
          </div>
          <div className="mt-5 grid gap-4 sm:grid-cols-3">
            <div className="rounded-2xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">Active now</div>
              <div className="mt-2 text-3xl font-semibold text-stone-900 dark:text-stone-100">{activeTGOs.length}</div>
              <Text className="mt-1 text-sm">These three skills drive the current assignment review.</Text>
            </div>
            <div className="rounded-2xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">Unlocked next</div>
              <div className="mt-2 text-3xl font-semibold text-stone-900 dark:text-stone-100">{upcomingTGOs.length}</div>
              <Text className="mt-1 text-sm">These are eligible choices for the next assignment cycle.</Text>
            </div>
            <div className="rounded-2xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-950 dark:text-white">Priority skills</div>
              <div className="mt-2 flex flex-wrap gap-2">
                {prioritySkills.slice(0, 3).map((skill) => (
                  <Badge key={skill} color="amber">
                    {skill}
                  </Badge>
                ))}
              </div>
            </div>
          </div>
        </WorkspaceCard>

        <WorkspaceCard>
          <Subheading>Track profile</Subheading>
          {profile ? (
            <dl className="mt-4 space-y-4 text-sm text-zinc-700 dark:text-zinc-300">
              <div>
                <dt className="font-semibold text-zinc-950 dark:text-white">Writing type</dt>
                <dd className="mt-1">{profile.writing_type}</dd>
              </div>
              <div>
                <dt className="font-semibold text-zinc-950 dark:text-white">Experience level</dt>
                <dd className="mt-1">{profile.experience_level}</dd>
              </div>
              <div>
                <dt className="font-semibold text-zinc-950 dark:text-white">Desired tone</dt>
                <dd className="mt-1">{profile.desired_tone}</dd>
              </div>
              <div>
                <dt className="font-semibold text-zinc-950 dark:text-white">Goals</dt>
                <dd className="mt-1">{profile.writing_goals}</dd>
              </div>
            </dl>
          ) : (
            <Text className="mt-3">This track was seeded without a persisted onboarding profile.</Text>
          )}
        </WorkspaceCard>
      </div>

      <div className="grid gap-8 xl:grid-cols-[1.35fr_1fr]">
        <WorkspaceCard>
          <div className="flex items-center justify-between gap-4">
            <Subheading>Skill map</Subheading>
            <Text className="text-sm">Three active at a time. Unlocks flow from mastered work.</Text>
          </div>
          <div className="mt-6 space-y-8">
            {stages.map((stage) => (
              <section key={`${stage.order}-${stage.title}`}>
                <div className="mb-4 flex items-center gap-3">
                  <Badge color="zinc">Stage {stage.order + 1}</Badge>
                  <h3 className="text-sm font-semibold uppercase tracking-[0.18em] text-zinc-500 dark:text-zinc-400">{stage.title}</h3>
                </div>
                <div className="grid gap-4 lg:grid-cols-2">
                  {stage.items.map((tgo) => (
                    <article
                      key={tgo.code}
                      className="rounded-2xl border border-stone-200 bg-stone-50 p-5 dark:border-white/10 dark:bg-white/5"
                    >
                      <div className="flex items-start justify-between gap-4">
                        <div>
                          <div className="text-base font-semibold text-zinc-950 dark:text-white">{tgo.title}</div>
                          <Text className="mt-2 text-sm">{tgo.description}</Text>
                        </div>
                        <Badge color={statusColor(tgo.status)}>{statusLabel(tgo.status)}</Badge>
                      </div>
                      {tgo.mastery_hint ? (
                        <div className="mt-4 rounded-xl bg-white/70 px-3 py-3 text-sm text-zinc-700 ring-1 ring-stone-200 dark:bg-black/10 dark:text-zinc-300 dark:ring-white/10">
                          <span className="font-semibold text-zinc-950 dark:text-white">Mastery marker:</span> {tgo.mastery_hint}
                        </div>
                      ) : null}
                      {tgo.prerequisites && tgo.prerequisites.length > 0 ? (
                        <div className="mt-4">
                          <div className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500 dark:text-zinc-400">Prerequisites</div>
                          <div className="mt-2 flex flex-wrap gap-2">
                            {tgo.prerequisites.map((prereq) => (
                              <Badge key={prereq} color="zinc">
                                {prereq}
                              </Badge>
                            ))}
                          </div>
                        </div>
                      ) : null}
                    </article>
                  ))}
                </div>
              </section>
            ))}
          </div>
        </WorkspaceCard>

        <div className="space-y-8">
          <WorkspaceCard>
            <Subheading>Active skills</Subheading>
            <div className="mt-4 space-y-3">
              {activeTGOs.map((tgo) => (
                <div key={tgo.code} className="rounded-xl border border-blue-200 bg-blue-50 px-4 py-3 dark:border-blue-500/20 dark:bg-blue-500/10">
                  <div className="flex items-center gap-2 text-sm font-semibold text-blue-900 dark:text-blue-200">
                    <SparklesIcon className="size-4" />
                    {tgo.title}
                  </div>
                  <Text className="mt-1 text-sm">{tgo.description}</Text>
                  {tgo.progress_mode === 'percent' ? (
                    <div className="mt-3">
                      <div className="flex items-center justify-between gap-3 text-xs font-medium uppercase tracking-[0.16em] text-blue-800/80 dark:text-blue-200/80">
                        <span>Mastery progress</span>
                        <span>{tgo.mastery_percent ?? 0}%</span>
                      </div>
                      <div className="mt-2 h-2 rounded-full bg-blue-200/70 dark:bg-blue-200/15">
                        <div className="h-2 rounded-full bg-blue-800 dark:bg-blue-200" style={{ width: `${tgo.mastery_percent ?? 0}%` }} />
                      </div>
                    </div>
                  ) : (
                    <Text className="mt-3 text-sm">
                      <span className="font-semibold text-blue-900 dark:text-blue-200">Mastery state:</span>{' '}
                      {tgo.mastery_stage ?? 'emerging'}
                    </Text>
                  )}
                </div>
              ))}
            </div>
          </WorkspaceCard>

          <WorkspaceCard>
            <Subheading>Mastered skills</Subheading>
            <div className="mt-4 space-y-3">
              {completedTGOs.length === 0 ? <Text>No mastered skills yet.</Text> : null}
              {completedTGOs.map((tgo) => (
                <div key={tgo.code} className="rounded-xl border border-green-200 bg-green-50 px-4 py-3 dark:border-green-500/20 dark:bg-green-500/10">
                  <div className="flex items-center gap-2 text-sm font-semibold text-green-900 dark:text-green-200">
                    <CheckBadgeIcon className="size-4" />
                    {tgo.title}
                  </div>
                  <Text className="mt-1 text-sm">{tgo.description}</Text>
                </div>
              ))}
            </div>
          </WorkspaceCard>

          <WorkspaceCard>
            <Subheading>Locked path</Subheading>
            <div className="mt-4 space-y-3">
              {stages.flatMap((stage) => stage.items).filter((tgo) => tgo.status === 'locked').slice(0, 5).map((tgo) => (
                <div key={tgo.code} className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="flex items-center gap-2 text-sm font-semibold text-zinc-900 dark:text-white">
                    <LockClosedIcon className="size-4" />
                    {tgo.title}
                  </div>
                  <Text className="mt-1 text-sm">{tgo.prerequisites?.length ? `Requires ${tgo.prerequisites.join(', ')}` : 'Locked until this branch opens.'}</Text>
                </div>
              ))}
            </div>
          </WorkspaceCard>
        </div>
      </div>
    </div>
  )
}
