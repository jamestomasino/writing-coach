import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

const providersPath = resolve('src/components/app-providers.tsx')
const content = readFileSync(providersPath, 'utf8')

if (!/NextIntlClientProvider[\s\S]*timeZone\s*=\s*"[^"]+"/.test(content)) {
  console.error('i18n timezone guard failed: NextIntlClientProvider must set a global timeZone')
  process.exit(1)
}

console.log('i18n timezone guard passed')
