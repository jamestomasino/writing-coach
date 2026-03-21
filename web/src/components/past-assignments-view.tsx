'use client'

import { Badge } from '@/components/badge'
import { Button } from '@/components/button'
import { CardHeader } from '@/components/card-header'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { getAssignments } from '@/lib/api'
import type { AssignmentSummary } from '@/lib/types'
import { useRequiredAppSession } from '@/lib/use-required-app-session'
import { ArrowRightIcon } from '@heroicons/react/16/solid'
import { useEffect, useMemo, useState } from 'react'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'
import { formatLocalDateTime } from '@/lib/datetime'

export function PastAssignmentsView() {
  const { session, loading: sessionLoading, error: sessionError } = useRequiredAppSession('/assignments')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [assignments, setAssignments] = useState<AssignmentSummary[]>([])

  useEffect(() => {
    let cancelled = false
    async function load() {
      if (!session) {
        return
      }
      try {
        const items = await getAssignments()
        if (!cancelled) {
          setAssignments(items)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load assignment archive')
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
  }, [session])

  const currentAssignment = useMemo(() => assignments.find((item) => item.is_current), [assignments])
  const pastAssignments = useMemo(() => assignments.filter((item) => !item.is_current), [assignments])

  if (sessionLoading || loading) {
    return <LoadingState label="Loading past assignments…" />
  }
  if (sessionError || error) {
    return <AppErrorState title="Archive unavailable" error={sessionError ?? error ?? 'Archive unavailable'} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Assignments"
        title="Past assignments"
        intro="Look back at earlier assignments from the active track and revisit your drafts, feedback, and revisions."
        actions={
          <Button href="/" outline>
            Current assignment
          </Button>
        }
      />

      {currentAssignment ? (
        <WorkspaceCard>
          <CardHeader
            eyebrow="Current chain"
            title={currentAssignment.title}
            description="Your current assignment is still active. You can open its full history here."
            actions={
              <Button href={`/assignments/${currentAssignment.current_exercise_id}`} plain>
                View current timeline
              </Button>
            }
          />
        </WorkspaceCard>
      ) : null}

      {pastAssignments.length === 0 ? (
        <EmptyState
          title="No past assignments yet"
          body="Older assignments will show up here once you move on from the current one."
          actionHref="/"
          actionLabel="Open current assignment"
        />
      ) : (
        <div className="grid gap-6 xl:grid-cols-2">
          {pastAssignments.map((assignment) => (
            <WorkspaceCard key={assignment.root_exercise_id}>
              <CardHeader
                eyebrow={assignment.latest_step_label}
                title={assignment.title}
                description={`Latest activity ${formatLocalDateTime(assignment.latest_activity) ?? assignment.latest_activity}.`}
                actions={
                  <Button href={`/assignments/${assignment.current_exercise_id}`} color="dark/zinc">
                    Open timeline
                    <ArrowRightIcon />
                  </Button>
                }
              />
              <div className="mt-5 flex flex-wrap gap-2">
                {assignment.tgos.map((tgo) => (
                  <Badge key={tgo} color="zinc">
                    {tgo}
                  </Badge>
                ))}
              </div>
              <div className="mt-5 grid gap-4 sm:grid-cols-4">
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400">
                    Prompts
                  </div>
                  <Text className="mt-2 text-sm">{assignment.exercise_count}</Text>
                </div>
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400">
                    Drafts
                  </div>
                  <Text className="mt-2 text-sm">{assignment.draft_count}</Text>
                </div>
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400">
                    Feedback
                  </div>
                  <Text className="mt-2 text-sm">{assignment.review_count}</Text>
                </div>
                <div className="rounded-xl border border-stone-200 bg-stone-50 px-4 py-3 dark:border-white/10 dark:bg-white/5">
                  <div className="text-xs font-semibold tracking-[0.16em] text-zinc-500 uppercase dark:text-zinc-400">
                    Revisions
                  </div>
                  <Text className="mt-2 text-sm">{assignment.revision_count}</Text>
                </div>
              </div>
            </WorkspaceCard>
          ))}
        </div>
      )}
    </div>
  )
}
