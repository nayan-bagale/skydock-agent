import type { Envelope } from '../../electron/agent-protocol'
import type { AgentConnectionStatus } from '../../electron/zmq-event-bus'

export { createQueueLatestHandler, createQueueLatestZmqHandler } from './queue-latest'

/**
 * Renderer-side client for the Go agent over ZeroMQ.
 *
 * React cannot talk to ZMQ itself. The path is:
 *
 *   UI (this class)
 *     → window.electron.zmq      (preload contextBridge)
 *     → IPC `agent:emit` / `agent:event`
 *     → Electron main ZmqEventBus
 *     → ipc:///tmp/skydock-agent.sock    (Go agent DEALER socket)
 *
 * Use the exported `zmq` singleton. Do not call `window.electron.zmq` from UI code.
 * Event names live in `electron/agent-protocol.ts` (keep in sync with Go `internal/zmq/events.go`).
 */
export class Zmq {
  private static instance: Zmq | null = null

  private constructor() {}

  static getInstance(): Zmq {
    if (!Zmq.instance) {
      Zmq.instance = new Zmq()
    }
    return Zmq.instance
  }

  /**
   * Send an event/command to the agent.
   *
   * @param name Event name, e.g. `agent:ready`.
   * @param data Optional JSON-serializable payload.
   * @param withAck If true, wait for the agent's ack envelope (or timeout).
   *                If false, fire-and-forget and resolve with `null`.
   */
  emit(name: string, data?: unknown, withAck = false): Promise<unknown> {
    return window.electron.zmq.emit(name, data, withAck)
  }

  /**
   * Subscribe to one event name. Main forwards every bus event; preload
   * then matches `envelope.name`.
   *
   * @returns Unsubscribe function — call it from a React effect cleanup.
   */
  on(name: string, listener: (envelope: Envelope) => void): () => void {
    return window.electron.zmq.on(name, listener)
  }

  /** Drop every `agent:event` listener. Prefer the unsubscribe from `on()`. */
  off(): void {
    window.electron.zmq.off()
  }

  /** Whether main currently has a live ZMQ session with the agent. */
  isConnected(): Promise<boolean> {
    return window.electron.zmq.isConnected()
  }

  /** Subscribe to connection attempts. Replays the latest status if one exists. */
  onStatus(listener: (status: AgentConnectionStatus) => void): () => void {
    return window.electron.zmq.onStatus(listener)
  }

  /** Start another 5-attempt connect cycle after the previous one failed. */
  retry(): Promise<void> {
    return window.electron.zmq.retry()
  }
}

export const zmq = Zmq.getInstance()
