import { render, screen, act } from '@testing-library/react'
import { RealtimeProvider, useRealtime } from '@/lib/realtime-context'

const TestComponent = () => {
  const { providers, connectionStatus } = useRealtime()
  return (
    <div>
      <div data-testid="status">{connectionStatus}</div>
      <div data-testid="count">{Object.keys(providers).length}</div>
    </div>
  )
}

describe('RealtimeProvider', () => {
  it('should provide initial state', () => {
    render(
      <RealtimeProvider>
        <TestComponent />
      </RealtimeProvider>
    )

    expect(screen.getByTestId('status')).toHaveTextContent('disconnected')
    expect(screen.getByTestId('count')).toHaveTextContent('0')
  })
})