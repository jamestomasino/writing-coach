'use client'

import { getAssignmentTimeline } from '@/lib/api'
import type { AssignmentTimeline } from '@/lib/types'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import { useEffect, useState } from 'react'

export function useAssignmentTimeline(exerciseId: number) {
  const { session, loading: sessionLoading, error: sessionError } = useRequiredAppSession(`/assignments/${exerciseId}`)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [assignment, setAssignment] = useState<AssignmentTimeline | null>(null)
  const [selectedStepID, setSelectedStepID] = useState('')

  useEffect(() => {
    let cancelled = false

    async function load() {
      if (!session) {
        return
      }
      try {
        const data = await getAssignmentTimeline(exerciseId)
        if (cancelled) {
          return
        }
        setAssignment(data)
        setSelectedStepID(data.latest_step_id ?? data.steps[0]?.id ?? '')
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load assignment timeline')
        }
      } finally {
        if (!cancelled) {
          setLoading(false)
        }
      }
    }

    void load()
    return () => {
      cancelled = true
    }
  }, [exerciseId, session])

  function selectStep(stepID: string) {
    setSelectedStepID(stepID)
    document.getElementById(stepID)?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }

  return {
    sessionLoading,
    sessionError,
    loading,
    error,
    assignment,
    selectedStepID,
    selectStep,
  }
}
