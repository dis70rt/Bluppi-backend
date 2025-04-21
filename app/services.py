import httpx
import logging
from typing import List, Dict, Any, Optional

from . import config
from .models import Track

log = logging.getLogger(__name__)

class TrackRepository:
    def __init__(self, client: httpx.AsyncClient):
        self.client = client

    async def search_itunes_tracks(self, query: str) -> List[Track]:
        params = {
            "term": query,
            "media": "music",
            "entity": "song",
            "limit": config.DEFAULT_SEARCH_LIMIT,
        }
        try:
            response = await self.client.get(config.ITUNES_BASE_URL, params=params)
            response.raise_for_status()
            data = response.json()

            if data and "results" in data:
                results = data["results"]
                return [
                    Track.from_itunes_json(item)
                    for item in results
                    if isinstance(item, dict)
                ]
            else:
                log.warning(
                    f"iTunes search for '{query}' returned unexpected data: {data}"
                )
                return []
        except httpx.RequestError as e:
            log.error(f"iTunes search network error for query '{query}': {e}")
            raise ConnectionError(f"Network error connecting to iTunes: {e}")
        except httpx.HTTPStatusError as e:
            log.error(
                f"iTunes search HTTP error for query '{query}': Status {e.response.status_code} - {e.response.text}"
            )
            raise ValueError(
                f"iTunes API returned error: Status {e.response.status_code}"
            )
        except Exception as e:
            log.exception(f"Unexpected error during iTunes search for '{query}': {e}")
            raise ValueError(f"Failed to process iTunes response: {e}")

    async def fetch_lastfm_track_info(
        self, artist: str, track_name: str
    ) -> Optional[Dict[str, Any]]:

        if not config.LASTFM_API_KEY_SET:
            log.warning("Last.fm API key not set. Skipping enrichment.")
            return None
        if not artist or not track_name:
            return None

        params = {
            "method": "track.getInfo",
            "api_key": config.LASTFM_API_KEY,
            "artist": artist,
            "track": track_name,
            "format": "json",
            "autocorrect": 1,
        }
        try:
            response = await self.client.get(config.LASTFM_BASE_URL, params=params)
            response.raise_for_status()
            data = response.json()

            if data and "track" in data and isinstance(data["track"], dict):
                track_data = data["track"]

                return {
                    "listeners": track_data.get("listeners"),
                    "playcount": track_data.get("playcount"),
                    "url": track_data.get("url"),
                    "toptags": track_data.get("toptags"),
                }
            elif data and "error" in data:
                log.warning(
                    f"Last.fm API error for '{artist} - {track_name}': {data.get('message', 'Unknown error')}"
                )
                return None
            else:
                log.warning(
                    f"Last.fm track.getInfo for '{artist} - {track_name}' returned unexpected data: {data}"
                )
                return None
        except httpx.RequestError as e:
            log.error(
                f"Last.fm track.getInfo network error for '{artist} - {track_name}': {e}"
            )
            return None
        except httpx.HTTPStatusError as e:

            if e.response.status_code != 404:
                log.error(
                    f"Last.fm track.getInfo HTTP error for '{artist} - {track_name}': Status {e.response.status_code} - {e.response.text}"
                )
            return None
        except Exception as e:
            log.exception(
                f"Unexpected error during Last.fm track.getInfo for '{artist} - {track_name}': {e}"
            )
            return None

    async def fetch_top_tracks_by_genre(
        self, genre: str, limit: int = config.DEFAULT_SUGGESTION_LIMIT
    ) -> List[Track]:

        if not config.LASTFM_API_KEY_SET:
            log.warning("Last.fm API key not set. Skipping genre suggestions.")
            return []
        if not genre:
            return []

        params = {
            "method": "tag.getTopTracks",
            "tag": genre,
            "api_key": config.LASTFM_API_KEY,
            "format": "json",
            "limit": limit,
        }
        try:
            response = await self.client.get(config.LASTFM_BASE_URL, params=params)
            response.raise_for_status()
            data = response.json()

            if (
                data
                and isinstance(data.get("toptracks"), dict)
                and isinstance(data["toptracks"].get("track"), list)
            ):
                results = data["toptracks"]["track"]

                tracks = []
                for item in results:
                    if isinstance(item, dict):
                        track = Track.from_lastfm_top_track_json(item)

                        if genre not in track.genres:
                            track.genres.append(genre)
                        tracks.append(track)
                return tracks
            else:
                log.warning(
                    f"Last.fm tag.getTopTracks for genre '{genre}' returned unexpected data: {data}"
                )
                return []
        except httpx.RequestError as e:
            log.error(
                f"Last.fm tag.getTopTracks network error for genre '{genre}': {e}"
            )
            return []
        except httpx.HTTPStatusError as e:

            log.error(
                f"Last.fm tag.getTopTracks HTTP error for genre '{genre}': Status {e.response.status_code} - {e.response.text}"
            )
            return []
        except Exception as e:
            log.exception(
                f"Unexpected error during Last.fm tag.getTopTracks for '{genre}': {e}"
            )
            return []
