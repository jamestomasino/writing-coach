'use client'

import { useEffect, useMemo, useState } from 'react'
import { Badge } from '@/components/badge'
import { Heading, Subheading } from '@/components/heading'
import { Text } from '@/components/text'
import { getDashboard, getOnboarding, getSession, getTree } from '@/lib/api'
import type { Dashboard, OnboardingState, Tree } from '@/lib/types'
import { EmptyState, LoadingState } from './status-state'
import { WorkspaceCard } from './workspace-card'

type TreeNodeStatus = 'active' | 'completed' | 'unlocked' | 'locked'

type SkillNode = {
  code: string
  title: string
  description: string
  stage: string
  stage_order: number
  prerequisites: string[]
  prerequisiteCodes: string[]
  unlocks: string[]
  mastery_hint?: string
  status: TreeNodeStatus
}

type StageRegion = {
  stage: string
  order: number
  title: string
  nodes: SkillNode[]
  activeCount: number
  unlockedCount: number
  completedCount: number
  x: number
  y: number
  links: string[]
}

function stageTitle(stage: string) {
  return stage.replace(/-/g, ' ')
}

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

function statusTone(status: TreeNodeStatus) {
  switch (status) {
    case 'active':
      return 'border-cyan-400/60 bg-cyan-500/12 text-cyan-50 shadow-[0_0_0_1px_rgba(34,211,238,0.14),0_0_24px_rgba(34,211,238,0.18)]'
    case 'completed':
      return 'border-emerald-400/55 bg-emerald-500/12 text-emerald-50 shadow-[0_0_0_1px_rgba(52,211,153,0.14),0_0_22px_rgba(52,211,153,0.14)]'
    case 'unlocked':
      return 'border-amber-300/55 bg-amber-500/10 text-amber-50 shadow-[0_0_0_1px_rgba(251,191,36,0.12),0_0_20px_rgba(245,158,11,0.14)]'
    default:
      return 'border-white/10 bg-zinc-950/72 text-zinc-200 opacity-78'
  }
}

function stageTone(region: StageRegion) {
  if (region.activeCount > 0) {
    return 'border-cyan-400/55 bg-cyan-500/12 text-cyan-50 shadow-[0_0_0_1px_rgba(34,211,238,0.14),0_0_30px_rgba(34,211,238,0.16)]'
  }
  if (region.unlockedCount > 0) {
    return 'border-amber-300/50 bg-amber-500/10 text-amber-50 shadow-[0_0_0_1px_rgba(251,191,36,0.12),0_0_24px_rgba(245,158,11,0.14)]'
  }
  if (region.completedCount === region.nodes.length && region.nodes.length > 0) {
    return 'border-emerald-400/55 bg-emerald-500/12 text-emerald-50 shadow-[0_0_0_1px_rgba(52,211,153,0.14),0_0_24px_rgba(52,211,153,0.14)]'
  }
  return 'border-white/10 bg-zinc-950/78 text-zinc-100'
}

function buildTreeModel(tree: Tree, dashboard: Dashboard) {
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

  const nodes: SkillNode[] = treeTGOs.map((tgo) => ({
    code: tgo.code,
    title: tgo.title,
    description: tgo.description,
    stage: tgo.stage,
    stage_order: tgo.stage_order,
    prerequisites: (tgo.prerequisites ?? []).map((code) => titleByCode.get(code) ?? code),
    prerequisiteCodes: tgo.prerequisites ?? [],
    unlocks: unlocks.get(tgo.code) ?? [],
    mastery_hint: tgo.mastery_hint,
    status: active.has(tgo.code) ? 'active' : completed.has(tgo.code) ? 'completed' : unlocked.has(tgo.code) ? 'unlocked' : 'locked',
  }))

  const stageMap = new Map<string, StageRegion>()
  for (const node of nodes) {
    const current = stageMap.get(node.stage)
    if (!current) {
      stageMap.set(node.stage, {
        stage: node.stage,
        order: node.stage_order,
        title: stageTitle(node.stage),
        nodes: [node],
        activeCount: node.status === 'active' ? 1 : 0,
        unlockedCount: node.status === 'unlocked' ? 1 : 0,
        completedCount: node.status === 'completed' ? 1 : 0,
        x: 0,
        y: 0,
        links: [],
      })
      continue
    }
    current.order = Math.min(current.order, node.stage_order)
    current.nodes.push(node)
    if (node.status === 'active') {
      current.activeCount++
    }
    if (node.status === 'unlocked') {
      current.unlockedCount++
    }
    if (node.status === 'completed') {
      current.completedCount++
    }
  }

  const regions = [...stageMap.values()].sort((a, b) => a.order - b.order)
  const stageIndex = new Map(regions.map((region, index) => [region.stage, index]))
  const regionLinks = new Map<string, Set<string>>()
  for (const region of regions) {
    regionLinks.set(region.stage, new Set())
  }
  for (const node of nodes) {
    for (const prerequisiteCode of node.prerequisiteCodes) {
      const prerequisite = nodes.find((candidate) => candidate.code === prerequisiteCode)
      if (!prerequisite || prerequisite.stage === node.stage) {
        continue
      }
      regionLinks.get(prerequisite.stage)?.add(node.stage)
    }
  }

  const columns = 3
  const regionWidth = 240
  const regionHeight = 196
  const columnGap = 220
  const rowGap = 190
  const rowOffset = 110

  regions.forEach((region, index) => {
    const row = Math.floor(index / columns)
    const col = index % columns
    region.x = col * columnGap + (row % 2 === 1 ? rowOffset : 0)
    region.y = row * rowGap
    region.links = [...(regionLinks.get(region.stage) ?? [])]
  })

  return {
    nodes,
    regions: regions.map((region) => ({
      ...region,
      nodes: region.nodes.sort((a, b) => a.title.localeCompare(b.title)),
    })),
    stageIndex,
    regionWidth,
    regionHeight,
  }
}

function StageRegionHex({ region }: { region: StageRegion }) {
  return (
    <div
      className={`absolute flex h-[196px] w-[240px] flex-col justify-center border px-6 py-5 text-center [clip-path:polygon(25%_4%,75%_4%,100%_50%,75%_96%,25%_96%,0_50%)] ${stageTone(region)}`}
      style={{ left: region.x, top: region.y }}
    >
      <div className="text-[11px] font-semibold uppercase tracking-[0.22em] text-zinc-300/90">Region</div>
      <div className="mt-2 text-base font-semibold leading-6">{region.title}</div>
      <div className="mt-4 flex items-center justify-center gap-3 text-xs uppercase tracking-[0.16em] text-zinc-300/85">
        <span>{region.nodes.length} skills</span>
        <span>{region.activeCount} active</span>
      </div>
    </div>
  )
}

function SkillHex({
  node,
  selected,
  onSelect,
}: {
  node: SkillNode
  selected: boolean
  onSelect: (code: string) => void
}) {
  return (
    <button
      type="button"
      onClick={() => onSelect(node.code)}
      className={`relative flex h-[118px] w-[136px] items-center justify-center border px-3 text-center text-sm font-semibold leading-5 transition-all duration-200 [clip-path:polygon(25%_6%,75%_6%,100%_50%,75%_94%,25%_94%,0_50%)] ${statusTone(
        node.status,
      )} ${selected ? 'ring-2 ring-white/60 ring-offset-2 ring-offset-zinc-950' : 'hover:border-white/30'}`}
    >
      <span>{node.title}</span>
    </button>
  )
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

  const model = useMemo(() => {
    if (!tree || !dashboard) {
      return null
    }
    return buildTreeModel(tree, dashboard)
  }, [dashboard, tree])

  useEffect(() => {
    if (!model || selectedCode) {
      return
    }
    const firstActive = model.nodes.find((node) => node.status === 'active')
    setSelectedCode(firstActive?.code ?? model.nodes[0]?.code ?? null)
  }, [model, selectedCode])

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
  if (error || !tree || !dashboard || !model) {
    return <EmptyState title="Tree unavailable" body={error ?? 'Could not load the current tree.'} actionHref="/" actionLabel="Back to assignment" />
  }

  const activeCount = dashboard.active_tgos?.length ?? 0
  const completedCount = dashboard.completed_tgos?.length ?? 0
  const unlockedCount = dashboard.upcoming_tgos?.length ?? 0
  const selected = selectedCode ? model.nodes.find((node) => node.code === selectedCode) ?? null : null
  const overviewHeight = (Math.floor((model.regions.length - 1) / 3) + 1) * 190 + 220
  const overviewWidth = 760

  return (
    <div className="space-y-8">
      <header className="flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
        <div>
          <Heading>Skill map</Heading>
          <Text className="mt-2 max-w-3xl">
            The top map shows how the major regions of the track connect. The page then moves downward through each stage so you can inspect the individual skill network one layer at a time.
          </Text>
        </div>
        <div className="flex flex-wrap gap-2 text-sm text-zinc-600 dark:text-zinc-300">
          <Badge color="blue">{activeCount} active</Badge>
          <Badge color="green">{completedCount} mastered</Badge>
          <Badge color="amber">{unlockedCount} unlocked next</Badge>
        </div>
      </header>

      <WorkspaceCard className="overflow-hidden border border-white/10 bg-[radial-gradient(circle_at_top,_rgba(34,211,238,0.14),_transparent_22%),linear-gradient(180deg,rgba(24,24,27,0.98),rgba(9,9,11,0.98))] text-white">
        <div className="grid gap-8 xl:grid-cols-[minmax(0,1fr)_22rem]">
          <div className="overflow-x-auto">
            <div className="relative mx-auto min-w-[760px]" style={{ height: `${overviewHeight}px`, width: `${overviewWidth}px` }}>
              <svg className="absolute inset-0 h-full w-full" viewBox={`0 0 ${overviewWidth} ${overviewHeight}`} fill="none" xmlns="http://www.w3.org/2000/svg">
                {model.regions.flatMap((region) =>
                  region.links.map((targetStage) => {
                    const target = model.regions.find((candidate) => candidate.stage === targetStage)
                    if (!target) {
                      return null
                    }
                    const startX = region.x + model.regionWidth / 2
                    const startY = region.y + model.regionHeight / 2
                    const endX = target.x + model.regionWidth / 2
                    const endY = target.y + model.regionHeight / 2
                    const midY = (startY + endY) / 2
                    return (
                      <path
                        key={`${region.stage}-${target.stage}`}
                        d={`M ${startX} ${startY} C ${startX} ${midY}, ${endX} ${midY}, ${endX} ${endY}`}
                        stroke="rgba(255,255,255,0.14)"
                        strokeWidth="2"
                      />
                    )
                  }),
                )}
              </svg>
              {model.regions.map((region) => (
                <StageRegionHex key={region.stage} region={region} />
              ))}
            </div>
          </div>

          <div className="space-y-5">
            <div className="text-xs font-semibold uppercase tracking-[0.22em] text-zinc-400">Selected skill</div>
            {selected ? (
              <div className="space-y-5">
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

                {selected.mastery_hint ? (
                  <div className="rounded-2xl border border-emerald-300/18 bg-emerald-400/10 p-4">
                    <div className="text-[11px] font-semibold uppercase tracking-[0.16em] text-emerald-100/80">Mastery marker</div>
                    <Text className="mt-2 text-sm text-emerald-50">{selected.mastery_hint}</Text>
                  </div>
                ) : null}
              </div>
            ) : (
              <Text className="text-sm text-zinc-400">Select a skill below to inspect it.</Text>
            )}
          </div>
        </div>
      </WorkspaceCard>

      <div className="space-y-8">
        {model.regions.map((region, index) => (
          <WorkspaceCard key={region.stage}>
            <div className="flex flex-col gap-3 lg:flex-row lg:items-end lg:justify-between">
              <div>
                <div className="text-[11px] font-semibold uppercase tracking-[0.2em] text-zinc-500 dark:text-zinc-400">Stage {index + 1}</div>
                <Subheading className="mt-2">{region.title}</Subheading>
                <Text className="mt-2">
                  {region.nodes.length} skills in this region. Active, unlocked, and mastered skills stay color-coded inside the stage grid.
                </Text>
              </div>
              <div className="flex flex-wrap gap-2">
                <Badge color="blue">{region.activeCount} active</Badge>
                <Badge color="amber">{region.unlockedCount} unlocked</Badge>
                <Badge color="green">{region.completedCount} mastered</Badge>
              </div>
            </div>

            <div className="mt-6 space-y-4">
              {Array.from({ length: Math.ceil(region.nodes.length / 4) }, (_, rowIndex) => {
                const rowNodes = region.nodes.slice(rowIndex * 4, rowIndex * 4 + 4)
                return (
                  <div key={`${region.stage}-${rowIndex}`} className={`flex flex-wrap gap-3 ${rowIndex % 2 === 1 ? 'md:pl-[76px]' : ''}`}>
                    {rowNodes.map((node) => (
                      <SkillHex key={node.code} node={node} selected={selectedCode === node.code} onSelect={setSelectedCode} />
                    ))}
                  </div>
                )
              })}
            </div>
          </WorkspaceCard>
        ))}
      </div>
    </div>
  )
}
