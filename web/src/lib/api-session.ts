import { request } from './api-core'
import type { AuthSession } from './types'

export function getSession() {
  return request<AuthSession>('/api/auth/session')
}
