'use client'

import { useEffect, useMemo, useState } from 'react'
import {
  ArrowTrendingDownIcon,
  ArrowTrendingUpIcon,
  CheckBadgeIcon,
  ExclamationTriangleIcon,
  SparklesIcon,
} from '@heroicons/react/16/solid'
import { Badge } from '@/components/badge'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { getDashboard, getSession, getTree } from '@/lib/api'
import type { Dashboard, Tree } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

function rankWidth(index: number, total: number) {
  if (total <= 1) {
    return '100%'
  }
  const step = 42 / Math.max(total - 1, 1)
  return `${Math.max(58, Math.round(100 - index * step))}%`
}

function stageTitle(stage: string) {
  return stage.replace(/-/g, ' ')
}

function stageCompletion(tree: Tree, completedCodes: Set<string>, activeCodes: Set<string>) {
  const groups = new Map<string, { label: string; total: number; completed: number; active: number }>()
  for (const tgo of tree.tgos ?? []) {
    const key = `${tgo.stage_order}:${tgo.stage}`
    if (!groups.has(key)) {
      groups.set(key, {
        label: stageTitle(tgo.stage),
        total: 0,
        completed: 0,
        active: 0,
      })
    }
    const group = groups.get(key)
    if (!group) {
      continue
    }
    group.total++
    if (completedCodes.has(tgo.code)) {
      group.completed++
    }
    if (activeCodes.has(tgo.code)) {
      group.active++
    }
  }
  return [...groups.values()]
}

export function ProgressView() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [tree, setTree] = useState<Tree | null>(null)
  const [needsOnboarding, setNeedsOnboarding] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const session = await getSession()
        if (!session.onboarding_complete) {
          if (!cancelled) {
            setNeedsOnboarding(true)
            setDashboard(null)
            setTree(null)
          }
          return
        }
        const state = await getDashboard()
        let treeData: Tree | null = null
        if (session.active_tree_slug) {
          treeData = await getTree(session.active_tree_slug)
        }
        if (!cancelled) {
          setDashboard(state)
          setTree(treeData)
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

  const completedCodes = useMemo(() => new Set((dashboard?.completed_tgos ?? []).map((tgo) => tgo.code)), [dashboard])
  const activeCodes = useMemo(() => new Set((dashboard?.active_tgos ?? []).map((tgo) => tgo.code)), [dashboard])
  const totalSkills = (tree?.tgos ?? []).length || (dashboard?.completed_tgos ?? []).length
  const completedCount = (dashboard?.completed_tgos ?? []).length
  const completionRatio = totalSkills > 0 ? Math.round((completedCount / totalSkills) * 100) : 0
  const stages = useMemo(() => (tree ? stageCompletion(tree, completedCodes, activeCodes) : []), [activeCodes, completedCodes, tree])

  if (loading) {
    return <LoadingState label="Loading progress board…" />
  }
  if (needsOnboarding) {
    return (
      <EmptyState
        title="Build your starter path first"
        body="Progress becomes meaningful once your skill map has been created. Start by setting your writing goals and recommended opening skills."
        actionHref="/onboarding"
        actionLabel="Set starter path"
      />
    )
  }
  if (error || !dashboard) {
    return <EmptyState title="Progress unavailable" body={error ?? 'Could not load progress board.'} actionHref="/" actionLabel="Back to assignment" />
  }

  const activeTGOs = dashboard.active_tgos ?? []
  const completedTGOs = dashboard.completed_tgos ?? []
  const upcomingTGOs = dashboard.upcoming_tgos ?? []
  const strongestSkills = dashboard.strongest_skills ?? []
  const weakestSkills = dashboard.weakest_skills ?? []
  const history = dashboard.history ?? []
  const recurringWeaknesses = dashboard.recurring_weaknesses ?? []
  const recurringFindings = dashboard.recurring_findings ?? []
  const recurringCompletedSlips = dashboard.recurring_completed_slips ?? []
  const headerSkills = activeTGOs.length > 0 ? activeTGOs.map((tgo) => tgo.title) : tree?.priority_skills ?? []

  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <Heading>Progress board</Heading>
          <Text className="mt-2 max-w-3xl">
            Active skills remain the primary measure. Mastered skills track durable gains, and regression watch keeps earlier wins from quietly eroding.
          </Text>
        </div>
        <div className="flex flex-wrap gap-2">
          {headerSkills.slice(0, 3).map((skill) => (
            <Badge key={skill} color="amber">
              {skill}
            </Badge>
          ))}
        </div>
      </header>

      <div className="grid gap-8 xl:grid-cols-[1.35fr_1fr]">
        <WorkspaceCard>
          <div className="flex items-start justify-between gap-4">
            <div>
              <Subheading>Track completion</Subheading>
              <Text className="mt-2">This is the broad curriculum view: how much of the current skill map has been mastered and how much is currently under active practice.</Text>
            </div>
            <Badge color="green">{completionRatio}% complete</Badge>
          </div>
          <div className="mt-6">
            <div className="h-3 rounded-full bg-stone-200 dark:bg-white/10">
              <div className="h-3 rounded-full bg-stone-800 dark:bg-stone-200" style={{ width: `${completionRatio}%` }} />
            </div>
            <div className="mt-3 flex flex-wrap items-center gap-4 text-sm text-zinc-600 dark:text-zinc-300">
              <span>{completedCount} completed</span>
              <span>{activeTGOs.length} active</span>
              <span>{upcomingTGOs.length} unlocked next</span>
              <span>{Math.max(totalSkills - completedCount - activeTGOs.length, 0)} still ahead</span>
            </div>
          </div>
          {stages.length > 0 ? (
            <div className="mt-8 grid gap-4 lg:grid-cols-2">
              {stages.map((stage) => {
                const stagePercent = stage.total > 0 ? Math.round((stage.completed / stage.total) * 100) : 0
                return (
                  <div key={stage.label} className="rounded-2xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
                    <div className="flex items-center justify-between gap-4">
                      <div className="text-sm font-semibold uppercase tracking-[0.18em] text-zinc-500 dark:text-zinc-400">{stage.label}</div>
                      <Badge color="zinc">{stage.completed}/{stage.total}</Badge>
                    </div>
                    <div className="mt-3 h-2 rounded-full bg-stone-200 dark:bg-white/10">
                      <div className="h-2 rounded-full bg-stone-700 dark:bg-stone-200" style={{ width: `${stagePercent}%` }} />
                    </div>
                    <Text className="mt-3 text-sm">
                      {stage.active > 0 ? `${stage.active} active now.` : 'No active skills in this stage.'}
                    </Text>
                  </div>
                )
              })}
            </div>
          ) : null}
        </WorkspaceCard>

        <WorkspaceCard>
          <Subheading>Current coaching state</Subheading>
          <div className="mt-5 space-y-4">
            <div className="rounded-2xl border border-blue-200 bg-blue-50 p-4 dark:border-blue-500/20 dark:bg-blue-500/10">
              <div className="flex items-center gap-2 text-sm font-semibold text-blue-900 dark:text-blue-200">
                <SparklesIcon className="size-4" />
                Active skills
              </div>
              <div className="mt-3 space-y-3">
                {activeTGOs.map((tgo) => (
                  <div key={tgo.code} className="rounded-xl border border-blue-200/80 bg-white/70 px-3 py-3 dark:border-blue-400/20 dark:bg-black/10">
                    <div className="flex items-center justify-between gap-3">
                      <span className="text-sm font-semibold text-blue-950 dark:text-blue-100">{tgo.title}</span>
                      <Badge color="blue">{tgo.stage}</Badge>
                    </div>
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
                      <Text className="mt-2 text-sm">
                        <span className="font-semibold text-blue-950 dark:text-blue-100">Mastery state:</span>{' '}
                        {tgo.mastery_stage ?? 'emerging'}
                      </Text>
                    )}
                  </div>
                ))}
              </div>
            </div>
            <div className="rounded-2xl border border-green-200 bg-green-50 p-4 dark:border-green-500/20 dark:bg-green-500/10">
              <div className="flex items-center gap-2 text-sm font-semibold text-green-900 dark:text-green-200">
                <CheckBadgeIcon className="size-4" />
                Mastered skills
              </div>
              <Text className="mt-2 text-sm">
                {completedTGOs.length === 0 ? 'No skills have been mastered yet.' : `${completedTGOs.length} skills have been marked mastered.`}
              </Text>
            </div>
            <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-500/20 dark:bg-amber-500/10">
              <div className="flex items-center gap-2 text-sm font-semibold text-amber-900 dark:text-amber-200">
                <ExclamationTriangleIcon className="size-4" />
                Regression watch
              </div>
              <Text className="mt-2 text-sm">
                {recurringCompletedSlips.length === 0
                  ? 'No mastered-skill regressions are currently being flagged.'
                  : `${recurringCompletedSlips.length} maintenance issues need attention.`}
              </Text>
            </div>
          </div>
        </WorkspaceCard>
      </div>

      <div className="grid gap-8 xl:grid-cols-2">
        <WorkspaceCard>
          <Subheading>Skill signals</Subheading>
          <Text className="mt-2">These are ranked signals, not absolute scores. They help you see where recent reviews consistently read strongest versus where the coaching load remains heaviest.</Text>
          <div className="mt-6 space-y-6">
            <div>
              <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-zinc-950 dark:text-white">
                <ArrowTrendingUpIcon className="size-4 text-green-600" />
                Strongest
              </div>
              <div className="space-y-3">
                {strongestSkills.map((item, index) => (
                  <div key={item}>
                    <div className="mb-1 flex items-center justify-between gap-4 text-sm">
                      <span className="text-zinc-900 dark:text-white">{item}</span>
                      <span className="text-zinc-500 dark:text-zinc-400">signal</span>
                    </div>
                    <div className="h-2 rounded-full bg-stone-200 dark:bg-white/10">
                      <div className="h-2 rounded-full bg-green-600" style={{ width: rankWidth(index, strongestSkills.length) }} />
                    </div>
                  </div>
                ))}
              </div>
            </div>
            <div>
              <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-zinc-950 dark:text-white">
                <ArrowTrendingDownIcon className="size-4 text-amber-600" />
                Weakest
              </div>
              <div className="space-y-3">
                {weakestSkills.map((item, index) => (
                  <div key={item}>
                    <div className="mb-1 flex items-center justify-between gap-4 text-sm">
                      <span className="text-zinc-900 dark:text-white">{item}</span>
                      <span className="text-zinc-500 dark:text-zinc-400">attention</span>
                    </div>
                    <div className="h-2 rounded-full bg-stone-200 dark:bg-white/10">
                      <div className="h-2 rounded-full bg-amber-600" style={{ width: rankWidth(index, weakestSkills.length) }} />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </WorkspaceCard>

        <WorkspaceCard>
          <Subheading>Recent activity timeline</Subheading>
          <Text className="mt-2">This compresses the recent coaching loop into a readable sequence so you can see what the system has emphasized lately.</Text>
          <ol className="mt-6 space-y-4">
            {history.length === 0 ? <li className="text-sm text-zinc-600 dark:text-zinc-300">No recent history yet.</li> : null}
            {history.map((item, index) => (
              <li key={`${item}-${index}`} className="relative pl-6">
                <span className="absolute left-0 top-1.5 size-2 rounded-full bg-stone-800 dark:bg-stone-200" />
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 text-sm text-zinc-700 dark:border-white/10 dark:bg-white/5 dark:text-zinc-300">
                  {item}
                </div>
              </li>
            ))}
          </ol>
        </WorkspaceCard>
      </div>

      <div className="grid gap-8 xl:grid-cols-3">
        <WorkspaceCard>
          <Subheading>Unlocked next</Subheading>
          <div className="mt-4 flex flex-wrap gap-2">
            {upcomingTGOs.length === 0 ? <Text>No new skill unlocks yet.</Text> : null}
            {upcomingTGOs.map((tgo) => (
              <Badge key={tgo.code} color="amber">
                {tgo.title}
              </Badge>
            ))}
          </div>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>Recurring weaknesses</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {recurringWeaknesses.length === 0 ? <li>No repeating weaknesses have been detected yet.</li> : null}
            {recurringWeaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>Analyzer trends</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {recurringFindings.length === 0 ? <li>No repeated analyzer findings yet.</li> : null}
            {recurringFindings.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>
    </div>
  )
}
