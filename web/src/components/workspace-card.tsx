import clsx from 'clsx'

export function WorkspaceCard({
  className,
  children,
}: {
  className?: string
  children: React.ReactNode
}) {
  return (
    <section
      className={clsx(
        className,
        'rounded-2xl border border-stone-200 bg-white p-6 shadow-sm ring-1 ring-black/2 dark:border-white/10 dark:bg-zinc-900'
      )}
    >
      {children}
    </section>
  )
}
