import { expect, test } from '@playwright/test'
import { closeContext, createFirstAssignment, createTrack, newUserPage, openTrackMenu } from './helpers'

function uniqueUserSlug(prefix: string, workerIndex: number) {
  return `${prefix}-${workerIndex}-${Date.now()}`
}

test('redirects incomplete users from gated routes into onboarding', async ({ browser }, testInfo) => {
  const userSlug = uniqueUserSlug('e2e-gate', testInfo.workerIndex)
  const { context, page } = await newUserPage(browser, userSlug)
  try {
    await page.goto('/progress')
    await expect(page).toHaveURL(/\/onboarding/)
    await expect(page.getByText('Next, create your first track')).toBeVisible()
  } finally {
    await closeContext(context)
  }
})

test('creates the first track and enters the first assignment flow', async ({ browser }, testInfo) => {
  const userSlug = uniqueUserSlug('e2e-first', testInfo.workerIndex)
  const firstTrackSlug = `${userSlug}-track`
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
    await expect(page.getByTestId(`track-option-${firstTrackSlug}`)).toBeVisible()
  } finally {
    await closeContext(context)
  }
})

test('guards unsaved drafts during track switching and switches when confirmed', async ({ browser }, testInfo) => {
  const userSlug = uniqueUserSlug('e2e-switch', testInfo.workerIndex)
  const firstTrackSlug = `${userSlug}-track`
  const secondTrackSlug = `${firstTrackSlug}-2`
  const { context, page } = await newUserPage(browser, userSlug)
  try {
    await page.goto('/onboarding')
    await createTrack(page, 'switch-one')
    await createFirstAssignment(page)

    await page.goto('/onboarding?mode=create')
    await createTrack(page, 'switch-two')
    await expect(page).toHaveURL(/\/new-assignment/)
    await expect(page.getByText('Finish by generating your first assignment')).toBeVisible()

    await openTrackMenu(page)
    await page.getByTestId(`track-option-${firstTrackSlug}`).click()
    await expect(page.getByTestId('active-track-button')).not.toContainText('No assignments yet')
    await page.goto('/')
    await expect(page.getByTestId('draft-textarea')).toBeVisible()

    const draft = page.getByTestId('draft-textarea')
    await draft.fill('This is an unsaved draft that should stay put.')

    page.once('dialog', (dialog) => dialog.dismiss())
    await openTrackMenu(page)
    await page.getByTestId(`track-option-${secondTrackSlug}`).click()
    await expect(draft).toHaveValue('This is an unsaved draft that should stay put.')
    await expect(page.getByTestId('draft-textarea')).toBeVisible()

    page.once('dialog', (dialog) => dialog.accept())
    await openTrackMenu(page)
    await page.getByTestId(`track-option-${secondTrackSlug}`).click()
    await expect(page.getByTestId('active-track-button')).toContainText('No assignments yet')
    await expect(page).toHaveURL(/\/new-assignment/)
    await expect(page.getByText('Finish by generating your first assignment')).toBeVisible()
  } finally {
    await closeContext(context)
  }
})

test('archives the current track and refreshes the sidebar track list', async ({ browser }, testInfo) => {
  const userSlug = uniqueUserSlug('e2e-archive', testInfo.workerIndex)
  const firstTrackSlug = `${userSlug}-track`
  const secondTrackSlug = `${firstTrackSlug}-2`
  const { context, page } = await newUserPage(browser, userSlug)
  try {
    await page.goto('/onboarding')
    await createTrack(page, 'archive-one')
    await createFirstAssignment(page)

    await page.goto('/onboarding?mode=create')
    await createTrack(page, 'archive-two')
    await expect(page).toHaveURL(/\/new-assignment/)
    await createFirstAssignment(page)

    await page.goto('/onboarding?mode=edit')
    page.once('dialog', (dialog) => dialog.accept())
    await page.getByTestId('archive-track-button').click()

    await expect(page).toHaveURL(/\/$/)
    await openTrackMenu(page)
    await expect(page.locator('[data-testid^="track-option-"]')).toHaveCount(1)
    await expect(page.getByTestId(`track-option-${firstTrackSlug}`)).toBeVisible()
    await expect(page.getByTestId(`track-option-${secondTrackSlug}`)).toHaveCount(0)
  } finally {
    await closeContext(context)
  }
})
