import fs from 'node:fs'
import path from 'node:path'
import vm from 'node:vm'
import ts from 'typescript'

const repoRoot = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..', '..')
const webRoot = path.resolve(repoRoot, 'web')
const worksheetPath = path.resolve(repoRoot, '.tmp', 'objective-example-worksheet.md')

const objectiveDetailsPath = path.resolve(webRoot, 'src/lib/objective-details.ts')
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

function esc(value) {
  return String(value ?? '').replaceAll('|', '\\|').replaceAll('\n', ' ')
}

function main() {
  const objectiveModule = loadTsModule(objectiveDetailsPath)
  const buildObjectiveDetail = objectiveModule.buildObjectiveDetail

  if (typeof buildObjectiveDetail !== 'function') {
    throw new Error('objective-details export `buildObjectiveDetail` is missing')
  }

  const nodes = parseObjectiveNodes(treeCatalogPath)
  nodes.sort((a, b) => a.code.localeCompare(b.code))

  const lines = []
  lines.push('# Objective Example Worksheet')
  lines.push('')
  lines.push(`Generated: ${new Date().toISOString()}`)
  lines.push(`Objectives: ${nodes.length}`)
  lines.push('')
  lines.push('| Code | Title | Skill | Good Example | Needs Work Example | Sources |')
  lines.push('| --- | --- | --- | --- | --- | --- |')

  for (const node of nodes) {
    const detail = buildObjectiveDetail(node)
    const sources = (detail.exampleSources ?? []).map((item) => `${item.label} (${item.url})`).join('; ')
    lines.push(
      `| ${esc(node.code)} | ${esc(node.title)} | ${esc(node.skill_name)} | ${esc(detail.goodExample)} | ${esc(
        detail.badExample
      )} | ${esc(sources)} |`
    )
  }

  fs.mkdirSync(path.dirname(worksheetPath), { recursive: true })
  fs.writeFileSync(worksheetPath, `${lines.join('\n')}\n`)
  console.log(`Wrote ${worksheetPath}`)
}

main()
