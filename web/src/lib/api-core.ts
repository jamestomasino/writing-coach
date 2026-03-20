'use client'

type ErrorBody = { error?: string }

export async function request<T>(path: string, init?: RequestInit): Promise<T> {
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
