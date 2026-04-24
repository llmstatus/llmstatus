import { WebSocketClient } from '@/lib/websocket'

type EventCallback = (data?: Event | MessageEvent) => void

describe('WebSocketClient', () => {
  let mockWS: {
    send: jest.Mock
    close: jest.Mock
    addEventListener: jest.Mock
    removeEventListener: jest.Mock
    readyState: number
    _triggerEvent: (event: string, data?: Event | MessageEvent) => void
  }
  let originalWebSocket: typeof global.WebSocket

  beforeEach(() => {
    const eventListeners: Record<string, EventCallback[]> = {}

    mockWS = {
      send: jest.fn(),
      close: jest.fn(),
      addEventListener: jest.fn((event: string, callback: EventCallback) => {
        if (!eventListeners[event]) {
          eventListeners[event] = []
        }
        eventListeners[event].push(callback)

        // Auto-trigger 'open' event for connection simulation
        if (event === 'open') {
          setTimeout(() => callback(), 0)
        }
      }),
      removeEventListener: jest.fn(),
      readyState: 1, // OPEN
      _triggerEvent: (event: string, data?: Event | MessageEvent) => {
        if (eventListeners[event]) {
          eventListeners[event].forEach(callback => callback(data))
        }
      }
    }

    originalWebSocket = global.WebSocket
    const MockWebSocket = jest.fn(() => mockWS) as unknown as typeof WebSocket & {
      OPEN: number; CLOSED: number; CONNECTING: number; CLOSING: number
    }
    MockWebSocket.OPEN = 1
    MockWebSocket.CLOSED = 3
    MockWebSocket.CONNECTING = 0
    MockWebSocket.CLOSING = 2
    global.WebSocket = MockWebSocket
  })

  afterEach(() => {
    global.WebSocket = originalWebSocket
    jest.clearAllMocks()
  })

  it('should establish connection and handle messages', async () => {
    const client = new WebSocketClient('ws://localhost:8081/ws')
    const messageHandler = jest.fn()

    client.onMessage(messageHandler)
    await client.connect()

    expect(global.WebSocket).toHaveBeenCalledWith('ws://localhost:8081/ws')
  })

  it('should handle incoming messages', async () => {
    const client = new WebSocketClient('ws://localhost:8081/ws')
    const messageHandler = jest.fn()

    client.onMessage(messageHandler)
    await client.connect()

    // Simulate incoming message
    const messageEvent = new MessageEvent('message', {
      data: JSON.stringify({ type: 'status_update', provider_id: 'openai', status: 'operational' })
    })

    // Trigger the message event using our mock helper
    mockWS._triggerEvent('message', messageEvent)

    expect(messageHandler).toHaveBeenCalledWith({
      type: 'status_update',
      provider_id: 'openai',
      status: 'operational'
    })
  })

  it('should subscribe to topic', async () => {
    const client = new WebSocketClient('ws://localhost:8081/ws')
    await client.connect()

    client.subscribe('provider:openai')

    expect(mockWS.send).toHaveBeenCalledWith(
      JSON.stringify({ type: 'subscribe', topic: 'provider:openai' }),
    )
  })

  it('should unsubscribe from topic', async () => {
    const client = new WebSocketClient('ws://localhost:8081/ws')
    await client.connect()

    client.unsubscribe('provider:openai')

    expect(mockWS.send).toHaveBeenCalledWith(
      JSON.stringify({ type: 'unsubscribe', topic: 'provider:openai' }),
    )
  })

  it('should handle reconnection with exponential backoff', async () => {
    jest.useFakeTimers()

    const client = new WebSocketClient('ws://localhost:8081/ws')

    // Connect first
    const connectPromise = client.connect()
    jest.runAllTimers() // Run the setTimeout for the open event
    await connectPromise

    // Trigger close event to start reconnection
    mockWS._triggerEvent('close', new Event('close'))

    // First reconnection attempt after 1 second
    jest.advanceTimersByTime(1000)
    jest.runAllTimers() // Run the setTimeout for the reconnection open event

    expect(global.WebSocket).toHaveBeenCalledTimes(2)

    jest.useRealTimers()
  })

  it('should disconnect gracefully', async () => {
    const client = new WebSocketClient('ws://localhost:8081/ws')
    await client.connect()

    client.disconnect()

    expect(mockWS.close).toHaveBeenCalled()
  })
})
