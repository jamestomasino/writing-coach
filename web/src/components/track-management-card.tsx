'use client'

import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { useTranslations } from 'next-intl'
import { WorkspaceCard } from './workspace-card'

export function TrackManagementCard({
  archiving,
  onArchive,
}: {
  archiving: boolean
  onArchive: () => void
}) {
  const t = useTranslations('trackManagementCard')
  return (
    <WorkspaceCard>
      <CardHeader
        eyebrow={t('eyebrow')}
        title={t('title')}
        description={t('description')}
      />
      <div className="mt-5 flex justify-end">
        <Button color="red" onClick={onArchive} disabled={archiving} data-testid="archive-track-button">
          {archiving ? t('archiving') : t('archive')}
        </Button>
      </div>
    </WorkspaceCard>
  )
}
