import messages from '@/messages/en.json'

export const defaultLocale = 'en'

export const localeMessages = {
  en: messages,
} as const

export type AppLocale = keyof typeof localeMessages
