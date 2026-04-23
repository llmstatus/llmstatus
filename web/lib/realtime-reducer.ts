export interface ProviderStatus {
  id: string
  status: 'operational' | 'degraded' | 'down'
  lastUpdated: number
}

export interface RealtimeState {
  providers: Record<string, ProviderStatus>
  connectionStatus: 'connected' | 'disconnected' | 'reconnecting'
  subscriptions: Set<string>
}

export type RealtimeAction =
  | { type: 'CONNECTION_STATUS'; status: RealtimeState['connectionStatus'] }
  | { type: 'PROVIDER_UPDATE'; provider: ProviderStatus }
  | { type: 'SUBSCRIBE'; topic: string }
  | { type: 'UNSUBSCRIBE'; topic: string }

export const initialState: RealtimeState = {
  providers: {},
  connectionStatus: 'disconnected',
  subscriptions: new Set(),
}

export function realtimeReducer(
  state: RealtimeState,
  action: RealtimeAction
): RealtimeState {
  switch (action.type) {
    case 'CONNECTION_STATUS':
      return { ...state, connectionStatus: action.status }

    case 'PROVIDER_UPDATE':
      return {
        ...state,
        providers: {
          ...state.providers,
          [action.provider.id]: action.provider,
        },
      }

    case 'SUBSCRIBE':
      return {
        ...state,
        subscriptions: new Set([...state.subscriptions, action.topic]),
      }

    case 'UNSUBSCRIBE':
      const newSubscriptions = new Set(state.subscriptions)
      newSubscriptions.delete(action.topic)
      return {
        ...state,
        subscriptions: newSubscriptions,
      }

    default:
      return state
  }
}