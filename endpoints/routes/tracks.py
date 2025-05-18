import json
from fastapi import APIRouter, Depends, Query, HTTPException
from typing import Optional
import logging
import httpx

from app.models import TrackSearchResponse
from app.recommendation import RecommendationResponse, recommend_track
from app.services import TrackRepository
from app import logic
from database.config import SynqItDB
from endpoints.core import get_http_client, get_redis, get_track_repository

import redis.asyncio as aioredis

router = APIRouter(tags=["Tracks"])
log = logging.getLogger(__name__)


@router.get(
    "/api/v1/search",
    response_model=TrackSearchResponse,
)
async def search_tracks(
    query: str = Query(
        ...,
        min_length=1,
        title="Search Query",
        description="The search term for tracks (e.g., artist, song title).",
    ),
    limit: Optional[int] = Query(
        100,
        gt=0,
        le=100,
        title="Result Limit",
        description="Maximum number of search results to return (default: 10, max: 100).",
    ),
    repo: TrackRepository = Depends(get_track_repository),
    redis: aioredis.Redis = Depends(get_redis),
):

    cache_key = f"search:{query}"
    cached = await redis.get(cache_key)

    if cached and limit is None:
        print(f"[yellow]Cache hit for query: '{query}'[/yellow]")
        return TrackSearchResponse.model_validate_json(cached)

    log.info(f"Cache miss for search query '{query}'")
    try:

        results: TrackSearchResponse = await logic.search_enrich_and_sort_tracks(
            query, repo, limit
        )
        await redis.set(cache_key, results.model_dump_json(), ex=3600)
        return results

    except ConnectionError as e:
        log.error(
            f"Search failed due to connection error for query '{query}': {e}",
            exc_info=True,
        )
        raise HTTPException(
            status_code=503, detail=f"Could not connect to external services: {e}"
        )

    except ValueError as e:
        log.error(
            f"Search failed due to value error for query '{query}': {e}", exc_info=True
        )
        raise HTTPException(status_code=400, detail=f"Invalid data processing: {e}")

    except Exception as e:
        log.exception(
            f"An unexpected error occurred during search for query '{query}': {e}"
        )
        raise HTTPException(
            status_code=500, detail="An internal server error occurred."
        )


@router.post("/api/v1/write-track", description="Write a track to the database.")
def write_track(track: SynqItDB.Track):
    try:
        response = SynqItDB.Track.write(track)
        if response["status"] == "success":
            return response
        else:
            raise HTTPException(
                status_code=response["status_code"], detail=response["message"]
            )
    except Exception as e:
        log.error(f"Failed to write track: {e}")
        raise HTTPException(status_code=500, detail="Failed to write track")


@router.get("/api/v1/track/{track_id}")
async def read_track(track_id: int, redis: aioredis.Redis = Depends(get_redis)):

    cache_key = f"track:{track_id}"
    cached = await redis.get(cache_key)
    if cached:
        log.info(f"Cache hit for track {track_id}")
        return json.loads(cached)

    log.info(f"Cache miss for track {track_id}")
    try:
        response = SynqItDB.Track.read(track_id)
        if response["status"] == "success":
            return response
        else:
            raise HTTPException(
                status_code=response["status_code"], detail=response["message"]
            )
    except Exception as e:
        log.error(f"Failed to read track: {e}")
        raise HTTPException(status_code=500, detail="Failed to read track")


@router.get("/api/v1/recommendations", response_model=RecommendationResponse)
async def recommend_next_track(
    artist: str = Query(..., title="Artist Name"),
    track: str = Query(..., title="Track Name"),
    redis: aioredis.Redis = Depends(get_redis),
    client: httpx.AsyncClient = Depends(get_http_client),
):

    # cache_key = f"recommendation:{artist.lower()}:{track.lower()}"
    # cached = await redis.get(cache_key)
    # if cached:
    #     log.info(f"Cache hit for recommendation: '{artist} - {track}'")
    #     return RecommendationResponse.model_validate_json(cached)

    recommended_track = await recommend_track(client, artist, track)

    if not recommended_track:
        return HTTPException(status_code=404, detail="No recommendations found")

    return recommended_track
