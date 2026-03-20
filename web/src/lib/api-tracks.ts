import { request } from './api-core'
import type { UserTrack } from './types'

export async function listTracks() {
  const payload = await request<{ tracks: UserTrack[] }>('/api/tracks')
  return payload.tracks
}

export async function setActiveTrack(treeSlug: string) {
  const payload = await request<{ tracks: UserTrack[] }>('/api/tracks/active', {
    method: 'PUT',
    body: JSON.stringify({ tree_slug: treeSlug }),
  })
  return payload.tracks
}

export async function archiveTrack(treeSlug: string) {
  const payload = await request<{ tracks: UserTrack[] }>(`/api/tracks/${treeSlug}/archive`, {
    method: 'POST',
    body: JSON.stringify({}),
  })
  return payload.tracks
}
