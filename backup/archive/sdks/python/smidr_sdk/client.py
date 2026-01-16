"""SmidrClient - Python client for interacting with the Smidr gRPC daemon."""

from __future__ import annotations

import grpc
from typing import AsyncIterator, Optional, List

# Import generated protobuf code
# The protobuf files are generated in the generated/ subdirectory
from smidr_sdk.generated import (
    builds_pb2,
    builds_pb2_grpc,
    logs_pb2,
    logs_pb2_grpc,
    artifacts_pb2,
    artifacts_pb2_grpc,
    common_pb2,
)


class SmidrClient:
    """Client for interacting with the Smidr gRPC daemon."""

    def __init__(self, address: str):
        """Create a new SmidrClient connected to the specified daemon address.

        Args:
            address: The daemon address (e.g., "localhost:50051")
        """
        # Remove http:// or https:// prefix if present
        if address.startswith("http://"):
            address = address[7:]
        elif address.startswith("https://"):
            address = address[8:]

        self._channel = grpc.aio.insecure_channel(address)
        self._build_client = builds_pb2_grpc.BuildServiceStub(self._channel)
        self._log_client = logs_pb2_grpc.LogServiceStub(self._channel)
        self._artifact_client = artifacts_pb2_grpc.ArtifactServiceStub(self._channel)

    @property
    def builds(self) -> builds_pb2_grpc.BuildServiceStub:
        """Get the BuildService client for direct access to build operations."""
        return self._build_client

    @property
    def logs(self) -> logs_pb2_grpc.LogServiceStub:
        """Get the LogService client for direct access to log streaming operations."""
        return self._log_client

    @property
    def artifacts(self) -> artifacts_pb2_grpc.ArtifactServiceStub:
        """Get the ArtifactService client for direct access to artifact operations."""
        return self._artifact_client

    async def start_build(
        self,
        config_path: str,
        target: str,
        customer: Optional[str] = None,
        force_clean: bool = False,
        force_image_rebuild: bool = False,
        environment_variables: Optional[dict[str, str]] = None,
    ) -> builds_pb2.BuildStatusResponse:
        """Start a new build with the specified configuration.

        Args:
            config_path: Path to the smidr.yaml configuration file
            target: Build target (e.g., "core-image-minimal")
            customer: Optional customer identifier
            force_clean: Force a clean rebuild
            force_image_rebuild: Force image rebuild only
            environment_variables: Optional dictionary of environment variables

        Returns:
            Build status response with build ID and initial state
        """
        request = builds_pb2.StartBuildRequest(
            config=config_path,
            target=target,
            customer=customer or "",
            force_clean=force_clean,
            force_image_rebuild=force_image_rebuild,
        )

        if environment_variables:
            request.environment_variables.update(environment_variables)

        return await self._build_client.StartBuild(request)

    async def get_build_status(self, build_id: str) -> builds_pb2.BuildStatusResponse:
        """Get the current status of a build.

        Args:
            build_id: The build ID to query

        Returns:
            Build status response with current state and metadata
        """
        request = builds_pb2.BuildStatusRequest(
            build_identifier=common_pb2.BuildIdentifier(build_id=build_id)
        )

        return await self._build_client.GetBuildStatus(request)

    async def list_builds(
        self,
        states: Optional[List[common_pb2.BuildState]] = None,
        page_size: int = 0,
        customer: Optional[str] = None,
        include_deleted: bool = False,
    ) -> builds_pb2.ListBuildsResponse:
        """List all builds matching the specified filters.

        Args:
            states: Optional list of build states to filter by
            page_size: Maximum number of builds to return (0 = all)
            customer: Optional customer identifier filter
            include_deleted: Include deleted builds in the response

        Returns:
            List of builds
        """
        request = builds_pb2.ListBuildsRequest(page_size=page_size, include_deleted=include_deleted)

        if states:
            request.state_filter.extend(states)

        if customer:
            request.customer = customer

        return await self._build_client.ListBuilds(request)

    async def cancel_build(self, build_id: str) -> builds_pb2.CancelBuildResponse:
        """Cancel a running build.

        Args:
            build_id: The build ID to cancel

        Returns:
            Cancellation response
        """
        request = builds_pb2.CancelBuildRequest(
            build_identifier=common_pb2.BuildIdentifier(build_id=build_id)
        )

        return await self._build_client.CancelBuild(request)

    async def list_artifacts(self, build_id: str) -> artifacts_pb2.ListArtifactsResponse:
        """List artifacts for a completed build.

        Args:
            build_id: The build ID to query

        Returns:
            List of artifacts with metadata
        """
        request = artifacts_pb2.ListArtifactsRequest(
            build_identifier=common_pb2.BuildIdentifier(build_id=build_id)
        )

        return await self._artifact_client.ListArtifacts(request)

    async def stream_logs(
        self, build_id: str, follow: bool = False
    ) -> AsyncIterator[logs_pb2.LogEntry]:
        """Stream logs from a build in real-time.

        Args:
            build_id: The build ID to stream logs from
            follow: Whether to continue streaming new logs as they arrive

        Yields:
            Log entries as they arrive
        """
        request = logs_pb2.StreamBuildLogsRequest(
            build_identifier=common_pb2.BuildIdentifier(build_id=build_id),
            follow=follow,
        )

        async for log_entry in self._log_client.StreamBuildLogs(request):
            yield log_entry

    async def close(self):
        """Close the gRPC channel and release resources."""
        await self._channel.close()

    async def __aenter__(self):
        """Async context manager entry."""
        return self

    async def __aexit__(self, exc_type, exc_val, exc_tb):
        """Async context manager exit."""
        await self.close()

    def __enter__(self):
        """Synchronous context manager entry (for compatibility)."""
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        """Synchronous context manager exit."""
        # Note: This is not ideal for sync context manager with async channel,
        # but provides compatibility. Users should prefer async context manager.
        import asyncio

        try:
            loop = asyncio.get_event_loop()
            if loop.is_running():
                # If loop is running, schedule close
                asyncio.create_task(self.close())
            else:
                loop.run_until_complete(self.close())
        except RuntimeError:
            # No event loop, create one
            asyncio.run(self.close())
