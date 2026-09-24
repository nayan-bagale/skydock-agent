import { useEffect, useState } from 'react'

import type { AgentReadyPayload } from '../../electron/agent-protocol'
import { EventAgentHeartbeat, EventAgentReady } from '../../electron/agent-protocol'
import type { AgentConnectionState } from '../../electron/zmq-event-bus'
import { EventGetRecentActivity } from '../types/agent-events'
import { zmq } from '../services/zmq'

/** Live connection flag for UI. Status comes from the main-process retry cycle. */
export function useAgentConnection() {
  const [status, setStatus] = useState<AgentConnectionState>('connecting')
  const [attempt, setAttempt] = useState(0)
  const [agentVersion, setAgentVersion] = useState<string | undefined>()

  useEffect(() => {
    const fetchRecentActivity = () => {
      void zmq.emit(EventGetRecentActivity, {}, true).then(
        (payload) => {
          console.log('GET_RECENT_ACTIVITY', payload)
        },
        (err) => {
          console.warn('GET_RECENT_ACTIVITY failed', err)
        },
      )
    }

    const onReady = zmq.on(EventAgentReady, (envelope) => {
      setStatus('connected')
      setAttempt(0)
      const data = envelope.data as AgentReadyPayload | undefined
      if (data?.version) {
        setAgentVersion(data.version)
      }
      fetchRecentActivity()
    })

    const onHeartbeat = zmq.on(EventAgentHeartbeat, () => {
      setStatus('connected')
      setAttempt(0)
    })

    const onStatus = zmq.onStatus((next) => {
      setStatus(next.state)
      setAttempt(next.attempt)
    })

    return () => {
      onReady()
      onHeartbeat()
      onStatus()
    }
  }, [])

  return {
    connected: status === 'connected',
    status,
    attempt,
    agentVersion,
    retry: () => zmq.retry(),
  }
}
