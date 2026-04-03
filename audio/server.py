from fastapi import FastAPI
from app.services.routes import router as api_router

app = FastAPI(
    title="yt-dlp Service",
    version="1.0.0",
)

app.include_router(api_router)

@app.get("/health")
def health():
    return {"status": "ok"}
