import { Button } from '@/components/button'
import { Callout } from '@/components/callout'
import { LoadingState } from '@/components/status-state'
import { KratosFlowView } from '@/components/kratos-flow-view'
import { ResetDataCard } from '@/components/reset-data-card'
import type { Metadata } from 'next'
import { Suspense } from 'react'

export const metadata: Metadata = {
  title: 'Settings',
}

export const dynamic = 'force-dynamic'

export default function SettingsPage() {
  return (
    <div className="grid w-full max-w-lg grid-cols-1 gap-8">
      <Suspense fallback={<LoadingState label="Loading account flow…" />}>
        <KratosFlowView kind="settings" />
      </Suspense>
      <Callout
        title="AI setup"
        body="Choose where AI help comes from, and manage the key used for assignments and feedback."
        actions={
          <Button href="/ai-settings" outline>
            Open AI settings
          </Button>
        }
      />
      <ResetDataCard />
    </div>
  )
}
