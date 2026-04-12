import type { Metadata } from 'next'
import { redirect } from 'next/navigation'

export const metadata: Metadata = {
  title: 'Skill objective guide',
}

export default async function TGODetailPage({ params }: { params: Promise<{ code: string }> }) {
  const { code } = await params
  redirect(`/skills/${encodeURIComponent(code)}`)
}
