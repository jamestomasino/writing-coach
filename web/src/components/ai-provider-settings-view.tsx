'use client'

import { FormEvent, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { Callout } from '@/components/callout'
import { CardHeader } from '@/components/card-header'
import { Checkbox, CheckboxField } from '@/components/checkbox'
import { Field, FieldGroup, Label } from '@/components/fieldset'
import { Input } from '@/components/input'
import { PageHeader } from '@/components/page-header'
import { Select } from '@/components/select'
import { Text } from '@/components/text'
import { deleteAISettings, getAISettings, saveAISettings, validateAISettings } from '@/lib/api'
import type { AIProviderSettings } from '@/lib/types'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
import { useToast } from './toast-provider'
import { WorkspaceCard } from './workspace-card'

const providerOptions = [
  { value: 'anthropic', label: 'Anthropic' },
  { value: 'gemini', label: 'Gemini' },
  { value: 'openai', label: 'OpenAI' },
  { value: 'groq', label: 'Groq' },
  { value: 'xai', label: 'xAI' },
]

const providerMetadata = {
  anthropic: {
    defaultBaseURL: 'https://api.anthropic.com/v1',
    promptModel: 'claude-sonnet-4-20250514',
    reviewModel: 'claude-sonnet-4-20250514',
    apiKeyHint: 'Anthropic API key',
    note: 'Uses Anthropic’s native Messages API with structured tool output.',
    modelHelp: 'Claude Sonnet is the recommended default here. Swap it only if you know the replacement supports the same structured outputs.',
    validationHelp: 'Validation checks the Anthropic API with this key and endpoint. Keep the default base URL unless you are routing through a compatible proxy.',
  },
  gemini: {
    defaultBaseURL: 'https://generativelanguage.googleapis.com/v1beta',
    promptModel: 'gemini-2.5-flash',
    reviewModel: 'gemini-2.5-flash',
    apiKeyHint: 'Gemini API key',
    note: 'Uses Gemini’s native generateContent API with JSON schema output.',
    modelHelp: 'Gemini 2.5 Flash is the light default. If you switch models, make sure the replacement still handles structured JSON responses well.',
    validationHelp: 'Validation checks the Gemini API with this key and endpoint. The default base URL is the normal Google Generative Language endpoint.',
  },
  openai: {
    defaultBaseURL: 'https://api.openai.com/v1',
    promptModel: 'gpt-5-mini',
    reviewModel: 'gpt-5-mini',
    apiKeyHint: 'OpenAI API key',
    note: 'Uses the OpenAI Responses API directly.',
    modelHelp: 'GPT-5 Mini is the default for both prompt and review generation. You can override it, but use a model that supports structured responses.',
    validationHelp: 'Validation checks the OpenAI API with this key and endpoint. Leave the base URL alone unless you are intentionally routing to a compatible gateway.',
  },
  groq: {
    defaultBaseURL: 'https://api.groq.com/openai/v1',
    promptModel: 'gpt-5-mini',
    reviewModel: 'gpt-5-mini',
    apiKeyHint: 'Groq API key',
    note: 'Uses Groq’s OpenAI-compatible endpoint.',
    modelHelp: 'Use a Groq-hosted model that behaves well with OpenAI-style structured responses. Replace the default only if you know the target model is available on your Groq account.',
    validationHelp: 'Validation checks Groq’s OpenAI-compatible endpoint. If you override the base URL, it needs to speak the same API shape.',
  },
  xai: {
    defaultBaseURL: 'https://api.x.ai/v1',
    promptModel: 'gpt-5-mini',
    reviewModel: 'gpt-5-mini',
    apiKeyHint: 'xAI API key',
    note: 'Uses xAI’s OpenAI-compatible endpoint.',
    modelHelp: 'Use an xAI model that supports the OpenAI-compatible response shape expected by the app. Override the model only if you have a specific supported target in mind.',
    validationHelp: 'Validation checks xAI’s OpenAI-compatible endpoint. Keep the default base URL unless you are routing through a compatible proxy.',
  },
} as const

function providerLabel(value?: string) {
  return providerOptions.find((item) => item.value === value)?.label ?? value ?? 'System provider'
}

function providerStatus(settings: AIProviderSettings | null) {
  if (settings && !settings.personal_provider_storage_available) {
    return {
      label: settings.system_fallback ? 'System provider' : 'Storage unavailable',
      detail: settings.system_fallback
        ? 'The app can still use the shared provider, but saving personal provider keys is disabled on this server.'
        : 'Saving personal provider keys is disabled on this server and no shared provider is available.',
      badgeColor: settings.system_fallback ? ('amber' as const) : ('rose' as const),
    }
  }
  if (!settings?.ready) {
    return {
      label: 'Setup required',
      detail: 'Add a personal provider or enable a shared fallback before generation can run.',
      badgeColor: 'rose' as const,
    }
  }
  if (settings.enabled && settings.provider) {
    return {
      label: providerLabel(settings.provider),
      detail: 'Your personal provider is active for future generation.',
      badgeColor: 'emerald' as const,
    }
  }
  return {
    label: 'System provider',
    detail: settings?.system_fallback ? 'The app will use the shared provider while it remains available.' : 'No shared provider is available.',
    badgeColor: settings?.system_fallback ? ('amber' as const) : ('rose' as const),
  }
}

function classifyProviderIssue(message: string | null) {
  const text = (message ?? '').toLowerCase()
  if (!text) {
    return null
  }
  if (text.includes('rejected this api key')) {
    return {
      title: 'API key rejected',
      body: 'This provider did not accept the key. Check that you copied the full key and that it belongs to the selected provider.',
      tone: 'danger' as const,
    }
  }
  if (text.includes('out of quota') || text.includes('billing')) {
    return {
      title: 'Quota or billing issue',
      body: 'The provider account cannot be used right now. Check quota, credits, or billing on the provider side.',
      tone: 'warning' as const,
    }
  }
  if (text.includes('rate-limiting')) {
    return {
      title: 'Provider is rate-limiting requests',
      body: 'Try again in a moment, or switch to another provider if you have one.',
      tone: 'warning' as const,
    }
  }
  if (text.includes('temporarily unavailable') || text.includes('timed out')) {
    return {
      title: 'Provider is temporarily unavailable',
      body: 'This usually means an endpoint or upstream issue. Confirm the base URL and try again shortly.',
      tone: 'warning' as const,
    }
  }
  return {
    title: 'Provider issue',
    body: message ?? 'The provider could not be used right now.',
    tone: 'danger' as const,
  }
}

function formatLocalTimestamp(value?: string) {
  if (!value) {
    return null
  }
  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(date)
}

export function AIProviderSettingsView({ required = false, nextPath }: { required?: boolean; nextPath?: string }) {
  const router = useRouter()
  const toast = useToast()
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [settings, setSettings] = useState<AIProviderSettings | null>(null)
  const [provider, setProvider] = useState('openai')
  const [apiKey, setAPIKey] = useState('')
  const [baseURL, setBaseURL] = useState('')
  const [promptModel, setPromptModel] = useState('')
  const [reviewModel, setReviewModel] = useState('')
  const [enabled, setEnabled] = useState(true)
  const [saving, setSaving] = useState(false)
  const [validating, setValidating] = useState(false)
  const status = providerStatus(settings)
  const validatedAt = formatLocalTimestamp(settings?.validated_at)
  const personalStorageAvailable = settings?.personal_provider_storage_available ?? true
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

  async function handleValidate() {
    try {
      setValidating(true)
      setError(null)
      const result = await validateAISettings({
        provider,
        api_key: apiKey,
        base_url_override: baseURL,
        prompt_model_override: promptModel,
        review_model_override: reviewModel,
        enabled,
      })
      toast.success(`Connection looks good for ${providerLabel(result.settings.provider)}.`, 'Provider check')
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Could not validate AI settings'
      const details = classifyProviderIssue(message)
      toast.error(details?.body ?? message, details?.title ?? 'Provider issue')
    } finally {
      setValidating(false)
    }
  }

  async function handleSave(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    try {
      setSaving(true)
      setError(null)
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
      toast.success('AI provider settings saved.')
      if (required && current.ready) {
        router.push(nextPath && nextPath.startsWith('/') ? nextPath : '/')
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Could not save AI settings'
      const details = classifyProviderIssue(message)
      toast.error(details?.body ?? message, details?.title ?? 'Provider issue')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    try {
      setSaving(true)
      setError(null)
      const result = await deleteAISettings()
      setSettings(result.settings)
      setProvider('openai')
      setAPIKey('')
      setBaseURL('')
      setPromptModel('')
      setReviewModel('')
      setEnabled(true)
      toast.success('Personal provider removed. The app will use the system provider when available.')
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Could not remove AI settings'
      toast.error(message, 'Provider issue')
    } finally {
      setSaving(false)
    }
  }

  function handleApplyDefaults() {
    setBaseURL(selectedProvider.defaultBaseURL)
    setPromptModel(selectedProvider.promptModel)
    setReviewModel(selectedProvider.reviewModel)
  }

  if (loading) {
    return <LoadingState label="Loading AI provider settings…" />
  }
  if (error && !settings) {
    return <AppErrorState title="AI settings unavailable" error={error} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Settings"
        title="AI provider"
        intro="Connect your own provider credentials for assignment and review generation. You can keep using the shared system provider while it remains available."
        actions={
          <Button href="/" outline>
            Current assignment
          </Button>
        }
      />

      <WorkspaceCard>
        <CardHeader eyebrow="Status" title="Current provider status" />
        <div className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
          <div className="flex items-center justify-between gap-3">
            <p>
              Effective provider: <span className="font-semibold text-zinc-950 dark:text-white">{status.label}</span>
            </p>
            <Badge color={status.badgeColor}>{settings?.enabled && settings?.provider ? 'Personal' : settings?.system_fallback ? 'Shared' : 'Needed'}</Badge>
          </div>
          <p>{status.detail}</p>
          <p>
            System fallback: <span className="font-semibold text-zinc-950 dark:text-white">{settings?.system_fallback ? 'available' : 'not available'}</span>
          </p>
          {settings?.has_key ? (
            <p>
              Saved key: <span className="font-semibold text-zinc-950 dark:text-white">••••{settings.key_last4}</span>
            </p>
          ) : null}
          {validatedAt ? (
            <p>
              Last checked: <span className="font-semibold text-zinc-950 dark:text-white">{validatedAt}</span>
            </p>
          ) : null}
        </div>
      </WorkspaceCard>

      {!settings?.ready ? (
        <Callout title="AI setup required">
          Add a provider key here before trying to generate assignments or reviews.
        </Callout>
      ) : null}

      {!personalStorageAvailable ? (
        <Callout
          title="Personal provider storage is unavailable"
          body={settings?.system_fallback
            ? 'This server is not configured to store personal provider keys. You can keep using the shared provider, but users cannot save their own keys here until the server sets WRITING_COACH_AI_KEY_SECRET.'
            : 'This server is not configured to store personal provider keys, and no shared provider fallback is available. An operator needs to set WRITING_COACH_AI_KEY_SECRET before users can save their own keys.'}
          tone={settings?.system_fallback ? 'warning' : 'danger'}
        />
      ) : null}

      {required && !settings?.ready ? (
        <Callout title="Setup needed to continue">
          This workspace needs an AI provider before assignment and feedback generation can run.
        </Callout>
      ) : null}

      {settings?.last_validation_error ? (
        <Callout
          title="Last validation issue"
          body={settings.last_validation_error}
          tone="warning"
        />
      ) : null}

      {personalStorageAvailable ? (
        <WorkspaceCard>
          <form className="space-y-8" onSubmit={handleSave}>
            <CardHeader
              eyebrow="Connection"
              title="Personal provider"
              description="Supported providers include Anthropic, Gemini, OpenAI, Groq, and xAI. Advanced fields are optional."
            />

            <Callout
              title={`${providerLabel(provider)} defaults`}
              body={`${selectedProvider.note} Suggested base URL: ${selectedProvider.defaultBaseURL}. Suggested prompt model: ${selectedProvider.promptModel}. Suggested review model: ${selectedProvider.reviewModel}.`}
              tone="info"
              actions={
                <Button type="button" outline onClick={handleApplyDefaults}>
                  Use suggested defaults
                </Button>
              }
            />

            <FieldGroup>
              <Field>
                <Label>Provider</Label>
                <Select value={provider} onChange={(event) => setProvider(event.target.value)}>
                  {providerOptions.map((item) => (
                    <option key={item.value} value={item.value}>
                      {item.label}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field>
                <Label>API key</Label>
                <Input
                  type="password"
                  value={apiKey}
                  onChange={(event) => setAPIKey(event.target.value)}
                  placeholder={settings?.has_key ? `Enter a new key to replace the saved ${selectedProvider.apiKeyHint}` : `Paste your ${selectedProvider.apiKeyHint}`}
                />
                <Text className="mt-2 text-sm">
                  {settings?.has_key ? 'Leave this blank to keep using the saved key. Keys are encrypted before they are stored.' : 'Keys are encrypted before they are stored.'}
                </Text>
                <Text className="mt-2 text-sm">{selectedProvider.validationHelp}</Text>
              </Field>
            </FieldGroup>

            <FieldGroup>
              <Field>
                <Label>Base URL override</Label>
                <Input value={baseURL} onChange={(event) => setBaseURL(event.target.value)} placeholder={selectedProvider.defaultBaseURL} />
                <Text className="mt-2 text-sm">Leave this blank unless you need a proxy or compatible custom endpoint.</Text>
              </Field>
              <Field>
                <Label>Prompt model override</Label>
                <Input value={promptModel} onChange={(event) => setPromptModel(event.target.value)} placeholder={selectedProvider.promptModel} />
              </Field>
              <Field>
                <Label>Review model override</Label>
                <Input value={reviewModel} onChange={(event) => setReviewModel(event.target.value)} placeholder={selectedProvider.reviewModel} />
              </Field>
            </FieldGroup>

            <Text className="text-sm">{selectedProvider.modelHelp}</Text>

            <CheckboxField>
              <Checkbox checked={enabled} onChange={setEnabled} />
              <Label>Use this provider for future generation actions</Label>
            </CheckboxField>

            <div className="flex flex-wrap gap-3">
              <Button type="button" outline onClick={handleValidate} disabled={validating || (apiKey.trim() === '' && !settings?.has_key)}>
                {validating ? 'Checking provider…' : 'Validate connection'}
              </Button>
              <Button type="submit" color="dark/zinc" disabled={saving || (apiKey.trim() === '' && !settings?.has_key)}>
                {saving ? 'Saving…' : 'Save provider'}
              </Button>
              {settings?.has_key ? (
                <Button type="button" plain onClick={handleDelete} disabled={saving}>
                  Remove personal provider
                </Button>
              ) : null}
            </div>
          </form>
        </WorkspaceCard>
      ) : null}
    </div>
  )
}
