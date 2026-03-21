'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { useTranslations } from 'next-intl'
import { Button } from '@/components/button'
import { Callout } from '@/components/callout'
import { resetAccountData } from '@/lib/api'

export function ResetDataCard() {
  const t = useTranslations('resetDataCard')
  const router = useRouter()
  const [working, setWorking] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleReset() {
    if (!window.confirm(t('confirm'))) {
      return
    }

    try {
      setWorking(true)
      setError(null)
      await resetAccountData()
      router.push('/onboarding')
      router.refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : t('error'))
    } finally {
      setWorking(false)
    }
  }

  return (
    <Callout
      tone="danger"
      eyebrow={t('eyebrow')}
      title={t('title')}
      body={t('body')}
      actions={
        <Button color="rose" onClick={handleReset} disabled={working}>
          {working ? t('resetting') : t('action')}
        </Button>
      }
    >
      {error ? <div className="text-sm text-rose-700 dark:text-rose-200">{error}</div> : null}
    </Callout>
  )
}
