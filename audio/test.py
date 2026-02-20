from ytmusicapi import YTMusic
import json

ytmusic = YTMusic()

song = ytmusic.get_watch_playlist("7gBadWs9Bu8")

with open("song.json", "w") as file:
    json.dump(song, file, indent=4)
