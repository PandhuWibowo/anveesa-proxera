export type WSEventType = 'metrics' | 'alert' | 'status_change' | 'pong' | 'connected'

export interface WSMessage<T = unknown> {
  type: WSEventType
  payload: T
}

export type WSHandler<T = unknown> = (payload: T) => void

export interface ProxeraWSOptions {
  url?: string
  reconnectDelayMs?: number
  maxReconnectAttempts?: number
}

export class ProxeraWebSocket {
  private ws: WebSocket | null = null
  private handlers = new Map<WSEventType, WSHandler[]>()
  private subscribedServerIds = new Set<string>()
  private reconnectAttempts = 0
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private destroyed = false

  private readonly url: string
  private readonly reconnectDelayMs: number
  private readonly maxReconnectAttempts: number

  constructor(options: ProxeraWSOptions = {}) {
    const wsBase =
      options.url ??
      (import.meta.env.VITE_WS_URL as string | undefined) ??
      (location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host

    this.url = wsBase.replace(/\/+$/, '') + '/ws'
    this.reconnectDelayMs = options.reconnectDelayMs ?? 3000
    this.maxReconnectAttempts = options.maxReconnectAttempts ?? 10
    this.connect()
  }

  private connect() {
    if (this.destroyed) return

    this.ws = new WebSocket(this.url)

    this.ws.onopen = () => {
      this.reconnectAttempts = 0
      this.emit('connected', {})

      // Re-subscribe after reconnect
      if (this.subscribedServerIds.size > 0) {
        this.subscribe([...this.subscribedServerIds])
      }
    }

    this.ws.onmessage = (event: MessageEvent) => {
      try {
        const msg: WSMessage = JSON.parse(event.data)
        this.emit(msg.type, msg.payload)
      } catch {
        // ignore malformed messages
      }
    }

    this.ws.onclose = () => {
      if (!this.destroyed) this.scheduleReconnect()
    }

    this.ws.onerror = () => {
      this.ws?.close()
    }
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) return
    this.reconnectTimer = setTimeout(() => {
      this.reconnectAttempts++
      this.connect()
    }, this.reconnectDelayMs * Math.min(this.reconnectAttempts + 1, 5))
  }

  private emit(type: WSEventType, payload: unknown) {
    const fns = this.handlers.get(type)
    fns?.forEach((fn) => fn(payload))
  }

  private send(msg: { type: string; payload: unknown }) {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(msg))
    }
  }

  on<T>(type: WSEventType, handler: WSHandler<T>): () => void {
    const list = this.handlers.get(type) ?? []
    list.push(handler as WSHandler)
    this.handlers.set(type, list)
    return () => {
      const updated = (this.handlers.get(type) ?? []).filter((h) => h !== handler)
      this.handlers.set(type, updated)
    }
  }

  subscribe(serverIds: string[], channel = 'metrics') {
    serverIds.forEach((id) => this.subscribedServerIds.add(id))
    this.send({ type: 'subscribe', payload: { serverIds, channel } })
  }

  unsubscribe(serverIds: string[]) {
    serverIds.forEach((id) => this.subscribedServerIds.delete(id))
    this.send({ type: 'unsubscribe', payload: { serverIds } })
  }

  ping() {
    this.send({ type: 'ping', payload: {} })
  }

  destroy() {
    this.destroyed = true
    if (this.reconnectTimer) clearTimeout(this.reconnectTimer)
    this.ws?.close()
    this.handlers.clear()
  }
}
