# SkyDock agent service

Go sync agent (filesystem watcher, SQLite state, reconciler).

## Prerequisites

- Go 1.26+
- ZMQ is implemented with pure-Go [`go-zeromq/zmq4`](https://github.com/go-zeromq/zmq4) (no libzmq needed to build the agent)

## Run locally

```bash
cd apps/agent-service
go run .
```

The agent binds a ZMQ **ROUTER** on `tcp://127.0.0.1:17300` by default. Override with:

```bash
export SKYDOCK_AGENT_ZMQ_URL=tcp://127.0.0.1:17300
```

Logs should include `ZMQ listening` when the ZMQ server is up.

## SQLite

```bash
open -a "DB Browser for SQLite" "$HOME/Library/Application Support/SkyDock/database/skydock.sqlite3"
```

## Desktop dev (two terminals)

1. **Agent:** `yarn agent:dev` (from repo root) or `go run .` in this directory.
2. **Desktop:** `yarn desktop:dev` — Electron connects to the agent; it does not start the agent process.

Both sides should use the same `SKYDOCK_AGENT_ZMQ_URL` if you change the port.
