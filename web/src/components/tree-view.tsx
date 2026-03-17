'use client'

import { useEffect, useMemo, useState } from 'react'
import {
  Background,
  Controls,
  Handle,
  MiniMap,
  Position,
  ReactFlow,
  ReactFlowProvider,
  type Edge,
  type Node,
  type NodeProps,
} from '@xyflow/react'
import { Badge } from '@/components/badge'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { getDashboard, getOnboarding, getSession, getTree } from '@/lib/api'
import type { Dashboard, OnboardingState, Tree } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

type TreeNodeStatus = 'active' | 'completed' | 'unlocked' | 'locked'

type HexCoord = {
  q: number
  r: number
}

type GraphNodeData = {
  code: string
  title: string
  description: string
  stage: string
  stage_order: number
  prerequisiteCodes: string[]
  prerequisites: string[]
  unlocks: string[]
  mastery_hint?: string
  status: TreeNodeStatus
  selected: boolean
}

const nodeTypes = {
  skillNode: SkillTreeNode,
}

const nodeWidth = 156
const nodeHeight = 136
const hexSize = 116

function statusLabel(status: TreeNodeStatus) {
  switch (status) {
    case 'active':
      return 'Active'
    case 'completed':
      return 'Mastered'
    case 'unlocked':
      return 'Unlocked'
    default:
      return 'Locked'
  }
}

function stageTitle(stage: string) {
  return stage.replace(/-/g, ' ')
}

function statusTone(status: TreeNodeStatus) {
  switch (status) {
    case 'active':
      return {
        shell:
          'border-cyan-400/60 bg-cyan-500/12 shadow-[0_0_0_1px_rgba(34,211,238,0.18),0_0_34px_rgba(34,211,238,0.2)]',
        accent: '#22d3ee',
      }
    case 'completed':
      return {
        shell:
          'border-emerald-400/55 bg-emerald-500/12 shadow-[0_0_0_1px_rgba(52,211,153,0.18),0_0_32px_rgba(52,211,153,0.16)]',
        accent: '#34d399',
      }
    case 'unlocked':
      return {
        shell:
          'border-amber-300/50 bg-amber-500/10 shadow-[0_0_0_1px_rgba(251,191,36,0.14),0_0_28px_rgba(245,158,11,0.16)]',
        accent: '#f59e0b',
      }
    default:
      return {
        shell: 'border-white/10 bg-zinc-950/80 opacity-84',
        accent: '#71717a',
      }
  }
}

function edgeColor(sourceStatus: TreeNodeStatus, targetStatus: TreeNodeStatus) {
  if (sourceStatus === 'completed' && targetStatus === 'completed') {
    return '#34d399'
  }
  if (targetStatus === 'active' || sourceStatus === 'active') {
    return '#22d3ee'
  }
  if (targetStatus === 'unlocked') {
    return '#f59e0b'
  }
  return '#3f3f46'
}

function SkillTreeNode({ data }: NodeProps<Node<GraphNodeData>>) {
  const tone = statusTone(data.status)
  return (
    <div
      className={`group relative flex h-[136px] w-[156px] items-center justify-center border px-4 py-4 text-center transition-all duration-300 [clip-path:polygon(25%_6%,75%_6%,100%_50%,75%_94%,25%_94%,0_50%)] ${tone.shell} ${
        data.selected ? 'ring-2 ring-white/60 ring-offset-2 ring-offset-zinc-950' : ''
      }`}
    >
      <Handle type="target" position={Position.Left} className="!h-2 !w-2 !border-none !bg-transparent" />
      <Handle type="source" position={Position.Right} className="!h-2 !w-2 !border-none !bg-transparent" />
      <div className="pointer-events-none absolute inset-0 bg-[radial-gradient(circle_at_50%_50%,rgba(255,255,255,0.14),transparent_66%)]" />
      <div className="relative text-sm font-semibold leading-5 text-white">{data.title}</div>
      {data.selected ? (
        <div className="pointer-events-none absolute inset-[8px] border border-white/45 [clip-path:polygon(25%_6%,75%_6%,100%_50%,75%_94%,25%_94%,0_50%)]" />
      ) : null}
    </div>
  )
}

function hexDirections(): HexCoord[] {
  return [
    { q: 1, r: 0 },
    { q: 1, r: -1 },
    { q: 0, r: -1 },
    { q: -1, r: 0 },
    { q: -1, r: 1 },
    { q: 0, r: 1 },
  ]
}

function addHex(a: HexCoord, b: HexCoord): HexCoord {
  return { q: a.q + b.q, r: a.r + b.r }
}

function scaleHex(coord: HexCoord, factor: number): HexCoord {
  return { q: coord.q * factor, r: coord.r * factor }
}

function axialToPixel(coord: HexCoord) {
  return {
    x: hexSize * Math.sqrt(3) * (coord.q + coord.r / 2),
    y: hexSize * 1.5 * coord.r,
  }
}

function hexDistance(coord: HexCoord) {
  return Math.max(Math.abs(coord.q), Math.abs(coord.r), Math.abs(-coord.q - coord.r))
}

function ringCoords(radius: number) {
  if (radius === 0) {
    return [{ q: 0, r: 0 }]
  }
  const coords: HexCoord[] = []
  const directions = hexDirections()
  let cursor = scaleHex(directions[4], radius)
  for (let side = 0; side < 6; side++) {
    for (let step = 0; step < radius; step++) {
      coords.push(cursor)
      cursor = addHex(cursor, directions[side])
    }
  }
  return coords
}

function coordAngle(coord: HexCoord) {
  const pixel = axialToPixel(coord)
  return Math.atan2(pixel.y, pixel.x)
}

function averageAngles(angles: number[]) {
  if (angles.length === 0) {
    return 0
  }
  const x = angles.reduce((sum, angle) => sum + Math.cos(angle), 0)
  const y = angles.reduce((sum, angle) => sum + Math.sin(angle), 0)
  return Math.atan2(y, x)
}

function buildStageRings(count: number, nextRing: number) {
  const rings: number[] = []
  let capacity = 0
  while (capacity < count) {
    rings.push(nextRing)
    capacity += ringCoords(nextRing).length
    nextRing++
  }
  return { rings, nextRing }
}

function orderRingPositions(rings: number[]) {
  return rings
    .flatMap((radius) => ringCoords(radius))
    .sort((a, b) => {
      const distanceScore = hexDistance(a) - hexDistance(b)
      if (distanceScore !== 0) {
        return distanceScore
      }
      return coordAngle(a) - coordAngle(b)
    })
}

function buildGraph(tree: Tree, dashboard: Dashboard, selectedCode: string | null) {
  const treeTGOs = tree.tgos ?? []
  const active = new Set((dashboard.active_tgos ?? []).map((tgo) => tgo.code))
  const completed = new Set((dashboard.completed_tgos ?? []).map((tgo) => tgo.code))
  const unlocked = new Set((dashboard.upcoming_tgos ?? []).map((tgo) => tgo.code))
  const titleByCode = new Map(treeTGOs.map((tgo) => [tgo.code, tgo.title]))

  const unlocks = new Map<string, string[]>()
  for (const tgo of treeTGOs) {
    for (const prerequisite of tgo.prerequisites ?? []) {
      const next = unlocks.get(prerequisite) ?? []
      next.push(titleByCode.get(tgo.code) ?? tgo.code)
      unlocks.set(prerequisite, next)
    }
  }

  const dataByCode = new Map<string, GraphNodeData>()
  for (const tgo of treeTGOs) {
    const status: TreeNodeStatus = active.has(tgo.code)
      ? 'active'
      : completed.has(tgo.code)
        ? 'completed'
        : unlocked.has(tgo.code)
          ? 'unlocked'
          : 'locked'
    dataByCode.set(tgo.code, {
      code: tgo.code,
      title: tgo.title,
      description: tgo.description,
      stage: tgo.stage,
      stage_order: tgo.stage_order,
      prerequisiteCodes: tgo.prerequisites ?? [],
      prerequisites: (tgo.prerequisites ?? []).map((code) => titleByCode.get(code) ?? code),
      unlocks: unlocks.get(tgo.code) ?? [],
      mastery_hint: tgo.mastery_hint,
      status,
      selected: selectedCode === tgo.code,
    })
  }

  const stageMap = new Map<string, { order: number; nodes: GraphNodeData[] }>()
  for (const data of dataByCode.values()) {
    const existing = stageMap.get(data.stage)
    if (!existing) {
      stageMap.set(data.stage, { order: data.stage_order, nodes: [data] })
      continue
    }
    existing.order = Math.min(existing.order, data.stage_order)
    existing.nodes.push(data)
  }

  const orderedStages = [...stageMap.entries()]
    .sort((a, b) => a[1].order - b[1].order)
    .map(([stage, value], index) => ({ stage, order: value.order, nodes: value.nodes, index }))

  const assignedCoords = new Map<string, HexCoord>()
  let nextRing = 0
  for (const stage of orderedStages) {
    const allocation = buildStageRings(stage.nodes.length, nextRing)
    const openCoords = orderRingPositions(allocation.rings)
    const nodes = [...stage.nodes].sort((a, b) => {
      const aAngles = a.prerequisiteCodes
        .map((code) => assignedCoords.get(code))
        .filter((coord): coord is HexCoord => Boolean(coord))
        .map(coordAngle)
      const bAngles = b.prerequisiteCodes
        .map((code) => assignedCoords.get(code))
        .filter((coord): coord is HexCoord => Boolean(coord))
        .map(coordAngle)
      const aAngle = aAngles.length > 0 ? averageAngles(aAngles) : stage.index * 0.18
      const bAngle = bAngles.length > 0 ? averageAngles(bAngles) : stage.index * 0.18
      const angleScore = aAngle - bAngle
      if (Math.abs(angleScore) > 0.01) {
        return angleScore
      }
      const statusOrder: Record<TreeNodeStatus, number> = { active: 0, unlocked: 1, completed: 2, locked: 3 }
      const statusScore = statusOrder[a.status] - statusOrder[b.status]
      return statusScore !== 0 ? statusScore : a.title.localeCompare(b.title)
    })
    for (let index = 0; index < nodes.length; index++) {
      assignedCoords.set(nodes[index].code, openCoords[index])
    }
    nextRing = allocation.nextRing
  }

  const nodes: Node<GraphNodeData>[] = treeTGOs.map((tgo) => {
    const coord = assignedCoords.get(tgo.code) ?? { q: 0, r: 0 }
    const pixel = axialToPixel(coord)
    return {
      id: tgo.code,
      type: 'skillNode',
      draggable: false,
      selectable: false,
      connectable: false,
      position: {
        x: pixel.x - nodeWidth / 2,
        y: pixel.y - nodeHeight / 2,
      },
      data: dataByCode.get(tgo.code)!,
    }
  })

  const edges: Edge[] = []
  for (const tgo of treeTGOs) {
    for (const prerequisite of tgo.prerequisites ?? []) {
      const source = dataByCode.get(prerequisite)
      const target = dataByCode.get(tgo.code)
      if (!source || !target) {
        continue
      }
      const stroke = edgeColor(source.status, target.status)
      edges.push({
        id: `${prerequisite}->${tgo.code}`,
        source: prerequisite,
        target: tgo.code,
        type: 'straight',
        animated: target.status === 'active' || target.status === 'unlocked',
        style: {
          stroke,
          strokeWidth: target.status === 'locked' ? 1.3 : 2.1,
          opacity: target.status === 'locked' ? 0.26 : 0.86,
        },
      })
    }
  }

  return { nodes, edges, dataByCode }
}

export function TreeView() {
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [tree, setTree] = useState<Tree | null>(null)
  const [dashboard, setDashboard] = useState<Dashboard | null>(null)
  const [onboarding, setOnboarding] = useState<OnboardingState | null>(null)
  const [selectedCode, setSelectedCode] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false

    async function load() {
      try {
        const session = await getSession()
        if (!session.authenticated) {
          throw new Error('Sign in to inspect your skill map')
        }
        if (!session.onboarding_complete || !session.active_tree_slug) {
          if (!cancelled) {
            setOnboarding({ onboarding_complete: false })
          }
          return
        }

        const [treeData, dashboardData, onboardingData] = await Promise.all([
          getTree(session.active_tree_slug),
          getDashboard(),
          getOnboarding(),
        ])

        if (!cancelled) {
          setTree(treeData)
          setDashboard(dashboardData)
          setOnboarding(onboardingData)
          setError(null)
        }
      } catch (err) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Could not load skill map')
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

  const graph = useMemo(() => {
    if (!tree || !dashboard) {
      return null
    }
    return buildGraph(tree, dashboard, selectedCode)
  }, [dashboard, selectedCode, tree])

  useEffect(() => {
    if (!tree || !dashboard || selectedCode) {
      return
    }
    const defaultCode = dashboard.active_tgos?.[0]?.code ?? tree.tgos?.[0]?.code ?? null
    if (defaultCode) {
      setSelectedCode(defaultCode)
    }
  }, [dashboard, selectedCode, tree])

  if (loading) {
    return <LoadingState label="Loading skill map…" />
  }
  if (!onboarding?.onboarding_complete) {
    return (
      <EmptyState
        title="Build your track first"
        body="You need to complete the starter path setup before the app can show your skill map."
        actionHref="/onboarding"
        actionLabel="Open starter path"
      />
    )
  }
  if (error || !tree || !dashboard || !graph) {
    return <EmptyState title="Tree unavailable" body={error ?? 'Could not load the current tree.'} actionHref="/" actionLabel="Back to assignment" />
  }

  const activeCount = dashboard.active_tgos?.length ?? 0
  const completedCount = dashboard.completed_tgos?.length ?? 0
  const unlockedCount = dashboard.upcoming_tgos?.length ?? 0
  const selected = selectedCode ? graph.dataByCode.get(selectedCode) ?? null : null

  return (
    <ReactFlowProvider>
      <div className="space-y-8">
        <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
          <div>
            <Heading>Skill map</Heading>
            <Text className="mt-2 max-w-3xl">
              This page is the game board. The core cluster anchors the center, and each later stage expands outward as the track opens.
            </Text>
          </div>
          <div className="flex flex-wrap gap-2 text-sm text-zinc-600 dark:text-zinc-300">
            <Badge color="blue">{activeCount} active</Badge>
            <Badge color="green">{completedCount} mastered</Badge>
            <Badge color="amber">{unlockedCount} unlocked next</Badge>
          </div>
        </header>

        <WorkspaceCard className="overflow-hidden border border-white/10 bg-[radial-gradient(circle_at_50%_42%,rgba(34,211,238,0.14),transparent_18%),radial-gradient(circle_at_50%_58%,rgba(245,158,11,0.1),transparent_28%),linear-gradient(180deg,rgba(24,24,27,0.98),rgba(9,9,11,0.98))] p-0 text-white">
          <div className="grid min-h-[46rem] gap-0 xl:grid-cols-[minmax(0,1fr)_22rem]">
            <div className="relative min-h-[38rem]">
              <ReactFlow
                nodes={graph.nodes}
                edges={graph.edges}
                nodeTypes={nodeTypes}
                fitView
                fitViewOptions={{ padding: 0.16 }}
                nodesDraggable={false}
                nodesConnectable={false}
                elementsSelectable={false}
                minZoom={0.38}
                maxZoom={1.7}
                colorMode="dark"
                onNodeClick={(_, node) => setSelectedCode(node.id)}
                proOptions={{ hideAttribution: true }}
                className="bg-transparent"
              >
                <Background gap={28} size={1} color="rgba(255,255,255,0.06)" />
                <MiniMap
                  pannable
                  zoomable
                  nodeColor={(node) => statusTone((node.data as GraphNodeData).status).accent}
                  maskColor="rgba(0,0,0,0.42)"
                  style={{
                    background: 'rgba(9, 9, 11, 0.92)',
                    border: '1px solid rgba(255,255,255,0.1)',
                  }}
                />
                <Controls
                  showInteractive={false}
                  className="[&_button]:!border-white/10 [&_button]:!bg-zinc-950/92 [&_button]:!text-zinc-100 [&_button:hover]:!bg-zinc-900"
                />
              </ReactFlow>
            </div>

            <aside className="border-t border-white/10 bg-black/24 xl:border-l xl:border-t-0">
              <div className="p-5">
                <div className="text-xs font-semibold uppercase tracking-[0.22em] text-zinc-400">Selected skill</div>
                {selected ? (
                  <div className="mt-4 space-y-5">
                    <div>
                      <div className="flex items-start justify-between gap-3">
                        <Subheading className="text-white">{selected.title}</Subheading>
                        <div className="rounded-full border border-white/12 bg-white/8 px-3 py-1 text-[11px] font-semibold uppercase tracking-[0.16em] text-zinc-200">
                          {statusLabel(selected.status)}
                        </div>
                      </div>
                      <Text className="mt-3 text-sm text-zinc-300">{selected.description}</Text>
                    </div>

                    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
                      <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                        <div className="text-[11px] font-semibold uppercase tracking-[0.16em] text-zinc-400">Stage</div>
                        <Text className="mt-2 text-sm text-zinc-100">{stageTitle(selected.stage)}</Text>
                      </div>
                      <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                        <div className="text-[11px] font-semibold uppercase tracking-[0.16em] text-zinc-400">Unlocks</div>
                        <Text className="mt-2 text-sm text-zinc-100">{selected.unlocks.length}</Text>
                      </div>
                    </div>

                    <div>
                      <div className="text-[11px] font-semibold uppercase tracking-[0.16em] text-zinc-400">Requires</div>
                      <div className="mt-3 flex flex-wrap gap-2">
                        {selected.prerequisites.length === 0 ? (
                          <div className="rounded-full border border-white/10 bg-white/6 px-3 py-1 text-xs text-zinc-300">Seed node</div>
                        ) : (
                          selected.prerequisites.map((prerequisite) => (
                            <div key={prerequisite} className="rounded-full border border-white/10 bg-white/6 px-3 py-1 text-xs text-zinc-200">
                              {prerequisite}
                            </div>
                          ))
                        )}
                      </div>
                    </div>

                    <div>
                      <div className="text-[11px] font-semibold uppercase tracking-[0.16em] text-zinc-400">Unlock pathway</div>
                      <div className="mt-3 flex flex-wrap gap-2">
                        {selected.unlocks.length === 0 ? (
                          <div className="rounded-full border border-white/10 bg-white/6 px-3 py-1 text-xs text-zinc-300">Terminal node</div>
                        ) : (
                          selected.unlocks.map((unlock) => (
                            <div key={unlock} className="rounded-full border border-amber-300/20 bg-amber-400/10 px-3 py-1 text-xs text-amber-100">
                              {unlock}
                            </div>
                          ))
                        )}
                      </div>
                    </div>

                    {selected.mastery_hint ? (
                      <div className="rounded-2xl border border-emerald-300/18 bg-emerald-400/10 p-4">
                        <div className="text-[11px] font-semibold uppercase tracking-[0.16em] text-emerald-100/80">Mastery marker</div>
                        <Text className="mt-2 text-sm text-emerald-50">{selected.mastery_hint}</Text>
                      </div>
                    ) : null}
                  </div>
                ) : (
                  <Text className="mt-3 text-sm text-zinc-400">Select a node to inspect its role in the tree.</Text>
                )}
              </div>
            </aside>
          </div>
        </WorkspaceCard>
      </div>
    </ReactFlowProvider>
  )
}
