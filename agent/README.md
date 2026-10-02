# Agent

Long-running process on each managed server. Talks to the control plane over HTTPS.

## Quick run

```bash
cp .env.example .env
# API_URL=http://<api-host>:8080
# SERVER_TOKEN=<from server create / DB>
go run .
```

## Build for API bootstrap

From this directory (outputs consumed by `api/provision`):

```bash
GOOS=linux GOARCH=amd64 go build -o ../api/bin/agent-linux-amd64 .
GOOS=linux GOARCH=arm64 go build -o ../api/bin/agent-linux-arm64 .
```

## Docs

- Protocol + install paths: [../docs/AGENT.md](../docs/AGENT.md)
- Local / Proxmox: [../docs/DEVELOPMENT.md](../docs/DEVELOPMENT.md)
