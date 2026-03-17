'use client'

import { useEffect, useState } from 'react'
import { EmptyState, LoadingState } from '@/components/status-state'

export const dynamic = 'force-dynamic'

export default function LogoutPage() {
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function logout() {
      try {
        const response = await fetch('/.ory/kratos/public/self-service/logout/browser', {
          credentials: 'include',
          headers: {
            Accept: 'application/json',
          },
        })
        if (!response.ok) {
          throw new Error(`${response.status} ${response.statusText}`)
        }
        const payload = (await response.json()) as { logout_url?: string }
        if (!payload.logout_url) {
          throw new Error('Kratos did not return a logout URL.')
        }
        if (!cancelled) {
          window.location.replace(payload.logout_url)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not sign out.')
        }
      }
    }

    void logout()
    return () => {
      cancelled = true
    }
  }, [])

  if (error) {
    return (
      <EmptyState
        title="Sign out unavailable"
        body={error}
        actionHref="/"
        actionLabel="Back to home"
      />
    )
  }

  return <LoadingState label="Signing out…" />
}
