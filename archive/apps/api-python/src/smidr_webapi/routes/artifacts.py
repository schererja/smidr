"""Artifact management route handlers."""

from fastapi import APIRouter, HTTPException
import grpc
from smidr_sdk import SmidrClient
from smidr_webapi.models import ListArtifactsResponse, ArtifactItem

router = APIRouter(prefix="/api/builds", tags=["artifacts"])


def get_client() -> SmidrClient:
    """Get SmidrClient instance (dependency injection)."""
    from smidr_webapi.config import config

    return SmidrClient(config.SMIDR_DAEMON_URL)


@router.get(
    "/{build_id}/artifacts",
    response_model=ListArtifactsResponse,
    summary="List artifacts for a specific build",
)
async def list_artifacts(build_id: str):
    """List artifacts for a completed build."""
    client = get_client()
    try:
        response = await client.list_artifacts(build_id)

        artifacts = [
            ArtifactItem(
                name=artifact.name,
                size_bytes=artifact.size_bytes,
                download_url=artifact.download_url,
                checksum=artifact.checksum,
            )
            for artifact in response.artifacts
        ]

        return ListArtifactsResponse(build_id=build_id, artifacts=artifacts)
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.NOT_FOUND:
            raise HTTPException(status_code=404, detail="Build not found") from e
        raise HTTPException(status_code=500, detail=str(e)) from e
    finally:
        await client.close()
