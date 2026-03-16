import { Suspense } from 'react'
import { CurrentAssignmentView } from '@/components/current-assignment-view'
import { LoadingState } from '@/components/status-state'

export default function Home() {
  return (
    <Suspense fallback={<LoadingState label="Loading current assignment…" />}>
      <CurrentAssignmentView />
    </Suspense>
  )
}
