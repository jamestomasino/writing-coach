import { expect, test } from '@playwright/test'
import { closeContext, createFirstAssignment, createTrack, newUserPage, openTrackMenu, switchToOtherTrack } from './helpers'

function uniqueUserSlug(prefix: string, workerIndex: number) {
  return `${prefix}-${workerIndex}-${Date.now()}`
}

test('redirects incomplete users from gated routes into onboarding', async ({ browser }, testInfo) => {
  const userSlug = uniqueUserSlug('e2e-gate', testInfo.workerIndex)
  const { context, page } = await newUserPage(browser, userSlug)
  try {
    await page.goto('/progress')
    await expect(page).toHaveURL(/\/onboarding/)
    await expect(page.getByText('Next, create your first practice path')).toBeVisible()
  } finally {
    await closeContext(context)
  }
})

test('creates the first track and enters the first assignment flow', async ({ browser }, testInfo) => {
  const userSlug = uniqueUserSlug('e2e-first', testInfo.workerIndex)
  const { context, page } = await newUserPage(browser, userSlug)
  try {
    await page.goto('/onboarding')
    await createTrack(page, 'first')
    await expect(page).toHaveURL(/\/new-assignment/)
    await expect(page.getByText('Finish by generating your first assignment')).toBeVisible()

    await createFirstAssignment(page)
    await expect(page.getByTestId('draft-textarea')).toBeVisible()

    await openTrackMenu(page)
    await expect(page.locator('[data-testid^="track-option-"]')).toHaveCount(1)
    await expect(page.locator('[data-testid^="track-option-"]').first()).toBeVisible()
  } finally {
    await closeContext(context)
  }
})

test('guards unsaved drafts during track switching and switches when confirmed', async ({ browser }, testInfo) => {
  const userSlug = uniqueUserSlug('e2e-switch', testInfo.workerIndex)
  const { context, page } = await newUserPage(browser, userSlug)
  try {
    await page.goto('/onboarding')
    await createTrack(page, 'switch-one')
    await createFirstAssignment(page)

    await page.goto('/onboarding?mode=create')
    await createTrack(page, 'switch-two', { writingTypeIndex: 2 })
    await expect(page).toHaveURL(/\/(?:new-assignment)?$/)
    await openTrackMenu(page)
    await expect(page.locator('[data-testid^="track-option-"]')).toHaveCount(2)
    await page.keyboard.press('Escape')

    const activeTrackBefore = ((await page.getByTestId('active-track-button').textContent()) ?? '').trim()
    await openTrackMenu(page)
    await switchToOtherTrack(page)
    await expect(page.getByTestId('active-track-button')).not.toHaveText(activeTrackBefore)
    await page.goto('/')
    await expect(page.getByTestId('draft-textarea')).toBeVisible()

    const draft = page.getByTestId('draft-textarea')
    await draft.fill('This is an unsaved draft that should stay put.')
    const activeTrackWithDraft = ((await page.getByTestId('active-track-button').textContent()) ?? '').trim()

    page.once('dialog', (dialog) => dialog.dismiss())
    await openTrackMenu(page)
    await switchToOtherTrack(page, { requireChange: false })
    await expect(draft).toHaveValue('This is an unsaved draft that should stay put.')
    await expect(page.getByTestId('draft-textarea')).toBeVisible()
    await expect(page.getByTestId('active-track-button')).toHaveText(activeTrackWithDraft)

    page.once('dialog', (dialog) => dialog.accept())
    await openTrackMenu(page)
    await switchToOtherTrack(page)
    await expect(page.getByTestId('active-track-button')).not.toHaveText(activeTrackWithDraft)
    await expect(page).toHaveURL(/\/$/)
  } finally {
    await closeContext(context)
  }
})

test('archives the current track and refreshes the sidebar track list', async ({ browser }, testInfo) => {
  const userSlug = uniqueUserSlug('e2e-archive', testInfo.workerIndex)
  const { context, page } = await newUserPage(browser, userSlug)
  try {
    await page.goto('/onboarding')
    await createTrack(page, 'archive-one')
    await createFirstAssignment(page)

    await page.goto('/onboarding?mode=create')
    await createTrack(page, 'archive-two', { writingTypeIndex: 2 })
    await expect(page).toHaveURL(/\/$/)
    await openTrackMenu(page)
    await expect(page.locator('[data-testid^="track-option-"]')).toHaveCount(2)
    await page.keyboard.press('Escape')

    await page.goto('/onboarding?mode=edit')
    if (!(await page.getByTestId('archive-track-button').isVisible().catch(() => false))) {
      await page.goto('/')
      await openTrackMenu(page)
      await switchToOtherTrack(page)
      await page.goto('/onboarding?mode=edit')
    }
    await expect(page.getByTestId('archive-track-button')).toBeVisible()
    page.once('dialog', (dialog) => dialog.accept())
    await page.getByTestId('archive-track-button').click()

    await expect(page).toHaveURL(/\/$/)
    await openTrackMenu(page)
    await expect(page.locator('[data-testid^="track-option-"]')).toHaveCount(1)
    await expect(page.locator('[data-testid^="track-option-"]').first()).toBeVisible()
  } finally {
    await closeContext(context)
  }
})
