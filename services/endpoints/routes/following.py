from fastapi import APIRouter, HTTPException, Path, Query, status

from services.database.db_users import UserDB


router = APIRouter(tags=["FollowController"])


@router.post(
    "/api/v1/user/{follower_id}/follow/{followee_id}",
    status_code=status.HTTP_201_CREATED,
)
async def follow_user(
    follower_id: str = Path(..., description="ID of the user who is following"),
    followee_id: str = Path(..., description="ID of the user to be followed"),
):
    result = UserDB.follow_user(follower_id, followee_id)
    if result["status"] == "error":
        raise HTTPException(status_code=result["status_code"], detail=result["message"])
    return {"message": result["message"]}


@router.delete("/api/v1/user/{follower_id}/unfollow/{followee_id}")
async def unfollow_user(
    follower_id: str = Path(..., description="ID of the user who is unfollowing"),
    followee_id: str = Path(..., description="ID of the user to be unfollowed"),
):
    result = UserDB.unfollow_user(follower_id, followee_id)
    if result["status"] == "error":
        raise HTTPException(status_code=result["status_code"], detail=result["message"])
    return {"message": result["message"]}


@router.get("/api/v1/user/{user_id}/followers")
async def get_followers(
    user_id: str, limit: int = Query(20, ge=1, le=100), offset: int = Query(0, ge=0)
):
    result = UserDB.get_followers(user_id, limit, offset)
    if result["status"] == "error":
        raise HTTPException(status_code=result["status_code"], detail=result["message"])
    return {"followers": result["followers"], "total": result["total"]}


@router.get("/api/v1/user/{user_id}/following")
async def get_following(
    user_id: str, limit: int = Query(20, ge=1, le=100), offset: int = Query(0, ge=0)
):
    result = UserDB.get_following(user_id, limit, offset)
    if result["status"] == "error":
        raise HTTPException(status_code=result["status_code"], detail=result["message"])
    return {"following": result["following"], "total": result["total"]}

@router.get("/api/v1/user/{user_id}/is-following/{other_user_id}")
async def is_following(
    user_id: str = Path(..., description="ID of the user"),
    other_user_id: str = Path(..., description="ID of the other user"),
):
    result = UserDB.is_following(user_id, other_user_id)
    if result["status"] == "error":
        raise HTTPException(status_code=result["status_code"], detail=result["message"])
    return {"is_following": result["is_following"]}