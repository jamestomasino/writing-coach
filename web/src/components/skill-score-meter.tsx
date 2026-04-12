import type { SkillScore } from '@/lib/types'
import { useTranslations } from 'next-intl'

function toneClasses(score: number) {
  switch (score) {
    case 5:
      return {
        filled: 'bg-emerald-500 dark:bg-emerald-300',
        empty: 'bg-stone-200 dark:bg-white/10',
        text: 'text-emerald-700 dark:text-emerald-200',
      }
    case 4:
      return {
        filled: 'bg-lime-500 dark:bg-lime-300',
        empty: 'bg-stone-200 dark:bg-white/10',
        text: 'text-lime-700 dark:text-lime-200',
      }
    case 3:
      return {
        filled: 'bg-amber-500 dark:bg-amber-300',
        empty: 'bg-stone-200 dark:bg-white/10',
        text: 'text-amber-700 dark:text-amber-200',
      }
    case 2:
      return {
        filled: 'bg-orange-500 dark:bg-orange-300',
        empty: 'bg-stone-200 dark:bg-white/10',
        text: 'text-orange-700 dark:text-orange-200',
      }
    default:
      return {
        filled: 'bg-rose-500 dark:bg-rose-300',
        empty: 'bg-stone-200 dark:bg-white/10',
        text: 'text-rose-700 dark:text-rose-200',
      }
  }
}

export function SkillScoreMeter({
  score,
  compact = false,
}: {
  score: Pick<SkillScore, 'skill' | 'score' | 'score_source' | 'score_version' | 'score_evidence'>
  compact?: boolean
}) {
  const t = useTranslations('skillScoreMeter')
  const value = Math.max(1, Math.min(5, score.score))
  const tone = toneClasses(value)
  const source = typeof score.score_source === 'string' ? score.score_source.trim() : ''
  const version = typeof score.score_version === 'string' ? score.score_version.trim() : ''
  const evidence = score.score_evidence && typeof score.score_evidence === 'object' ? score.score_evidence : undefined
  const appliedRules = Array.isArray(evidence?.applied_rules)
    ? evidence.applied_rules.filter((item): item is string => typeof item === 'string' && item.trim().length > 0)
    : []
  const domain = typeof evidence?.domain === 'string' ? evidence.domain : ''
  const findings = typeof evidence?.finding_count === 'number' ? evidence.finding_count : undefined
  const meaning = value >= 4 ? t('meaningStrong') : value === 3 ? t('meaningMixed') : t('meaningNeedsWork')
  const signalLabel = t('signalLabel', { name: score.skill })

  return (
    <div className={compact ? 'space-y-2' : 'space-y-2.5'}>
      <div className="flex items-center justify-between gap-4">
        <span className="font-semibold capitalize text-zinc-950 dark:text-white">{signalLabel}</span>
        <span className={`text-sm font-semibold ${tone.text}`}>{value}/5</span>
      </div>
      {source || version ? (
        <div className="flex flex-wrap items-center gap-2 text-xs text-zinc-600 dark:text-zinc-300">
          {source ? <span className="rounded-full border border-stone-200 px-2 py-0.5 dark:border-white/10">{source}</span> : null}
          {version ? <span className="rounded-full border border-stone-200 px-2 py-0.5 dark:border-white/10">{version}</span> : null}
        </div>
      ) : null}
      <div className="grid grid-cols-5 gap-1.5">
        {Array.from({ length: 5 }, (_, index) => (
          <div key={index} className={`h-2 rounded-full ${index < value ? tone.filled : tone.empty}`} />
        ))}
      </div>
      {!compact ? <div className="text-xs text-zinc-600 dark:text-zinc-400">{meaning}</div> : null}
      {!compact && evidence ? (
        <details className="rounded-lg border border-stone-200 px-3 py-2 text-xs text-zinc-700 dark:border-white/10 dark:text-zinc-300">
          <summary className="cursor-pointer font-medium text-zinc-900 dark:text-white">{t('whyThisScore')}</summary>
          <div className="mt-2 space-y-1">
            <div>{t('evidenceIntro')}</div>
            {domain ? <div>{t('domainLabel', { value: domain })}</div> : null}
            {typeof findings === 'number' ? <div>{t('findingsLabel', { count: findings })}</div> : null}
            {appliedRules.length > 0 ? (
              <div>
                <div>{t('checksLabel')}</div>
                <ul className="ml-4 list-disc">
                  {appliedRules.slice(0, 3).map((rule) => (
                    <li key={rule}>{rule}</li>
                  ))}
                </ul>
              </div>
            ) : null}
            <div className="pt-1 text-zinc-600 dark:text-zinc-400">{t('voiceNote')}</div>
          </div>
        </details>
      ) : null}
    </div>
  )
}
