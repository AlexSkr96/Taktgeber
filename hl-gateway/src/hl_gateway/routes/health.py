from typing import Any

from fastapi import APIRouter, Depends

from ..dependencies import get_hyperliquid_client
from ..hyperliquid_client import HyperliquidGatewayClient

router = APIRouter(tags=["health"])


@router.get("/health")
async def health(client: HyperliquidGatewayClient = Depends(get_hyperliquid_client)) -> dict[str, Any]:
    return {
        "status": "ok",
        "service": "hl-gateway",
        "details": client.health_details(),
    }
