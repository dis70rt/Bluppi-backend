from pydantic import BaseModel
from typing import Optional, Dict, Any, List
from datetime import datetime

from synqit_db import SynqItDB


class UserBase(BaseModel):
    id: str
    username: str
    email: str
    name: str
    bio: Optional[str] = None
    country: Optional[str] = None
    phone: Optional[str] = None
    profile_pic: Optional[str] = None
    favorite_genres: Optional[List[str]] = None


class UserCreate(UserBase):
    pass


class UserUpdate(BaseModel):
    username: Optional[str] = None
    email: Optional[str] = None
    name: Optional[str] = None
    bio: Optional[str] = None
    country: Optional[str] = None
    phone: Optional[str] = None
    profile_pic: Optional[str] = None
    favorite_genres: Optional[List[str]] = None


class UserResponse(BaseModel):
    id: str
    username: str
    email: str
    name: str
    bio: Optional[str] = None
    country: Optional[str] = None
    phone: Optional[str] = None
    profile_pic: Optional[str] = None
    favorite_genres: List[str] = []
    follower_count: int = 0
    following_count: int = 0
    created_at: datetime


class UserDB:
    @staticmethod
    def test_connection():
        try:
            with SynqItDB() as db:
                db.cursor.execute("SELECT 1")
                result = db.cursor.fetchone()
                return {
                    "status": "success",
                    "message": "Database connection successful",
                }
        except Exception as e:
            return {
                "status": "error",
                "message": f"Database connection failed: {str(e)}",
            }

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
                        "message": "Username already exists",
                    }

                db.cursor.execute(
                    "SELECT id FROM users WHERE email = %s", (user.email,)
                )
                if db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": "Email already exists",
                    }

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user.id,))
                if db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": f"User with ID {user.id} already exists",
                    }

                query = """
                    INSERT INTO users (
                        id, email, username, name, bio, country, phone, profile_pic, favorite_genres
                    ) VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
                    RETURNING id, email, username, name, bio, country, phone, profile_pic, 
                              favorite_genres, follower_count, following_count, created_at
                """

                db.cursor.execute(
                    query,
                    (
                        user.id,
                        user.email,
                        user.username,
                        user.name,
                        user.bio,
                        user.country,
                        user.phone,
                        user.profile_pic,
                        user.favorite_genres or [],
                    ),
                )
                db.connection.commit()

                result = db.cursor.fetchone()
                return {
                    "status": "success",
                    "status_code": 201,
                    "user": {
                        "id": result[0],
                        "email": result[1],
                        "username": result[2],
                        "name": result[3],
                        "bio": result[4],
                        "country": result[5],
                        "phone": result[6],
                        "profile_pic": result[7],
                        "favorite_genres": result[8],
                        "follower_count": result[9],
                        "following_count": result[10],
                        "created_at": result[11],
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
                    SELECT id, email, username, name, bio, country, phone, profile_pic, 
                           favorite_genres, follower_count, following_count, created_at 
                    FROM users WHERE id = %s
                """
                db.cursor.execute(query, (user_id,))
                result = db.cursor.fetchone()

                if not result:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                return {
                    "status": "success",
                    "status_code": 200,
                    "user": {
                        "id": result[0],
                        "email": result[1],
                        "username": result[2],
                        "name": result[3],
                        "bio": result[4],
                        "country": result[5],
                        "phone": result[6],
                        "profile_pic": result[7],
                        "favorite_genres": result[8],
                        "follower_count": result[9],
                        "following_count": result[10],
                        "created_at": result[11],
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
                    SELECT id, email, username, name, bio, country, phone, profile_pic, 
                           favorite_genres, follower_count, following_count, created_at
                    FROM users WHERE username = %s
                """
                db.cursor.execute(query, (username,))
                result = db.cursor.fetchone()

                if not result:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with username '{username}' not found",
                    }

                return {
                    "status": "success",
                    "status_code": 200,
                    "user": {
                        "id": result[0],
                        "email": result[1],
                        "username": result[2],
                        "name": result[3],
                        "bio": result[4],
                        "country": result[5],
                        "phone": result[6],
                        "profile_pic": result[7],
                        "favorite_genres": result[8],
                        "follower_count": result[9],
                        "following_count": result[10],
                        "created_at": result[11],
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
                        "message": f"User with ID {user_id} not found",
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
                            "message": "Username already taken",
                        }
                    update_fields.append("username = %s")
                    values.append(user_update.username)

                if user_update.email is not None:
                    db.cursor.execute(
                        "SELECT id FROM users WHERE email = %s AND id != %s",
                        (user_update.email, user_id),
                    )
                    if db.cursor.fetchone():
                        return {
                            "status": "error",
                            "status_code": 409,
                            "message": "Email already taken",
                        }
                    update_fields.append("email = %s")
                    values.append(user_update.email)

                if user_update.name is not None:
                    update_fields.append("name = %s")
                    values.append(user_update.name)

                if user_update.bio is not None:
                    update_fields.append("bio = %s")
                    values.append(user_update.bio)

                if user_update.country is not None:
                    update_fields.append("country = %s")
                    values.append(user_update.country)

                if user_update.phone is not None:
                    update_fields.append("phone = %s")
                    values.append(user_update.phone)

                if user_update.profile_pic is not None:
                    update_fields.append("profile_pic = %s")
                    values.append(user_update.profile_pic)

                if user_update.favorite_genres is not None:
                    update_fields.append("favorite_genres = %s")
                    values.append(user_update.favorite_genres)

                if not update_fields:
                    return {
                        "status": "success",
                        "status_code": 200,
                        "message": "No fields to update",
                    }

                values.append(user_id)
                query = f"""
                    UPDATE users 
                    SET {', '.join(update_fields)}
                    WHERE id = %s
                    RETURNING id, email, username, name, bio, country, phone, profile_pic, 
                              favorite_genres, follower_count, following_count, created_at
                """

                db.cursor.execute(query, values)
                db.connection.commit()

                result = db.cursor.fetchone()
                return {
                    "status": "success",
                    "status_code": 200,
                    "user": {
                        "id": result[0],
                        "email": result[1],
                        "username": result[2],
                        "name": result[3],
                        "bio": result[4],
                        "country": result[5],
                        "phone": result[6],
                        "profile_pic": result[7],
                        "favorite_genres": result[8],
                        "follower_count": result[9],
                        "following_count": result[10],
                        "created_at": result[11],
                    },
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
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute("DELETE FROM users WHERE id = %s", (user_id,))
                db.connection.commit()

                return {
                    "status": "success",
                    "status_code": 200,
                    "message": f"User with ID {user_id} successfully deleted",
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
                    SELECT id, email, username, name, bio, country, phone, profile_pic, 
                           favorite_genres, follower_count, following_count, created_at
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
                        {
                            "id": row[0],
                            "email": row[1],
                            "username": row[2],
                            "name": row[3],
                            "bio": row[4],
                            "country": row[5],
                            "phone": row[6],
                            "profile_pic": row[7],
                            "favorite_genres": row[8],
                            "follower_count": row[9],
                            "following_count": row[10],
                            "created_at": row[11],
                        }
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

    @staticmethod
    def username_exists(username: str) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                db.cursor.execute(
                    "SELECT id FROM users WHERE username = %s", (username,)
                )
                result = db.cursor.fetchone()
                return {
                    "status": "success",
                    "status_code": 200,
                    "exists": result is not None,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error checking username: {str(e)}",
                }

    @staticmethod
    def email_exists(email: str) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                db.cursor.execute("SELECT id FROM users WHERE email = %s", (email,))
                result = db.cursor.fetchone()
                return {
                    "status": "success",
                    "status_code": 200,
                    "exists": result is not None,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error checking email: {str(e)}",
                }

    @staticmethod
    def follow_user(follower_id: str, followee_id: str) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                if follower_id == followee_id:
                    return {
                        "status": "error",
                        "status_code": 400,
                        "message": "Users cannot follow themselves",
                    }

                db.cursor.execute(
                    "SELECT id FROM users WHERE id IN (%s, %s)",
                    (follower_id, followee_id),
                )
                results = db.cursor.fetchall()
                if len(results) != 2:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "One or both users not found",
                    }

                db.cursor.execute(
                    "SELECT 1 FROM follows WHERE follower_id = %s AND followee_id = %s",
                    (follower_id, followee_id),
                )
                if db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": "Already following this user",
                    }

                db.cursor.execute(
                    "INSERT INTO follows (follower_id, followee_id) VALUES (%s, %s)",
                    (follower_id, followee_id),
                )
                db.connection.commit()

                return {
                    "status": "success",
                    "status_code": 201,
                    "message": "Successfully followed user",
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error following user: {str(e)}",
                }

    @staticmethod
    def unfollow_user(follower_id: str, followee_id: str) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute(
                    "SELECT 1 FROM follows WHERE follower_id = %s AND followee_id = %s",
                    (follower_id, followee_id),
                )
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "Not following this user",
                    }

                db.cursor.execute(
                    "DELETE FROM follows WHERE follower_id = %s AND followee_id = %s",
                    (follower_id, followee_id),
                )
                db.connection.commit()

                return {
                    "status": "success",
                    "status_code": 200,
                    "message": "Successfully unfollowed user",
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error unfollowing user: {str(e)}",
                }

    @staticmethod
    def get_followers(user_id: str, limit: int = 20, offset: int = 0) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute(
                    """
                    SELECT u.id, u.username, u.name, u.profile_pic
                    FROM follows f
                    JOIN users u ON f.follower_id = u.id
                    WHERE f.followee_id = %s
                    ORDER BY f.created_at DESC
                    LIMIT %s OFFSET %s
                    """,
                    (user_id, limit, offset),
                )

                followers = [
                    {
                        "id": row[0],
                        "username": row[1],
                        "name": row[2],
                        "profile_pic": row[3],
                    }
                    for row in db.cursor.fetchall()
                ]

                db.cursor.execute(
                    "SELECT COUNT(*) FROM follows WHERE followee_id = %s", (user_id,)
                )
                total_count = db.cursor.fetchone()[0]

                return {
                    "status": "success",
                    "status_code": 200,
                    "followers": followers,
                    "total": total_count,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving followers: {str(e)}",
                }

    @staticmethod
    def get_following(user_id: str, limit: int = 20, offset: int = 0) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute(
                    """
                    SELECT u.id, u.username, u.name, u.profile_pic
                    FROM follows f
                    JOIN users u ON f.followee_id = u.id
                    WHERE f.follower_id = %s
                    ORDER BY f.created_at DESC
                    LIMIT %s OFFSET %s
                    """,
                    (user_id, limit, offset),
                )

                following = [
                    {
                        "id": row[0],
                        "username": row[1],
                        "name": row[2],
                        "profile_pic": row[3],
                    }
                    for row in db.cursor.fetchall()
                ]

                db.cursor.execute(
                    "SELECT COUNT(*) FROM follows WHERE follower_id = %s", (user_id,)
                )
                total_count = db.cursor.fetchone()[0]

                return {
                    "status": "success",
                    "status_code": 200,
                    "following": following,
                    "total": total_count,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving following: {str(e)}",
                }
    
    @staticmethod
    def is_following(follower_id: str, followee_id: str) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:
                db.cursor.execute(
                    "SELECT 1 FROM follows WHERE follower_id = %s AND followee_id = %s",
                    (follower_id, followee_id),
                )
                result = db.cursor.fetchone()
                return {
                    "status": "success",
                    "status_code": 200,
                    "is_following": result is not None,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error checking following status: {str(e)}",
                }

    @staticmethod
    def like_track(user_id: str, track_id: int) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute(
                    """
                    INSERT INTO user_track (user_id, track_id, interaction_type)
                    VALUES (%s, %s, 'liked')
                    ON CONFLICT (user_id, track_id, interaction_type)
                    DO NOTHING
                    """,
                    (user_id, track_id),
                )
                db.connection.commit()

                if db.cursor.rowcount == 0:
                    return {
                        "status": "success",
                        "status_code": 200,
                        "message": "Track already liked",
                    }

                return {
                    "status": "success",
                    "status_code": 201,
                    "message": "Track liked successfully",
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error liking track: {str(e)}",
                }

    @staticmethod
    def unlike_track(user_id: str, track_id: int) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute(
                    """
                    SELECT 1 FROM user_track 
                    WHERE user_id = %s AND track_id = %s AND interaction_type = 'liked'
                    """,
                    (user_id, track_id),
                )
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "Track not liked",
                    }

                db.cursor.execute(
                    """
                    DELETE FROM user_track 
                    WHERE user_id = %s AND track_id = %s AND interaction_type = 'liked'
                    """,
                    (user_id, track_id),
                )
                db.connection.commit()

                return {
                    "status": "success",
                    "status_code": 200,
                    "message": "Track unliked successfully",
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error unliking track: {str(e)}",
                }

    @staticmethod
    def get_liked_tracks(
        user_id: str, limit: int = 20, offset: int = 0
    ) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute(
                    """
                    SELECT track_id, interacted_at
                    FROM user_track
                    WHERE user_id = %s AND interaction_type = 'liked'
                    ORDER BY interacted_at DESC
                    LIMIT %s OFFSET %s
                    """,
                    (user_id, limit, offset),
                )

                liked_tracks = [
                    {"track_id": row[0], "liked_at": row[1]}
                    for row in db.cursor.fetchall()
                ]

                db.cursor.execute(
                    """
                    SELECT COUNT(*) FROM user_track 
                    WHERE user_id = %s AND interaction_type = 'liked'
                    """,
                    (user_id,),
                )
                total_count = db.cursor.fetchone()[0]

                return {
                    "status": "success",
                    "status_code": 200,
                    "tracks": liked_tracks,
                    "total": total_count,
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving liked tracks: {str(e)}",
                }

    @staticmethod
    def add_recent_search(user_id: str, query: str) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute(
                    """
                    INSERT INTO recent_searches (user_id, query)
                    VALUES (%s, %s)
                    """,
                    (user_id, query),
                )
                db.connection.commit()

                return {
                    "status": "success",
                    "status_code": 201,
                    "message": "Search recorded successfully",
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error recording search: {str(e)}",
                }

    @staticmethod
    def get_recent_searches(user_id: str, limit: int = 10) -> Dict[str, Any]:
        with SynqItDB() as db:
            try:

                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found",
                    }

                db.cursor.execute(
                    """
                    SELECT query, searched_at
                    FROM recent_searches
                    WHERE user_id = %s
                    ORDER BY searched_at DESC
                    LIMIT %s
                    """,
                    (user_id, limit),
                )

                searches = [
                    {"query": row[0], "searched_at": row[1]}
                    for row in db.cursor.fetchall()
                ]

                return {"status": "success", "status_code": 200, "searches": searches}
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving recent searches: {str(e)}",
                }
