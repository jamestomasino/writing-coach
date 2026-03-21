import type enMessages from '@/messages/en.json'

export const defaultLocale = 'en'
export const localeCookieName = 'writing_coach_locale'

export const localeLabels = {
  en: 'English',
} as const

export type AppLocale = keyof typeof localeLabels

type MessagesShape = typeof enMessages

export const availableLocales = Object.keys(localeLabels) as AppLocale[]

export async function getLocaleMessages(locale: AppLocale): Promise<MessagesShape> {
  switch (locale) {
    case 'en':
    default:
      return (await import('@/messages/en.json')).default
  }
}

export function isSupportedLocale(value: string | null | undefined): value is AppLocale {
  if (!value) {
    return false
  }
  return availableLocales.includes(value as AppLocale)
}

export function resolveLocale(preferred: string | null | undefined, acceptLanguageHeader?: string | null): AppLocale {
  if (isSupportedLocale(preferred)) {
    return preferred
  }

  const header = (acceptLanguageHeader ?? '').trim().toLowerCase()
  if (header) {
    const tags = header
      .split(',')
      .map((part) => part.split(';')[0]?.trim())
      .filter(Boolean)
    for (const tag of tags) {
      if (isSupportedLocale(tag)) {
        return tag
      }
      const base = tag.split('-')[0]
      if (isSupportedLocale(base)) {
        return base
      }
    }
  }

  return defaultLocale
}
