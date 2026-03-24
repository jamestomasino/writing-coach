'use client'

import { BaseEdge, type Edge, type EdgeProps } from '@xyflow/react'
import type { CSSProperties } from 'react'
import type { LayoutPoint } from './tree-layout'

export type EdgeBridge = {
  x: number
  y: number
  segmentIndex: number
}

export type SkillTreeEdgeData = {
  points: LayoutPoint[]
  bridges: EdgeBridge[]
}

const bridgeRadius = 11
const bridgeLift = 9

export function SkillTreeEdge({
  id,
  data,
  style,
  interactionWidth,
}: EdgeProps<Edge<SkillTreeEdgeData>>) {
  const points = data?.points ?? []
  if (points.length < 2) {
    return null
  }

  const color = typeof style?.stroke === 'string' ? style.stroke : '#71717a'
  const strokeWidth = typeof style?.strokeWidth === 'number' ? style.strokeWidth : 2
  const opacity = typeof style?.opacity === 'number' ? style.opacity : 1
  const sharedStyle: CSSProperties = {
    stroke: color,
    strokeWidth,
    opacity,
  }

  const backgroundPath = buildBridgePath(points, data?.bridges ?? [], bridgeRadius, bridgeLift)
  const strokePath = buildBridgePath(points, data?.bridges ?? [], bridgeRadius, bridgeLift - 1)
  const arrow = buildArrowHead(points, strokeWidth)

  return (
    <>
      <BaseEdge
        id={`${id}-backdrop`}
        path={backgroundPath}
        interactionWidth={0}
        style={{
          stroke: 'rgba(9, 9, 11, 0.92)',
          strokeWidth: strokeWidth + 4,
          opacity,
          strokeLinecap: 'round',
          strokeLinejoin: 'round',
        }}
      />
      <BaseEdge
        id={id}
        path={strokePath}
        interactionWidth={interactionWidth}
        style={{
          ...sharedStyle,
          strokeLinecap: 'round',
          strokeLinejoin: 'round',
        }}
      />
      {arrow ? (
        <>
          <path
            d={arrow.path}
            transform={arrow.transform}
            fill="none"
            stroke="rgba(9, 9, 11, 0.92)"
            strokeWidth={strokeWidth + 3}
            strokeLinecap="round"
            strokeLinejoin="round"
            opacity={opacity}
          />
          <path
            d={arrow.path}
            transform={arrow.transform}
            fill="none"
            stroke={color}
            strokeWidth={strokeWidth}
            strokeLinecap="round"
            strokeLinejoin="round"
            opacity={opacity}
          />
        </>
      ) : null}
    </>
  )
}

function buildBridgePath(points: LayoutPoint[], bridges: EdgeBridge[], radius: number, lift: number) {
  const bridgesBySegment = new Map<number, EdgeBridge[]>()

  for (const bridge of bridges) {
    const next = bridgesBySegment.get(bridge.segmentIndex) ?? []
    next.push(bridge)
    bridgesBySegment.set(bridge.segmentIndex, next)
  }

  let path = `M ${points[0].x} ${points[0].y}`

  for (let i = 0; i < points.length - 1; i++) {
    const current = points[i]
    const next = points[i + 1]
    const segmentBridges = bridgesBySegment.get(i)

    if (!segmentBridges || segmentBridges.length === 0 || Math.abs(current.y - next.y) > 0.5) {
      path += ` L ${next.x} ${next.y}`
      continue
    }

    const direction = Math.sign(next.x - current.x) || 1
    const sorted = [...segmentBridges].sort((a, b) => (direction > 0 ? a.x - b.x : b.x - a.x))
    let cursorX = current.x
    let lastBridgeCenter: number | null = null

    for (const bridge of sorted) {
      if (lastBridgeCenter !== null && Math.abs(bridge.x - lastBridgeCenter) < radius * 2.4) {
        continue
      }
      const startX = bridge.x - direction * radius
      const endX = bridge.x + direction * radius
      path += ` L ${startX} ${current.y}`
      path += ` Q ${bridge.x} ${current.y - lift} ${endX} ${current.y}`
      cursorX = endX
      lastBridgeCenter = bridge.x
    }

    if (!nearlyEqual(cursorX, next.x)) {
      path += ` L ${next.x} ${next.y}`
    }
  }

  return path
}

function nearlyEqual(a: number, b: number) {
  return Math.abs(a - b) < 0.5
}

function buildArrowHead(points: LayoutPoint[], strokeWidth: number) {
  if (points.length < 2) {
    return null
  }

  const tip = points[points.length - 1]
  const base = points[points.length - 2]
  const angle = (Math.atan2(tip.y - base.y, tip.x - base.x) * 180) / Math.PI
  const size = Math.max(9, strokeWidth * 4.5)

  return {
    path: `M ${-size} ${-size * 0.58} L 0 0 L ${-size} ${size * 0.58}`,
    transform: `translate(${tip.x}, ${tip.y}) rotate(${angle})`,
  }
}
