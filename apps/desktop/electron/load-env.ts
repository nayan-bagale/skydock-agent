import fs from 'node:fs'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

/**
 * Load apps/agent-service/.env into process.env before the ZMQ client starts.
 * Existing environment variables are left unchanged.
 */
function loadAgentEnv() {
  const here = path.dirname(fileURLToPath(import.meta.url))
  const candidates = [
    path.join(here, '../../agent-service/.env'),
    path.join(here, '../agent-service/.env'),
  ]

  const file = candidates.find((candidate) => fs.existsSync(candidate))
  if (!file) {
    return
  }

  const text = fs.readFileSync(file, 'utf8')
  for (const line of text.split('\n')) {
    const trimmed = line.trim()
    if (trimmed === '' || trimmed.startsWith('#')) {
      continue
    }

    const body = trimmed.startsWith('export ') ? trimmed.slice('export '.length) : trimmed
    const eq = body.indexOf('=')
    if (eq <= 0) {
      continue
    }

    const key = body.slice(0, eq).trim()
    let value = body.slice(eq + 1).trim()
    if (
      (value.startsWith('"') && value.endsWith('"')) ||
      (value.startsWith("'") && value.endsWith("'"))
    ) {
      value = value.slice(1, -1)
    }

    if (process.env[key] === undefined) {
      process.env[key] = value
    }
  }
}

loadAgentEnv()
