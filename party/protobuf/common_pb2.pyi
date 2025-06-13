from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from typing import ClassVar as _ClassVar

DESCRIPTOR: _descriptor.FileDescriptor

class RoomStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    ACTIVE: _ClassVar[RoomStatus]
    INACTIVE: _ClassVar[RoomStatus]
    CLOSED: _ClassVar[RoomStatus]

class RoomVisibility(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PUBLIC: _ClassVar[RoomVisibility]
    PRIVATE: _ClassVar[RoomVisibility]

class RoomMemberRole(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    HOST: _ClassVar[RoomMemberRole]
    MODERATOR: _ClassVar[RoomMemberRole]
    PARTICIPANT: _ClassVar[RoomMemberRole]

class PlaybackStatus(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PLAYING: _ClassVar[PlaybackStatus]
    PAUSED: _ClassVar[PlaybackStatus]

class EventType(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
    __slots__ = ()
    PLAY: _ClassVar[EventType]
    PAUSE: _ClassVar[EventType]
    SEEK: _ClassVar[EventType]
    SKIP: _ClassVar[EventType]
ACTIVE: RoomStatus
INACTIVE: RoomStatus
CLOSED: RoomStatus
PUBLIC: RoomVisibility
PRIVATE: RoomVisibility
HOST: RoomMemberRole
MODERATOR: RoomMemberRole
PARTICIPANT: RoomMemberRole
PLAYING: PlaybackStatus
PAUSED: PlaybackStatus
PLAY: EventType
PAUSE: EventType
SEEK: EventType
SKIP: EventType
