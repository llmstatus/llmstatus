import useSWR from 'swr'
import { useRealtime } from '@/lib/realtime-context'
import { api, type Provider } from '@/lib/api-client'
import { useEffect } from 'react'

export function useProviderStatus() {
  const { data, error, mutate } = useSWR('/api/providers', api.providers)
  const { providers: realtimeProviders, subscribe, unsubscribe } = useRealtime()

  // Subscribe to real-time provider updates
  useEffect(() => {
    subscribe('providers:all')
    return () => unsubscribe('providers:all')
  }, [subscribe, unsubscribe])

  // Merge SWR data with real-time updates
  const mergedData = data?.map(provider => {
    const realtimeUpdate = realtimeProviders[provider.id]
    if (realtimeUpdate && realtimeUpdate.lastUpdated > new Date(provider.last_updated).getTime()) {
      return {
        ...provider,
        status: realtimeUpdate.status,
        last_updated: new Date(realtimeUpdate.lastUpdated).toISOString(),
      }
    }
    return provider
  })

  return {
    data: mergedData,
    error,
    isLoading: !error && !data,
    mutate,
  }
}

export function useProviderDetail(id: string) {
  const { data, error, mutate } = useSWR(`/api/providers/${id}`, () => api.provider(id))
  const { providers: realtimeProviders, subscribe, unsubscribe } = useRealtime()

  // Subscribe to specific provider updates
  useEffect(() => {
    subscribe(`provider:${id}`)
    return () => unsubscribe(`provider:${id}`)
  }, [id, subscribe, unsubscribe])

  // Merge SWR data with real-time updates
  const mergedData = data && realtimeProviders[id] &&
    realtimeProviders[id].lastUpdated > new Date(data.last_updated).getTime()
    ? {
        ...data,
        status: realtimeProviders[id].status,
        last_updated: new Date(realtimeProviders[id].lastUpdated).toISOString(),
      }
    : data

  return {
    data: mergedData,
    error,
    isLoading: !error && !data,
    mutate,
  }
}