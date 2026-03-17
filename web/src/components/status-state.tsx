import { Button } from '@/components/button'
import { Heading } from '@/components/heading'
import { Text } from '@/components/text'
import { WorkspaceCard } from './workspace-card'

export function EmptyState({
  title,
  body,
  actionHref,
  actionLabel,
}: {
  title: string
  body: string
  actionHref?: string
  actionLabel?: string
}) {
  return (
    <div className="rounded-2xl border border-dashed border-stone-300 bg-stone-50 p-10 text-center dark:border-white/10 dark:bg-white/5">
      <Heading level={2}>{title}</Heading>
      <Text className="mx-auto mt-3 max-w-2xl">{body}</Text>
      {actionHref && actionLabel ? (
        <div className="mt-6">
          <Button href={actionHref} color="dark/zinc">
            {actionLabel}
          </Button>
        </div>
      ) : null}
    </div>
  )
}

export function LoadingState({ label = 'Loading workspace…' }: { label?: string }) {
  return (
    <div className="rounded-2xl border border-stone-200 bg-white p-8 text-center text-sm text-zinc-500 shadow-sm dark:border-white/10 dark:bg-zinc-900 dark:text-zinc-400">
      {label}
    </div>
  )
}

export function TaskProgressState({
  title,
  body,
  steps,
}: {
  title: string
  body: string
  steps: string[]
}) {
  return (
    <WorkspaceCard
      className="border-amber-200 bg-amber-50 dark:border-amber-500/20 dark:bg-amber-500/10"
    >
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
          className="min-w-0 rounded-2xl border border-amber-300/70 bg-white/70 px-4 py-4 dark:border-amber-400/20 dark:bg-black/10 lg:w-80"
          role="status"
          aria-live="polite"
        >
          <div className="text-xs font-semibold uppercase tracking-[0.18em] text-amber-900 dark:text-amber-100">
            Progress
          </div>
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
