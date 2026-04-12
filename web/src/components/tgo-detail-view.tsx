'use client'

import { useEffect, useMemo, useState } from 'react'
import { Badge } from '@/components/badge'
import { PageHeader } from '@/components/page-header'
import { Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import { getSkillGraph } from '@/lib/api'
import { buildObjectiveConcepts } from '@/lib/objective-concepts'
import { buildObjectiveDetail } from '@/lib/objective-details'
import type { SkillGraphNode } from '@/lib/types'
import { AppErrorState, EmptyState, LoadingState } from './status-state'

export function TGODetailView({ code }: { code: string }) {
  const [nodes, setNodes] = useState<SkillGraphNode[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    async function load() {
      try {
        setLoading(true)
        const graph = await getSkillGraph()
        if (cancelled) {
          return
        }
        setNodes(graph.nodes ?? [])
      } catch (err) {
        if (cancelled) {
          return
        }
        setError(err instanceof Error ? err.message : 'Could not load objective guide.')
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
  }, [])

  const conceptData = useMemo(() => buildObjectiveConcepts(nodes), [nodes])
  const concept = conceptData.conceptByKey.get(code) ?? conceptData.conceptByKey.get(conceptData.conceptByCode.get(code) ?? '')
  const node = concept?.representative
  const objectiveDetail = node ? buildObjectiveDetail(node) : null

  if (loading) {
    return <LoadingState label="Loading objective guide..." />
  }

  if (error) {
    return <AppErrorState title="Skill objective guide unavailable" error={error} />
  }

  if (!node) {
    return <EmptyState title="Skill objective not found" body="This skill objective key is not in the current skill graph." actionHref="/skills" actionLabel="Back to library" />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Skill objective guide"
        title={concept?.title ?? node.title}
        intro={concept?.description ?? node.description}
        actions={<Badge color="zinc">{concept?.key ?? node.code}</Badge>}
      />

      {node.mastery_hint ? (
        <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-400/30 dark:bg-emerald-500/10">
          <div className="text-sm font-semibold text-emerald-900 dark:text-emerald-200">Mastery marker</div>
          <Text className="mt-2 text-sm text-emerald-900 dark:text-emerald-100">{node.mastery_hint}</Text>
        </div>
      ) : null}

      {objectiveDetail ? (
        <WorkspaceCard>
          <Subheading>What this objective trains</Subheading>
          <Text className="mt-3">{objectiveDetail.objectiveGoal}</Text>
          <Subheading className="mt-6">Why it matters now</Subheading>
          <Text className="mt-3">{objectiveDetail.whyThisObjective}</Text>

          <Subheading className="mt-6">What success looks like</Subheading>
          <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            {objectiveDetail.successLooksLike.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>

          <div className="mt-6 grid gap-4 lg:grid-cols-2">
            <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-400/30 dark:bg-emerald-500/10">
              <div className="text-sm font-semibold text-emerald-900 dark:text-emerald-200">Good example</div>
              <Text className="mt-2 text-sm text-emerald-900 dark:text-emerald-100">{objectiveDetail.goodExample}</Text>
            </div>
            <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-400/30 dark:bg-amber-500/10">
              <div className="text-sm font-semibold text-amber-900 dark:text-amber-200">Needs work example</div>
              <Text className="mt-2 text-sm text-amber-900 dark:text-amber-100">{objectiveDetail.badExample}</Text>
            </div>
          </div>
        </WorkspaceCard>
      ) : null}

    </div>
  )
}
