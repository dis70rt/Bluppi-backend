import os
import asyncio
import random
from typing import Optional, List, Dict, Any

import httpx
from dotenv import load_dotenv
from pydantic import BaseModel, Field

from ..app import logic
from ..app.models import Track

load_dotenv()

LASTFM_API_KEY = os.getenv("LASTFM_API_KEY")
LASTFM_BASE_URL = "http://ws.audioscrobbler.com/2.0/"

class RecommendedTrack(BaseModel):
    name: str = Field(...)
    artist: str = Field(...)
    match: Optional[float] = Field(None)
    playcount: Optional[int] = Field(None)

class RecommendationResponse(BaseModel):
    recommendation: Track

async def get_similar_tracks(client: httpx.AsyncClient, artist: str, track: str, limit: int = 10) -> List[Dict[str, Any]]:
    params = {"method": "track.getsimilar", "artist": artist, "track": track, "api_key": LASTFM_API_KEY, "format": "json", "limit": limit}
    resp = await client.get(LASTFM_BASE_URL, params=params)
    resp.raise_for_status()
    data = resp.json().get('similartracks', {}).get('track', [])
    return [{"name": t.get('name'), "artist": t.get('artist', {}).get('name'), "playcount": int(t.get('playcount', 0)), "match": float(t.get('match', 0.0))} for t in data]

async def get_artist_top_tracks(client: httpx.AsyncClient, artist: str, limit: int = 10) -> List[Dict[str, Any]]:
    params = {"method": "artist.gettoptracks", "artist": artist, "api_key": LASTFM_API_KEY, "format": "json", "limit": limit}
    resp = await client.get(LASTFM_BASE_URL, params=params)
    resp.raise_for_status()
    data = resp.json().get('toptracks', {}).get('track', [])
    return [{"name": t.get('name'), "artist": t.get('artist', {}).get('name'), "playcount": int(t.get('playcount', 0)), "match": None} for t in data]

async def recommend_track(client: httpx.AsyncClient, artist: str, track: str, limit: int = 10) -> RecommendationResponse:
    candidates = await get_similar_tracks(client, artist, track, limit)
    if not candidates:
        candidates = await get_artist_top_tracks(client, artist, limit)
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
    
    rec = await logic.search_enrich_and_sort_tracks(
        f"{choice['artist']} - {choice['name']}",
        logic.TrackRepository(client),
        limit=1,
    )

    print(f"Recommended track: {rec.results[0]}")
    return RecommendationResponse(recommendation=rec.results[0])
