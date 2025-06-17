import uuid
from typing import Union, Optional

def str_to_uuid(value: Union[str, uuid.UUID]) -> Optional[uuid.UUID]:
    if isinstance(value, uuid.UUID):
        return value
    
    if isinstance(value, str):
        try:
            return uuid.UUID(value)
        except ValueError:
            return None   
    return None

def uuid_to_str(value: Union[str, uuid.UUID]) -> Optional[str]:
    if isinstance(value, str):
        try:
            uuid.UUID(value)
            return value
        except ValueError:
            return None
    
    if isinstance(value, uuid.UUID):
        return str(value)
    return None