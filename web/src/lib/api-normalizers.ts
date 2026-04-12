import type { AssignmentTimeline, Dashboard, PlaygroundReview, Review, Tree } from './types'

export function arrayOrEmpty<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : []
}

export function normalizeDashboard(payload: Dashboard): Dashboard {
  return {
    ...payload,
    active_tgos: arrayOrEmpty(payload.active_tgos),
    completed_tgos: arrayOrEmpty(payload.completed_tgos),
    upcoming_tgos: arrayOrEmpty(payload.upcoming_tgos),
    progress_lines: arrayOrEmpty(payload.progress_lines),
    strongest_skills: arrayOrEmpty(payload.strongest_skills),
    weakest_skills: arrayOrEmpty(payload.weakest_skills),
    recurring_weaknesses: arrayOrEmpty(payload.recurring_weaknesses),
    recurring_findings: arrayOrEmpty(payload.recurring_findings),
    recurring_completed_slips: arrayOrEmpty(payload.recurring_completed_slips),
    history: arrayOrEmpty(payload.history),
  }
}

export function normalizeTree(tree: Tree): Tree {
  return {
    ...tree,
    seed_codes: arrayOrEmpty(tree.seed_codes),
    priority_skills: arrayOrEmpty(tree.priority_skills),
    tgos: arrayOrEmpty(tree.tgos).map((tgo) => ({
      ...tgo,
      prerequisites: arrayOrEmpty(tgo.prerequisites),
    })),
  }
}

export function normalizeReview(review: Review): Review {
  return {
    ...review,
    strengths: arrayOrEmpty(review.strengths),
    weaknesses: arrayOrEmpty(review.weaknesses),
    analyzer_findings: arrayOrEmpty(review.analyzer_findings),
    skill_scores: arrayOrEmpty(review.skill_scores),
    objective_scores: arrayOrEmpty(review.objective_scores),
    tgo_assessments: arrayOrEmpty(review.tgo_assessments),
    completed_tgo_checks: arrayOrEmpty(review.completed_tgo_checks),
    annotations: arrayOrEmpty(review.annotations),
    artifacts: review.artifacts
      ? {
          ...review.artifacts,
          annotations: arrayOrEmpty(review.artifacts.annotations),
        }
      : undefined,
  }
}

export function normalizeAssignmentTimeline(timeline: AssignmentTimeline): AssignmentTimeline {
  return {
    ...timeline,
    steps: arrayOrEmpty(timeline.steps).map((step) => ({
      ...step,
      review: step.review ? normalizeReview(step.review) : undefined,
    })),
  }
}

export function normalizePlaygroundReview(item: PlaygroundReview): PlaygroundReview {
  return {
    ...item,
    review: normalizeReview(item.review),
  }
}
