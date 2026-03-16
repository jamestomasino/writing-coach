'use client'

import { useEffect, useState } from 'react'
import { Badge } from '@/components/badge'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { getComparison, getSubmission } from '@/lib/api'
import type { Comparison, Submission } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

export function CompareView({ submissionId }: { submissionId: number }) {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [comparison, setComparison] = useState<Comparison | null>(null)
  const [submission, setSubmission] = useState<Submission | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        const [comparisonData, submissionData] = await Promise.all([getComparison(submissionId), getSubmission(submissionId)])
        if (!cancelled) {
          setComparison(comparisonData)
          setSubmission(submissionData)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load comparison')
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
  }, [submissionId])

  if (loading) {
    return <LoadingState label="Loading comparison…" />
  }
  if (error || !comparison || !submission) {
    return <EmptyState title="Comparison unavailable" body={error ?? 'The requested comparison is not available yet.'} actionHref="/" actionLabel="Back to assignment" />
  }

  return (
    <div className="space-y-8">
      <header>
        <Heading>Revision compare</Heading>
        <Text className="mt-2 max-w-3xl">{comparison.summary}</Text>
      </header>

      <div className="grid gap-8 lg:grid-cols-3">
        <WorkspaceCard>
          <Subheading>Draft</Subheading>
          <Text className="mt-2">Revision draft #{submission.draft_number}</Text>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>Word delta</Subheading>
          <div className="mt-3">
            <Badge color={comparison.word_delta >= 0 ? 'green' : 'amber'}>
              {comparison.word_delta >= 0 ? `+${comparison.word_delta}` : comparison.word_delta}
            </Badge>
          </div>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>Persistence check</Subheading>
          <Text className="mt-2">{comparison.persisting_weaknesses.length} weaknesses still carrying forward.</Text>
        </WorkspaceCard>
      </div>

      <div className="grid gap-8 xl:grid-cols-2">
        <WorkspaceCard>
          <Subheading>Addressed weaknesses</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {comparison.addressed_weaknesses.length === 0 ? <li>No weaknesses were marked resolved yet.</li> : null}
            {comparison.addressed_weaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
        <WorkspaceCard>
          <Subheading>Persisting weaknesses</Subheading>
          <ul className="mt-4 space-y-3 text-sm text-zinc-700 dark:text-zinc-300">
            {comparison.persisting_weaknesses.length === 0 ? <li>No persisting weaknesses were flagged.</li> : null}
            {comparison.persisting_weaknesses.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
      </div>
    </div>
  )
}
