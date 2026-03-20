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
        title="AI provider"
        body="Manage the provider and key used for assignment and feedback generation."
        actions={
          <Button href="/ai-settings" outline>
            Open AI provider settings
          </Button>
        }
      />
      <ResetDataCard />
    </div>
  )
}
