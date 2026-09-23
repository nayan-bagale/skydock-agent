import { contextBridge, ipcRenderer } from 'electron'

import type { Envelope } from './agent-protocol'

const AGENT_EVENT_CHANNEL = 'agent:event'

contextBridge.exposeInMainWorld('ipcRenderer', {
  on(...args: Parameters<typeof ipcRenderer.on>) {
    const [channel, listener] = args
    return ipcRenderer.on(channel, (event, ...args) => listener(event, ...args))
  },
  off(...args: Parameters<typeof ipcRenderer.off>) {
    const [channel, ...omit] = args
    return ipcRenderer.off(channel, ...omit)
  },
  send(...args: Parameters<typeof ipcRenderer.send>) {
    const [channel, ...omit] = args
    return ipcRenderer.send(channel, ...omit)
  },
  invoke(...args: Parameters<typeof ipcRenderer.invoke>) {
    const [channel, ...omit] = args
    return ipcRenderer.invoke(channel, ...omit)
  },
  getOS() {
    return process.platform
  },
})

contextBridge.exposeInMainWorld('agentBus', {
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
})
