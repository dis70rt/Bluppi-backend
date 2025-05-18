

from typing import List, Optional
from fastapi import APIRouter, HTTPException, Query, status
from rich import print_json

from database.users import UserCreate, UserDB, UserResponse, UserUpdate

router = APIRouter(tags=["Users"])

@router.post("/api/v1/user", status_code=status.HTTP_201_CREATED, response_model=UserResponse)
async def create_user(user: UserCreate):
    result = UserDB.create(user)
    if result["status"] == "error":
        raise HTTPException(
            status_code=result["status_code"],
            detail=result["message"]
        )
    return result["user"]

# @router.get("/api/v1/user", response_model=UserResponse)
# async def get_user(
#     id: Optional[str] = Query(
#         None, 
#         title="User ID",
#         description="The userId of the user to retrieve",
#     )):
#     result = UserDB.get_by_id(id)
#     if result["status"] == "error":
#         raise HTTPException(
#             status_code=result["status_code"],
#             detail=result["message"]
#         )
#     return result["user"]

@router.get("/api/v1/user", response_model=UserResponse)
async def get_user_by_username(username: Optional[str] = Query(
        None, 
        title="Username",
        description="The username of the user to retrieve",
    )):
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
    for user in result.get("users", []):
        user["created_at"] = user["created_at"].isoformat()
        user["updated_at"] = user["updated_at"].isoformat()
    
    return {"users": result.get("users", []), "count": result.get("count", 0)}

@router.get("/api/v1/test")
async def test_endpoint():
    return UserDB.test_connection()