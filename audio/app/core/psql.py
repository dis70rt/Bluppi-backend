import os
from psycopg_pool import ConnectionPool


class PSQL:
    _pool: ConnectionPool | None = None

    @classmethod
    def init(cls):
        if cls._pool is not None:
            return

        db_host = os.getenv("DB_HOST", "postgres")
        db_port = os.getenv("DB_PORT", "5432")
        db_name = os.getenv("DB_NAME", "bluppi_music")
        db_user = os.getenv("DB_USER", "ethernode")
        db_password = os.getenv("DB_PASSWORD", "password")

        dsn = f"postgresql://{db_user}:{db_password}@{db_host}:{db_port}/{db_name}"

        cls._pool = ConnectionPool(conninfo=dsn)

    @classmethod
    def pool(cls) -> ConnectionPool:
        if cls._pool is None:
            cls.init()
        return cls._pool

    @classmethod
    def update_video_id(cls, track_id: str, video_id: str):
        query = """
            UPDATE tracks
            SET video_id = %s
            WHERE track_id = %s
        """

        with cls.pool().connection() as conn:
            with conn.cursor() as cur:
                cur.execute(query, (video_id, track_id))
            conn.commit()
