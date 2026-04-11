import { request } from './api-core'
import type {
  AIProviderEvent,
  AIProviderEventFilters,
  AIProviderEventSummary,
  AdminNotification,
  CalibrationRun,
  CalibrationSettings,
  PedagogyIntegrity,
  UserRecord,
} from './types'

export function listAdmins() {
  return request<{ admins: string[] }>('/api/admins')
}

export async function listUsers() {
  const payload = await request<{ users: UserRecord[] }>('/api/users')
  return payload.users
}

export async function provisionUser(input: { slug: string; name: string }) {
  return request('/api/users', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}

export async function getAdminAIProviderEvents(limit = 100, hours = 24, provider = '', event = '') {
  const params = new URLSearchParams({
    limit: String(limit),
    hours: String(hours),
  })
  if (provider) {
    params.set('provider', provider)
  }
  if (event) {
    params.set('event', event)
  }
  return request<{ summary: AIProviderEventSummary; events: AIProviderEvent[]; filters: AIProviderEventFilters }>(
    `/api/admin/ai-provider-events?${params.toString()}`
  )
}

export async function getAdminCalibrationDashboard(limit = 20, notificationsLimit = 50) {
  const params = new URLSearchParams({
    limit: String(limit),
    notifications_limit: String(notificationsLimit),
  })
  return request<{
    runs: CalibrationRun[]
    notifications: AdminNotification[]
    unread_count: number
    settings: CalibrationSettings
  }>(`/api/admin/calibration?${params.toString()}`)
}

export async function runAdminCalibration() {
  return request<{ run: CalibrationRun }>('/api/admin/calibration/run', { method: 'POST' })
}

export async function getAdminPedagogyIntegrity(hours = 24 * 7) {
  const params = new URLSearchParams({
    hours: String(hours),
  })
  return request<{ integrity: PedagogyIntegrity }>(`/api/admin/pedagogy-integrity?${params.toString()}`)
}

export async function markAdminCalibrationNotificationRead(id: number) {
  return request(`/api/admin/calibration/notifications/${id}/read`, { method: 'POST' })
}

export async function markAdminCalibrationRunRead(runID: number) {
  return request(`/api/admin/calibration/runs/${runID}/read`, { method: 'POST' })
}

export async function setAdminCalibrationRunApproval(runID: number, status: 'pending' | 'approved' | 'rejected', notes = '') {
  return request(`/api/admin/calibration/runs/${runID}/approval`, {
    method: 'POST',
    body: JSON.stringify({ status, notes }),
  })
}
