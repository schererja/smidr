#!/bin/bash
# Smidr Agent Installation Script
# Usage: sudo ./install.sh [--control-plane-url URL] [--token TOKEN]

set -e

CONTROL_PLANE_URL="${CONTROL_PLANE_URL:-}"
TOKEN="${TOKEN:-}"

# Parse command line arguments
while [[ $# -gt 0 ]]; do
  case $1 in
    --control-plane-url)
      CONTROL_PLANE_URL="$2"
      shift 2
      ;;
    --token)
      TOKEN="$2"
      shift 2
      ;;
    *)
      echo "Unknown option: $1"
      echo "Usage: $0 [--control-plane-url URL] [--token TOKEN]"
      exit 1
      ;;
  esac
done

if [ -z "$CONTROL_PLANE_URL" ]; then
  echo "Error: --control-plane-url is required"
  echo "Usage: $0 --control-plane-url https://your-control-plane.example.com [--token TOKEN]"
  exit 1
fi

echo "Installing Smidr Agent..."

# Create smidr user and group
if ! id -u smidr > /dev/null 2>&1; then
  echo "Creating smidr user..."
  useradd --system --no-create-home --shell /usr/sbin/nologin smidr
fi

# Create directories
echo "Creating directories..."
mkdir -p /etc/smidr
mkdir -p /usr/local/bin

# Install binary
echo "Installing agent binary..."
if [ ! -f "smidr-agent" ]; then
  echo "Error: smidr-agent binary not found in current directory"
  exit 1
fi
cp smidr-agent /usr/local/bin/
chmod +x /usr/local/bin/smidr-agent

# Initialize configuration
echo "Initializing configuration..."
/usr/local/bin/smidr-agent init --config /etc/smidr/agent.yaml

# Update config with control plane URL
echo "Configuring control plane URL..."
sed -i "s|control_plane_url:.*|control_plane_url: \"$CONTROL_PLANE_URL\"|" /etc/smidr/agent.yaml

if [ -n "$TOKEN" ]; then
  echo "Configuring enrollment token..."
  sed -i "s|token:.*|token: \"$TOKEN\"|" /etc/smidr/agent.yaml
fi

# Set permissions
chown -R smidr:smidr /etc/smidr
chmod 700 /etc/smidr
chmod 600 /etc/smidr/agent.yaml

# Install systemd service
echo "Installing systemd service..."
if [ ! -f "systemd/smidr-agent.service" ]; then
  echo "Error: systemd/smidr-agent.service not found"
  exit 1
fi
cp systemd/smidr-agent.service /etc/systemd/system/
systemctl daemon-reload

# Enable and start service
echo "Enabling and starting service..."
systemctl enable smidr-agent
systemctl start smidr-agent

echo ""
echo "✓ Smidr Agent installed successfully!"
echo ""
echo "Service status:"
systemctl status smidr-agent --no-pager || true
echo ""
echo "To view logs: journalctl -u smidr-agent -f"
echo "To check status: systemctl status smidr-agent"
