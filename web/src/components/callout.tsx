import clsx from 'clsx'
import { Eyebrow } from '@/components/eyebrow'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'

const toneClasses = {
  neutral: 'border-stone-200 bg-stone-50 dark:border-white/10 dark:bg-white/5',
  info: 'border-cyan-200 bg-cyan-50 dark:border-cyan-500/20 dark:bg-cyan-500/10',
  active: 'border-blue-200 bg-blue-50 dark:border-blue-500/20 dark:bg-blue-500/10',
  success: 'border-green-200 bg-green-50 dark:border-green-500/20 dark:bg-green-500/10',
  warning: 'border-amber-200 bg-amber-50 dark:border-amber-500/20 dark:bg-amber-500/10',
  danger: 'border-rose-200 bg-rose-50 dark:border-rose-500/20 dark:bg-rose-500/10',
} as const

const eyebrowTones = {
  neutral: 'zinc',
  info: 'cyan',
  active: 'amber',
  success: 'emerald',
  warning: 'amber',
  danger: 'amber',
} as const

export function Callout({
  tone = 'neutral',
  eyebrow,
  title,
  body,
  actions,
  className,
  titleLevel = 'subheading',
  children,
}: {
  tone?: keyof typeof toneClasses
  eyebrow?: React.ReactNode
  title?: React.ReactNode
  body?: React.ReactNode
  actions?: React.ReactNode
  className?: string
  titleLevel?: 'subheading' | 'heading'
  children?: React.ReactNode
}) {
  return (
    <div className={clsx('rounded-2xl border p-4', toneClasses[tone], className)}>
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          {eyebrow ? <Eyebrow tone={eyebrowTones[tone]}>{eyebrow}</Eyebrow> : null}
          {title
            ? titleLevel === 'heading'
              ? <Heading level={2} className="text-base/6">{title}</Heading>
              : <Subheading>{title}</Subheading>
            : null}
          {body ? <Text className="mt-2">{body}</Text> : null}
        </div>
        {actions ? <div className="shrink-0">{actions}</div> : null}
      </div>
      {children ? <div className="mt-4">{children}</div> : null}
    </div>
  )
}
