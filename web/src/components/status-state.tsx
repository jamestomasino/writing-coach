import { Button } from '@/components/button'
import { Heading } from '@/components/heading'
import { Text } from '@/components/text'

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
