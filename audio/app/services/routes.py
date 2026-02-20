from fastapi import APIRouter, HTTPException
from app.services.youtube import YouTubeService
from app.schema import SearchResponse, AudioResponse
from app.core.redis import redis_client
from app.core.psql import PSQL

router = APIRouter()
CACHE_TTL = 3600

@router.get("/search", response_model=SearchResponse)
def search(artist: str, title: str, track_id: str):
    cache_key = f"search:{artist}:{title}"
    cached = redis_client.get(cache_key)

    if cached:
        video_id, audio_url = cached.split("|")
        return SearchResponse(video_id=video_id, audio_url=audio_url)

    vid = YouTubeService.search_video_id(title, artist)
    if not vid:
        raise HTTPException(status_code=404, detail="Video not found")

    audio_url = YouTubeService.get_audio_url(vid)

    if not audio_url:
        raise HTTPException(status_code=404, detail="Audio not available")
    
    redis_client.setex(cache_key, CACHE_TTL, f"{vid}|{audio_url}")
    PSQL.update_video_id(track_id=track_id, video_id=vid)
    return SearchResponse(video_id=vid, audio_url=audio_url)


@router.get("/audio/{video_id}", response_model=AudioResponse)
def audio(video_id: str):
    cache_key = f"audio:{video_id}"
    cached = redis_client.get(cache_key)
    
    if cached:
        return AudioResponse(video_id=video_id, audio_url=cached)
    
    url = YouTubeService.get_audio_url(video_id)

    if not url:
        raise HTTPException(status_code=404, detail="Audio not available")
    
    redis_client.setex(cache_key, CACHE_TTL, url)
    return AudioResponse(video_id=video_id, audio_url=url)
