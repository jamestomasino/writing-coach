'use client'

import { getDashboard, getOnboarding, getTree } from '@/lib/api'
import type { Dashboard, OnboardingState, Tree } from '@/lib/types'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import { useEffect, useState } from 'react'

type UseTrackDashboardDataOptions = {
  requireActiveTree?: boolean
  loadErrorMessage: string
}

type TrackDashboardData = {
  dashboard: Dashboard | null
  onboarding: OnboardingState | null
  tree: Tree | null
}

export function useTrackDashboardData(
  path: string,
  { requireActiveTree = false, loadErrorMessage }: UseTrackDashboardDataOptions
) {
  const { session, loading: sessionLoading, error: sessionError } = useRequiredAppSession(path)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [data, setData] = useState<TrackDashboardData>({
    dashboard: null,
    onboarding: null,
    tree: null,
  })

  useEffect(() => {
    let cancelled = false

    async function load() {
      if (!session) {
        return
      }
      try {
        const activeTreeSlug = session.active_tree_slug
        if (requireActiveTree && !activeTreeSlug) {
          throw new Error('No active track selected')
        }

        const [dashboard, onboarding, tree] = await Promise.all([
          getDashboard(),
          getOnboarding(),
          activeTreeSlug ? getTree(activeTreeSlug) : Promise.resolve(null),
        ])

        if (!cancelled) {
          setData({ dashboard, onboarding, tree })
          setError(null)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : loadErrorMessage)
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
  }, [loadErrorMessage, requireActiveTree, session])

  return {
    session,
    sessionLoading,
    sessionError,
    loading,
    error,
    dashboard: data.dashboard,
    onboarding: data.onboarding,
    tree: data.tree,
  }
}
