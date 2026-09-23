import type { BrowserWindow } from 'electron'
import { ipcMain } from 'electron'

import type { ZmqEventBus } from './zmq-event-bus'

export const AGENT_EVENT_CHANNEL = 'agent:event'

export function registerAgentIpc(
  bus: ZmqEventBus,
  getWindow: () => BrowserWindow | null,
): void {
  bus.onAny((_data, envelope) => {
    getWindow()?.webContents.send(AGENT_EVENT_CHANNEL, envelope)
  })

  ipcMain.handle(
    'agent:emit',
    async (_event, name: string, data?: unknown, withAck?: boolean) => {
      if (withAck) {
        return await bus.emitWithAck(name, data)
      }
      bus.emit(name, data)
      return null
    },
  )

  ipcMain.handle('agent:is-connected', () => bus.isConnected())
}
