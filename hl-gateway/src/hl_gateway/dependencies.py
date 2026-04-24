from functools import lru_cache

from .hyperliquid_client import HyperliquidGatewayClient
from .settings import get_settings
from .ws_proxy import WebSocketProxy


@lru_cache
def get_hyperliquid_client() -> HyperliquidGatewayClient:
    return HyperliquidGatewayClient(get_settings())


@lru_cache
def get_ws_proxy() -> WebSocketProxy:
    return WebSocketProxy(get_settings())
