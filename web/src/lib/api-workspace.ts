import { request } from './api-core'
import { arrayOrEmpty, normalizeAssignmentTimeline, normalizeDashboard, normalizePlaygroundReview, normalizeReview, normalizeTree } from './api-normalizers'
import type { AIJob, AssignmentSummary, AssignmentTimeline, Comparison, Dashboard, Exercise, PlaygroundDraft, PlaygroundReview, PlaygroundReviewInput, PlaygroundSession, Review, SkillGraph, Submission, Tree } from './types'

export function getDashboard() {
  return request<Dashboard>('/api/dashboard').then(normalizeDashboard)
}

export async function getTree(slug: string) {
  const payload = await request<{ tree: Tree }>(`/api/trees/${slug}`)
  return normalizeTree(payload.tree)
}

export async function getSkillGraph() {
  const payload = await request<{ graph: SkillGraph }>('/api/skill-graph')
  return payload.graph
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

export async function reopenAssignment(exerciseId: number) {
  await request<{ ok: boolean }>(`/api/assignments/${exerciseId}/reopen`, {
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

export async function createPlaygroundSession(input: PlaygroundReviewInput) {
  const payload = await request<{ session: PlaygroundSession }>('/api/playground/sessions', {
    method: 'POST',
    body: JSON.stringify({
      content: input.content,
      writing_language: input.writing_language ?? '',
      writing_type: input.writing_type ?? '',
      assignment_format: input.assignment_format ?? '',
      coaching_brief: input.coaching_brief ?? '',
    }),
  })
  return payload.session
}

export async function updatePlaygroundSession(sessionId: number, input: PlaygroundReviewInput) {
  const payload = await request<{ session: PlaygroundSession }>(`/api/playground/sessions/${sessionId}`, {
    method: 'PUT',
    body: JSON.stringify({
      content: input.content,
      writing_language: input.writing_language ?? '',
      writing_type: input.writing_type ?? '',
      assignment_format: input.assignment_format ?? '',
      coaching_brief: input.coaching_brief ?? '',
    }),
  })
  return payload.session
}

export async function getPlaygroundSessions(limit = 50, cursor?: number) {
  const query = new URLSearchParams()
  query.set('limit', String(limit))
  if (typeof cursor === 'number' && Number.isFinite(cursor) && cursor > 0) {
    query.set('cursor', String(cursor))
  }
  const payload = await request<{ sessions: PlaygroundSession[]; next_cursor?: number }>(`/api/playground/sessions?${query.toString()}`)
  return {
    sessions: arrayOrEmpty(payload.sessions),
    nextCursor: typeof payload.next_cursor === 'number' ? payload.next_cursor : undefined,
  }
}

export async function getPlaygroundSession(sessionId: number) {
  const payload = await request<{ session: PlaygroundSession }>(`/api/playground/sessions/${sessionId}`)
  return payload.session
}

export async function createPlaygroundSessionReview(sessionId: number) {
  const payload = await request<{ job: AIJob }>(`/api/playground/sessions/${sessionId}/reviews`, {
    method: 'POST',
    body: '{}',
  })
  return payload.job
}

export async function createPlaygroundDraft(sessionId: number) {
  const payload = await request<{ draft: PlaygroundDraft; session: PlaygroundSession }>(`/api/playground/sessions/${sessionId}/drafts`, {
    method: 'POST',
    body: '{}',
  })
  return payload
}

export async function getPlaygroundSessionDrafts(sessionId: number, limit = 50) {
  const payload = await request<{ drafts: PlaygroundDraft[] }>(`/api/playground/sessions/${sessionId}/drafts?limit=${limit}`)
  return arrayOrEmpty(payload.drafts)
}

export async function getPlaygroundSessionReviews(sessionId: number, limit = 20) {
  const payload = await request<{ reviews: PlaygroundReview[] }>(`/api/playground/sessions/${sessionId}/reviews?limit=${limit}`)
  return arrayOrEmpty(payload.reviews).map(normalizePlaygroundReview)
}

export async function getPlaygroundReview(reviewId: number) {
  const payload = await request<{ review: PlaygroundReview }>(`/api/playground/reviews/${reviewId}`)
  return normalizePlaygroundReview(payload.review)
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
