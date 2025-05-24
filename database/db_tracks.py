from typing import List, Dict, Any, Optional
import psycopg2
from pydantic import BaseModel
from datetime import datetime

from database.config import SynqItDB


class Track(BaseModel):
    id: int
    title: str
    artist: str
    album: str
    duration: int
    genres: List[str]
    image_url: str
    preview_url: str
    video_id: str
    listeners: int
    play_count: int
    popularity: int


class TrackDB:
    @staticmethod
    def write(track: Track) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                query = """
                    INSERT INTO tracks (
                        id, title, artist, album, duration, genre,
                        image_url, preview_url, video_id,
                        listeners, play_count, popularity
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s)
                    ON CONFLICT (id) DO NOTHING
                """

                values = (
                    track.id,
                    track.title,
                    track.artist,
                    track.album,
                    track.duration,
                    track.genres,
                    track.image_url,
                    track.preview_url,
                    track.video_id,
                    track.listeners,
                    track.play_count,
                    track.popularity,
                )
                db.cursor.execute(query, values)
                db.connection.commit()
                return {
                    "status": "success",
                    "status_code": 200,
                    "message": f"Track ID {track.id} inserted successfully.",
                }
            except psycopg2.Error as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error inserting track: {str(e)}",
                }

    @staticmethod
    def read(track_id: int) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                query = "SELECT * FROM tracks WHERE id = %s"
                db.cursor.execute(query, (track_id,))
                result = db.cursor.fetchone()
                if result:
                    return {
                        "status": "success",
                        "status_code": 200,
                        "track": {
                            "trackId": result[0],
                            "trackName": result[1],
                            "artistName": result[2],
                            "albumName": result[3],
                            "duration": result[4],
                            "genres": result[5],
                            "imageUrl": result[6],
                            "previewUrl": result[7],
                            "videoId": result[8],
                            "listeners": result[9],
                            "playcount": result[10],
                            "popularity": result[11],
                        },
                    }
                else:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"Track ID {track_id} not found.",
                    }
            except psycopg2.Error as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error reading track: {str(e)}",
                }

    @staticmethod
    def search(query: str, limit: int = 20, offset: int = 0) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                search_term = f"%{query}%"
                db.cursor.execute(
                    """
                    SELECT id, title, artist, album, duration, genre, image_url, 
                           preview_url, video_id, listeners, play_count, popularity 
                    FROM tracks 
                    WHERE title ILIKE %s OR artist ILIKE %s OR album ILIKE %s
                    ORDER BY popularity DESC
                    LIMIT %s OFFSET %s
                    """,
                    (search_term, search_term, search_term, limit, offset),
                )

                results = db.cursor.fetchall()
                tracks = []

                for row in results:
                    tracks.append(
                        {
                            "trackId": row[0],
                            "trackName": row[1],
                            "artistName": row[2],
                            "albumName": row[3],
                            "duration": row[4],
                            "genres": row[5],
                            "imageUrl": row[6],
                            "previewUrl": row[7],
                            "videoId": row[8],
                            "listeners": row[9],
                            "playcount": row[10],
                            "popularity": row[11],
                        }
                    )

                db.cursor.execute(
                    """
                    SELECT COUNT(*) FROM tracks 
                    WHERE title ILIKE %s OR artist ILIKE %s OR album ILIKE %s
                    """,
                    (search_term, search_term, search_term),
                )
                total_count = db.cursor.fetchone()[0]

                return {
                    "status": "success",
                    "status_code": 200,
                    "tracks": tracks,
                    "total": total_count,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error searching tracks: {str(e)}",
                }

    @staticmethod
    def add_track_history(user_id: str, track_id: int) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute("SELECT id FROM tracks WHERE id = %s", (track_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"Track with ID {track_id} not found",
                    }

                db.cursor.execute(
                    """
                    INSERT INTO history_tracks (user_id, track_id, played_at)
                    VALUES (%s, %s, NOW())
                    """,
                    (user_id, track_id),
                )

                db.cursor.execute(
                    """
                    INSERT INTO user_track (user_id, track_id, interaction_type, play_count)
                    VALUES (%s, %s, 'most_played', 1)
                    ON CONFLICT (user_id, track_id, interaction_type)
                    DO UPDATE SET 
                        play_count = user_track.play_count + 1,
                        interacted_at = NOW()
                    """,
                    (user_id, track_id),
                )

                db.cursor.execute(
                    """
                    INSERT INTO user_track (user_id, track_id, interaction_type)
                    VALUES (%s, %s, 'last_played')
                    ON CONFLICT (user_id, track_id, interaction_type)
                    DO UPDATE SET interacted_at = NOW()
                    """,
                    (user_id, track_id),
                )

                db.connection.commit()

                return {
                    "status": "success",
                    "status_code": 201,
                    "message": "Track added to history",
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error adding track to history: {str(e)}",
                }

    @staticmethod
    def get_track_history(
        user_id: str, limit: int = 20, offset: int = 0
    ) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute(
                    """
                    SELECT h.track_id, h.played_at, 
                           t.title, t.artist, t.album, t.image_url
                    FROM history_tracks h
                    JOIN tracks t ON h.track_id = t.id
                    WHERE h.user_id = %s
                    ORDER BY h.played_at DESC
                    LIMIT %s OFFSET %s
                    """,
                    (user_id, limit, offset),
                )

                history = []
                for row in db.cursor.fetchall():
                    history.append(
                        {
                            "track_id": row[0],
                            "played_at": row[1],
                            "title": row[2],
                            "artist": row[3],
                            "album": row[4],
                            "image_url": row[5],
                        }
                    )

                db.cursor.execute(
                    "SELECT COUNT(*) FROM history_tracks WHERE user_id = %s", (user_id,)
                )
                total_count = db.cursor.fetchone()[0]

                return {
                    "status": "success",
                    "status_code": 200,
                    "history": history,
                    "total": total_count,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving track history: {str(e)}",
                }

    @staticmethod
    def like_track(user_id: str, track_id: int) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute("SELECT id FROM tracks WHERE id = %s", (track_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"Track with ID {track_id} not found",
                    }

                db.cursor.execute(
                    """
                    INSERT INTO user_track (user_id, track_id, interaction_type)
                    VALUES (%s, %s, 'liked')
                    ON CONFLICT (user_id, track_id, interaction_type)
                    DO NOTHING
                    """,
                    (user_id, track_id),
                )
                db.connection.commit()

                if db.cursor.rowcount == 0:
                    return {
                        "status": "success",
                        "status_code": 200,
                        "message": "Track already liked",
                    }

                return {
                    "status": "success",
                    "status_code": 201,
                    "message": "Track liked successfully",
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error liking track: {str(e)}",
                }

    @staticmethod
    def unlike_track(user_id: str, track_id: int) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute(
                    """
                    SELECT 1 FROM user_track 
                    WHERE user_id = %s AND track_id = %s AND interaction_type = 'liked'
                    """,
                    (user_id, track_id),
                )
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "Track not liked",
                    }

                db.cursor.execute(
                    """
                    DELETE FROM user_track 
                    WHERE user_id = %s AND track_id = %s AND interaction_type = 'liked'
                    """,
                    (user_id, track_id),
                )
                db.connection.commit()

                return {
                    "status": "success",
                    "status_code": 200,
                    "message": "Track unliked successfully",
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error unliking track: {str(e)}",
                }

    @staticmethod
    def get_liked_tracks(
        user_id: str, limit: int = 20, offset: int = 0
    ) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute(
                    """
                    SELECT ut.track_id, ut.interacted_at, 
                           t.title, t.artist, t.album, t.image_url
                    FROM user_track ut
                    JOIN tracks t ON ut.track_id = t.id
                    WHERE ut.user_id = %s AND ut.interaction_type = 'liked'
                    ORDER BY ut.interacted_at DESC
                    LIMIT %s OFFSET %s
                    """,
                    (user_id, limit, offset),
                )

                liked_tracks = []
                for row in db.cursor.fetchall():
                    liked_tracks.append(
                        {
                            "track_id": row[0],
                            "liked_at": row[1],
                            "title": row[2],
                            "artist": row[3],
                            "album": row[4],
                            "image_url": row[5],
                        }
                    )

                db.cursor.execute(
                    """
                    SELECT COUNT(*) FROM user_track 
                    WHERE user_id = %s AND interaction_type = 'liked'
                    """,
                    (user_id,),
                )
                total_count = db.cursor.fetchone()[0]

                return {
                    "status": "success",
                    "status_code": 200,
                    "tracks": liked_tracks,
                    "total": total_count,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving liked tracks: {str(e)}",
                }

    @staticmethod
    def get_popular_tracks(limit: int = 20) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                db.cursor.execute(
                    """
                    SELECT id, title, artist, album, duration, genre, 
                           image_url, preview_url, video_id, 
                           listeners, play_count, popularity
                    FROM tracks
                    ORDER BY popularity DESC
                    LIMIT %s
                    """,
                    (limit,),
                )

                popular_tracks = []
                for row in db.cursor.fetchall():
                    popular_tracks.append(
                        {
                            "trackId": row[0],
                            "trackName": row[1],
                            "artistName": row[2],
                            "albumName": row[3],
                            "duration": row[4],
                            "genres": row[5],
                            "imageUrl": row[6],
                            "previewUrl": row[7],
                            "videoId": row[8],
                            "listeners": row[9],
                            "playcount": row[10],
                            "popularity": row[11],
                        }
                    )

                return {
                    "status": "success",
                    "status_code": 200,
                    "tracks": popular_tracks,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving popular tracks: {str(e)}",
                }

    @staticmethod
    def get_tracks_by_genre(
        genre: str, limit: int = 20, offset: int = 0
    ) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                db.cursor.execute(
                    """
                    SELECT id, title, artist, album, duration, genre, 
                           image_url, preview_url, video_id, 
                           listeners, play_count, popularity
                    FROM tracks
                    WHERE %s = ANY(genre)
                    ORDER BY popularity DESC
                    LIMIT %s OFFSET %s
                    """,
                    (genre, limit, offset),
                )

                genre_tracks = []
                for row in db.cursor.fetchall():
                    genre_tracks.append(
                        {
                            "trackId": row[0],
                            "trackName": row[1],
                            "artistName": row[2],
                            "albumName": row[3],
                            "duration": row[4],
                            "genres": row[5],
                            "imageUrl": row[6],
                            "previewUrl": row[7],
                            "videoId": row[8],
                            "listeners": row[9],
                            "playcount": row[10],
                            "popularity": row[11],
                        }
                    )

                db.cursor.execute(
                    """
                    SELECT COUNT(*) FROM tracks
                    WHERE %s = ANY(genre)
                    """,
                    (genre,),
                )
                total_count = db.cursor.fetchone()[0]

                return {
                    "status": "success",
                    "status_code": 200,
                    "tracks": genre_tracks,
                    "total": total_count,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving tracks by genre: {str(e)}",
                }
            
    @staticmethod
    def clear_track_history(user_id: str) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found"
                    }
                    
                db.cursor.execute(
                    """
                    DELETE FROM history_tracks 
                    WHERE user_id = %s
                    """,
                    (user_id,)
                )
                db.connection.commit()
                
                return {
                    "status": "success",
                    "status_code": 200,
                    "message": "History cleared successfully"
                }
            except Exception as e:
                db.connection.rollback()
                print(f"Error clearing history: {str(e)}")
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error clearing history: {str(e)}"
                }    
