import { render, screen, waitFor } from '@testing-library/react'
import useSWR from 'swr'
import { SWRProvider } from '@/lib/swr-config'

const TestComponent = () => {
  const { data, error } = useSWR('/api/providers', () =>
    Promise.resolve([{ id: 'openai', status: 'operational' }])
  )

  if (error) return <div>Error</div>
  if (!data) return <div>Loading</div>
  return <div data-testid="data">{data[0].id}</div>
}

describe('SWRProvider', () => {
  it('should provide SWR configuration', async () => {
    render(
      <SWRProvider>
        <TestComponent />
      </SWRProvider>
    )

    await waitFor(() => {
      expect(screen.getByTestId('data')).toHaveTextContent('openai')
    })
  })
})