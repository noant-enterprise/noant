import { test, expect } from '@playwright/test'

test.describe('Login page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/login')
    await page.waitForSelector('h1', { timeout: 10000 })
  })

  test('renders email and password fields', async ({ page }) => {
    await expect(page.getByPlaceholder('you@company.com')).toBeVisible()
    await expect(page.getByPlaceholder('********')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible()
  })

  test('shows heading and subtitle', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Welcome back' })).toBeVisible()
    await expect(page.getByText('Sign in to manage your AI customer support')).toBeVisible()
  })

  test('has link to signup page', async ({ page }) => {
    const signupLink = page.getByRole('link', { name: 'Get started' })
    await expect(signupLink).toBeVisible()
    await signupLink.click()
    await expect(page).toHaveURL(/\/signup/)
  })

  test('has link to forgot password', async ({ page }) => {
    const forgotLink = page.getByRole('link', { name: 'Forgot?' })
    await expect(forgotLink).toBeVisible()
    await forgotLink.click()
    await expect(page).toHaveURL(/\/forgot-password/)
  })

  test('shows toast validation error when submitting empty fields', async ({ page }) => {
    await page.getByRole('button', { name: 'Sign in' }).click()
    await expect(page.locator('[class*="fixed"]').filter({ hasText: 'Please fill in all fields' })).toBeVisible()
  })

  test('password visibility toggle works', async ({ page }) => {
    const passwordInput = page.getByPlaceholder('********')
    await expect(passwordInput).toHaveAttribute('type', 'password')

    const toggle = page.getByRole('button', { name: /show password/i })
    await toggle.click()
    await expect(passwordInput).toHaveAttribute('type', 'text')

    await page.getByRole('button', { name: /hide password/i }).click()
    await expect(passwordInput).toHaveAttribute('type', 'password')
  })

  test('can type into email and password fields', async ({ page }) => {
    await page.getByPlaceholder('you@company.com').fill('test@example.com')
    await expect(page.getByPlaceholder('you@company.com')).toHaveValue('test@example.com')

    await page.getByPlaceholder('********').fill('mypassword')
    await expect(page.getByPlaceholder('********')).toHaveValue('mypassword')
  })
})

test.describe('Signup page', () => {
  test.beforeEach(async ({ page }) => {
    await page.goto('/signup')
    await page.waitForSelector('h1', { timeout: 10000 })
  })

  test('renders all form fields', async ({ page }) => {
    await expect(page.getByPlaceholder('John')).toBeVisible()
    await expect(page.getByPlaceholder('Doe')).toBeVisible()
    await expect(page.getByPlaceholder('you@company.com')).toBeVisible()
    await expect(page.getByPlaceholder('Min 8 chars')).toBeVisible()
    await expect(page.getByRole('button', { name: 'Create account' })).toBeVisible()
  })

  test('shows heading and subtitle', async ({ page }) => {
    await expect(page.getByRole('heading', { name: 'Get started' })).toBeVisible()
    await expect(page.getByText('Create your AI customer support agent')).toBeVisible()
  })

  test('has link to login page', async ({ page }) => {
    const loginLink = page.getByRole('link', { name: 'Sign in' })
    await expect(loginLink).toBeVisible()
    await loginLink.click()
    await expect(page).toHaveURL(/\/login/)
  })

  test('shows toast validation error when submitting empty fields', async ({ page }) => {
    await page.getByRole('button', { name: 'Create account' }).click()
    await expect(page.locator('[class*="fixed"]').filter({ hasText: 'Please fill in all required fields' })).toBeVisible()
  })

  test('can fill in all required fields', async ({ page }) => {
    await page.getByPlaceholder('John').fill('John')
    await page.getByPlaceholder('Doe').fill('Doe')
    await page.getByPlaceholder('you@company.com').fill('john@example.com')
    await page.getByPlaceholder('Min 8 chars').fill('securepass123')

    await expect(page.getByPlaceholder('John')).toHaveValue('John')
    await expect(page.getByPlaceholder('Doe')).toHaveValue('Doe')
    await expect(page.getByPlaceholder('you@company.com')).toHaveValue('john@example.com')
    await expect(page.getByPlaceholder('Min 8 chars')).toHaveValue('securepass123')
  })

  test('company field is optional', async ({ page }) => {
    const companyInput = page.locator('input[placeholder="Acme Inc"]')
    await expect(companyInput).toBeVisible()
  })
})

test.describe('Auth layout', () => {
  test('login page shows branding on desktop viewport', async ({ page }) => {
    await page.setViewportSize({ width: 1280, height: 720 })
    await page.goto('/login')
    await page.waitForSelector('h1', { timeout: 10000 })
    await expect(page.locator('text=noant').first()).toBeVisible()
  })

  test('login page shows mobile logo', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/login')
    await page.waitForSelector('h1', { timeout: 10000 })
    await expect(page.locator('.lg\\:hidden >> text=noant').first()).toBeVisible()
  })
})
