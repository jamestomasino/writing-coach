'use client'

import { Badge } from '@/components/badge'
import { Text } from '@/components/text'

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
}: {
  providerNote?: string
  compact?: boolean
}) {
  const parsed = parseProviderNote(providerNote)
  if (!parsed) {
    return null
  }

  return (
    <div className={compact ? 'flex flex-wrap items-center gap-2' : 'space-y-2'}>
      <div className="flex flex-wrap items-center gap-2">
        <Badge color="zinc">{parsed.sourceLabel}</Badge>
        {parsed.providerLabel ? <Badge color="cyan">{parsed.providerLabel}</Badge> : null}
        {parsed.modelLabel ? <Badge color="amber">{parsed.modelLabel}</Badge> : null}
      </div>
      {!compact && parsed.providerLabel ? (
        <Text className="text-sm">
          Created with {parsed.sourceLabel.toLowerCase()} {parsed.providerLabel}
          {parsed.modelLabel ? ` using ${parsed.modelLabel}` : ''}.
        </Text>
      ) : null}
    </div>
  )
}
