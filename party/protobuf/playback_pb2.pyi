from protobuf import common_pb2 as _common_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class PlaybackState(_message.Message):
    __slots__ = ("room_id", "current_track_id", "position_ms", "status", "updated_at")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    CURRENT_TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    UPDATED_AT_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    current_track_id: str
    position_ms: int
    status: _common_pb2.PlaybackStatus
    updated_at: _timestamp_pb2.Timestamp
    def __init__(self, room_id: _Optional[str] = ..., current_track_id: _Optional[str] = ..., position_ms: _Optional[int] = ..., status: _Optional[_Union[_common_pb2.PlaybackStatus, str]] = ..., updated_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class PlayRequest(_message.Message):
    __slots__ = ("room_id", "track_id", "position_ms")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    track_id: str
    position_ms: int
    def __init__(self, room_id: _Optional[str] = ..., track_id: _Optional[str] = ..., position_ms: _Optional[int] = ...) -> None: ...

class PauseRequest(_message.Message):
    __slots__ = ("room_id",)
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    def __init__(self, room_id: _Optional[str] = ...) -> None: ...

class SeekRequest(_message.Message):
    __slots__ = ("room_id", "position_ms")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    position_ms: int
    def __init__(self, room_id: _Optional[str] = ..., position_ms: _Optional[int] = ...) -> None: ...

class SkipRequest(_message.Message):
    __slots__ = ("room_id",)
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    def __init__(self, room_id: _Optional[str] = ...) -> None: ...

class GetPlaybackStateRequest(_message.Message):
    __slots__ = ("room_id",)
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    def __init__(self, room_id: _Optional[str] = ...) -> None: ...

class SyncPlaybackRequest(_message.Message):
    __slots__ = ("room_id", "user_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    user_id: str
    def __init__(self, room_id: _Optional[str] = ..., user_id: _Optional[str] = ...) -> None: ...
