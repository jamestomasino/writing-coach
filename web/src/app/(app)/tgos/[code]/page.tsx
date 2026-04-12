import { TGODetailView } from '@/components/tgo-detail-view'
import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Objective guide',
}

export default async function TGODetailPage({ params }: { params: Promise<{ code: string }> }) {
  const { code } = await params
  return <TGODetailView code={decodeURIComponent(code)} />
}
