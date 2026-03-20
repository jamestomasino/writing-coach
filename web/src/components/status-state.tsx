import { Button } from '@/components/button'
import { Eyebrow } from '@/components/eyebrow'
import { Heading } from '@/components/heading'
import { Text } from '@/components/text'
import { WorkspaceCard } from './workspace-card'

type EmptyStateAction = {
  href: string
  label: string
  outline?: boolean
  plain?: boolean
}

export function EmptyState({
  title,
  body,
  actionHref,
  actionLabel,
  actions,
}: {
  title: string
  body: string
  actionHref?: string
  actionLabel?: string
  actions?: EmptyStateAction[]
}) {
  const resolvedActions = actions ?? (actionHref && actionLabel ? [{ href: actionHref, label: actionLabel }] : [])

  return (
    <div className="rounded-2xl border border-dashed border-stone-300 bg-stone-50 p-10 text-center dark:border-white/10 dark:bg-white/5">
      <Heading level={2}>{title}</Heading>
      <Text className="mx-auto mt-3 max-w-2xl">{body}</Text>
      {resolvedActions.length > 0 ? (
        <div className="mt-6 flex flex-wrap justify-center gap-3">
          {resolvedActions.map((action, index) => {
            const key = `${action.href}-${action.label}`

            if (action.plain) {
              return (
                <Button key={key} href={action.href} plain>
                  {action.label}
                </Button>
              )
            }

            if (action.outline || index !== 0) {
              return (
                <Button key={key} href={action.href} outline>
                  {action.label}
                </Button>
              )
            }

            return (
              <Button key={key} href={action.href} color="dark/zinc">
                {action.label}
              </Button>
            )
          })}
        </div>
      ) : null}
    </div>
  )
}

function classifyErrorMessage(message: string) {
  const normalized = message.toLowerCase()
  if (
    normalized.includes('unauthorized') ||
    normalized.includes('sign in required') ||
    normalized.includes('forbidden') ||
    normalized.includes('401') ||
    normalized.includes('403')
  ) {
    return {
      title: 'Sign in required',
      body: 'This page is only available inside an authenticated coaching session.',
      actions: [
        { href: '/login', label: 'Sign in' },
        { href: '/about', label: 'Back to home', outline: true },
      ],
    }
  }
  if (normalized.includes('not found') || normalized.includes('unavailable') || normalized.includes('404')) {
    return {
      title: 'Page unavailable',
      body: 'The page or record you requested could not be found in this account context.',
      actions: [
        { href: '/', label: 'Active track' },
        { href: '/assignments', label: 'Past assignments', outline: true },
        { href: '/about', label: 'About', outline: true },
      ],
    }
  }
  if (
    normalized.includes('500') ||
    normalized.includes('502') ||
    normalized.includes('503') ||
    normalized.includes('504') ||
    normalized.includes('internal server error')
  ) {
    return {
      title: 'Something went wrong',
      body: 'The app hit a server-side problem while trying to load this page.',
      actions: [
        { href: '/', label: 'Active track' },
        { href: '/about', label: 'About', outline: true },
      ],
    }
  }
  return {
    title: 'Something went wrong',
    body: message,
    actions: [
      { href: '/', label: 'Active track' },
      { href: '/about', label: 'About', outline: true },
    ],
  }
}

export function AppErrorState({
  error,
  title,
  body,
}: {
  error: string | null | undefined
  title?: string
  body?: string
}) {
  const details = classifyErrorMessage(error ?? body ?? 'Something went wrong')
  return <EmptyState title={title ?? details.title} body={body ?? details.body} actions={details.actions} />
}

export function LoadingState({ label = 'Loading workspace…' }: { label?: string }) {
  return (
    <div className="rounded-2xl border border-stone-200 bg-white p-8 text-center text-sm text-zinc-500 shadow-sm dark:border-white/10 dark:bg-zinc-900 dark:text-zinc-400">
      {label}
    </div>
  )
}

export function TaskProgressState({ title, body, steps }: { title: string; body: string; steps: string[] }) {
  return (
    <WorkspaceCard className="border-amber-200 bg-amber-50 dark:border-amber-500/20 dark:bg-amber-500/10">
      <div className="flex flex-col gap-5 lg:flex-row lg:items-start lg:justify-between">
        <div className="max-w-2xl">
          <div className="flex items-center gap-3">
            <span
              className="inline-flex size-3 animate-pulse rounded-full bg-amber-500 dark:bg-amber-300"
              aria-hidden="true"
            />
            <Heading level={2} className="text-base/6">
              {title}
            </Heading>
          </div>
          <Text className="mt-2" aria-live="polite">
            {body}
          </Text>
        </div>
        <div
          className="min-w-0 rounded-2xl border border-amber-300/70 bg-white/70 px-4 py-4 lg:w-80 dark:border-amber-400/20 dark:bg-black/10"
          role="status"
          aria-live="polite"
        >
          <Eyebrow tone="amber">Progress</Eyebrow>
          <ol className="mt-3 space-y-3">
            {steps.map((step, index) => (
              <li key={step} className="flex items-start gap-3 text-sm text-amber-950 dark:text-amber-50">
                <span className="mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full border border-amber-400 bg-amber-100 text-xs font-semibold dark:border-amber-300/30 dark:bg-amber-300/10">
                  {index + 1}
                </span>
                <span>{step}</span>
              </li>
            ))}
          </ol>
        </div>
      </div>
    </WorkspaceCard>
  )
}
