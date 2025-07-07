from protobuf import common_pb2 as _common_pb2
from protobuf import track_pb2 as _track_pb2
from protobuf import playback_pb2 as _playback_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class Room(_message.Message):
    __slots__ = ("id", "name", "description", "room_code", "status", "visibility", "invite_only", "members", "host_user_id", "playback_state", "current_track", "current_position_ms", "last_position_update", "track_start_time", "is_playing")
    ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    ROOM_CODE_FIELD_NUMBER: _ClassVar[int]
    STATUS_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_FIELD_NUMBER: _ClassVar[int]
    INVITE_ONLY_FIELD_NUMBER: _ClassVar[int]
    MEMBERS_FIELD_NUMBER: _ClassVar[int]
    HOST_USER_ID_FIELD_NUMBER: _ClassVar[int]
    PLAYBACK_STATE_FIELD_NUMBER: _ClassVar[int]
    CURRENT_TRACK_FIELD_NUMBER: _ClassVar[int]
    CURRENT_POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    LAST_POSITION_UPDATE_FIELD_NUMBER: _ClassVar[int]
    TRACK_START_TIME_FIELD_NUMBER: _ClassVar[int]
    IS_PLAYING_FIELD_NUMBER: _ClassVar[int]
    id: str
    name: str
    description: str
    room_code: str
    status: _common_pb2.RoomStatus
    visibility: _common_pb2.RoomVisibility
    invite_only: bool
    members: _containers.RepeatedCompositeFieldContainer[RoomMember]
    host_user_id: str
    playback_state: _playback_pb2.PlaybackState
    current_track: _track_pb2.Track
    current_position_ms: int
    last_position_update: int
    track_start_time: int
    is_playing: bool
    def __init__(self, id: _Optional[str] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., room_code: _Optional[str] = ..., status: _Optional[_Union[_common_pb2.RoomStatus, str]] = ..., visibility: _Optional[_Union[_common_pb2.RoomVisibility, str]] = ..., invite_only: bool = ..., members: _Optional[_Iterable[_Union[RoomMember, _Mapping]]] = ..., host_user_id: _Optional[str] = ..., playback_state: _Optional[_Union[_playback_pb2.PlaybackState, _Mapping]] = ..., current_track: _Optional[_Union[_track_pb2.Track, _Mapping]] = ..., current_position_ms: _Optional[int] = ..., last_position_update: _Optional[int] = ..., track_start_time: _Optional[int] = ..., is_playing: bool = ...) -> None: ...

class CreateRoomRequest(_message.Message):
    __slots__ = ("name", "description", "room_code", "visibility", "invite_only", "host_user_id")
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    ROOM_CODE_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_FIELD_NUMBER: _ClassVar[int]
    INVITE_ONLY_FIELD_NUMBER: _ClassVar[int]
    HOST_USER_ID_FIELD_NUMBER: _ClassVar[int]
    name: str
    description: str
    room_code: str
    visibility: _common_pb2.RoomVisibility
    invite_only: bool
    host_user_id: str
    def __init__(self, name: _Optional[str] = ..., description: _Optional[str] = ..., room_code: _Optional[str] = ..., visibility: _Optional[_Union[_common_pb2.RoomVisibility, str]] = ..., invite_only: bool = ..., host_user_id: _Optional[str] = ...) -> None: ...

class GetRoomRequest(_message.Message):
    __slots__ = ("room_id",)
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    def __init__(self, room_id: _Optional[str] = ...) -> None: ...

class GetRoomByCodeRequest(_message.Message):
    __slots__ = ("room_code",)
    ROOM_CODE_FIELD_NUMBER: _ClassVar[int]
    room_code: str
    def __init__(self, room_code: _Optional[str] = ...) -> None: ...

class UpdateRoomRequest(_message.Message):
    __slots__ = ("room_id", "name", "description", "visibility", "invite_only")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    NAME_FIELD_NUMBER: _ClassVar[int]
    DESCRIPTION_FIELD_NUMBER: _ClassVar[int]
    VISIBILITY_FIELD_NUMBER: _ClassVar[int]
    INVITE_ONLY_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    name: str
    description: str
    visibility: _common_pb2.RoomVisibility
    invite_only: bool
    def __init__(self, room_id: _Optional[str] = ..., name: _Optional[str] = ..., description: _Optional[str] = ..., visibility: _Optional[_Union[_common_pb2.RoomVisibility, str]] = ..., invite_only: bool = ...) -> None: ...

class DeleteRoomRequest(_message.Message):
    __slots__ = ("room_id",)
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    def __init__(self, room_id: _Optional[str] = ...) -> None: ...

class ListRoomsRequest(_message.Message):
    __slots__ = ("visibility_filter", "host_user_id_filter", "include_private_rooms_if_member", "page_size", "page_token")
    VISIBILITY_FILTER_FIELD_NUMBER: _ClassVar[int]
    HOST_USER_ID_FILTER_FIELD_NUMBER: _ClassVar[int]
    INCLUDE_PRIVATE_ROOMS_IF_MEMBER_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    visibility_filter: _common_pb2.RoomVisibility
    host_user_id_filter: str
    include_private_rooms_if_member: bool
    page_size: int
    page_token: str
    def __init__(self, visibility_filter: _Optional[_Union[_common_pb2.RoomVisibility, str]] = ..., host_user_id_filter: _Optional[str] = ..., include_private_rooms_if_member: bool = ..., page_size: _Optional[int] = ..., page_token: _Optional[str] = ...) -> None: ...

class ListRoomsResponse(_message.Message):
    __slots__ = ("rooms", "next_page_token")
    ROOMS_FIELD_NUMBER: _ClassVar[int]
    NEXT_PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    rooms: _containers.RepeatedCompositeFieldContainer[Room]
    next_page_token: str
    def __init__(self, rooms: _Optional[_Iterable[_Union[Room, _Mapping]]] = ..., next_page_token: _Optional[str] = ...) -> None: ...

class JoinRoomRequest(_message.Message):
    __slots__ = ("room_id", "room_code", "user_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    ROOM_CODE_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    room_code: str
    user_id: str
    def __init__(self, room_id: _Optional[str] = ..., room_code: _Optional[str] = ..., user_id: _Optional[str] = ...) -> None: ...

class LeaveRoomRequest(_message.Message):
    __slots__ = ("room_id", "user_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    user_id: str
    def __init__(self, room_id: _Optional[str] = ..., user_id: _Optional[str] = ...) -> None: ...

class InviteToRoomRequest(_message.Message):
    __slots__ = ("room_id", "inviter_user_id", "invitee_user_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    INVITER_USER_ID_FIELD_NUMBER: _ClassVar[int]
    INVITEE_USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    inviter_user_id: str
    invitee_user_id: str
    def __init__(self, room_id: _Optional[str] = ..., inviter_user_id: _Optional[str] = ..., invitee_user_id: _Optional[str] = ...) -> None: ...

class KickFromRoomRequest(_message.Message):
    __slots__ = ("room_id", "performing_user_id", "kicked_user_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    PERFORMING_USER_ID_FIELD_NUMBER: _ClassVar[int]
    KICKED_USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    performing_user_id: str
    kicked_user_id: str
    def __init__(self, room_id: _Optional[str] = ..., performing_user_id: _Optional[str] = ..., kicked_user_id: _Optional[str] = ...) -> None: ...

class RoomMember(_message.Message):
    __slots__ = ("room_id", "user_id", "role", "invited_by", "joined_at", "left_at")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    ROLE_FIELD_NUMBER: _ClassVar[int]
    INVITED_BY_FIELD_NUMBER: _ClassVar[int]
    JOINED_AT_FIELD_NUMBER: _ClassVar[int]
    LEFT_AT_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    user_id: str
    role: _common_pb2.RoomMemberRole
    invited_by: str
    joined_at: _timestamp_pb2.Timestamp
    left_at: _timestamp_pb2.Timestamp
    def __init__(self, room_id: _Optional[str] = ..., user_id: _Optional[str] = ..., role: _Optional[_Union[_common_pb2.RoomMemberRole, str]] = ..., invited_by: _Optional[str] = ..., joined_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., left_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...
