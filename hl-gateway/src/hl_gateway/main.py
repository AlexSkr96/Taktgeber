from contextlib import asynccontextmanager

import structlog
from fastapi import FastAPI, WebSocket

from .dependencies import get_ws_proxy
from .logging_config import configure_logging
from .routes.health import router as health_router
from .routes.trading import router as trading_router
from .settings import get_settings

logger = structlog.get_logger("hl_gateway.main")


@asynccontextmanager
async def lifespan(_: FastAPI):
    settings = get_settings()
    configure_logging(settings.log_level)
    logger.info(
        "service_starting",
        environment=settings.app_env,
        base_url=settings.hyperliquid_base_url,
    )
    yield
    logger.info("service_stopping")


app = FastAPI(title="hl-gateway", version="0.1.0", lifespan=lifespan)
app.include_router(health_router)
app.include_router(trading_router)


@app.get("/")
async def root() -> dict[str, str]:
    return {"service": "hl-gateway", "status": "running"}


@app.websocket("/ws")
async def websocket_proxy(websocket: WebSocket) -> None:
    await get_ws_proxy().proxy(websocket)
