#!/bin/bash
set -e

# Generate development HTTPS certificate for Docker deployment
CERT_DIR="certs"
CERT_PASSWORD="${CERT_PASSWORD:-development}"

echo "Generating development HTTPS certificate..."

# Create certs directory if it doesn't exist
mkdir -p "$CERT_DIR"

# Generate certificate
dotnet dev-certs https -ep "$CERT_DIR/aspnetapp.pfx" -p "$CERT_PASSWORD" --trust

echo "Certificate generated at $CERT_DIR/aspnetapp.pfx"
echo "Password: $CERT_PASSWORD"
echo ""
echo "To use a different password, set the CERT_PASSWORD environment variable:"
echo "  export CERT_PASSWORD=your-password"
echo "  ./scripts/generate-dev-cert.sh"
