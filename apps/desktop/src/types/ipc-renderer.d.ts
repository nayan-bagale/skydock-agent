/**
 * Renderer-side API mirrored from `electron/preload.ts` (contextBridge).
 * Add new methods here and in preload when extending the bridge.
 */
import type { Envelope } from '../../electron/agent-protocol'
import type { AgentConnectionStatus } from '../../electron/zmq-event-bus'

interface ElectronBridge {
  on: import('electron').IpcRenderer['on']
  off: import('electron').IpcRenderer['off']
  send: import('electron').IpcRenderer['send']
  invoke: import('electron').IpcRenderer['invoke']
  getOS: () => NodeJS.Platform
  zmq: ZmqBridge
}

interface ZmqBridge {
  on: (name: string, listener: (envelope: Envelope) => void) => () => void
  off: () => void
  emit: (name: string, data?: unknown, withAck?: boolean) => Promise<unknown>
  isConnected: () => Promise<boolean>
  onStatus: (listener: (status: AgentConnectionStatus) => void) => () => void
  retry: () => Promise<void>
}

declare global {
  interface Window {
    electron: ElectronBridge
  }
}

export type { ElectronBridge, ZmqBridge }
