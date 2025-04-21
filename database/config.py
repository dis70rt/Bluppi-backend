import os
from typing import List
import psycopg2
from dotenv import load_dotenv
from pydantic import BaseModel

load_dotenv()

class SynqItDB:
    def __init__(self):
        self.connection = None
        self.cursor = None

    def __enter__(self):
        self.connection = psycopg2.connect(
            dbname=os.getenv("DB_NAME"),
            user=os.getenv("DB_USER"),
            password=os.getenv("DB_PASSWORD"),
            host=os.getenv("DB_HOST"),
            port=os.getenv("DB_PORT"),
        )
        self.cursor = self.connection.cursor()
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        if self.cursor:
            self.cursor.close()
        if self.connection:
            self.connection.close()

    def close(self):
        if self.cursor:
            self.cursor.close()
        if self.connection:
            self.connection.close()

    class Track(BaseModel):
        id: int
        title: str
        artist: str
        album: str
        duration: int
        genres: List[str]
        image_url: str
        preview_url: str
        youtube_url: str
        listeners: int
        play_count: int
        popularity: int

        @classmethod
        def write(cls, track: "SynqItDB.Track"):
            with SynqItDB() as db:
                try:

                    query = """
                        INSERT INTO tracks (
                            id, title, artist, album, duration, genre,
                            image_url, preview_url, youtube_url,
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
                        track.youtube_url,
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
                        "message": f"Error inserting track: {e}",
                    }

        @classmethod
        def read(cls, track_id: int):
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
                                "ytUrl": result[8],
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
                        "message": f"Error reading track: {e}",
                    }
