from protobuf import track_pb2 as _track_pb2
from protobuf import common_pb2 as _common_pb2
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class SyncMeasurement(_message.Message):
    __slots__ = ("client_send_ms", "server_receive_ms", "server_send_ms", "client_receive_ms", "client_id")
    CLIENT_SEND_MS_FIELD_NUMBER: _ClassVar[int]
    SERVER_RECEIVE_MS_FIELD_NUMBER: _ClassVar[int]
    SERVER_SEND_MS_FIELD_NUMBER: _ClassVar[int]
    CLIENT_RECEIVE_MS_FIELD_NUMBER: _ClassVar[int]
    CLIENT_ID_FIELD_NUMBER: _ClassVar[int]
    client_send_ms: int
    server_receive_ms: int
    server_send_ms: int
    client_receive_ms: int
    client_id: str
    def __init__(self, client_send_ms: _Optional[int] = ..., server_receive_ms: _Optional[int] = ..., server_send_ms: _Optional[int] = ..., client_receive_ms: _Optional[int] = ..., client_id: _Optional[str] = ...) -> None: ...

class SyncResponse(_message.Message):
    __slots__ = ("server_timestamp_ms", "recommended_adjustment", "estimated_offset_ms")
    SERVER_TIMESTAMP_MS_FIELD_NUMBER: _ClassVar[int]
    RECOMMENDED_ADJUSTMENT_FIELD_NUMBER: _ClassVar[int]
    ESTIMATED_OFFSET_MS_FIELD_NUMBER: _ClassVar[int]
    server_timestamp_ms: int
    recommended_adjustment: float
    estimated_offset_ms: int
    def __init__(self, server_timestamp_ms: _Optional[int] = ..., recommended_adjustment: _Optional[float] = ..., estimated_offset_ms: _Optional[int] = ...) -> None: ...

class SyncRequest(_message.Message):
    __slots__ = ("client_send_ms",)
    CLIENT_SEND_MS_FIELD_NUMBER: _ClassVar[int]
    client_send_ms: int
    def __init__(self, client_send_ms: _Optional[int] = ...) -> None: ...

class SyncReply(_message.Message):
    __slots__ = ("server_receive_ms", "server_send_ms")
    SERVER_RECEIVE_MS_FIELD_NUMBER: _ClassVar[int]
    SERVER_SEND_MS_FIELD_NUMBER: _ClassVar[int]
    server_receive_ms: int
    server_send_ms: int
    def __init__(self, server_receive_ms: _Optional[int] = ..., server_send_ms: _Optional[int] = ...) -> None: ...

class PlaybackCommand(_message.Message):
    __slots__ = ("type", "server_timestamp_ms", "position_ms", "track", "playback_rate")
    class CommandType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        PLAY: _ClassVar[PlaybackCommand.CommandType]
        PAUSE: _ClassVar[PlaybackCommand.CommandType]
        SEEK: _ClassVar[PlaybackCommand.CommandType]
        TRACK_CHANGE: _ClassVar[PlaybackCommand.CommandType]
        ADJUST_RATE: _ClassVar[PlaybackCommand.CommandType]
    PLAY: PlaybackCommand.CommandType
    PAUSE: PlaybackCommand.CommandType
    SEEK: PlaybackCommand.CommandType
    TRACK_CHANGE: PlaybackCommand.CommandType
    ADJUST_RATE: PlaybackCommand.CommandType
    TYPE_FIELD_NUMBER: _ClassVar[int]
    SERVER_TIMESTAMP_MS_FIELD_NUMBER: _ClassVar[int]
    POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    PLAYBACK_RATE_FIELD_NUMBER: _ClassVar[int]
    type: PlaybackCommand.CommandType
    server_timestamp_ms: int
    position_ms: int
    track: _track_pb2.Track
    playback_rate: float
    def __init__(self, type: _Optional[_Union[PlaybackCommand.CommandType, str]] = ..., server_timestamp_ms: _Optional[int] = ..., position_ms: _Optional[int] = ..., track: _Optional[_Union[_track_pb2.Track, _Mapping]] = ..., playback_rate: _Optional[float] = ...) -> None: ...

class ClientState(_message.Message):
    __slots__ = ("client_timestamp_ms", "playback_position_ms", "current_track_id", "status", "current_playback_rate", "buffer_health_ms")
    CLIENT_TIMESTAMP_MS_FIELD_NUMBER: _ClassVar[int]
    PLAYBACK_POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    CURRENT_TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    CURRENT_PLAYBACK_RATE_FIELD_NUMBER: _ClassVar[int]
    BUFFER_HEALTH_MS_FIELD_NUMBER: _ClassVar[int]
    client_timestamp_ms: int
    playback_position_ms: int
    current_track_id: str
    status: _common_pb2.PlaybackStatus
    current_playback_rate: float
    buffer_health_ms: int
    def __init__(self, client_timestamp_ms: _Optional[int] = ..., playback_position_ms: _Optional[int] = ..., current_track_id: _Optional[str] = ..., status: _Optional[_Union[_common_pb2.PlaybackStatus, str]] = ..., current_playback_rate: _Optional[float] = ..., buffer_health_ms: _Optional[int] = ...) -> None: ...
