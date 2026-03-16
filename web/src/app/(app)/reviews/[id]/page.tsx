import { ReviewView } from '@/components/review-view'

export default async function ReviewPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params
  return <ReviewView reviewId={Number(id)} />
}
