'use client'

import { deleteAISettings, getAISettings, saveAISettings, validateAISettings } from '@/lib/api'
import type { AIProviderSettings } from '@/lib/types'
import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'

export const providerOptions = [
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'groq', label: 'Groq' },
  { value: 'xai', label: 'xAI' },
]

export const providerMetadata = {
  anthropic: {
    defaultBaseURL: 'https://api.anthropic.com/v1',
    promptModel: 'claude-sonnet-4-20250514',
    reviewModel: 'claude-sonnet-4-20250514',
    promptModels: ['claude-sonnet-4-20250514'],
    reviewModels: ['claude-sonnet-4-20250514'],
    apiKeyHint: 'Anthropic API key',
    note: 'Uses Anthropic’s native Messages API with structured tool output.',
    modelHelp:
      'Claude Sonnet is the recommended default here. Swap it only if you know the replacement supports the same structured outputs.',
    validationHelp:
      'Validation checks the Anthropic API with this key and endpoint. Keep the default base URL unless you are routing through a compatible proxy.',
  },
  gemini: {
    defaultBaseURL: 'https://generativelanguage.googleapis.com/v1beta',
    promptModel: 'gemini-2.5-flash',
    reviewModel: 'gemini-2.5-flash',
    promptModels: ['gemini-2.5-flash', 'gemini-2.5-pro'],
    reviewModels: ['gemini-2.5-flash', 'gemini-2.5-pro'],
    apiKeyHint: 'Gemini API key',
    note: 'Uses Gemini’s native generateContent API with JSON schema output.',
    modelHelp:
      'Gemini 2.5 Flash is the light default. If you switch models, make sure the replacement still handles structured JSON responses well.',
    validationHelp:
      'Validation checks the Gemini API with this key and endpoint. The default base URL is the normal Google Generative Language endpoint.',
  },
  openai: {
    defaultBaseURL: 'https://api.openai.com/v1',
    promptModel: 'gpt-5-mini',
    reviewModel: 'gpt-5-mini',
    promptModels: ['gpt-5-mini', 'gpt-5'],
    reviewModels: ['gpt-5-mini', 'gpt-5'],
    apiKeyHint: 'OpenAI API key',
    note: 'Uses the OpenAI Responses API directly.',
    modelHelp:
      'GPT-5 Mini is the default for both prompt and review generation. You can override it, but use a model that supports structured responses.',
    validationHelp:
      'Validation checks the OpenAI API with this key and endpoint. Leave the base URL alone unless you are intentionally routing to a compatible gateway.',
  },
  groq: {
    defaultBaseURL: 'https://api.groq.com/openai/v1',
    promptModel: 'gpt-5-mini',
    reviewModel: 'gpt-5-mini',
    promptModels: ['gpt-5-mini'],
    reviewModels: ['gpt-5-mini'],
    apiKeyHint: 'Groq API key',
    note: 'Uses Groq’s OpenAI-compatible endpoint.',
    modelHelp:
      'Use a Groq-hosted model that behaves well with OpenAI-style structured responses. Replace the default only if you know the target model is available on your Groq account.',
    validationHelp:
      'Validation checks Groq’s OpenAI-compatible endpoint. If you override the base URL, it needs to speak the same API shape.',
  },
  xai: {
    defaultBaseURL: 'https://api.x.ai/v1',
    promptModel: 'gpt-5-mini',
    reviewModel: 'gpt-5-mini',
    promptModels: ['gpt-5-mini'],
    reviewModels: ['gpt-5-mini'],
    apiKeyHint: 'xAI API key',
    note: 'Uses xAI’s OpenAI-compatible endpoint.',
    modelHelp:
      'Use an xAI model that supports the OpenAI-compatible response shape expected by the app. Override the model only if you have a specific supported target in mind.',
    validationHelp:
      'Validation checks xAI’s OpenAI-compatible endpoint. Keep the default base URL unless you are routing through a compatible proxy.',
  },
} as const

export const DEFAULT_MODEL_OPTION = '__default__'
export const CUSTOM_MODEL_OPTION = '__custom__'

type AIProviderSettingsControllerOptions = {
  required?: boolean
  nextPath?: string
}

type SaveInput = {
  provider: string
  api_key: string
  base_url_override: string
  prompt_model_override: string
  review_model_override: string
  enabled: boolean
}

export function useAIProviderSettings({ required = false, nextPath }: AIProviderSettingsControllerOptions) {
  const router = useRouter()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [settings, setSettings] = useState<AIProviderSettings | null>(null)
  const [provider, setProvider] = useState('openai')
  const [apiKey, setAPIKey] = useState('')
  const [baseURL, setBaseURL] = useState('')
  const [promptModel, setPromptModel] = useState('')
  const [reviewModel, setReviewModel] = useState('')
  const [customPromptModel, setCustomPromptModel] = useState(false)
  const [customReviewModel, setCustomReviewModel] = useState(false)
  const [enabled, setEnabled] = useState(true)
  const [saving, setSaving] = useState(false)
  const [validating, setValidating] = useState(false)

  const selectedProvider = providerMetadata[provider as keyof typeof providerMetadata] ?? providerMetadata.openai

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const current = await getAISettings()
        if (cancelled) {
          return
        }
        setSettings(current)
        setProvider(current.provider ?? 'openai')
        setBaseURL(current.base_url_override ?? '')
        setPromptModel(current.prompt_model_override ?? '')
        setReviewModel(current.review_model_override ?? '')
        setEnabled(current.enabled)
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load AI settings')
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
    if (promptModel.trim() === '') {
      setCustomPromptModel(false)
    } else if (!(selectedProvider.promptModels as readonly string[]).includes(promptModel.trim())) {
      setCustomPromptModel(true)
    }

    if (reviewModel.trim() === '') {
      setCustomReviewModel(false)
    } else if (!(selectedProvider.reviewModels as readonly string[]).includes(reviewModel.trim())) {
      setCustomReviewModel(true)
    }
  }, [promptModel, reviewModel, selectedProvider.promptModels, selectedProvider.reviewModels])

  async function validateConnection() {
    setValidating(true)
    setError(null)
    try {
      return await validateAISettings({
        provider,
        api_key: apiKey,
        base_url_override: baseURL,
        prompt_model_override: promptModel,
        review_model_override: reviewModel,
        enabled,
      })
    } finally {
      setValidating(false)
    }
  }

  async function save() {
    setSaving(true)
    setError(null)
    try {
      const current = await saveAISettings({
        provider,
        api_key: apiKey,
        base_url_override: baseURL,
        prompt_model_override: promptModel,
        review_model_override: reviewModel,
        enabled,
      })
      setSettings(current)
      setAPIKey('')
      if (required && current.ready) {
        router.push(nextPath && nextPath.startsWith('/') ? nextPath : '/')
      }
      return current
    } finally {
      setSaving(false)
    }
  }

  async function remove() {
    setSaving(true)
    setError(null)
    try {
      const result = await deleteAISettings()
      setSettings(result.settings)
      setProvider('openai')
      setAPIKey('')
      setBaseURL('')
      setPromptModel('')
      setReviewModel('')
      setEnabled(true)
      setCustomPromptModel(false)
      setCustomReviewModel(false)
      return result
    } finally {
      setSaving(false)
    }
  }

  function applyDefaults() {
    setBaseURL(selectedProvider.defaultBaseURL)
    setPromptModel('')
    setReviewModel('')
    setCustomPromptModel(false)
    setCustomReviewModel(false)
  }

  function handlePromptModelChange(value: string) {
    if (value === DEFAULT_MODEL_OPTION) {
      setPromptModel('')
      setCustomPromptModel(false)
      return
    }
    if (value === CUSTOM_MODEL_OPTION) {
      setCustomPromptModel(true)
      if ((selectedProvider.promptModels as readonly string[]).includes(promptModel)) {
        setPromptModel('')
      }
      return
    }
    setPromptModel(value)
    setCustomPromptModel(false)
  }

  function handleReviewModelChange(value: string) {
    if (value === DEFAULT_MODEL_OPTION) {
      setReviewModel('')
      setCustomReviewModel(false)
      return
    }
    if (value === CUSTOM_MODEL_OPTION) {
      setCustomReviewModel(true)
      if ((selectedProvider.reviewModels as readonly string[]).includes(reviewModel)) {
        setReviewModel('')
      }
      return
    }
    setReviewModel(value)
    setCustomReviewModel(false)
  }

  return {
    loading,
    error,
    settings,
    provider,
    setProvider,
    apiKey,
    setAPIKey,
    baseURL,
    setBaseURL,
    promptModel,
    setPromptModel,
    reviewModel,
    setReviewModel,
    customPromptModel,
    customReviewModel,
    enabled,
    setEnabled,
    saving,
    validating,
    selectedProvider,
    validateConnection,
    save,
    remove,
    applyDefaults,
    handlePromptModelChange,
    handleReviewModelChange,
  }
}
