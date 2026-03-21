'use client'

import { FormEvent } from 'react'
import { useTranslations } from 'next-intl'
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

function providerLabel(t: ReturnType<typeof useTranslations<'aiProviderSettingsView'>>, value?: string) {
  return providerOptions.find((item) => item.value === value)?.label ?? value ?? t('systemProvider')
}

function providerStatus(t: ReturnType<typeof useTranslations<'aiProviderSettingsView'>>, settings: AIProviderSettings | null) {
  if (settings && !settings.personal_provider_storage_available) {
    return {
      label: settings.system_fallback ? t('systemProvider') : t('storageUnavailable'),
      detail: settings.system_fallback
        ? t('statusDetailSharedStorageUnavailable')
        : t('statusDetailStorageUnavailableNoShared'),
      badgeColor: settings.system_fallback ? ('amber' as const) : ('rose' as const),
    }
  }
  if (!settings?.ready) {
    return {
      label: t('setupRequired'),
      detail: t('statusDetailSetupRequired'),
      badgeColor: 'rose' as const,
    }
  }
  if (settings.enabled && settings.provider) {
    return {
      label: providerLabel(t, settings.provider),
      detail: t('statusDetailPersonalActive'),
      badgeColor: 'emerald' as const,
    }
  }
  return {
    label: t('systemProvider'),
    detail: settings?.system_fallback ? t('statusDetailSharedActive') : t('statusDetailNoShared'),
    badgeColor: settings?.system_fallback ? ('amber' as const) : ('rose' as const),
  }
}

function classifyProviderIssue(t: ReturnType<typeof useTranslations<'aiProviderSettingsView'>>, message: string | null) {
  const text = (message ?? '').toLowerCase()
  if (!text) {
    return null
  }
  if (text.includes('rejected this api key')) {
    return {
      title: t('issueApiKeyRejectedTitle'),
      body: t('issueApiKeyRejectedBody'),
      tone: 'danger' as const,
    }
  }
  if (text.includes('out of quota') || text.includes('billing')) {
    return {
      title: t('issueQuotaTitle'),
      body: t('issueQuotaBody'),
      tone: 'warning' as const,
    }
  }
  if (text.includes('rate-limiting')) {
    return {
      title: t('issueRateLimitingTitle'),
      body: t('issueRateLimitingBody'),
      tone: 'warning' as const,
    }
  }
  if (text.includes('temporarily unavailable') || text.includes('timed out')) {
    return {
      title: t('issueTemporarilyUnavailableTitle'),
      body: t('issueTemporarilyUnavailableBody'),
      tone: 'warning' as const,
    }
  }
  return {
    title: t('issueGenericTitle'),
    body: message ?? t('issueGenericBody'),
    tone: 'danger' as const,
  }
}

export function AIProviderSettingsView({ required = false, nextPath }: { required?: boolean; nextPath?: string }) {
  const t = useTranslations('aiProviderSettingsView')
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
  const status = providerStatus(t, settings)
  const validatedAt = formatLocalDateTime(settings?.validated_at)
  const personalStorageAvailable = settings?.personal_provider_storage_available ?? true
  const promptModelIsPreset = selectedProvider.promptModels.includes(promptModel as never)
  const reviewModelIsPreset = selectedProvider.reviewModels.includes(reviewModel as never)
  const promptModelSelectValue = customPromptModel || (promptModel.trim() !== '' && !promptModelIsPreset) ? CUSTOM_MODEL_OPTION : promptModel.trim() === '' ? DEFAULT_MODEL_OPTION : promptModel
  const reviewModelSelectValue = customReviewModel || (reviewModel.trim() !== '' && !reviewModelIsPreset) ? CUSTOM_MODEL_OPTION : reviewModel.trim() === '' ? DEFAULT_MODEL_OPTION : reviewModel

  async function handleValidate() {
    try {
      const result = await validateConnection()
      toast.success(t('connectionLooksGood', { provider: providerLabel(t, result.settings.provider) }), t('providerCheck'))
    } catch (err) {
      const message = err instanceof Error ? err.message : t('validateError')
      const details = classifyProviderIssue(t, message)
      toast.error(details?.body ?? message, details?.title ?? t('issueGenericTitle'))
    }
  }

  async function handleSave(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    try {
      await save()
      toast.success(t('settingsSaved'))
    } catch (err) {
      const message = err instanceof Error ? err.message : t('saveError')
      const details = classifyProviderIssue(t, message)
      toast.error(details?.body ?? message, details?.title ?? t('issueGenericTitle'))
    }
  }

  async function handleDelete() {
    try {
      await remove()
      toast.success(t('personalProviderRemoved'))
    } catch (err) {
      const message = err instanceof Error ? err.message : t('removeError')
      toast.error(message, t('issueGenericTitle'))
    }
  }

  if (loading) {
    return <LoadingState label={t('loading')} />
  }
  if (error && !settings) {
    return <AppErrorState title={t('unavailableTitle')} error={error} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow={required ? t('eyebrowSetup') : t('eyebrowSettings')}
        title={t('title')}
        intro={
          required
            ? t('introRequired')
            : t('introOptional')
        }
      />

      {required ? (
        <Callout
          tone="active"
          eyebrow={t('onboardingEyebrow')}
          title={t('firstConnectTitle')}
          body={t('firstConnectBody')}
        >
          <ul className="space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            <li>{t('firstConnectListOne')}</li>
            <li>{t('firstConnectListTwo')}</li>
            <li>{t('firstConnectListThree')}</li>
          </ul>
        </Callout>
      ) : null}

      <WorkspaceCard>
        <CardHeader eyebrow={t('statusEyebrow')} title={t('statusTitle')} />
        <div className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
          <div className="flex items-center justify-between gap-3">
            <p>
              {t('effectiveProvider')}{' '}
              <span className="font-semibold text-zinc-950 dark:text-white">{status.label}</span>
            </p>
            <Badge color={status.badgeColor}>
              {settings?.enabled && settings?.provider ? t('badgePersonal') : settings?.system_fallback ? t('badgeShared') : t('badgeNeeded')}
            </Badge>
          </div>
          <p>{status.detail}</p>
          <p>
            {t('systemFallback')}{' '}
            <span className="font-semibold text-zinc-950 dark:text-white">{settings?.system_fallback ? t('available') : t('notAvailable')}</span>
          </p>
          {settings?.has_key ? (
            <p>
              {t('savedKey')}{' '}
              <span className="font-semibold text-zinc-950 dark:text-white">••••{settings.key_last4}</span>
            </p>
          ) : null}
          {validatedAt ? (
            <p>
              {t('lastChecked')}{' '}
              <span className="font-semibold text-zinc-950 dark:text-white">{validatedAt}</span>
            </p>
          ) : null}
        </div>
      </WorkspaceCard>

      {!settings?.ready ? (
        <Callout title={t('setupRequiredCalloutTitle')}>
          {t('setupRequiredCalloutBody')}
        </Callout>
      ) : null}

      {!personalStorageAvailable ? (
        <Callout
          title={t('personalStorageUnavailableTitle')}
          body={settings?.system_fallback
            ? t('statusDetailSharedStorageUnavailable')
            : t('personalStorageUnavailableNoFallback')}
          tone={settings?.system_fallback ? 'warning' : 'danger'}
        />
      ) : null}

      {required && !settings?.ready ? (
        <Callout title={t('continueSetupTitle')}>
          {t('continueSetupBody')}
        </Callout>
      ) : null}

      {settings?.last_validation_error ? (
        <Callout
          title={t('lastValidationIssueTitle')}
          body={settings.last_validation_error}
          tone="warning"
        />
      ) : null}

      {personalStorageAvailable ? (
        <WorkspaceCard>
          <form className="space-y-8" onSubmit={handleSave}>
            <CardHeader
              eyebrow={t('connectionEyebrow')}
              title={t('personalProviderTitle')}
              description={t('personalProviderDescription')}
            />

            <Callout
              title={t('providerDefaultsTitle', { provider: providerLabel(t, provider) })}
              body={t('providerDefaultsBody', {
                note: selectedProvider.note,
                baseURL: selectedProvider.defaultBaseURL,
                promptModel: selectedProvider.promptModel,
                reviewModel: selectedProvider.reviewModel,
              })}
              tone="info"
              actions={
                <Button type="button" outline onClick={applyDefaults}>
                  {t('useSuggestedDefaults')}
                </Button>
              }
            />

            <FieldGroup>
              <Field>
                <Label>{t('providerField')}</Label>
                <Select value={provider} onChange={(event) => setProvider(event.target.value)}>
                  {providerOptions.map((item) => (
                    <option key={item.value} value={item.value}>
                      {item.label}
                    </option>
                  ))}
                </Select>
              </Field>
              <Field>
                <Label>{t('apiKeyField')}</Label>
                <Input
                  type="password"
                  value={apiKey}
                  onChange={(event) => setAPIKey(event.target.value)}
                  placeholder={settings?.has_key ? t('apiKeyPlaceholderReplace', { hint: selectedProvider.apiKeyHint }) : t('apiKeyPlaceholderPaste', { hint: selectedProvider.apiKeyHint })}
                />
                <Text className="mt-2 text-sm">
                  {settings?.has_key ? t('apiKeyHelpKeepSaved') : t('apiKeyHelpEncrypted')}
                </Text>
                <Text className="mt-2 text-sm">{selectedProvider.validationHelp}</Text>
              </Field>
            </FieldGroup>

            <FieldGroup>
              <Field>
                <Label>{t('baseUrlOverrideField')}</Label>
                <Input value={baseURL} onChange={(event) => setBaseURL(event.target.value)} placeholder={selectedProvider.defaultBaseURL} />
                <Text className="mt-2 text-sm">{t('baseUrlOverrideHelp')}</Text>
              </Field>
              <Field>
                <Label>{t('promptModelOverrideField')}</Label>
                <Select value={promptModelSelectValue} onChange={(event) => handlePromptModelChange(event.target.value)}>
                  <option value={DEFAULT_MODEL_OPTION}>{t('useSuggestedPromptDefault', { model: selectedProvider.promptModel })}</option>
                  {selectedProvider.promptModels.map((model) => (
                    <option key={model} value={model}>
                      {model}
                    </option>
                  ))}
                  <option value={CUSTOM_MODEL_OPTION}>{t('customModel')}</option>
                </Select>
                {promptModelSelectValue === CUSTOM_MODEL_OPTION ? (
                  <div className="mt-2">
                    <Input value={promptModel} onChange={(event) => setPromptModel(event.target.value)} placeholder={selectedProvider.promptModel} />
                  </div>
                ) : null}
              </Field>
              <Field>
                <Label>{t('reviewModelOverrideField')}</Label>
                <Select value={reviewModelSelectValue} onChange={(event) => handleReviewModelChange(event.target.value)}>
                  <option value={DEFAULT_MODEL_OPTION}>{t('useSuggestedReviewDefault', { model: selectedProvider.reviewModel })}</option>
                  {selectedProvider.reviewModels.map((model) => (
                    <option key={model} value={model}>
                      {model}
                    </option>
                  ))}
                  <option value={CUSTOM_MODEL_OPTION}>{t('customModel')}</option>
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
              <Label>{t('useProviderForFuture')}</Label>
            </CheckboxField>

            <div className="flex flex-wrap gap-3">
              <Button type="button" outline onClick={handleValidate} disabled={validating || (apiKey.trim() === '' && !settings?.has_key)}>
                {validating ? t('checkingProvider') : t('validateConnection')}
              </Button>
              <Button type="submit" color="dark/zinc" disabled={saving || (apiKey.trim() === '' && !settings?.has_key)}>
                {saving ? t('saving') : t('saveProvider')}
              </Button>
              {settings?.has_key ? (
                <Button type="button" plain onClick={handleDelete} disabled={saving}>
                  {t('removePersonalProvider')}
                </Button>
              ) : null}
            </div>
          </form>
        </WorkspaceCard>
      ) : null}
    </div>
  )
}
