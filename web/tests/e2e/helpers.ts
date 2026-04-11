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
  await page.getByLabel('What kind of writing is this for?').selectOption({ index: 1 })
  await page.getByLabel('What kind of assignments do you want?').selectOption({ index: 1 })
  await page.getByLabel('Experience level').selectOption({ index: 1 })
  await page.getByLabel('How hard should this push you?').selectOption({ index: 1 })
  await page.getByLabel('Who are you writing for?').fill(`Readers ${suffix}`)
  await page.getByLabel('What topics or situations should this draw from?').fill(`Subject matter ${suffix}`)
  await page.getByLabel('What tone are you aiming for?').fill(`Direct and clear ${suffix}`)
  await page.getByLabel('What do you want to get better at?').fill(`Build a stronger writing practice for ${suffix}.`)
  await page.getByRole('checkbox', { name: 'word choice' }).check()
  await page.getByRole('checkbox', { name: 'build revision discipline' }).check()
}

export async function createTrack(page: Page, suffix: string, options?: { writingTypeIndex?: number }) {
  await fillTrackForm(page, suffix)
  if (typeof options?.writingTypeIndex === 'number') {
    await page.getByLabel('What kind of writing is this for?').selectOption({ index: options.writingTypeIndex })
  }
  await Promise.all([
    page.getByTestId('save-track-button').click(),
    page.waitForURL(/\/(?:new-assignment)?$/),
  ])
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

export async function switchToOtherTrack(page: Page, options?: { requireChange?: boolean }) {
  const requireChange = options?.requireChange ?? true
  const activeLabel = ((await page.getByTestId('active-track-button').textContent()) ?? '').trim()
  const trackOptions = page.locator('[data-testid^="track-option-"]')
  const count = await trackOptions.count()
  for (let index = 0; index < count; index += 1) {
    await trackOptions.nth(index).click()
    await page.waitForTimeout(150)
    const nextLabel = ((await page.getByTestId('active-track-button').textContent()) ?? '').trim()
    if (nextLabel !== '' && nextLabel !== activeLabel) {
      return true
    }
    if (index < count - 1) {
      await openTrackMenu(page)
    }
  }
  if (requireChange) {
    throw new Error('No non-active track option found to switch to')
  }
  return false
}
