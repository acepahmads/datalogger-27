#!/usr/bin/env bash
# Datalogger Analysis Application - Raspberry Pi & Linux 1-Click Update Script
# Automatically detects installation type (/opt/datalogger systemd service or local folder)
# Pulls, installs binary & web assets, and restarts service.

set -e

echo "=========================================================="
echo "  DATALOGGER ANALYSIS APPLICATION - RASPBERRY PI UPDATER"
echo "=========================================================="

INSTALL_DIR="/opt/datalogger"
SERVICE_NAME="datalogger.service"

ARCH=$(uname -m)
echo "Detected CPU Architecture: $ARCH"

# 1. Check if git repo needs pulling or if already pulled
if [ -d ".git" ]; then
  echo "[1/4] Pulling latest commits from GitHub..."
  git pull origin main || true
fi

# 2. Stop systemd service if active
if systemctl is-active --quiet "$SERVICE_NAME" 2>/dev/null; then
  echo "[2/4] Stopping active $SERVICE_NAME..."
  sudo systemctl stop "$SERVICE_NAME" || true
fi

# 3. Deploy binary and web assets
echo "[3/4] Updating binary and frontend assets..."

TARGET_BIN="./datalogger"
if [ -d "$INSTALL_DIR" ]; then
  TARGET_BIN="$INSTALL_DIR/datalogger"
fi

UPDATED=0
if [ -f "./bin/datalogger-linux-arm64" ] && [ "$ARCH" = "aarch64" -o "$ARCH" = "arm64" ]; then
  echo "Deploying ARM64 binary to $TARGET_BIN..."
  sudo cp "./bin/datalogger-linux-arm64" "$TARGET_BIN"
  UPDATED=1
elif [ -f "./bin/datalogger-linux-armv7" ] && [ "$ARCH" = "armv7l" -o "$ARCH" = "armhf" ]; then
  echo "Deploying ARMv7 binary to $TARGET_BIN..."
  sudo cp "./bin/datalogger-linux-armv7" "$TARGET_BIN"
  UPDATED=1
elif [ -f "./bin/datalogger-linux-amd64" ] && [ "$ARCH" = "x86_64" ]; then
  echo "Deploying Linux x86_64 binary to $TARGET_BIN..."
  sudo cp "./bin/datalogger-linux-amd64" "$TARGET_BIN"
  UPDATED=1
fi

if [ "$UPDATED" -eq 0 ]; then
  if command -v go &> /dev/null; then
    echo "Compiling native binary on target machine using Go..."
    sudo CGO_ENABLED=0 go build -ldflags "-s -w" -o "$TARGET_BIN" ./cmd/main.go
  else
    echo "Error: Pre-built binary not matched and Go compiler not found."
    exit 1
  fi
fi

sudo chmod +x "$TARGET_BIN"

# Copy web/dist to install directory if /opt/datalogger exists
if [ -d "$INSTALL_DIR" ]; then
  sudo mkdir -p "$INSTALL_DIR/web"
  if [ -d "./web/dist" ]; then
    echo "Copying web/dist assets to $INSTALL_DIR/web/dist..."
    sudo cp -r ./web/dist "$INSTALL_DIR/web/"
  fi
fi

# 4. Restart service or start binary
echo "[4/4] Restarting Datalogger service..."
if [ -f "/etc/systemd/system/$SERVICE_NAME" ]; then
  sudo systemctl daemon-reload
  sudo systemctl restart "$SERVICE_NAME"
  sleep 2
  if sudo systemctl is-active --quiet "$SERVICE_NAME"; then
    echo "SUCCESS: $SERVICE_NAME is RUNNING!"
  else
    echo "WARNING: Service failed to start. Showing logs:"
    sudo journalctl -u "$SERVICE_NAME" -n 20 --no-pager
  fi
else
  echo "Service file not installed in systemd."
  echo "You can run directly with: sudo $TARGET_BIN"
fi

IP_ADDR=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "localhost")
echo "=========================================================="
echo "  UPDATE COMPLETED SUCCESSFULLY!"
echo "  URL: http://$IP_ADDR:8080"
echo "  Tip: Press Ctrl + Shift + R (Hard Refresh) in browser"
echo "=========================================================="
