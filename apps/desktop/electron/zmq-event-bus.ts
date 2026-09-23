import { randomUUID } from 'node:crypto'
import * as zmq from 'zeromq'

import {
  type AgentReadyPayload,
  DEFAULT_ZMQ_URL,
  EventAgentHeartbeat,
  EventAgentReady,
  type Envelope,
  parseEnvelope,
  PROTOCOL_VERSION,
} from './agent-protocol'

type EventHandler = (data: unknown, envelope: Envelope) => void

const ACK_TIMEOUT_MS = 15_000
const HELLO_TIMEOUT_MS = 3_000
const HEARTBEAT_STALE_MS = 12_000
const WATCHDOG_INTERVAL_MS = 2_000
const MAX_BACKOFF_MS = 30_000

export class ZmqEventBus {
  private socket: zmq.Dealer | null = null
  private readonly url: string
  private readonly listeners = new Map<string, Set<EventHandler>>()
  private readonly anyListeners = new Set<EventHandler>()
  private readonly pendingAcks = new Map<
    string,
    {
      resolve: (data: unknown) => void
      reject: (err: Error) => void
      timer: ReturnType<typeof setTimeout>
    }
  >()
  private sendQueue: Promise<void> = Promise.resolve()
  private receiveLoopRunning = false
  private reconnectBackoffMs = 1_000
  private shouldRun = false
  private connected = false
  private lastEventAt = 0
  private handshakeInFlight = false
  private watchdog: ReturnType<typeof setInterval> | undefined

  constructor(url = process.env.SKYDOCK_AGENT_ZMQ_URL ?? DEFAULT_ZMQ_URL) {
    this.url = url
  }

  isConnected(): boolean {
    return this.connected
  }

  on(name: string, handler: EventHandler): void {
    let set = this.listeners.get(name)
    if (!set) {
      set = new Set()
      this.listeners.set(name, set)
    }
    set.add(handler)
  }

  off(name: string, handler: EventHandler): void {
    this.listeners.get(name)?.delete(handler)
  }

  onAny(handler: EventHandler): () => void {
    this.anyListeners.add(handler)
    return () => {
      this.anyListeners.delete(handler)
    }
  }

  async connect(): Promise<void> {
    this.shouldRun = true
    this.startWatchdog()
    await this.openSocket()
  }

  async disconnect(): Promise<void> {
    this.shouldRun = false
    this.stopWatchdog()
    this.connected = false
    if (this.socket) {
      await this.socket.close()
      this.socket = null
    }
    for (const [, pending] of this.pendingAcks) {
      clearTimeout(pending.timer)
      pending.reject(new Error('bus disconnected'))
    }
    this.pendingAcks.clear()
  }

  emit(name: string, data?: unknown): void {
    void this.enqueueSend(() => this.sendEvent(name, data))
  }

  emitWithAck<T = unknown>(name: string, data?: unknown, timeoutMs = ACK_TIMEOUT_MS): Promise<T> {
    return this.enqueueSend(async () => {
      const id = randomUUID()
      return new Promise<T>((resolve, reject) => {
        const timer = setTimeout(() => {
          this.pendingAcks.delete(id)
          reject(new Error(`ack timeout for ${name}`))
        }, timeoutMs)
        this.pendingAcks.set(id, {
          resolve: (payload) => resolve(payload as T),
          reject,
          timer,
        })
        void this.sendEvent(name, data, id)
      })
    })
  }

  private enqueueSend<T>(fn: () => Promise<T>): Promise<T> {
    const run = this.sendQueue.then(fn)
    this.sendQueue = run.then(
      () => undefined,
      () => undefined,
    )
    return run
  }

  private async openSocket(): Promise<void> {
    while (this.shouldRun) {
      try {
        const socket = new zmq.Dealer()
        this.socket = socket
        await socket.connect(this.url)
        this.reconnectBackoffMs = 1_000
        console.log('[agent-bus] socket connected', this.url)

        if (!this.receiveLoopRunning) {
          this.receiveLoopRunning = true
          void this.receiveLoop(socket)
        }

        await this.handshake()
        return
      } catch (err) {
        this.markDisconnected('connect_failed')
        console.error('[agent-bus] connect failed', err)
        if (!this.shouldRun) {
          return
        }
        await this.sleep(this.reconnectBackoffMs)
        this.reconnectBackoffMs = Math.min(this.reconnectBackoffMs * 2, MAX_BACKOFF_MS)
      }
    }
  }

  private async handshake(): Promise<boolean> {
    if (this.handshakeInFlight || !this.socket) {
      return this.connected
    }
    this.handshakeInFlight = true
    try {
      const ack = await this.emitWithAck<AgentReadyPayload>(
        EventAgentReady,
        {
          appVersion: process.env.npm_package_version ?? '0.0.0',
        },
        HELLO_TIMEOUT_MS,
      )
      if (!ack?.ok) {
        return false
      }
      this.connected = true
      this.lastEventAt = Date.now()
      const env: Envelope = {
        v: PROTOCOL_VERSION,
        kind: 'event',
        name: EventAgentReady,
        data: ack,
      }
      this.dispatch(EventAgentReady, ack, env)
      return true
    } catch (err) {
      console.warn('[agent-bus] hello handshake failed', err)
      return false
    } finally {
      this.handshakeInFlight = false
    }
  }

  private startWatchdog(): void {
    if (this.watchdog) {
      return
    }
    this.watchdog = setInterval(() => {
      void this.watchdogTick()
    }, WATCHDOG_INTERVAL_MS)
  }

  private stopWatchdog(): void {
    if (this.watchdog) {
      clearInterval(this.watchdog)
      this.watchdog = undefined
    }
  }

  private async watchdogTick(): Promise<void> {
    if (!this.shouldRun || !this.socket) {
      return
    }
    const stale = Date.now() - this.lastEventAt > HEARTBEAT_STALE_MS
    if (this.connected && stale) {
      this.markDisconnected('heartbeat_stale')
    }
    if (!this.connected) {
      await this.handshake()
    }
  }

  private markDisconnected(reason: string): void {
    const wasConnected = this.connected
    this.connected = false
    if (!wasConnected) {
      return
    }
    console.log('[agent-bus] disconnected', reason)
  }

  private async receiveLoop(socket: zmq.Dealer): Promise<void> {
    try {
      for await (const parts of socket) {
        const last = parts[parts.length - 1]
        const raw = last ? last.toString() : ''
        try {
          const env = parseEnvelope(raw)
          if (env.kind === 'ack' && env.id) {
            const pending = this.pendingAcks.get(env.id)
            if (pending) {
              clearTimeout(pending.timer)
              this.pendingAcks.delete(env.id)
              pending.resolve(env.data)
            }
            continue
          }
          if (env.kind === 'event' && env.name) {
            if (env.name === EventAgentHeartbeat || env.name === EventAgentReady) {
              this.connected = true
              this.lastEventAt = Date.now()
            }
            this.dispatch(env.name, env.data, env)
          }
        } catch (parseErr) {
          console.warn('[agent-bus] invalid envelope', parseErr)
        }
      }
    } catch (err) {
      if (this.shouldRun) {
        console.error('[agent-bus] receive loop ended', err)
        this.markDisconnected('socket_closed')
        this.receiveLoopRunning = false
        void this.reconnect()
      }
    }
  }

  private async reconnect(): Promise<void> {
    if (this.socket) {
      try {
        await this.socket.close()
      } catch {
        /* ignore */
      }
      this.socket = null
    }
    if (this.shouldRun) {
      await this.sleep(this.reconnectBackoffMs)
      this.reconnectBackoffMs = Math.min(this.reconnectBackoffMs * 2, MAX_BACKOFF_MS)
      await this.openSocket()
    }
  }

  private dispatch(name: string, data: unknown, envelope?: Envelope): void {
    const env =
      envelope ??
      ({ v: PROTOCOL_VERSION, kind: 'event', name, data } satisfies Envelope)
    for (const handler of this.anyListeners) {
      try {
        handler(data, env)
      } catch (err) {
        console.error('[agent-bus] handler error', name, err)
      }
    }
    const handlers = this.listeners.get(name)
    if (!handlers) {
      return
    }
    for (const handler of handlers) {
      try {
        handler(data, env)
      } catch (err) {
        console.error('[agent-bus] handler error', name, err)
      }
    }
  }

  private async sendEvent(name: string, data: unknown, id?: string): Promise<void> {
    if (!this.socket) {
      throw new Error('ZMQ socket not connected')
    }
    const envelope: Envelope = {
      v: PROTOCOL_VERSION,
      kind: 'event',
      name,
      data,
      ...(id ? { id } : {}),
    }
    await this.socket.send(JSON.stringify(envelope))
  }

  private sleep(ms: number): Promise<void> {
    return new Promise((resolve) => setTimeout(resolve, ms))
  }
}

export const agentBus = new ZmqEventBus()
