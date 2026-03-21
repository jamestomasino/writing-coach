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
import { useTranslations } from 'next-intl'
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
  const t = useTranslations('progressView')
  const { sessionLoading, sessionError, loading, error, dashboard, tree, onboarding } = useTrackDashboardData(
    '/progress',
    { loadErrorMessage: t('loadError') }
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
    return <LoadingState label={t('loading')} />
  }
  if (sessionError || error || !dashboard) {
    return <AppErrorState title={t('unavailableTitle')} error={sessionError ?? error ?? t('unavailableBody')} />
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
  const draftCount = dashboard.draft_count ?? history.length
  const revisionCount = dashboard.revision_count ?? 0
  const profile = onboarding?.profile
  const profileCards = profile
    ? [
        { label: t('writingType'), value: profile.writing_type },
        { label: t('experienceLevel'), value: profile.experience_level },
        { label: t('desiredTone'), value: profile.desired_tone },
        { label: t('goals'), value: profile.writing_goals },
      ]
    : []

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={t('eyebrow')}
        title={tree?.title ?? t('defaultTitle')}
        intro={tree?.description ?? t('defaultIntro')}
      />

      {profile ? (
        <section aria-label={t('profileAria')} className="grid gap-4 md:grid-cols-2 xl:grid-cols-4">
          {profileCards.map((item) => (
            <WorkspaceCard key={item.label} className="p-5">
              <Eyebrow>{item.label}</Eyebrow>
              <Text className="mt-3 text-sm leading-6 text-zinc-900 dark:text-white">{item.value}</Text>
            </WorkspaceCard>
          ))}
        </section>
      ) : (
        <WorkspaceCard>
          <Subheading>{t('profileTitle')}</Subheading>
          <Text className="mt-3">{t('profileMissing')}</Text>
        </WorkspaceCard>
      )}

      <div className="grid gap-8 xl:grid-cols-[1.35fr_1fr]">
        <WorkspaceCard>
          <CardHeader
            eyebrow={t('progressEyebrow')}
            title={t('progressTitle')}
            description={t('progressDescription')}
            actions={<Badge color="green">{t('complete', { percent: completionRatio })}</Badge>}
          />
          <div className="mt-6">
            <div className="h-3 rounded-full bg-stone-200 dark:bg-white/10">
              <div
                className="h-3 rounded-full bg-stone-800 dark:bg-stone-200"
                style={{ width: `${completionRatio}%` }}
              />
            </div>
            <div className="mt-3 flex flex-wrap items-center gap-4 text-sm text-zinc-600 dark:text-zinc-300">
              <span>{t('completedCount', { count: completedCount })}</span>
              <span>{t('activeCount', { count: activeTGOs.length })}</span>
              <span>{t('readyNextCount', { count: upcomingTGOs.length })}</span>
              <span>{t('stillAheadCount', { count: Math.max(totalSkills - completedCount - activeTGOs.length, 0) })}</span>
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
                      {stage.active > 0 ? t('activeNow', { count: stage.active }) : t('noActiveInStage')}
                    </Text>
                  </div>
                )
              })}
            </div>
          ) : null}
        </WorkspaceCard>

        <WorkspaceCard>
          <CardHeader eyebrow={t('rightNowEyebrow')} title={t('rightNowTitle')} />
          <div className="mt-5 space-y-4">
            <div className="rounded-2xl border border-blue-200 bg-blue-50 p-4 dark:border-blue-500/20 dark:bg-blue-500/10">
              <div className="flex items-center gap-2 text-sm font-semibold text-blue-900 dark:text-blue-200">
                <SparklesIcon className="size-4" />
                {t('activeSkills')}
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
                {t('masteredSkills')}
              </div>
              <Text className="mt-2 text-sm">
                {completedTGOs.length === 0 ? t('noMasteredSkills') : t('masteredSkillsCount', { count: completedTGOs.length })}
              </Text>
            </div>
            <div className="rounded-2xl border border-stone-200 bg-stone-50 p-4 dark:border-white/10 dark:bg-white/5">
              <div className="text-sm font-semibold text-zinc-900 dark:text-white">{t('assignmentsCompleted')}</div>
              <div className="mt-4 grid grid-cols-3 gap-3">
                <div className="rounded-xl border border-stone-200 bg-white/70 px-3 py-3 text-center dark:border-white/10 dark:bg-black/10">
                  <div className="text-[11px] font-semibold uppercase tracking-[0.18em] text-zinc-500 dark:text-zinc-400">
                    {t('draftsStat')}
                  </div>
                  <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{draftCount}</div>
                </div>
                <div className="rounded-xl border border-stone-200 bg-white/70 px-3 py-3 text-center dark:border-white/10 dark:bg-black/10">
                  <div className="text-[11px] font-semibold uppercase tracking-[0.18em] text-zinc-500 dark:text-zinc-400">
                    {t('revisionsStat')}
                  </div>
                  <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{revisionCount}</div>
                </div>
                <div className="rounded-xl border border-stone-200 bg-white/70 px-3 py-3 text-center dark:border-white/10 dark:bg-black/10">
                  <div className="text-[11px] font-semibold uppercase tracking-[0.18em] text-zinc-500 dark:text-zinc-400">
                    {t('completedStat')}
                  </div>
                  <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{completedAssignments}</div>
                </div>
              </div>
            </div>
            <div className="rounded-2xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-500/20 dark:bg-amber-500/10">
              <div className="flex items-center gap-2 text-sm font-semibold text-amber-900 dark:text-amber-200">
                <ExclamationTriangleIcon className="size-4" />
                {t('olderSkillsToWatch')}
              </div>
              <Text className="mt-2 text-sm">
                {recurringCompletedSlips.length === 0
                  ? t('noOlderSkillsSlipping')
                  : t('olderSkillsNeedAttention', { count: recurringCompletedSlips.length })}
              </Text>
            </div>
          </div>
        </WorkspaceCard>
      </div>

      <div className="grid gap-8 xl:grid-cols-2">
        <WorkspaceCard>
          <Subheading>{t('recentPatternsTitle')}</Subheading>
          <Text className="mt-2">{t('recentPatternsBody')}</Text>
          <div className="mt-6 space-y-6">
            <div>
              <div className="mb-3 flex items-center gap-2 text-sm font-semibold text-zinc-950 dark:text-white">
                <ArrowTrendingUpIcon className="size-4 text-green-600" />
                {t('lookingStrongest')}
              </div>
              <div className="space-y-3">
                {strongestSkills.map((item, index) => (
                  <div key={item}>
                    <div className="mb-1 flex items-center justify-between gap-4 text-sm">
                      <span className="text-zinc-900 dark:text-white">{item}</span>
                      <span className="text-zinc-500 dark:text-zinc-400">{t('signal')}</span>
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
                {t('needsAttention')}
              </div>
              <div className="space-y-3">
                {weakestSkills.map((item, index) => (
                  <div key={item}>
                    <div className="mb-1 flex items-center justify-between gap-4 text-sm">
                      <span className="text-zinc-900 dark:text-white">{item}</span>
                      <span className="text-zinc-500 dark:text-zinc-400">{t('attention')}</span>
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
          <Subheading>{t('recentAssignmentsTitle')}</Subheading>
          <Text className="mt-2">{t('recentAssignmentsBody')}</Text>
          <ol className="mt-6 space-y-4">
            {history.length === 0 ? (
              <li className="text-sm text-zinc-600 dark:text-zinc-300">{t('noRecentHistory')}</li>
            ) : null}
            {history.map((item, index) => (
              <li key={`${item.title}-${index}`} className="relative pl-6">
                <span className="absolute top-1.5 left-0 size-2 rounded-full bg-stone-800 dark:bg-stone-200" />
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="text-sm font-medium text-zinc-900 dark:text-white">{item.title}</div>
                  <div className="mt-3 flex flex-wrap gap-2">
                    {(item.tgos ?? []).map((tgo) => (
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
          <Subheading>{t('unlockedNextTitle')}</Subheading>
          <div className="mt-4 flex flex-wrap gap-2">
            {upcomingTGOs.length === 0 ? <Text>{t('noNewSkillUnlocks')}</Text> : null}
            {upcomingTGOs.map((tgo) => (
              <Badge key={tgo.code} color="amber">
                {tgo.title}
              </Badge>
            ))}
          </div>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>{t('recurringWeaknessesTitle')}</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {recurringWeaknesses.length === 0 ? <li>{t('noRecurringWeaknesses')}</li> : null}
            {recurringWeaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>{t('analyzerTrendsTitle')}</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {recurringFindings.length === 0 ? <li>{t('noAnalyzerTrends')}</li> : null}
            {recurringFindings.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>
    </div>
  )
}
