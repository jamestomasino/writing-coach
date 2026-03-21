'use client'

import { getAdminAIProviderEvents, getSession, listAdmins, listUsers } from '@/lib/api'
import type { AIProviderEvent, AIProviderEventFilters, AIProviderEventSummary, AuthSession, UserRecord } from '@/lib/types'
import { useEffect, useState } from 'react'

export function useAdminWorkspace() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [session, setSession] = useState<AuthSession | null>(null)
  const [admins, setAdmins] = useState<string[]>([])
  const [users, setUsers] = useState<UserRecord[]>([])
  const [providerSummary, setProviderSummary] = useState<AIProviderEventSummary | null>(null)
  const [providerEvents, setProviderEvents] = useState<AIProviderEvent[]>([])
  const [providerFilters, setProviderFilters] = useState<AIProviderEventFilters | null>(null)
  const [selectedHours, setSelectedHours] = useState('24')
  const [selectedProvider, setSelectedProvider] = useState('')
  const [selectedEvent, setSelectedEvent] = useState('')
  const [loadingProviderActivity, setLoadingProviderActivity] = useState(false)

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
          setProviderFilters(providerData.filters)
          setSelectedHours(String(providerData.filters.hours))
          setSelectedProvider(providerData.filters.provider ?? '')
          setSelectedEvent(providerData.filters.event ?? '')
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

  useEffect(() => {
    if (!session?.is_admin || loading) {
      return
    }

    let cancelled = false

    async function loadProviderActivity() {
      try {
        setLoadingProviderActivity(true)
        const providerData = await getAdminAIProviderEvents(100, Number(selectedHours) || 24, selectedProvider, selectedEvent)
        if (!cancelled) {
          setProviderSummary(providerData.summary)
          setProviderEvents(providerData.events)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load AI provider activity')
        }
      } finally {
        if (!cancelled) {
          setLoadingProviderActivity(false)
        }
      }
    }

    void loadProviderActivity()
    return () => {
      cancelled = true
    }
  }, [loading, selectedEvent, selectedHours, selectedProvider, session?.is_admin])

  return {
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
  }
}
