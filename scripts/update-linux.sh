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

# 1. Check if git repo needs pulling or if already pulled (auto-detect active branch safely even under sudo)
if [ -d ".git" ]; then
  REPO_DIR=$(pwd -P 2>/dev/null || pwd)
  run_git() {
    if [ -n "$SUDO_USER" ] && [ "$EUID" -eq 0 ]; then
      sudo -u "$SUDO_USER" git -c safe.directory="$REPO_DIR" "$@"
    else
      git -c safe.directory="$REPO_DIR" "$@"
    fi
  }

  if [ -n "$1" ]; then
    CURRENT_BRANCH="$1"
  else
    CURRENT_BRANCH=$(run_git symbolic-ref --short HEAD 2>/dev/null || run_git rev-parse --abbrev-ref HEAD 2>/dev/null || sed -n 's|^ref: refs/heads/||p' .git/HEAD 2>/dev/null)
  fi

  if [ -n "$CURRENT_BRANCH" ] && [ "$CURRENT_BRANCH" != "HEAD" ]; then
    echo "[1/4] Pulling latest commits from GitHub (active branch: $CURRENT_BRANCH)..."
    run_git pull origin "$CURRENT_BRANCH" || true
  else
    echo "[1/4] Detached HEAD detected; skipping automatic git pull."
  fi
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
CANDIDATE_BIN=""
if [ -f "./bin/datalogger-linux-arm64" ] && [ "$ARCH" = "aarch64" -o "$ARCH" = "arm64" ]; then
  CANDIDATE_BIN="./bin/datalogger-linux-arm64"
elif [ -f "./bin/datalogger-linux-armv7" ] && [ "$ARCH" = "armv7l" -o "$ARCH" = "armhf" ]; then
  CANDIDATE_BIN="./bin/datalogger-linux-armv7"
elif [ -f "./bin/datalogger-linux-amd64" ] && [ "$ARCH" = "x86_64" ]; then
  CANDIDATE_BIN="./bin/datalogger-linux-amd64"
fi

# If Go compiler is installed and .go sources have newer commits than prebuilt binary, compile fresh binary
PREFER_NATIVE_BUILD=0
if [ -d ".git" ] && [ -n "$CANDIDATE_BIN" ] && command -v go &> /dev/null; then
  LAST_GO_COMMIT=$(run_git log -n 1 --format=%H -- cmd internal pkg 2>/dev/null || echo "")
  LAST_BIN_COMMIT=$(run_git log -n 1 --format=%H -- "$CANDIDATE_BIN" 2>/dev/null || echo "")
  if [ -n "$LAST_GO_COMMIT" ] && [ "$LAST_GO_COMMIT" != "$LAST_BIN_COMMIT" ]; then
    ANCESTOR_CHECK=0
    run_git merge-base --is-ancestor "$LAST_GO_COMMIT" "$LAST_BIN_COMMIT" 2>/dev/null && ANCESTOR_CHECK=1 || true
    if [ "$ANCESTOR_CHECK" -eq 0 ]; then
      echo "Detected Go source commits newer than $CANDIDATE_BIN; compiling fresh native binary..."
      PREFER_NATIVE_BUILD=1
    fi
  fi
fi

if [ "$PREFER_NATIVE_BUILD" -eq 0 ] && [ -n "$CANDIDATE_BIN" ]; then
  echo "Deploying pre-built binary ($CANDIDATE_BIN) to $TARGET_BIN..."
  sudo cp "$CANDIDATE_BIN" "$TARGET_BIN"
  UPDATED=1
fi

if [ "$UPDATED" -eq 0 ]; then
  if command -v go &> /dev/null; then
    echo "Compiling native binary on target machine using Go..."
    sudo CGO_ENABLED=0 go build -ldflags "-s -w" -o "$TARGET_BIN" ./cmd/main.go
  elif [ -n "$CANDIDATE_BIN" ]; then
    echo "Go compiler not found; falling back to pre-built binary ($CANDIDATE_BIN)..."
    sudo cp "$CANDIDATE_BIN" "$TARGET_BIN"
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
    echo "Deploying fresh web/dist assets to $INSTALL_DIR/web/dist..."
    CURRENT_DIR=$(pwd -P 2>/dev/null || pwd)
    INSTALL_REAL=$(cd "$INSTALL_DIR" 2>/dev/null && pwd -P || echo "$INSTALL_DIR")
    if [ "$CURRENT_DIR" != "$INSTALL_REAL" ]; then
      sudo rm -rf "$INSTALL_DIR/web/dist"
      sudo cp -r ./web/dist "$INSTALL_DIR/web/dist"
    fi
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
