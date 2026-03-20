'use client'

import { useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { CurrentAssignmentView } from '@/components/current-assignment-view'
import { LoadingState } from '@/components/status-state'
import { getSession } from '@/lib/api'
import { requiredSetupPath } from '@/lib/onboarding-funnel'

export default function Home() {
  const router = useRouter()
  const [authenticated, setAuthenticated] = useState<boolean | null>(null)

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const session = await getSession()
        if (cancelled) {
          return
        }
        if (session.authenticated) {
          const nextPath = requiredSetupPath(session, '/')
          if (nextPath) {
            router.replace(nextPath)
            return
          }
          setAuthenticated(true)
          return
        }
        setAuthenticated(false)
        router.replace('/about')
      } catch {
        if (cancelled) {
          return
        }
        setAuthenticated(false)
        router.replace('/about')
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [router])

  if (authenticated !== true) {
    return <LoadingState label="Loading workspace…" />
  }

  return (
    <CurrentAssignmentView />
  )
}
