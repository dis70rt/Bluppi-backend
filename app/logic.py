import asyncio
from typing import List
from .yt_services import YouTubeDLP as ytDLP

from .models import Track, TrackSearchResponse
from .services import TrackRepository

async def search_enrich_and_sort_tracks(
    query: str, repository: TrackRepository
) -> TrackSearchResponse:
    if not query:
        return TrackSearchResponse(results=[], suggestedTracks=[])

    try:
        itunes_results: List[Track] = await repository.search_itunes_tracks(query)
    except Exception:
        return TrackSearchResponse(results=[], suggestedTracks=[])

    if not itunes_results:
        return TrackSearchResponse(results=[], suggestedTracks=[])

    lastfm_infos = await asyncio.gather(
        *(
            repository.fetch_lastfm_track_info(t.artistName, t.trackName)
            for t in itunes_results
        ),
        return_exceptions=True,
    )

    for track, lm in zip(itunes_results, lastfm_infos):
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
