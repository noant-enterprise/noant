import { test, expect } from '@playwright/test'

// These tests require an authenticated session with the backend running.
// Without auth, ProtectedRoute redirects to /login.

test.describe('Dashboard (requires auth)', () => {
  test.skip(
    ({ browserName }) => true,
    'Requires backend running and authenticated session'
  )

  test('loads dashboard with sidebar', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page.locator('nav')).toBeVisible()
  })

  test('sidebar has navigation links', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page.getByRole('link', { name: 'Overview' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Conversations' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Insights' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Settings' })).toBeVisible()
  })

  test('sidebar navigates to chats page', async ({ page }) => {
    await page.goto('/dashboard')
    await page.getByRole('link', { name: 'Conversations' }).click()
    await expect(page).toHaveURL(/\/chats/)
  })

  test('sidebar navigates to settings page', async ({ page }) => {
    await page.goto('/dashboard')
    await page.getByRole('link', { name: 'Settings' }).click()
    await expect(page).toHaveURL(/\/settings/)
  })

  test('sidebar navigates to channels page', async ({ page }) => {
    await page.goto('/dashboard')
    await page.getByRole('link', { name: 'Your channels' }).click()
    await expect(page).toHaveURL(/\/channels/)
  })

  test('dashboard overview shows stat cards', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page.getByText('Conversations today')).toBeVisible()
    await expect(page.getByText('Resolved auto')).toBeVisible()
  })

  test('dashboard overview shows quick actions', async ({ page }) => {
    await page.goto('/dashboard')
    await expect(page.getByRole('link', { name: 'Inbox' })).toBeVisible()
    await expect(page.getByRole('link', { name: 'Channels' })).toBeVisible()
  })
})

test.describe('Dashboard (unauthenticated redirect)', () => {
  test('redirects to login when not authenticated', async ({ page }) => {
    // Clear any existing auth state
    await page.goto('/login')
    await page.evaluate(() => {
      localStorage.removeItem('noant_token')
      localStorage.removeItem('noant_refresh')
    })

    // Now try to access a protected route
    await page.goto('/dashboard')
    await expect(page).toHaveURL(/\/login/)
  })
})
