import { test, expect } from '@playwright/test'

test.describe('Homepage', () => {
  test('should load and display provider grid', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByRole('heading', { name: 'LLM Status Monitor' })).toBeVisible()
    await expect(page.getByText('Real-time monitoring of AI API providers')).toBeVisible()

    // Wait for provider grid to load
    await expect(page.locator('[data-testid="provider-grid"]')).toBeVisible()
  })

  test('should display metrics cards', async ({ page }) => {
    await page.goto('/')

    await expect(page.getByText('Overall Uptime')).toBeVisible()
    await expect(page.getByText('Avg Response Time')).toBeVisible()
    await expect(page.getByText('Active Incidents')).toBeVisible()
  })

  test('should be accessible', async ({ page }) => {
    await page.goto('/')

    // Test keyboard navigation
    await page.keyboard.press('Tab')
    await expect(page.locator(':focus')).toBeVisible()

    // Test skip link
    await page.keyboard.press('Tab')
    await expect(page.getByText('Skip to main content')).toBeFocused()
  })

  test('should handle loading states', async ({ page }) => {
    await page.goto('/')

    // Should show loading skeleton initially
    await expect(page.locator('[role="status"][aria-label="loading"]')).toBeVisible()

    // Should show actual data after loading
    await expect(page.locator('[data-testid="provider-card"]').first()).toBeVisible()
  })
})