import type { Metadata } from 'next'
import { TGODetailView } from '@/components/tgo-detail-view'

type Params = {
  slug: string
}

export const metadata: Metadata = {
  title: 'Skill Objectives',
}

export default async function SkillObjectiveDetailPage({ params }: { params: Promise<Params> }) {
  const { slug } = await params
  return <TGODetailView code={decodeURIComponent(slug)} />
}
