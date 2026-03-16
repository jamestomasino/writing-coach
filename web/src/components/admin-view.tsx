'use client'

import { FormEvent, useEffect, useState } from 'react'
import { Button } from '@/components/button'
import { Field, FieldGroup, Label } from '@/components/fieldset'
import { Heading, Subheading } from '@/components/heading'
import { Input } from '@/components/input'
import { Text } from '@/components/text'
import { listAdmins, listUsers, provisionUser } from '@/lib/api'
import type { UserRecord } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function AdminView() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [admins, setAdmins] = useState<string[]>([])
  const [users, setUsers] = useState<UserRecord[]>([])
  const [slug, setSlug] = useState('')
  const [name, setName] = useState('')
  const [saving, setSaving] = useState(false)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const [adminData, userData] = await Promise.all([listAdmins(), listUsers()])
        if (!cancelled) {
          setAdmins(adminData.admins)
          setUsers(userData)
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
  if (error && users.length === 0) {
    return <EmptyState title="Admin workspace unavailable" body={error} actionHref="/" actionLabel="Back to assignment" />
  }

  return (
    <div className="space-y-8">
      <header>
        <Heading>Admin</Heading>
        <Text className="mt-2 max-w-3xl">
          Admin scope stays intentionally narrow. For now, this screen provisions user records ahead of first login and exposes the current admin allowlist.
        </Text>
      </header>

      {error ? <EmptyState title="Admin action failed" body={error} /> : null}

      <div className="grid gap-8 xl:grid-cols-[1.4fr_1fr]">
        <WorkspaceCard>
          <Subheading>Provision user</Subheading>
          <Text className="mt-2">This prepares the internal user record. Browser authentication still runs through Kratos registration and login.</Text>
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
          <Subheading>Admin allowlist</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {admins.map((admin) => (
              <li key={admin}>• {admin}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>

      <WorkspaceCard>
        <Subheading>Provisioned users</Subheading>
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
