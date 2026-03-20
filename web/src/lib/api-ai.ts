import { request } from './api-core'
import type { AIProviderSettings } from './types'

export async function getAISettings() {
  const payload = await request<{ settings: AIProviderSettings }>('/api/ai/settings')
  return payload.settings
}

export async function validateAISettings(input: {
  provider: string
  api_key: string
  base_url_override?: string
  prompt_model_override?: string
  review_model_override?: string
  enabled: boolean
}) {
  const payload = await request<{ valid: boolean; settings: AIProviderSettings }>('/api/ai/settings/validate', {
    method: 'POST',
    body: JSON.stringify(input),
  })
  return payload
}

export async function saveAISettings(input: {
  provider: string
  api_key: string
  base_url_override?: string
  prompt_model_override?: string
  review_model_override?: string
  enabled: boolean
}) {
  const payload = await request<{ settings: AIProviderSettings }>('/api/ai/settings', {
    method: 'PUT',
    body: JSON.stringify(input),
  })
  return payload.settings
}

export async function deleteAISettings() {
  const payload = await request<{ deleted: boolean; settings: AIProviderSettings }>('/api/ai/settings', {
    method: 'DELETE',
  })
  return payload
}
