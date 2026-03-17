import clsx from 'clsx'
import { Eyebrow } from '@/components/eyebrow'
import { Heading } from '@/components/heading'
import { Text } from '@/components/text'

export function PageHeader({
  eyebrow,
  title,
  intro,
  actions,
  className,
}: {
  eyebrow?: React.ReactNode
  title: React.ReactNode
  intro?: React.ReactNode
  actions?: React.ReactNode
  className?: string
}) {
  return (
    <header className={clsx('flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between', className)}>
      <div>
        {eyebrow ? <Eyebrow>{eyebrow}</Eyebrow> : null}
        <Heading>{title}</Heading>
        {intro ? <Text className="mt-2 max-w-3xl">{intro}</Text> : null}
      </div>
      {actions ? <div className="flex flex-wrap gap-2">{actions}</div> : null}
    </header>
  )
}
