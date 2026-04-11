import type { TGO } from '@/lib/types'

export type SkillLevelUpState = {
  mode: 'ready' | 'building_history' | 'consolidating'
  evidenceCount: number
  remainingHistory: number
}

export function skillLevelUpState(tgo: Pick<TGO, 'mastery_stage' | 'mastery_evidence_count'>): SkillLevelUpState {
  const stage = (tgo.mastery_stage ?? '').trim().toLowerCase()
  const evidenceCount = Math.max(0, tgo.mastery_evidence_count ?? 0)
  if (stage === 'mastery evidence') {
    return {
      mode: 'ready',
      evidenceCount,
      remainingHistory: 0,
    }
  }
  const remainingHistory = Math.max(0, 3 - evidenceCount)
  if (remainingHistory > 0) {
    return {
      mode: 'building_history',
      evidenceCount,
      remainingHistory,
    }
  }
  return {
    mode: 'consolidating',
    evidenceCount,
    remainingHistory: 0,
  }
}
