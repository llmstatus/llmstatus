import type { WebSocketMessage } from './types'

export class WebSocketClient {
  private ws: WebSocket | null = null
  private url: string
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private reconnectDelay = 1000
  private messageHandlers: ((data: WebSocketMessage) => void)[] = []
  private reconnectTimeout: NodeJS.Timeout | null = null

  constructor(url: string) {
    this.url = url
  }

  async connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      try {
        this.ws = new WebSocket(this.url)

        this.ws.addEventListener('open', () => {
          this.reconnectAttempts = 0
          resolve()
        })

        this.ws.addEventListener('message', (event: MessageEvent) => {
          try {
            const data = JSON.parse(event.data) as WebSocketMessage
            this.messageHandlers.forEach((handler) => handler(data))
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error)
          }
        })

        this.ws.addEventListener('close', () => {
          this.handleReconnect()
        })

        this.ws.addEventListener('error', (event: Event) => {
          console.error('WebSocket error:', event)
          reject(new Error('WebSocket connection failed'))
        })
      } catch (error) {
        reject(error)
      }
    })
  }

  private handleReconnect(): void {
    if (this.reconnectAttempts < this.maxReconnectAttempts) {
      this.reconnectAttempts++
      const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
      const cappedDelay = Math.min(delay, 30000) // Max 30 seconds

      this.reconnectTimeout = setTimeout(() => {
        this.connect().catch((error) => {
          console.error('Reconnection attempt failed:', error)
        })
      }, cappedDelay)
    }
  }

  subscribe(topic: string): void {
    this.send({ type: 'subscribe', topic })
  }

  unsubscribe(topic: string): void {
    this.send({ type: 'unsubscribe', topic })
  }

  onMessage(handler: (data: WebSocketMessage) => void): void {
    this.messageHandlers.push(handler)
  }

  private send(data: WebSocketMessage): void {
    if (this.ws && this.ws.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(data))
    }
  }

  disconnect(): void {
    if (this.reconnectTimeout) {
      clearTimeout(this.reconnectTimeout)
      this.reconnectTimeout = null
    }
    this.ws?.close()
  }
}