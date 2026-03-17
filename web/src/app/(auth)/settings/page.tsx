import { LoadingState } from '@/components/status-state'
import { KratosFlowView } from '@/components/kratos-flow-view'
import type { Metadata } from 'next'
import { Suspense } from 'react'

export const metadata: Metadata = {
  title: 'Settings',
}

export const dynamic = 'force-dynamic'

export default function SettingsPage() {
  return (
    <Suspense fallback={<LoadingState label="Loading account flow…" />}>
      <KratosFlowView kind="settings" />
    </Suspense>
  )
}
