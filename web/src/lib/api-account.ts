import { request } from './api-core'

export function resetAccountData() {
  return request<{ ok: boolean }>('/api/account/reset', {
    method: 'POST',
    body: JSON.stringify({}),
  })
}
