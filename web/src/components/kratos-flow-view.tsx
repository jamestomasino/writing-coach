'use client'

import { useEffect, useMemo, useState } from 'react'
import { useRouter, useSearchParams } from 'next/navigation'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { Checkbox, CheckboxField } from '@/components/checkbox'
import { Eyebrow } from '@/components/eyebrow'
import { AppErrorState, EmptyState, LoadingState } from '@/components/status-state'
import { Field, FieldGroup, Label } from '@/components/fieldset'
import { Heading, Subheading } from '@/components/heading'
import { Input } from '@/components/input'
import { PageHeader } from '@/components/page-header'
import { Text, TextLink } from '@/components/text'

type FlowKind = 'login' | 'registration' | 'verification' | 'recovery' | 'settings'

type KratosMessage = {
  id: number
  text: string
  type: 'error' | 'info' | 'success'
}

type KratosNode = {
  type: string
  group?: string
  attributes: {
    name?: string
    type?: string
    value?: string
    required?: boolean
    disabled?: boolean
    checked?: boolean
  }
  meta?: {
    label?: {
      text?: string
    }
  }
  messages?: KratosMessage[]
}

type KratosFlow = {
  id: string
  ui: {
    action: string
    method: string
    messages?: KratosMessage[]
    nodes: KratosNode[]
  }
}

const kindMeta: Record<
  FlowKind,
  {
    title: string
    intro: string
    initPath: string
    getPath: string
    submitLabel: string
    alternateHref?: string
    alternateLabel?: string
    alternateLead?: string
  }
> = {
  login: {
    title: 'Sign in to your account',
    intro: 'Use your account email and password to continue into the workshop.',
    initPath: 'login',
    getPath: 'login',
    submitLabel: 'Sign in',
    alternateHref: '/register',
    alternateLabel: 'Create an account',
    alternateLead: 'Need an account?',
  },
  registration: {
    title: 'Create your account',
    intro: 'Create a writing coach account. Email verification and password security are handled by Kratos behind the scenes.',
    initPath: 'registration',
    getPath: 'registration',
    submitLabel: 'Create account',
    alternateHref: '/login',
    alternateLabel: 'Sign in',
    alternateLead: 'Already have an account?',
  },
  verification: {
    title: 'Verify your email',
    intro: 'Enter the verification code from your email or request a fresh verification message.',
    initPath: 'verification',
    getPath: 'verification',
    submitLabel: 'Verify email',
  },
  recovery: {
    title: 'Recover your account',
    intro: 'Request a recovery code by email, then set a new password through the recovery flow.',
    initPath: 'recovery',
    getPath: 'recovery',
    submitLabel: 'Send recovery code',
  },
  settings: {
    title: 'Account settings',
    intro: 'Manage your account credentials and verified traits here.',
    initPath: 'settings',
    getPath: 'settings',
    submitLabel: 'Save settings',
  },
}

async function fetchFlow(path: string) {
  const response = await fetch(path, {
    credentials: 'include',
    headers: { Accept: 'application/json' },
  })
  if (!response.ok) {
    let message = `${response.status} ${response.statusText}`
    try {
      const payload = (await response.json()) as {
        error_text?: string
        error_detail?: { message?: string }
        error?: string | { message?: string }
      }
      if (typeof payload.error_text === 'string') {
        message = payload.error_text
      } else if (typeof payload.error === 'string') {
        message = payload.error
      } else if (payload.error_detail?.message) {
        message = payload.error_detail.message
      } else if (typeof payload.error === 'object' && payload.error?.message) {
        message = payload.error.message
      }
    } catch {}
    throw new Error(message)
  }
  return (await response.json()) as KratosFlow
}

function fieldLabel(node: KratosNode) {
  return node.meta?.label?.text ?? node.attributes.name ?? ''
}

function groupNodes(nodes: KratosNode[]) {
  const hidden: KratosNode[] = []
  const visible: KratosNode[] = []
  const submit: KratosNode[] = []

  for (const node of nodes) {
    const type = node.attributes.type ?? node.type
    if (type === 'hidden') {
      hidden.push(node)
      continue
    }
    if (type === 'submit') {
      submit.push(node)
      continue
    }
    visible.push(node)
  }

  return { hidden, visible, submit }
}

function FlowMessages({ messages }: { messages?: KratosMessage[] }) {
  if (!messages || messages.length === 0) {
    return null
  }
  return (
    <div className="space-y-3">
      {messages.map((message) => (
        <div
          key={`${message.id}-${message.text}`}
          className={`rounded-xl border px-4 py-3 text-sm ${
            message.type === 'error'
              ? 'border-rose-200 bg-rose-50 text-rose-900 dark:border-rose-500/20 dark:bg-rose-500/10 dark:text-rose-100'
              : message.type === 'success'
                ? 'border-emerald-200 bg-emerald-50 text-emerald-900 dark:border-emerald-500/20 dark:bg-emerald-500/10 dark:text-emerald-100'
                : 'border-stone-200 bg-stone-50 text-zinc-800 dark:border-white/10 dark:bg-white/5 dark:text-zinc-200'
          }`}
        >
          {message.text}
        </div>
      ))}
    </div>
  )
}

export function KratosFlowView({ kind }: { kind: FlowKind }) {
  const router = useRouter()
  const searchParams = useSearchParams()
  const flowID = searchParams.get('flow')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [flow, setFlow] = useState<KratosFlow | null>(null)

  const meta = kindMeta[kind]

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        setLoading(true)
        setError(null)
        let nextFlow: KratosFlow
        if (flowID) {
          nextFlow = await fetchFlow(`/.ory/kratos/public/self-service/${meta.getPath}/flows?id=${encodeURIComponent(flowID)}`)
        } else {
          nextFlow = await fetchFlow(`/.ory/kratos/public/self-service/${meta.initPath}/browser`)
          if (!cancelled) {
            const params = new URLSearchParams(searchParams.toString())
            params.set('flow', nextFlow.id)
            router.replace(`?${params.toString()}`)
          }
        }
        if (!cancelled) {
          setFlow(nextFlow)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load authentication flow')
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
  }, [flowID, meta.getPath, meta.initPath, router, searchParams])

  const grouped = useMemo(() => groupNodes(flow?.ui.nodes ?? []), [flow])

  if (loading) {
    return <LoadingState label="Loading account flow…" />
  }
  if (error || !flow) {
    return <AppErrorState title="Authentication unavailable" error={error ?? 'Could not load the account flow.'} />
  }

  const submitNode = grouped.submit[0]
  const submitLabel = submitNode ? fieldLabel(submitNode) || meta.submitLabel : meta.submitLabel

  return (
    <div className="grid w-full max-w-lg grid-cols-1 gap-8">
      <PageHeader eyebrow={kind === 'settings' ? 'Settings' : 'Account'} title={meta.title} intro={meta.intro} />

      <FlowMessages messages={flow.ui.messages} />

      <form action={flow.ui.action} method={flow.ui.method} className="space-y-6">
        {grouped.hidden.map((node, index) => (
          <input
            key={`${node.attributes.name ?? 'hidden'}-${index}`}
            type="hidden"
            name={node.attributes.name}
            value={node.attributes.value ?? ''}
          />
        ))}

        <FieldGroup>
          {grouped.visible.map((node, index) => {
            const name = node.attributes.name ?? `${kind}-${index}`
            const type = node.attributes.type ?? node.type
            const label = fieldLabel(node)

            if (type === 'checkbox') {
              return (
                <CheckboxField key={name}>
                  <Checkbox
                    name={node.attributes.name}
                    defaultChecked={Boolean(node.attributes.checked)}
                    disabled={node.attributes.disabled}
                    value={node.attributes.value ?? 'true'}
                  />
                  <Label>{label}</Label>
                </CheckboxField>
              )
            }

            return (
              <Field key={name}>
                {label ? <Label>{label}</Label> : null}
                <Input
                  name={node.attributes.name}
                  type={type === 'submit' ? 'text' : (type as 'email' | 'number' | 'password' | 'search' | 'tel' | 'text' | 'url')}
                  defaultValue={type === 'password' ? '' : (node.attributes.value ?? '')}
                  required={node.attributes.required}
                  disabled={node.attributes.disabled}
                  autoComplete={name.includes('password') ? 'current-password' : name.includes('email') ? 'email' : undefined}
                />
                <FlowMessages messages={node.messages} />
              </Field>
            )
          })}
        </FieldGroup>

        <Button
          type="submit"
          color="dark/zinc"
          className="w-full"
          name={submitNode?.attributes.name}
          value={submitNode?.attributes.value ?? ''}
          disabled={submitNode?.attributes.disabled}
        >
          {submitLabel}
        </Button>
      </form>

      {kind === 'verification' || kind === 'recovery' ? (
        <div className="space-y-2 rounded-2xl border border-stone-200 bg-stone-50 px-4 py-4 dark:border-white/10 dark:bg-white/5">
          <CardHeader eyebrow="Account flow" title="Need a different account flow?" />
          <Text className="mt-2 text-sm">
            <TextLink href="/login">Return to sign in</TextLink>
            {' · '}
            <TextLink href="/register">Create a new account</TextLink>
          </Text>
        </div>
      ) : null}

      {meta.alternateHref && meta.alternateLabel && meta.alternateLead ? (
        <Text>
          {meta.alternateLead} <TextLink href={meta.alternateHref}>{meta.alternateLabel}</TextLink>
        </Text>
      ) : null}
    </div>
  )
}
