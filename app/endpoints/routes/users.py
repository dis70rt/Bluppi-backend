from fastapi import APIRouter, HTTPException, Path, status, Query

from pydantic import BaseModel
from app.database.db_users import UserCreate, UserDB, UserResponse, UserUpdate
from app.utils.uuid_helper import str_to_uuid

router = APIRouter(tags=["Users"])

class TrackInteraction(BaseModel):
    track_id: str

@router.post("/api/v1/user", status_code=status.HTTP_201_CREATED, operation_id="create_user")
async def create_user(user: UserCreate):
    result = UserDB.create(user)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return result

@router.get("/api/v1/user/{user_id}", operation_id="get_user_by_id")
async def get_user_by_id(user_id: str):
    result = UserDB.get_by_id(user_id)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return result

@router.get("/api/v1/user/username/{username}", operation_id="get_user_by_username")
async def get_user_by_username(username: str):
    result = UserDB.get_by_username(username)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return result

@router.patch("/api/v1/user/{user_id}", operation_id="update_user")
async def update_user(user_id: str, user_update: UserUpdate):
    result = UserDB.update(user_id, user_update)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return result

@router.delete("/api/v1/user/{user_id}", status_code=status.HTTP_200_OK, operation_id="delete_user")
async def delete_user(user_id: str):
    result = UserDB.delete(user_id)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"message": result["message"]}

@router.get("/api/v1/users/search", operation_id="search_users")
async def search_users(q: str = "", limit: int = 10, offset: int = 0):
    result = UserDB.search(q, limit, offset)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {
        "users": result["users"], 
        "total": result["total"]
    }

@router.get("/api/v1/test", operation_id="test_connection")
async def test_endpoint():
    result = UserDB.test_connection()
    if result["status"] == "error":
        raise HTTPException(status_code=500, detail=result["message"])
    return result

@router.get("/api/v1/user/check/username/{username}", operation_id="check_username")
async def check_username(username: str):
    result = UserDB.username_exists(username)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"exists": result["exists"]}

@router.get("/api/v1/user/check/email/{email}", operation_id="check_email")
async def check_email(email: str):
    result = UserDB.email_exists(email)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"exists": result["exists"]}

@router.post("/api/v1/user/{user_id}/like", status_code=status.HTTP_201_CREATED, operation_id="user_like_track")
async def user_like_track(user_id: str, track: TrackInteraction):
    track_uuid = str_to_uuid(track.track_id)
    if not track_uuid:
        raise HTTPException(status_code=400, detail="Invalid track ID format")
    
    result = UserDB.like_track(user_id, track_uuid)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"message": result["message"]}

@router.delete("/api/v1/user/{user_id}/unlike/{track_id}", operation_id="user_unlike_track")
async def unlike_track(user_id: str, track_id: str):
    track_uuid = str_to_uuid(track_id)
    if not track_uuid:
        raise HTTPException(status_code=400, detail="Invalid track ID format")
    
    result = UserDB.unlike_track(user_id, track_uuid)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"message": result["message"]}

@router.get("/api/v1/user/{user_id}/likes", operation_id="user_liked_tracks")
async def get_liked_tracks(
    user_id: str, 
    limit: int = Query(20, ge=1, le=100), 
    offset: int = Query(0, ge=0)
):
    result = UserDB.get_liked_tracks(user_id, limit, offset)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {
        "tracks": result["tracks"],
        "total": result["total"]
    }

@router.get("/api/v1/user/{user_id}/stats", operation_id="get_user_stats")
async def get_user_stats(user_id: str):
    result = UserDB.get_user_stats(user_id)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return result["stats"]

@router.post("/api/v1/user/{user_id}/search", operation_id="add_recent_search")
async def add_recent_search(user_id: str, search_data: dict):
    query = search_data.get("query", "")
    if not query:
        raise HTTPException(status_code=400, detail="Query is required")
    
    result = UserDB.add_recent_search(user_id, query)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"message": result["message"]}

@router.get("/api/v1/user/{user_id}/recent-searches", operation_id="get_recent_searches")
async def get_recent_searches(
    user_id: str,
    limit: int = Query(10, ge=1, le=50)
):
    result = UserDB.get_recent_searches(user_id, limit)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"searches": result["searches"]}