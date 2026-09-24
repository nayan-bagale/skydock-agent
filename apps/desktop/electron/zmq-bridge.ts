import type { IpcRenderer } from 'electron'

import type { Envelope } from './agent-protocol'
import type { AgentConnectionStatus } from './zmq-event-bus'

const AGENT_EVENT_CHANNEL = 'agent:event'
const AGENT_STATUS_CHANNEL = 'agent:status'

export function createZmqBridge(ipcRenderer: IpcRenderer) {
  return {
    on(name: string, listener: (envelope: Envelope) => void) {
      const wrapped = (_event: Electron.IpcRendererEvent, envelope: Envelope) => {
        if (envelope.name === name) {
          listener(envelope)
        }
      }
      ipcRenderer.on(AGENT_EVENT_CHANNEL, wrapped)
      return () => ipcRenderer.off(AGENT_EVENT_CHANNEL, wrapped)
    },
    off() {
      ipcRenderer.removeAllListeners(AGENT_EVENT_CHANNEL)
    },
    emit(name: string, data?: unknown, withAck?: boolean) {
      return ipcRenderer.invoke('agent:emit', name, data, withAck ?? false)
    },
    isConnected() {
      return ipcRenderer.invoke('agent:is-connected') as Promise<boolean>
    },
    onStatus(listener: (status: AgentConnectionStatus) => void) {
      let sawLive = false
      const wrapped = (_event: Electron.IpcRendererEvent, status: AgentConnectionStatus) => {
        sawLive = true
        listener(status)
      }
      ipcRenderer.on(AGENT_STATUS_CHANNEL, wrapped)
      void ipcRenderer.invoke('agent:get-status').then((status: AgentConnectionStatus | null) => {
        if (status && !sawLive) {
          listener(status)
        }
      })
      return () => {
        ipcRenderer.off(AGENT_STATUS_CHANNEL, wrapped)
      }
    },
    retry() {
      return ipcRenderer.invoke('agent:retry') as Promise<void>
    },
  }
}
