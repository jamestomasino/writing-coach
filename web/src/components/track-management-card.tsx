'use client'

import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { WorkspaceCard } from './workspace-card'

export function TrackManagementCard({
  archiving,
  onArchive,
}: {
  archiving: boolean
  onArchive: () => void
}) {
  return (
    <WorkspaceCard>
      <CardHeader
        eyebrow="Practice path settings"
        title="Archive this practice path"
        description="This hides the practice path from the switcher but keeps its assignment history and progress."
      />
      <div className="mt-5 flex justify-end">
        <Button color="red" onClick={onArchive} disabled={archiving} data-testid="archive-track-button">
          {archiving ? 'Archiving…' : 'Archive practice path'}
        </Button>
      </div>
    </WorkspaceCard>
  )
}
