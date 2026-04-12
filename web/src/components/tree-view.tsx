'use client'

import { ArrowsPointingInIcon, ArrowsPointingOutIcon } from '@heroicons/react/20/solid'
import { InformationCircleIcon } from '@heroicons/react/16/solid'
import { useTranslations } from 'next-intl'
import { Badge } from '@/components/badge'
import { SkillTreeEdge, type EdgeBridge, type SkillTreeEdgeData } from '@/components/tree-edge'
import { Eyebrow } from '@/components/eyebrow'
import { Subheading } from '@/components/heading'
import { Link } from '@/components/link'
import { PageHeader } from '@/components/page-header'
import { Text } from '@/components/text'
import { layoutTreeGraph, type LayoutEdgeRoute, type LayoutPoint } from '@/components/tree-layout'
import { objectiveConceptKey } from '@/lib/objective-concepts'
import type { Dashboard, Tree } from '@/lib/types'
import { useTrackDashboardData } from '@/lib/use-track-dashboard-data'
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
import { useEffect, useId, useRef, useState } from 'react'
import { AppErrorState, EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

type TreeNodeStatus = 'active' | 'completed' | 'unlocked' | 'locked'
type ObjectiveRef = {
  code: string
  title: string
}

type GraphNodeData = {
  code: string
  title: string
  description: string
  stage: string
  stage_order: number
  prerequisites: ObjectiveRef[]
  unlocks: ObjectiveRef[]
  mastery_hint?: string
  status: TreeNodeStatus
  selected: boolean
  tooltipSide: 'left' | 'right'
}

const nodeTypes = {
  skillNode: SkillTreeNode,
}

const edgeTypes = {
  skillTreeEdge: SkillTreeEdge,
}

const nodeWidth = 184
const nodeHeight = 76
const tooltipFlipThreshold = 0.72

type TreeGraph = {
  nodes: Node<GraphNodeData>[]
  edges: Edge<SkillTreeEdgeData>[]
  dataByCode: Map<string, GraphNodeData>
}

function statusLabel(status: TreeNodeStatus) {
  switch (status) {
    case 'active':
      return 'active'
    case 'completed':
      return 'mastered'
    case 'unlocked':
      return 'unlocked'
    default:
      return 'locked'
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
          'border-cyan-400/60 bg-cyan-500/12 shadow-[0_0_0_1px_rgba(34,211,238,0.18),0_0_36px_rgba(34,211,238,0.18)]',
        badge: 'bg-cyan-400/18 text-cyan-100 ring-cyan-300/30',
        accent: '#22d3ee',
      }
    case 'completed':
      return {
        shell:
          'border-emerald-400/55 bg-emerald-500/12 shadow-[0_0_0_1px_rgba(52,211,153,0.18),0_0_34px_rgba(52,211,153,0.16)]',
        badge: 'bg-emerald-400/18 text-emerald-100 ring-emerald-300/30',
        accent: '#34d399',
      }
    case 'unlocked':
      return {
        shell:
          'border-amber-300/50 bg-amber-500/10 shadow-[0_0_0_1px_rgba(251,191,36,0.14),0_0_32px_rgba(245,158,11,0.14)]',
        badge: 'bg-amber-300/18 text-amber-100 ring-amber-200/30',
        accent: '#f59e0b',
      }
    default:
      return {
        shell: 'border-white/10 bg-zinc-950/78 opacity-80',
        badge: 'bg-white/8 text-zinc-300 ring-white/10',
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
  return '#71717a'
}

function SkillTreeNode({ data }: NodeProps<Node<GraphNodeData>>) {
  const tone = statusTone(data.status)
  return (
    <div
      className={`group relative w-[184px] overflow-visible rounded-[24px] border px-4 py-3 text-left transition-all duration-300 ${tone.shell} ${
        data.selected ? 'ring-2 ring-white/60 ring-offset-2 ring-offset-zinc-950' : ''
      }`}
    >
      <Handle type="target" position={Position.Left} className="!h-3 !w-3 !border-none !bg-transparent" />
      <Handle type="source" position={Position.Right} className="!h-3 !w-3 !border-none !bg-transparent" />
      <div className="flex min-h-[52px] items-center justify-between gap-3">
        <div className="min-w-0">
          <div className="pr-2 text-[13px] leading-5 font-semibold text-white">{data.title}</div>
        </div>
        <div className="flex shrink-0 items-center gap-2">
          <div
            className="h-2.5 w-2.5 rounded-full ring-1 ring-white/20"
            style={{ backgroundColor: tone.accent }}
            aria-hidden="true"
          />
        </div>
      </div>
      <div
        className={`pointer-events-none absolute top-1/2 z-30 hidden w-72 -translate-y-1/2 rounded-3xl border border-white/12 bg-zinc-950/96 p-4 text-white shadow-[0_24px_80px_rgba(0,0,0,0.45)] group-hover:block group-focus-within:block ${
          data.tooltipSide === 'left' ? 'right-[calc(100%+1rem)]' : 'left-[calc(100%+1rem)]'
        }`}
      >
        <Eyebrow tone="white" className="text-[10px] tracking-[0.24em]">
          {stageTitle(data.stage)}
        </Eyebrow>
        <div className="mt-2 flex items-start justify-between gap-3">
          <div className="min-w-0 text-sm leading-5 font-semibold text-white">{data.title}</div>
          <div
            className={`shrink-0 rounded-full px-2 py-1 text-[9px] font-semibold tracking-[0.16em] uppercase ring-1 ${tone.badge}`}
          >
            {statusLabel(data.status)}
          </div>
        </div>
        <Text className="mt-3 text-[13px] leading-5 text-zinc-300">{data.description}</Text>
        <div className="mt-4 flex items-center justify-between gap-4 text-[10px] tracking-[0.16em] text-zinc-400 uppercase">
          <span>{data.prerequisites.length} prereq</span>
          <span>{data.unlocks.length} unlocks</span>
        </div>
      </div>
    </div>
  )
}

async function buildGraph(
  tree: Tree,
  dashboard: Dashboard,
  selectedCode: string | null
): Promise<TreeGraph> {
  const treeTGOs = tree.tgos ?? []
  const active = new Set((dashboard.active_tgos ?? []).map((tgo) => tgo.code))
  const completed = new Set((dashboard.completed_tgos ?? []).map((tgo) => tgo.code))
  const unlocked = new Set((dashboard.upcoming_tgos ?? []).map((tgo) => tgo.code))
  const titleByCode = new Map(treeTGOs.map((tgo) => [tgo.code, tgo.title]))

  const unlocks = new Map<string, ObjectiveRef[]>()
  for (const tgo of treeTGOs) {
    for (const prerequisite of tgo.prerequisites ?? []) {
      const next = unlocks.get(prerequisite) ?? []
      next.push({
        code: tgo.code,
        title: titleByCode.get(tgo.code) ?? tgo.code,
      })
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
    const data: GraphNodeData = {
      code: tgo.code,
      title: tgo.title,
      description: tgo.description,
      stage: tgo.stage,
      stage_order: tgo.stage_order,
      prerequisites: (tgo.prerequisites ?? []).map((code) => ({
        code,
        title: titleByCode.get(code) ?? code,
      })),
      unlocks: unlocks.get(tgo.code) ?? [],
      mastery_hint: tgo.mastery_hint,
      status,
      selected: selectedCode === tgo.code,
      tooltipSide: 'right',
    }
    dataByCode.set(tgo.code, data)
  }

  const layoutNodes = treeTGOs.map((tgo) => ({
    id: tgo.code,
    width: nodeWidth,
    height: nodeHeight,
  }))
  const layoutEdges: { source: string; target: string }[] = []
  for (const tgo of treeTGOs) {
    for (const prerequisite of tgo.prerequisites ?? []) {
      layoutEdges.push({
        source: prerequisite,
        target: tgo.code,
      })
    }
  }
  const layout = await layoutTreeGraph({
    strategy: 'elk',
    nodes: layoutNodes,
    edges: layoutEdges,
  })
  const positionByCode = new Map(layout.nodes.map((node) => [node.id, node]))
  const routeById = new Map(layout.edges.map((edge) => [edge.id, edge]))
  const bridgesByEdge = computeEdgeBridges(layout.edges, dataByCode)
  const maxNodeX = Math.max(...layout.nodes.map((node) => node.x), 0)

  for (const [code, data] of dataByCode.entries()) {
    const positioned = positionByCode.get(code)
    if (!positioned) {
      continue
    }
    data.tooltipSide = maxNodeX === 0 || positioned.x / maxNodeX < tooltipFlipThreshold ? 'right' : 'left'
  }

  const nodes: Node<GraphNodeData>[] = treeTGOs.map((tgo) => {
    const positioned = positionByCode.get(tgo.code) ?? { x: 0, y: 0 }
    return {
      id: tgo.code,
      type: 'skillNode',
      draggable: false,
      selectable: false,
      connectable: false,
      position: {
        x: positioned.x,
        y: positioned.y,
      },
      data: dataByCode.get(tgo.code)!,
    }
  })

  const edges: Edge<SkillTreeEdgeData>[] = []
  for (const tgo of treeTGOs) {
    for (const prerequisite of tgo.prerequisites ?? []) {
      const source = dataByCode.get(prerequisite)
      const target = dataByCode.get(tgo.code)
      if (!source || !target) {
        continue
      }
      const edgeId = `${prerequisite}->${tgo.code}`
      const stroke = edgeColor(source.status, target.status)
      edges.push({
        id: edgeId,
        source: prerequisite,
        target: tgo.code,
        type: 'skillTreeEdge',
        animated: target.status === 'active' || target.status === 'unlocked',
        data: {
          points: routeById.get(edgeId)?.points ?? [],
          bridges: bridgesByEdge.get(edgeId) ?? [],
        } satisfies SkillTreeEdgeData,
        style: {
          stroke,
          strokeWidth: target.status === 'locked' ? 1.8 : 2.4,
          opacity: target.status === 'locked' ? 0.58 : 0.92,
        },
      })
    }
  }

  return { nodes, edges, dataByCode }
}

function computeEdgeBridges(routes: LayoutEdgeRoute[], dataByCode: Map<string, GraphNodeData>) {
  const bridgesByEdge = new Map<string, EdgeBridge[]>()
  const segments = routes.flatMap((route) => getOrthogonalSegments(route))

  for (let i = 0; i < segments.length; i++) {
    for (let j = i + 1; j < segments.length; j++) {
      const first = segments[i]
      const second = segments[j]

      if (first.edgeId === second.edgeId) {
        continue
      }
      if (sharesEndpoint(first, second)) {
        continue
      }
      if (first.orientation === second.orientation) {
        continue
      }

      const horizontal = first.orientation === 'horizontal' ? first : second
      const vertical = first.orientation === 'vertical' ? first : second
      const crossing = orthogonalCrossing(horizontal.start, horizontal.end, vertical.start, vertical.end)
      if (!crossing) {
        continue
      }

      const top = chooseBridgeOwner(horizontal.edgeId, vertical.edgeId, dataByCode)
      if (top !== horizontal.edgeId) {
        continue
      }

      const existing = bridgesByEdge.get(horizontal.edgeId) ?? []
      existing.push({
        x: crossing.x,
        y: crossing.y,
        segmentIndex: horizontal.segmentIndex,
      })
      bridgesByEdge.set(horizontal.edgeId, existing)
    }
  }

  return bridgesByEdge
}

function getOrthogonalSegments(route: LayoutEdgeRoute) {
  const segments: Array<{
    edgeId: string
    source: string
    target: string
    start: LayoutPoint
    end: LayoutPoint
    orientation: 'horizontal' | 'vertical'
    segmentIndex: number
  }> = []

  for (let i = 0; i < route.points.length - 1; i++) {
    const start = route.points[i]
    const end = route.points[i + 1]
    if (nearlyEqual(start.x, end.x) && nearlyEqual(start.y, end.y)) {
      continue
    }
    if (nearlyEqual(start.y, end.y)) {
      segments.push({
        edgeId: route.id,
        source: route.source,
        target: route.target,
        start,
        end,
        orientation: 'horizontal',
        segmentIndex: i,
      })
      continue
    }
    if (nearlyEqual(start.x, end.x)) {
      segments.push({
        edgeId: route.id,
        source: route.source,
        target: route.target,
        start,
        end,
        orientation: 'vertical',
        segmentIndex: i,
      })
    }
  }

  return segments
}

function sharesEndpoint(
  first: { source: string; target: string },
  second: { source: string; target: string }
) {
  return (
    first.source === second.source ||
    first.source === second.target ||
    first.target === second.source ||
    first.target === second.target
  )
}

function orthogonalCrossing(
  horizontalStart: LayoutPoint,
  horizontalEnd: LayoutPoint,
  verticalStart: LayoutPoint,
  verticalEnd: LayoutPoint
) {
  const minX = Math.min(horizontalStart.x, horizontalEnd.x)
  const maxX = Math.max(horizontalStart.x, horizontalEnd.x)
  const minY = Math.min(verticalStart.y, verticalEnd.y)
  const maxY = Math.max(verticalStart.y, verticalEnd.y)
  const x = verticalStart.x
  const y = horizontalStart.y

  if (x <= minX + 8 || x >= maxX - 8) {
    return null
  }
  if (y <= minY + 8 || y >= maxY - 8) {
    return null
  }

  return { x, y }
}

function chooseBridgeOwner(firstEdgeId: string, secondEdgeId: string, dataByCode: Map<string, GraphNodeData>) {
  const firstWeight = edgePriority(firstEdgeId, dataByCode)
  const secondWeight = edgePriority(secondEdgeId, dataByCode)
  if (firstWeight !== secondWeight) {
    return firstWeight > secondWeight ? firstEdgeId : secondEdgeId
  }
  return firstEdgeId < secondEdgeId ? firstEdgeId : secondEdgeId
}

function edgePriority(edgeId: string, dataByCode: Map<string, GraphNodeData>) {
  const [sourceId, targetId] = edgeId.split('->')
  const source = dataByCode.get(sourceId)
  const target = dataByCode.get(targetId)

  return statusWeight(source?.status) + statusWeight(target?.status) + (target?.selected ? 5 : 0)
}

function statusWeight(status?: TreeNodeStatus) {
  switch (status) {
    case 'active':
      return 6
    case 'completed':
      return 4
    case 'unlocked':
      return 3
    case 'locked':
      return 1
    default:
      return 0
  }
}

function nearlyEqual(a: number, b: number) {
  return Math.abs(a - b) < 0.5
}

export function TreeView() {
  const t = useTranslations('treeView')
  const { sessionLoading, sessionError, loading, error, tree, dashboard } = useTrackDashboardData('/tree', {
    requireActiveTree: true,
    loadErrorMessage: t('loadError'),
  })
  const [selectedCode, setSelectedCode] = useState<string | null>(null)
  const [graph, setGraph] = useState<TreeGraph | null>(null)
  const [isFullscreen, setIsFullscreen] = useState(false)
  const [fullscreenAnnouncement, setFullscreenAnnouncement] = useState('')
  const flowPaneRef = useRef<HTMLDivElement | null>(null)
  const flowPaneId = useId()
  const effectiveSelectedCode = selectedCode ?? dashboard?.active_tgos?.[0]?.code ?? tree?.tgos?.[0]?.code ?? null

  useEffect(() => {
    let cancelled = false

    async function loadGraph() {
      if (!tree || !dashboard) {
        setGraph(null)
        return
      }

      const nextGraph = await buildGraph(tree, dashboard, effectiveSelectedCode)
      if (!cancelled) {
        setGraph(nextGraph)
      }
    }

    void loadGraph()

    return () => {
      cancelled = true
    }
  }, [dashboard, effectiveSelectedCode, tree])

  useEffect(() => {
    function syncFullscreenState() {
      const nextIsFullscreen = document.fullscreenElement === flowPaneRef.current
      setIsFullscreen(nextIsFullscreen)
      setFullscreenAnnouncement(
        nextIsFullscreen ? t('fullscreenEnteredAnnouncement') : t('fullscreenExitedAnnouncement'),
      )
    }

    document.addEventListener('fullscreenchange', syncFullscreenState)
    return () => {
      document.removeEventListener('fullscreenchange', syncFullscreenState)
    }
  }, [t])

  if (sessionLoading || loading || (tree && dashboard && !graph)) {
    return <LoadingState label={t('loading')} />
  }
  if (sessionError || error || !tree || !dashboard || !graph) {
    return <AppErrorState title={t('unavailableTitle')} error={sessionError ?? error ?? t('unavailableBody')} />
  }

  const activeCount = dashboard.active_tgos?.length ?? 0
  const completedCount = dashboard.completed_tgos?.length ?? 0
  const unlockedCount = dashboard.upcoming_tgos?.length ?? 0
  const selected = effectiveSelectedCode ? (graph.dataByCode.get(effectiveSelectedCode) ?? null) : null

  async function toggleFullscreen() {
    if (!flowPaneRef.current) {
      return
    }

    if (document.fullscreenElement === flowPaneRef.current) {
      await document.exitFullscreen()
      return
    }

    await flowPaneRef.current.requestFullscreen()
  }

  return (
    <ReactFlowProvider>
      <div className="space-y-8">
        <PageHeader
          eyebrow={t('eyebrow')}
          title={t('title')}
          intro={t('intro')}
          actions={
            <>
              <Badge color="blue">{t('activeBadge', { count: activeCount })}</Badge>
              <Badge color="green">{t('masteredBadge', { count: completedCount })}</Badge>
              <Badge color="amber">{t('unlockedBadge', { count: unlockedCount })}</Badge>
            </>
          }
        />

        <WorkspaceCard className="overflow-hidden border border-white/10 bg-[radial-gradient(circle_at_top_left,_rgba(34,211,238,0.16),_transparent_30%),radial-gradient(circle_at_bottom_right,_rgba(245,158,11,0.12),_transparent_28%),linear-gradient(180deg,rgba(24,24,27,0.98),rgba(9,9,11,0.98))] p-0 text-white">
          <div className="grid min-h-[46rem] gap-0 xl:grid-cols-[minmax(0,1fr)_22rem]">
            <div
              id={flowPaneId}
              ref={flowPaneRef}
              role="region"
              aria-label={t('flowRegionLabel')}
              className="relative min-h-[38rem] bg-zinc-950"
            >
              <ReactFlow
                nodes={graph.nodes}
                edges={graph.edges}
                nodeTypes={nodeTypes}
                edgeTypes={edgeTypes}
                fitView
                fitViewOptions={{ padding: 0.18 }}
                nodesDraggable={false}
                nodesConnectable={false}
                elementsSelectable={false}
                minZoom={0.4}
                maxZoom={1.6}
                colorMode="dark"
                onNodeClick={(_, node) => setSelectedCode(node.id)}
                proOptions={{ hideAttribution: true }}
                className="bg-transparent"
              >
                <Background gap={24} size={1} color="rgba(255,255,255,0.06)" />
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
              <button
                type="button"
                onClick={() => void toggleFullscreen()}
                className="absolute bottom-3 left-[3.25rem] z-20 inline-flex h-8 w-8 items-center justify-center rounded-md border border-white/10 bg-zinc-950/92 text-zinc-100 shadow-sm transition hover:bg-zinc-900 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-cyan-300 focus-visible:ring-offset-2 focus-visible:ring-offset-zinc-950"
                aria-label={isFullscreen ? t('exitFullscreenLabel') : t('enterFullscreenLabel')}
                aria-controls={flowPaneId}
                aria-pressed={isFullscreen}
                title={isFullscreen ? t('exitFullscreenTitle') : t('enterFullscreenTitle')}
              >
                <span className="sr-only">{isFullscreen ? t('exitFullscreenLabel') : t('enterFullscreenLabel')}</span>
                {isFullscreen ? <ArrowsPointingInIcon className="h-4 w-4" /> : <ArrowsPointingOutIcon className="h-4 w-4" />}
              </button>
              <div className="sr-only" aria-live="polite" aria-atomic="true">
                {fullscreenAnnouncement}
              </div>
            </div>

            <aside className="border-t border-white/10 bg-black/24 xl:border-t-0 xl:border-l">
              <div className="p-5">
                <Eyebrow tone="white" className="tracking-[0.22em]">
                  {t('selectedSkill')}
                </Eyebrow>
                {selected ? (
                  <div className="mt-4 space-y-5">
                    <div>
                      <div className="flex items-start justify-between gap-3">
                        <div className="flex items-center gap-2">
                          <Subheading className="text-white">{selected.title}</Subheading>
                          <Link
                            href={`/skills/${encodeURIComponent(objectiveConceptKey(selected.title))}`}
                            aria-label={`Open ${selected.title} details`}
                            className="inline-flex items-center justify-center rounded-full border border-white/20 bg-white/10 p-0.5 text-zinc-200 data-hover:text-white"
                          >
                            <InformationCircleIcon className="size-4" aria-hidden="true" />
                          </Link>
                        </div>
                        <div className="rounded-full border border-white/12 bg-white/8 px-3 py-1 text-[11px] font-semibold tracking-[0.16em] text-zinc-200 uppercase">
                          {t(`status.${statusLabel(selected.status)}`)}
                        </div>
                      </div>
                      <Text className="mt-3 text-sm text-zinc-300">{selected.description}</Text>
                    </div>

                    <div className="grid gap-3 sm:grid-cols-2 xl:grid-cols-1">
                      <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                        <Eyebrow tone="white" className="text-[11px] tracking-[0.16em]">
                          {t('stage')}
                        </Eyebrow>
                        <Text className="mt-2 text-sm text-zinc-100">{stageTitle(selected.stage)}</Text>
                      </div>
                      <div className="rounded-2xl border border-white/10 bg-white/5 p-4">
                        <Eyebrow tone="white" className="text-[11px] tracking-[0.16em]">
                          {t('unlocks')}
                        </Eyebrow>
                        <Text className="mt-2 text-sm text-zinc-100">{selected.unlocks.length}</Text>
                      </div>
                    </div>

                    <div>
                      <Eyebrow tone="white" className="text-[11px] tracking-[0.16em]">
                        {t('requires')}
                      </Eyebrow>
                      <div className="mt-3 flex flex-wrap gap-2">
                        {selected.prerequisites.length === 0 ? (
                          <div className="rounded-full border border-white/10 bg-white/6 px-3 py-1 text-xs text-zinc-300">
                            {t('seedNode')}
                          </div>
                        ) : (
                          selected.prerequisites.map((prerequisite) => (
                            <div
                              key={prerequisite.code}
                              className="inline-flex items-center gap-1 rounded-full border border-white/10 bg-white/6 px-3 py-1 text-xs text-zinc-200"
                            >
                              <span>{prerequisite.title}</span>
                              <Link
                                href={`/skills/${encodeURIComponent(objectiveConceptKey(prerequisite.title))}`}
                                aria-label={`Open ${prerequisite.title} details`}
                                className="inline-flex items-center justify-center rounded-full p-0.5 text-zinc-300 data-hover:text-white"
                              >
                                <InformationCircleIcon className="size-3.5" aria-hidden="true" />
                              </Link>
                            </div>
                          ))
                        )}
                      </div>
                    </div>

                    <div>
                      <Eyebrow tone="white" className="text-[11px] tracking-[0.16em]">
                        {t('unlockPathway')}
                      </Eyebrow>
                      <div className="mt-3 flex flex-wrap gap-2">
                        {selected.unlocks.length === 0 ? (
                          <div className="rounded-full border border-white/10 bg-white/6 px-3 py-1 text-xs text-zinc-300">
                            {t('terminalNode')}
                          </div>
                        ) : (
                          selected.unlocks.map((unlock) => (
                            <div
                              key={unlock.code}
                              className="inline-flex items-center gap-1 rounded-full border border-amber-300/20 bg-amber-400/10 px-3 py-1 text-xs text-amber-100"
                            >
                              <span>{unlock.title}</span>
                              <Link
                                href={`/skills/${encodeURIComponent(objectiveConceptKey(unlock.title))}`}
                                aria-label={`Open ${unlock.title} details`}
                                className="inline-flex items-center justify-center rounded-full p-0.5 text-amber-100/90 data-hover:text-white"
                              >
                                <InformationCircleIcon className="size-3.5" aria-hidden="true" />
                              </Link>
                            </div>
                          ))
                        )}
                      </div>
                    </div>

                    {selected.mastery_hint ? (
                      <div className="rounded-2xl border border-emerald-300/18 bg-emerald-400/10 p-4">
                        <Eyebrow tone="emerald" className="text-[11px] tracking-[0.16em] dark:text-emerald-100/80">
                          {t('masteryMarker')}
                        </Eyebrow>
                        <Text className="mt-2 text-sm text-emerald-50">{selected.mastery_hint}</Text>
                      </div>
                    ) : null}
                  </div>
                ) : (
                  <Text className="mt-3 text-sm text-zinc-400">{t('selectNode')}</Text>
                )}
              </div>
            </aside>
          </div>
        </WorkspaceCard>
      </div>
    </ReactFlowProvider>
  )
}
