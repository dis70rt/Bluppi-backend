import httpx
import logging
from fastapi import Depends, HTTPException

import redis.asyncio as aioredis
from ..utils.services import TrackRepository

log = logging.getLogger(__name__)

async def get_http_client() -> httpx.AsyncClient:
    from app.server import app
    if not hasattr(app.state, "http_client") or not app.state.http_client:
        log.error("HTTP client unavailable")
        raise HTTPException(500, "Internal server error: HTTP client unavailable")
    return app.state.http_client

async def get_redis() -> aioredis.Redis:
    from app.server import app
    if not hasattr(app.state, "redis") or app.state.redis is None:
        log.error("Redis client unavailable")
        raise HTTPException(500, "Internal server error: Redis unavailable")
    return app.state.redis

async def get_track_repository(
        client: httpx.AsyncClient = Depends(get_http_client)
) -> TrackRepository:
    return TrackRepository(client=client)