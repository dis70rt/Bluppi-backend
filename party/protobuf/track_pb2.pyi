from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Track(_message.Message):
    __slots__ = ("track_id", "title", "artist", "image_url", "duration_ms")
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    TITLE_FIELD_NUMBER: _ClassVar[int]
    ARTIST_FIELD_NUMBER: _ClassVar[int]
    IMAGE_URL_FIELD_NUMBER: _ClassVar[int]
    DURATION_MS_FIELD_NUMBER: _ClassVar[int]
    track_id: str
    title: str
    artist: str
    image_url: str
    duration_ms: int
    def __init__(self, track_id: _Optional[str] = ..., title: _Optional[str] = ..., artist: _Optional[str] = ..., image_url: _Optional[str] = ..., duration_ms: _Optional[int] = ...) -> None: ...

class GetTrackRequest(_message.Message):
    __slots__ = ("track_id",)
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    track_id: str
    def __init__(self, track_id: _Optional[str] = ...) -> None: ...

class SearchTracksRequest(_message.Message):
    __slots__ = ("query", "page_size", "page_token")
    QUERY_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    query: str
    page_size: int
    page_token: str
    def __init__(self, query: _Optional[str] = ..., page_size: _Optional[int] = ..., page_token: _Optional[str] = ...) -> None: ...

class SearchTracksResponse(_message.Message):
    __slots__ = ("tracks", "next_page_token")
    TRACKS_FIELD_NUMBER: _ClassVar[int]
    NEXT_PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    tracks: _containers.RepeatedCompositeFieldContainer[Track]
    next_page_token: str
    def __init__(self, tracks: _Optional[_Iterable[_Union[Track, _Mapping]]] = ..., next_page_token: _Optional[str] = ...) -> None: ...
