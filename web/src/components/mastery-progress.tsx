import type { TGO } from '@/lib/types'
import { Text } from '@/components/text'

const stageOrder = ['emerging', 'developing', 'strong control', 'mastery evidence'] as const

function stageSegments(stage?: string) {
  const normalized = (stage ?? 'emerging').toLowerCase()
  const index = stageOrder.indexOf(normalized as (typeof stageOrder)[number])
  return index >= 0 ? index + 1 : 1
}

function stageLabel(stage?: string) {
  return stage ?? 'emerging'
}

export function MasteryProgress({
  tgo,
  tone = 'neutral',
}: {
  tgo: Pick<TGO, 'progress_mode' | 'mastery_percent' | 'mastery_stage' | 'mastery_evidence_count'>
  tone?: 'neutral' | 'blue'
}) {
  const percent = Math.max(0, Math.min(100, tgo.mastery_percent ?? 0))
  const filledSegments = stageSegments(tgo.mastery_stage)
  const label = stageLabel(tgo.mastery_stage)
  const evidence = tgo.mastery_evidence_count ?? 0

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
          <span>Mastery</span>
          <span>{percent}%</span>
        </div>
        <div className={`mt-2 h-2 rounded-full ${palette.track}`}>
          <div className={`h-2 rounded-full ${palette.fill}`} style={{ width: `${percent}%` }} />
        </div>
        {evidence > 0 ? (
          <Text className={`mt-2 text-sm ${palette.meta}`}>
            {label} across {evidence} review{evidence === 1 ? '' : 's'}
          </Text>
        ) : null}
      </div>
    )
  }

  return (
    <div className="mt-3">
      <div className={`text-xs font-medium uppercase tracking-[0.16em] ${palette.text}`}>Mastery</div>
      <div className="mt-2 grid grid-cols-4 gap-1.5">
        {stageOrder.map((segment, index) => (
          <div
            key={segment}
            className={`h-2 rounded-full ${index < filledSegments ? palette.fill : palette.emptySegment}`}
          />
        ))}
      </div>
      <Text className={`mt-2 text-sm ${palette.meta}`}>
        <span className={`font-semibold ${palette.body}`}>{label}</span>
        {evidence > 0 ? ` across ${evidence} review${evidence === 1 ? '' : 's'}` : ''}
      </Text>
    </div>
  )
}
