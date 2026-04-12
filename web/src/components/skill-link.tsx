import clsx from 'clsx'
import { Link } from '@/components/link'
import { hasSkillDetail, skillHref } from '@/lib/skill-details'

type SkillLinkProps = {
  skill: string
  className?: string
  children?: React.ReactNode
}

export function SkillLink({ skill, className, children }: SkillLinkProps) {
  const value = skill.trim()
  const label = children ?? value

  if (!value || !hasSkillDetail(value)) {
    return <span className={className}>{label}</span>
  }

  return (
    <Link
      href={skillHref(value)}
      className={clsx(
        className,
        'underline decoration-zinc-300 decoration-2 underline-offset-2 data-hover:decoration-zinc-600 dark:decoration-zinc-600 dark:data-hover:decoration-zinc-200'
      )}
    >
      {label}
    </Link>
  )
}
