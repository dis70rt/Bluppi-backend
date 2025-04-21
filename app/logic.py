import asyncio
import logging
from typing import List, Optional

from .models import Track, TrackSearchResponse
from .services import TrackRepository

log = logging.getLogger(__name__)

async def search_enrich_and_sort_tracks(
    query: str, repository: TrackRepository
) -> TrackSearchResponse:
    
    if not query:
        return TrackSearchResponse(results=[], suggestedTracks=[])

    log.info(f"Starting track search for query: '{query}'")

    try:
        itunes_results: List[Track] = await repository.search_itunes_tracks(query)
    except (ConnectionError, ValueError) as e:
        log.error(f"Failed to get initial iTunes results for '{query}': {e}")

        return TrackSearchResponse(results=[], suggestedTracks=[])

    if not itunes_results:
        log.info(f"No iTunes results found for '{query}'.")
        return TrackSearchResponse(results=[], suggestedTracks=[])

    log.info(
        f"Found {len(itunes_results)} results from iTunes for '{query}'. Starting enrichment."
    )

    enrich_tasks = []
    for track in itunes_results:
        enrich_tasks.append(
            repository.fetch_lastfm_track_info(track.artistName, track.trackName)
        )

    lastfm_infos = await asyncio.gather(*enrich_tasks, return_exceptions=True)

    enriched_results: List[Track] = []
    for i, track in enumerate(itunes_results):
        lastfm_data = lastfm_infos[i]
        if isinstance(lastfm_data, dict):
            try:
                track.enrich_with_lastfm(lastfm_data)
            except Exception as e:
                log.error(
                    f"Error enriching track '{track.trackName}' with data {lastfm_data}: {e}"
                )
        elif isinstance(lastfm_data, Exception):
            log.error(
                f"Exception during Last.fm fetch for '{track.artistName} - {track.trackName}': {lastfm_data}"
            )
        enriched_results.append(track)

    log.info(
        f"Enrichment complete for '{query}'. {len(enriched_results)} tracks processed."
    )

    sorted_results = sorted(
        enriched_results, key=lambda t: t.popularityScore, reverse=True
    )

    best_result_for_logging: Optional[Track] = (
        sorted_results[0] if sorted_results else None
    )
    if best_result_for_logging:
        log.info(
            f"Best result for '{query}' (based on score): '{best_result_for_logging.trackName}' by {best_result_for_logging.artistName} (Score: {best_result_for_logging.popularityScore})"
        )
    else:
        log.info(f"Could not determine a best result for '{query}' after sorting.")

    return TrackSearchResponse(results=sorted_results, suggestedTracks=[])
