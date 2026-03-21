'use client'

import { FormEvent } from 'react'
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
import { formatLocalDateTime } from '@/lib/datetime'
import {
  CUSTOM_MODEL_OPTION,
  DEFAULT_MODEL_OPTION,
  providerOptions,
  useAIProviderSettings,
} from '@/lib/use-ai-provider-settings'
import type { AIProviderSettings } from '@/lib/types'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
import { useToast } from './toast-provider'
import { WorkspaceCard } from './workspace-card'

function providerLabel(value?: string) {
  return providerOptions.find((item) => item.value === value)?.label ?? value ?? 'System provider'
}

function providerStatus(settings: AIProviderSettings | null) {
  if (settings && !settings.personal_provider_storage_available) {
    return {
      label: settings.system_fallback ? 'System provider' : 'Storage unavailable',
      detail: settings.system_fallback
        ? 'You can keep using the shared provider, but personal provider keys are not available here right now.'
        : 'Personal provider keys are not available here right now, and no shared provider is available.',
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
      detail: 'Your personal provider is active.',
      badgeColor: 'emerald' as const,
    }
  }
  return {
    label: 'System provider',
    detail: settings?.system_fallback ? 'The shared provider is active for now.' : 'No shared provider is available.',
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
      body: 'That key was rejected. Check that you copied it fully and chose the right provider.',
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
      body: 'Check the base URL and try again in a moment.',
      tone: 'warning' as const,
    }
  }
  return {
    title: 'Provider issue',
    body: message ?? 'The provider could not be used right now.',
    tone: 'danger' as const,
  }
}

export function AIProviderSettingsView({ required = false, nextPath }: { required?: boolean; nextPath?: string }) {
  const toast = useToast()
  const {
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
  } = useAIProviderSettings({ required, nextPath })
  const status = providerStatus(settings)
  const validatedAt = formatLocalDateTime(settings?.validated_at)
  const personalStorageAvailable = settings?.personal_provider_storage_available ?? true
  const promptModelIsPreset = selectedProvider.promptModels.includes(promptModel as never)
  const reviewModelIsPreset = selectedProvider.reviewModels.includes(reviewModel as never)
  const promptModelSelectValue = customPromptModel || (promptModel.trim() !== '' && !promptModelIsPreset) ? CUSTOM_MODEL_OPTION : promptModel.trim() === '' ? DEFAULT_MODEL_OPTION : promptModel
  const reviewModelSelectValue = customReviewModel || (reviewModel.trim() !== '' && !reviewModelIsPreset) ? CUSTOM_MODEL_OPTION : reviewModel.trim() === '' ? DEFAULT_MODEL_OPTION : reviewModel

  async function handleValidate() {
    try {
      const result = await validateConnection()
      toast.success(`Connection looks good for ${providerLabel(result.settings.provider)}.`, 'Provider check')
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Could not validate AI settings'
      const details = classifyProviderIssue(message)
      toast.error(details?.body ?? message, details?.title ?? 'Provider issue')
    }
  }

  async function handleSave(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    try {
      await save()
      toast.success('AI provider settings saved.')
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Could not save AI settings'
      const details = classifyProviderIssue(message)
      toast.error(details?.body ?? message, details?.title ?? 'Provider issue')
    }
  }

  async function handleDelete() {
    try {
      await remove()
      toast.success('Personal provider removed. The shared provider will be used when available.')
    } catch (err) {
      const message = err instanceof Error ? err.message : 'Could not remove AI settings'
      toast.error(message, 'Provider issue')
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
        eyebrow={required ? 'Step 1 of 3 · AI setup' : 'Settings'}
        title="AI provider"
        intro={
          required
            ? 'Connect an AI provider before creating your first practice path. Assignment generation and review both depend on it.'
            : 'Connect your own provider for assignment and review generation. You can keep using the shared provider while it is available.'
        }
      />

      {required ? (
        <Callout
          tone="active"
          eyebrow="Onboarding"
          title="First, connect an AI provider"
          body="You only need one working provider to continue."
        >
          <ul className="space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            <li>Choose a provider and paste a valid API key.</li>
            <li>Keep the default base URL unless you are using a compatible proxy.</li>
            <li>Save here to continue to Step 2 of 3: create your first practice path.</li>
          </ul>
        </Callout>
      ) : null}

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
            ? 'You can keep using the shared provider, but personal provider keys are not available here right now.'
            : 'Personal provider keys are not available here right now, and no shared provider fallback is available.'}
          tone={settings?.system_fallback ? 'warning' : 'danger'}
        />
      ) : null}

      {required && !settings?.ready ? (
        <Callout title="Setup needed to continue">
          Add a working provider before you create assignments or get feedback.
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
              description="Pick a provider and add your key. Advanced fields are optional."
            />

            <Callout
              title={`${providerLabel(provider)} defaults`}
              body={`${selectedProvider.note} Suggested base URL: ${selectedProvider.defaultBaseURL}. Suggested prompt model: ${selectedProvider.promptModel}. Suggested review model: ${selectedProvider.reviewModel}.`}
              tone="info"
              actions={
                <Button type="button" outline onClick={applyDefaults}>
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
                <Select value={promptModelSelectValue} onChange={(event) => handlePromptModelChange(event.target.value)}>
                  <option value={DEFAULT_MODEL_OPTION}>Use suggested default ({selectedProvider.promptModel})</option>
                  {selectedProvider.promptModels.map((model) => (
                    <option key={model} value={model}>
                      {model}
                    </option>
                  ))}
                  <option value={CUSTOM_MODEL_OPTION}>Custom model…</option>
                </Select>
                {promptModelSelectValue === CUSTOM_MODEL_OPTION ? (
                  <div className="mt-2">
                    <Input value={promptModel} onChange={(event) => setPromptModel(event.target.value)} placeholder={selectedProvider.promptModel} />
                  </div>
                ) : null}
              </Field>
              <Field>
                <Label>Review model override</Label>
                <Select value={reviewModelSelectValue} onChange={(event) => handleReviewModelChange(event.target.value)}>
                  <option value={DEFAULT_MODEL_OPTION}>Use suggested default ({selectedProvider.reviewModel})</option>
                  {selectedProvider.reviewModels.map((model) => (
                    <option key={model} value={model}>
                      {model}
                    </option>
                  ))}
                  <option value={CUSTOM_MODEL_OPTION}>Custom model…</option>
                </Select>
                {reviewModelSelectValue === CUSTOM_MODEL_OPTION ? (
                  <div className="mt-2">
                    <Input value={reviewModel} onChange={(event) => setReviewModel(event.target.value)} placeholder={selectedProvider.reviewModel} />
                  </div>
                ) : null}
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
