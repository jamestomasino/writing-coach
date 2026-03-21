'use client'

import { Badge } from '@/components/badge'
import { CardHeader } from '@/components/card-header'
import { Field, FieldGroup, Label } from '@/components/fieldset'
import { PageHeader } from '@/components/page-header'
import { Select } from '@/components/select'
import { Text } from '@/components/text'
import { formatLocalDateTime } from '@/lib/datetime'
import { useAdminWorkspace } from '@/lib/use-admin-workspace'
import { useTranslations } from 'next-intl'
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

function userLabel(userSlug: string | undefined, userID: number) {
  if (userSlug && userSlug.trim() !== '') {
    return `${userSlug} · #${userID}`
  }
  return `#${userID}`
}

export function AdminView() {
  const t = useTranslations('adminView')
  const {
    loading,
    error,
    setError,
    session,
    admins,
    users,
    providerSummary,
    providerEvents,
    providerFilters,
    selectedHours,
    setSelectedHours,
    selectedProvider,
    setSelectedProvider,
    selectedEvent,
    setSelectedEvent,
    loadingProviderActivity,
  } = useAdminWorkspace()

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
            <Select value={selectedHours} onChange={(event) => setSelectedHours(event.target.value)}>
              <option value="1">{t('lastHour')}</option>
              <option value="6">{t('last6Hours')}</option>
              <option value="24">{t('last24Hours')}</option>
              <option value="72">{t('last3Days')}</option>
              <option value="168">{t('last7Days')}</option>
            </Select>
          </Field>
          <Field>
            <Label>{t('provider')}</Label>
            <Select value={selectedProvider} onChange={(event) => setSelectedProvider(event.target.value)}>
              <option value="">{t('allProviders')}</option>
              {providerFilters?.providers.map((provider) => (
                <option key={provider} value={provider}>
                  {provider}
                </option>
              ))}
            </Select>
          </Field>
          <Field>
            <Label>{t('eventType')}</Label>
            <Select value={selectedEvent} onChange={(event) => setSelectedEvent(event.target.value)}>
              <option value="">{t('allEventTypes')}</option>
              {providerFilters?.events.map((eventName) => (
                <option key={eventName} value={eventName}>
                  {humanizeEventLabel(eventName)}
                </option>
              ))}
            </Select>
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
                  {providerSummary.provider_counts.map((item) => (
                    <Badge key={item.label} color="zinc">
                      {item.label} {item.count}
                    </Badge>
                  ))}
                </div>
              </div>
              <div>
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">{t('topCategories')}</div>
                <div className="mt-3 flex flex-wrap gap-2">
                  {providerSummary.category_counts.map((item) => (
                    <Badge key={item.label} color="zinc">
                      {(item.label === 'uncategorized' ? t('uncategorized') : humanizeCategoryLabel(item.label))} {item.count}
                    </Badge>
                  ))}
                </div>
              </div>
            </div>

            <div className="mt-6 space-y-3">
              {providerEvents.map((event) => (
                <div key={event.id} className="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 dark:border-white/10 dark:bg-white/5">
                  <div className="flex flex-wrap items-center gap-2">
                    <div className="font-semibold text-zinc-950 dark:text-white">{humanizeEventLabel(event.event)}</div>
                    <Badge color={categoryBadgeColor(event.category)}>
                      {event.provider || t('unknown')}
                    </Badge>
                    {event.category ? <Badge color={categoryBadgeColor(event.category)}>{event.category === 'uncategorized' ? t('uncategorized') : humanizeCategoryLabel(event.category)}</Badge> : null}
                    {event.status_code ? <Badge color="zinc">{t('httpStatus', { status: event.status_code })}</Badge> : null}
                  </div>
                  <div className="mt-2 text-sm text-zinc-700 dark:text-zinc-300">
                    {t('user')}: <span className="font-medium text-zinc-950 dark:text-white">{userLabel(event.user_slug, event.user_id)}</span>
                  </div>
                  <div className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">
                    {formatLocalDateTime(event.created_at) ?? event.created_at}
                  </div>
                  {event.details && Object.keys(event.details).length > 0 ? (
                    <pre className="mt-3 overflow-x-auto rounded-xl bg-white/70 px-3 py-3 text-xs text-zinc-700 dark:bg-black/20 dark:text-zinc-300">
                      {JSON.stringify(event.details, null, 2)}
                    </pre>
                  ) : null}
                </div>
              ))}
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
    </div>
  )
}
