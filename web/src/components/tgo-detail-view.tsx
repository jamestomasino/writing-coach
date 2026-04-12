'use client'

import { useEffect, useMemo, useState } from 'react'
import { Badge } from '@/components/badge'
import { Link } from '@/components/link'
import { PageHeader } from '@/components/page-header'
import { Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import { getSkillGraph } from '@/lib/api'
import { buildObjectiveDetail } from '@/lib/objective-details'
import { getSkillDetailByName } from '@/lib/skill-details'
import type { SkillGraphNode } from '@/lib/types'
import { SkillLink } from './skill-link'
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

  const nodeByCode = useMemo(() => new Map(nodes.map((item) => [item.code, item])), [nodes])
  const node = nodeByCode.get(code)
  const skillDetail = node?.skill_name ? getSkillDetailByName(node.skill_name) : undefined
  const objectiveDetail = node ? buildObjectiveDetail(node, skillDetail) : null

  if (loading) {
    return <LoadingState label="Loading objective guide..." />
  }

  if (error) {
    return <AppErrorState title="Objective guide unavailable" error={error} />
  }

  if (!node) {
    return <EmptyState title="Objective not found" body="This objective code is not in the current skill graph." actionHref="/tgos" actionLabel="Back to library" />
  }

  const prereqNodes = (node.prerequisites ?? []).map((value) => nodeByCode.get(value)).filter((value): value is SkillGraphNode => Boolean(value))
  const unlockNodes = (node.unlocks ?? []).map((value) => nodeByCode.get(value)).filter((value): value is SkillGraphNode => Boolean(value))

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Objective guide"
        title={node.title}
        intro={node.description}
        actions={<Badge color="zinc">{node.code}</Badge>}
      />

      <WorkspaceCard>
        <div className="grid gap-4 md:grid-cols-3">
          <div>
            <Subheading>Track</Subheading>
            <Text className="mt-2 text-sm">{node.source_tree_title}</Text>
          </div>
          <div>
            <Subheading>Stage</Subheading>
            <Text className="mt-2 text-sm">{node.stage.replace(/-/g, ' ')}</Text>
          </div>
          <div>
            <Subheading>Skill family</Subheading>
            <Text className="mt-2 text-sm">
              {node.skill_name ? <SkillLink skill={node.skill_name}>{node.skill_name}</SkillLink> : 'Not set'}
            </Text>
          </div>
        </div>
        {node.mastery_hint ? (
          <div className="mt-4 rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-400/30 dark:bg-emerald-500/10">
            <div className="text-sm font-semibold text-emerald-900 dark:text-emerald-200">Mastery marker</div>
            <Text className="mt-2 text-sm text-emerald-900 dark:text-emerald-100">{node.mastery_hint}</Text>
          </div>
        ) : null}
      </WorkspaceCard>

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

          <Subheading className="mt-6">Next revision moves</Subheading>
          <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            {objectiveDetail.revisionMoves.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>

          <Subheading className="mt-6">Deterministic assessment focus</Subheading>
          <ul className="mt-3 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            {objectiveDetail.assessmentFocus.map((item) => (
              <li key={item}>• {item}</li>
            ))}
          </ul>
        </WorkspaceCard>
      ) : null}

      <WorkspaceCard>
        <Subheading>Dependencies</Subheading>
        <div className="mt-3 flex flex-wrap gap-2">
          {prereqNodes.length === 0 ? <Text className="text-sm">Seed objective (no prerequisites).</Text> : null}
          {prereqNodes.map((item) => (
            <Link
              key={item.code}
              href={`/tgos/${encodeURIComponent(item.code)}`}
              className="rounded-full border border-stone-200 bg-stone-50 px-3 py-1 text-xs text-zinc-900 data-hover:bg-white dark:border-white/10 dark:bg-white/5 dark:text-zinc-100"
            >
              {item.title}
            </Link>
          ))}
        </div>

        <Subheading className="mt-6">Unlocks</Subheading>
        <div className="mt-3 flex flex-wrap gap-2">
          {unlockNodes.length === 0 ? <Text className="text-sm">Terminal objective (does not unlock another objective).</Text> : null}
          {unlockNodes.map((item) => (
            <Link
              key={item.code}
              href={`/tgos/${encodeURIComponent(item.code)}`}
              className="rounded-full border border-amber-300/40 bg-amber-100/50 px-3 py-1 text-xs text-amber-900 data-hover:bg-amber-50 dark:border-amber-300/30 dark:bg-amber-500/10 dark:text-amber-100"
            >
              {item.title}
            </Link>
          ))}
        </div>
      </WorkspaceCard>

      {skillDetail ? (
        <WorkspaceCard>
          <Subheading>Skill family crosswalk</Subheading>
          <Text className="mt-3">{skillDetail.whatItMeans}</Text>
          <div className="mt-4 grid gap-4 lg:grid-cols-2">
            <div className="rounded-xl border border-emerald-200 bg-emerald-50 p-4 dark:border-emerald-400/30 dark:bg-emerald-500/10">
              <div className="text-sm font-semibold text-emerald-900 dark:text-emerald-200">Stronger example</div>
              <Text className="mt-2 text-sm text-emerald-900 dark:text-emerald-100">{skillDetail.strongExample}</Text>
            </div>
            <div className="rounded-xl border border-amber-200 bg-amber-50 p-4 dark:border-amber-400/30 dark:bg-amber-500/10">
              <div className="text-sm font-semibold text-amber-900 dark:text-amber-200">Needs work example</div>
              <Text className="mt-2 text-sm text-amber-900 dark:text-amber-100">{skillDetail.weakExample}</Text>
            </div>
          </div>
          <ul className="mt-4 space-y-2 text-sm text-zinc-700 dark:text-zinc-300">
            {skillDetail.revisionMoves.map((move) => (
              <li key={move}>• {move}</li>
            ))}
          </ul>
        </WorkspaceCard>
      ) : null}
    </div>
  )
}
