from __future__ import annotations

import asyncio
import json
from typing import Any

import structlog
import websockets
from fastapi import WebSocket, WebSocketDisconnect
from websockets.exceptions import ConnectionClosed

from .settings import Settings


class ClientDisconnected(Exception):
    pass


class UpstreamDisconnected(Exception):
    pass


class WebSocketProxy:
    def __init__(self, settings: Settings) -> None:
        self._settings = settings
        self._logger = structlog.get_logger("hl_gateway.ws_proxy")

    async def proxy(self, client_socket: WebSocket) -> None:
        await client_socket.accept()

        subscriptions: dict[str, str] = {}
        backoff = self._settings.ws_reconnect_initial_seconds

        while True:
            try:
                async with websockets.connect(
                    self._settings.hyperliquid_ws_url,
                    ping_interval=20,
                    ping_timeout=20,
                    close_timeout=5,
                ) as upstream_socket:
                    self._logger.info("upstream_connected", url=self._settings.hyperliquid_ws_url)
                    await self._replay_subscriptions(upstream_socket, subscriptions)
                    backoff = self._settings.ws_reconnect_initial_seconds

                    should_reconnect = await self._bridge(client_socket, upstream_socket, subscriptions)
                    if not should_reconnect:
                        return
            except ClientDisconnected:
                return
            except Exception as exc:
                self._logger.warning("upstream_connection_error", error=str(exc), retry_in_seconds=backoff)

            await asyncio.sleep(backoff)
            backoff = min(backoff * 2, self._settings.ws_reconnect_max_seconds)

    async def _bridge(
        self,
        client_socket: WebSocket,
        upstream_socket: Any,
        subscriptions: dict[str, str],
    ) -> bool:
        client_to_upstream = asyncio.create_task(
            self._pipe_client_to_upstream(client_socket, upstream_socket, subscriptions)
        )
        upstream_to_client = asyncio.create_task(self._pipe_upstream_to_client(client_socket, upstream_socket))

        done, pending = await asyncio.wait(
            {client_to_upstream, upstream_to_client},
            return_when=asyncio.FIRST_EXCEPTION,
        )

        for task in pending:
            task.cancel()
        if pending:
            await asyncio.gather(*pending, return_exceptions=True)

        for task in done:
            exc = task.exception()
            if exc is None:
                continue
            if isinstance(exc, ClientDisconnected):
                return False
            if isinstance(exc, UpstreamDisconnected):
                return True
            raise exc

        return True

    async def _pipe_client_to_upstream(
        self,
        client_socket: WebSocket,
        upstream_socket: Any,
        subscriptions: dict[str, str],
    ) -> None:
        try:
            while True:
                message = await client_socket.receive_text()
                self._track_subscription_state(message, subscriptions)
                await upstream_socket.send(message)
        except WebSocketDisconnect as exc:
            raise ClientDisconnected from exc
        except ConnectionClosed as exc:
            raise UpstreamDisconnected from exc

    async def _pipe_upstream_to_client(self, client_socket: WebSocket, upstream_socket: Any) -> None:
        try:
            async for message in upstream_socket:
                if isinstance(message, bytes):
                    await client_socket.send_bytes(message)
                else:
                    await client_socket.send_text(message)
        except ConnectionClosed as exc:
            raise UpstreamDisconnected from exc
        except WebSocketDisconnect as exc:
            raise ClientDisconnected from exc
        except RuntimeError as exc:
            raise ClientDisconnected from exc

        raise UpstreamDisconnected

    async def _replay_subscriptions(self, upstream_socket: Any, subscriptions: dict[str, str]) -> None:
        if not subscriptions:
            return
        for message in subscriptions.values():
            await upstream_socket.send(message)

    @staticmethod
    def _track_subscription_state(message: str, subscriptions: dict[str, str]) -> None:
        try:
            payload = json.loads(message)
        except json.JSONDecodeError:
            return

        method = payload.get("method")
        subscription = payload.get("subscription")
        if not isinstance(subscription, dict):
            return

        subscription_key = json.dumps(subscription, sort_keys=True)
        if method == "subscribe":
            subscriptions[subscription_key] = message
        elif method == "unsubscribe":
            subscriptions.pop(subscription_key, None)
