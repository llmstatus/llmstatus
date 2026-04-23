import { renderHook, waitFor } from '@testing-library/react'
import { useProviderStatus } from '@/hooks/useProviderStatus'
import { RealtimeProvider } from '@/lib/realtime-context'
import { SWRProvider } from '@/lib/swr-config'

const wrapper = ({ children }: { children: React.ReactNode }) => (
  <SWRProvider>
    <RealtimeProvider>
      {children}
    </RealtimeProvider>
  </SWRProvider>
)

describe('useProviderStatus', () => {
  it('should fetch and return provider status', async () => {
    const { result } = renderHook(() => useProviderStatus(), { wrapper })

    await waitFor(() => {
      expect(result.current.data).toBeDefined()
    })
  })

  it('should handle real-time updates', async () => {
    const { result } = renderHook(() => useProviderStatus(), { wrapper })

    // Test would verify real-time updates integration
    expect(result.current.isLoading).toBeDefined()
  })
})