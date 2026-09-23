import type { Envelope } from '../../electron/agent-protocol'

/**
 * Wraps an async handler with a queue-latest concurrency lock.
 *
 * - If the handler is idle, the call runs immediately.
 * - If the handler is already running, only the LATEST call is queued.
 *   Any previously queued call is dropped (only the most recent matters).
 * - When the running call finishes, the queued call (if any) is executed.
 *
 * @param handler - The async function to wrap
 * @returns A wrapped function with the same signature (sync return)
 *
 * @example
 * ```ts
 * const onHeartbeat = zmq.on(
 *   EventAgentHeartbeat,
 *   createQueueLatestHandler(async (envelope: Envelope) => {
 *     await processHeartbeat(envelope)
 *   }),
 * )
 * ```
 */
export function createQueueLatestHandler<T>(
  handler: (message: T) => Promise<void>,
): (message: T) => void {
  let processing = false
  let hasPending = false
  let pending: T

  const run = async (message: T): Promise<void> => {
    processing = true
    try {
      await handler(message)
    } catch (err) {
      console.error('[zmq] queue-latest handler error', err)
    } finally {
      processing = false

      if (hasPending) {
        const next = pending
        hasPending = false
        void run(next)
      }
    }
  }

  return (message: T) => {
    if (processing) {
      pending = message
      hasPending = true
      return
    }
    void run(message)
  }
}

/** Queue-latest lock for a renderer ZMQ listener. Only the newest envelope is kept while one is in flight. */
export function createQueueLatestZmqHandler(
  handler: (envelope: Envelope) => Promise<void>,
): (envelope: Envelope) => void {
  return createQueueLatestHandler(handler)
}
