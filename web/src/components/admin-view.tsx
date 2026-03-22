'use client'

import { Badge, BadgeButton } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Dialog, DialogActions, DialogBody, DialogTitle } from '@/components/dialog'
import { Field, FieldGroup, Label } from '@/components/fieldset'
import { PageHeader } from '@/components/page-header'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/table'
import { Text } from '@/components/text'
import { formatLocalDateTime } from '@/lib/datetime'
import { useAdminWorkspace } from '@/lib/use-admin-workspace'
import { useTranslations } from 'next-intl'
import { useState } from 'react'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

function humanizeEventLabel(value: string) {
  return value.replaceAll('_', ' ')
}

function humanizeCategoryLabel(value: string) {
  if (value === 'uncategorized') {
    return 'Uncategorized'
  }
  return value.replaceAll('_', ' ')
}

function categoryBadgeColor(value?: string) {
  switch (value) {
    case 'auth':
    case 'quota':
      return 'rose'
    case 'local_rate_limit':
    case 'rate_limit':
    case 'timeout':
      return 'amber'
    case 'settings':
      return 'blue'
    case 'provider':
      return 'cyan'
    case 'generation':
      return 'green'
    default:
      return 'zinc'
  }
}

function providerBadgeColor(value?: string) {
  switch ((value ?? '').toLowerCase()) {
    case 'openai':
      return 'green'
    case 'anthropic':
      return 'amber'
    case 'gemini':
      return 'blue'
    case 'groq':
      return 'fuchsia'
    case 'xai':
      return 'purple'
    case 'unknown':
      return 'zinc'
    default:
      return 'cyan'
  }
}

function filterChipClass(active: boolean) {
  return active ? 'ring-1 ring-inset ring-current' : 'opacity-55 hover:opacity-100'
}

type EventSortKey = 'created_at' | 'provider' | 'event' | 'category' | 'status_code' | 'user'

function compareValues(left: string | number, right: string | number) {
  if (typeof left === 'number' && typeof right === 'number') {
    return left - right
  }
  return String(left).localeCompare(String(right))
}

function userLabel(userSlug: string | undefined, userID: number) {
  if (userSlug && userSlug.trim() !== '') {
    return `${userSlug} · #${userID}`
  }
  return `#${userID}`
}

export function AdminView() {
  const t = useTranslations('adminView')
  const [sortKey, setSortKey] = useState<EventSortKey>('created_at')
  const [sortDirection, setSortDirection] = useState<'asc' | 'desc'>('desc')
  const [selectedDetailsEvent, setSelectedDetailsEvent] = useState<(typeof providerEvents)[number] | null>(null)
  const {
    loading,
    error,
    setError,
    session,
    admins,
    users,
    providerSummary,
    providerEvents,
    selectedHours,
    setSelectedHours,
    selectedProvider,
    setSelectedProvider,
    selectedEvent,
    setSelectedEvent,
    loadingProviderActivity,
  } = useAdminWorkspace()

  const sortedProviderEvents = [...providerEvents].sort((left, right) => {
    let result = 0
    switch (sortKey) {
      case 'provider':
        result = compareValues(left.provider || 'unknown', right.provider || 'unknown')
        break
      case 'event':
        result = compareValues(humanizeEventLabel(left.event), humanizeEventLabel(right.event))
        break
      case 'category':
        result = compareValues(left.category || 'uncategorized', right.category || 'uncategorized')
        break
      case 'status_code':
        result = compareValues(left.status_code || 0, right.status_code || 0)
        break
      case 'user':
        result = compareValues(userLabel(left.user_slug, left.user_id), userLabel(right.user_slug, right.user_id))
        break
      case 'created_at':
      default:
        result = compareValues(Date.parse(left.created_at) || 0, Date.parse(right.created_at) || 0)
        break
    }
    if (result === 0) {
      result = compareValues(right.id, left.id)
    }
    return sortDirection === 'asc' ? result : -result
  })

  function toggleSort(nextKey: EventSortKey) {
    if (sortKey === nextKey) {
      setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc')
      return
    }
    setSortKey(nextKey)
    setSortDirection(nextKey === 'created_at' ? 'desc' : 'asc')
  }

  function sortIndicator(key: EventSortKey) {
    if (sortKey !== key) {
      return '↕'
    }
    return sortDirection === 'asc' ? '↑' : '↓'
  }

  if (loading) {
    return <LoadingState label={t('loading')} />
  }
  if (!session?.is_admin) {
    return (
      <EmptyState
        title={t('unavailableTitle')}
        body={t('unavailableBody')}
        actionHref="/"
        actionLabel={t('backToAssignment')}
      />
    )
  }
  if (error && users.length === 0) {
    return <AppErrorState title={t('unavailableTitle')} error={error} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={t('pageEyebrow')}
        title={t('pageTitle')}
        intro={t('pageIntro')}
      />

      {error ? <EmptyState title={t('actionFailedTitle')} body={error} /> : null}

      <div className="grid gap-8 xl:grid-cols-[1.4fr_1fr]">
        <WorkspaceCard>
          <CardHeader eyebrow={t('accessEyebrow')} title={t('accessTitle')} />
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {admins.map((admin) => (
              <li key={admin}>• {admin}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>

      <WorkspaceCard>
        <CardHeader
          eyebrow={t('aiOpsEyebrow')}
          title={t('providerActivityTitle')}
          description={t('providerActivityDescription')}
        />
        <FieldGroup className="mt-5 md:grid-cols-3">
          <Field>
            <Label>{t('window')}</Label>
            <select
              value={selectedHours}
              onChange={(event) => setSelectedHours(event.target.value)}
              className="block w-full rounded-xl border border-white/10 bg-white/5 px-4 py-3 text-zinc-950 dark:text-white"
            >
              <option value="1">{t('lastHour')}</option>
              <option value="6">{t('last6Hours')}</option>
              <option value="24">{t('last24Hours')}</option>
              <option value="72">{t('last3Days')}</option>
              <option value="168">{t('last7Days')}</option>
            </select>
          </Field>
        </FieldGroup>
        {providerSummary ? (
          <>
            {loadingProviderActivity ? <Text className="mt-4 text-sm">{t('refreshingActivity')}</Text> : null}
            <div className="mt-5 grid gap-4 md:grid-cols-4">
              <div className="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 dark:border-white/10 dark:bg-white/5">
                <div className="text-xs uppercase tracking-[0.2em] text-zinc-500">{t('events')}</div>
                <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{providerSummary.total}</div>
                <div className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
                  {t('since', { time: formatLocalDateTime(providerSummary.since) ?? providerSummary.since })}
                </div>
              </div>
              <div className="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 dark:border-white/10 dark:bg-white/5">
                <div className="text-xs uppercase tracking-[0.2em] text-zinc-500">{t('validationFailures')}</div>
                <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{providerSummary.validation_failures}</div>
              </div>
              <div className="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 dark:border-white/10 dark:bg-white/5">
                <div className="text-xs uppercase tracking-[0.2em] text-zinc-500">{t('validationThrottles')}</div>
                <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{providerSummary.validation_rate_limit}</div>
              </div>
              <div className="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 dark:border-white/10 dark:bg-white/5">
                <div className="text-xs uppercase tracking-[0.2em] text-zinc-500">{t('fallbacks')}</div>
                <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{providerSummary.fallbacks}</div>
              </div>
            </div>

            <div className="mt-6 grid gap-6 xl:grid-cols-[1fr_1fr]">
              <div>
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('topProviders')}</div>
                <div className="mt-3 flex flex-wrap gap-2">
                  <BadgeButton
                    color="emerald"
                    className={filterChipClass(selectedProvider === '')}
                    onClick={() => setSelectedProvider('')}
                  >
                    {t('allProviders')} {providerSummary.total}
                  </BadgeButton>
                  {providerSummary.provider_counts.map((item) => (
                    <BadgeButton
                      key={item.label}
                      color={providerBadgeColor(item.label)}
                      className={filterChipClass(selectedProvider === item.label)}
                      onClick={() => setSelectedProvider(selectedProvider === item.label ? '' : item.label)}
                    >
                      {item.label} {item.count}
                    </BadgeButton>
                  ))}
                </div>
              </div>
              <div>
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('eventType')}</div>
                <div className="mt-3 flex flex-wrap gap-2">
                  <BadgeButton
                    color="sky"
                    className={filterChipClass(selectedEvent === '')}
                    onClick={() => setSelectedEvent('')}
                  >
                    {t('allEventTypes')} {providerSummary.total}
                  </BadgeButton>
                  {providerSummary.event_counts.map((item) => (
                    <BadgeButton
                      key={item.label}
                      color="sky"
                      className={filterChipClass(selectedEvent === item.label)}
                      onClick={() => setSelectedEvent(selectedEvent === item.label ? '' : item.label)}
                    >
                      {humanizeEventLabel(item.label)} {item.count}
                    </BadgeButton>
                  ))}
                </div>
              </div>
            </div>

            <div className="mt-6 max-h-[42rem] overflow-y-auto pr-2">
              <Table bleed dense striped grid>
                <TableHead>
                  <TableRow>
                    <TableHeader>
                      <button type="button" className="cursor-pointer font-medium" onClick={() => toggleSort('created_at')}>
                        Time {sortIndicator('created_at')}
                      </button>
                    </TableHeader>
                    <TableHeader>
                      <button type="button" className="cursor-pointer font-medium" onClick={() => toggleSort('provider')}>
                        Provider {sortIndicator('provider')}
                      </button>
                    </TableHeader>
                    <TableHeader>
                      <button type="button" className="cursor-pointer font-medium" onClick={() => toggleSort('event')}>
                        Event {sortIndicator('event')}
                      </button>
                    </TableHeader>
                    <TableHeader>
                      <button type="button" className="cursor-pointer font-medium" onClick={() => toggleSort('category')}>
                        Category {sortIndicator('category')}
                      </button>
                    </TableHeader>
                    <TableHeader>
                      <button type="button" className="cursor-pointer font-medium" onClick={() => toggleSort('status_code')}>
                        HTTP {sortIndicator('status_code')}
                      </button>
                    </TableHeader>
                    <TableHeader>
                      <button type="button" className="cursor-pointer font-medium" onClick={() => toggleSort('user')}>
                        {t('user')} {sortIndicator('user')}
                      </button>
                    </TableHeader>
                    <TableHeader>Details</TableHeader>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {sortedProviderEvents.map((event) => (
                    <TableRow key={event.id}>
                      <TableCell className="text-sm text-zinc-600 dark:text-zinc-400">
                        {formatLocalDateTime(event.created_at) ?? event.created_at}
                      </TableCell>
                      <TableCell>
                        <Badge color={providerBadgeColor(event.provider)}>{event.provider || t('unknown')}</Badge>
                      </TableCell>
                      <TableCell className="font-medium text-zinc-950 dark:text-white">
                        {humanizeEventLabel(event.event)}
                      </TableCell>
                      <TableCell>
                        <Badge color={categoryBadgeColor(event.category)}>
                          {event.category === 'uncategorized' || !event.category ? t('uncategorized') : humanizeCategoryLabel(event.category)}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {event.status_code ? <Badge color="zinc">{event.status_code}</Badge> : <span className="text-zinc-500">-</span>}
                      </TableCell>
                      <TableCell className="text-sm text-zinc-700 dark:text-zinc-300">
                        {userLabel(event.user_slug, event.user_id)}
                      </TableCell>
                      <TableCell>
                        {event.details && Object.keys(event.details).length > 0 ? (
                          <Button plain onClick={() => setSelectedDetailsEvent(event)}>
                            View
                          </Button>
                        ) : (
                          <span className="text-zinc-500">-</span>
                        )}
                      </TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </div>
          </>
        ) : (
          <Text className="mt-4 text-sm">{t('noProviderActivity')}</Text>
        )}
      </WorkspaceCard>

      <WorkspaceCard>
        <CardHeader eyebrow={t('directoryEyebrow')} title={t('directoryTitle')} />
        <div className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
          {users.map((user) => (
            <div key={user.id} className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
              <div className="font-semibold text-zinc-950 dark:text-white">{user.name}</div>
              <div>{user.slug}</div>
            </div>
          ))}
        </div>
      </WorkspaceCard>

      <Dialog open={selectedDetailsEvent !== null} onClose={() => setSelectedDetailsEvent(null)} size="3xl">
        <DialogTitle>Event details</DialogTitle>
        <DialogBody>
          {selectedDetailsEvent ? (
            <div className="space-y-4">
              <div className="flex flex-wrap items-center gap-2">
                <Badge color={providerBadgeColor(selectedDetailsEvent.provider)}>{selectedDetailsEvent.provider || t('unknown')}</Badge>
                <Badge color={categoryBadgeColor(selectedDetailsEvent.category)}>
                  {selectedDetailsEvent.category === 'uncategorized' || !selectedDetailsEvent.category
                    ? t('uncategorized')
                    : humanizeCategoryLabel(selectedDetailsEvent.category)}
                </Badge>
                <Badge color="zinc">{humanizeEventLabel(selectedDetailsEvent.event)}</Badge>
                {selectedDetailsEvent.status_code ? <Badge color="zinc">{t('httpStatus', { status: selectedDetailsEvent.status_code })}</Badge> : null}
              </div>
              <div className="text-sm text-zinc-700 dark:text-zinc-300">
                {t('user')}: <span className="font-medium text-zinc-950 dark:text-white">{userLabel(selectedDetailsEvent.user_slug, selectedDetailsEvent.user_id)}</span>
              </div>
              <div className="text-sm text-zinc-600 dark:text-zinc-400">
                {formatLocalDateTime(selectedDetailsEvent.created_at) ?? selectedDetailsEvent.created_at}
              </div>
              <pre className="overflow-x-auto rounded-xl bg-white/70 px-3 py-3 text-xs text-zinc-700 dark:bg-black/20 dark:text-zinc-300">
                {JSON.stringify(selectedDetailsEvent.details ?? {}, null, 2)}
              </pre>
            </div>
          ) : null}
        </DialogBody>
        <DialogActions>
          <Button color="dark/zinc" onClick={() => setSelectedDetailsEvent(null)}>
            Close
          </Button>
        </DialogActions>
      </Dialog>
    </div>
  )
}
