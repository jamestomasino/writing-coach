'use client'

import { CurrentAssignmentView } from '@/components/current-assignment-view'
import { LoadingState } from '@/components/status-state'
import { Suspense } from 'react'

export default function Home() {
  return (
    <Suspense fallback={<LoadingState />}>
      <CurrentAssignmentView />
    </Suspense>
  )
}
