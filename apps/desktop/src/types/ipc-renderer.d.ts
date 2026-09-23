/**
 * Renderer-side API mirrored from `electron/preload.ts` (contextBridge).
 * Add new methods here and in preload when extending the bridge.
 */
import type { Envelope } from '../../electron/agent-protocol'

interface IpcRendererBridge {
  on: import('electron').IpcRenderer['on']
  off: import('electron').IpcRenderer['off']
  send: import('electron').IpcRenderer['send']
  invoke: import('electron').IpcRenderer['invoke']
  getOS: () => NodeJS.Platform
}

interface AgentBusBridge {
  on: (name: string, listener: (envelope: Envelope) => void) => () => void
  off: () => void
  emit: (name: string, data?: unknown, withAck?: boolean) => Promise<unknown>
  isConnected: () => Promise<boolean>
}

declare global {
  interface Window {
    ipcRenderer: IpcRendererBridge
    agentBus: AgentBusBridge
  }
}

export type { AgentBusBridge, IpcRendererBridge }
