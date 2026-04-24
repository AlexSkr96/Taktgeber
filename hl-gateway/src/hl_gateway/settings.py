from functools import lru_cache

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    app_name: str = "hl-gateway"
    app_env: str = "development"
    log_level: str = "INFO"

    hyperliquid_base_url: str = "https://api.hyperliquid-testnet.xyz"
    hyperliquid_ws_url: str = "wss://api.hyperliquid-testnet.xyz/ws"
    hyperliquid_private_key: str | None = None
    hyperliquid_account_address: str | None = None

    redis_url: str = "redis://redis:6379/0"
    request_timeout_seconds: float = 10.0

    ws_reconnect_initial_seconds: float = 1.0
    ws_reconnect_max_seconds: float = 30.0

    mock_trading: bool = False

    model_config = SettingsConfigDict(
        env_prefix="HL_GATEWAY_",
        case_sensitive=False,
    )


@lru_cache
def get_settings() -> Settings:
    return Settings()
