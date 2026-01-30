from ytmusicapi import YTMusic
import re
import uuid
from typing import Dict, Any, Optional

class YTMusicService:
    def __init__(self):
        self.ytmusic = YTMusic()
    
    def parse_views_to_int(self, views_str: Optional[str]) -> int:
        if not views_str:
            return 0
        views_str = views_str.lower()
        num = re.match(r'[\d\.]+', views_str)
        if not num:
            return 0
        num = float(num.group(0))
        if 'k' in views_str:
            return int(num * 1000)
        if 'm' in views_str:
            return int(num * 1000000)
        if 'b' in views_str:
            return int(num * 1000000000)
        return int(num)
    
    def _generate_uuid_from_video_id(self, video_id: str) -> str:
        """Generate a consistent UUID from a video ID string"""
        if not video_id:
            return str(uuid.uuid4())
        
        namespace = uuid.UUID('a3a721a9-5094-4c9a-a836-454fe02d8f2d')
        track_uuid = uuid.uuid5(namespace, video_id)
        return str(track_uuid)
    
    def search_tracks(self, query: str, limit: int = 20, offset: int = 0) -> Dict[str, Any]:
        try:
            page = (offset // limit) + 1
            per_page = limit
            total_needed = page * per_page
            
            results = self.ytmusic.search(query, filter="songs", limit=total_needed)
            
            start_idx = (page - 1) * per_page
            end_idx = start_idx + per_page
            
            paginated_results = results[start_idx:end_idx] if results else []

            formatted_tracks = []
            for item in paginated_results:
                artist_names = "Unknown Artist"
                if item.get("artists") and isinstance(item["artists"], list) and len(item["artists"]) > 0:
                    artist_names = ", ".join([artist.get("name", "") for artist in item["artists"] if artist.get("name")])
                    if not artist_names:
                        artist_names = item["artists"][0].get("name", "Unknown Artist")
                
                album_name = None
                if item.get("album") and isinstance(item["album"], dict):
                    album_name = item["album"].get("name")
                
                image_url = None
                if item.get("thumbnails") and isinstance(item["thumbnails"], list) and len(item["thumbnails"]) > 0:
                    image_url = item["thumbnails"][-1].get("url")
                
                video_id = item.get("videoId", "")
                
                track_data = {
                    "trackId": self._generate_uuid_from_video_id(video_id),  # UUID string instead of integer
                    "trackName": item.get("title"),
                    "artistName": artist_names,
                    "albumName": album_name,
                    "duration": item.get("duration_seconds"),
                    "genres": [],
                    "imageUrl": image_url,
                    "previewUrl": None,
                    "videoId": video_id,
                    "listeners": 0,
                    "playcount": self.parse_views_to_int(item.get("views")),
                }
                formatted_tracks.append(track_data)
            
            return {
                "status": "success",
                "status_code": 200,
                "tracks": formatted_tracks,
                "total": len(results) if results else 0,
                "limit": limit,
                "offset": offset
            }
            
        except Exception as e:
            return {
                "status": "error",
                "status_code": 500,
                "message": f"Error searching tracks: {str(e)}",
                "tracks": [],
                "total": 0
            }
    
    def get_recommendations(self, video_id: str, limit: int = 5) -> Dict[str, Any]:
        try:
            if not video_id:
                return {
                    "status": "error",
                    "status_code": 400,
                    "message": "Video ID is required",
                    "tracks": []
                }
            
            radio_playlist = self.ytmusic.get_watch_playlist(videoId=video_id, limit=limit+1)
            
            if not radio_playlist or "tracks" not in radio_playlist:
                return {
                    "status": "error",
                    "status_code": 404,
                    "message": "No recommendations found",
                    "tracks": []
                }
                
            recommended_tracks = []
            for item in radio_playlist.get("tracks", []):
                if item.get("videoId") == video_id:
                    continue
                    
                artist_names = "Unknown Artist"
                if item.get("artists") and isinstance(item["artists"], list) and len(item["artists"]) > 0:
                    artist_names = ", ".join([artist.get("name", "") for artist in item["artists"] if artist.get("name")])
                    if not artist_names:
                        artist_names = item["artists"][0].get("name", "Unknown Artist")
                
                album_name = None
                if item.get("album") and isinstance(item["album"], dict):
                    album_name = item["album"].get("name")
                
                image_url = None
                if item.get("thumbnail") and isinstance(item["thumbnail"], list) and len(item["thumbnail"]) > 0:
                    image_url = item["thumbnail"][-1].get("url")
                
                current_video_id = item.get("videoId", "")
                
                track_data = {
                    "trackId": self._generate_uuid_from_video_id(current_video_id), 
                    "trackName": item.get("title"),
                    "artistName": artist_names,
                    "albumName": album_name,
                    "duration": item.get("duration_seconds"),
                    "genres": [],
                    "imageUrl": image_url,
                    "previewUrl": None,
                    "videoId": current_video_id,
                    "listeners": 0,
                    "playcount": self.parse_views_to_int(item.get("views")),
                    "popularity": 0
                }
                recommended_tracks.append(track_data)
                
                if len(recommended_tracks) >= limit:
                    break
            
            return {
                "status": "success",
                "status_code": 200,
                "tracks": recommended_tracks,
                "total": len(recommended_tracks)
            }
            
        except Exception as e:
            return {
                "status": "error",
                "status_code": 500,
                "message": f"Error getting recommendations: {str(e)}",
                "tracks": []
            }
    
    def get_mood_playlist(self, params):
        self.ytmusic.get_mood_playlists(params=params)
        