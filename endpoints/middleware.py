import json
import time
from fastapi import Request
from fastapi.responses import JSONResponse


async def ping_middleware(request: Request, call_next):
    start_time = time.time()
    response = await call_next(request)
    process_time = round((time.time() - start_time) * 1000, 2)

    if response.headers.get("content-type") == "application/json":
        body = b""
        async for chunk in response.body_iterator:
            body += chunk
        try:
            data = json.loads(body)
            data["ping_ms"] = process_time
            return JSONResponse(content=data, status_code=response.status_code)
        except Exception:
            return response
    else:
        return response
