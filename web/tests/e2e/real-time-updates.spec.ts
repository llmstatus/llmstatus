import { test, expect } from '@playwright/test'

test.describe('Real-time Updates', () => {
  test('should establish WebSocket connection', async ({ page }) => {
    // Mock WebSocket for testing
    await page.addInitScript(() => {
      class MockWebSocket {
        constructor(url: string) {
          setTimeout(() => {
            this.onopen?.({ type: 'open' })
            this.onmessage?.({
              type: 'message',
              data: JSON.stringify({ type: 'connected' })
            })
          }, 100)
        }

        send(data: string) {
          // Mock send
        }

        close() {
          this.onclose?.({ type: 'close' })
        }

        onopen: ((event: any) => void) | null = null
        onmessage: ((event: any) => void) | null = null
        onclose: ((event: any) => void) | null = null
        onerror: ((event: any) => void) | null = null
      }

      (window as any).WebSocket = MockWebSocket
    })

    await page.goto('/')

    // Verify connection status indicator
    await expect(page.locator('[data-testid="connection-status"]')).toHaveText('Connected')
  })

  test('should receive and display real-time status updates', async ({ page }) => {
    await page.goto('/')

    // Wait for initial load
    await expect(page.locator('[data-testid="provider-card"]').first()).toBeVisible()

    // Simulate real-time update
    await page.evaluate(() => {
      window.dispatchEvent(new CustomEvent('websocket-message', {
        detail: {
          type: 'provider:update',
          data: { id: 'openai', status: 'degraded' }
        }
      }))
    })

    // Verify UI updated
    await expect(page.locator('[data-testid="provider-openai"] [data-testid="status-indicator"]'))
      .toHaveAttribute('data-status', 'degraded')
  })

  test('should handle connection loss and reconnection', async ({ page }) => {
    await page.goto('/')

    // Simulate connection loss
    await page.evaluate(() => {
      window.dispatchEvent(new Event('offline'))
    })

    await expect(page.locator('[data-testid="connection-status"]')).toHaveText('Disconnected')

    // Simulate reconnection
    await page.evaluate(() => {
      window.dispatchEvent(new Event('online'))
    })

    await expect(page.locator('[data-testid="connection-status"]')).toHaveText('Connected')
  })

  test('should display optimistic updates', async ({ page }) => {
    await page.goto('/providers/openai')

    // Trigger an action that causes optimistic update
    await page.getByRole('button', { name: /refresh/i }).click()

    // Verify optimistic state
    await expect(page.locator('[data-testid="loading-indicator"]')).toBeVisible()

    // Wait for server confirmation
    await expect(page.locator('[data-testid="loading-indicator"]')).not.toBeVisible()
  })
})