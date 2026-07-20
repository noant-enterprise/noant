import { test, expect } from '@playwright/test'

// These tests require an authenticated session with the backend running.

test.describe('Chats page (requires auth)', () => {
  test.skip(
    ({ browserName }) => true,
    'Requires backend running and authenticated session'
  )

  test('loads chats page', async ({ page }) => {
    await page.goto('/chats')
    // The conversation list panel should be visible
    await expect(page.locator('text=Messages').first()).toBeVisible()
  })

  test('shows AI assistant button', async ({ page }) => {
    await page.goto('/chats')
    await expect(page.getByText('Noant AI')).toBeVisible()
    await expect(page.getByText('Chat with your assistant')).toBeVisible()
  })

  test('shows conversation list or empty state', async ({ page }) => {
    await page.goto('/chats')
    // Either conversations are shown or the empty state
    const hasConversations = await page.locator('button[class*="w-full flex items-center gap-3"]').count()
    const hasEmptyState = await page.getByText('No conversations found').isVisible()
    expect(hasConversations > 0 || hasEmptyState).toBeTruthy()
  })

  test('has search input', async ({ page }) => {
    await page.goto('/chats')
    await expect(page.getByPlaceholder('Search conversations...')).toBeVisible()
  })

  test('has filter buttons', async ({ page }) => {
    await page.goto('/chats')
    await expect(page.getByRole('button', { name: 'all' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'active' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'escalated' })).toBeVisible()
    await expect(page.getByRole('button', { name: 'resolved' })).toBeVisible()
  })

  test('shows empty chat area when no conversation selected', async ({ page }) => {
    await page.goto('/chats')
    // When no conversation is selected, the right panel shows AI Assistant prompt
    await expect(page.getByText('Click the Noant AI button in the sidebar to start chatting.')).toBeVisible()
  })
})

test.describe('Chats page (unauthenticated redirect)', () => {
  test('redirects to login when not authenticated', async ({ page }) => {
    await page.goto('/login')
    await page.evaluate(() => {
      localStorage.removeItem('noant_token')
      localStorage.removeItem('noant_refresh')
    })

    await page.goto('/chats')
    await expect(page).toHaveURL(/\/login/)
  })
})
