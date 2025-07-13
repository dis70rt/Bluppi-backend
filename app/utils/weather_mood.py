import requests
import redis
import os
from google import genai
from datetime import datetime
from ytmusicapi import YTMusic
from dotenv import load_dotenv
load_dotenv(override=True)

REDIS_HOST = os.getenv("REDIS_HOST")
REDIS_PORT = os.getenv("REDIS_PORT")
GEMINI_API_KEY = os.getenv("GEMINI_API_KEY")
CACHE_TTL_SECONDS = 86400

gemini_client = genai.Client()
redis_client = redis.StrictRedis(host=REDIS_HOST, port=REDIS_PORT, db=0, decode_responses=True)
ytmusic = YTMusic()

def get_open_meteo_weather(lat, lon):
    url = (
        f"https://api.open-meteo.com/v1/forecast"
        f"?latitude={lat}&longitude={lon}"
        f"&current=temperature_2m,weathercode,cloudcover,is_day"
        f"&timezone=auto"
    )
    res = requests.get(url)
    res.raise_for_status()
    current = res.json()["current"]
    return {
        "temperature": current.get("temperature_2m"),
        "weathercode": current.get("weathercode"),
        "cloudcover": current.get("cloudcover"),
        "is_day": current.get("is_day")
    }

def classify_time_of_day(is_day=None, client_time_iso=None):
    if client_time_iso:
        dt = datetime.fromisoformat(client_time_iso.replace('Z', '+00:00'))
        hour = dt.hour
        if 5 <= hour < 12:
            return "morning"
        elif 12 <= hour < 17:
            return "afternoon"
        elif 17 <= hour < 21:
            return "evening"
        else:
            return "night"
    return "day" if is_day else "night"

def map_weathercode_to_condition(code):
    if code == 0:
        return "Clear"
    elif 1 <= code <= 3:
        return "Partly Cloudy"
    elif code in (45, 48):
        return "Fog"
    elif 51 <= code <= 67:
        return "Drizzle"
    elif 71 <= code <= 77:
        return "Snow"
    elif 80 <= code <= 82:
        return "Rain"
    elif 95 <= code <= 99:
        return "Thunderstorm"
    else:
        return "Other"

def map_condition_and_time_to_mood(condition, time_of_day):
    if condition in ("Clear", "Partly Cloudy"):
        if time_of_day == "morning":
            return "Energize"
        elif time_of_day == "afternoon":
            return "Focus"
        elif time_of_day == "evening":
            return "Relax"
        elif time_of_day == "night":
            return "Sleep"
        else:
            return "Energize" if time_of_day == "day" else "Sleep"
    elif condition == "Fog":
        return "Chill"
    elif condition in ("Rain", "Drizzle"):
        return "Romance"
    elif condition == "Snow":
        return "Feel good"
    elif condition == "Thunderstorm":
        return "Chill"
    else:
        return "Focus"

def make_cache_key(mood, weather, time_of_day):
    return f"heading:{mood}:{weather}:{time_of_day}"

def get_cached_heading(mood, weather, time_of_day):
    key = make_cache_key(mood, weather, time_of_day)
    return redis_client.get(key)

def set_cached_heading(mood, weather, time_of_day, heading):
    key = make_cache_key(mood, weather, time_of_day)
    redis_client.setex(key, CACHE_TTL_SECONDS, heading)

def generate_heading_llm(mood, weather, time_of_day):
    prompt = f"""Generate a short, attractive 2-4 word heading for a music playlist.
    
    Context:
    - Mood: {mood}
    - Weather: {weather}
    - Time: {time_of_day}
    
    Requirements:
    - Must be 2-4 words only
    - Catchy and appealing for music app users
    - Should reflect the mood and weather
        
    Return only the heading, nothing else."""
    try:
        response = gemini_client.models.generate_content(
            model="models/gemma-3-27b-it",
            contents=prompt,
        )
        return response.text.strip()
    except Exception as e:
        fallback_headings = {
            ("Energize", "Clear", "morning"): "Morning Energy",
            ("Focus", "Clear", "afternoon"): "Sunny Focus",
            ("Relax", "Clear", "evening"): "Evening Chill",
            ("Sleep", "Clear", "night"): "Starlit Dreams",
            ("Chill", "Fog", "morning"): "Misty Morning",
            ("Romance", "Rain", "evening"): "Rainy Romance",
            ("Feel good", "Snow", "afternoon"): "Snowy Bliss",
            ("Chill", "Thunderstorm", "night"): "Storm Chill"
        }
        return fallback_headings.get((mood, weather, time_of_day), "Good Vibes")

def get_mood_and_genre_params():
    browse = ytmusic.get_mood_categories()
    result = {}
    for category in browse:
        result[category['title']] = [
            {"title": item['title'], "params": item['params']} for item in category['params']
        ]
    return result

def generate_playlist_mood_heading(lat, lon, client_time_iso=None):
    weather_data = get_open_meteo_weather(lat, lon)
    time_of_day = classify_time_of_day(weather_data["is_day"], client_time_iso)
    condition = map_weathercode_to_condition(weather_data["weathercode"])
    mood = map_condition_and_time_to_mood(condition, time_of_day)
    cached_heading = get_cached_heading(mood, condition, time_of_day)
    if cached_heading:
        heading = cached_heading
    else:
        heading = generate_heading_llm(mood, condition, time_of_day)
        set_cached_heading(mood, condition, time_of_day, heading)
    params_data = get_mood_and_genre_params()
    return {
        "latitude": lat,
        "longitude": lon,
        "temperature": weather_data["temperature"],
        "condition": condition,
        "time_of_day": time_of_day,
        "mood": mood,
        "heading": heading,
        "params": params_data
    }
