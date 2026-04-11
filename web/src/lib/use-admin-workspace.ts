'use client'

import { getAdminAIProviderEvents, getSession, listAdmins, listUsers } from '@/lib/api'
import {
  getAdminCalibrationDashboard,
  getAdminPedagogyIntegrity,
  markAdminCalibrationNotificationRead,
  markAdminCalibrationRunRead,
  setAdminCalibrationRunApproval,
  runAdminCalibration,
} from '@/lib/api-admin'
import type {
  AIProviderEvent,
  AIProviderEventFilters,
  AIProviderEventSummary,
  AdminNotification,
  AuthSession,
  CalibrationRun,
  CalibrationSettings,
  PedagogyIntegrity,
  UserRecord,
} from '@/lib/types'
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
  const [calibrationRuns, setCalibrationRuns] = useState<CalibrationRun[]>([])
  const [calibrationNotifications, setCalibrationNotifications] = useState<AdminNotification[]>([])
  const [calibrationUnreadCount, setCalibrationUnreadCount] = useState(0)
  const [calibrationSettings, setCalibrationSettings] = useState<CalibrationSettings | null>(null)
  const [pedagogyIntegrity, setPedagogyIntegrity] = useState<PedagogyIntegrity | null>(null)
  const [loadingCalibration, setLoadingCalibration] = useState(false)
  const [runningCalibration, setRunningCalibration] = useState(false)

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
        const [adminData, userData, providerData, calibrationData, integrityData] = await Promise.all([
          listAdmins(),
          listUsers(),
          getAdminAIProviderEvents(),
          getAdminCalibrationDashboard(),
          getAdminPedagogyIntegrity(),
        ])
        if (!cancelled) {
          setAdmins(adminData.admins)
          setUsers(userData)
          setProviderSummary(providerData.summary)
          setProviderEvents(providerData.events)
          setProviderFilters(providerData.filters)
          setSelectedHours(String(providerData.filters.hours))
          setSelectedProvider(providerData.filters.provider ?? '')
          setSelectedEvent(providerData.filters.event ?? '')
          setCalibrationRuns(calibrationData.runs)
          setCalibrationNotifications(calibrationData.notifications)
          setCalibrationUnreadCount(calibrationData.unread_count)
          setCalibrationSettings(calibrationData.settings)
          setPedagogyIntegrity(integrityData.integrity)
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

  async function refreshCalibrationDashboard() {
    if (!session?.is_admin) {
      return
    }
    try {
      setLoadingCalibration(true)
      const calibrationData = await getAdminCalibrationDashboard()
      const integrityData = await getAdminPedagogyIntegrity()
      setCalibrationRuns(calibrationData.runs)
      setCalibrationNotifications(calibrationData.notifications)
      setCalibrationUnreadCount(calibrationData.unread_count)
      setCalibrationSettings(calibrationData.settings)
      setPedagogyIntegrity(integrityData.integrity)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not load calibration dashboard')
    } finally {
      setLoadingCalibration(false)
    }
  }

  async function triggerCalibrationRun() {
    try {
      setRunningCalibration(true)
      await runAdminCalibration()
      await refreshCalibrationDashboard()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not trigger calibration run')
    } finally {
      setRunningCalibration(false)
    }
  }

  async function markCalibrationNotificationRead(notificationID: number) {
    try {
      await markAdminCalibrationNotificationRead(notificationID)
      await refreshCalibrationDashboard()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not mark notification as read')
    }
  }

  async function markCalibrationRunRead(runID: number) {
    try {
      await markAdminCalibrationRunRead(runID)
      await refreshCalibrationDashboard()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not mark run as read')
    }
  }

  async function setCalibrationRunApproval(runID: number, status: 'pending' | 'approved' | 'rejected', notes = '') {
    try {
      await setAdminCalibrationRunApproval(runID, status, notes)
      await refreshCalibrationDashboard()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not update run approval')
    }
  }

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
    calibrationRuns,
    calibrationNotifications,
    calibrationUnreadCount,
    calibrationSettings,
    pedagogyIntegrity,
    loadingCalibration,
    runningCalibration,
    triggerCalibrationRun,
    markCalibrationNotificationRead,
    markCalibrationRunRead,
    setCalibrationRunApproval,
    refreshCalibrationDashboard,
  }
}
