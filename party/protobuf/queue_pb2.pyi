from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf import empty_pb2 as _empty_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union

DESCRIPTOR: _descriptor.FileDescriptor

class RoomQueueItem(_message.Message):
    __slots__ = ("room_id", "position", "track_id", "added_by", "added_at")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    POSITION_FIELD_NUMBER: _ClassVar[int]
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    ADDED_BY_FIELD_NUMBER: _ClassVar[int]
    ADDED_AT_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    position: int
    track_id: str
    added_by: str
    added_at: _timestamp_pb2.Timestamp
    def __init__(self, room_id: _Optional[str] = ..., position: _Optional[int] = ..., track_id: _Optional[str] = ..., added_by: _Optional[str] = ..., added_at: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class AddToQueueRequest(_message.Message):
    __slots__ = ("room_id", "track_id", "user_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    track_id: str
    user_id: str
    def __init__(self, room_id: _Optional[str] = ..., track_id: _Optional[str] = ..., user_id: _Optional[str] = ...) -> None: ...

class RemoveFromQueueRequest(_message.Message):
    __slots__ = ("room_id", "track_id")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    track_id: str
    def __init__(self, room_id: _Optional[str] = ..., track_id: _Optional[str] = ...) -> None: ...

class GetQueueRequest(_message.Message):
    __slots__ = ("room_id",)
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    def __init__(self, room_id: _Optional[str] = ...) -> None: ...

class GetQueueResponse(_message.Message):
    __slots__ = ("items",)
    ITEMS_FIELD_NUMBER: _ClassVar[int]
    items: _containers.RepeatedCompositeFieldContainer[RoomQueueItem]
    def __init__(self, items: _Optional[_Iterable[_Union[RoomQueueItem, _Mapping]]] = ...) -> None: ...

class ClearQueueRequest(_message.Message):
    __slots__ = ("room_id",)
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    def __init__(self, room_id: _Optional[str] = ...) -> None: ...
