'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import { Button } from '@/components/button'
import { Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { resetAccountData } from '@/lib/api'

export function ResetDataCard() {
  const router = useRouter()
  const [working, setWorking] = useState(false)
  const [error, setError] = useState<string | null>(null)

  async function handleReset() {
    if (!window.confirm('Reset all coaching data for this account? This keeps your login, but deletes onboarding, assignments, submissions, reviews, and progress history.')) {
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
    <div className="rounded-2xl border border-rose-200 bg-rose-50 p-6 dark:border-rose-500/20 dark:bg-rose-500/10">
      <Subheading>Reset coaching data</Subheading>
      <Text className="mt-2">
        This clears your onboarding profile, generated track, assignments, submissions, reviews, and progress history. Your account and sign-in remain intact.
      </Text>
      {error ? <Text className="mt-3 text-sm text-rose-700 dark:text-rose-200">{error}</Text> : null}
      <div className="mt-4">
        <Button color="rose" onClick={handleReset} disabled={working}>
          {working ? 'Resetting…' : 'Reset all user data'}
        </Button>
      </div>
    </div>
  )
}
