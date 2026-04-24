# Taktgeber

> *German: "the one who sets the pace"*

Taktgeber is a modular, containerized algorithmic trading system targeting the Hyperliquid exchange. It is built around a clear separation of concerns — market connectivity, trading logic, data persistence, and control interface are each owned by a dedicated service, coordinated via a shared internal network.

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Podman Internal Network                  │
│                                                             │
│  ┌──────────────┐      ┌─────────────────────────────────┐  │
│  │  hl-gateway  │      │          algo-engine            │  │
│  │  (Python)    │◄────►│          (Go)                   │  │
│  │              │      │                                 │  │
│  │ Hyperliquid  │      │  Strategy logic                 │  │
│  │ SDK wrapper  │      │  Risk management                │  │
│  │ REST + WS    │      │  Order execution                │  │
│  └──────────────┘      │  DB writes (orders, fills)      │  │
│                        └─────────────────────────────────┘  │
│         ▲                             ▲                      │
│         │                             │                      │
│  ┌──────┴─────────────────────────────┴──────────────────┐  │
│  │                    dashboard (Go)                      │  │
│  │          HTMX UI · SSE live updates · controls        │  │
│  └────────────────────────────────────────────────────────┘ │
│                                                             │
│  ┌──────────────────┐      ┌──────────────────────────────┐ │
│  │      redis       │      │         postgres             │ │
│  │  live state      │      │  durable trade history       │ │
│  │  pub/sub bus     │      │  P&L snapshots · config      │ │
│  │  kill switch     │      │  orders · fills              │ │
│  └──────────────────┘      └──────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
```

---

## Services

### `hl-gateway` — Python
The exchange connectivity layer. Wraps the official Hyperliquid Python SDK and exposes a clean internal HTTP + WebSocket API for the rest of the system. No other service talks to Hyperliquid directly.

**Responsibilities**
- REST endpoints for order placement, cancellation, and account state
- WebSocket proxy for real-time market data and order updates
- Reconnection logic for upstream WebSocket drops
- Health check endpoint

**Key dependencies:** `hyperliquid-python-sdk`, `fastapi`, `uvicorn`, `websockets`, `redis`

---

### `algo-engine` — Go
The core of the system. Consumes market data from the gateway, runs strategy logic, enforces risk rules, and executes orders. The only service that writes trade records to Postgres.

**Responsibilities**
- Market data consumption via WebSocket
- Strategy execution (pluggable strategy interface)
- Risk management (position limits, max drawdown, kill switch polling)
- Order execution with retry and idempotency
- Writing orders, fills, and P&L snapshots to Postgres
- Publishing state updates to Redis for the dashboard

**Key dependencies:** `pgx`, `go-redis`, `nhooyr.io/websocket`, `golang-migrate`

**Internal packages**
```
internal/
├── gateway/    # HTTP + WS client for hl-gateway
├── strategy/   # Strategy interface + implementations
├── risk/       # Risk checks and kill switch
├── executor/   # Order placement with idempotency
├── db/         # Postgres connection pool and queries
└── state/      # Redis client for shared state
```

---

### `dashboard` — Go
The operator control plane. A server-rendered web UI for monitoring and controlling the system in real time.

**Responsibilities**
- Display current positions, balances, and open orders
- Show order and fill history (from Postgres)
- P&L chart over time
- Start/stop strategy controls
- Flip the kill switch via Redis
- Live updates via SSE (Server-Sent Events) from Redis pub/sub

**Key dependencies:** `net/http`, `templ` (or `html/template`), HTMX, Chart.js (CDN)

---

### `redis`
Shared in-memory state and pub/sub bus. Fast, ephemeral. Not a source of truth.

**Stores**
- Current positions and balances (refreshed by engine)
- Kill switch flag (`taktgeber:kill_switch`)
- Live event stream (pub/sub channel for dashboard SSE)

---

### `postgres`
Durable record of everything that matters. The source of truth for historical data.

**Tables**

| Table | Purpose |
|---|---|
| `orders` | Every order placed, with status and strategy tag |
| `fills` | Every fill received, linked to an order |
| `pnl_snapshots` | Periodic P&L snapshots per strategy |
| `strategy_runs` | Start/stop events with parameters |

---

## Repository Structure

```
taktgeber/
├── Makefile                  # build, up, down, logs, migrate, test
├── go.work                   # Go multi-module workspace
├── .env.example              # All required env vars documented
│
├── hl-gateway/
│   ├── Dockerfile
│   ├── pyproject.toml
│   └── src/
│       ├── main.py
│       ├── routes/
│       └── ws_proxy.py
│
├── algo-engine/
│   ├── Dockerfile
│   ├── go.mod
│   └── internal/
│       ├── gateway/
│       ├── strategy/
│       ├── risk/
│       ├── executor/
│       ├── db/
│       └── state/
│
├── dashboard/
│   ├── Dockerfile
│   ├── go.mod
│   └── internal/
│       ├── handlers/
│       └── templates/
│
└── infra/
    ├── podman-compose.yml
    ├── redis.conf
    └── migrations/
        ├── 000001_init.up.sql
        └── 000001_init.down.sql
```

---

## Infrastructure

Taktgeber runs entirely under **Podman** with `podman-docker` compatibility and `podman-compose`. All services communicate over a named internal network. No service except the dashboard is exposed to the host by default.

### Container summary

| Service | Base image | Exposed (internal) | Exposed (host) |
|---|---|---|---|
| `hl-gateway` | `python:3.12-slim` | `:8000` | — |
| `algo-engine` | `gcr.io/distroless/static` | `:9000` (metrics) | — |
| `dashboard` | `gcr.io/distroless/static` | `:8080` | `:8080` |
| `redis` | `redis:7-alpine` | `:6379` | — |
| `postgres` | `postgres:16-alpine` | `:5432` | — |

### Key practices
- API credentials are passed via env file or Podman secrets — never baked into images
- Postgres data lives on a named volume, never a bind mount
- All services define health checks; `algo-engine` depends on both `hl-gateway` and `postgres` being healthy before starting
- Redis and Postgres are not reachable from outside the Podman network
- Migrations run as an explicit `make migrate` step, not automatically on engine boot

---

## Safety

- **Kill switch** — a Redis key (`taktgeber:kill_switch`) that the engine polls on every cycle. The dashboard can flip it to immediately halt all order activity without restarting any container.
- **Risk layer** — the engine enforces position size limits and max drawdown thresholds before any order reaches the executor.
- **Idempotent orders** — every order is tagged with a `client_oid` generated before the network call, so retries cannot double-place.
- **Structured logging** — all services emit JSON logs (`log/slog` in Go, `structlog` in Python) for easy aggregation.

---

## Build & Operations

```bash
# Start everything
make up

# Tail logs
make logs

# Run DB migrations
make migrate

# Stop everything
make down

# Run tests
make test
```

---

## Launch (Current Scaffold)

Use this flow to boot the project with the current repository scaffold.

```bash
# 1) Create local environment file
cp .env.example .env

# 2) Fill in credentials / settings (at least review POSTGRES_PASSWORD)
$EDITOR .env

# 3) Start the stack
docker compose -f infra/podman-compose.yml up --build -d

# 4) Confirm service health
docker compose -f infra/podman-compose.yml ps
curl http://127.0.0.1:8000/health

# 5) Follow logs
docker compose -f infra/podman-compose.yml logs -f hl-gateway

# 6) Stop everything
docker compose -f infra/podman-compose.yml down
```

`HL_GATEWAY_MOCK_TRADING=true` is enabled by default in `.env.example`, so the gateway can start without a live trading key during local development.

---

## Development Phases

| Phase | Scope |
|---|---|
| 1 | Project scaffold, Makefile, podman-compose, env setup |
| 2 | `hl-gateway` — FastAPI wrapper, WS proxy, health check |
| 3 | `algo-engine` — gateway client, first strategy, risk layer, DB writes |
| 4 | `dashboard` — positions view, order history, kill switch control |
| 5 | Observability — Prometheus metrics, P&L chart, structured logs |
| 6 | Hardening — integration tests, reconnection stress testing, resource limits |
