"""Log streaming route handlers."""

import json
from fastapi import APIRouter, HTTPException, Request
from fastapi.responses import StreamingResponse
import grpc
from sse_starlette.sse import EventSourceResponse
from smidr_sdk import SmidrClient

router = APIRouter(prefix="/api/builds", tags=["logs"])


def get_client() -> SmidrClient:
    """Get SmidrClient instance (dependency injection)."""
    from smidr_webapi.config import config

    return SmidrClient(config.SMIDR_DAEMON_URL)


async def stream_logs_generator(build_id: str, follow: bool, request: Request):
    """Generator function for streaming logs via Server-Sent Events."""
    client = get_client()
    try:
        async for log_entry in client.stream_logs(build_id, follow=follow):
            # Check if client disconnected
            if await request.is_disconnected():
                break

            # Format as SSE event
            data = {
                "stream": log_entry.stream,
                "message": log_entry.message,
                "timestamp": log_entry.timestamp_unix_seconds,
            }
            yield {
                "event": "log",
                "data": json.dumps(data),
            }
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.NOT_FOUND:
            yield {
                "event": "error",
                "data": json.dumps({"error": "Build not found"}),
            }
        else:
            yield {
                "event": "error",
                "data": json.dumps({"error": str(e)}),
            }
    except Exception as e:
        yield {
            "event": "error",
            "data": json.dumps({"error": str(e)}),
        }
    finally:
        await client.close()


@router.get("/{build_id}/logs", summary="Stream build logs in real-time using Server-Sent Events")
async def stream_build_logs(
    build_id: str, follow: bool = False, request: Request = None
):  # noqa: E501
    """Stream build logs in real-time using Server-Sent Events (SSE).

    Note: This endpoint is not fully documented in OpenAPI/Swagger as SSE
    doesn't work well with standard OpenAPI documentation.
    """
    return EventSourceResponse(stream_logs_generator(build_id, follow, request))
