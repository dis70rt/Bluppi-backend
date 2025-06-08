from fastapi import APIRouter, Depends, Query, HTTPException
from typing import Optional
import logging
from ...endpoints.core import get_redis

import redis.asyncio as aioredis

from ...app import yt_services
ytDLP = yt_services.YouTubeDLP()


router = APIRouter(tags=["Audio"])
log = logging.getLogger(__name__)


@router.get("/api/v1/audio-stream")
async def audio_stream(
    q: str | None = Query(
        None,
        min_length=1,
        title="Search Query",
        description="The search term for tracks (e.g., artist, song title).",
    ),
    id: str | None = Query(
        None,
        min_length=1,
        title="YouTube Video ID",
        description="Direct YouTube video ID for the track.",
    ),
    redis: aioredis.Redis = Depends(get_redis),
):
    if (q is None and id is None) or (q is not None and id is not None):
        raise HTTPException(
            status_code=400,
            detail="Please provide exactly one of `q` (search term) or `id` (video ID).",
        )

    if q is not None:
        cache_key = f"audio_url_search:{q.lower()}"
        if cached := await redis.get(cache_key):
            videoId, audioUrl = cached.split("|", 1)
            return {"videoId": videoId, "audioUrl": audioUrl, "cached": True}

        videoId = ytDLP.search_video_id(query=q)
        audioUrl = ytDLP.get_audio_url(video_id=videoId)
        if not videoId or not audioUrl:
            raise HTTPException(status_code=404, detail="Audio stream not found")

        await redis.set(cache_key, f"{videoId}|{audioUrl}", ex=5 * 3600)
        return {"videoId": videoId, "audioUrl": audioUrl, "cached": False}

    else:
        cache_key = f"audio_url_id:{id.lower()}"
        if cached := await redis.get(cache_key):
            return {"audioUrl": cached, "cached": True}

        audioUrl = ytDLP.get_audio_url(video_id=id)
        if not audioUrl:
            raise HTTPException(status_code=404, detail="Audio stream not found")

        await redis.set(cache_key, audioUrl, ex=5 * 3600)
        return {"audioUrl": audioUrl, "cached": False}
