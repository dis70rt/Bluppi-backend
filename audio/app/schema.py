from pydantic import BaseModel, Field
from typing import Optional

class SearchRequest(BaseModel):
    artist: str = Field(..., example="Charlie Puth")
    title: str = Field(..., example="Attention")
    trackId: str = Field(..., example="Track ID")

class SearchResponse(BaseModel):
    video_id: str = Field(..., example="dQw4w9WgXcQ")
    audio_url: str = Field(
        ...,
        example="https://rr1---sn-abc.googlevideo.com/videoplayback?...",
    )

class AudioResponse(BaseModel):
    video_id: str = Field(..., example="dQw4w9WgXcQ")
    audio_url: Optional[str] = Field(
        None,
        example="https://rr1---sn-abc.googlevideo.com/videoplayback?...",
    )

class ErrorResponse(BaseModel):
    detail: str = Field(..., example="Video not found")
