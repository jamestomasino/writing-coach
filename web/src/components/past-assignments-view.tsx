'use client'

import { useTranslations } from 'next-intl'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { getAssignments, getDashboard } from '@/lib/api'
import type { AssignmentSummary, Dashboard, TGO } from '@/lib/types'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import { ArrowRightIcon } from '@heroicons/react/16/solid'
import { useEffect, useMemo, useState } from 'react'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'
import { formatLocalDateTime } from '@/lib/datetime'

function badgeColorForMasteryStage(stage?: string): React.ComponentProps<typeof Badge>['color'] {
  switch ((stage ?? '').trim().toLowerCase()) {
    case 'mastery evidence':
      return 'green'
    case 'strong control':
      return 'cyan'
    case 'developing':
      return 'amber'
    default:
      return 'zinc'
  }
}

function badgeColorForTGO(title: string, byTitle: Map<string, TGO>): React.ComponentProps<typeof Badge>['color'] {
  const tgo = byTitle.get(title.trim().toLowerCase())
  if (!tgo) {
    return 'zinc'
  }
  return badgeColorForMasteryStage(tgo.mastery_stage)
}

export function PastAssignmentsView() {
  const t = useTranslations('pastAssignmentsView')
  const { session, loading: sessionLoading, error: sessionError } = useRequiredAppSession('/assignments')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [assignments, setAssignments] = useState<AssignmentSummary[]>([])
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      if (!session) {
        return
      }
      try {
        const [items, dashboardState] = await Promise.all([getAssignments(), getDashboard()])
        if (!cancelled) {
          setAssignments(items)
          setDashboard(dashboardState)
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
  }, [session, t])

  const currentAssignment = useMemo(() => assignments.find((item) => item.is_current), [assignments])
  const pastAssignments = useMemo(() => assignments.filter((item) => !item.is_current), [assignments])
  const tgoByTitle = useMemo(() => {
    const entries = new Map<string, TGO>()
    for (const tgo of dashboard?.completed_tgos ?? []) {
      entries.set(tgo.title.trim().toLowerCase(), tgo)
    }
    for (const tgo of dashboard?.upcoming_tgos ?? []) {
      if (!entries.has(tgo.title.trim().toLowerCase())) {
        entries.set(tgo.title.trim().toLowerCase(), tgo)
      }
    }
    for (const tgo of dashboard?.active_tgos ?? []) {
      entries.set(tgo.title.trim().toLowerCase(), tgo)
    }
    return entries
  }, [dashboard])

  if (sessionLoading || loading) {
    return <LoadingState label={t('loading')} />
  }
  if (sessionError || error) {
    return <AppErrorState title={t('unavailableTitle')} error={sessionError ?? error ?? t('unavailableBody')} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={t('eyebrow')}
        title={t('title')}
        intro={t('intro')}
      />

      {currentAssignment ? (
        <WorkspaceCard>
          <CardHeader
            eyebrow={t('currentAssignmentEyebrow')}
            title={currentAssignment.title}
            description={t('currentAssignmentDescription')}
            actions={
              <Button href={`/assignments/${currentAssignment.current_exercise_id}`} plain>
                {t('viewCurrentTimeline')}
              </Button>
            }
          />
        </WorkspaceCard>
      ) : null}

      {pastAssignments.length === 0 ? (
        <EmptyState
          title={t('emptyTitle')}
          body={t('emptyBody')}
          actionHref="/"
          actionLabel={t('openCurrentAssignment')}
        />
      ) : (
        <div className="grid gap-6 xl:grid-cols-2">
          {pastAssignments.map((assignment) => (
            <WorkspaceCard key={assignment.root_exercise_id}>
              <CardHeader
                title={assignment.title}
                description={t('latestActivity', { datetime: formatLocalDateTime(assignment.latest_activity) ?? assignment.latest_activity })}
                actions={
                  <Button href={`/assignments/${assignment.current_exercise_id}`} color="dark/zinc">
                    {t('openTimeline')}
                    <ArrowRightIcon />
                  </Button>
                }
              />
              <div className="mt-5 flex flex-wrap gap-2">
                {assignment.tgos.map((tgo) => (
                  <Badge key={tgo} color={badgeColorForTGO(tgo, tgoByTitle)}>
                    {tgo}
                  </Badge>
                ))}
              </div>
              <div className="mt-5 grid gap-4 sm:grid-cols-4">
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400">
                    {t('promptsStat')}
                  </div>
                  <Text className="mt-2 text-sm">{assignment.exercise_count}</Text>
                </div>
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400">
                    {t('firstDraftsStat')}
                  </div>
                  <Text className="mt-2 text-sm">{assignment.draft_count}</Text>
                </div>
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400">
                    {t('feedbackStat')}
                  </div>
                  <Text className="mt-2 text-sm">{assignment.review_count}</Text>
                </div>
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400">
                    {t('revisionsStat')}
                  </div>
                  <Text className="mt-2 text-sm">{assignment.revision_count}</Text>
                </div>
              </div>
            </WorkspaceCard>
          ))}
        </div>
      )}
    </div>
  )
}
