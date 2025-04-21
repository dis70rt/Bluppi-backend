from dotenv import load_dotenv
from base64 import b64encode
import os

import requests

load_dotenv()
CLIENT_ID = os.getenv("SPOTIFY_CLIENT_ID")
CLIENT_SECRET = os.getenv("SPOTIFY_CLIENT_SECRET")


class Spotify:
    def __init__(self):
        pass

    def refresh(self, refresh_token: str):
        auth = f"{CLIENT_ID}:{CLIENT_SECRET}"
        b64_auth = b64encode(auth.encode()).decode()

        headers = {
            "Authorization": f"Basic {b64_auth}",
            "Content-Type": "application/x-www-form-urlencoded",
        }

        post = {"grant_type": "refresh_token", "refresh_token": refresh_token}

        response = requests.post(
            "https://accounts.spotify.com/api/token", headers=headers, data=post
        )
        if response.status_code == 200:
            return response.json()
        else:
            raise Exception(f"Error refreshing token: {response.status_code}")
        
    def last_played_track(self, access_token: str):
        headers = {
            "Authorization": f"Bearer {access_token}",
        }

        response = requests.get(
            "https://api.spotify.com/v1/me/player/recently-played", headers=headers
        )
        if response.status_code == 200:
            return response.json()
        else:
            raise Exception(f"Error fetching last played: {response.status_code}")