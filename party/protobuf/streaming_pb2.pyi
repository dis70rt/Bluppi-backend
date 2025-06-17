from protobuf import common_pb2 as _common_pb2
from protobuf import track_pb2 as _track_pb2
from protobuf import user_pb2 as _user_pb2
from protobuf import room_pb2 as _room_pb2
from protobuf import playback_pb2 as _playback_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RoomStreamUpdate(_message.Message):
    __slots__ = ("room_id", "timestamp", "playback_update", "member_update", "room_settings_update", "room_status_update")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    PLAYBACK_UPDATE_FIELD_NUMBER: _ClassVar[int]
    MEMBER_UPDATE_FIELD_NUMBER: _ClassVar[int]
    ROOM_SETTINGS_UPDATE_FIELD_NUMBER: _ClassVar[int]
    ROOM_STATUS_UPDATE_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    timestamp: _timestamp_pb2.Timestamp
    playback_update: PlaybackUpdate
    member_update: MemberUpdate
    room_settings_update: RoomSettingsUpdate
    room_status_update: RoomStatusUpdate
    def __init__(self, room_id: _Optional[str] = ..., timestamp: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., playback_update: _Optional[_Union[PlaybackUpdate, _Mapping]] = ..., member_update: _Optional[_Union[MemberUpdate, _Mapping]] = ..., room_settings_update: _Optional[_Union[RoomSettingsUpdate, _Mapping]] = ..., room_status_update: _Optional[_Union[RoomStatusUpdate, _Mapping]] = ...) -> None: ...

class PlaybackUpdate(_message.Message):
    __slots__ = ("track_change", "play_state", "seek", "skip")
    TRACK_CHANGE_FIELD_NUMBER: _ClassVar[int]
    PLAY_STATE_FIELD_NUMBER: _ClassVar[int]
    SEEK_FIELD_NUMBER: _ClassVar[int]
    SKIP_FIELD_NUMBER: _ClassVar[int]
    track_change: TrackChangeEvent
    play_state: PlayStateEvent
    seek: SeekEvent
    skip: SkipEvent
    def __init__(self, track_change: _Optional[_Union[TrackChangeEvent, _Mapping]] = ..., play_state: _Optional[_Union[PlayStateEvent, _Mapping]] = ..., seek: _Optional[_Union[SeekEvent, _Mapping]] = ..., skip: _Optional[_Union[SkipEvent, _Mapping]] = ...) -> None: ...

class TrackChangeEvent(_message.Message):
    __slots__ = ("current_track", "position_ms")
    CURRENT_TRACK_FIELD_NUMBER: _ClassVar[int]
    POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    current_track: _track_pb2.Track
    position_ms: int
    def __init__(self, current_track: _Optional[_Union[_track_pb2.Track, _Mapping]] = ..., position_ms: _Optional[int] = ...) -> None: ...

class PlayStateEvent(_message.Message):
    __slots__ = ("status", "position_ms")
    STATUS_FIELD_NUMBER: _ClassVar[int]
    POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    status: _common_pb2.PlaybackStatus
    position_ms: int
    def __init__(self, status: _Optional[_Union[_common_pb2.PlaybackStatus, str]] = ..., position_ms: _Optional[int] = ...) -> None: ...

class SeekEvent(_message.Message):
    __slots__ = ("position_ms",)
    POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    position_ms: int
    def __init__(self, position_ms: _Optional[int] = ...) -> None: ...

class SkipEvent(_message.Message):
    __slots__ = ("new_track",)
    NEW_TRACK_FIELD_NUMBER: _ClassVar[int]
    new_track: _track_pb2.Track
    def __init__(self, new_track: _Optional[_Union[_track_pb2.Track, _Mapping]] = ...) -> None: ...

class MemberUpdate(_message.Message):
    __slots__ = ("member_join", "member_leave", "role_change")
    MEMBER_JOIN_FIELD_NUMBER: _ClassVar[int]
    MEMBER_LEAVE_FIELD_NUMBER: _ClassVar[int]
    ROLE_CHANGE_FIELD_NUMBER: _ClassVar[int]
    member_join: MemberJoinEvent
    member_leave: MemberLeaveEvent
    role_change: MemberRoleChangeEvent
    def __init__(self, member_join: _Optional[_Union[MemberJoinEvent, _Mapping]] = ..., member_leave: _Optional[_Union[MemberLeaveEvent, _Mapping]] = ..., role_change: _Optional[_Union[MemberRoleChangeEvent, _Mapping]] = ...) -> None: ...

class MemberJoinEvent(_message.Message):
    __slots__ = ("user", "role")
    USER_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    user: _user_pb2.User
    role: _common_pb2.RoomMemberRole
    def __init__(self, user: _Optional[_Union[_user_pb2.User, _Mapping]] = ..., role: _Optional[_Union[_common_pb2.RoomMemberRole, str]] = ...) -> None: ...

class MemberLeaveEvent(_message.Message):
    __slots__ = ("user_id",)
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    def __init__(self, user_id: _Optional[str] = ...) -> None: ...

class MemberRoleChangeEvent(_message.Message):
    __slots__ = ("user_id", "new_role")
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    NEW_ROLE_FIELD_NUMBER: _ClassVar[int]
    user_id: str
    new_role: _common_pb2.RoomMemberRole
    def __init__(self, user_id: _Optional[str] = ..., new_role: _Optional[_Union[_common_pb2.RoomMemberRole, str]] = ...) -> None: ...

class RoomSettingsUpdate(_message.Message):
    __slots__ = ("name", "description", "visibility", "invite_only")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_FIELD_NUMBER: _ClassVar[int]
    INVITE_ONLY_FIELD_NUMBER: _ClassVar[int]
    name: str
    description: str
    visibility: _common_pb2.RoomVisibility
    invite_only: bool
    def __init__(self, name: _Optional[str] = ..., description: _Optional[str] = ..., visibility: _Optional[_Union[_common_pb2.RoomVisibility, str]] = ..., invite_only: bool = ...) -> None: ...

class RoomStatusUpdate(_message.Message):
    __slots__ = ("status", "reason")
    STATUS_FIELD_NUMBER: _ClassVar[int]
    REASON_FIELD_NUMBER: _ClassVar[int]
    status: _common_pb2.RoomStatus
    reason: str
    def __init__(self, status: _Optional[_Union[_common_pb2.RoomStatus, str]] = ..., reason: _Optional[str] = ...) -> None: ...

class JoinRoomStreamRequest(_message.Message):
    __slots__ = ("room_id", "user_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    user_id: str
    def __init__(self, room_id: _Optional[str] = ..., user_id: _Optional[str] = ...) -> None: ...

class RoomStreamSnapshot(_message.Message):
    __slots__ = ("room_info", "current_playback", "current_track", "members", "member_count")
    ROOM_INFO_FIELD_NUMBER: _ClassVar[int]
    CURRENT_PLAYBACK_FIELD_NUMBER: _ClassVar[int]
    CURRENT_TRACK_FIELD_NUMBER: _ClassVar[int]
    MEMBERS_FIELD_NUMBER: _ClassVar[int]
    MEMBER_COUNT_FIELD_NUMBER: _ClassVar[int]
    room_info: _room_pb2.Room
    current_playback: _playback_pb2.PlaybackState
    current_track: _track_pb2.Track
    members: _containers.RepeatedCompositeFieldContainer[_user_pb2.User]
    member_count: int
    def __init__(self, room_info: _Optional[_Union[_room_pb2.Room, _Mapping]] = ..., current_playback: _Optional[_Union[_playback_pb2.PlaybackState, _Mapping]] = ..., current_track: _Optional[_Union[_track_pb2.Track, _Mapping]] = ..., members: _Optional[_Iterable[_Union[_user_pb2.User, _Mapping]]] = ..., member_count: _Optional[int] = ...) -> None: ...

class StreamRoomUpdatesRequest(_message.Message):
    __slots__ = ("room_id", "user_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    user_id: str
    def __init__(self, room_id: _Optional[str] = ..., user_id: _Optional[str] = ...) -> None: ...
