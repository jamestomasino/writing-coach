import { expect, type Browser, type BrowserContext, type Page } from '@playwright/test'

const apiToken = 'e2e-api-token'

export async function newUserPage(browser: Browser, userSlug: string) {
  const context = await browser.newContext({
    extraHTTPHeaders: {
      Authorization: `Bearer ${apiToken}`,
      'X-Writing-Coach-User': userSlug,
    },
  })
  const page = await context.newPage()
  return { context, page }
}

export async function closeContext(context: BrowserContext) {
  try {
    await context.close()
  } catch {}
}

export async function fillTrackForm(page: Page, suffix: string) {
  await expect(page.getByTestId('track-form')).toBeVisible()
  await page.getByLabel('Primary writing domain').selectOption({ index: 1 })
  await page.getByLabel('Common assignment format').selectOption({ index: 1 })
  await page.getByLabel('Experience level').selectOption({ index: 1 })
  await page.getByLabel('Difficulty and intensity').selectOption({ index: 1 })
  await page.getByLabel('Target audience').fill(`Readers ${suffix}`)
  await page.getByLabel('Typical subject matter').fill(`Subject matter ${suffix}`)
  await page.getByLabel('Tone target').fill(`Direct and clear ${suffix}`)
  await page.getByLabel('Writing goals').fill(`Build a stronger writing practice for ${suffix}.`)
  await page.getByText('word choice', { exact: true }).click()
  await page.getByText('build revision discipline', { exact: true }).click()
}

export async function createTrack(page: Page, suffix: string) {
  await fillTrackForm(page, suffix)
  await page.getByTestId('save-track-button').click()
}

export async function createFirstAssignment(page: Page) {
  await expect(page).toHaveURL(/\/new-assignment/)
  const skillButtons = page.locator('[data-testid^="skill-option-"]')
  const selectedCount = page.getByText(/of 3 skills selected\./)
  await expect(skillButtons).toHaveCount(3)

  for (let index = 0; index < 3; index += 1) {
    const countText = (await selectedCount.textContent()) ?? ''
    if (countText.startsWith('3 of 3')) {
      break
    }
    const button = skillButtons.nth(index)
    const className = await button.getAttribute('class')
    if (!className?.includes('bg-stone-900')) {
      await button.click()
    }
  }

  await expect(selectedCount).toHaveText('3 of 3 skills selected.')
  await page.getByTestId('generate-assignment-button').click()
  await expect(page.getByTestId('accept-assignment-button')).toBeVisible()
  await page.getByTestId('accept-assignment-button').click()
  await expect(page).toHaveURL(/\/$/)
}

export async function openTrackMenu(page: Page) {
  await page.getByTestId('active-track-button').click()
}
