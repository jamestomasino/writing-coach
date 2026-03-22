import { defineConfig, devices } from '@playwright/test'
import path from 'node:path'

const rootDir = path.resolve(__dirname, '..')
const testDataDir = path.join(rootDir, '.writing-coach-e2e')
const apiPort = 18080
const webPort = 13001
const apiToken = 'e2e-api-token'

export default defineConfig({
  testDir: './tests/e2e',
  timeout: 60_000,
  expect: {
    timeout: 10_000,
  },
  fullyParallel: true,
  reporter: 'list',
  use: {
    baseURL: `http://127.0.0.1:${webPort}`,
    trace: 'on-first-retry',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
  webServer: [
    {
      command: `rm -rf ${testDataDir} && go run ./cmd/writing-coach serve`,
      cwd: rootDir,
      url: `http://127.0.0.1:${apiPort}/api/ready`,
      reuseExistingServer: !process.env.CI,
      stdout: 'pipe',
      stderr: 'pipe',
      env: {
        ...process.env,
        WRITING_COACH_HTTP_ADDR: `127.0.0.1:${apiPort}`,
        WRITING_COACH_DATA_DIR: testDataDir,
        WRITING_COACH_DATABASE_URL: path.join(testDataDir, 'writing-coach-e2e.db'),
        WRITING_COACH_DEFAULT_TREE_SLUG: 'global-writing-skill-graph',
        WRITING_COACH_API_TOKEN: apiToken,
        WRITING_COACH_AI_KEY_SECRET: 'e2e-ai-key-secret',
        OPENAI_API_KEY: 'e2e-fallback-key',
        OPENAI_BASE_URL: 'http://127.0.0.1:9/v1',
      },
    },
    {
      command: `npm run dev -- --hostname 127.0.0.1 --port ${webPort}`,
      cwd: __dirname,
      url: `http://127.0.0.1:${webPort}/about`,
      reuseExistingServer: !process.env.CI,
      stdout: 'pipe',
      stderr: 'pipe',
      env: {
        ...process.env,
        WRITING_COACH_API_INTERNAL_URL: `http://127.0.0.1:${apiPort}`,
      },
    },
  ],
})
