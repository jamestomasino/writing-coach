import type { Dashboard, TGO } from '@/lib/types'

export type LevelUpGuidanceMode = 'hold' | 'revise' | 'consolidate'

export type LevelUpGuidance = {
  mode: LevelUpGuidanceMode
  earlyCount: number
  strongCount: number
  totalCount: number
}

function isEarlyStage(tgo: TGO): boolean {
  if (tgo.progress_mode === 'percent') {
    return (tgo.mastery_percent ?? 0) < 70
  }
  const stage = (tgo.mastery_stage ?? '').trim().toLowerCase()
  return stage === '' || stage === 'emerging' || stage === 'developing'
}

function isStrongStage(tgo: TGO): boolean {
  if (tgo.progress_mode === 'percent') {
    return (tgo.mastery_percent ?? 0) >= 70
  }
  const stage = (tgo.mastery_stage ?? '').trim().toLowerCase()
  return stage === 'strong control' || stage === 'mastery evidence'
}

export function computeLevelUpGuidance(dashboard: Dashboard): LevelUpGuidance {
  const active = dashboard.active_tgos ?? []
  const totalCount = active.length
  const earlyCount = active.filter(isEarlyStage).length
  const strongCount = active.filter(isStrongStage).length

  if (dashboard.curriculum_state.progression_hold_active) {
    return { mode: 'hold', earlyCount, strongCount, totalCount }
  }
  if (earlyCount > 0) {
    return { mode: 'revise', earlyCount, strongCount, totalCount }
  }
  return { mode: 'consolidate', earlyCount, strongCount, totalCount }
}
