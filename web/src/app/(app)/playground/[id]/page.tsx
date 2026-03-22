import { PlaygroundView } from '@/components/playground-view'

export const metadata = {
  title: 'Playground',
}

export default async function PlaygroundSessionPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  return <PlaygroundView sessionId={Number(id)} />
}
