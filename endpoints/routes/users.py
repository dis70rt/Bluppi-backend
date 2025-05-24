from fastapi import APIRouter, HTTPException, Path, status, Query

from pydantic import BaseModel
from database.db_users import UserCreate, UserDB, UserResponse, UserUpdate

router = APIRouter(tags=["Users"])
class TrackInteraction(BaseModel):
    track_id: int

@router.post("/api/v1/user", status_code=status.HTTP_201_CREATED, response_model=UserResponse)
async def create_user(user: UserCreate):
    result = UserDB.create(user)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return result["user"]

@router.get("/api/v1/user/{user_id}", response_model=UserResponse)
async def get_user_by_id(user_id: str):
    result = UserDB.get_by_id(user_id)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return result["user"]

@router.get("/api/v1/user/username/{username}", response_model=UserResponse)
async def get_user_by_username(username: str):
    result = UserDB.get_by_username(username)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return result["user"]

@router.patch("/api/v1/user/{user_id}", response_model=UserResponse)
async def update_user(user_id: str, user_update: UserUpdate):
    result = UserDB.update(user_id, user_update)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return result["user"]

@router.delete("/api/v1/user/{user_id}", status_code=status.HTTP_200_OK)
async def delete_user(user_id: str):
    result = UserDB.delete(user_id)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"message": result["message"]}

@router.get("/api/v1/users/search")
async def search_users(q: str = "", limit: int = 10, offset: int = 0):
    result = UserDB.search(q, limit, offset)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"users": result["users"], "count": result["count"]}

@router.get("/api/v1/test")
async def test_endpoint():
    result = UserDB.test_connection()
    if result["status"] == "error":
        raise HTTPException(status_code=500, detail=result["message"])
    return result

@router.get("/api/v1/user/check/username/{username}")
async def check_username(username: str):
    result = UserDB.username_exists(username)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"exists": result["exists"]}

@router.get("/api/v1/user/check/email/{email}")
async def check_email(email: str):
    result = UserDB.email_exists(email)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"exists": result["exists"]}


@router.post("/api/v1/user/{user_id}/like", status_code=status.HTTP_201_CREATED)
async def like_track(user_id: str, track: TrackInteraction):
    result = UserDB.like_track(user_id, track.track_id)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"message": result["message"]}

@router.delete("/api/v1/user/{user_id}/unlike/{track_id}")
async def unlike_track(user_id: str, track_id: int):
    result = UserDB.unlike_track(user_id, track_id)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return {"message": result["message"]}

@router.get("/api/v1/user/{user_id}/likes")
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