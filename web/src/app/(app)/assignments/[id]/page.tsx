import { AssignmentTimelineView } from '@/components/assignment-timeline-view'

export default async function AssignmentPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  return <AssignmentTimelineView exerciseId={Number(id)} />
}
