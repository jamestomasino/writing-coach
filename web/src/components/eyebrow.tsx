import clsx from 'clsx'

const toneClasses = {
  amber: 'text-amber-700 dark:text-amber-300',
  zinc: 'text-zinc-500 dark:text-zinc-400',
  white: 'text-zinc-400 dark:text-zinc-400',
  cyan: 'text-cyan-900 dark:text-cyan-100',
  emerald: 'text-emerald-700 dark:text-emerald-200',
} as const

export function Eyebrow({
  className,
  tone = 'amber',
  ...props
}: React.ComponentPropsWithoutRef<'div'> & {
  tone?: keyof typeof toneClasses
}) {
  return (
    <div
      {...props}
      className={clsx(
        'text-xs font-semibold uppercase tracking-[0.18em]',
        toneClasses[tone],
        className
      )}
    />
  )
}
