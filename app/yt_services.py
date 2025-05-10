from typing import Optional
import yt_dlp

class YouTubeDLP:
    @staticmethod
    def search_video_id(query: str) -> Optional[str]:
        ydl_opts = {
            'quiet': True,
            'extract_flat': True,
            'default_search': 'ytsearch1',
            'skip_download': True,
        }
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(f"ytsearch1:{query}", download=False)
            entries = info.get('entries') or []
            return entries[0].get('id') if entries else None

    @staticmethod
    def get_audio_url(video_id: str) -> Optional[str]:
        if not video_id:
            return None
        ydl_opts = {
            'quiet': True,
            'format': 'bestaudio/best',
            'skip_download': True,
        }
        with yt_dlp.YoutubeDL(ydl_opts) as ydl:
            info = ydl.extract_info(f"https://www.youtube.com/watch?v={video_id}", download=False)
            return info.get('url')
        
if __name__ == "__main__":
    # Example usage
    video_id = YouTubeDLP.search_video_id("how long")
    if video_id:
        audio_url = YouTubeDLP.get_audio_url(video_id)
        print(f"Video ID: {video_id}, Audio URL: {audio_url}")
    else:
        print("No video found.")