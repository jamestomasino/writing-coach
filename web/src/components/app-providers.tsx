'use client'

import type React from 'react'
import { ToastProvider } from '@/components/toast-provider'

export function AppProviders({ children }: { children: React.ReactNode }) {
  return <ToastProvider>{children}</ToastProvider>
}
