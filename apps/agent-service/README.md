# SkyDock agent service

Go sync agent (filesystem watcher, SQLite state, reconciler).

## Prerequisites

- Go 1.26+
- ZMQ is implemented with pure-Go [`go-zeromq/zmq4`](https://github.com/go-zeromq/zmq4) (no libzmq needed to build the agent)

## Run locally

```bash
cd apps/agent-service
cp .env.example .env
go run .
```

Settings live in `.env` (`SKYDOCK_WATCH_DIRS`, `SKYDOCK_DB_PATH`, and `SKYDOCK_AGENT_ZMQ_URL`). See `.env.example`. The agent binds a ZMQ **ROUTER** on `ipc:///tmp/skydock-agent.sock` unless `SKYDOCK_AGENT_ZMQ_URL` is set.

Logs should include `ZMQ listening` when the ZMQ server is up.

## SQLite

```bash
open -a "DB Browser for SQLite" "$HOME/Library/Application Support/SkyDock/database/skydock.sqlite3"
```

## Desktop dev (two terminals)

1. **Agent:** `yarn agent:dev` (from repo root) or `go run .` in this directory.
2. **Desktop:** `yarn desktop:dev` — Electron connects to the agent; it does not start the agent process.

Desktop reads `apps/agent-service/.env` for `SKYDOCK_AGENT_ZMQ_URL` so both sides use the same endpoint.
