'use client'

import { useEffect, useMemo, useState } from 'react'
import { Link } from '@/components/link'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { WorkspaceCard } from '@/components/workspace-card'
import { Input } from '@/components/input'
import { getSkillGraph } from '@/lib/api'
import type { SkillGraphNode } from '@/lib/types'
import { LoadingState, AppErrorState } from './status-state'

export function TGOLibraryView() {
  const [nodes, setNodes] = useState<SkillGraphNode[]>([])
  const [query, setQuery] = useState('')
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
        setError(err instanceof Error ? err.message : 'Could not load objective guides.')
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

  const normalized = query.trim().toLowerCase()
  const uniqueObjectives = useMemo(() => {
    const byCode = new Map<string, SkillGraphNode>()
    for (const node of nodes) {
      if (!byCode.has(node.code)) {
        byCode.set(node.code, node)
      }
    }
    return Array.from(byCode.values())
  }, [nodes])

  const filtered = useMemo(() => {
    const list = [...uniqueObjectives].sort((a, b) => {
      const familyCompare = (a.skill_name ?? '').localeCompare(b.skill_name ?? '')
      if (familyCompare !== 0) {
        return familyCompare
      }
      if (a.stage_order !== b.stage_order) {
        return a.stage_order - b.stage_order
      }
      return a.title.localeCompare(b.title)
    })
    if (!normalized) {
      return list
    }
    return list.filter((item) => {
      const haystack = `${item.title} ${item.code} ${item.description} ${item.skill_name ?? ''} ${item.stage}`.toLowerCase()
      return haystack.includes(normalized)
    })
  }, [uniqueObjectives, normalized])

  const limited = filtered.slice(0, 250)

  if (loading) {
    return <LoadingState label="Loading objective guides..." />
  }

  if (error) {
    return <AppErrorState title="Objective guides unavailable" error={error} />
  }

  return (
    <div className="space-y-8">
      <PageHeader
        eyebrow="Objective guides"
        title="Skill objective library"
        intro="Each skill objective has one shared guide page. Search by title, code, or skill family."
      />

      <WorkspaceCard>
        <div className="grid gap-4 lg:grid-cols-[2fr_1fr] lg:items-end">
          <div>
            <label htmlFor="tgo-search" className="text-sm font-semibold text-zinc-900 dark:text-white">
              Search objectives
            </label>
            <div className="mt-2">
              <Input
                id="tgo-search"
                type="search"
                value={query}
                onChange={(event) => setQuery(event.target.value)}
                placeholder="Try: scene, claim, memoir, structure..."
              />
            </div>
          </div>
          <div className="text-sm text-zinc-600 dark:text-zinc-300">
            <div>Total objectives: {uniqueObjectives.length}</div>
            <div>Matches: {filtered.length}</div>
            {filtered.length > limited.length ? <div>Showing first {limited.length} matches. Refine search to narrow further.</div> : null}
          </div>
        </div>
      </WorkspaceCard>

      <WorkspaceCard>
        <div className="space-y-3">
          {limited.length === 0 ? <Text>No matching objectives found.</Text> : null}
          {limited.map((node) => (
            <Link
              key={node.code}
              href={`/tgos/${encodeURIComponent(node.code)}`}
              className="block rounded-xl border border-stone-200 bg-stone-50 p-4 data-hover:border-stone-300 data-hover:bg-white dark:border-white/10 dark:bg-white/5 dark:data-hover:border-white/20"
            >
              <div className="flex flex-wrap items-center justify-between gap-3">
                <div className="font-semibold text-zinc-900 dark:text-white">
                  {node.title}
                  <span className="ml-2 text-xs font-normal text-zinc-500 dark:text-zinc-400">({node.code})</span>
                </div>
                <div className="text-xs text-zinc-500 dark:text-zinc-400">
                  {node.skill_name ? `${node.skill_name} • ` : ''}
                  {node.stage.replace(/-/g, ' ')}
                </div>
              </div>
              <Text className="mt-2 text-sm">{node.description}</Text>
            </Link>
          ))}
        </div>
      </WorkspaceCard>
    </div>
  )
}
