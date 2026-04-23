export interface StatusUpdate {
  type: string
  provider_id: string
  status: 'operational' | 'degraded' | 'down'
  timestamp: number
}

export interface SubscriptionMessage {
  type: 'subscribe' | 'unsubscribe'
  topic: string
}

export interface WebSocketMessage {
  type: string
  [key: string]: any
}