from typing import Any

from fastapi import APIRouter, Depends, HTTPException
from fastapi.concurrency import run_in_threadpool

from ..dependencies import get_hyperliquid_client
from ..hyperliquid_client import GatewayConfigurationError, GatewayOperationError, HyperliquidGatewayClient
from ..models import CancelOrderRequest, PlaceOrderRequest

router = APIRouter(prefix="/api/v1", tags=["trading"])


@router.post("/orders")
async def place_order(
    payload: PlaceOrderRequest,
    client: HyperliquidGatewayClient = Depends(get_hyperliquid_client),
) -> dict[str, Any]:
    try:
        result = await run_in_threadpool(client.place_order, payload)
    except GatewayConfigurationError as exc:
        raise HTTPException(status_code=503, detail=str(exc)) from exc
    except GatewayOperationError as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc

    return {"status": "ok", "result": result}


@router.post("/orders/cancel")
async def cancel_order(
    payload: CancelOrderRequest,
    client: HyperliquidGatewayClient = Depends(get_hyperliquid_client),
) -> dict[str, Any]:
    try:
        result = await run_in_threadpool(client.cancel_order, payload)
    except GatewayConfigurationError as exc:
        raise HTTPException(status_code=503, detail=str(exc)) from exc
    except GatewayOperationError as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc

    return {"status": "ok", "result": result}


@router.get("/account/state")
async def get_account_state(
    client: HyperliquidGatewayClient = Depends(get_hyperliquid_client),
) -> dict[str, Any]:
    try:
        result = await run_in_threadpool(client.account_state)
    except GatewayConfigurationError as exc:
        raise HTTPException(status_code=503, detail=str(exc)) from exc
    except GatewayOperationError as exc:
        raise HTTPException(status_code=502, detail=str(exc)) from exc

    return {"status": "ok", "result": result}
