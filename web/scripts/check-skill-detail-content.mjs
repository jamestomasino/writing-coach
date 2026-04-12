import fs from 'node:fs'
import path from 'node:path'
import vm from 'node:vm'
import ts from 'typescript'

const repoRoot = path.resolve(path.dirname(new URL(import.meta.url).pathname), '..', '..')
const webRoot = path.resolve(repoRoot, 'web')

const skillDetailsPath = path.resolve(webRoot, 'src/lib/skill-details.ts')
const skillDefinitionsPath = path.resolve(repoRoot, 'internal/domain/skills.go')

function loadTsModule(filePath) {
  const source = fs.readFileSync(filePath, 'utf8')
  const compiled = ts.transpileModule(source, {
    compilerOptions: {
      module: ts.ModuleKind.CommonJS,
      target: ts.ScriptTarget.ES2020,
      esModuleInterop: true,
    },
  }).outputText

  const cjsModule = { exports: {} }
  const localRequire = (name) => {
    throw new Error(`Unexpected require in transpiled module: ${name}`)
  }
  const context = vm.createContext({ module: cjsModule, exports: cjsModule.exports, require: localRequire, console })
  vm.runInContext(compiled, context, { filename: filePath })
  return cjsModule.exports
}

function parseGoSkillDefinitions(filePath) {
  const source = fs.readFileSync(filePath, 'utf8')
  const pattern = /\{Name:\s*"([^"]+)",\s*Tier:\s*SkillTier(\w+)\}/g
  const results = []
  let match
  while ((match = pattern.exec(source)) !== null) {
    const tierToken = match[2].toLowerCase()
    const tier = tierToken === 'specialty' ? 'specialty' : tierToken === 'domain' ? 'domain' : 'core'
    results.push({ name: match[1], tier })
  }
  return results
}

function words(text) {
  return (text.match(/[A-Za-z]+/g) ?? []).length
}

function sentenceChunks(text) {
  return text
    .split(/[.!?]/)
    .map((item) => item.trim())
    .filter(Boolean)
}

function hasActionVerb(line) {
  return /^(add|cut|keep|name|move|split|replace|remove|read|pick|mark|write|answer|choose|link|show|track|run|use|start|reorder|standardize|underline|give|explain|raise|do|tag)\b/i.test(
    line.trim()
  )
}

function main() {
  const skillModule = loadTsModule(skillDetailsPath)
  const allSkillDetails = skillModule.allSkillDetails

  if (!Array.isArray(allSkillDetails) || allSkillDetails.length === 0) {
    throw new Error('skill-details export `allSkillDetails` is missing or empty')
  }

  const canonicalSkills = parseGoSkillDefinitions(skillDefinitionsPath)
  const canonicalByName = new Map(canonicalSkills.map((item) => [item.name, item]))
  const detailByName = new Map(allSkillDetails.map((item) => [item.name, item]))

  const failures = []

  for (const item of canonicalSkills) {
    if (!detailByName.has(item.name)) {
      failures.push(`Missing skill detail for: ${item.name}`)
    }
  }

  for (const item of allSkillDetails) {
    const canonical = canonicalByName.get(item.name)
    if (!canonical) {
      failures.push(`Extra skill detail not in canonical skill list: ${item.name}`)
      continue
    }
    if (canonical.tier !== item.tier) {
      failures.push(`Tier mismatch for ${item.name}: detail=${item.tier} canonical=${canonical.tier}`)
    }

    const requiredFields = [
      'oneLine',
      'whatItMeans',
      'whyItMatters',
      'strongExample',
      'weakExample',
      'coachTip',
    ]

    for (const field of requiredFields) {
      if (typeof item[field] !== 'string' || item[field].trim().length === 0) {
        failures.push(`${item.name}: missing required text field ${field}`)
      }
    }

    if (!Array.isArray(item.lookFor) || item.lookFor.length < 2) {
      failures.push(`${item.name}: lookFor must contain at least 2 checks`)
    }
    if (!Array.isArray(item.revisionMoves) || item.revisionMoves.length < 2 || item.revisionMoves.length > 4) {
      failures.push(`${item.name}: revisionMoves must contain 2-4 action items`)
    }
    if (!Array.isArray(item.contentSources) || item.contentSources.length === 0) {
      failures.push(`${item.name}: contentSources must not be empty`)
    }

    if (typeof item.oneLine === 'string' && words(item.oneLine) > 18) {
      failures.push(`${item.name}: oneLine is too long (${words(item.oneLine)} words)`)
    }

    for (const field of ['whatItMeans', 'whyItMatters']) {
      const value = item[field]
      if (typeof value === 'string') {
        const chunks = sentenceChunks(value)
        if (chunks.length > 3) {
          failures.push(`${item.name}: ${field} should be 1-3 short sentences`)
        }
        const longest = Math.max(0, ...chunks.map((chunk) => words(chunk)))
        if (longest > 22) {
          failures.push(`${item.name}: ${field} has a sentence that is too long (${longest} words)`)
        }
      }
    }

    if (typeof item.strongExample === 'string' && !/^Strong:/i.test(item.strongExample.trim())) {
      failures.push(`${item.name}: strongExample must begin with "Strong:"`)
    }
    if (typeof item.weakExample === 'string' && !/^Weak:/i.test(item.weakExample.trim())) {
      failures.push(`${item.name}: weakExample must begin with "Weak:"`)
    }

    if (Array.isArray(item.revisionMoves)) {
      for (const [index, move] of item.revisionMoves.entries()) {
        if (!hasActionVerb(move)) {
          failures.push(`${item.name}: revisionMoves[${index}] should start with an action verb`)
        }
      }
    }

    const bannedPhrases = ['loser', 'stupid', 'obviously', 'just write better', 'always do this']
    const inspectFields = [item.oneLine, item.whatItMeans, item.whyItMatters, item.coachTip, ...item.revisionMoves]
      .filter((value) => typeof value === 'string')
      .join(' ')
      .toLowerCase()

    for (const phrase of bannedPhrases) {
      if (inspectFields.includes(phrase)) {
        failures.push(`${item.name}: contains banned phrase "${phrase}"`)
      }
    }
  }

  if (failures.length > 0) {
    console.error('Skill detail content check failed:\n')
    for (const failure of failures) {
      console.error(`- ${failure}`)
    }
    process.exit(1)
  }

  console.log(`Skill detail content check passed for ${allSkillDetails.length} skills.`)
}

main()
