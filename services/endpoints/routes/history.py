from fastapi import APIRouter, HTTPException, Path, Query, status

from ...database.db_tracks import TrackDB
from ...database.db_users import UserDB
from ...endpoints.routes.tracks import TrackInteraction

router = APIRouter(tags=["History"])

@router.post("/api/v1/user/{user_id}/history", status_code=status.HTTP_201_CREATED)
async def add_track_history(user_id: str, track: TrackInteraction):
    result = TrackDB.add_track_history(user_id, track.track_id)
    if result["status"] == "error":
        raise HTTPException(status_code=result["status_code"], detail=result["message"])
    return {"message": result["message"]}

@router.get("/api/v1/user/{user_id}/history")
async def get_track_history(
    user_id: str, limit: int = Query(20, ge=1, le=100), offset: int = Query(0, ge=0)
):
    result = TrackDB.get_track_history(user_id, limit, offset)
    if result["status"] == "error":
        raise HTTPException(status_code=result["status_code"], detail=result["message"])
    return {"history": result["history"], "total": result["total"]}

@router.delete("/api/v1/user/{user_id}/history")
async def clear_track_history(
    user_id: str = Path(..., description="ID of the user")
):
    result = TrackDB.clear_track_history(user_id)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"message": result["message"]}
