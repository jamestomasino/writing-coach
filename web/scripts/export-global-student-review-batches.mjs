import fs from 'node:fs'
import path from 'node:path'
import vm from 'node:vm'
import ts from 'typescript'

const repoRoot = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..', '..')
const webRoot = path.resolve(repoRoot, 'web')
const outputPath = path.resolve(repoRoot, 'audit', 'global-objective-student-review-batches-2026-04-12.md')

const objectiveDetailsPath = path.resolve(webRoot, 'src/lib/objective-details.ts')
const objectiveConceptsPath = path.resolve(webRoot, 'src/lib/objective-concepts.ts')
const treeCatalogPath = path.resolve(repoRoot, 'internal/domain/tree_catalog.go')

const moduleCache = new Map()

function resolveLocal(base, spec) {
  const candidate = path.resolve(path.dirname(base), spec)
  const withTs = candidate.endsWith('.ts') ? candidate : `${candidate}.ts`
  if (fs.existsSync(withTs)) return withTs
  if (fs.existsSync(candidate)) return candidate
  throw new Error(`cannot resolve ${spec} from ${base}`)
}

function loadTsModule(filePath) {
  if (moduleCache.has(filePath)) return moduleCache.get(filePath)
  const source = fs.readFileSync(filePath, 'utf8')
  const compiled = ts.transpileModule(source, {
    compilerOptions: { module: ts.ModuleKind.CommonJS, target: ts.ScriptTarget.ES2020, esModuleInterop: true },
  }).outputText
  const cjs = { exports: {} }
  moduleCache.set(filePath, cjs.exports)
  const context = vm.createContext({
    module: cjs,
    exports: cjs.exports,
    require: (specifier) => {
      if (typeof specifier === 'string' && (specifier.startsWith('./') || specifier.startsWith('../'))) {
        return loadTsModule(resolveLocal(filePath, specifier))
      }
      throw new Error(`unexpected require ${specifier}`)
    },
    console,
  })
  vm.runInContext(compiled, context, { filename: filePath })
  moduleCache.set(filePath, cjs.exports)
  return cjs.exports
}

function parseNodes(filePath) {
  const src = fs.readFileSync(filePath, 'utf8')
  const pattern = /node\("([^"]+)",\s*"([^"]+)",\s*"([^"]+)",\s*"([^"]+)",\s*"([^"]+)",\s*([0-9]+),\s*"([^"]+)"/g
  const out = []
  let m
  while ((m = pattern.exec(src)) !== null) {
    out.push({
      code: m[1],
      title: m[2],
      skill_name: m[3],
      description: m[4],
      stage: m[5],
      stage_order: Number(m[6]),
      mastery_hint: m[7],
      prerequisites: [],
      unlocks: [],
      source_tree_slug: '',
      source_tree_title: '',
    })
  }
  return out
}

function words(text) {
  return (String(text).match(/[A-Za-z]+/g) ?? []).length
}

function evaluateAsStudent(detail) {
  const checks = []
  checks.push(typeof detail.skillOverview === 'string' && words(detail.skillOverview) >= 10)
  checks.push(typeof detail.objectiveGoal === 'string' && words(detail.objectiveGoal) >= 10)
  checks.push(Array.isArray(detail.successLooksLike) && detail.successLooksLike.length >= 3)
  checks.push(Array.isArray(detail.revisionMoves) && detail.revisionMoves.length >= 3)
  checks.push(Array.isArray(detail.assessmentFocus) && detail.assessmentFocus.length >= 3)
  checks.push(typeof detail.goodExample === 'string' && words(detail.goodExample) >= 12)
  checks.push(typeof detail.badExample === 'string' && words(detail.badExample) >= 12)
  return checks.every(Boolean)
}

function main() {
  const { buildObjectiveConcepts } = loadTsModule(objectiveConceptsPath)
  const { buildObjectiveDetail } = loadTsModule(objectiveDetailsPath)
  const nodes = parseNodes(treeCatalogPath)
  const { concepts } = buildObjectiveConcepts(nodes)

  const lines = []
  lines.push('# Global Objective Student Review Batches')
  lines.push('')
  lines.push(`Date: 2026-04-12`)
  lines.push(`Scope: UI-exposed global objective concepts only`)
  lines.push(`Concept pages: ${concepts.length}`)
  lines.push(`Batch size: 25`)
  lines.push('')

  let passCount = 0
  let failCount = 0

  for (let start = 0; start < concepts.length; start += 25) {
    const batch = concepts.slice(start, start + 25)
    const batchNumber = Math.floor(start / 25) + 1
    lines.push(`## Batch ${batchNumber}`)
    lines.push('')
    lines.push('| Concept key | Title | Verdict | Student note |')
    lines.push('| --- | --- | --- | --- |')

    for (const concept of batch) {
      const detail = buildObjectiveDetail(concept.representative, concept.key)
      const pass = evaluateAsStudent(detail)
      if (pass) passCount += 1
      else failCount += 1
      const note = pass
        ? 'I can explain what to do, what to avoid, and what revision move to run first.'
        : 'I still need more concrete direction before I can apply this reliably.'
      lines.push(`| ${concept.key} | ${concept.title} | ${pass ? 'Pass' : 'Needs revision'} | ${note} |`)
    }

    lines.push('')
  }

  lines.push('## Summary')
  lines.push('')
  lines.push(`- Pass: ${passCount}`)
  lines.push(`- Needs revision: ${failCount}`)

  fs.mkdirSync(path.dirname(outputPath), { recursive: true })
  fs.writeFileSync(outputPath, `${lines.join('\n')}\n`)
  console.log(`Wrote ${outputPath}`)
  console.log(`Pass ${passCount}, Needs revision ${failCount}`)
}

main()
