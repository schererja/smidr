"""Pydantic models for request/response validation."""

from typing import Optional
from pydantic import BaseModel


class StartBuildRequest(BaseModel):
    """Request model for starting a build."""

    config_path: str
    target: str
    customer: Optional[str] = None
    force_clean: Optional[bool] = False
    force_image_rebuild: Optional[bool] = False
    environment_variables: Optional[dict[str, str]] = None


class BuildStatusResponse(BaseModel):
    """Response model for build status."""

    build_id: str
    state: str
    start_time: Optional[str] = None
    end_time: Optional[str] = None
    error_message: Optional[str] = None
    exit_code: Optional[int] = None


class BuildListItem(BaseModel):
    """Model for a build in a list."""

    build_id: str
    state: str
    start_time: str
    target: str


class ListBuildsResponse(BaseModel):
    """Response model for listing builds."""

    builds: list[BuildListItem]


class CancelBuildResponse(BaseModel):
    """Response model for canceling a build."""

    success: bool
    message: str


class ArtifactItem(BaseModel):
    """Model for an artifact."""

    name: str
    size_bytes: int
    download_url: str
    checksum: str


class ListArtifactsResponse(BaseModel):
    """Response model for listing artifacts."""

    build_id: str
    artifacts: list[ArtifactItem]


class ErrorResponse(BaseModel):
    """Error response model."""

    error: str

