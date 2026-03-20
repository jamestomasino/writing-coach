import { request } from './api-core'
import type { AIProviderEvent, AIProviderEventFilters, AIProviderEventSummary, UserRecord } from './types'

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
