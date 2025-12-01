"""Main FastAPI application for Smidr WebAPI."""

import sys
from pathlib import Path

# Add the src directory to the path so we can import smidr_webapi
src_path = Path(__file__).parent.parent.parent
if str(src_path) not in sys.path:
    sys.path.insert(0, str(src_path))

# Add the SDK to the path so we can import it
# In production, the SDK should be installed as a package
# We need to add the SDK directory so the 'generated' package can be found
sdk_path = Path(__file__).parent.parent.parent.parent / "sdks" / "python"
if sdk_path.exists():
    # Add the SDK directory itself (not the parent) so 'generated' is importable
    if str(sdk_path) not in sys.path:
        sys.path.insert(0, str(sdk_path))

from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from smidr_webapi.config import config
from smidr_webapi.routes import builds, artifacts, logs

# Create FastAPI app
app = FastAPI(
    title="Smidr API",
    version="v1",
    description="REST API server for Smidr gRPC daemon",
    docs_url="/swagger" if config.DEBUG else None,
    redoc_url="/redoc" if config.DEBUG else None,
)

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=config.CORS_ORIGINS if "*" not in config.CORS_ORIGINS else ["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include routers
app.include_router(builds.router)
app.include_router(artifacts.router)
app.include_router(logs.router)


@app.get("/")
async def root():
    """Root endpoint."""
    return {
        "message": "Smidr API",
        "version": "v1",
        "docs": "/swagger" if config.DEBUG else "disabled",
    }


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "healthy"}


def main():
    """Main entry point for running the server."""
    import uvicorn

    # Use the app directly instead of string reference to avoid path issues
    uvicorn.run(
        app,
        host=config.HOST,
        port=config.PORT,
        reload=config.DEBUG,
    )


if __name__ == "__main__":
    main()
