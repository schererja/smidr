"""Configuration management for Smidr WebAPI."""

import os
from typing import Optional
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()


class Config:
    """Application configuration."""

    # Smidr daemon URL
    SMIDR_DAEMON_URL: str = os.getenv("SMIDR_DAEMON_URL", "http://localhost:50051")

    # Server configuration
    HOST: str = os.getenv("HOST", "0.0.0.0")
    PORT: int = int(os.getenv("PORT", "8000"))
    DEBUG: bool = os.getenv("DEBUG", "false").lower() == "true"

    # CORS
    CORS_ORIGINS: list[str] = os.getenv(
        "CORS_ORIGINS", "*"
    ).split(",")  # Comma-separated list of origins


config = Config()

