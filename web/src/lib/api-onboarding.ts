import { request } from './api-core'
import type { OnboardingOptions, OnboardingState } from './types'

export function getOnboarding() {
  return request<OnboardingState>('/api/onboarding')
}

export function getOnboardingOptions() {
  return request<OnboardingOptions>('/api/onboarding/options')
}

export function saveOnboarding(input: {
  mode?: 'create' | 'edit'
  writing_language: string
  writing_type: string
  assignment_format: string
  target_audience: string
  subject_matter: string
  experience_level: string
  desired_tone: string
  biggest_weaknesses: string[]
  desired_outcomes: string[]
  difficulty_intensity: string
  writing_goals: string
}) {
  return request<OnboardingState>('/api/onboarding', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}
