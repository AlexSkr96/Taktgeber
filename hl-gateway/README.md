# hl-gateway

`hl-gateway` is the Hyperliquid connectivity layer for Taktgeber.  
It exposes internal HTTP and WebSocket APIs so the rest of the stack can trade and consume market data without talking to Hyperliquid directly.

## Endpoints

- `GET /health` — service health and runtime mode
- `POST /api/v1/orders` — place order
- `POST /api/v1/orders/cancel` — cancel by order id or client oid
- `GET /api/v1/account/state` — user state + open orders
- `WS /ws` — upstream Hyperliquid WebSocket proxy with reconnect/subscription replay

## Configuration

All settings use the `HL_GATEWAY_` env prefix (see root `.env.example`):

- `HL_GATEWAY_HYPERLIQUID_BASE_URL`
- `HL_GATEWAY_HYPERLIQUID_WS_URL`
- `HL_GATEWAY_HYPERLIQUID_PRIVATE_KEY`
- `HL_GATEWAY_HYPERLIQUID_ACCOUNT_ADDRESS`
- `HL_GATEWAY_MOCK_TRADING`

`HL_GATEWAY_MOCK_TRADING=true` allows local startup without a private key.

## Local run

From repo root:

```bash
podman compose -f infra/podman-compose.yml up --build -d hl-gateway
curl http://127.0.0.1:8000/health
```
