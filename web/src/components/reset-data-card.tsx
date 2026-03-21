'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/button'
import { Callout } from '@/components/callout'
import { resetAccountData } from '@/lib/api'

export function ResetDataCard() {
  const router = useRouter()
  const [working, setWorking] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleReset() {
    if (!window.confirm('Reset all writing progress for this account? Your sign-in will stay the same.')) {
      return
    }

    try {
      setWorking(true)
      setError(null)
      await resetAccountData()
      router.push('/onboarding')
      router.refresh()
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Could not reset account data')
    } finally {
      setWorking(false)
    }
  }

  return (
    <Callout
      tone="danger"
      eyebrow="Danger zone"
      title="Reset coaching data"
      body="Delete your practice path setup, assignments, reviews, and progress history. Your account and sign-in stay intact."
      actions={
        <Button color="rose" onClick={handleReset} disabled={working}>
          {working ? 'Resetting…' : 'Reset all coaching data'}
        </Button>
      }
    >
      {error ? <div className="text-sm text-rose-700 dark:text-rose-200">{error}</div> : null}
    </Callout>
  )
}
