from pydantic import BaseModel, Field, HttpUrl, field_validator, computed_field
from typing import Optional, List, Any      
import uuid
import logging

log = logging.getLogger(__name__)

def _parse_int_safe(value: Any) -> Optional[int]:
    if isinstance(value, int):
        return value
    if isinstance(value, str) and value.isdigit():
        try:
            return int(value)
        except ValueError:
            pass

    return None


def _get_lastfm_image_url(images: Optional[List[dict]]) -> Optional[str]:
    if not images or not isinstance(images, list):
        return None
    url = None
    try:

        preferred_sizes = ["extralarge", "large", "medium", "small"]
        image_map = {
            img.get("size"): img.get("#text") for img in images if isinstance(img, dict)
        }

        for size in preferred_sizes:
            if image_map.get(size):
                url = image_map[size]
                break

        if not url and images:
            last_image = images[-1]
            if isinstance(last_image, dict):
                url = last_image.get("#text")

        return url
    except Exception as e:
        log.warning(f"Error parsing Last.fm image data: {e}", exc_info=True)
        return None

class Track(BaseModel):
    trackId: Optional[uuid.UUID] = None
    artistName: str = Field(default="Unknown Artist")
    trackName: str = Field(default="Unknown Track")
    albumName: Optional[str] = None
    imageUrl: Optional[HttpUrl] = None
    previewUrl: Optional[HttpUrl] = None
    videoId: Optional[str] = None
    # audioUrl: Optional[HttpUrl] = None
    genres: List[str] = Field(default_factory=list)
    duration: Optional[int] = None
    listeners: Optional[int] = None
    playcount: Optional[int] = None
    popularity: int = Field(default=0)

    @field_validator("listeners", "playcount", mode="before")
    @classmethod
    def validate_int_string(cls, value: Any) -> Optional[int]:
        parsed = _parse_int_safe(value)

        if value is not None and value != "" and parsed is None:
            log.warning(
                f"Could not parse integer from value: '{value}' ({type(value)})"
            )
        return parsed

    @field_validator("imageUrl", "previewUrl", mode="before")
    @classmethod
    def validate_url(cls, value: Any) -> Optional[str]:
        if isinstance(value, str) and value.strip():

            if value.startswith("http://") or value.startswith("https://"):
                return value
            else:
                log.warning(f"Value '{value}' is not a valid http/https URL.")
        elif value is not None and value != "":
            log.warning(
                f"URL field received non-string value: '{value}' ({type(value)})"
            )
        return None

    @computed_field
    @property
    def popularityScore(self) -> int:
        return self.playcount or self.listeners or 0

    @classmethod
    def from_itunes_json(cls, json_data: dict) -> "Track":
        if not isinstance(json_data, dict):
            log.error(f"Expected dict for from_itunes_json, got {type(json_data)}")

            return cls()

        genre = json_data.get("primaryGenreName")
        return cls(
            trackId=json_data.get("trackId"),
            artistName=json_data.get("artistName", "Unknown Artist"),
            trackName=json_data.get("trackName", "Unknown Track"),
            albumName=json_data.get("collectionName"),
            imageUrl=json_data.get("artworkUrl100"),
            previewUrl=json_data.get("previewUrl"),
            genres=[genre] if genre else [],
        )

    @classmethod
    def from_lastfm_top_track_json(cls, json_data: dict) -> "Track":
        if not isinstance(json_data, dict):
            log.error(
                f"Expected dict for from_lastfm_top_track_json, got {type(json_data)}"
            )
            return cls()

        artist_info = json_data.get("artist", {})
        if not isinstance(artist_info, dict):
            artist_info = {}

        genre_tags = []

        artwork_url = _get_lastfm_image_url(json_data.get("image"))

        return cls(
            trackId=None,
            artistName=artist_info.get("name", "Unknown Artist"),
            trackName=json_data.get("name", "Unknown Track"),
            albumName=None,
            imageUrl=artwork_url,
            previewUrl=None,
            videoId=None,
            genres=genre_tags,
            listeners=_parse_int_safe(json_data.get("listeners")),
            playcount=_parse_int_safe(json_data.get("playcount")),
            lastFmUrl=json_data.get("url"),
        )

    def enrich_with_lastfm(self, lastfm_data: dict):
        if not isinstance(lastfm_data, dict):
            log.warning(
                f"Received non-dict data for Last.fm enrichment: {type(lastfm_data)}"
            )
            return

        self.listeners = _parse_int_safe(lastfm_data.get("listeners")) or self.listeners
        self.playcount = _parse_int_safe(lastfm_data.get("playcount")) or self.playcount

        new_url = lastfm_data.get("url")
        if isinstance(new_url, str) and (
            new_url.startswith("http://") or new_url.startswith("https://")
        ):

            try:
                self.lastFmUrl = HttpUrl(new_url)
            except Exception:
                log.warning(
                    f"Failed to validate Last.fm URL '{new_url}' during enrichment."
                )

        elif new_url:
            log.warning(
                f"Received invalid Last.fm URL format during enrichment: '{new_url}'"
            )

        lastfm_genres = []
        toptags_data = lastfm_data.get("toptags")

        tag_list = []
        if isinstance(toptags_data, dict):
            tag_list = toptags_data.get("tag", [])

        if isinstance(tag_list, list):
            for tag in tag_list:

                if isinstance(tag, dict):
                    tag_name = tag.get("name")
                    if isinstance(tag_name, str) and tag_name:
                        lastfm_genres.append(tag_name.lower())

        existing_genres_lower = {g.lower() for g in self.genres}
        for genre in lastfm_genres:
            if genre not in existing_genres_lower:
                self.genres.append(genre)

                original_case_genre = genre
                for tag in tag_list:
                    if isinstance(tag, dict) and tag.get("name", "").lower() == genre:
                        original_case_genre = tag.get("name")
                        break
                if original_case_genre not in self.genres:
                    self.genres.append(original_case_genre)

class TrackSearchResponse(BaseModel):
    results: List[Track]
    suggestedTracks: List[Track] = Field(default_factory=list) 
    query: str
    limit: int
    offset: int = 0
    total: int = 0
