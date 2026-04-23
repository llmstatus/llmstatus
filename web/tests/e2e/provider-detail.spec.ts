import { test, expect } from '@playwright/test'

test.describe('Provider Detail Page', () => {
  test('should display provider information', async ({ page }) => {
    await page.goto('/providers/openai')

    await expect(page.getByRole('heading', { name: /openai/i })).toBeVisible()
    await expect(page.locator('[data-testid="status-indicator"]')).toBeVisible()
    await expect(page.locator('[data-testid="metrics-section"]')).toBeVisible()
  })

  test('should display time series chart', async ({ page }) => {
    await page.goto('/providers/openai')

    await expect(page.locator('[data-testid="time-series-chart"]')).toBeVisible()

    // Test chart interactions
    const chart = page.locator('[data-testid="time-series-chart"] svg')
    await chart.hover()
    await expect(page.locator('[data-testid="chart-tooltip"]')).toBeVisible()
  })

  test('should handle mobile viewport', async ({ page }) => {
    await page.setViewportSize({ width: 375, height: 667 })
    await page.goto('/providers/openai')

    await expect(page.getByRole('heading', { name: /openai/i })).toBeVisible()

    // Test mobile-specific interactions
    const chart = page.locator('[data-testid="time-series-chart"]')
    await chart.tap()
  })

  test('should navigate back to homepage', async ({ page }) => {
    await page.goto('/providers/openai')

    await page.getByRole('link', { name: /back to home/i }).click()
    await expect(page).toHaveURL('/')
  })
})