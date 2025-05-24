import asyncio
from typing import Dict, List
from .yt_services import YouTubeDLP as ytDLP

from .models import Track, TrackSearchResponse
from .services import TrackRepository

async def search_enrich_and_sort_tracks(
    query: str, repository: TrackRepository, limit: int
) -> TrackSearchResponse:
    if not query:
        return TrackSearchResponse(results=[], suggestedTracks=[])

    try:
        itunes_results: List[Track] = await repository.search_itunes_tracks(query, limit)
    except Exception:
        return TrackSearchResponse(results=[], suggestedTracks=[])

    if not itunes_results:
        return TrackSearchResponse(results=[], suggestedTracks=[])
    
    unique_tracks: Dict[str, Track] = {}
    for track in itunes_results:
        unique_key = f"{track.artistName.lower()}:{track.trackName.lower()}"
        if unique_key not in unique_tracks:
            unique_tracks[unique_key] = track
    
    unique_track_list = list(unique_tracks.values())
    
    if len(unique_track_list) > limit:
        unique_track_list = unique_track_list[:limit]

    lastfm_infos = await asyncio.gather(
        *(
            repository.fetch_lastfm_track_info(t.artistName, t.trackName)
            for t in itunes_results
        ),
        return_exceptions=True,
    )

    for track, lm in zip(unique_track_list, lastfm_infos):
        if isinstance(lm, dict):
            try:
                track.enrich_with_lastfm(lm)
            except Exception:
                pass

    # async def fetch_yt(track: Track):
    #     term = f"{track.artistName} - {track.trackName}"
    #     vid = await asyncio.to_thread(ytDLP.search_video_id, term)
    #     track.videoId = vid

    # await asyncio.gather(*(fetch_yt(t) for t in itunes_results))

    sorted_list = sorted(
        itunes_results, key=lambda t: getattr(t, "popularityScore", 0), reverse=True
    )

    return TrackSearchResponse(results=sorted_list, suggestedTracks=[])
