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
      body: 'Sign in to open this page.',
      actions: [
        { href: '/login', label: 'Sign in' },
        { href: '/about', label: 'Back to home', outline: true },
      ],
    }
  }
  if (normalized.includes('not found') || normalized.includes('unavailable') || normalized.includes('404')) {
    return {
      title: 'Page unavailable',
      body: 'The page or record you requested could not be found in this practice path.',
      actions: [
        { href: '/', label: 'Current assignment' },
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
      body: 'A server error came back while loading this page.',
      actions: [
        { href: '/', label: 'Current assignment' },
        { href: '/about', label: 'About', outline: true },
      ],
    }
  }
  return {
    title: 'Something went wrong',
    body: message,
    actions: [
      { href: '/', label: 'Current assignment' },
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
    <WorkspaceCard className="border-cyan-200 bg-cyan-50 dark:border-cyan-500/20 dark:bg-cyan-500/10">
      <div className="flex flex-col gap-5 lg:flex-row lg:items-center lg:justify-between">
        <div className="flex items-start gap-4">
          <div
            className="mt-0.5 inline-flex size-10 shrink-0 items-center justify-center rounded-full border border-cyan-300 bg-white/80 dark:border-cyan-400/20 dark:bg-black/10"
            aria-hidden="true"
          >
            <span className="size-5 animate-spin rounded-full border-2 border-cyan-700/25 border-t-cyan-700 dark:border-cyan-200/25 dark:border-t-cyan-200" />
          </div>
          <div className="max-w-2xl">
            <Eyebrow tone="cyan">Working</Eyebrow>
            <Heading level={2} className="mt-1 text-base/6">
              {title}
            </Heading>
            <Text className="mt-2" aria-live="polite">
              {body}
            </Text>
          </div>
        </div>
        <div
          className="min-w-0 rounded-2xl border border-cyan-300/70 bg-white/70 px-4 py-4 lg:w-80 dark:border-cyan-400/20 dark:bg-black/10"
          role="status"
          aria-live="polite"
        >
          <Eyebrow tone="cyan">Progress</Eyebrow>
          <div className="mt-3 space-y-2" aria-hidden="true">
            <div className="h-2 w-full animate-pulse rounded-full bg-cyan-200/80 dark:bg-cyan-200/15" />
            <div className="h-2 w-5/6 animate-pulse rounded-full bg-cyan-200/70 [animation-delay:120ms] dark:bg-cyan-200/12" />
            <div className="h-2 w-2/3 animate-pulse rounded-full bg-cyan-200/60 [animation-delay:240ms] dark:bg-cyan-200/10" />
          </div>
          <ol className="mt-4 space-y-3">
            {steps.map((step, index) => (
              <li key={step} className="flex items-start gap-3 text-sm text-cyan-950 dark:text-cyan-50">
                <span className="mt-0.5 inline-flex size-6 shrink-0 items-center justify-center rounded-full border border-cyan-400 bg-cyan-100 text-xs font-semibold dark:border-cyan-300/30 dark:bg-cyan-300/10">
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
