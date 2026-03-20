import type { SkillScore } from '@/lib/types'

function toneClasses(score: number) {
  if (score >= 4) {
    return {
      filled: 'bg-emerald-500 dark:bg-emerald-300',
      empty: 'bg-stone-200 dark:bg-white/10',
      text: 'text-emerald-700 dark:text-emerald-200',
    }
  }
  if (score === 3) {
    return {
      filled: 'bg-amber-500 dark:bg-amber-300',
      empty: 'bg-stone-200 dark:bg-white/10',
      text: 'text-amber-700 dark:text-amber-200',
    }
  }
  return {
    filled: 'bg-rose-500 dark:bg-rose-300',
    empty: 'bg-stone-200 dark:bg-white/10',
    text: 'text-rose-700 dark:text-rose-200',
  }
}

export function SkillScoreMeter({
  score,
  compact = false,
}: {
  score: Pick<SkillScore, 'skill' | 'score'>
  compact?: boolean
}) {
  const value = Math.max(1, Math.min(5, score.score))
  const tone = toneClasses(value)

  return (
    <div className={compact ? 'space-y-2' : 'space-y-2.5'}>
      <div className="flex items-center justify-between gap-4">
        <span className="font-semibold capitalize text-zinc-950 dark:text-white">{score.skill}</span>
        <span className={`text-sm font-semibold ${tone.text}`}>{value}/5</span>
      </div>
      <div className="grid grid-cols-5 gap-1.5">
        {Array.from({ length: 5 }, (_, index) => (
          <div key={index} className={`h-2 rounded-full ${index < value ? tone.filled : tone.empty}`} />
        ))}
      </div>
    </div>
  )
}
