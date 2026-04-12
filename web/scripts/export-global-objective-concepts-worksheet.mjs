import fs from 'node:fs'
import path from 'node:path'
import vm from 'node:vm'
import ts from 'typescript'

const repoRoot = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..', '..')
const webRoot = path.resolve(repoRoot, 'web')
const outputPath = path.resolve(repoRoot, '.tmp', 'global-objective-concepts-worksheet.md')

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

function esc(value) {
  return String(value ?? '').replaceAll('|', '\\|').replaceAll('\n', ' ')
}

function main() {
  const { buildObjectiveConcepts } = loadTsModule(objectiveConceptsPath)
  const { buildObjectiveDetail } = loadTsModule(objectiveDetailsPath)
  const nodes = parseNodes(treeCatalogPath)
  const { concepts } = buildObjectiveConcepts(nodes)

  const lines = []
  lines.push('# Global Objective Concepts Worksheet')
  lines.push('')
  lines.push(`Generated: ${new Date().toISOString()}`)
  lines.push(`Concept pages: ${concepts.length}`)
  lines.push('')
  lines.push('| Concept key | Title | Skill | Representative code | Good Example | Needs Work Example |')
  lines.push('| --- | --- | --- | --- | --- | --- |')

  for (const concept of concepts) {
    const detail = buildObjectiveDetail(concept.representative, concept.key)
    lines.push(
      `| ${esc(concept.key)} | ${esc(concept.title)} | ${esc(concept.skill_name)} | ${esc(concept.representative.code)} | ${esc(detail.goodExample)} | ${esc(detail.badExample)} |`
    )
  }

  fs.mkdirSync(path.dirname(outputPath), { recursive: true })
  fs.writeFileSync(outputPath, `${lines.join('\n')}\n`)
  console.log(`Wrote ${outputPath}`)
}

main()
