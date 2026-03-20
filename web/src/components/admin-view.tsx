'use client'

import { FormEvent, useEffect, useState } from 'react'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Field, FieldGroup, Label } from '@/components/fieldset'
import { Input } from '@/components/input'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { getAdminAIProviderEvents, getSession, listAdmins, listUsers, provisionUser } from '@/lib/api'
import type { AIProviderEvent, AIProviderEventSummary, AuthSession, UserRecord } from '@/lib/types'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

function formatLocalTimestamp(value: string) {
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

function humanizeEventLabel(value: string) {
  return value.replaceAll('_', ' ')
}

export function AdminView() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [session, setSession] = useState<AuthSession | null>(null)
  const [admins, setAdmins] = useState<string[]>([])
  const [users, setUsers] = useState<UserRecord[]>([])
  const [providerSummary, setProviderSummary] = useState<AIProviderEventSummary | null>(null)
  const [providerEvents, setProviderEvents] = useState<AIProviderEvent[]>([])
  const [slug, setSlug] = useState('')
  const [name, setName] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const sessionData = await getSession()
        if (!cancelled) {
          setSession(sessionData)
        }
        if (!sessionData.is_admin) {
          return
        }
        const [adminData, userData, providerData] = await Promise.all([listAdmins(), listUsers(), getAdminAIProviderEvents()])
        if (!cancelled) {
          setAdmins(adminData.admins)
          setUsers(userData)
          setProviderSummary(providerData.summary)
          setProviderEvents(providerData.events)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load admin workspace')
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

  async function handleProvision(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    try {
      setSaving(true)
      setError(null)
      await provisionUser({ slug, name })
      setUsers(await listUsers())
      setSlug('')
      setName('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not provision user')
    } finally {
      setSaving(false)
    }
  }

  if (loading) {
    return <LoadingState label="Loading admin workspace…" />
  }
  if (!session?.is_admin) {
    return (
      <EmptyState
        title="Admin workspace unavailable"
        body="Admin access required."
        actionHref="/"
        actionLabel="Back to assignment"
      />
    )
  }
  if (error && users.length === 0) {
    return <AppErrorState title="Admin workspace unavailable" error={error} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Operations"
        title="Admin"
        intro="Manage account access, provision users, and keep an eye on AI provider health."
      />

      {error ? <EmptyState title="Admin action failed" body={error} /> : null}

      <div className="grid gap-8 xl:grid-cols-[1.4fr_1fr]">
        <WorkspaceCard>
          <CardHeader
            eyebrow="Provisioning"
            title="Provision user"
            description="This prepares the internal user record. Browser authentication now runs through the app’s own branded account flows backed by Kratos."
          />
          <form className="mt-5" onSubmit={handleProvision}>
            <FieldGroup>
              <Field>
                <Label>Name</Label>
                <Input value={name} onChange={(event) => setName(event.target.value)} required />
              </Field>
              <Field>
                <Label>Slug</Label>
                <Input value={slug} onChange={(event) => setSlug(event.target.value)} required />
              </Field>
            </FieldGroup>
            <div className="mt-5">
              <Button type="submit" color="dark/zinc" disabled={saving}>
                {saving ? 'Saving…' : 'Provision user'}
              </Button>
            </div>
          </form>
        </WorkspaceCard>

        <WorkspaceCard>
          <CardHeader eyebrow="Access" title="Admin allowlist" />
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {admins.map((admin) => (
              <li key={admin}>• {admin}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>

      <WorkspaceCard>
        <CardHeader
          eyebrow="AI operations"
          title="Provider activity"
          description="Recent validation failures, local throttling, provider resolution, and fallback events."
        />
        {providerSummary ? (
          <>
            <div className="mt-5 grid gap-4 md:grid-cols-4">
              <div className="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 dark:border-white/10 dark:bg-white/5">
                <div className="text-xs uppercase tracking-[0.2em] text-zinc-500">Events</div>
                <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{providerSummary.total}</div>
                <div className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">Since {providerSummary.since}</div>
              </div>
              <div className="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 dark:border-white/10 dark:bg-white/5">
                <div className="text-xs uppercase tracking-[0.2em] text-zinc-500">Validation failures</div>
                <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{providerSummary.validation_failures}</div>
              </div>
              <div className="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 dark:border-white/10 dark:bg-white/5">
                <div className="text-xs uppercase tracking-[0.2em] text-zinc-500">Validation throttles</div>
                <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{providerSummary.validation_rate_limit}</div>
              </div>
              <div className="rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 dark:border-white/10 dark:bg-white/5">
                <div className="text-xs uppercase tracking-[0.2em] text-zinc-500">Fallbacks</div>
                <div className="mt-2 text-2xl font-semibold text-zinc-950 dark:text-white">{providerSummary.fallbacks}</div>
              </div>
            </div>

            <div className="mt-6 grid gap-6 xl:grid-cols-[1fr_1fr]">
              <div>
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">Top providers</div>
                <div className="mt-3 flex flex-wrap gap-2">
                  {providerSummary.provider_counts.map((item) => (
                    <Badge key={item.label} color="zinc">
                      {item.label} {item.count}
                    </Badge>
                  ))}
                </div>
              </div>
              <div>
                <div className="text-sm font-semibold text-zinc-950 dark:text-white">Top categories</div>
                <div className="mt-3 flex flex-wrap gap-2">
                  {providerSummary.category_counts.map((item) => (
                    <Badge key={item.label} color="zinc">
                      {item.label} {item.count}
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
                    <Badge color={event.category === 'auth' || event.category === 'quota' ? 'rose' : event.category === 'local_rate_limit' || event.category === 'rate_limit' ? 'amber' : 'zinc'}>
                      {event.provider || 'unknown'}
                    </Badge>
                    {event.category ? <Badge color="zinc">{event.category}</Badge> : null}
                    {event.status_code ? <Badge color="zinc">HTTP {event.status_code}</Badge> : null}
                  </div>
                  <div className="mt-2 text-sm text-zinc-700 dark:text-zinc-300">
                    User: <span className="font-medium text-zinc-950 dark:text-white">{event.user_slug || `#${event.user_id}`}</span>
                  </div>
                  <div className="mt-1 text-sm text-zinc-600 dark:text-zinc-400">{formatLocalTimestamp(event.created_at)}</div>
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
          <Text className="mt-4 text-sm">No provider activity has been recorded yet.</Text>
        )}
      </WorkspaceCard>

      <WorkspaceCard>
        <CardHeader eyebrow="Directory" title="Provisioned users" />
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
