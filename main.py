import logging
from contextlib import asynccontextmanager
import json

import httpx
from fastapi import FastAPI, Query, HTTPException, Depends

from app import logic
from app.models import TrackSearchResponse
from app.services import TrackRepository

from database.config import SynqItDB
from rich import print

import redis.asyncio as aioredis 

from app import yt_services
ytDLP = yt_services.YouTubeDLP()

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
log = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    log.info("Application startup: Creating HTTPX client.")

    client = httpx.AsyncClient(timeout=10.0)

    redis_client = aioredis.Redis(
        host="localhost", port=6379, db=0,
        encoding="utf-8", decode_responses=True
    )

    app.state.http_client = client
    app.state.redis = redis_client

    yield

    log.info("Application shutdown: Closing HTTPX client.")
    await client.aclose()
    await redis_client.close()


app = FastAPI(
    title="Music Track Search API",
    description="Searches iTunes, enriches with Last.fm, and returns sorted tracks.",
    version="1.0.0",
    lifespan=lifespan,
)

async def get_http_client() -> httpx.AsyncClient:
    if not hasattr(app.state, "http_client") or not app.state.http_client:
        log.error("HTTP client unavailable")
        raise HTTPException(500, "Internal server error: HTTP client unavailable")
    return app.state.http_client

async def get_redis() -> aioredis.Redis:
    if not hasattr(app.state, "redis") or app.state.redis is None:
        log.error("Redis client unavailable")
        raise HTTPException(500, "Internal server error: Redis unavailable")
    return app.state.redis


async def get_track_repository(
        client: httpx.AsyncClient = Depends(get_http_client)
) -> TrackRepository:
    return TrackRepository(client=client)


@app.get(
    "/api/v1/search",
    response_model=TrackSearchResponse,
    tags=["Search"],
)
async def search_tracks(
    query: str = Query(
        ...,
        min_length=1,
        title="Search Query",
        description="The search term for tracks (e.g., artist, song title).",
    ),
    repo: TrackRepository = Depends(get_track_repository),
    redis: aioredis.Redis = Depends(get_redis),
):
    
    cache_key = f"search:{query}"
    cached = await redis.get(cache_key)

    if cached:
        print(f"[yellow]Cache hit for query: '{query}'[/yellow]")
        return TrackSearchResponse.model_validate_json(cached)

    log.info(f"Cache miss for search query '{query}'")
    try:

        results: TrackSearchResponse = await logic.search_enrich_and_sort_tracks(query, repo)
        await redis.set(cache_key, results.model_dump_json(), ex=3600)
        return results
    
    except ConnectionError as e:
        log.error(f"Search failed due to connection error for query '{query}': {e}",exc_info=True,)
        raise HTTPException(status_code=503, detail=f"Could not connect to external services: {e}")
    
    except ValueError as e:
        log.error(f"Search failed due to value error for query '{query}': {e}", exc_info=True)
        raise HTTPException(status_code=400, detail=f"Invalid data processing: {e}")
    
    except Exception as e:
        log.exception(f"An unexpected error occurred during search for query '{query}': {e}")
        raise HTTPException(status_code=500, detail="An internal server error occurred.")


@app.get("/", tags=["Health"], include_in_schema=False)
async def read_root():
    return {"message": "Welcome to the Music Track Search API!"}


@app.post("/api/v1/write-track", tags=["WriteTrack"])
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


@app.get("/api/v1/track/{track_id}", tags=["Health"])
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
    

@app.get("/api/v1/audio-stream", tags=["audio"])
async def audio_stream(
    q: str | None = Query(
        None,
        min_length=1,
        title="Search Query",
        description="The search term for tracks (e.g., artist, song title)."
    ),
    id: str | None = Query(
        None,
        min_length=1,
        title="YouTube Video ID",
        description="Direct YouTube video ID for the track."
    ),
    redis: aioredis.Redis = Depends(get_redis),
):
    if (q is None and id is None) or (q is not None and id is not None):
        # require exactly one of them
        raise HTTPException(
            status_code=400,
            detail="Please provide exactly one of `q` (search term) or `id` (video ID)."
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

    else:  # id is not None
        cache_key = f"audio_url_id:{id.lower()}"
        if cached := await redis.get(cache_key):
            return {"audioUrl": cached, "cached": True}

        audioUrl = ytDLP.get_audio_url(video_id=id)
        if not audioUrl:
            raise HTTPException(status_code=404, detail="Audio stream not found")

        await redis.set(cache_key, audioUrl, ex=5 * 3600)
        return {"audioUrl": audioUrl, "cached": False}