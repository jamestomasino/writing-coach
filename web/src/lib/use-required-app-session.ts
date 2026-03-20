'use client'

import { getSession } from '@/lib/api'
import { requiredSetupPath } from '@/lib/onboarding-funnel'
import type { AuthSession } from '@/lib/types'
import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'

type Result = {
  session: AuthSession | null
  loading: boolean
  error: string | null
}

type State = Result & {
  session: AuthSession | null
  loading: boolean
  error: string | null
  pathname: string
}

export function useRequiredAppSession(pathname: string): Result {
  const router = useRouter()
  const [state, setState] = useState<State>({
    session: null,
    loading: true,
    error: null,
    pathname,
  })

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const session = await getSession()
        if (cancelled) {
          return
        }
        if (!session.authenticated) {
          router.replace('/about')
          return
        }
        const nextPath = requiredSetupPath(session, pathname)
        if (nextPath) {
          router.replace(nextPath)
          return
        }
        setState({
          session,
          loading: false,
          error: null,
          pathname,
        })
      } catch (err) {
        if (cancelled) {
          return
        }
        setState({
          session: null,
          loading: false,
          error: err instanceof Error ? err.message : 'Could not load session',
          pathname,
        })
      }
    }

    void load()

    return () => {
      cancelled = true
    }
  }, [pathname, router])

  if (state.pathname !== pathname) {
    return {
      session: null,
      loading: true,
      error: null,
    }
  }

  return {
    session: state.session,
    loading: state.loading,
    error: state.error,
  }
}
