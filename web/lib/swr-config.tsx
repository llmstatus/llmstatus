'use client'

import { SWRConfig } from 'swr'
import { ReactNode } from 'react'
import { fetcher } from './api-client'
import { useRealtime } from './realtime-context'

interface SWRProviderProps {
  children: ReactNode
}

export function SWRProvider({ children }: SWRProviderProps) {
  return (
    <SWRConfig
      value={{
        fetcher,
        revalidateOnFocus: false,
        revalidateOnReconnect: true,
        refreshInterval: 30000, // 30 second fallback polling
        dedupingInterval: 5000,
        errorRetryCount: 3,
        errorRetryInterval: 1000,
        onError: (error) => {
          console.error('SWR Error:', error)
        },
        use: [realtimeMiddleware],
      }}
    >
      {children}
    </SWRConfig>
  )
}

// SWR middleware to integrate with real-time updates
// eslint-disable-next-line @typescript-eslint/no-explicit-any
function realtimeMiddleware(useSWRNext: any) {
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  return (key: string, fetcher: any, config: any) => {
    const swr = useSWRNext(key, fetcher, config)

    // This would be enhanced to listen to real-time updates
    // and trigger revalidation when relevant data changes
    return swr
  }
}