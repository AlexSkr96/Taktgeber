from __future__ import annotations

from dataclasses import dataclass
from typing import Any

import structlog
from eth_account import Account
from hyperliquid.exchange import Exchange
from hyperliquid.info import Info
from hyperliquid.utils.types import Cloid

from .models import CancelOrderRequest, PlaceOrderRequest
from .settings import Settings


class GatewayConfigurationError(RuntimeError):
    pass


class GatewayOperationError(RuntimeError):
    pass


@dataclass(slots=True)
class _MockOrder:
    oid: int
    coin: str
    is_buy: bool
    size: float
    limit_price: float
    client_oid: str | None


class HyperliquidGatewayClient:
    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._logger = structlog.get_logger("hl_gateway.client")
        self._account_address = settings.hyperliquid_account_address
        self._info: Info | None = None
        self._exchange: Exchange | None = None

        self._mock_orders: dict[int, _MockOrder] = {}
        self._mock_next_oid = 1

        if settings.mock_trading:
            self._logger.warning("mock_mode_enabled")

    def health_details(self) -> dict[str, Any]:
        return {
            "mode": "mock" if self._settings.mock_trading else "live",
            "base_url": self._settings.hyperliquid_base_url,
            "ws_url": self._settings.hyperliquid_ws_url,
            "trading_enabled": self._settings.mock_trading or self._settings.hyperliquid_private_key is not None,
            "has_account_address": self._account_address is not None,
        }

    def place_order(self, payload: PlaceOrderRequest) -> dict[str, Any]:
        if self._settings.mock_trading:
            return self._mock_place_order(payload)

        exchange = self._ensure_exchange()
        cloid = self._as_cloid(payload.client_oid)

        try:
            if cloid is None:
                return exchange.order(
                    payload.coin,
                    payload.is_buy,
                    payload.size,
                    payload.limit_price,
                    payload.order_type,
                    reduce_only=payload.reduce_only,
                )

            return exchange.order(
                payload.coin,
                payload.is_buy,
                payload.size,
                payload.limit_price,
                payload.order_type,
                reduce_only=payload.reduce_only,
                cloid=cloid,
            )
        except Exception as exc:
            raise GatewayOperationError(f"order placement failed: {exc}") from exc

    def cancel_order(self, payload: CancelOrderRequest) -> dict[str, Any]:
        if self._settings.mock_trading:
            return self._mock_cancel_order(payload)

        exchange = self._ensure_exchange()

        try:
            if payload.order_id is not None:
                return exchange.cancel(payload.coin, payload.order_id)

            cloid = self._as_cloid(payload.client_oid)
            if cloid is None:
                raise GatewayConfigurationError("client_oid is required when order_id is not provided")
            return exchange.cancel_by_cloid(payload.coin, cloid)
        except GatewayConfigurationError:
            raise
        except Exception as exc:
            raise GatewayOperationError(f"order cancellation failed: {exc}") from exc

    def account_state(self) -> dict[str, Any]:
        if self._settings.mock_trading:
            return {
                "account_address": self._account_address,
                "user_state": {"marginSummary": {}, "assetPositions": []},
                "open_orders": list(self._mock_orders.values()),
            }

        if self._account_address is None:
            raise GatewayConfigurationError("HL_GATEWAY_HYPERLIQUID_ACCOUNT_ADDRESS must be set")

        info = self._ensure_info()
        try:
            return {
                "account_address": self._account_address,
                "user_state": info.user_state(self._account_address),
                "open_orders": info.open_orders(self._account_address),
            }
        except Exception as exc:
            raise GatewayOperationError(f"failed to query account state: {exc}") from exc

    def _ensure_info(self) -> Info:
        if self._info is not None:
            return self._info

        try:
            self._info = Info(
                self._settings.hyperliquid_base_url,
                skip_ws=True,
                timeout=self._settings.request_timeout_seconds,
            )
            return self._info
        except Exception as exc:
            raise GatewayOperationError(f"failed to initialize Hyperliquid info client: {exc}") from exc

    def _ensure_exchange(self) -> Exchange:
        if self._exchange is not None:
            return self._exchange

        if self._settings.hyperliquid_private_key is None:
            raise GatewayConfigurationError("HL_GATEWAY_HYPERLIQUID_PRIVATE_KEY must be set for trading endpoints")

        try:
            account = Account.from_key(self._settings.hyperliquid_private_key)
        except Exception as exc:
            raise GatewayConfigurationError("invalid HL_GATEWAY_HYPERLIQUID_PRIVATE_KEY value") from exc

        if self._account_address is None:
            self._account_address = account.address

        try:
            self._exchange = Exchange(
                account,
                self._settings.hyperliquid_base_url,
                account_address=self._account_address,
                timeout=self._settings.request_timeout_seconds,
            )
            return self._exchange
        except Exception as exc:
            raise GatewayOperationError(f"failed to initialize Hyperliquid exchange client: {exc}") from exc

    @staticmethod
    def _as_cloid(client_oid: str | None) -> Cloid | None:
        if client_oid is None:
            return None
        return Cloid.from_str(client_oid)

    def _mock_place_order(self, payload: PlaceOrderRequest) -> dict[str, Any]:
        oid = self._mock_next_oid
        self._mock_next_oid += 1

        self._mock_orders[oid] = _MockOrder(
            oid=oid,
            coin=payload.coin,
            is_buy=payload.is_buy,
            size=payload.size,
            limit_price=payload.limit_price,
            client_oid=payload.client_oid,
        )

        return {
            "status": "ok",
            "response": {
                "type": "mock",
                "data": {
                    "statuses": [
                        {
                            "resting": {
                                "oid": oid,
                            }
                        }
                    ]
                },
            },
        }

    def _mock_cancel_order(self, payload: CancelOrderRequest) -> dict[str, Any]:
        found_oid: int | None = None
        if payload.order_id is not None and payload.order_id in self._mock_orders:
            found_oid = payload.order_id
        elif payload.client_oid is not None:
            for oid, order in self._mock_orders.items():
                if order.client_oid == payload.client_oid:
                    found_oid = oid
                    break

        if found_oid is None:
            return {
                "status": "err",
                "error": "order_not_found",
            }

        del self._mock_orders[found_oid]
        return {
            "status": "ok",
            "response": {"type": "mock", "data": {"cancelled_oid": found_oid}},
        }
