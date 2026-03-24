import dagre from 'dagre'
import ELK from 'elkjs/lib/elk.bundled'

export type TreeLayoutStrategy = 'dagre' | 'elk'

type LayoutNodeInput = {
  id: string
  width: number
  height: number
}

type LayoutEdgeInput = {
  source: string
  target: string
}

type LayoutGraphInput = {
  strategy: TreeLayoutStrategy
  nodes: LayoutNodeInput[]
  edges: LayoutEdgeInput[]
}

export type LayoutPoint = {
  x: number
  y: number
}

export type LayoutNodePosition = {
  id: string
  x: number
  y: number
}

export type LayoutEdgeRoute = {
  id: string
  source: string
  target: string
  points: LayoutPoint[]
}

export type TreeLayoutResult = {
  nodes: LayoutNodePosition[]
  edges: LayoutEdgeRoute[]
}

const dagreGraphDefaults = {
  rankdir: 'LR',
  nodesep: 36,
  ranksep: 128,
  marginx: 24,
  marginy: 24,
}

const elk = new ELK()

export async function layoutTreeGraph({ strategy, nodes, edges }: LayoutGraphInput): Promise<TreeLayoutResult> {
  if (strategy === 'elk') {
    return layoutWithElk(nodes, edges)
  }
  return layoutWithDagre(nodes, edges)
}

function layoutWithDagre(nodes: LayoutNodeInput[], edges: LayoutEdgeInput[]): TreeLayoutResult {
  const graph = new dagre.graphlib.Graph()
  graph.setDefaultEdgeLabel(() => ({}))
  graph.setGraph(dagreGraphDefaults)

  for (const node of nodes) {
    graph.setNode(node.id, { width: node.width, height: node.height })
  }

  for (const edge of edges) {
    graph.setEdge(edge.source, edge.target)
  }

  dagre.layout(graph)

  return {
    nodes: nodes.map((node) => {
      const positioned = graph.node(node.id)
      return {
        id: node.id,
        x: positioned.x - node.width / 2,
        y: positioned.y - node.height / 2,
      }
    }),
    edges: graph.edges().map((edge) => ({
      id: `${edge.v}->${edge.w}`,
      source: edge.v,
      target: edge.w,
      points: simplifyPolyline((graph.edge(edge).points ?? []).map(toPoint)),
    })),
  }
}

async function layoutWithElk(nodes: LayoutNodeInput[], edges: LayoutEdgeInput[]): Promise<TreeLayoutResult> {
  const graph = (await elk.layout({
    id: 'skill-tree',
    layoutOptions: {
      'elk.algorithm': 'layered',
      'elk.direction': 'RIGHT',
      'elk.edgeRouting': 'ORTHOGONAL',
      'elk.layered.considerModelOrder': 'NODES_AND_EDGES',
      'elk.layered.crossingMinimization.strategy': 'LAYER_SWEEP',
      'elk.layered.nodePlacement.strategy': 'BRANDES_KOEPF',
      'elk.layered.spacing.nodeNodeBetweenLayers': '128',
      'elk.padding': '[top=24,left=24,bottom=24,right=24]',
      'elk.spacing.nodeNode': '36',
    },
    children: nodes.map((node) => ({
      id: node.id,
      width: node.width,
      height: node.height,
    })),
    edges: edges.map((edge) => ({
      id: `${edge.source}->${edge.target}`,
      sources: [edge.source],
      targets: [edge.target],
    })),
  })) as {
    children?: Array<{ id: string; x?: number; y?: number }>
    edges?: Array<{
      id: string
      sources: string[]
      targets: string[]
      sections?: Array<{
        startPoint?: { x: number; y: number }
        bendPoints?: Array<{ x: number; y: number }>
        endPoint?: { x: number; y: number }
      }>
    }>
  }

  return {
    nodes: (graph.children ?? []).map((node) => ({
      id: node.id,
      x: node.x ?? 0,
      y: node.y ?? 0,
    })),
    edges: (graph.edges ?? []).map((edge) => {
      const section = edge.sections?.[0]
      const points = simplifyPolyline(
        [section?.startPoint, ...(section?.bendPoints ?? []), section?.endPoint]
          .filter((point): point is { x: number; y: number } => point != null)
          .map(toPoint)
      )

      return {
        id: edge.id,
        source: edge.sources[0],
        target: edge.targets[0],
        points,
      }
    }),
  }
}

function toPoint(point: { x: number; y: number }): LayoutPoint {
  return { x: point.x, y: point.y }
}

function simplifyPolyline(points: LayoutPoint[]): LayoutPoint[] {
  const simplified: LayoutPoint[] = []

  for (const point of points) {
    const last = simplified[simplified.length - 1]
    if (last && nearlyEqual(last.x, point.x) && nearlyEqual(last.y, point.y)) {
      continue
    }

    const prev = simplified[simplified.length - 2]
    if (prev && last && isCollinear(prev, last, point)) {
      simplified[simplified.length - 1] = point
      continue
    }

    simplified.push(point)
  }

  return simplified
}

function isCollinear(a: LayoutPoint, b: LayoutPoint, c: LayoutPoint) {
  return (nearlyEqual(a.x, b.x) && nearlyEqual(b.x, c.x)) || (nearlyEqual(a.y, b.y) && nearlyEqual(b.y, c.y))
}

function nearlyEqual(a: number, b: number) {
  return Math.abs(a - b) < 0.5
}
