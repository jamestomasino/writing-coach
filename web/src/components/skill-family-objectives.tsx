'use client'

import { useEffect, useMemo, useState } from 'react'
import { Link } from '@/components/link'
import { Text } from '@/components/text'
import { getSkillGraph } from '@/lib/api'
import { buildObjectiveConcepts } from '@/lib/objective-concepts'
import type { SkillGraphNode } from '@/lib/types'
import { AppErrorState, LoadingState } from './status-state'

export function SkillFamilyObjectives({ familyName }: { familyName: string }) {
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
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load skill objectives.')
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
  }, [])

  const objectives = useMemo(() => {
    const key = familyName.trim().toLowerCase()
    const { concepts } = buildObjectiveConcepts(nodes)
    return concepts.filter((item) => (item.skill_name ?? '').trim().toLowerCase() === key)
  }, [familyName, nodes])

  if (loading) {
    return <LoadingState label="Loading skill objectives..." />
  }

  if (error) {
    return <AppErrorState title="Skill objectives unavailable" error={error} />
  }

  if (objectives.length === 0) {
    return <Text className="mt-3 text-sm">No active objective concepts currently map to this skill family.</Text>
  }

  return (
    <ul className="mt-3 space-y-2 text-sm">
      {objectives.map((item) => (
        <li key={item.key}>
          <Link
            href={`/tgos/${encodeURIComponent(item.key)}`}
            className="inline-flex items-center gap-2 rounded-lg border border-stone-200 bg-stone-50 px-3 py-1.5 text-zinc-900 data-hover:bg-white dark:border-white/10 dark:bg-white/5 dark:text-zinc-100"
          >
            <span className="font-medium">{item.title}</span>
            <span className="text-xs text-zinc-500 dark:text-zinc-400">({item.key})</span>
          </Link>
        </li>
      ))}
    </ul>
  )
}
