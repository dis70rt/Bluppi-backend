from typing import Optional
import yt_dlp
from ytmusicapi import YTMusic

class YouTubeService:
    _ytmusic = YTMusic()

    # _search_ydl = yt_dlp.YoutubeDL({
    #     "quiet": True,
    #     "extract_flat": True,
    #     "default_search": "ytsearch1",
    #     "skip_download": True,
    # })

    _audio_ydl = yt_dlp.YoutubeDL({
        "quiet": True,
        "format": "bestaudio/best",
        "skip_download": True,
        "noplaylist": True,
    })

    # @staticmethod
    # def search_video_id(query: str) -> Optional[str]:
    #     info = YouTubeDLP._search_ydl.extract_info(
    #         f"ytsearch1:{query}",
    #         download=False
    #     )
    #     entries = info.get("entries") or []
    #     return entries[0].get("id") if entries else None

    @staticmethod
    def search_video_id(title: str, artist: str) -> Optional[str]:
        query = f"{title} by {artist}"

        results = YouTubeService._ytmusic.search(
            query,
            filter="songs",
            limit=1
        )

        return results[0]["videoId"] if results else None

    @staticmethod
    def get_audio_url(video_id: str) -> Optional[str]:
        if not video_id:
            return None

        info = YouTubeService._audio_ydl.extract_info(
            f"https://www.youtube.com/watch?v={video_id}",
            download=False
        )

        return info.get("url")
