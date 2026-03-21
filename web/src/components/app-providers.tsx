'use client'

import type React from 'react'
import { ToastProvider } from '@/components/toast-provider'
import { NextIntlClientProvider } from 'next-intl'

export function AppProviders({
  children,
  locale,
  messages,
}: {
  children: React.ReactNode
  locale: string
  messages: Record<string, unknown>
}) {
  return (
    <NextIntlClientProvider locale={locale} messages={messages}>
      <ToastProvider>{children}</ToastProvider>
    </NextIntlClientProvider>
  )
}
