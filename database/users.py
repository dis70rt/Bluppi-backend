from pydantic import BaseModel
from typing import Optional, Dict, Any
from datetime import datetime

from database.config import SynqItDB
from rich import print


class UserBase(BaseModel):
    id: str
    username: str
    name: str
    bio: Optional[str] = None
    avatar_url: Optional[str] = None


class UserCreate(UserBase):
    pass


class UserUpdate(BaseModel):
    username: Optional[str] = None
    name: Optional[str] = None
    bio: Optional[str] = None
    avatar_url: Optional[str] = None


class UserResponse(BaseModel):
    id: str
    username: str
    name: str
    bio: Optional[str] = None
    avatar_url: Optional[str] = None
    created_at: datetime
    updated_at: datetime


class UserDB:
    @staticmethod
    def test_connection():
        try:
            with SynqItDB() as db:
                db.cursor.execute("SELECT 1")
                result = db.cursor.fetchone()
                return {"status": "success", "message": "Database connection successful"}
        except Exception as e:
            return {"status": "error", "message": f"Database connection failed: {str(e)}"}

    @staticmethod
    def create(user: UserCreate) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                db.cursor.execute(
                    "SELECT id FROM users WHERE username = %s", (user.username,)
                )
                if db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": "Username already exists.",
                    }

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user.id,))
                if db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": f"User with ID {user.id} already exists.",
                    }

                query = """
                    INSERT INTO users (
                        id, username, name, bio, avatar_url
                    ) VALUES (%s, %s, %s, %s, %s)
                    RETURNING id, username, name, bio, avatar_url, created_at, updated_at
                """

                db.cursor.execute(
                    query,
                    (user.id, user.username, user.name, user.bio, user.avatar_url),
                )
                db.connection.commit()

                result = db.cursor.fetchone()
                return {
                    "status": "success",
                    "status_code": 201,
                    "user": {
                        "id": result[0],
                        "username": result[1],
                        "name": result[2],
                        "bio": result[3],
                        "avatar_url": result[4],
                        "created_at": result[5],
                        "updated_at": result[6],
                    },
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error creating user: {str(e)}",
                }

    @staticmethod
    def get_by_id(user_id: str) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                query = """
                    SELECT id, username, name, bio, avatar_url, created_at, updated_at 
                    FROM users WHERE id = %s
                """
                db.cursor.execute(query, (user_id,))
                result = db.cursor.fetchone()

                if not result:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found.",
                    }

                return {
                    "status": "success",
                    "status_code": 200,
                    "user": {
                        "id": result[0],
                        "username": result[1],
                        "name": result[2],
                        "bio": result[3],
                        "avatar_url": result[4],
                        "created_at": result[5],
                        "updated_at": result[6],
                    },
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving user: {str(e)}",
                }

    @staticmethod
    def get_by_username(username: str) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                
                query = """
                    SELECT id, username, name, bio, avatar_url, created_at, updated_at 
                    FROM users WHERE username = %s
                """
                db.cursor.execute(query, (username,))
                result = db.cursor.fetchone()

                if not result:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with username '{username}' not found.",
                    }

                return {
                    "status": "success",
                    "status_code": 200,
                    "user": {
                        "id": result[0],
                        "username": result[1],
                        "name": result[2],
                        "bio": result[3],
                        "avatar_url": result[4],
                        "created_at": result[5],
                        "updated_at": result[6],
                    },
                }
            except Exception as e:
                
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving user: {str(e)}",
                }

    @staticmethod
    def update(user_id: str, user_update: UserUpdate) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found.",
                    }

                update_fields = []
                values = []

                if user_update.username is not None:
                    db.cursor.execute(
                        "SELECT id FROM users WHERE username = %s AND id != %s",
                        (user_update.username, user_id),
                    )
                    if db.cursor.fetchone():
                        return {
                            "status": "error",
                            "status_code": 409,
                            "message": "Username already taken.",
                        }
                    update_fields.append("username = %s")
                    values.append(user_update.username)

                if user_update.name is not None:
                    update_fields.append("name = %s")
                    values.append(user_update.name)

                if user_update.bio is not None:
                    update_fields.append("bio = %s")
                    values.append(user_update.bio)

                if user_update.avatar_url is not None:
                    update_fields.append("avatar_url = %s")
                    values.append(user_update.avatar_url)

                update_fields.append("updated_at = CURRENT_TIMESTAMP")

                values.append(user_id)

                if update_fields:
                    query = f"""
                        UPDATE users 
                        SET {', '.join(update_fields)}
                        WHERE id = %s
                        RETURNING id, username, name, bio, avatar_url, created_at, updated_at
                    """

                    db.cursor.execute(query, values)
                    db.connection.commit()

                    result = db.cursor.fetchone()
                    return {
                        "status": "success",
                        "status_code": 200,
                        "user": {
                            "id": result[0],
                            "username": result[1],
                            "name": result[2],
                            "bio": result[3],
                            "avatar_url": result[4],
                            "created_at": result[5],
                            "updated_at": result[6],
                        },
                    }
                else:
                    return {
                        "status": "success",
                        "status_code": 200,
                        "message": "No fields to update.",
                    }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error updating user: {str(e)}",
                }

    @staticmethod
    def delete(user_id: str) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found.",
                    }

                db.cursor.execute("DELETE FROM users WHERE id = %s", (user_id,))
                db.connection.commit()

                return {
                    "status": "success",
                    "status_code": 200,
                    "message": f"User with ID {user_id} successfully deleted.",
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error deleting user: {str(e)}",
                }

    @staticmethod
    def search(query: str, limit: int = 10, offset: int = 0) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                search_term = f"%{query}%"
                db.cursor.execute(
                    """
                    SELECT id, username, name, bio, avatar_url, created_at, updated_at 
                    FROM users 
                    WHERE username ILIKE %s OR name ILIKE %s OR bio ILIKE %s
                    ORDER BY username
                    LIMIT %s OFFSET %s
                    """,
                    (search_term, search_term, search_term, limit, offset),
                )

                results = db.cursor.fetchall()
                users = []

                for row in results:
                    users.append(
                        UserResponse(
                        id=row[0],
                        username=row[1],
                        name=row[2],
                        bio=row[3],
                        avatar_url=row[4],
                        created_at=row[5],
                        updated_at=row[6]
                    ).model_dump()
                    )

                return {
                    "status": "success",
                    "status_code": 200,
                    "users": users,
                    "count": len(users),
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error searching users: {str(e)}",
                }