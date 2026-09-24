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

/**
 * Main-process client for the Go agent over ZeroMQ.
 *
 * The renderer never touches this socket. It talks to Electron IPC
 * (`agent-ipc.ts`), and this class is the only DEALER peer of
 * `ipc:///tmp/skydock-agent.sock`.
 *
 * Lifecycle: connect() opens the socket, sends `agent:ready`, and treats an
 * ack with `ok: true` as connected. A watchdog then watches heartbeats. After
 * five failed tries the status becomes `failed` until retryConnection().
 *
 * Outbound messages are JSON envelopes (`agent-protocol.ts`). Pass `withAck`
 * to wait for the matching ack. Inbound events go to every `on()` listener;
 * `agent-ipc.ts` forwards them to the window.
 */

type EventHandler = (data: unknown, envelope: Envelope) => void

export type AgentConnectionState = 'connecting' | 'connected' | 'failed'

export interface AgentConnectionStatus {
  state: AgentConnectionState
  /** 1-based attempt while connecting or failed. 0 once connected. */
  attempt: number
}

/** How long emit(..., true) waits for an ack before rejecting. */
const ACK_TIMEOUT_MS = 15_000
/** No heartbeat or ready event for this long means the agent is gone. */
const HEARTBEAT_STALE_MS = 12_000
const WATCHDOG_INTERVAL_MS = 2_000
/** One connect() or retryConnection() cycle. Further tries wait for retry(). */
const RETRY_LIMIT = 5
const RETRY_INTERVAL_MS = 2_000

export class ZmqEventBus {
  private socket: zmq.Dealer | null = null
  private readonly url: string
  /** Every inbound event. Filtering by name happens in the renderer bridge. */
  private readonly listeners = new Set<EventHandler>()
  /** Ack id -> promise waiting inside emit(..., true). */
  private readonly pendingAcks = new Map<
    string,
    {
      resolve: (data: unknown) => void
      reject: (err: Error) => void
      timer: ReturnType<typeof setTimeout>
    }
  >()
  private readonly statusListeners = new Set<(status: AgentConnectionStatus) => void>()
  /**
   * Chains sends so two envelopes are never written at once.
   * A failed send does not block the next one.
   */
  private sendQueue: Promise<void> = Promise.resolve()
  private receiveLoopRunning = false
  /** False after disconnect() so retries and the receive loop stop. */
  private shouldRun = false
  /** True only after a successful `agent:ready` ack, or a live heartbeat. */
  private connected = false
  /** Timestamp of the last heartbeat or ready event. The watchdog uses this. */
  private lastEventAt = 0
  private handshakeInFlight = false
  private attempts = 0
  /** Guards runAttempts() so connect, retry, and the watchdog cannot overlap. */
  private retrying = false
  /** True after RETRY_LIMIT failures. Cleared by retryConnection() or a heartbeat. */
  private attemptsExhausted = false
  private lastStatus: AgentConnectionStatus | null = null
  private watchdog: ReturnType<typeof setInterval> | undefined

  constructor(url = process.env.SKYDOCK_AGENT_ZMQ_URL ?? DEFAULT_ZMQ_URL) {
    this.url = url
  }

  /** Live session with the agent. Socket-open alone is not enough. */
  isConnected(): boolean {
    return this.connected
  }

  /** Latest status, including one published before the renderer subscribed. */
  getStatus(): AgentConnectionStatus | null {
    return this.lastStatus
  }

  /** Subscribe to connecting / connected / failed. Returns unsubscribe. */
  onStatus(handler: (status: AgentConnectionStatus) => void): () => void {
    this.statusListeners.add(handler)
    return () => {
      this.statusListeners.delete(handler)
    }
  }

  /**
   * Start another 5-attempt cycle after the previous one failed.
   * No-op while already connected or already retrying.
   */
  async retryConnection(): Promise<void> {
    if (this.connected || this.retrying) {
      return
    }
    this.attempts = 0
    this.attemptsExhausted = false
    this.shouldRun = true
    await this.runAttempts()
  }

  /**
   * Listen to every inbound event. The handler receives the payload and the
   * full envelope. Call the returned function to remove this listener.
   */
  on(handler: EventHandler): () => void {
    this.listeners.add(handler)
    return () => {
      this.listeners.delete(handler)
    }
  }

  /** App startup. Starts the watchdog and the first connect cycle. */
  async connect(): Promise<void> {
    this.shouldRun = true
    this.attempts = 0
    this.attemptsExhausted = false
    this.startWatchdog()
    await this.runAttempts()
  }

  /** App quit. Closes the socket and rejects any emit still waiting for an ack. */
  async disconnect(): Promise<void> {
    this.shouldRun = false
    this.stopWatchdog()
    this.connected = false
    this.receiveLoopRunning = false
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

  /**
   * Send an event to the agent.
   *
   * withAck false: queue the send and resolve with null. Does not wait.
   * withAck true: include an id, wait for the ack with that id, and resolve
   * with the ack payload. Rejects after ACK_TIMEOUT_MS.
   */
  emit<T = unknown>(name: string, data?: unknown, withAck = false): Promise<T | null> {
    if (!withAck) {
      void this.enqueueSend(() => this.sendEvent(name, data))
      return Promise.resolve(null)
    }
    return this.enqueueSend(async () => {
      const id = randomUUID()
      return new Promise<T>((resolve, reject) => {
        const timer = setTimeout(() => {
          this.pendingAcks.delete(id)
          reject(new Error(`ack timeout for ${name}`))
        }, ACK_TIMEOUT_MS)
        this.pendingAcks.set(id, {
          resolve: (payload) => resolve(payload as T),
          reject,
          timer,
        })
        void this.sendEvent(name, data, id)
      })
    })
  }

  /** Run `fn` only after the previous send has settled. */
  private enqueueSend<T>(fn: () => Promise<T>): Promise<T> {
    const run = this.sendQueue.then(fn)
    this.sendQueue = run.then(
      () => undefined,
      () => undefined,
    )
    return run
  }

  /**
   * Open the socket and complete the hello handshake, up to RETRY_LIMIT times.
   * Publishes `connecting` on each try and `failed` when the cycle ends still down.
   * Returns immediately if another cycle is running, attempts are exhausted,
   * the bus was stopped, or it is already connected.
   */
  private async runAttempts(): Promise<void> {
    if (this.retrying || this.attemptsExhausted || !this.shouldRun || this.connected) {
      return
    }
    this.retrying = true
    try {
      while (this.shouldRun && this.attempts < RETRY_LIMIT) {
        if (this.connected) {
          this.finishConnected()
          return
        }
        this.attempts += 1
        this.publishStatus('connecting', this.attempts)
        if (!this.socket) {
          const opened = await this.openSocketOnce()
          if (!opened) {
            if (this.attempts < RETRY_LIMIT && this.shouldRun) {
              await new Promise((resolve) => setTimeout(resolve, RETRY_INTERVAL_MS))
            }
            continue
          }
        }
        if (await this.handshake()) {
          this.finishConnected()
          return
        }
        // A heartbeat can mark us connected while the handshake is in flight.
        if (this.connected) {
          this.finishConnected()
          return
        }
        if (this.attempts < RETRY_LIMIT && this.shouldRun) {
          await new Promise((resolve) => setTimeout(resolve, RETRY_INTERVAL_MS))
        }
      }
      if (!this.shouldRun || this.connected) {
        if (this.connected) {
          this.finishConnected()
        }
        return
      }
      this.attemptsExhausted = true
      this.publishStatus('failed', this.attempts)
    } finally {
      this.retrying = false
    }
  }

  /** Reset the retry budget and tell listeners the session is up. */
  private finishConnected(): void {
    this.attempts = 0
    this.attemptsExhausted = false
    this.publishStatus('connected', 0)
  }

  /** Connect the DEALER socket once and start reading frames from it. */
  private async openSocketOnce(): Promise<boolean> {
    try {
      const socket = new zmq.Dealer()
      this.socket = socket
      await socket.connect(this.url)
      console.log('[agent-bus] socket connected', this.url)

      if (!this.receiveLoopRunning) {
        this.receiveLoopRunning = true
        void this.receiveLoop(socket)
      }
      return true
    } catch (err) {
      console.error('[agent-bus] connect failed', err)
      this.socket = null
      this.receiveLoopRunning = false
      this.markDisconnected('connect_failed')
      return false
    }
  }

  /**
   * Ask the agent to identify itself. Connected means the ack has `ok: true`.
   * The ack is also dispatched locally so the UI sees `agent:ready` even when
   * the agent only answered the handshake and did not push a separate event.
   */
  private async handshake(): Promise<boolean> {
    if (this.handshakeInFlight || !this.socket) {
      return this.connected
    }
    this.handshakeInFlight = true
    try {
      const ack = await this.emit<AgentReadyPayload>(
        EventAgentReady,
        {
          appVersion: process.env.npm_package_version ?? '0.0.0',
        },
        true,
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

  /** Poll for a missing heartbeat. Safe to call more than once. */
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

  /** If the agent has gone quiet, drop the session and start a new connect cycle. */
  private async watchdogTick(): Promise<void> {
    if (!this.shouldRun || this.retrying || this.attemptsExhausted || !this.connected) {
      return
    }
    const stale = Date.now() - this.lastEventAt > HEARTBEAT_STALE_MS
    if (!stale) {
      return
    }
    this.markDisconnected('heartbeat_stale')
    this.attempts = 0
    this.attemptsExhausted = false
    await this.runAttempts()
  }

  /** Clear `connected`. Logs only the transition from up to down. */
  private markDisconnected(reason: string): void {
    const wasConnected = this.connected
    this.connected = false
    if (!wasConnected) {
      return
    }
    console.log('[agent-bus] disconnected', reason)
  }

  /**
   * Read until the socket closes. An ack completes the matching emit().
   * An event is delivered to listeners. Heartbeat and ready refresh the
   * watchdog clock and can recover a session that the handshake had not
   * marked connected yet.
   */
  private async receiveLoop(socket: zmq.Dealer): Promise<void> {
    try {
      for await (const parts of socket) {
        // DEALER frames: the payload is the last part.
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
              const wasConnected = this.connected
              this.connected = true
              this.lastEventAt = Date.now()
              if (!wasConnected) {
                this.finishConnected()
              }
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
        const wasConnected = this.connected
        this.markDisconnected('socket_closed')
        this.receiveLoopRunning = false
        this.socket = null
        if (wasConnected) {
          this.attempts = 0
          this.attemptsExhausted = false
          void this.runAttempts()
        }
      }
    }
  }

  /** Call every `on()` listener. One throwing handler does not stop the rest. */
  private dispatch(name: string, data: unknown, envelope?: Envelope): void {
    const env =
      envelope ??
      ({ v: PROTOCOL_VERSION, kind: 'event', name, data } satisfies Envelope)
    for (const handler of this.listeners) {
      try {
        handler(data, env)
      } catch (err) {
        console.error('[agent-bus] handler error', name, err)
      }
    }
  }

  /** JSON-encode one envelope. `id` is set only when the caller is waiting for an ack. */
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

  /** Remember the status and notify onStatus listeners. */
  private publishStatus(state: AgentConnectionState, attempt: number): void {
    const status: AgentConnectionStatus = { state, attempt }
    this.lastStatus = status
    for (const handler of this.statusListeners) {
      try {
        handler(status)
      } catch (err) {
        console.error('[agent-bus] status handler error', err)
      }
    }
  }
}

/** Process-wide bus. main.ts connects it on ready and disconnects it on quit. */
export const agentBus = new ZmqEventBus()
