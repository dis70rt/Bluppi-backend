import random
import uuid
from typing import Optional, List, Dict, Any
import logging

import httpx
from pydantic import BaseModel, Field

from .ytmusicService import YTMusicService
from .models import Track

log = logging.getLogger(__name__)

class RecommendedTrack(BaseModel):
    name: str = Field(...)
    artist: str = Field(...)
    match: Optional[float] = Field(None)
    playcount: Optional[int] = Field(None)

class RecommendationResponse(BaseModel):
    recommendation: Track

async def get_yt_recommendations(client: httpx.AsyncClient, video_id: str, limit: int = 10) -> List[Dict[str, Any]]:
    ytmusic = YTMusicService()    
    recommendations = ytmusic.get_recommendations(video_id=video_id, limit=limit)
    
    if recommendations["status"] != "success" or not recommendations["tracks"]:
        return []
    
    result = []
    for rec in recommendations["tracks"]:
        max_playcount = 10000000
        playcount = rec.get("playcount", 0)
        match = min(playcount / max_playcount, 1.0) if playcount > 0 else 0.5
        
        image_url = rec.get("imageUrl")
        if not image_url and rec.get("videoId"):
            image_url = f"https://img.youtube.com/vi/{rec.get('videoId')}/hqdefault.jpg"
        
        result.append({
            "name": rec.get("trackName", ""),
            "artist": rec.get("artistName", ""),
            "playcount": playcount,
            "match": match,
            "videoId": rec.get("videoId", ""),
            "imageUrl": image_url,
            "albumName": rec.get("albumName", ""),
            "trackId": str(uuid.uuid4())
        })
    
    return result

def convert_to_track(recommendation: Dict[str, Any]) -> Track:
    image_url = recommendation.get("imageUrl")
    if not image_url and recommendation.get("videoId"):
        video_id = recommendation.get("videoId")
        image_url = f"https://img.youtube.com/vi/{video_id}/hqdefault.jpg"
    
    return Track(
        trackId=recommendation.get("trackId", str(uuid.uuid4())),
        trackName=recommendation.get("name", ""),
        artistName=recommendation.get("artist", ""),
        albumName=recommendation.get("albumName", ""),
        imageUrl=image_url,
        videoId=recommendation.get("videoId", ""),
        genres=[],
        duration=None,
        previewUrl=None,
        listeners=0,
        popularity=0,
        playcount=recommendation.get("playcount", 0)
    )

async def recommend_track(client: httpx.AsyncClient, video_id: str, limit: int = 10) -> Optional[RecommendationResponse]:
    candidates = await get_yt_recommendations(client, video_id, limit)
    
    if not candidates:
        return None
    
    matches = [c.get('match') or 0.0 for c in candidates]
    plays = [c.get('playcount') or 0 for c in candidates]
    max_play = max(plays) if plays else 0
    
    weights = []
    for m, p in zip(matches, plays):
        w_play = p / max_play if max_play > 0 else 0
        weights.append(m * 0.5 + w_play * 0.5)
    
    total = sum(weights)
    if total <= 0:
        choice = random.choice(candidates)
    else:
        choice = random.choices(candidates, weights=weights, k=1)[0]
    
    track_obj = convert_to_track(choice)
    
    return RecommendationResponse(recommendation=track_obj)