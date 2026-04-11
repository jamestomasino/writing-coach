'use client'

import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { PageHeader } from '@/components/page-header'
import { AppErrorState, EmptyState, LoadingState } from '@/components/status-state'
import { Text } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import { formatLocalDateTime } from '@/lib/datetime'
import { getPlaygroundSessions } from '@/lib/api'
import type { PlaygroundSession } from '@/lib/types'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import { useTranslations } from 'next-intl'
import { useEffect, useState } from 'react'

export function PlaygroundHistoryView() {
  const t = useTranslations('playgroundHistoryView')
  const { loading: sessionLoading, error: sessionError } = useRequiredAppSession('/playground/history')
  const [loading, setLoading] = useState(true)
  const [loadingMore, setLoadingMore] = useState(false)
  const [error, setError] = useState<string | null>(null)
  const [sessions, setSessions] = useState<PlaygroundSession[]>([])
  const [nextCursor, setNextCursor] = useState<number | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      setLoading(true)
      try {
        const next = await getPlaygroundSessions()
        if (!cancelled) {
          setSessions(next.sessions)
          setNextCursor(next.nextCursor ?? null)
          setError(null)
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
  }, [t])

  async function handleLoadMore() {
    if (loadingMore || !nextCursor) {
      return
    }
    setLoadingMore(true)
    try {
      const next = await getPlaygroundSessions(50, nextCursor)
      const seen = new Set(sessions.map((item) => item.id))
      const merged = sessions.concat(next.sessions.filter((item) => !seen.has(item.id)))
      setSessions(merged)
      setNextCursor(next.nextCursor ?? null)
      setError(null)
    } catch (err) {
      setError(err instanceof Error ? err.message : t('loadError'))
    } finally {
      setLoadingMore(false)
    }
  }

  if (sessionLoading || loading) {
    return <LoadingState label={t('loading')} />
  }
  if (sessionError || error) {
    return <AppErrorState title={t('unavailableTitle')} error={sessionError ?? error ?? t('loadError')} />
  }
  if (sessions.length === 0) {
    return (
      <EmptyState
        title={t('emptyTitle')}
        body={t('emptyBody')}
        actionHref="/playground"
        actionLabel={t('startNewReview')}
      />
    )
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={t('eyebrow')}
        title={t('title')}
        intro={t('intro')}
        actions={
          <Button href="/playground" color="dark/zinc">
            {t('startNewReview')}
          </Button>
        }
      />

      <div className="grid gap-6">
        {sessions.map((session) => (
          <WorkspaceCard key={session.id}>
            <CardHeader
              eyebrow={session.assignment_format || session.writing_type || t('eyebrow')}
              title={session.title}
              description={session.latest_review_at ? t('latestReview', { datetime: formatLocalDateTime(session.latest_review_at) ?? session.latest_review_at }) : t('updated', { datetime: formatLocalDateTime(session.updated_at) ?? session.updated_at })}
              actions={
                <Button href={`/playground/${session.id}`} color="dark/zinc">
                  {t('openSession')}
                </Button>
              }
            />
            <div className="mt-4 space-y-4">
              <div className="grid gap-3 sm:grid-cols-3">
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-2.5 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500 dark:text-zinc-400">Reviews</div>
                  <Text className="mt-1.5 text-sm">
                    {session.review_count === 0 ? t('reviewCountZero') : session.review_count === 1 ? t('reviewCountOne') : t('reviewCount', { count: session.review_count })}
                  </Text>
                </div>
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-2.5 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500 dark:text-zinc-400">{t('format')}</div>
                  <Text className="mt-1.5 text-sm">{session.assignment_format || '-'}</Text>
                </div>
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-2.5 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500 dark:text-zinc-400">{t('writingType')}</div>
                  <Text className="mt-1.5 text-sm">{session.writing_type || '-'}</Text>
                </div>
              </div>
              <Text className="text-sm">{session.content}</Text>
            </div>
          </WorkspaceCard>
        ))}
      </div>
      {nextCursor ? (
        <div className="flex justify-center">
          <Button onClick={handleLoadMore} outline disabled={loadingMore}>
            {loadingMore ? t('loadingMore') : t('loadMore')}
          </Button>
        </div>
      ) : null}
    </div>
  )
}
