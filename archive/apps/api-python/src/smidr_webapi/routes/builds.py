"""Build management route handlers."""

from datetime import datetime
from typing import Optional
from fastapi import APIRouter, HTTPException, Query
import grpc
from smidr_sdk import SmidrClient
from smidr_sdk.generated import common_pb2
from smidr_webapi.models import (
    StartBuildRequest,
    BuildStatusResponse,
    ListBuildsResponse,
    BuildListItem,
    CancelBuildResponse,
)

router = APIRouter(prefix="/api/builds", tags=["builds"])


def get_client() -> SmidrClient:
    """Get SmidrClient instance (dependency injection)."""
    from smidr_webapi.config import config

    return SmidrClient(config.SMIDR_DAEMON_URL)


@router.post("", response_model=dict, summary="Start a new Yocto build")
async def start_build(request: StartBuildRequest):
    """Start a new build with the specified configuration."""
    client = get_client()
    try:
        response = await client.start_build(
            config_path=request.config_path,
            target=request.target,
            customer=request.customer,
            force_clean=request.force_clean or False,
            force_image_rebuild=request.force_image_rebuild or False,
            environment_variables=request.environment_variables,
        )
        return {"buildId": response.build_identifier.build_id}
    finally:
        await client.close()


@router.get(
    "/{build_id}", response_model=BuildStatusResponse, summary="Get the status of a specific build"
)
async def get_build_status(build_id: str):
    """Get the current status of a build."""
    client = get_client()
    try:
        status = await client.get_build_status(build_id)

        start_time = None
        if status.timestamps.start_time_unix_seconds > 0:
            start_time = datetime.fromtimestamp(
                status.timestamps.start_time_unix_seconds
            ).isoformat()

        end_time = None
        if status.timestamps.end_time_unix_seconds > 0:
            end_time = datetime.fromtimestamp(status.timestamps.end_time_unix_seconds).isoformat()

        return BuildStatusResponse(
            build_id=status.build_identifier.build_id,
            state=common_pb2.BuildState.Name(status.state),
            start_time=start_time,
            end_time=end_time,
            error_message=status.error_message if status.error_message else None,
            exit_code=status.exit_code if status.exit_code != 0 else None,
        )
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.NOT_FOUND:
            raise HTTPException(status_code=404, detail="Build not found") from e
        raise HTTPException(status_code=500, detail=str(e)) from e
    finally:
        await client.close()


@router.get(
    "", response_model=ListBuildsResponse, summary="List all builds with optional filtering"
)
async def list_builds(
    page_size: Optional[int] = Query(None, description="Maximum number of builds to return"),
    state: Optional[str] = Query(None, description="Filter by build state"),
    customer: Optional[str] = Query(None, description="Filter by customer"),
):
    """List all builds, optionally filtered by state or customer."""
    client = get_client()
    try:
        # Parse state filter if provided
        states = None
        if state:
            try:
                # Convert string to BuildState enum
                build_state = common_pb2.BuildState.Value(f"BUILD_STATE_{state.upper()}")
                states = [build_state]
            except (ValueError, KeyError):
                # Invalid state, ignore filter
                pass

        response = await client.list_builds(
            states=states,
            page_size=page_size or 0,
            customer=customer,
        )

        builds = []
        for b in response.builds:
            start_time = datetime.fromtimestamp(b.timestamps.start_time_unix_seconds).isoformat()

            builds.append(
                BuildListItem(
                    build_id=b.build_identifier.build_id,
                    state=common_pb2.BuildState.Name(b.build_state),
                    start_time=start_time,
                    target=b.target_image,
                )
            )

        return ListBuildsResponse(builds=builds)
    finally:
        await client.close()


@router.delete("/{build_id}", response_model=CancelBuildResponse, summary="Cancel a running build")
async def cancel_build(build_id: str):
    """Cancel a running build."""
    client = get_client()
    try:
        response = await client.cancel_build(build_id)
        return CancelBuildResponse(success=response.success, message=response.message)
    except grpc.RpcError as e:
        if e.code() == grpc.StatusCode.NOT_FOUND:
            raise HTTPException(status_code=404, detail="Build not found") from e
        raise HTTPException(status_code=500, detail=str(e)) from e
    finally:
        await client.close()
