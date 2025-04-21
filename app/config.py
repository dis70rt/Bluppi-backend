import os
from dotenv import load_dotenv

load_dotenv()

LASTFM_API_KEY = os.getenv("LASTFM_API_KEY")
LASTFM_BASE_URL = "http://ws.audioscrobbler.com/2.0/"
LASTFM_API_KEY_SET = LASTFM_API_KEY != "YOUR_LASTFM_API_KEY"

ITUNES_BASE_URL = "https://itunes.apple.com/search"

DEFAULT_SEARCH_LIMIT = 100
DEFAULT_SUGGESTION_LIMIT = 10
SUGGESTIONS_TO_RETURN = 5