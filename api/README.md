# API (control plane)

Go + Gin + GORM + Postgres.

## Run

```bash
cp .env.example .env
# configure DB_* and PUBLIC_API_URL
mkdir -p bin
(cd ../agent && GOOS=linux GOARCH=amd64 go build -o ../api/bin/agent-linux-amd64 .)
go run .
```

See [../docs/DEVELOPMENT.md](../docs/DEVELOPMENT.md) for env vars and Proxmox testing.

## Layout

```text
api/
  main.go
  app/
    routes/          # /api/v1/*
    controllers/
    models/
    middleware/      # user JWT + agent token
    policies/        # RBAC
    provision/       # SSH test + agent install
    sshrun/          # SSH dial/run/upload
    resources/       # response shaping
    requests/        # validation
  config/            # env + DB
  bin/               # agent-linux-amd64, agent-linux-arm64
```

## Notable routes

- Auth: `/api/v1/auth/*`
- Orgs / members / invites: `/api/v1/organizations/*`
- Servers: `/api/v1/organizations/:org/servers` (+ `test-connection`, `:id/bootstrap`)
- Agent: `/api/v1/agent/heartbeat`, `/jobs/next`, `/jobs/:id/result`

User responses use `{ success, message, data }`. Agent protocol endpoints return bare JSON (see [../docs/AGENT.md](../docs/AGENT.md)).
