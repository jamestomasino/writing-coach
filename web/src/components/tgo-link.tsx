import clsx from 'clsx'
import { Link } from '@/components/link'

type TgoLinkProps = {
  code: string
  label?: string
  className?: string
}

export function TgoLink({ code, label, className }: TgoLinkProps) {
  const value = code.trim()
  if (!value) {
    return <span className={className}>{label ?? code}</span>
  }
  return (
    <Link
      href={`/tgos/${encodeURIComponent(value)}`}
      className={clsx(
        className,
        'underline decoration-zinc-300 decoration-2 underline-offset-2 data-hover:decoration-zinc-600 dark:decoration-zinc-600 dark:data-hover:decoration-zinc-200'
      )}
    >
      {label ?? code}
    </Link>
  )
}
