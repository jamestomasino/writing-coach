#!/usr/bin/env node

const fs = require('fs')
const os = require('os')
const path = require('path')
const { execFileSync } = require('child_process')
const ELK = require('../web/node_modules/elkjs/lib/elk.bundled')
const { chromium } = require('../web/node_modules/playwright')

const slug = process.argv[2] || 'story-craft-track'
const outputDir = process.argv[3] || path.join(os.homedir(), 'tmp')

const nodeWidth = 184
const nodeHeight = 76
const padding = 48

main().catch((error) => {
  console.error(error)
  process.exit(1)
})

async function main() {
  fs.mkdirSync(outputDir, { recursive: true })

  const tree = exportBuiltInTree(slug, outputDir)
  const elk = new ELK()
  const unlocks = buildUnlocks(tree)
  const byCode = Object.fromEntries(
    tree.tgos.map((tgo) => [
      tgo.code,
      {
        ...tgo,
        unlocks: unlocks.get(tgo.code) ?? [],
        status: statusFor(tgo.prerequisites ?? [], unlocks.get(tgo.code) ?? []),
      },
    ])
  )

  const graph = await elk.layout({
    id: 'skill-tree',
    layoutOptions: {
      'elk.algorithm': 'layered',
      'elk.direction': 'RIGHT',
      'elk.edgeRouting': 'ORTHOGONAL',
      'elk.layered.considerModelOrder': 'NODES_AND_EDGES',
      'elk.layered.crossingMinimization.strategy': 'LAYER_SWEEP',
      'elk.layered.nodePlacement.strategy': 'BRANDES_KOEPF',
      'elk.layered.spacing.nodeNodeBetweenLayers': '128',
      'elk.padding': '[top=48,left=48,bottom=48,right=48]',
      'elk.spacing.nodeNode': '36',
    },
    children: tree.tgos.map((tgo) => ({ id: tgo.code, width: nodeWidth, height: nodeHeight })),
    edges: tree.tgos.flatMap((tgo) =>
      (tgo.prerequisites ?? []).map((prereq) => ({
        id: `${prereq}->${tgo.code}`,
        sources: [prereq],
        targets: [tgo.code],
      }))
    ),
  })

  const nodes = (graph.children ?? []).map((node) => ({
    id: node.id,
    x: node.x ?? 0,
    y: node.y ?? 0,
    w: nodeWidth,
    h: nodeHeight,
    ...byCode[node.id],
  }))
  const routes = (graph.edges ?? []).map((edge) => {
    const section = edge.sections?.[0]
    return {
      id: edge.id,
      source: edge.sources[0],
      target: edge.targets[0],
      points: simplifyPolyline(
        [section?.startPoint, ...(section?.bendPoints ?? []), section?.endPoint]
          .filter(Boolean)
          .map((point) => ({ x: point.x, y: point.y }))
      ),
    }
  })
  const bridgeMap = buildBridgeMap(routes, byCode)

  const width = Math.ceil(Math.max(...nodes.map((node) => node.x + node.w), 1600) + padding)
  const height = Math.ceil(Math.max(...nodes.map((node) => node.y + node.h), 900) + padding)
  const edgeSvg = routes
    .map((route) => renderEdge(route, bridgeMap.get(route.id) ?? [], byCode))
    .join('\n')
  const nodeSvg = nodes.map(renderNode).join('\n')

  const html = `<!doctype html><html><head><meta charset="utf-8" />
  <style>
    body { margin: 0; background: #09090b; }
    .frame { width: ${width}px; min-height: ${height}px; background: radial-gradient(circle at top left, rgba(34,211,238,0.16), transparent 30%), radial-gradient(circle at bottom right, rgba(245,158,11,0.12), transparent 28%), linear-gradient(180deg, rgba(24,24,27,0.98), rgba(9,9,11,0.98)); color: white; font-family: ui-sans-serif, system-ui, sans-serif; position: relative; overflow: hidden; }
    .title { padding: 26px 32px 8px; font-size: 28px; font-weight: 800; }
    .intro { padding: 0 32px 18px; max-width: 860px; color: #a1a1aa; font-size: 14px; line-height: 1.5; }
    .legend { position: absolute; top: 24px; right: 24px; display: flex; gap: 8px; }
    .pill { padding: 6px 10px; border-radius: 999px; font-size: 11px; letter-spacing: 0.12em; text-transform: uppercase; border: 1px solid rgba(255,255,255,0.1); }
    .active { background: rgba(34,211,238,0.14); color: #d9fbff; }
    .unlocked { background: rgba(245,158,11,0.14); color: #fff3d1; }
    .locked { background: rgba(255,255,255,0.06); color: #d4d4d8; }
    svg { display: block; }
  </style></head><body>
    <div class="frame">
      <div class="title">${esc(tree.title)}</div>
      <div class="intro">Exported built-in tree map using the live ELK layout, simplified cards, and bridge/underpass edge treatment.</div>
      <div class="legend"><div class="pill active">Seed / Active</div><div class="pill unlocked">Unlocked Branch</div><div class="pill locked">Locked</div></div>
      <svg width="${width}" height="${height}" viewBox="0 0 ${width} ${height}" xmlns="http://www.w3.org/2000/svg"><g transform="translate(0, 36)">${edgeSvg}${nodeSvg}</g></svg>
    </div>
  </body></html>`

  const htmlPath = path.join(outputDir, `${slug}-map.html`)
  const pngPath = path.join(outputDir, `${slug}-map.png`)
  fs.writeFileSync(htmlPath, html)

  const browser = await chromium.launch({ headless: true })
  const page = await browser.newPage({
    viewport: { width: Math.min(width, 2000), height: Math.min(height, 1400) },
    deviceScaleFactor: 2,
  })
  await page.goto(`file://${htmlPath}`)
  await page.screenshot({ path: pngPath, fullPage: true })
  await browser.close()

  console.log(JSON.stringify({ slug, htmlPath, pngPath, width, height }, null, 2))
}

function exportBuiltInTree(slug, outputDir) {
  const goPath = path.join(process.cwd(), 'tree-export-one.go')
  const jsonPath = path.join(outputDir, `${slug}.json`)
  const goSource = `package main
import (
  "encoding/json"
  "os"
  "github.com/tomasino/writing-coach/internal/domain"
)
type outTGO struct { Code string \`json:"code"\`; Title string \`json:"title"\`; Description string \`json:"description"\`; Stage string \`json:"stage"\`; StageOrder int \`json:"stage_order"\`; MasteryHint string \`json:"mastery_hint"\`; Prerequisites []string \`json:"prerequisites"\` }
type outTree struct { Slug string \`json:"slug"\`; Title string \`json:"title"\`; Description string \`json:"description"\`; TGOs []outTGO \`json:"tgos"\` }
func main() {
  tree, ok := domain.BuiltInTreeBySlug(${JSON.stringify(slug)})
  if !ok { panic("tree not found") }
  out := outTree{Slug: tree.Slug, Title: tree.Title, Description: tree.Description, TGOs: make([]outTGO, 0, len(tree.TGOs))}
  for _, tgo := range tree.TGOs {
    out.TGOs = append(out.TGOs, outTGO{Code: tgo.Code, Title: tgo.Title, Description: tgo.Description, Stage: tgo.Stage, StageOrder: tgo.StageOrder, MasteryHint: tgo.MasteryHint, Prerequisites: append([]string(nil), tgo.Prerequisites...)})
  }
  f, err := os.Create(${JSON.stringify(jsonPath)})
  if err != nil { panic(err) }
  defer f.Close()
  if err := json.NewEncoder(f).Encode(out); err != nil { panic(err) }
}`

  fs.writeFileSync(goPath, goSource)
  try {
    execFileSync('go', ['run', goPath], { cwd: process.cwd(), stdio: 'inherit' })
  } finally {
    fs.rmSync(goPath, { force: true })
  }

  return JSON.parse(fs.readFileSync(jsonPath, 'utf8'))
}

function buildUnlocks(tree) {
  const unlocks = new Map()
  for (const tgo of tree.tgos) {
    for (const prereq of tgo.prerequisites ?? []) {
      const next = unlocks.get(prereq) ?? []
      next.push(tgo.code)
      unlocks.set(prereq, next)
    }
  }
  return unlocks
}

function statusFor(prereqs, unlocks) {
  if (prereqs.length === 0) {
    return 'active'
  }
  if (unlocks.length === 0) {
    return 'locked'
  }
  return 'unlocked'
}

function tone(status) {
  switch (status) {
    case 'active':
      return { shell: '#0e3a43', border: '#22d3ee', accent: '#22d3ee' }
    case 'completed':
      return { shell: '#113326', border: '#34d399', accent: '#34d399' }
    case 'unlocked':
      return { shell: '#3d2a0c', border: '#f59e0b', accent: '#f59e0b' }
    default:
      return { shell: '#111217', border: '#71717a', accent: '#71717a' }
  }
}

function edgeColor(sourceStatus, targetStatus) {
  if (sourceStatus === 'completed' && targetStatus === 'completed') return '#34d399'
  if (targetStatus === 'active' || sourceStatus === 'active') return '#22d3ee'
  if (targetStatus === 'unlocked') return '#f59e0b'
  return '#71717a'
}

function simplifyPolyline(points) {
  const simplified = []
  for (const point of points) {
    const last = simplified[simplified.length - 1]
    const prev = simplified[simplified.length - 2]
    if (last && nearlyEqual(last.x, point.x) && nearlyEqual(last.y, point.y)) continue
    if (
      prev &&
      ((nearlyEqual(prev.x, last.x) && nearlyEqual(last.x, point.x)) ||
        (nearlyEqual(prev.y, last.y) && nearlyEqual(last.y, point.y)))
    ) {
      simplified[simplified.length - 1] = point
      continue
    }
    simplified.push(point)
  }
  return simplified
}

function buildBridgeMap(routes, byCode) {
  const out = new Map()
  const segments = routes.flatMap(getSegments)
  for (let i = 0; i < segments.length; i++) {
    for (let j = i + 1; j < segments.length; j++) {
      const a = segments[i]
      const b = segments[j]
      if (a.edgeId === b.edgeId || sharesEndpoint(a, b) || a.orientation === b.orientation) continue
      const h = a.orientation === 'horizontal' ? a : b
      const v = a.orientation === 'vertical' ? a : b
      const c = crossing(h.start, h.end, v.start, v.end)
      if (!c) continue
      if (chooseBridgeOwner(h.edgeId, v.edgeId, byCode) !== h.edgeId) continue
      const next = out.get(h.edgeId) ?? []
      next.push({ x: c.x, y: c.y, segmentIndex: h.segmentIndex })
      out.set(h.edgeId, next)
    }
  }
  return out
}

function renderEdge(route, bridges, byCode) {
  const target = byCode[route.target]
  const source = byCode[route.source]
  const stroke = edgeColor(source.status, target.status)
  const strokeWidth = target.status === 'locked' ? 1.8 : 2.4
  const opacity = target.status === 'locked' ? 0.58 : 0.92
  const d = buildBridgePath(route.points, bridges)
  const a = buildArrow(route.points, strokeWidth)
  return `
    <path d="${d}" fill="none" stroke="rgba(9,9,11,0.92)" stroke-width="${strokeWidth + 4}" stroke-linecap="round" stroke-linejoin="round" opacity="${opacity}" />
    <path d="${d}" fill="none" stroke="${stroke}" stroke-width="${strokeWidth}" stroke-linecap="round" stroke-linejoin="round" opacity="${opacity}" />
    <path d="${a.path}" transform="${a.transform}" fill="none" stroke="rgba(9,9,11,0.92)" stroke-width="${strokeWidth + 3}" stroke-linecap="round" stroke-linejoin="round" opacity="${opacity}" />
    <path d="${a.path}" transform="${a.transform}" fill="none" stroke="${stroke}" stroke-width="${strokeWidth}" stroke-linecap="round" stroke-linejoin="round" opacity="${opacity}" />`
}

function renderNode(node) {
  const t = tone(node.status)
  return `
    <g transform="translate(${node.x}, ${node.y})">
      <rect x="0" y="0" width="${node.w}" height="${node.h}" rx="24" fill="${t.shell}" stroke="${t.border}" stroke-width="1.5" />
      <text x="16" y="32" fill="#ffffff" font-size="13" font-weight="700" font-family="ui-sans-serif, system-ui, sans-serif">${esc(node.title)}</text>
      <circle cx="${node.w - 18}" cy="${node.h / 2}" r="5" fill="${t.accent}" stroke="rgba(255,255,255,0.2)" stroke-width="1" />
    </g>`
}

function getSegments(route) {
  const segments = []
  for (let i = 0; i < route.points.length - 1; i++) {
    const start = route.points[i]
    const end = route.points[i + 1]
    if (nearlyEqual(start.x, end.x) && nearlyEqual(start.y, end.y)) continue
    if (nearlyEqual(start.y, end.y)) {
      segments.push({ edgeId: route.id, source: route.source, target: route.target, start, end, orientation: 'horizontal', segmentIndex: i })
    } else if (nearlyEqual(start.x, end.x)) {
      segments.push({ edgeId: route.id, source: route.source, target: route.target, start, end, orientation: 'vertical', segmentIndex: i })
    }
  }
  return segments
}

function sharesEndpoint(a, b) {
  return a.source === b.source || a.source === b.target || a.target === b.source || a.target === b.target
}

function crossing(hs, he, vs, ve) {
  const minX = Math.min(hs.x, he.x)
  const maxX = Math.max(hs.x, he.x)
  const minY = Math.min(vs.y, ve.y)
  const maxY = Math.max(vs.y, ve.y)
  const x = vs.x
  const y = hs.y
  if (x <= minX + 8 || x >= maxX - 8) return null
  if (y <= minY + 8 || y >= maxY - 8) return null
  return { x, y }
}

function chooseBridgeOwner(edgeA, edgeB, byCode) {
  const weight = (id) => {
    const [source, target] = id.split('->')
    const sw = { active: 6, completed: 4, unlocked: 3, locked: 1 }[byCode[source].status] ?? 0
    const tw = { active: 6, completed: 4, unlocked: 3, locked: 1 }[byCode[target].status] ?? 0
    return sw + tw
  }
  const aw = weight(edgeA)
  const bw = weight(edgeB)
  if (aw !== bw) return aw > bw ? edgeA : edgeB
  return edgeA < edgeB ? edgeA : edgeB
}

function buildBridgePath(points, bridges) {
  const radius = 11
  const lift = 9
  const bySegment = new Map()
  for (const bridge of bridges) {
    const next = bySegment.get(bridge.segmentIndex) ?? []
    next.push(bridge)
    bySegment.set(bridge.segmentIndex, next)
  }
  let path = `M ${points[0].x} ${points[0].y}`
  for (let i = 0; i < points.length - 1; i++) {
    const a = points[i]
    const b = points[i + 1]
    const segmentBridges = bySegment.get(i)
    if (!segmentBridges || !nearlyEqual(a.y, b.y)) {
      path += ` L ${b.x} ${b.y}`
      continue
    }
    const direction = Math.sign(b.x - a.x) || 1
    const sorted = [...segmentBridges].sort((m, n) => (direction > 0 ? m.x - n.x : n.x - m.x))
    let cursor = a.x
    let last = null
    for (const bridge of sorted) {
      if (last !== null && Math.abs(bridge.x - last) < radius * 2.4) continue
      const startX = bridge.x - direction * radius
      const endX = bridge.x + direction * radius
      path += ` L ${startX} ${a.y}`
      path += ` Q ${bridge.x} ${a.y - lift} ${endX} ${a.y}`
      cursor = endX
      last = bridge.x
    }
    if (!nearlyEqual(cursor, b.x)) path += ` L ${b.x} ${b.y}`
  }
  return path
}

function buildArrow(points, strokeWidth) {
  const tip = points[points.length - 1]
  const base = points[points.length - 2] ?? tip
  const angle = (Math.atan2(tip.y - base.y, tip.x - base.x) * 180) / Math.PI
  const size = Math.max(9, strokeWidth * 4.5)
  return {
    path: `M ${-size} ${-size * 0.58} L 0 0 L ${-size} ${size * 0.58}`,
    transform: `translate(${tip.x}, ${tip.y}) rotate(${angle})`,
  }
}

function nearlyEqual(a, b) {
  return Math.abs(a - b) < 0.5
}

function esc(value) {
  return String(value).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;')
}
