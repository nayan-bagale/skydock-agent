# skydock-agent

Monorepo: Go **agent-service** + Electron **desktop** client.

## Development

The desktop app uses the native `zeromq` npm package in the Electron main process; run `yarn` from the repo root after cloning.

Run the agent and desktop in **two terminals** (desktop connects only; it does not spawn the agent):

```bash
# Terminal 1
yarn agent:dev

# Terminal 2
SKYDOCK_AGENT_ZMQ_URL=tcp://127.0.0.1:17300 yarn desktop:dev
```

Default ZMQ URL: `tcp://127.0.0.1:17300` on loopback. Set `SKYDOCK_AGENT_ZMQ_URL` on both processes to use another port.

See [apps/agent-service/README.md](apps/agent-service/README.md) and [apps/desktop/README.md](apps/desktop/README.md) for details.
