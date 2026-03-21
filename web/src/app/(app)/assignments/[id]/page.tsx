import { AssignmentTimelineView } from '@/components/assignment-timeline-view'

export default async function AssignmentPage({
  params,
  searchParams,
}: {
  params: Promise<{ id: string }>
  searchParams: Promise<{ completed?: string }>
}) {
  const { id } = await params
  const { completed } = await searchParams
  return <AssignmentTimelineView exerciseId={Number(id)} showCompletionState={completed === '1'} />
}
