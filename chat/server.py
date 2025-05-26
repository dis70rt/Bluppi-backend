import uuid
import orjson as json

from chat.cassandra import CassandraClient
from datetime import datetime, timezone, timedelta
from fastapi import FastAPI, WebSocket
from fastapi.responses import HTMLResponse, JSONResponse
from fastapi.middleware.cors import CORSMiddleware
from typing import Dict, Set
from uuid import UUID

class ConnectionManager:
    def __init__(self):
        self.user_connections: Dict[str, WebSocket] = {}
        self.conversation_participants: Dict[str, Set[str]] = {}
    
    async def connect(self, websocket: WebSocket, user_id: str):
        await websocket.accept()
        self.user_connections[user_id] = websocket
    
    async def disconnect(self, user_id: str):
        if user_id in self.user_connections:
            del self.user_connections[user_id]

    async def send_to_user(self, user_id: str, message: str):
        if user_id in self.user_connections:
            await self.user_connections[user_id].send_text(json.dumps(message))
    
    async def send_to_conversation(self, conversation_id: str, message: dict):
        if conversation_id in self.conversation_participants:
            task = []
            for user_id in self.conversation_participants[conversation_id]:
                if user_id in self.user_connections:
                    task.append(self.send_to_user(user_id, message))
        

manager = ConnectionManager()
cassandra = CassandraClient()
ist_timezone = timezone(timedelta(hours=5, minutes=30))

app = FastAPI(
    title="Chat Server",
    description="A simple chat server using FastAPI and WebSockets.",
    version="1.0.0",
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
    return {
        "title": app.title,
        "description": app.description,
        "version": app.version
    }

@app.post("/ws/v1/conversations")
async def create_conversation(conversation: dict):
    conversation_id = str(uuid.uuid4())
    is_group = conversation.get("is_group", False)
    conversation_name = conversation.get("conversation_name", None)
    participants = conversation.get("participants", [])

    # TODO: Cassandra implementaion later, to write and read


@app.websocket("/ws/v1/{user_id}")
async def websocket_endpoint(websocket: WebSocket, user_id: str):
    await manager.connect(websocket, user_id)
    try:
        while True:
            data_text = await websocket.receive_text()
            data = json.loads(data_text)

            if data["type"] == "message":
                conversation_id = data["conversation_id"]
                message_id = str(uuid.uuid4())
                message_text = data["message"]
                now = datetime.now(ist_timezone)

                cassandra.insert_message(
                    conversation_id=UUID(conversation_id),
                    message_id=UUID(message_id),
                    sender_id=UUID(user_id),
                    message_text=message_text,
                    created_at=now
                )

                response = {
                    "type": "message",
                    "conversation_id": conversation_id,
                    "message_id": message_id,
                    "sender_id": user_id,
                    "message_text": message_text,
                    "created_at": now.isoformat()
                }

                await manager.send_to_conversation(conversation_id, response)

                for participant in manager.conversation_participants.get(conversation_id, set()):
                    status = "sent"
                    if participant == user_id:
                        status = "seen"
                    
                    cassandra.insert_message_status(
                        message_id=UUID(message_id),
                        user_id=UUID(participant),
                        status=status,
                        )
            
            elif data["type"] == "status_update":
                message_id = data["message_id"]
                status = data["status"]
                
                cassandra.update_message_status(
                    message_id=UUID(message_id),
                    user_id=UUID(user_id),
                    status=status,
                    updated_at=datetime.now(ist_timezone)
                )

                msg = cassandra.select_message_info(message_id=UUID(message_id))
                if msg:
                    await manager.send_to_user(
                        str(msg.sender_id), 
                        {
                            "type": "status_update",
                            "message_id": message_id,
                            "user_id": user_id,
                            "status": status,
                            "conversation_id": str(msg.conversation_id),
                        }
                    )
            
    except Exception as e:
        print(f"WebSocket error for user {user_id}: {e}")
    finally:
        await manager.disconnect(user_id)    

                
