#!/bin/bash
# Startup script for Smidr Python WebAPI

# Get the directory where this script is located
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Check if virtual environment exists
if [ ! -d "venv" ]; then
    echo "Virtual environment not found. Creating one..."
    python3 -m venv venv
    source venv/bin/activate
    echo "Installing WebAPI dependencies..."
    pip install -r requirements.txt
    echo "Installing SDK in development mode..."
    cd ../../sdks/python
    pip install -e .
    cd "$SCRIPT_DIR"
    echo "Setup complete!"
else
    # Activate virtual environment
    source venv/bin/activate

    # Check if SDK is installed, if not install it
    if ! python3 -c "import smidr_sdk" 2>/dev/null; then
        echo "SDK not found. Installing SDK in development mode..."
        cd ../../sdks/python
        pip install -e .
        cd "$SCRIPT_DIR"
    fi
fi

# Set PYTHONPATH to include src directory (fallback if not installed)
export PYTHONPATH="${SCRIPT_DIR}/src:${PYTHONPATH}"

# Add SDK to path (fallback if not installed)
export PYTHONPATH="${SCRIPT_DIR}/../../sdks/python:${PYTHONPATH}"

# Run uvicorn
uvicorn smidr_webapi.main:app --host 0.0.0.0 --port 8000 --reload

