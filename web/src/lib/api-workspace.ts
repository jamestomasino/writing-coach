import { request } from './api-core'
import { arrayOrEmpty, normalizeAssignmentTimeline, normalizeDashboard, normalizeReview, normalizeTree } from './api-normalizers'
import type { AIJob, AssignmentSummary, AssignmentTimeline, Comparison, Dashboard, Exercise, PlaygroundReviewInput, Review, Submission, Tree } from './types'

export function getDashboard() {
  return request<Dashboard>('/api/dashboard').then(normalizeDashboard)
}

export async function getTree(slug: string) {
  const payload = await request<{ tree: Tree }>(`/api/trees/${slug}`)
  return normalizeTree(payload.tree)
}

export async function getExercises(limit = 10) {
  const payload = await request<{ exercises: Exercise[] }>(`/api/exercises?limit=${limit}`)
  return payload.exercises
}

export async function getExercise(exerciseId: number) {
  const payload = await request<{ exercise: Exercise }>(`/api/exercises/${exerciseId}`)
  return payload.exercise
}

export async function getAssignmentTimeline(exerciseId: number) {
  const payload = await request<{ assignment: AssignmentTimeline }>(`/api/assignments/${exerciseId}`)
  return normalizeAssignmentTimeline(payload.assignment)
}

export async function getAssignments() {
  const payload = await request<{ assignments: AssignmentSummary[] }>('/api/assignments')
  return arrayOrEmpty(payload.assignments)
}

export async function getSubmissions(exerciseId: number, limit = 10) {
  const payload = await request<{ submissions: Submission[] }>(`/api/submissions?exercise_id=${exerciseId}&limit=${limit}`)
  return payload.submissions
}

export async function getSubmission(submissionId: number) {
  const payload = await request<{ submission: Submission }>(`/api/submissions/${submissionId}`)
  return payload.submission
}

export async function getReviews(submissionId: number, limit = 10) {
  const payload = await request<{ reviews: Review[] }>(`/api/reviews?submission_id=${submissionId}&limit=${limit}`)
  return payload.reviews.map(normalizeReview)
}

export async function getReview(reviewId: number) {
  const payload = await request<{ review: Review }>(`/api/reviews/${reviewId}`)
  return normalizeReview(payload.review)
}

export async function getComparison(submissionId: number, against?: number) {
  const query = against ? `?submission_id=${submissionId}&against=${against}` : `?submission_id=${submissionId}`
  const payload = await request<{ comparison: Comparison }>(`/api/compare${query}`)
  return payload.comparison
}

export async function createAssignment(tgoCodes?: string[]) {
  const payload = await request<{ job: AIJob }>('/api/prompts/next', {
    method: 'POST',
    body: JSON.stringify(tgoCodes && tgoCodes.length > 0 ? { tgo_codes: tgoCodes } : {}),
  })
  return payload.job
}

export async function acceptAssignment(exercise: Exercise) {
  const payload = await request<{ exercise: Exercise }>('/api/prompts/accept', {
    method: 'POST',
    body: JSON.stringify({
      title: exercise.title,
      brief: exercise.brief,
      constraints: exercise.constraints,
      focus_skills: exercise.focus_skills,
      tgo_codes: exercise.tgo_codes,
      success_criteria: exercise.success_criteria,
      generation_kind: exercise.generation_kind,
      provider_note: exercise.provider_note ?? '',
    }),
  })
  return payload.exercise
}

export async function createRevisionAssignment(submissionId: number) {
  const payload = await request<{ job: AIJob }>('/api/prompts/revise', {
    method: 'POST',
    body: JSON.stringify({ submission_id: submissionId }),
  })
  return payload.job
}

export async function closeAssignment(exerciseId: number) {
  await request<{ ok: boolean }>(`/api/assignments/${exerciseId}/close`, {
    method: 'POST',
    body: '{}',
  })
}

export async function submitDraft(input: { exerciseId: number; content: string; parentSubmissionId?: number }) {
  const payload = await request<{ submission: Submission }>('/api/submissions', {
    method: 'POST',
    body: JSON.stringify({
      exercise_id: input.exerciseId,
      parent_submission_id: input.parentSubmissionId ?? 0,
      content: input.content,
    }),
  })
  return payload.submission
}

export async function reviewSubmission(submissionId: number) {
  const payload = await request<{ job: AIJob }>('/api/reviews', {
    method: 'POST',
    body: JSON.stringify({ submission_id: submissionId }),
  })
  return payload.job
}

export async function getReviewJob(submissionId: number) {
  const payload = await request<{ job: AIJob }>(`/api/review-jobs?submission_id=${submissionId}`)
  return payload.job
}

export async function createPlaygroundReview(input: PlaygroundReviewInput) {
  const payload = await request<{ job: AIJob }>('/api/playground/review', {
    method: 'POST',
    body: JSON.stringify({
      content: input.content,
      writing_language: input.writing_language ?? '',
      writing_type: input.writing_type ?? '',
      assignment_format: input.assignment_format ?? '',
      coaching_brief: input.coaching_brief ?? '',
    }),
  })
  return payload.job
}

export async function getAIJob(jobId: number) {
  const payload = await request<{ job: AIJob }>(`/api/jobs/${jobId}`)
  const job = payload.job
  return {
    ...job,
    result: job.result
      ? {
          exercise: job.result.exercise,
          review: job.result.review ? normalizeReview(job.result.review) : undefined,
        }
      : undefined,
  }
}
