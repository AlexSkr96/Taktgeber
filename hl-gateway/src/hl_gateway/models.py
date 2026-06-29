import re
from typing import Any

from pydantic import BaseModel, Field, field_validator, model_validator


_CLOID_PATTERN = re.compile(r"^0x[a-fA-F0-9]{32}$")


class PlaceOrderRequest(BaseModel):
    coin: str = Field(min_length=1)
    is_buy: bool
    size: float = Field(gt=0)
    limit_price: float = Field(gt=0)
    order_type: dict[str, Any] = Field(default_factory=lambda: {"limit": {"tif": "Gtc"}})
    reduce_only: bool = False
    client_oid: str | None = None

    @field_validator("client_oid")
    @classmethod
    def validate_client_oid(cls, value: str | None) -> str | None:
        if value is None:
            return value
        if not _CLOID_PATTERN.fullmatch(value):
            raise ValueError("client_oid must be a 16-byte hex string (0x + 32 hex chars)")
        return value


class CancelOrderRequest(BaseModel):
    coin: str = Field(min_length=1)
    order_id: int | None = Field(default=None, gt=0)
    client_oid: str | None = None

    @field_validator("client_oid")
    @classmethod
    def validate_client_oid(cls, value: str | None) -> str | None:
        if value is None:
            return value
        if not _CLOID_PATTERN.fullmatch(value):
            raise ValueError("client_oid must be a 16-byte hex string (0x + 32 hex chars)")
        return value

    @model_validator(mode="after")
    def validate_identifier(self) -> "CancelOrderRequest":
        if self.order_id is None and self.client_oid is None:
            raise ValueError("one of order_id or client_oid is required")
        return self
