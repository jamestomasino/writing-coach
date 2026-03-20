'use client'

import { FormEvent, useEffect, useState } from 'react'
import { useRouter } from 'next/navigation'
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
import { WorkspaceCard } from './workspace-card'

const providerOptions = [
  { value: 'openai', label: 'OpenAI' },
  { value: 'groq', label: 'Groq' },
  { value: 'xai', label: 'xAI' },
]

function providerLabel(value?: string) {
  return providerOptions.find((item) => item.value === value)?.label ?? value ?? 'System provider'
}

export function AIProviderSettingsView({ required = false, nextPath }: { required?: boolean; nextPath?: string }) {
  const router = useRouter()
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
  const [validationMessage, setValidationMessage] = useState<string | null>(null)

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
      setValidationMessage(null)
      const result = await validateAISettings({
        provider,
        api_key: apiKey,
        base_url_override: baseURL,
        prompt_model_override: promptModel,
        review_model_override: reviewModel,
        enabled,
      })
      setValidationMessage(`Connection looks good for ${providerLabel(result.settings.provider)}.`)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not validate AI settings')
    } finally {
      setValidating(false)
    }
  }

  async function handleSave(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    try {
      setSaving(true)
      setError(null)
      setValidationMessage(null)
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
      setValidationMessage('AI provider settings saved.')
      if (required && current.ready) {
        router.push(nextPath && nextPath.startsWith('/') ? nextPath : '/')
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not save AI settings')
    } finally {
      setSaving(false)
    }
  }

  async function handleDelete() {
    try {
      setSaving(true)
      setError(null)
      setValidationMessage(null)
      const result = await deleteAISettings()
      setSettings(result.settings)
      setProvider('openai')
      setAPIKey('')
      setBaseURL('')
      setPromptModel('')
      setReviewModel('')
      setEnabled(true)
      setValidationMessage('Personal provider removed. The app will use the system provider when available.')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not remove AI settings')
    } finally {
      setSaving(false)
    }
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
          <p>
            Effective provider: <span className="font-semibold text-zinc-950 dark:text-white">{settings?.enabled && settings?.provider ? providerLabel(settings.provider) : 'System provider'}</span>
          </p>
          <p>
            System fallback: <span className="font-semibold text-zinc-950 dark:text-white">{settings?.system_fallback ? 'available' : 'not available'}</span>
          </p>
          {settings?.has_key ? (
            <p>
              Saved key: <span className="font-semibold text-zinc-950 dark:text-white">••••{settings.key_last4}</span>
            </p>
          ) : null}
        </div>
      </WorkspaceCard>

      {!settings?.ready ? (
        <Callout title="AI setup required">
          Add a provider key here before trying to generate assignments or reviews.
        </Callout>
      ) : null}

      {required && !settings?.ready ? (
        <Callout title="Setup needed to continue">
          This workspace needs an AI provider before assignment and feedback generation can run.
        </Callout>
      ) : null}

      {error ? (
        <EmptyState title="AI settings issue" body={error} />
      ) : null}

      {validationMessage ? (
        <Callout title="Provider check">{validationMessage}</Callout>
      ) : null}

      <WorkspaceCard>
        <form className="space-y-8" onSubmit={handleSave}>
          <CardHeader
            eyebrow="Connection"
            title="Personal provider"
            description="Phase 1 supports OpenAI, Groq, and xAI. Advanced fields are optional."
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
                placeholder={settings?.has_key ? 'Enter a new key to replace the saved one' : 'Paste your provider API key'}
              />
              <Text className="mt-2 text-sm">
                {settings?.has_key ? 'Leave this blank to keep using the saved key. Keys are encrypted before they are stored.' : 'Keys are encrypted before they are stored.'}
              </Text>
            </Field>
          </FieldGroup>

          <FieldGroup>
            <Field>
              <Label>Base URL override</Label>
              <Input value={baseURL} onChange={(event) => setBaseURL(event.target.value)} placeholder="Optional custom endpoint" />
            </Field>
            <Field>
              <Label>Prompt model override</Label>
              <Input value={promptModel} onChange={(event) => setPromptModel(event.target.value)} placeholder="Optional prompt model" />
            </Field>
            <Field>
              <Label>Review model override</Label>
              <Input value={reviewModel} onChange={(event) => setReviewModel(event.target.value)} placeholder="Optional review model" />
            </Field>
          </FieldGroup>

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
    </div>
  )
}
