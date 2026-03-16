'use client'

import type { AuthSession, Comparison, Dashboard, Exercise, Review, Submission, UserRecord } from './types'

type ErrorBody = { error?: string }

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(path, {
    ...init,
    credentials: 'include',
    headers: {
      'Content-Type': 'application/json',
      ...(init?.headers ?? {}),
    },
  })
  if (!response.ok) {
    let message = `${response.status} ${response.statusText}`
    try {
      const payload = (await response.json()) as ErrorBody
      if (payload.error) {
        message = payload.error
      }
    } catch {}
    throw new Error(message)
  }
  return response.json() as Promise<T>
}

export function getSession() {
  return request<AuthSession>('/api/auth/session')
}

export function getDashboard() {
  return request<Dashboard>('/api/dashboard')
}

export async function getExercises(limit = 10) {
  const payload = await request<{ exercises: Exercise[] }>(`/api/exercises?limit=${limit}`)
  return payload.exercises
}

export async function getExercise(exerciseId: number) {
  const payload = await request<{ exercise: Exercise }>(`/api/exercises/${exerciseId}`)
  return payload.exercise
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
  return payload.reviews
}

export async function getReview(reviewId: number) {
  const payload = await request<{ review: Review }>(`/api/reviews/${reviewId}`)
  return payload.review
}

export async function getComparison(submissionId: number, against?: number) {
  const query = against ? `?submission_id=${submissionId}&against=${against}` : `?submission_id=${submissionId}`
  const payload = await request<{ comparison: Comparison }>(`/api/compare${query}`)
  return payload.comparison
}

export async function createAssignment(tgoCodes?: string[]) {
  const payload = await request<{ exercise: Exercise }>('/api/prompts/next', {
    method: 'POST',
    body: JSON.stringify(tgoCodes && tgoCodes.length > 0 ? { tgo_codes: tgoCodes } : {}),
  })
  return payload.exercise
}

export async function createRevisionAssignment(submissionId: number) {
  const payload = await request<{ exercise: Exercise }>('/api/prompts/revise', {
    method: 'POST',
    body: JSON.stringify({ submission_id: submissionId }),
  })
  return payload.exercise
}

export async function submitDraft(input: {
  exerciseId: number
  content: string
  parentSubmissionId?: number
}) {
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
  const payload = await request<{ review: Review }>('/api/reviews', {
    method: 'POST',
    body: JSON.stringify({ submission_id: submissionId }),
  })
  return payload.review
}

export function listAdmins() {
  return request<{ admins: string[] }>('/api/admins')
}

export async function listUsers() {
  const payload = await request<{ users: UserRecord[] }>('/api/users')
  return payload.users
}

export async function provisionUser(input: { slug: string; name: string }) {
  return request('/api/users', {
    method: 'POST',
    body: JSON.stringify(input),
  })
}
