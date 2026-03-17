import clsx from 'clsx'
import { Eyebrow } from '@/components/eyebrow'
import { Subheading } from '@/components/heading'
import { Text } from '@/components/text'

export function CardHeader({
  eyebrow,
  title,
  description,
  actions,
  className,
}: {
  eyebrow?: React.ReactNode
  title: React.ReactNode
  description?: React.ReactNode
  actions?: React.ReactNode
  className?: string
}) {
  return (
    <div className={clsx('flex items-start justify-between gap-4', className)}>
      <div>
        {eyebrow ? <Eyebrow>{eyebrow}</Eyebrow> : null}
        <Subheading>{title}</Subheading>
        {description ? <Text className="mt-2">{description}</Text> : null}
      </div>
      {actions ? <div className="shrink-0">{actions}</div> : null}
    </div>
  )
}
