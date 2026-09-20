/**
 * Renderer-side API mirrored from `electron/preload.ts` (contextBridge).
 * Add new methods here and in preload when extending the bridge.
 */
interface IpcRendererBridge {
  on: import('electron').IpcRenderer['on']
  off: import('electron').IpcRenderer['off']
  send: import('electron').IpcRenderer['send']
  invoke: import('electron').IpcRenderer['invoke']
  getOS: () => NodeJS.Platform
}

declare global {
  interface Window {
    ipcRenderer: IpcRendererBridge
  }
}

export type { IpcRendererBridge }
