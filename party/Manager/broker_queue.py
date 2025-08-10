from celery import Celery
import os

redis_host = os.getenv("REDIS_HOST", "localhost")
redis_port = int(os.getenv("REDIS_PORT", 6379))
redis_db   = int(os.getenv("REDIS_DB", 0))

queue = Celery(
    "bluppi-queue",
    broker=f"redis://{redis_host}:{redis_port}/0",
    backend=f"redis://{redis_host}:{redis_port}/1"
)
