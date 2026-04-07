import nextPlugin from '@next/eslint-plugin-next'
import tseslint from 'typescript-eslint'

const config = [
  {
    ignores: ['.next/**', 'node_modules/**', 'test-results/**', 'playwright-report/**'],
  },
  {
    files: ['**/*.{js,jsx,mjs,ts,tsx,mts,cts}'],
    languageOptions: {
      parser: tseslint.parser,
      parserOptions: {
        ecmaVersion: 'latest',
        sourceType: 'module',
        ecmaFeatures: { jsx: true },
      },
    },
  },
  nextPlugin.configs['core-web-vitals'],
  {
    rules: {
      '@next/next/no-img-element': 'off',
    },
  },
]

export default config
