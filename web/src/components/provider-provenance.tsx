'use client'

import { Badge } from '@/components/badge'
import { Text } from '@/components/text'
import { useTranslations } from 'next-intl'

const deterministicTooling = ['Rubric Engine', 'Heuristic', 'Vale', 'LanguageTool', 'NLP']

function providerLabel(value: string) {
  switch (value) {
    case 'anthropic':
      return 'Anthropic'
    case 'gemini':
      return 'Gemini'
    case 'openai':
      return 'OpenAI'
    case 'groq':
      return 'Groq'
    case 'xai':
      return 'xAI'
    default:
      return value
  }
}

function isKnownProvider(value: string) {
  return ['anthropic', 'gemini', 'openai', 'groq', 'xai'].includes(value)
}

function titleCase(value: string) {
  return value
    .split(/[-_/]/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1))
    .join(' ')
}

function parseProviderNote(providerNote?: string) {
  if (!providerNote) {
    return null
  }
  const [head, detail = ''] = providerNote.split(' • ', 2)
  const trimmedHead = head.trim()
  if (trimmedHead === '') {
    return null
  }

  if (trimmedHead === 'deterministic' || trimmedHead === 'deterministic-fallback') {
    return {
      sourceLabel: trimmedHead === 'deterministic' ? 'Deterministic' : 'Fallback',
      providerLabel: '',
      modelLabel: detail.trim(),
    }
  }

  const [mode, provider] = trimmedHead.split('/', 2)
  if (!provider) {
    if (isKnownProvider(trimmedHead)) {
      return {
        sourceLabel: 'LLM',
        providerLabel: providerLabel(trimmedHead),
        modelLabel: detail.trim(),
      }
    }
    return {
      sourceLabel: titleCase(trimmedHead),
      providerLabel: '',
      modelLabel: detail.trim(),
    }
  }

  return {
    sourceLabel: mode === 'user' ? 'Personal' : mode === 'system' ? 'Shared' : titleCase(mode),
    providerLabel: providerLabel(provider),
    modelLabel: detail.trim(),
  }
}

export function ProviderProvenance({
  providerNote,
  compact = false,
  kind = 'default',
}: {
  providerNote?: string
  compact?: boolean
  kind?: 'default' | 'feedback'
}) {
  const t = useTranslations('providerProvenance')
  const parsed = parseProviderNote(providerNote)
  if (!parsed) {
    return null
  }

  const feedbackMode = kind === 'feedback'
  const showLLM = parsed.providerLabel || parsed.sourceLabel.toLowerCase() === 'llm' || parsed.modelLabel

  return (
    <div className={compact ? 'flex flex-wrap items-center gap-2' : 'space-y-2'}>
      <div className="flex flex-wrap items-center gap-2">
        {feedbackMode ? (
          <>
            <Badge color="green">{t('deterministic')}</Badge>
            {deterministicTooling.map((tool) => (
              <Badge key={tool} color="emerald">
                {tool}
              </Badge>
            ))}
            {showLLM ? (
              <>
                <Badge color="zinc">LLM (secondary)</Badge>
                {parsed.providerLabel ? <Badge color="zinc">{parsed.providerLabel}</Badge> : null}
                {parsed.modelLabel ? <Badge color="zinc">{parsed.modelLabel}</Badge> : null}
              </>
            ) : null}
          </>
        ) : (
          <>
            <Badge color="zinc">{parsed.sourceLabel}</Badge>
            {parsed.providerLabel ? <Badge color="cyan">{parsed.providerLabel}</Badge> : null}
            {parsed.modelLabel ? <Badge color="amber">{parsed.modelLabel}</Badge> : null}
          </>
        )}
      </div>
      {!compact && feedbackMode ? (
        <Text className="text-sm">{t('feedbackPipeline')}</Text>
      ) : null}
      {!compact && !feedbackMode && parsed.providerLabel ? (
        <Text className="text-sm">
          {parsed.modelLabel
            ? t('createdWithModel', {
                source: parsed.sourceLabel.toLowerCase(),
                provider: parsed.providerLabel,
                model: parsed.modelLabel,
              })
            : t('createdWith', {
                source: parsed.sourceLabel.toLowerCase(),
                provider: parsed.providerLabel,
              })}
        </Text>
      ) : null}
    </div>
  )
}
