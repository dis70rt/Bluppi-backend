import uuid
import orjson as json
import asyncio
from rich import print_json, print

from datetime import datetime
from fastapi import FastAPI, Query, WebSocket
from fastapi.responses import JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from typing import Dict, Optional, Set
from uuid import UUID
from contextlib import asynccontextmanager
from .psql_client import PostgresClient, IST_TIMEZONE


@asynccontextmanager
async def lifespan(app: FastAPI):
    try:
        yield
    finally:
        for user_id in list(manager.user_connections.keys()):
            await manager.disconnect(user_id)

class ConnectionManager:
    def __init__(self):
        self.user_connections: Dict[str, WebSocket] = {}
        self.conversation_participants: Dict[str, Set[str]] = {}

    async def connect(self, websocket: WebSocket, user_id: str):
        await websocket.accept()
        self.user_connections[user_id] = websocket

        conversations = postgres.select_user_conversations(user_id=user_id)
        for row in conversations:
            conv_id = row["conversation_id"] 
            conv_str = str(conv_id)
            if conv_str not in self.conversation_participants:
                self.conversation_participants[conv_str] = set()
            self.conversation_participants[conv_str].add(user_id)

    async def disconnect(self, user_id: str):
        if user_id in self.user_connections:
            del self.user_connections[user_id]

    async def send_to_user(self, user_id: str, message: dict):
        if user_id in self.user_connections:
            await self.user_connections[user_id].send_json(message)

    async def send_to_conversation(self, conversation_id: str, message: dict, exclude_user: str = None):
        if conversation_id in self.conversation_participants:
            tasks = []
            for user_id in self.conversation_participants[conversation_id]:
                if user_id in self.user_connections: #and user_id != exclude_user:
                    tasks.append(self.send_to_user(user_id, message))

            if tasks:
                await asyncio.gather(*tasks)


manager = ConnectionManager()
postgres = PostgresClient()

app = FastAPI(
    title="Chat Server",
    description="A simple chat server using FastAPI and WebSockets.",
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

@app.get("/", response_class=JSONResponse)
async def get():
    return {"title": app.title, "description": app.description, "version": app.version}


@app.post("/ws/v1/conversations")
async def create_conversation(conversation: dict):
    is_group = conversation.get("is_group", False)
    conversation_name = conversation.get("conversation_name", None)
    participants = conversation.get("participants", [])

    if participants is None or len(participants) < 2:
        return JSONResponse(
            {"error": "At least 2 participant is required."},
            status_code=400,
        )
    
    partStr = sorted(participants)
    blob = "-".join(partStr)
    conversation_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, blob))

    postgres.insert_conversations(
        conversation_id=UUID(conversation_id),
        conversation_name=conversation_name,
        is_group=is_group,
    )

    for user_id in participants:
        postgres.insert_conversation_participants(
            conversation_id=UUID(conversation_id), user_id=user_id
        )

    return JSONResponse(
        {
            "conversation_id": conversation_id,
            "status": "created",
            "code": 201,
        }
    )

@app.get("/conversations/{conversation_id}/messages")
async def get_messages(
    conversation_id: str, before_id: Optional[str] = None, limit: int = 20
):
    rows = postgres.select_messages(
        conversation_id=UUID(conversation_id),
        before_id=UUID(before_id) if before_id else None,
        limit=limit,
    )
    
    messages = []
    for row in rows:
        messages.append(
            {
                "message_id": str(row["message_id"]),
                "sender_id": str(row["sender_id"]),
                "message_text": row["message_text"],
                "created_at": row["created_at"].isoformat(),
                "type": row["type"],
            }
        )

    return {"messages": messages, "has_more": len(messages) == limit}

@app.get("/conversations/{conversation_id}/updates")
async def get_updates(conversation_id: str, since_timestamp: str):
    since_dt = datetime.fromisoformat(since_timestamp)

    rows = postgres.select_messages_since(
        conversation_id=UUID(conversation_id), since_timestamp=since_dt
    )

    new_messages = []
    for row in rows:
        new_messages.append(
            {
                "message_id": str(row.message_id),
                "sender_id": str(row.sender_id),
                "message_text": row.message_text,
                "created_at": row.created_at.isoformat(),
                "type": row.type
            }
        )

    return {"new_messages": new_messages}

@app.post("/ws/v1/conversations/get-or-create")
async def get_or_create_conversation(conversation: dict):
    participants = conversation.get("participants", [])
    conversation_name = conversation.get("conversation_name", None)
    is_group = conversation.get("is_group", False)
    
    if participants is None or len(participants) < 2:
        return JSONResponse(
            {"error": "At least 2 participants are required."},
            status_code=400,
        )
    
    partStr = sorted(participants)
    blob = "-".join(partStr)
    conversation_id = str(uuid.uuid5(uuid.NAMESPACE_DNS, blob))
    
    exists = postgres.conversation_exists(UUID(conversation_id))
    
    if not exists:
        postgres.insert_conversations(
            conversation_id=UUID(conversation_id),
            conversation_name=conversation_name,
            is_group=is_group,
        )
        
        for user_id in participants:
            postgres.insert_conversation_participants(
                conversation_id=UUID(conversation_id), user_id=user_id
            )
    
    # last_activity = datetime.now(IST_TIMEZONE)
    # last_message_id = None
    
    # rows = postgres.select_messages(
    #     conversation_id=UUID(conversation_id),
    #     limit=1,
    # )
    
    # if rows and len(rows) > 0:
    #     last_activity = rows[0]["created_at"]
    #     last_message_id = str(rows[0]["message_id"])
    
    return {
        "conversation_id": conversation_id,
        "conversation_name": conversation_name,
        "is_group": is_group,
        "participants": participants,
        "is_new": not exists
        # "last_activity": {
        #     "timestamp": last_activity.isoformat(),
        #     "message_id": last_message_id
        # }
    }    

@app.get("/{user_id}/conversations/")
async def get_conversations(user_id: str):
    rows = postgres.get_user_conversations(user_id=user_id)
    return {"conversations": rows}

@app.websocket("/ws/v1/{user_id}/{conversation_id}")
async def websocket_endpoint(websocket: WebSocket, user_id: str, conversation_id: str):
    await manager.connect(websocket, user_id)

    if conversation_id not in manager.conversation_participants:
        manager.conversation_participants[conversation_id] = set()
    manager.conversation_participants[conversation_id].add(user_id)

    try:
        while True:
            data_text = await websocket.receive_text()
            data = json.loads(data_text)

            if data["type"] in ["text", "track"]:
                message_id = data["message_id"]
                message_text = data["message"]
                message_type = data.get("type", "text")
                now = datetime.now(IST_TIMEZONE)

                postgres.insert_message(
                    conversation_id=UUID(conversation_id),
                    message_id=UUID(message_id),
                    sender_id=user_id,
                    message_text=message_text,
                    created_at=now,
                    type=message_type,
                )

                response = {
                    "type": message_type,
                    "conversation_id": conversation_id,
                    "message_id": message_id,
                    "sender_id": user_id,
                    "message_text": message_text,
                    "created_at": now.isoformat(),
                }

                await manager.send_to_conversation(conversation_id, response, exclude_user=user_id)

                for participant in manager.conversation_participants.get(
                    conversation_id, set()
                ):
                    status = "sent"
                    if participant == user_id:
                        status = "seen"

                    postgres.insert_message_status(
                        message_id=UUID(message_id),
                        user_id=participant,
                        status=status,
                    )

            elif data["type"] == "status_update":
                message_id = data["message_id"]
                status = data["status"]

                postgres.update_message_status(
                    message_id=UUID(message_id),
                    user_id=user_id,
                    status=status,
                    updated_at=datetime.now(IST_TIMEZONE),
                )

                msg = postgres.select_message_info(message_id=UUID(message_id))
                if msg:
                    await manager.send_to_user(
                        str(msg["sender_id"]),
                        {
                            "type": "status_update",
                            "message_id": message_id,
                            "user_id": user_id,
                            "status": status,
                            "conversation_id": str(msg["conversation_id"]),
                        },
                    )

    except Exception as e:
        print(f"[red]WebSocket error for user {user_id}: {e}[/red]")
    finally:
        await manager.disconnect(user_id)

