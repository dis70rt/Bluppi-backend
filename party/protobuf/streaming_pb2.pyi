from protobuf import track_pb2 as _track_pb2
from protobuf import common_pb2 as _common_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

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

class HostTrackCommand(_message.Message):
    __slots__ = ("room_id", "track", "start_at_server_ms", "start_position", "host_user_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    TRACK_FIELD_NUMBER: _ClassVar[int]
    START_AT_SERVER_MS_FIELD_NUMBER: _ClassVar[int]
    START_POSITION_FIELD_NUMBER: _ClassVar[int]
    HOST_USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    track: _track_pb2.Track
    start_at_server_ms: int
    start_position: int
    host_user_id: str
    def __init__(self, room_id: _Optional[str] = ..., track: _Optional[_Union[_track_pb2.Track, _Mapping]] = ..., start_at_server_ms: _Optional[int] = ..., start_position: _Optional[int] = ..., host_user_id: _Optional[str] = ...) -> None: ...

class HostPositionUpdate(_message.Message):
    __slots__ = ("room_id", "current_position_ms", "server_timestamp_ms", "host_user_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    CURRENT_POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    SERVER_TIMESTAMP_MS_FIELD_NUMBER: _ClassVar[int]
    HOST_USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    current_position_ms: int
    server_timestamp_ms: int
    host_user_id: str
    def __init__(self, room_id: _Optional[str] = ..., current_position_ms: _Optional[int] = ..., server_timestamp_ms: _Optional[int] = ..., host_user_id: _Optional[str] = ...) -> None: ...

class HostPlaybackCommand(_message.Message):
    __slots__ = ("room_id", "type", "execute_at_server_ms", "position_ms", "host_user_id")
    class CommandType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        PLAY: _ClassVar[HostPlaybackCommand.CommandType]
        PAUSE: _ClassVar[HostPlaybackCommand.CommandType]
        SEEK: _ClassVar[HostPlaybackCommand.CommandType]
        TRACK_CHANGE: _ClassVar[HostPlaybackCommand.CommandType]
        ADJUST_RATE: _ClassVar[HostPlaybackCommand.CommandType]
    PLAY: HostPlaybackCommand.CommandType
    PAUSE: HostPlaybackCommand.CommandType
    SEEK: HostPlaybackCommand.CommandType
    TRACK_CHANGE: HostPlaybackCommand.CommandType
    ADJUST_RATE: HostPlaybackCommand.CommandType
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    EXECUTE_AT_SERVER_MS_FIELD_NUMBER: _ClassVar[int]
    POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    HOST_USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    type: HostPlaybackCommand.CommandType
    execute_at_server_ms: int
    position_ms: int
    host_user_id: str
    def __init__(self, room_id: _Optional[str] = ..., type: _Optional[_Union[HostPlaybackCommand.CommandType, str]] = ..., execute_at_server_ms: _Optional[int] = ..., position_ms: _Optional[int] = ..., host_user_id: _Optional[str] = ...) -> None: ...

class HostCommand(_message.Message):
    __slots__ = ("track_command", "position_update", "control_command")
    TRACK_COMMAND_FIELD_NUMBER: _ClassVar[int]
    POSITION_UPDATE_FIELD_NUMBER: _ClassVar[int]
    CONTROL_COMMAND_FIELD_NUMBER: _ClassVar[int]
    track_command: HostTrackCommand
    position_update: HostPositionUpdate
    control_command: HostPlaybackCommand
    def __init__(self, track_command: _Optional[_Union[HostTrackCommand, _Mapping]] = ..., position_update: _Optional[_Union[HostPositionUpdate, _Mapping]] = ..., control_command: _Optional[_Union[HostPlaybackCommand, _Mapping]] = ...) -> None: ...

class MemberStatus(_message.Message):
    __slots__ = ("room_id", "user_id", "status", "actual_position_ms", "client_timestamp_ms", "error_message")
    class StatusType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        JOINED: _ClassVar[MemberStatus.StatusType]
        TRACK_LOADED: _ClassVar[MemberStatus.StatusType]
        READY_TO_PLAY: _ClassVar[MemberStatus.StatusType]
        PLAYING: _ClassVar[MemberStatus.StatusType]
        PAUSED: _ClassVar[MemberStatus.StatusType]
        SYNCING: _ClassVar[MemberStatus.StatusType]
        SYNCED: _ClassVar[MemberStatus.StatusType]
        DESYNC: _ClassVar[MemberStatus.StatusType]
        ERROR: _ClassVar[MemberStatus.StatusType]
    JOINED: MemberStatus.StatusType
    TRACK_LOADED: MemberStatus.StatusType
    READY_TO_PLAY: MemberStatus.StatusType
    PLAYING: MemberStatus.StatusType
    PAUSED: MemberStatus.StatusType
    SYNCING: MemberStatus.StatusType
    SYNCED: MemberStatus.StatusType
    DESYNC: MemberStatus.StatusType
    ERROR: MemberStatus.StatusType
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    ACTUAL_POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    CLIENT_TIMESTAMP_MS_FIELD_NUMBER: _ClassVar[int]
    ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    user_id: str
    status: MemberStatus.StatusType
    actual_position_ms: int
    client_timestamp_ms: int
    error_message: str
    def __init__(self, room_id: _Optional[str] = ..., user_id: _Optional[str] = ..., status: _Optional[_Union[MemberStatus.StatusType, str]] = ..., actual_position_ms: _Optional[int] = ..., client_timestamp_ms: _Optional[int] = ..., error_message: _Optional[str] = ...) -> None: ...

class ServerBroadcast(_message.Message):
    __slots__ = ("room_id", "type", "track_command", "position_update", "control_command", "new_host_user_id", "affected_user_id")
    class BroadcastType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        TRACK_COMMAND: _ClassVar[ServerBroadcast.BroadcastType]
        POSITION_UPDATE: _ClassVar[ServerBroadcast.BroadcastType]
        CONTROL_COMMAND: _ClassVar[ServerBroadcast.BroadcastType]
        HOST_CHANGED: _ClassVar[ServerBroadcast.BroadcastType]
        MEMBER_JOINED: _ClassVar[ServerBroadcast.BroadcastType]
        MEMBER_LEFT: _ClassVar[ServerBroadcast.BroadcastType]
    TRACK_COMMAND: ServerBroadcast.BroadcastType
    POSITION_UPDATE: ServerBroadcast.BroadcastType
    CONTROL_COMMAND: ServerBroadcast.BroadcastType
    HOST_CHANGED: ServerBroadcast.BroadcastType
    MEMBER_JOINED: ServerBroadcast.BroadcastType
    MEMBER_LEFT: ServerBroadcast.BroadcastType
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    TYPE_FIELD_NUMBER: _ClassVar[int]
    TRACK_COMMAND_FIELD_NUMBER: _ClassVar[int]
    POSITION_UPDATE_FIELD_NUMBER: _ClassVar[int]
    CONTROL_COMMAND_FIELD_NUMBER: _ClassVar[int]
    NEW_HOST_USER_ID_FIELD_NUMBER: _ClassVar[int]
    AFFECTED_USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    type: ServerBroadcast.BroadcastType
    track_command: HostTrackCommand
    position_update: HostPositionUpdate
    control_command: HostPlaybackCommand
    new_host_user_id: str
    affected_user_id: str
    def __init__(self, room_id: _Optional[str] = ..., type: _Optional[_Union[ServerBroadcast.BroadcastType, str]] = ..., track_command: _Optional[_Union[HostTrackCommand, _Mapping]] = ..., position_update: _Optional[_Union[HostPositionUpdate, _Mapping]] = ..., control_command: _Optional[_Union[HostPlaybackCommand, _Mapping]] = ..., new_host_user_id: _Optional[str] = ..., affected_user_id: _Optional[str] = ...) -> None: ...

class ServerResponse(_message.Message):
    __slots__ = ("type", "ready_member_count", "total_member_count", "member_statuses", "error_message")
    class ResponseType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        ACKNOWLEDGED: _ClassVar[ServerResponse.ResponseType]
        MEMBER_COUNT_UPDATE: _ClassVar[ServerResponse.ResponseType]
        MEMBER_STATUS_UPDATE: _ClassVar[ServerResponse.ResponseType]
        ERROR: _ClassVar[ServerResponse.ResponseType]
    ACKNOWLEDGED: ServerResponse.ResponseType
    MEMBER_COUNT_UPDATE: ServerResponse.ResponseType
    MEMBER_STATUS_UPDATE: ServerResponse.ResponseType
    ERROR: ServerResponse.ResponseType
    TYPE_FIELD_NUMBER: _ClassVar[int]
    READY_MEMBER_COUNT_FIELD_NUMBER: _ClassVar[int]
    TOTAL_MEMBER_COUNT_FIELD_NUMBER: _ClassVar[int]
    MEMBER_STATUSES_FIELD_NUMBER: _ClassVar[int]
    ERROR_MESSAGE_FIELD_NUMBER: _ClassVar[int]
    type: ServerResponse.ResponseType
    ready_member_count: int
    total_member_count: int
    member_statuses: _containers.RepeatedCompositeFieldContainer[MemberStatus]
    error_message: str
    def __init__(self, type: _Optional[_Union[ServerResponse.ResponseType, str]] = ..., ready_member_count: _Optional[int] = ..., total_member_count: _Optional[int] = ..., member_statuses: _Optional[_Iterable[_Union[MemberStatus, _Mapping]]] = ..., error_message: _Optional[str] = ...) -> None: ...
