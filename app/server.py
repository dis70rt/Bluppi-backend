import logging
from contextlib import asynccontextmanager

from fastapi.responses import FileResponse
import httpx
from fastapi import APIRouter, FastAPI

from rich import print
import datetime
import os

import redis.asyncio as aioredis
from dotenv import load_dotenv
load_dotenv(override=True)

redis_host = os.getenv("REDIS_HOST", "localhost")
redis_port = int(os.getenv("REDIS_PORT", 6379))
redis_db   = int(os.getenv("REDIS_DB",   0))

from .endpoints.middleware import ping_middleware
from .endpoints.routes import following, history, tracks, audio, users


logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
log = logging.getLogger(__name__)

router = APIRouter()

router.include_router(tracks.router)
router.include_router(users.router)
router.include_router(following.router)
router.include_router(history.router)
router.include_router(audio.router)

WELL_KNOWN_DIR = os.path.join(os.path.dirname(__file__), ".well-known")

@asynccontextmanager
async def lifespan(app: FastAPI):
    client = httpx.AsyncClient(timeout=10.0)

    redis_client = aioredis.Redis(
        host=redis_host, port=redis_port, db=redis_db,
        encoding="utf-8", decode_responses=True
    )

    app.state.http_client = client
    app.state.redis = redis_client

    yield

    log.info("Application shutdown: Closing HTTPX client.")
    await client.aclose()
    await redis_client.close()


app = FastAPI(
    title="Bluppi API",
    description="Searches iTunes, enriches with Last.fm, and returns sorted tracks.",
    version="1.0.0",
    lifespan=lifespan,
)

app.middleware("http")(ping_middleware)

@app.get("/", tags=["Health"], include_in_schema=False)
async def read_root():
    return {
        "message": "Welcome to the Bluppi API!",
        "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat().replace("+00:00", "Z"),
    }

@app.get("/.well-known/assetlinks.json")
async def get_android_asset_links():
    file_path = os.path.join(WELL_KNOWN_DIR, "assetlinks.json")
    return FileResponse(
        path=file_path, 
        media_type="application/json"
    )

app.include_router(router)

@app.get("/{username}")
async def user_profile(username: str):
    return {
        "message": f"Profile page for {username}",
        "download_app": "https://saikat.in"
    }



    


    
