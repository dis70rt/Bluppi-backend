import logging
from contextlib import asynccontextmanager

import httpx
from fastapi import FastAPI, Query, HTTPException, Depends

from app import logic
from app.models import TrackSearchResponse
from app.services import TrackRepository

from database.config import SynqItDB
from rich import print

logging.basicConfig(
    level=logging.INFO, format="%(asctime)s - %(name)s - %(levelname)s - %(message)s"
)
log = logging.getLogger(__name__)


@asynccontextmanager
async def lifespan(app: FastAPI):
    log.info("Application startup: Creating HTTPX client.")

    client = httpx.AsyncClient(timeout=10.0)
    app.state.http_client = client
    yield
    log.info("Application shutdown: Closing HTTPX client.")
    await client.aclose()


app = FastAPI(
    title="Music Track Search API",
    description="Searches iTunes, enriches with Last.fm, and returns sorted tracks.",
    version="1.0.0",
    lifespan=lifespan,
)


async def get_track_repository() -> TrackRepository:

    if not hasattr(app.state, "http_client") or not app.state.http_client:
        log.error("HTTPX client not available in app state during request!")

        raise HTTPException(
            status_code=500, detail="Internal server error: HTTP client unavailable"
        )
    client = app.state.http_client
    return TrackRepository(client=client)


@app.get(
    "/api/v1/search",
    response_model=TrackSearchResponse,
    tags=["Search"],
)
async def search_tracks(
    query: str = Query(
        ...,
        min_length=1,
        title="Search Query",
        description="The search term for tracks (e.g., artist, song title).",
    ),
    repo: TrackRepository = Depends(get_track_repository),
):

    log.info(f"Received search request for query: '{query}'")
    try:

        results: TrackSearchResponse = await logic.search_enrich_and_sort_tracks(
            query, repo
        )

        return results
    except ConnectionError as e:

        log.error(
            f"Search failed due to connection error for query '{query}': {e}",
            exc_info=True,
        )
        raise HTTPException(
            status_code=503, detail=f"Could not connect to external services: {e}"
        )
    except ValueError as e:
        log.error(
            f"Search failed due to value error for query '{query}': {e}", exc_info=True
        )

        raise HTTPException(status_code=400, detail=f"Invalid data processing: {e}")
    except Exception as e:

        log.exception(
            f"An unexpected error occurred during search for query '{query}': {e}"
        )
        raise HTTPException(
            status_code=500, detail="An internal server error occurred."
        )


@app.get("/", tags=["Health"], include_in_schema=False)
async def read_root():
    return {"message": "Welcome to the Music Track Search API!"}


@app.post("/api/v1/write_track", tags=["Health"])
def write_track(track: SynqItDB.Track):
    try:
        response = SynqItDB.Track.write(track)
        if response["status"] == "success":
            return response
        else:
            raise HTTPException(
                status_code=response["status_code"], detail=response["message"]
            )
    except Exception as e:
        log.error(f"Failed to write track: {e}")
        raise HTTPException(status_code=500, detail="Failed to write track")


@app.get("/api/v1/track/{track_id}", tags=["Health"])
def read_track(track_id: int):
    try:
        response = SynqItDB.Track.read(track_id)
        if response["status"] == "success":
            return response
        else:
            raise HTTPException(
                status_code=response["status_code"], detail=response["message"]
            )
    except Exception as e:
        log.error(f"Failed to read track: {e}")
        raise HTTPException(status_code=500, detail="Failed to read track")
