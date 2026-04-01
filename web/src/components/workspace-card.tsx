import clsx from 'clsx'
import type { ComponentPropsWithoutRef, ReactNode } from 'react'

export function WorkspaceCard({
  className,
  children,
  ...props
}: {
  className?: string
  children: ReactNode
} & ComponentPropsWithoutRef<'section'>) {
  return (
    <section
      {...props}
      className={clsx(
        className,
        'rounded-2xl border border-stone-200 bg-white p-6 shadow-sm ring-1 ring-black/2 dark:border-white/10 dark:bg-zinc-900'
      )}
    >
      {children}
    </section>
  )
}
