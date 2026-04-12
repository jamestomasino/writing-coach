import type { SkillGraphNode } from './types'

export type ObjectiveConcept = {
  key: string
  title: string
  description: string
  skill_name?: string
  stage: string
  stage_order: number
  codes: string[]
  representative: SkillGraphNode
}

const conceptAliasBySuffix: Record<string, string> = {
  'causal-thread': 'causal-clarity',
}

const conceptTitleByKey: Record<string, string> = {
  'causal-clarity': 'Causal Clarity',
}

const conceptAliasByTitle: Record<string, string> = {
  'causal thread': 'causal clarity',
}

function slugify(value: string) {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
}

function canonicalTitle(title: string) {
  const normalized = title.trim().toLowerCase()
  const mapped = conceptAliasByTitle[normalized] ?? normalized
  return mapped
}

export function objectiveConceptKey(codeOrTitle: string) {
  const normalized = codeOrTitle.trim().toLowerCase()
  if (!normalized) {
    return normalized
  }
  // Keep backward compatibility for direct code aliases.
  if (conceptAliasBySuffix[normalized]) {
    return conceptAliasBySuffix[normalized]
  }
  return slugify(canonicalTitle(normalized))
}

export function buildObjectiveConcepts(nodes: SkillGraphNode[]) {
  const conceptNodes = new Map<string, SkillGraphNode[]>()
  const conceptByCode = new Map<string, string>()

  for (const node of nodes) {
    const key = objectiveConceptKey(node.title)
    const list = conceptNodes.get(key) ?? []
    list.push(node)
    conceptNodes.set(key, list)
    conceptByCode.set(node.code, key)
  }

  const concepts: ObjectiveConcept[] = []
  for (const [key, list] of conceptNodes.entries()) {
    const sorted = [...list].sort((a, b) => {
      if (a.stage_order !== b.stage_order) {
        return a.stage_order - b.stage_order
      }
      return a.code.localeCompare(b.code)
    })
    const representative = sorted[0]
    const codes = sorted.map((item) => item.code)
    const canonical = canonicalTitle(representative.title)
    const defaultTitle = canonical
      .split(' ')
      .map((part) => (part ? part[0].toUpperCase() + part.slice(1) : part))
      .join(' ')
    concepts.push({
      key,
      title: conceptTitleByKey[key] ?? defaultTitle,
      description: representative.description,
      skill_name: representative.skill_name,
      stage: representative.stage,
      stage_order: representative.stage_order,
      codes,
      representative,
    })
  }

  concepts.sort((a, b) => {
    const familyCompare = (a.skill_name ?? '').localeCompare(b.skill_name ?? '')
    if (familyCompare !== 0) {
      return familyCompare
    }
    if (a.stage_order !== b.stage_order) {
      return a.stage_order - b.stage_order
    }
    return a.title.localeCompare(b.title)
  })

  const conceptByKey = new Map(concepts.map((item) => [item.key, item]))
  return { concepts, conceptByKey, conceptByCode }
}
