import json
from fastapi import APIRouter, Depends, Query, HTTPException, Path
from typing import Optional
import logging
import httpx
import uuid

from app.utils.models import TrackSearchResponse
from app.utils.recommendation import RecommendationResponse, recommend_track
# from utils.services import TrackRepository
# from utils import logic
from app.utils.ytmusicService import YTMusicService
from app.database.db_tracks import Track, TrackDB
from app.endpoints.core import get_http_client, get_redis
from app.utils.uuid_helper import str_to_uuid
from pydantic import BaseModel

import redis.asyncio as aioredis

router = APIRouter(tags=["Tracks"])
log = logging.getLogger(__name__)


class TrackInteraction(BaseModel):
    track_id: str

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
        20,
        gt=0,
        le=100,
        title="Result Limit",
        description="Maximum number of search results to return (default: 20, max: 100).",
    ),
    offset: int = Query(
        0, 
        ge=0,
        title="Result Offset",
        description="Number of results to skip before returning (for pagination)."
    ),
    redis: aioredis.Redis = Depends(get_redis),
):
    cache_key = f"ytmusic_search:{query}:{limit}:{offset}"
    cached = await redis.get(cache_key)

    if cached:
        log.info(f"Cache hit for YTMusic query: '{query}' (limit: {limit}, offset: {offset})")
        return TrackSearchResponse.model_validate_json(cached)

    log.info(f"Cache miss for YTMusic search query '{query}' (limit: {limit}, offset: {offset})")
    try:
        ytmusic_service = YTMusicService()
        result = ytmusic_service.search_tracks(query, limit, offset)
        
        if result["status"] == "error":
            raise HTTPException(
                status_code=result["status_code"],
                detail=result["message"]
            )
        
        track_search_response = TrackSearchResponse(
            results=result["tracks"],
            suggestedTracks=[],
            query=query,
            limit=limit,
            offset=offset,
            total=result["total"]
        )
        
        await redis.set(cache_key, track_search_response.model_dump_json(), ex=3600)
        return track_search_response
        
    except Exception as e:
        log.exception(f"An unexpected error occurred during YTMusic search for query '{query}': {e}")
        raise HTTPException(
            status_code=500, 
            detail="An internal server error occurred."
        )


# @router.get(
#     "/api/v1/search",
#     response_model=TrackSearchResponse,
# )
# async def search_tracks(
#     query: str = Query(
#         ...,
#         min_length=1,
#         title="Search Query",
#         description="The search term for tracks (e.g., artist, song title).",
#     ),
#     limit: Optional[int] = Query(
#         100,
#         gt=0,
#         le=100,
#         title="Result Limit",
#         description="Maximum number of search results to return (default: 10, max: 100).",
#     ),
#     repo: TrackRepository = Depends(get_track_repository),
#     redis: aioredis.Redis = Depends(get_redis),
# ):
#     cache_key = f"search:{query}"
#     cached = await redis.get(cache_key)

#     if cached and limit is None:
#         print(f"[yellow]Cache hit for query: '{query}'[/yellow]")
#         return TrackSearchResponse.model_validate_json(cached)

#     log.info(f"Cache miss for search query '{query}'")
#     try:
#         results: TrackSearchResponse = await logic.search_enrich_and_sort_tracks(
#             query, repo, limit
#         )
#         await redis.set(cache_key, results.model_dump_json(), ex=3600)
#         return results
#     except ConnectionError as e:
#         log.error(
#             f"Search failed due to connection error for query '{query}': {e}",
#             exc_info=True,
#         )
#         raise HTTPException(
#             status_code=503, detail=f"Could not connect to external services: {e}"
#         )
#     except ValueError as e:
#         log.error(
#             f"Search failed due to value error for query '{query}': {e}", exc_info=True
#         )
#         raise HTTPException(status_code=400, detail=f"Invalid data processing: {e}")
#     except Exception as e:
#         log.exception(
#             f"An unexpected error occurred during search for query '{query}': {e}"
#         )
#         raise HTTPException(
#             status_code=500, detail="An internal server error occurred."
#         )


@router.get("/api/v1/tracks/search")
async def search_tracks_db(q: str = "", limit: int = 20, offset: int = 0):
    result = TrackDB.search(q, limit, offset)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {
        "tracks": result["tracks"],
        "total": result["total"]
    }


@router.post("/api/v1/write-track", description="Write a track to the database.")
def write_track(track_data: dict):
    try:
        track_data['id'] = uuid.UUID(track_data['id'])
        if 'popularity' not in track_data:
            track_data['popularity'] = 0
        
        track = Track(**track_data)
        response = TrackDB.write(track)
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
async def read_track(track_id: str, redis: aioredis.Redis = Depends(get_redis)):
    track_uuid = str_to_uuid(track_id)
    if not track_uuid:
        raise HTTPException(status_code=400, detail="Invalid track ID format")
    
    cache_key = f"track:{track_uuid}"
    cached = await redis.get(cache_key)
    if cached:
        log.info(f"Cache hit for track {track_uuid}")
        return json.loads(cached)

    log.info(f"Cache miss for track {track_uuid}")
    try:
        response = TrackDB.read(track_uuid)
        if response["status"] == "success":
            await redis.set(cache_key, json.dumps(response), ex=3600)
            return response
        else:
            raise HTTPException(
                status_code=response["status_code"], detail=response["message"]
            )
    except Exception as e:
        log.error(f"Failed to read track: {e}")
        raise HTTPException(status_code=500, detail="Failed to read track")

@router.post("/api/v1/user/{user_id}/like", status_code=200)
async def like_track(user_id: str, track: TrackInteraction):
    track_uuid = str_to_uuid(track.track_id)
    if not track_uuid:
        raise HTTPException(status_code=400, detail="Invalid track ID format")
    
    result = TrackDB.like_track(user_id, track_uuid)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"message": result["message"]}


@router.delete("/api/v1/user/{user_id}/unlike/{track_id}")
async def unlike_track(user_id: str, track_id: str):
    track_uuid = str_to_uuid(track_id)
    if not track_uuid:
        raise HTTPException(status_code=400, detail="Invalid track ID format")
    
    result = TrackDB.unlike_track(user_id, track_uuid)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"message": result["message"]}


@router.get("/api/v1/user/{user_id}/likes")
async def get_liked_tracks(
    user_id: str, 
    limit: int = Query(20, ge=1, le=100), 
    offset: int = Query(0, ge=0)
):
    result = TrackDB.get_liked_tracks(user_id, limit, offset)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {
        "tracks": result["tracks"],
        "total": result["total"]
    }


@router.get("/api/v1/tracks/popular")
async def get_popular_tracks(limit: int = Query(20, ge=1, le=100)):
    result = TrackDB.get_popular_tracks(limit)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"tracks": result["tracks"]}


@router.get("/api/v1/tracks/genre/{genre}")
async def get_tracks_by_genre(
    genre: str = Path(..., description="Genre to filter tracks by"),
    limit: int = Query(20, ge=1, le=100),
    offset: int = Query(0, ge=0)
):
    result = TrackDB.get_tracks_by_genre(genre, limit, offset)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {
        "tracks": result["tracks"],
        "total": result["total"]
    }


@router.get("/api/v1/recommendations", response_model=RecommendationResponse)
async def recommend_next_track(
    videoId: str = Query(None, title="YouTube Video ID"),
    redis: aioredis.Redis = Depends(get_redis),
    client: httpx.AsyncClient = Depends(get_http_client),
):
    cache_key = f"rec:{videoId}"
    cached = await redis.get(cache_key)

    if cached:
        log.info(f"Cache hit for recommendation")
        return RecommendationResponse.model_validate_json(cached)

    recommended_track = await recommend_track(client, videoId)

    if not recommended_track:
        return HTTPException(status_code=404, detail="No recommendations found")
    
    await redis.set(cache_key, recommended_track.model_dump_json(), ex=3600)
    return recommended_track