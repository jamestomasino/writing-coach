export type RequestContext = {
  user_slug: string
  tree_slug: string
  user_id: number
  tree_id: number
}

export type AuthSession = {
  authenticated: boolean
  auth_mode: string
  is_admin: boolean
  onboarding_complete: boolean
  ai_provider_ready: boolean
  ai_effective_provider?: string
  ai_system_fallback: boolean
  ai_has_personal_key: boolean
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
  progress_mode?: string
  mastery_stage?: string
  mastery_percent?: number
  mastery_evidence_count?: number
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
  source_submission_id?: number
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
  tgo_title?: string
  status: string
  evidence: string
}

export type SkillScore = {
  skill: string
  score: number
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
  skill_scores: SkillScore[]
  tgo_assessments: TGOAssessment[]
  completed_tgo_checks: TGOAssessment[]
  annotations: ReviewAnnotation[]
  artifacts?: {
    analyzer_report?: Record<string, unknown>
    recommendation?: Record<string, unknown>
    comparison?: Comparison
    annotations?: ReviewAnnotation[]
  }
}

export type AIProviderSettings = {
  provider?: string
  base_url_override?: string
  prompt_model_override?: string
  review_model_override?: string
  enabled: boolean
  has_key: boolean
  key_last4?: string
  validated_at?: string
  last_validation_error?: string
  effective_provider: string
  system_fallback: boolean
  ready: boolean
}

export type ReviewAnnotation = {
  quote: string
  tgo_code: string
  tgo_title?: string
  category: string
  comment: string
  severity: string
}

export type ReviewJob = {
  id: number
  submission_id: number
  review_id?: number
  status: string
  attempt_count: number
  max_attempts: number
  last_error?: string
  created_at: string
  updated_at: string
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
  completed_assignments: number
  progress_lines: string[]
  strongest_skills: string[]
  weakest_skills: string[]
  recurring_weaknesses: string[]
  recurring_findings: string[]
  recurring_completed_slips: string[]
  history: {
    title: string
    tgos: string[]
  }[]
}

export type AssignmentTimelineStep = {
  id: string
  kind: string
  title: string
  label: string
  created_at: string
  exercise_id?: number
  submission_id?: number
  review_id?: number
  draft_number?: number
  exercise?: Exercise
  submission?: Submission
  review?: Review
}

export type AssignmentTimeline = {
  root_exercise_id: number
  current_exercise_id: number
  title: string
  is_current?: boolean
  latest_step_id?: string
  steps: AssignmentTimelineStep[]
}

export type AssignmentSummary = {
  root_exercise_id: number
  current_exercise_id: number
  title: string
  latest_activity: string
  latest_step_label: string
  exercise_count: number
  draft_count: number
  review_count: number
  revision_count: number
  tgos: string[]
  is_current?: boolean
}

export type Tree = {
  id: number
  slug: string
  title: string
  description: string
  seed_codes: string[]
  priority_skills: string[]
  tgos: TGO[]
  created_at?: string
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
  assignment_format: string
  target_audience: string
  subject_matter: string
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
  starter_tgo_codes?: string[]
  recommended_regions?: string[]
}

export type OnboardingOption = {
  value: string
  label: string
}

export type OnboardingOptions = {
  writing_domains: OnboardingOption[]
  assignment_formats: OnboardingOption[]
  experience_levels: OnboardingOption[]
  difficulty_levels: OnboardingOption[]
  weaknesses: OnboardingOption[]
  desired_outcomes: OnboardingOption[]
}
