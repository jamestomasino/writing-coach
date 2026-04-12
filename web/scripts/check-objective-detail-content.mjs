import fs from 'node:fs'
import path from 'node:path'
import vm from 'node:vm'
import ts from 'typescript'

const repoRoot = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..', '..')
const webRoot = path.resolve(repoRoot, 'web')

const objectiveDetailsPath = path.resolve(webRoot, 'src/lib/objective-details.ts')
const skillDetailsPath = path.resolve(webRoot, 'src/lib/skill-details.ts')
const treeCatalogPath = path.resolve(repoRoot, 'internal/domain/tree_catalog.go')

const moduleCache = new Map()

function resolveLocalModulePath(baseFile, specifier) {
  const candidate = path.resolve(path.dirname(baseFile), specifier)
  const withTs = candidate.endsWith('.ts') ? candidate : `${candidate}.ts`
  if (fs.existsSync(withTs)) {
    return withTs
  }
  if (fs.existsSync(candidate)) {
    return candidate
  }
  throw new Error(`Cannot resolve local module ${specifier} from ${baseFile}`)
}

function loadTsModule(filePath) {
  if (moduleCache.has(filePath)) {
    return moduleCache.get(filePath)
  }
  const source = fs.readFileSync(filePath, 'utf8')
  const compiled = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2020,
      esModuleInterop: true,
    },
  }).outputText

  const cjsModule = { exports: {} }
  moduleCache.set(filePath, cjsModule.exports)
  const context = vm.createContext({
    module: cjsModule,
    exports: cjsModule.exports,
    require: (specifier) => {
      if (typeof specifier === 'string' && (specifier.startsWith('./') || specifier.startsWith('../'))) {
        const resolved = resolveLocalModulePath(filePath, specifier)
        return loadTsModule(resolved)
      }
      throw new Error(`Unexpected require in transpiled module: ${specifier}`)
    },
    console,
  })
  vm.runInContext(compiled, context, { filename: filePath })
  moduleCache.set(filePath, cjsModule.exports)
  return cjsModule.exports
}

function parseObjectiveNodes(filePath) {
  const source = fs.readFileSync(filePath, 'utf8')
  const pattern = /node\("([^"]+)",\s*"([^"]+)",\s*"([^"]+)",\s*"([^"]+)",\s*"([^"]+)",\s*([0-9]+),\s*"([^"]+)"/g
  const out = []
  let match
  while ((match = pattern.exec(source)) !== null) {
    out.push({
      code: match[1],
      title: match[2],
      skill_name: match[3],
      description: match[4],
      stage: match[5],
      stage_order: Number(match[6]),
      mastery_hint: match[7],
      prerequisites: [],
      unlocks: [],
      source_tree_slug: '',
      source_tree_title: '',
    })
  }
  return out
}

function words(text) {
  return (text.match(/[A-Za-z]+/g) ?? []).length
}

function removeExampleLabel(text) {
  return text.replace(/^Good:\s*/i, '').replace(/^Needs work:\s*/i, '').trim()
}

function escapeRegex(value) {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
}

function containsObjectiveName(text, objectiveTitle) {
  const title = objectiveTitle.trim()
  if (!title) {
    return false
  }
  const variants = new Set([title])
  const normalized = title.replace(/[^A-Za-z0-9\\s]+/g, ' ').replace(/\\s+/g, ' ').trim()
  if (normalized) {
    variants.add(normalized)
  }
  for (const variant of variants) {
    const pattern = new RegExp(`\\\\b${escapeRegex(variant).replace(/\\s+/g, '\\\\s+')}\\\\b`, 'i')
    if (pattern.test(text)) {
      return true
    }
  }
  return false
}

function sentenceChunks(text) {
  return text
    .split(/[.!?]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function hasActionVerb(line) {
  const first = line
    .trim()
    .split(/\s+/)[0]
    ?.toLowerCase()
    .replace(/[^a-z-]/g, '')
  if (!first) {
    return false
  }
  const blocked = new Set(['the', 'a', 'an', 'this', 'that', 'when', 'if', 'because', 'while'])
  if (blocked.has(first)) {
    return false
  }
  return true
}

function main() {
  const objectiveModule = loadTsModule(objectiveDetailsPath)
  const skillModule = loadTsModule(skillDetailsPath)
  const buildObjectiveDetail = objectiveModule.buildObjectiveDetail
  const getSkillDetailByName = skillModule.getSkillDetailByName

  if (typeof buildObjectiveDetail !== 'function') {
    throw new Error('objective-details export `buildObjectiveDetail` is missing')
  }
  if (typeof getSkillDetailByName !== 'function') {
    throw new Error('skill-details export `getSkillDetailByName` is missing')
  }

  const nodes = parseObjectiveNodes(treeCatalogPath)
  if (nodes.length === 0) {
    throw new Error('No objective nodes parsed from tree catalog')
  }

  const failures = []
  const goodByText = new Map()
  const badByText = new Map()
  for (const node of nodes) {
    const family = getSkillDetailByName(node.skill_name)
    const detail = buildObjectiveDetail(node, family)

    const requiredFields = ['code', 'title', 'skillFamily', 'objectiveGoal', 'whyThisObjective', 'goodExample', 'badExample']
    for (const field of requiredFields) {
      if (typeof detail[field] !== 'string' || detail[field].trim().length === 0) {
        failures.push(`${node.code}: missing required text field ${field}`)
      }
    }

    if (!Array.isArray(detail.successLooksLike) || detail.successLooksLike.length < 2) {
      failures.push(`${node.code}: successLooksLike must contain at least 2 checks`)
    }
    if (!Array.isArray(detail.revisionMoves) || detail.revisionMoves.length < 2 || detail.revisionMoves.length > 4) {
      failures.push(`${node.code}: revisionMoves must contain 2-4 action items`)
    }
    if (!Array.isArray(detail.assessmentFocus) || detail.assessmentFocus.length < 2) {
      failures.push(`${node.code}: assessmentFocus must contain at least 2 checks`)
    }

    if (!/^Good:/i.test(detail.goodExample.trim())) {
      failures.push(`${node.code}: goodExample must begin with "Good:"`)
    }
    if (!/^Needs work:/i.test(detail.badExample.trim())) {
      failures.push(`${node.code}: badExample must begin with "Needs work:"`)
    }
    const goodBody = removeExampleLabel(detail.goodExample)
    const badBody = removeExampleLabel(detail.badExample)
    if (words(goodBody) < 12) {
      failures.push(`${node.code}: goodExample must include concrete detail (>= 12 words after label)`)
    }
    if (words(badBody) < 12) {
      failures.push(`${node.code}: badExample must include concrete detail (>= 12 words after label)`)
    }

    const normalizedGood = detail.goodExample.replace(/\s+/g, ' ').trim()
    const normalizedBad = detail.badExample.replace(/\s+/g, ' ').trim()
    const priorGood = goodByText.get(normalizedGood)
    const priorBad = badByText.get(normalizedBad)
    if (priorGood && priorGood !== node.code) {
      failures.push(`${node.code}: goodExample duplicates ${priorGood}; every objective example must be unique`)
    } else {
      goodByText.set(normalizedGood, node.code)
    }
    if (priorBad && priorBad !== node.code) {
      failures.push(`${node.code}: badExample duplicates ${priorBad}; every objective example must be unique`)
    } else {
      badByText.set(normalizedBad, node.code)
    }

    const stalePatterns = [
      /clearly demonstrates/i,
      /in a way the reader can track/i,
      /gestures at/i,
      /reader has to guess what changed/i,
    ]
    for (const pattern of stalePatterns) {
      if (pattern.test(detail.goodExample) || pattern.test(detail.badExample)) {
        failures.push(`${node.code}: example text still matches deprecated catch-all phrasing`)
      }
    }

    for (const field of ['objectiveGoal', 'whyThisObjective']) {
      const value = detail[field]
      if (typeof value !== 'string' || value.trim().length === 0) {
        failures.push(`${node.code}: ${field} must contain guidance text`)
      }
    }

    for (const [field, value] of [
      ['objectiveGoal', detail.objectiveGoal],
      ['whyThisObjective', detail.whyThisObjective],
      ['goodExample', detail.goodExample],
      ['badExample', detail.badExample],
    ]) {
      if (typeof value === 'string' && containsObjectiveName(value, node.title)) {
        failures.push(`${node.code}: ${field} must not repeat the objective title "${node.title}"`)
      }
    }
    if (Array.isArray(detail.successLooksLike)) {
      for (const [index, value] of detail.successLooksLike.entries()) {
        if (containsObjectiveName(value, node.title)) {
          failures.push(`${node.code}: successLooksLike[${index}] must not repeat the objective title "${node.title}"`)
        }
      }
    }

    for (const [index, move] of detail.revisionMoves.entries()) {
      if (!hasActionVerb(move)) {
        failures.push(`${node.code}: revisionMoves[${index}] should start with an action verb`)
      }
    }

    const bannedPhrases = ['loser', 'stupid', 'obviously', 'just write better', 'always do this']
    const inspectFields = [
      detail.objectiveGoal,
      detail.whyThisObjective,
      detail.goodExample,
      detail.badExample,
      ...detail.successLooksLike,
      ...detail.revisionMoves,
      ...detail.assessmentFocus,
    ]
      .filter((value) => typeof value === 'string')
      .join(' ')
      .toLowerCase()

    for (const phrase of bannedPhrases) {
      if (inspectFields.includes(phrase)) {
        failures.push(`${node.code}: contains banned phrase "${phrase}"`)
      }
    }
  }

  if (failures.length > 0) {
    console.error('Objective detail content check failed:\n')
    for (const failure of failures) {
      console.error(`- ${failure}`)
    }
    process.exit(1)
  }

  console.log(`Objective detail content check passed for ${nodes.length} objectives.`)
}

main()
