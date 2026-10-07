import { app } from 'electron'
import path from 'node:path'

import { EventAuthCallback } from './agent-protocol'
import type { ZmqEventBus } from './zmq-event-bus'

type CallbackPayload = {
  code: string
  state: string
}

let queued: CallbackPayload | null = null

export function registerSkydockProtocol(bus: ZmqEventBus): void {
  if (process.defaultApp) {
    if (process.argv.length >= 2) {
      app.setAsDefaultProtocolClient('skydock', process.execPath, [
        path.resolve(process.argv[1]),
      ])
    }
  } else {
    app.setAsDefaultProtocolClient('skydock')
  }

  app.on('open-url', (event, url) => {
    event.preventDefault()
    void handleCallbackUrl(bus, url)
  })
}

export function handleArgvForCallback(bus: ZmqEventBus, argv: string[]): void {
  for (const arg of argv) {
    if (arg.startsWith('skydock://')) {
      void handleCallbackUrl(bus, arg)
    }
  }
}

export function setupAuthCallbackDelivery(bus: ZmqEventBus): void {
  bus.onStatus((status) => {
    if (status.state === 'connected') {
      void flushQueuedCallback(bus)
    }
  })
}

function parseCallbackUrl(raw: string): CallbackPayload | null {
  try {
    const url = new URL(raw)
    if (url.protocol !== 'skydock:') {
      return null
    }
    if (url.hostname !== 'callback') {
      return null
    }
    const code = url.searchParams.get('code')
    const state = url.searchParams.get('state')
    if (!code || !state) {
      return null
    }
    return { code, state }
  } catch {
    return null
  }
}

async function handleCallbackUrl(bus: ZmqEventBus, raw: string): Promise<void> {
  const payload = parseCallbackUrl(raw)
  if (!payload) {
    return
  }
  queued = payload
  await tryDeliverCallback(bus)
}

async function flushQueuedCallback(bus: ZmqEventBus): Promise<void> {
  if (!queued) {
    return
  }
  await tryDeliverCallback(bus)
}

async function tryDeliverCallback(bus: ZmqEventBus): Promise<void> {
  if (!queued) {
    return
  }
  const connected = await bus.isConnected()
  if (!connected) {
    return
  }
  const payload = queued
  try {
    await bus.emit(EventAuthCallback, payload, true)
    queued = null
  } catch {
    // Keep queued until the agent is reachable.
  }
}
