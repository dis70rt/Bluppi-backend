from protobuf import common_pb2 as _common_pb2
from google.protobuf import timestamp_pb2 as _timestamp_pb2
from google.protobuf.internal import containers as _containers
from google.protobuf.internal import enum_type_wrapper as _enum_type_wrapper
from google.protobuf import descriptor as _descriptor
from google.protobuf import message as _message
from collections.abc import Iterable as _Iterable, Mapping as _Mapping
from typing import ClassVar as _ClassVar, Optional as _Optional, Union as _Union
import datetime

DESCRIPTOR: _descriptor.FileDescriptor

class PlaybackEventLog(_message.Message):
    __slots__ = ("event_id", "room_id", "user_id", "event_type", "event_payload", "timestamp")
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    EVENT_PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    event_id: int
    room_id: str
    user_id: str
    event_type: _common_pb2.EventType
    event_payload: str
    timestamp: _timestamp_pb2.Timestamp
    def __init__(self, event_id: _Optional[int] = ..., room_id: _Optional[str] = ..., user_id: _Optional[str] = ..., event_type: _Optional[_Union[_common_pb2.EventType, str]] = ..., event_payload: _Optional[str] = ..., timestamp: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class StreamRoomEventsRequest(_message.Message):
    __slots__ = ("room_id",)
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    def __init__(self, room_id: _Optional[str] = ...) -> None: ...

class RoomEvent(_message.Message):
    __slots__ = ("event_id", "room_id", "category", "user_id", "details_json", "timestamp")
    class EventCategory(int, metaclass=_enum_type_wrapper.EnumTypeWrapper):
        __slots__ = ()
        EVENT_CATEGORY_UNSPECIFIED: _ClassVar[RoomEvent.EventCategory]
        USER_JOINED: _ClassVar[RoomEvent.EventCategory]
        USER_LEFT: _ClassVar[RoomEvent.EventCategory]
        ROOM_SETTINGS_UPDATED: _ClassVar[RoomEvent.EventCategory]
    EVENT_CATEGORY_UNSPECIFIED: RoomEvent.EventCategory
    USER_JOINED: RoomEvent.EventCategory
    USER_LEFT: RoomEvent.EventCategory
    ROOM_SETTINGS_UPDATED: RoomEvent.EventCategory
    EVENT_ID_FIELD_NUMBER: _ClassVar[int]
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    CATEGORY_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    DETAILS_JSON_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    event_id: str
    room_id: str
    category: RoomEvent.EventCategory
    user_id: str
    details_json: str
    timestamp: _timestamp_pb2.Timestamp
    def __init__(self, event_id: _Optional[str] = ..., room_id: _Optional[str] = ..., category: _Optional[_Union[RoomEvent.EventCategory, str]] = ..., user_id: _Optional[str] = ..., details_json: _Optional[str] = ..., timestamp: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class StreamPlaybackEventsRequest(_message.Message):
    __slots__ = ("room_id",)
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    def __init__(self, room_id: _Optional[str] = ...) -> None: ...

class PlaybackEvent(_message.Message):
    __slots__ = ("room_id", "user_id", "event_type", "track_id", "position_ms", "timestamp")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    TRACK_ID_FIELD_NUMBER: _ClassVar[int]
    POSITION_MS_FIELD_NUMBER: _ClassVar[int]
    TIMESTAMP_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    user_id: str
    event_type: _common_pb2.EventType
    track_id: str
    position_ms: int
    timestamp: _timestamp_pb2.Timestamp
    def __init__(self, room_id: _Optional[str] = ..., user_id: _Optional[str] = ..., event_type: _Optional[_Union[_common_pb2.EventType, str]] = ..., track_id: _Optional[str] = ..., position_ms: _Optional[int] = ..., timestamp: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ...) -> None: ...

class LogEventRequest(_message.Message):
    __slots__ = ("room_id", "user_id", "event_type", "event_payload")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    USER_ID_FIELD_NUMBER: _ClassVar[int]
    EVENT_TYPE_FIELD_NUMBER: _ClassVar[int]
    EVENT_PAYLOAD_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    user_id: str
    event_type: _common_pb2.EventType
    event_payload: str
    def __init__(self, room_id: _Optional[str] = ..., user_id: _Optional[str] = ..., event_type: _Optional[_Union[_common_pb2.EventType, str]] = ..., event_payload: _Optional[str] = ...) -> None: ...

class GetEventHistoryRequest(_message.Message):
    __slots__ = ("room_id", "start_time", "end_time", "page_size", "page_token")
    ROOM_ID_FIELD_NUMBER: _ClassVar[int]
    START_TIME_FIELD_NUMBER: _ClassVar[int]
    END_TIME_FIELD_NUMBER: _ClassVar[int]
    PAGE_SIZE_FIELD_NUMBER: _ClassVar[int]
    PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    room_id: str
    start_time: _timestamp_pb2.Timestamp
    end_time: _timestamp_pb2.Timestamp
    page_size: int
    page_token: str
    def __init__(self, room_id: _Optional[str] = ..., start_time: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., end_time: _Optional[_Union[datetime.datetime, _timestamp_pb2.Timestamp, _Mapping]] = ..., page_size: _Optional[int] = ..., page_token: _Optional[str] = ...) -> None: ...

class GetEventHistoryResponse(_message.Message):
    __slots__ = ("events", "next_page_token")
    EVENTS_FIELD_NUMBER: _ClassVar[int]
    NEXT_PAGE_TOKEN_FIELD_NUMBER: _ClassVar[int]
    events: _containers.RepeatedCompositeFieldContainer[PlaybackEventLog]
    next_page_token: str
    def __init__(self, events: _Optional[_Iterable[_Union[PlaybackEventLog, _Mapping]]] = ..., next_page_token: _Optional[str] = ...) -> None: ...
