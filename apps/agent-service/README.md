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

Settings live in `.env` (`SKYDOCK_WATCH_DIRS`, `SKYDOCK_DB_PATH`, `SKYDOCK_AGENT_ZMQ_URL`, `SKYDOCK_API_URL`, `SKYDOCK_WEB_URL`, and `SKYDOCK_API_TIMEOUT`). See `.env.example`. The agent binds a ZMQ **ROUTER** on `ipc:///tmp/skydock-agent.sock` unless `SKYDOCK_AGENT_ZMQ_URL` is set.

`SKYDOCK_API_URL` is the SkyDock API base (default `http://localhost:3000/api/v1`). `SKYDOCK_WEB_URL` is the web app origin used to build the desktop login URL (default `http://localhost:5173`). `SKYDOCK_API_TIMEOUT` is the per-request deadline (default `30s`). Desktop PKCE (`AUTH_START`, `AUTH_CALLBACK`, …) runs in `internal/auth`; the refresh token is stored in the OS keychain. See [docs/pkce.md](../../docs/pkce.md).

Logs should include `ZMQ listening` when the ZMQ server is up, and `api client ready` with the base URL.

## SQLite

```bash
open -a "DB Browser for SQLite" "$HOME/Library/Application Support/SkyDock/database/skydock.sqlite3"
```

## Desktop dev (two terminals)

1. **Agent:** `yarn agent:dev` (from repo root) or `go run .` in this directory.
2. **Desktop:** `yarn desktop:dev` — Electron connects to the agent; it does not start the agent process.

Desktop reads `apps/agent-service/.env` for `SKYDOCK_AGENT_ZMQ_URL` so both sides use the same endpoint.
