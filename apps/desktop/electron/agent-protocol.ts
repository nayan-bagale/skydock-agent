/** Wire protocol v1 — keep in sync with apps/agent-service/internal/zmq/events.go */

export const PROTOCOL_VERSION = 1

export const EventAgentReady = 'agent:ready'
export const EventAgentHeartbeat = 'agent:heartbeat'

export const DEFAULT_ZMQ_URL = 'ipc:///tmp/skydock-agent.sock'

export type EnvelopeKind = 'event' | 'ack'

export interface Envelope<T = unknown> {
  v: number
  kind: EnvelopeKind
  name?: string
  data?: T
  id?: string
}

export interface AgentReadyPayload {
  ok: boolean
  name: string
  version: string
  pid: number
  roots: number
}

export interface AgentHeartbeatPayload {
  ts: string
}

export function parseEnvelope(raw: string): Envelope {
  const env = JSON.parse(raw) as Envelope
  if (env.v !== PROTOCOL_VERSION) {
    throw new Error(`Unsupported protocol version ${env.v}`)
  }
  if (env.kind !== 'event' && env.kind !== 'ack') {
    throw new Error(`Invalid kind ${env.kind}`)
  }
  if (env.kind === 'event' && !env.name) {
    throw new Error('Event envelope name is required')
  }
  return env
}
