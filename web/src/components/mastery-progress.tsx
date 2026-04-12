import type { TGO } from '@/lib/types'
import { Text } from '@/components/text'

const stageOrder = ['emerging', 'developing', 'strong control', 'mastery evidence'] as const
const stageLadder = ['unstarted', ...stageOrder] as const

function titleCase(value: string) {
  return value.replace(/\b\w/g, (match) => match.toUpperCase())
}

function stageSegments(stage?: string, evidenceCount?: number) {
  const label = stageLabel(stage, evidenceCount)
  if (label === 'unstarted') {
    return 0
  }
  const normalized = label.toLowerCase()
  const index = stageOrder.indexOf(normalized as (typeof stageOrder)[number])
  return index >= 0 ? index + 1 : 1
}

function stageLabel(stage?: string, evidenceCount?: number) {
  const normalized = (stage ?? '').trim().toLowerCase()
  if (!normalized && (evidenceCount ?? 0) === 0) {
    return 'unstarted'
  }
  if (normalized === 'emerging' && (evidenceCount ?? 0) === 0) {
    return 'unstarted'
  }
  return normalized || 'emerging'
}

function stageIndex(label: string) {
  return stageLadder.indexOf(label as (typeof stageLadder)[number])
}

function MasteryStageWithTooltip({
  label,
  tone,
}: {
  label: string
  tone: {
    body: string
    meta: string
  }
}) {
  const currentIndex = stageIndex(label)
  return (
    <span className="group/stage relative inline-flex items-center">
      <span className={`cursor-help border-b border-dotted border-current pb-px font-semibold ${tone.body}`}>
        Stage: {titleCase(label)}
      </span>
      <span
        role="tooltip"
        className="pointer-events-none absolute bottom-[calc(100%+0.5rem)] left-0 z-20 w-52 rounded-lg border border-stone-200 bg-white p-2 opacity-0 shadow-lg transition-opacity group-hover/stage:opacity-100 dark:border-white/10 dark:bg-zinc-900"
      >
        <span className="text-xs font-semibold uppercase tracking-[0.16em] text-zinc-500 dark:text-zinc-400">Mastery stages</span>
        <ul className="mt-1 space-y-1">
          {stageLadder.map((stage, index) => {
            const active = index === currentIndex
            return (
              <li
                key={stage}
                className={`rounded px-2 py-1 text-xs ${active ? 'bg-cyan-500/15 font-semibold text-cyan-900 dark:text-cyan-200' : 'text-zinc-600 dark:text-zinc-300'}`}
              >
                {titleCase(stage)}
              </li>
            )
          })}
        </ul>
      </span>
    </span>
  )
}

function stageFillClass(filledSegments: number, tone: 'neutral' | 'blue') {
  const level = Math.max(1, Math.min(stageOrder.length, filledSegments)) - 1
  if (tone === 'blue') {
    switch (level) {
      case 0:
        return 'bg-blue-500 dark:bg-blue-300'
      case 1:
        return 'bg-cyan-500 dark:bg-cyan-300'
      case 2:
        return 'bg-amber-400 dark:bg-amber-300'
      default:
        return 'bg-yellow-400 dark:bg-yellow-200'
    }
  }

  switch (level) {
    case 0:
      return 'bg-stone-500 dark:bg-stone-300'
    case 1:
      return 'bg-amber-500 dark:bg-amber-300'
    case 2:
      return 'bg-yellow-500 dark:bg-yellow-300'
    default:
      return 'bg-yellow-400 dark:bg-yellow-200'
  }
}

export function MasteryProgress({
  tgo,
  tone = 'neutral',
}: {
  tgo: Pick<TGO, 'progress_mode' | 'mastery_percent' | 'mastery_stage' | 'mastery_evidence_count'>
  tone?: 'neutral' | 'blue'
}) {
  const percent = Math.max(0, Math.min(100, tgo.mastery_percent ?? 0))
  const evidence = tgo.mastery_evidence_count ?? 0
  const label = stageLabel(tgo.mastery_stage, evidence)
  const filledSegments = stageSegments(tgo.mastery_stage, evidence)

  const palette =
    tone === 'blue'
      ? {
          track: 'bg-blue-200/70 dark:bg-blue-200/15',
          fill: 'bg-gradient-to-r from-blue-500 via-cyan-500 to-amber-300 dark:from-blue-300 dark:via-cyan-300 dark:to-amber-200',
          emptySegment: 'bg-blue-200/70 dark:bg-blue-200/15',
          text: 'text-blue-800/80 dark:text-blue-200/80',
          body: 'text-blue-950 dark:text-blue-100',
          meta: 'text-blue-900/80 dark:text-blue-200/80',
        }
      : {
          track: 'bg-stone-200 dark:bg-white/10',
          fill: 'bg-gradient-to-r from-stone-500 via-amber-500 to-yellow-400 dark:from-stone-300 dark:via-amber-300 dark:to-yellow-200',
          emptySegment: 'bg-stone-200 dark:bg-white/10',
          text: 'text-zinc-500 dark:text-zinc-400',
          body: 'text-zinc-950 dark:text-white',
          meta: 'text-zinc-600 dark:text-zinc-300',
        }

  if (tgo.progress_mode === 'percent') {
    return (
      <div className="mt-3">
        <div className={`flex items-center justify-between gap-4 text-xs font-medium uppercase tracking-[0.16em] ${palette.text}`}>
          <span>Mastery stage</span>
          <span>{percent}%</span>
        </div>
        <div className={`mt-2 h-2 rounded-full ${palette.track}`}>
          <div className={`h-2 rounded-full ${palette.fill}`} style={{ width: `${percent}%` }} />
        </div>
        <Text className={`mt-2 text-sm ${palette.meta}`}>
          <MasteryStageWithTooltip label={label} tone={{ body: palette.body, meta: palette.meta }} />
          {evidence > 0 ? ` across ${evidence} review${evidence === 1 ? '' : 's'}` : ''}
        </Text>
      </div>
    )
  }

  return (
    <div className="mt-3">
      <div className={`text-xs font-medium uppercase tracking-[0.16em] ${palette.text}`}>Mastery stage</div>
      <div className="mt-2 grid grid-cols-4 gap-1.5">
        {stageOrder.map((segment, index) => (
          <div
            key={segment}
            className={`h-2 rounded-full ${index < filledSegments ? stageFillClass(filledSegments, tone) : palette.emptySegment}`}
          />
        ))}
      </div>
      <Text className={`mt-2 text-sm ${palette.meta}`}>
        <MasteryStageWithTooltip label={label} tone={{ body: palette.body, meta: palette.meta }} />
        {evidence > 0 ? ` across ${evidence} review${evidence === 1 ? '' : 's'}` : ''}
      </Text>
    </div>
  )
}
