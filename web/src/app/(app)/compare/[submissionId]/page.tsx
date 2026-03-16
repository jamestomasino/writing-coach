import { CompareView } from '@/components/compare-view'

export default async function ComparePage({ params }: { params: Promise<{ submissionId: string }> }) {
  const { submissionId } = await params
  return <CompareView submissionId={Number(submissionId)} />
}
