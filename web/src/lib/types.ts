export type RequestContext = {
  user_slug: string
  tree_slug: string
  user_id: number
  tree_id: number
}

export type AuthSession = {
  authenticated: boolean
  auth_mode: string
  onboarding_complete: boolean
  active_tree_slug?: string
  identity?: {
    subject: string
    email?: string
    name?: string
  }
  context?: RequestContext
}

export type TGO = {
  id: number
  code: string
  title: string
  description: string
  stage: string
  stage_order: number
  active_slot?: number
  prerequisites?: string[]
  mastery_hint?: string
}

export type Exercise = {
  id: number
  title: string
  brief: string
  constraints: string[]
  focus_skills: string[]
  tgo_codes: string[]
  success_criteria: string[]
  generation_kind: string
  provider_note?: string
}

export type Submission = {
  id: number
  exercise_id: number
  parent_submission_id?: number
  draft_number: number
  content: string
  word_count: number
  created_at: string
}

export type TGOAssessment = {
  tgo_code: string
  status: string
  evidence: string
}

export type Review = {
  id: number
  submission_id: number
  review_kind: string
  provider_note?: string
  summary: string
  strengths: string[]
  weaknesses: string[]
  analyzer_findings: string[]
  next_focus: string
  metric_word_count: number
  tgo_assessments: TGOAssessment[]
  completed_tgo_checks: TGOAssessment[]
  artifacts?: {
    analyzer_report?: Record<string, unknown>
    recommendation?: Record<string, unknown>
    comparison?: Record<string, unknown>
  }
}

export type Comparison = {
  summary: string
  word_delta: number
  added_words: string[]
  removed_words: string[]
  addressed_weaknesses: string[]
  persisting_weaknesses: string[]
}

export type Dashboard = {
  context: RequestContext
  curriculum_state: {
    id: number
    current_focus: string
    difficulty_level: number
    last_review_id: number
    updated_at: string
  }
  active_tgos: TGO[]
  completed_tgos: TGO[]
  upcoming_tgos: TGO[]
  progress_lines: string[]
  strongest_skills: string[]
  weakest_skills: string[]
  recurring_weaknesses: string[]
  recurring_findings: string[]
  recurring_completed_slips: string[]
  history: string[]
}

export type UserRecord = {
  id: number
  slug: string
  name: string
  active_tree_slug?: string
  created_at: string
}

export type OnboardingProfile = {
  writing_type: string
  experience_level: string
  desired_tone: string
  biggest_weaknesses: string[]
  desired_outcomes: string[]
  difficulty_intensity: string
  writing_goals: string
  generated_tree_slug: string
  template_key: string
}

export type OnboardingState = {
  onboarding_complete: boolean
  profile?: OnboardingProfile
}
