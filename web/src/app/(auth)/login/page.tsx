import { LoadingState } from '@/components/status-state'
import { KratosFlowView } from '@/components/kratos-flow-view'
import type { Metadata } from 'next'
import { Suspense } from 'react'

export const metadata: Metadata = {
  title: 'Login',
}

export const dynamic = 'force-dynamic'

export default function LoginPage() {
  return (
    <Suspense fallback={<LoadingState label="Loading account flow…" />}>
      <KratosFlowView kind="login" />
    </Suspense>
  )
}
