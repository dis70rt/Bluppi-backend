class Track:
    def __init__(self, track_id, title, artist, album, duration):
        self.track_id = track_id
        self.title = title
        self.artist = artist
        self.album = album
        self.duration = duration

    def __repr__(self):
        return f"Track({self.track_id}, {self.title}, {self.artist}, {self.album}, {self.duration})"