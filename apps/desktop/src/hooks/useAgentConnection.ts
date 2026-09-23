import { useEffect, useState } from 'react'

import type { AgentReadyPayload } from '../../electron/agent-protocol'
import { EventAgentHeartbeat, EventAgentReady } from '../../electron/agent-protocol'
import { EventGetRecentActivity } from '../types/agent-events'
import { zmq } from '../services/zmq'

/** Agent heartbeat interval is 5s; treat the link as dead if none arrives in this window. */
const HEARTBEAT_STALE_MS = 12_000

/** Live connection flag for UI. Tracks agent:ready / heartbeat / disconnected over `zmq`. */
export function useAgentConnection() {
  const [connected, setConnected] = useState(false)
  const [agentVersion, setAgentVersion] = useState<string | undefined>()

  useEffect(() => {
    let staleTimer: ReturnType<typeof setTimeout> | undefined

    const markDisconnected = () => {
      setConnected(false)
      if (staleTimer) {
        clearTimeout(staleTimer)
        staleTimer = undefined
      }
    }

    const markConnected = () => {
      setConnected(true)
      if (staleTimer) {
        clearTimeout(staleTimer)
      }
      staleTimer = setTimeout(() => {
        setConnected(false)
      }, HEARTBEAT_STALE_MS)
    }

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
      markConnected()
      const data = envelope.data as AgentReadyPayload | undefined
      if (data?.version) {
        setAgentVersion(data.version)
      }
      fetchRecentActivity()
    })

    const onHeartbeat = zmq.on(EventAgentHeartbeat, () => {
      markConnected()
    })

    void zmq.isConnected().then((isUp) => {
      if (isUp) {
        markConnected()
        fetchRecentActivity()
      } else {
        markDisconnected()
      }
    })

    return () => {
      onReady()
      onHeartbeat()
      if (staleTimer) {
        clearTimeout(staleTimer)
      }
    }
  }, [])

  return { connected, agentVersion }
}
