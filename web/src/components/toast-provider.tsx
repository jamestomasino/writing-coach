'use client'

import type React from 'react'
import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'
import clsx from 'clsx'
import { Button } from '@/components/button'
import { Text } from '@/components/text'

type ToastTone = 'info' | 'success' | 'warning' | 'danger'

type ToastInput = {
  title?: string
  body: string
  tone?: ToastTone
  duration?: number
  actionLabel?: string
  onAction?: () => void
}

type ToastRecord = ToastInput & {
  id: number
}

type ToastContextValue = {
  showToast: (input: ToastInput) => void
  success: (body: string, title?: string) => void
  error: (body: string, title?: string) => void
  info: (body: string, title?: string) => void
  warning: (body: string, title?: string) => void
}

const toneClasses: Record<ToastTone, string> = {
  info: 'border-cyan-200 bg-cyan-50 dark:border-cyan-500/20 dark:bg-cyan-500/10',
  success: 'border-green-200 bg-green-50 dark:border-green-500/20 dark:bg-green-500/10',
  warning: 'border-amber-200 bg-amber-50 dark:border-amber-500/20 dark:bg-amber-500/10',
  danger: 'border-rose-200 bg-rose-50 dark:border-rose-500/20 dark:bg-rose-500/10',
}

const ToastContext = createContext<ToastContextValue | null>(null)

export function ToastProvider({ children }: { children: React.ReactNode }) {
  const [toasts, setToasts] = useState<ToastRecord[]>([])

  const dismissToast = useCallback((id: number) => {
    setToasts((current) => current.filter((toast) => toast.id !== id))
  }, [])

  const showToast = useCallback((input: ToastInput) => {
    const id = Date.now() + Math.floor(Math.random() * 1000)
    const toast: ToastRecord = {
      id,
      tone: input.tone ?? 'info',
      duration: input.duration ?? (input.tone === 'danger' ? 0 : 4500),
      ...input,
    }
    setToasts((current) => [...current.slice(-2), toast])
  }, [])

  useEffect(() => {
    const timers = toasts
      .filter((toast) => (toast.duration ?? 0) > 0)
      .map((toast) =>
        window.setTimeout(() => {
          dismissToast(toast.id)
        }, toast.duration)
      )
    return () => {
      for (const timer of timers) {
        window.clearTimeout(timer)
      }
    }
  }, [dismissToast, toasts])

  const value = useMemo<ToastContextValue>(
    () => ({
      showToast,
      success: (body, title = 'Done') => showToast({ body, title, tone: 'success' }),
      error: (body, title = 'Something went wrong') => showToast({ body, title, tone: 'danger' }),
      info: (body, title = 'Note') => showToast({ body, title, tone: 'info' }),
      warning: (body, title = 'Attention') => showToast({ body, title, tone: 'warning' }),
    }),
    [showToast]
  )

  return (
    <ToastContext.Provider value={value}>
      {children}
      <div className="pointer-events-none fixed inset-x-0 top-4 z-50 flex justify-center px-4 sm:justify-end sm:px-6">
        <div className="flex w-full max-w-sm flex-col gap-3">
          {toasts.map((toast) => (
            <div
              key={toast.id}
              className={clsx(
                'pointer-events-auto rounded-2xl border p-4 shadow-lg ring-1 ring-zinc-950/5 backdrop-blur-sm dark:ring-white/10',
                toneClasses[toast.tone ?? 'info']
              )}
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0">
                  {toast.title ? <div className="text-sm font-semibold text-zinc-950 dark:text-white">{toast.title}</div> : null}
                  <Text className="mt-1 text-sm">{toast.body}</Text>
                </div>
                <Button type="button" plain onClick={() => dismissToast(toast.id)}>
                  Dismiss
                </Button>
              </div>
              {toast.actionLabel && toast.onAction ? (
                <div className="mt-3">
                  <Button
                    type="button"
                    outline
                    onClick={() => {
                      toast.onAction?.()
                      dismissToast(toast.id)
                    }}
                  >
                    {toast.actionLabel}
                  </Button>
                </div>
              ) : null}
            </div>
          ))}
        </div>
      </div>
    </ToastContext.Provider>
  )
}

export function useToast() {
  const context = useContext(ToastContext)
  if (!context) {
    throw new Error('useToast must be used within a ToastProvider')
  }
  return context
}
