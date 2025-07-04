from pydantic import BaseModel
from typing import Optional, Dict, Any, List
from datetime import datetime
import psycopg2
import uuid 
from app.utils.bluppi_db import BluppiDB


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
        with BluppiDB() as db:
            try:
                db.cursor.execute("SELECT 1")
                return {
                    "status": "success",
                    "status_code": 200,
                    "message": "Database connection successful"
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Database connection failed: {str(e)}"
                }

    @staticmethod
    def create(user: UserCreate) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                query = """
                    INSERT INTO users (id, username, email, name, bio, country, phone, profile_pic, favorite_genres)
                    VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s)
                """
                values = (
                    user.id,
                    user.username,
                    user.email,
                    user.name,
                    user.bio,
                    user.country,
                    user.phone,
                    user.profile_pic,
                    user.favorite_genres or []
                )
                db.cursor.execute(query, values)
                db.connection.commit()
                return {
                    "status": "success",
                    "status_code": 201,
                    "message": "User created successfully",
                    "user_id": user.id
                }
            except psycopg2.IntegrityError as e:
                db.connection.rollback()
                if "username" in str(e):
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": "Username already exists"
                    }
                elif "email" in str(e):
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": "Email already exists"
                    }
                else:
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": "User already exists"
                    }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error creating user: {str(e)}"
                }

    @staticmethod
    def get_by_id(user_id: str) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                query = "SELECT * FROM users WHERE id = %s"
                db.cursor.execute(query, (user_id,))
                result = db.cursor.fetchone()
                if result:
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
                            "created_at": result[9],
                            "follower_count": result[10],
                            "following_count": result[11],
                        }
                    }
                else:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "User not found"
                    }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving user: {str(e)}"
                }

    @staticmethod
    def get_by_username(username: str) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                query = "SELECT * FROM users WHERE username = %s"
                db.cursor.execute(query, (username,))
                result = db.cursor.fetchone()
                if result:
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
                            "created_at": result[11]
                        }
                    }
                else:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "User not found"
                    }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving user: {str(e)}"
                }

    @staticmethod
    def update(user_id: str, user_update: UserUpdate) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                update_fields = []
                values = []
                
                for field, value in user_update.dict(exclude_unset=True).items():
                    update_fields.append(f"{field} = %s")
                    values.append(value)
                
                if not update_fields:
                    return {
                        "status": "error",
                        "status_code": 400,
                        "message": "No fields to update"
                    }
                
                values.append(user_id)
                query = f"UPDATE users SET {', '.join(update_fields)} WHERE id = %s"
                
                db.cursor.execute(query, values)
                db.connection.commit()
                
                if db.cursor.rowcount == 0:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "User not found"
                    }
                
                return {
                    "status": "success",
                    "status_code": 200,
                    "message": "User updated successfully"
                }
            except psycopg2.IntegrityError as e:
                db.connection.rollback()
                if "username" in str(e):
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": "Username already exists"
                    }
                elif "email" in str(e):
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": "Email already exists"
                    }
                else:
                    return {
                        "status": "error",
                        "status_code": 409,
                        "message": "Constraint violation"
                    }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error updating user: {str(e)}"
                }

    @staticmethod
    def delete(user_id: str) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                query = "DELETE FROM users WHERE id = %s"
                db.cursor.execute(query, (user_id,))
                db.connection.commit()
                
                if db.cursor.rowcount == 0:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "User not found"
                    }
                
                return {
                    "status": "success",
                    "status_code": 200,
                    "message": "User deleted successfully"
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error deleting user: {str(e)}"
                }

    @staticmethod
    def search(query: str, limit: int = 10, offset: int = 0) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                search_term = f"%{query}%"
                db.cursor.execute(
                    """
                    SELECT id, username, name, profile_pic, follower_count
                    FROM users 
                    WHERE username ILIKE %s OR name ILIKE %s
                    ORDER BY follower_count DESC
                    LIMIT %s OFFSET %s
                    """,
                    (search_term, search_term, limit, offset)
                )
                
                results = db.cursor.fetchall()
                users = []
                
                for row in results:
                    users.append({
                        "id": row[0],
                        "username": row[1],
                        "name": row[2],
                        "profile_pic": row[3],
                        "follower_count": row[4]
                    })
                
                db.cursor.execute(
                    """
                    SELECT COUNT(*) FROM users 
                    WHERE username ILIKE %s OR name ILIKE %s
                    """,
                    (search_term, search_term)
                )
                total_count = db.cursor.fetchone()[0]
                
                return {
                    "status": "success",
                    "status_code": 200,
                    "users": users,
                    "total": total_count
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error searching users: {str(e)}"
                }

    @staticmethod
    def username_exists(username: str) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                query = "SELECT EXISTS(SELECT 1 FROM users WHERE username = %s)"
                db.cursor.execute(query, (username,))
                exists = db.cursor.fetchone()[0]
                return {
                    "status": "success",
                    "status_code": 200,
                    "exists": exists
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error checking username: {str(e)}"
                }

    @staticmethod
    def email_exists(email: str) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                query = "SELECT EXISTS(SELECT 1 FROM users WHERE email = %s)"
                db.cursor.execute(query, (email,))
                exists = db.cursor.fetchone()[0]
                return {
                    "status": "success",
                    "status_code": 200,
                    "exists": exists
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error checking email: {str(e)}"
                }

    @staticmethod
    def follow_user(follower_id: str, followee_id: str) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                if follower_id == followee_id:
                    return {
                        "status": "error",
                        "status_code": 400,
                        "message": "Cannot follow yourself"
                    }
                
                db.cursor.execute("SELECT id FROM users WHERE id = %s", (followee_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "User to follow not found"
                    }
                
                query = """
                    INSERT INTO follows (follower_id, followee_id)
                    VALUES (%s, %s)
                    ON CONFLICT (follower_id, followee_id) DO NOTHING
                """
                db.cursor.execute(query, (follower_id, followee_id))
                db.connection.commit()
                
                if db.cursor.rowcount == 0:
                    return {
                        "status": "success",
                        "status_code": 200,
                        "message": "Already following this user"
                    }
                
                return {
                    "status": "success",
                    "status_code": 201,
                    "message": "User followed successfully"
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error following user: {str(e)}"
                }

    @staticmethod
    def unfollow_user(follower_id: str, followee_id: str) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                query = "DELETE FROM follows WHERE follower_id = %s AND followee_id = %s"
                db.cursor.execute(query, (follower_id, followee_id))
                db.connection.commit()
                
                if db.cursor.rowcount == 0:
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "Not following this user"
                    }
                
                return {
                    "status": "success",
                    "status_code": 200,
                    "message": "User unfollowed successfully"
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error unfollowing user: {str(e)}"
                }

    @staticmethod
    def get_followers(user_id: str, limit: int = 20, offset: int = 0) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                db.cursor.execute(
                    """
                    SELECT u.id, u.username, u.name, u.profile_pic, f.created_at
                    FROM follows f
                    JOIN users u ON f.follower_id = u.id
                    WHERE f.followee_id = %s
                    ORDER BY f.created_at DESC
                    LIMIT %s OFFSET %s
                    """,
                    (user_id, limit, offset)
                )
                
                followers = []
                for row in db.cursor.fetchall():
                    followers.append({
                        "id": row[0],
                        "username": row[1],
                        "name": row[2],
                        "profile_pic": row[3],
                        "followed_at": row[4]
                    })
                
                db.cursor.execute(
                    "SELECT COUNT(*) FROM follows WHERE followee_id = %s",
                    (user_id,)
                )
                total_count = db.cursor.fetchone()[0]
                
                return {
                    "status": "success",
                    "status_code": 200,
                    "followers": followers,
                    "total": total_count
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving followers: {str(e)}"
                }

    @staticmethod
    def get_following(user_id: str, limit: int = 20, offset: int = 0) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                db.cursor.execute(
                    """
                    SELECT u.id, u.username, u.name, u.profile_pic, f.created_at
                    FROM follows f
                    JOIN users u ON f.followee_id = u.id
                    WHERE f.follower_id = %s
                    ORDER BY f.created_at DESC
                    LIMIT %s OFFSET %s
                    """,
                    (user_id, limit, offset)
                )
                
                following = []
                for row in db.cursor.fetchall():
                    following.append({
                        "id": row[0],
                        "username": row[1],
                        "name": row[2],
                        "profile_pic": row[3],
                        "followed_at": row[4]
                    })
                
                db.cursor.execute(
                    "SELECT COUNT(*) FROM follows WHERE follower_id = %s",
                    (user_id,)
                )
                total_count = db.cursor.fetchone()[0]
                
                return {
                    "status": "success",
                    "status_code": 200,
                    "following": following,
                    "total": total_count
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving following: {str(e)}"
                }
    
    @staticmethod
    def is_following(follower_id: str, followee_id: str) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                query = """
                    SELECT EXISTS(
                        SELECT 1 FROM follows 
                        WHERE follower_id = %s AND followee_id = %s
                    )
                """
                db.cursor.execute(query, (follower_id, followee_id))
                is_following = db.cursor.fetchone()[0]
                
                return {
                    "status": "success",
                    "status_code": 200,
                    "is_following": is_following
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error checking follow status: {str(e)}"
                }

    @staticmethod
    def like_track(user_id: str, track_id: uuid.UUID) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found"
                    }

                db.cursor.execute("SELECT id FROM tracks WHERE id = %s", (track_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"Track with ID {track_id} not found"
                    }

                db.cursor.execute(
                    """
                    INSERT INTO user_track (user_id, track_id, interaction_type)
                    VALUES (%s, %s, 'liked')
                    ON CONFLICT (user_id, track_id, interaction_type)
                    DO NOTHING
                    """,
                    (user_id, track_id)
                )
                db.connection.commit()

                if db.cursor.rowcount == 0:
                    return {
                        "status": "success",
                        "status_code": 200,
                        "message": "Track already liked"
                    }

                return {
                    "status": "success",
                    "status_code": 201,
                    "message": "Track liked successfully"
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error liking track: {str(e)}"
                }

    @staticmethod
    def unlike_track(user_id: str, track_id: uuid.UUID) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                db.cursor.execute(
                    """
                    SELECT 1 FROM user_track 
                    WHERE user_id = %s AND track_id = %s AND interaction_type = 'liked'
                    """,
                    (user_id, track_id)
                )
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": "Track not liked"
                    }

                db.cursor.execute(
                    """
                    DELETE FROM user_track 
                    WHERE user_id = %s AND track_id = %s AND interaction_type = 'liked'
                    """,
                    (user_id, track_id)
                )
                db.connection.commit()

                return {
                    "status": "success",
                    "status_code": 200,
                    "message": "Track unliked successfully"
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error unliking track: {str(e)}"
                }

    @staticmethod
    def get_liked_tracks(user_id: str, limit: int = 20, offset: int = 0) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found"
                    }

                db.cursor.execute(
                    """
                    SELECT ut.track_id, ut.interacted_at, 
                           t.title, t.artist, t.album, t.image_url
                    FROM user_track ut
                    JOIN tracks t ON ut.track_id = t.id
                    WHERE ut.user_id = %s AND ut.interaction_type = 'liked'
                    ORDER BY ut.interacted_at DESC
                    LIMIT %s OFFSET %s
                    """,
                    (user_id, limit, offset)
                )

                liked_tracks = []
                for row in db.cursor.fetchall():
                    liked_tracks.append({
                        "track_id": str(row[0]),
                        "liked_at": row[1],
                        "title": row[2],
                        "artist": row[3],
                        "album": row[4],
                        "image_url": row[5]
                    })

                db.cursor.execute(
                    """
                    SELECT COUNT(*) FROM user_track 
                    WHERE user_id = %s AND interaction_type = 'liked'
                    """,
                    (user_id,)
                )
                total_count = db.cursor.fetchone()[0]

                return {
                    "status": "success",
                    "status_code": 200,
                    "tracks": liked_tracks,
                    "total": total_count
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving liked tracks: {str(e)}"
                }

    @staticmethod
    def get_user_stats(user_id: str) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found"
                    }

                db.cursor.execute(
                    """
                    SELECT 
                        (SELECT COUNT(*) FROM user_track WHERE user_id = %s AND interaction_type = 'liked') as liked_tracks,
                        (SELECT COUNT(*) FROM history_tracks WHERE user_id = %s) as total_plays,
                        (SELECT COUNT(*) FROM follows WHERE follower_id = %s) as following_count,
                        (SELECT COUNT(*) FROM follows WHERE followee_id = %s) as followers_count
                    """,
                    (user_id, user_id, user_id, user_id)
                )
                
                result = db.cursor.fetchone()
                
                return {
                    "status": "success",
                    "status_code": 200,
                    "stats": {
                        "liked_tracks": result[0],
                        "total_plays": result[1],
                        "following_count": result[2],
                        "followers_count": result[3]
                    }
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving user stats: {str(e)}"
                }

    @staticmethod
    def add_recent_search(user_id: str, query: str) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found"
                    }

                db.cursor.execute(
                    """
                    INSERT INTO recent_searches (user_id, query)
                    VALUES (%s, %s)
                    """,
                    (user_id, query)
                )
                db.connection.commit()

                return {
                    "status": "success",
                    "status_code": 201,
                    "message": "Search recorded successfully"
                }
            except Exception as e:
                db.connection.rollback()
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error recording search: {str(e)}"
                }

    @staticmethod
    def get_recent_searches(user_id: str, limit: int = 10) -> Dict[str, Any]:
        with BluppiDB() as db:
            try:
                db.cursor.execute("SELECT id FROM users WHERE id = %s", (user_id,))
                if not db.cursor.fetchone():
                    return {
                        "status": "error",
                        "status_code": 404,
                        "message": f"User with ID {user_id} not found"
                    }

                db.cursor.execute(
                    """
                    SELECT query, searched_at
                    FROM recent_searches
                    WHERE user_id = %s
                    ORDER BY searched_at DESC
                    LIMIT %s
                    """,
                    (user_id, limit)
                )

                searches = []
                for row in db.cursor.fetchall():
                    searches.append({
                        "query": row[0],
                        "searched_at": row[1]
                    })

                return {
                    "status": "success",
                    "status_code": 200,
                    "searches": searches
                }
            except Exception as e:
                return {
                    "status": "error",
                    "status_code": 500,
                    "message": f"Error retrieving recent searches: {str(e)}"
                }