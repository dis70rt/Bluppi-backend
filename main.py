import logging
from contextlib import asynccontextmanager

import httpx
from fastapi import APIRouter, FastAPI

from rich import print
import datetime

import redis.asyncio as aioredis

from endpoints.middleware import ping_middleware
from endpoints.routes import tracks, audio, users

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
log = logging.getLogger(__name__)

router = APIRouter()
router.include_router(tracks.router)
router.include_router(audio.router)
router.include_router(users.router)

@asynccontextmanager
async def lifespan(app: FastAPI):
    log.info("Application startup: Creating HTTPX client.")

    client = httpx.AsyncClient(timeout=10.0)

    redis_client = aioredis.Redis(
        host="localhost", port=6379, db=0,
        encoding="utf-8", decode_responses=True
    )

    app.state.http_client = client
    app.state.redis = redis_client

    yield

    log.info("Application shutdown: Closing HTTPX client.")
    await client.aclose()
    await redis_client.close()


app = FastAPI(
    title="SynqIt API",
    description="Searches iTunes, enriches with Last.fm, and returns sorted tracks.",
    version="1.0.0",
    lifespan=lifespan,
)

app.middleware("http")(ping_middleware)

@app.get("/", tags=["Health"], include_in_schema=False)
async def read_root():
    return {
        "message": "Welcome to the SynqIt API!",
        "timestamp": datetime.datetime.now(datetime.timezone.utc).isoformat().replace("+00:00", "Z"),
    }

app.include_router(router)

    


    
