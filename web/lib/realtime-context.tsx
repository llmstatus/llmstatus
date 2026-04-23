'use client'

import React, { createContext, useContext, useReducer, useEffect, useRef } from 'react'
import { WebSocketClient } from './websocket'
import { realtimeReducer, initialState, RealtimeState, RealtimeAction } from './realtime-reducer'
import type { WebSocketMessage } from './types'

interface RealtimeContextType {
  providers: RealtimeState['providers']
  connectionStatus: RealtimeState['connectionStatus']
  subscriptions: RealtimeState['subscriptions']
  subscribe: (topic: string) => void
  unsubscribe: (topic: string) => void
}

const RealtimeContext = createContext<RealtimeContextType | null>(null)

export function RealtimeProvider({ children }: { children: React.ReactNode }) {
  const [state, dispatch] = useReducer(realtimeReducer, initialState)
  const wsClient = useRef<WebSocketClient | null>(null)

  useEffect(() => {
    const client = new WebSocketClient('ws://localhost:8081/ws')
    wsClient.current = client

    client.onMessage((message: WebSocketMessage) => {
      if (message.type === 'status_update') {
        dispatch({
          type: 'PROVIDER_UPDATE',
          provider: {
            id: message.provider_id,
            status: message.status,
            lastUpdated: message.timestamp || Date.now(),
          },
        })
      }
    })

    client.connect()
      .then(() => {
        dispatch({ type: 'CONNECTION_STATUS', status: 'connected' })
      })
      .catch(() => {
        dispatch({ type: 'CONNECTION_STATUS', status: 'disconnected' })
      })

    return () => {
      client.disconnect()
    }
  }, [])

  const subscribe = (topic: string) => {
    wsClient.current?.subscribe(topic)
    dispatch({ type: 'SUBSCRIBE', topic })
  }

  const unsubscribe = (topic: string) => {
    wsClient.current?.unsubscribe(topic)
    dispatch({ type: 'UNSUBSCRIBE', topic })
  }

  const contextValue: RealtimeContextType = {
    providers: state.providers,
    connectionStatus: state.connectionStatus,
    subscriptions: state.subscriptions,
    subscribe,
    unsubscribe,
  }

  return (
    <RealtimeContext.Provider value={contextValue}>
      {children}
    </RealtimeContext.Provider>
  )
}

export function useRealtime(): RealtimeContextType {
  const context = useContext(RealtimeContext)
  if (!context) {
    throw new Error('useRealtime must be used within a RealtimeProvider')
  }
  return context
}