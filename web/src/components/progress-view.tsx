'use client'

import { Badge } from '@/components/badge'
import { CardHeader } from '@/components/card-header'
import { Eyebrow } from '@/components/eyebrow'
import { Subheading } from '@/components/heading'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import type { Tree } from '@/lib/types'
import { useTrackDashboardData } from '@/lib/use-track-dashboard-data'
import {
  ArrowTrendingDownIcon,
  ArrowTrendingUpIcon,
  CheckBadgeIcon,
  ExclamationTriangleIcon,
  SparklesIcon,
} from '@heroicons/react/16/solid'
import { useMemo } from 'react'
import { MasteryProgress } from './mastery-progress'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
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
  const groups = new Map<string, { label: string; order: number; total: number; completed: number; active: number }>()
  for (const tgo of tree.tgos ?? []) {
    const key = tgo.stage
    if (!groups.has(key)) {
      groups.set(key, {
        label: stageTitle(tgo.stage),
        order: tgo.stage_order,
        total: 0,
        completed: 0,
        active: 0,
      })
    }
    const group = groups.get(key)
    if (!group) {
      continue
    }
    group.order = Math.min(group.order, tgo.stage_order)
    group.total++
    if (completedCodes.has(tgo.code)) {
      group.completed++
    }
    if (activeCodes.has(tgo.code)) {
      group.active++
    }
  }
  return [...groups.values()].sort((a, b) => a.order - b.order)
}

export function ProgressView() {
  const { sessionLoading, sessionError, loading, error, dashboard, tree, onboarding } = useTrackDashboardData(
    '/progress',
    { loadErrorMessage: 'Could not load progress' }
  )

  const completedCodes = useMemo(() => new Set((dashboard?.completed_tgos ?? []).map((tgo) => tgo.code)), [dashboard])
  const activeCodes = useMemo(() => new Set((dashboard?.active_tgos ?? []).map((tgo) => tgo.code)), [dashboard])
  const totalSkills = (tree?.tgos ?? []).length || (dashboard?.completed_tgos ?? []).length
  const completedCount = (dashboard?.completed_tgos ?? []).length
  const completionRatio = totalSkills > 0 ? Math.round((completedCount / totalSkills) * 100) : 0
  const stages = useMemo(
    () => (tree ? stageCompletion(tree, completedCodes, activeCodes) : []),
    [activeCodes, completedCodes, tree]
  )

  if (sessionLoading || loading) {
    return <LoadingState label="Loading progress board…" />
  }
  if (sessionError || error || !dashboard) {
    return <AppErrorState title="Progress unavailable" error={sessionError ?? error ?? 'Could not load progress board.'} />
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
  const completedAssignments = dashboard.completed_assignments ?? 0
  const profile = onboarding?.profile
  const profileCards = profile
    ? [
        { label: 'Writing type', value: profile.writing_type },
        { label: 'Experience level', value: profile.experience_level },
        { label: 'Desired tone', value: profile.desired_tone },
        { label: 'Goals', value: profile.writing_goals },
      ]
    : []

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Progress"
        title={tree?.title ?? 'Progress board'}
        intro={tree?.description ?? 'Active skills remain the primary measure.'}
      />

      {profile ? (
        <section aria-label="Track profile" className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {profileCards.map((item) => (
            <WorkspaceCard key={item.label} className="p-5">
              <Eyebrow>{item.label}</Eyebrow>
              <Text className="mt-3 text-sm leading-6 text-zinc-900 dark:text-white">{item.value}</Text>
            </WorkspaceCard>
          ))}
        </section>
      ) : (
        <WorkspaceCard>
          <Subheading>Track profile</Subheading>
          <Text className="mt-3">This track was created without a saved profile.</Text>
        </WorkspaceCard>
      )}

      <div className="grid gap-8 xl:grid-cols-[1.35fr_1fr]">
        <WorkspaceCard>
          <CardHeader
            eyebrow="Progress"
            title="Track completion"
            description="See how much of this track you've practiced and how much you've mastered."
            actions={<Badge color="green">{completionRatio}% complete</Badge>}
          />
          <div className="mt-6">
            <div className="h-3 rounded-full bg-stone-200 dark:bg-white/10">
              <div
                className="h-3 rounded-full bg-stone-800 dark:bg-stone-200"
                style={{ width: `${completionRatio}%` }}
              />
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
                  <div
                    key={stage.label}
                    className="rounded-2xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5"
                  >
                    <div className="flex items-center justify-between gap-4">
                      <Eyebrow>{stage.label}</Eyebrow>
                      <Badge color="zinc">
                        {stage.completed}/{stage.total}
                      </Badge>
                    </div>
                    <div className="mt-3 h-2 rounded-full bg-stone-200 dark:bg-white/10">
                      <div
                        className="h-2 rounded-full bg-stone-700 dark:bg-stone-200"
                        style={{ width: `${stagePercent}%` }}
                      />
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
          <CardHeader eyebrow="Status" title="Current coaching state" />
          <div className="mt-5 space-y-4">
            <div className="rounded-2xl border border-blue-200 bg-blue-50 p-4 dark:border-blue-500/20 dark:bg-blue-500/10">
              <div className="flex items-center gap-2 text-sm font-semibold text-blue-900 dark:text-blue-200">
                <SparklesIcon className="size-4" />
                Active skills
              </div>
              <div className="mt-3 space-y-3">
                {activeTGOs.map((tgo) => (
                  <div
                    key={tgo.code}
                    className="rounded-xl border border-blue-200/80 bg-white/70 px-3 py-3 dark:border-blue-400/20 dark:bg-black/10"
                  >
                    <div className="flex items-center justify-between gap-3">
                      <span className="text-sm font-semibold text-blue-950 dark:text-blue-100">{tgo.title}</span>
                      <Badge color="blue">{tgo.stage}</Badge>
                    </div>
                    <MasteryProgress tgo={tgo} tone="blue" />
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
                {completedTGOs.length === 0
                  ? 'No skills have been mastered yet.'
                  : `${completedTGOs.length} skills have been marked mastered.`}
              </Text>
            </div>
            <div className="rounded-2xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-900 dark:text-white">Assignments completed</div>
              <Text className="mt-2 text-sm">
                {completedAssignments === 0
                  ? 'No assignments have been completed yet.'
                  : `${completedAssignments} assignments have been completed, including revision rounds.`}
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
          <Text className="mt-2">
            A quick read on where your recent work looks strongest and where it still needs attention.
          </Text>
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
                      <div
                        className="h-2 rounded-full bg-green-600"
                        style={{ width: rankWidth(index, strongestSkills.length) }}
                      />
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
                      <div
                        className="h-2 rounded-full bg-amber-600"
                        style={{ width: rankWidth(index, weakestSkills.length) }}
                      />
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </WorkspaceCard>

        <WorkspaceCard>
          <Subheading>Recent activity timeline</Subheading>
          <Text className="mt-2">
            A quick view of your recent assignments, reviews, and next focus.
          </Text>
          <ol className="mt-6 space-y-4">
            {history.length === 0 ? (
              <li className="text-sm text-zinc-600 dark:text-zinc-300">No recent history yet.</li>
            ) : null}
            {history.map((item, index) => (
              <li key={`${item.title}-${index}`} className="relative pl-6">
                <span className="absolute top-1.5 left-0 size-2 rounded-full bg-stone-800 dark:bg-stone-200" />
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="text-sm font-medium text-zinc-900 dark:text-white">Assignment: {item.title}</div>
                  <div className="mt-3 flex flex-wrap gap-2">
                    {item.tgos.map((tgo) => (
                      <Badge key={tgo} color="zinc">
                        {tgo}
                      </Badge>
                    ))}
                  </div>
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
