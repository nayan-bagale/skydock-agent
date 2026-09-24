import type { BrowserWindow } from 'electron'
import { ipcMain } from 'electron'

import type { ZmqEventBus } from './zmq-event-bus'

export const AGENT_EVENT_CHANNEL = 'agent:event'
export const AGENT_STATUS_CHANNEL = 'agent:status'

export function registerAgentIpc(
  bus: ZmqEventBus,
  getWindow: () => BrowserWindow | null,
): void {
  bus.on((_data, envelope) => {
    getWindow()?.webContents.send(AGENT_EVENT_CHANNEL, envelope)
  })

  bus.onStatus((status) => {
    getWindow()?.webContents.send(AGENT_STATUS_CHANNEL, status)
  })

  ipcMain.handle(
    'agent:emit',
    (_event, name: string, data?: unknown, withAck?: boolean) => {
      return bus.emit(name, data, withAck ?? false)
    },
  )

  ipcMain.handle('agent:is-connected', () => bus.isConnected())
  ipcMain.handle('agent:get-status', () => bus.getStatus())
  ipcMain.handle('agent:retry', () => bus.retryConnection())
}
