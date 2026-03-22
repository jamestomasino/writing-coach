import { expect, test } from '@playwright/test'
import { closeContext, createFirstAssignment, createTrack, newUserPage } from './helpers'

function uniqueUserSlug(prefix: string, workerIndex: number) {
  return `${prefix}-${workerIndex}-${Date.now()}`
}

test('saves playground drafts, reviews twice, and shows history with comparison', async ({ browser }, testInfo) => {
  const userSlug = uniqueUserSlug('e2e-playground', testInfo.workerIndex)
  const { context, page } = await newUserPage(browser, userSlug)
  try {
    await page.goto('/onboarding')
    await createTrack(page, 'playground')
    await createFirstAssignment(page)

    await page.goto('/playground')
    await expect(page.getByRole('heading', { name: 'New review' })).toBeVisible()

    const content = page.getByLabel('Text to review')
    await content.fill('Writing Coach helps writers improve through focused assignments and feedback they can use right away.')

    await page.getByRole('button', { name: 'Save draft' }).click()
    await expect(page).toHaveURL(/\/playground\/\d+$/)
    await expect(page.getByText(/Saved draft:/)).toBeVisible()

    await page.getByRole('button', { name: 'Review text' }).click()
    await expect(page.getByRole('heading', { name: 'Playground feedback' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Saved drafts' })).toBeVisible()
    await expect(page.getByRole('heading', { name: 'Earlier reviews' })).toBeVisible()

    await content.fill('Writing Coach helps writers improve through focused assignments, clearer revision steps, and feedback they can act on quickly.')
    await page.getByRole('button', { name: 'Save draft' }).click()
    await expect(page.getByText('Draft #2')).toBeVisible()

    await page.getByRole('button', { name: 'Review text' }).click()
    await expect(page.getByRole('heading', { name: 'What changed' })).toBeVisible()

    await page.goto('/playground/history')
    await expect(page.getByRole('heading', { name: 'History' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'Open' }).first()).toBeVisible()
  } finally {
    await closeContext(context)
  }
})
