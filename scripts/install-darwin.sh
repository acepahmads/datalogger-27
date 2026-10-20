#!/usr/bin/env bash
# ==============================================================================
# Datalogger Analysis Application - macOS Service & Binary Installer
# Supports macOS Intel (x86_64) and Apple Silicon (arm64 / M1-M4)
# Configures launchd daemon for auto-start on boot
# ==============================================================================

set -e

INSTALL_DIR="/usr/local/opt/datalogger"
DATA_DIR="/usr/local/var/datalogger"
LOG_DIR="/var/log/datalogger"
PLIST_FILE="/Library/LaunchDaemons/com.datalogger.analysis.plist"

echo "=========================================================="
echo "  DATALOGGER ANALYSIS APPLICATION - MACOS INSTALLER"
echo "  Target: macOS (Darwin) with launchd auto-start"
echo "=========================================================="

# Check OS
OS_NAME=$(uname -s)
if [ "$OS_NAME" != "Darwin" ]; then
  echo "Error: This installer is intended for macOS (Darwin), detected: $OS_NAME"
  exit 1
fi

if [ "$EUID" -ne 0 ]; then
  echo "Error: Please run as root (sudo bash install-darwin.sh)"
  exit 1
fi

# Handle uninstall flag
if [ "$1" = "--uninstall" ] || [ "$1" = "-u" ] || [ "$1" = "--remove" ]; then
  echo "Uninstalling Datalogger Application from macOS..."
  if [ -f "$PLIST_FILE" ]; then
    echo "Unloading and removing launchd service..."
    launchctl unload "$PLIST_FILE" 2>/dev/null || true
    rm -f "$PLIST_FILE"
  fi

  if [ -d "$INSTALL_DIR" ]; then
    echo "Removing application binaries and frontend assets from $INSTALL_DIR..."
    rm -rf "$INSTALL_DIR"
  fi

  echo ""
  read -p "Do you want to delete persistent telemetry and log files ($DATA_DIR and $LOG_DIR)? [y/N]: " CONFIRM_PURGE
  if [ "$CONFIRM_PURGE" = "y" ] || [ "$CONFIRM_PURGE" = "Y" ]; then
    rm -rf "$DATA_DIR" "$LOG_DIR"
    echo "Persistent data directories removed."
  else
    echo "Persistent data preserved in $DATA_DIR and $LOG_DIR."
  fi
  echo "Datalogger has been successfully uninstalled from this Mac."
  exit 0
fi

echo "[1/5] Creating application directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR/web"
mkdir -p "$DATA_DIR"
mkdir -p "$LOG_DIR"

ARCH=$(uname -m)
echo "[2/5] Detected Darwin architecture: $ARCH"

TARGET_BIN="$INSTALL_DIR/datalogger"
COPIED=0

if [ -f "./bin/datalogger-darwin-arm64" ] && [ "$ARCH" = "arm64" ]; then
  echo "Using pre-built Apple Silicon binary (bin/datalogger-darwin-arm64)..."
  cp "./bin/datalogger-darwin-arm64" "$TARGET_BIN"
  COPIED=1
elif [ -f "./bin/datalogger-darwin-amd64" ] && [ "$ARCH" = "x86_64" ]; then
  echo "Using pre-built macOS Intel binary (bin/datalogger-darwin-amd64)..."
  cp "./bin/datalogger-darwin-amd64" "$TARGET_BIN"
  COPIED=1
elif [ -f "./datalogger" ]; then
  echo "Using standalone datalogger binary in current directory..."
  cp "./datalogger" "$TARGET_BIN"
  COPIED=1
fi

if [ "$COPIED" -eq 0 ]; then
  if command -v go &> /dev/null; then
    echo "Compiling native Darwin binary on target Mac..."
    CGO_ENABLED=0 go build -ldflags "-s -w" -o "$TARGET_BIN" ./cmd/main.go
  else
    echo "Error: Pre-built binary not found for $ARCH and Go compiler not available."
    exit 1
  fi
fi

chmod +x "$TARGET_BIN"

# Copy web assets
echo "[3/5] Deploying Web UI production assets..."
if [ -d "./web/dist" ]; then
  cp -r ./web/dist "$INSTALL_DIR/web/"
elif [ -d "./dist" ]; then
  cp -r ./dist "$INSTALL_DIR/web/"
fi

# Copy default config if not existing
if [ ! -f "$INSTALL_DIR/config.json" ]; then
  if [ -f "./config.json" ]; then
    cp ./config.json "$INSTALL_DIR/config.json"
  fi
fi

# Install launchd service
echo "[4/5] Installing launchd service daemon..."
if [ -f "./scripts/com.datalogger.analysis.plist" ]; then
  cp "./scripts/com.datalogger.analysis.plist" "$PLIST_FILE"
elif [ -f "./com.datalogger.analysis.plist" ]; then
  cp "./com.datalogger.analysis.plist" "$PLIST_FILE"
else
  cat <<EOF > "$PLIST_FILE"
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.datalogger.analysis</string>
    <key>ProgramArguments</key>
    <array>
        <string>$TARGET_BIN</string>
    </array>
    <key>WorkingDirectory</key>
    <string>$INSTALL_DIR</string>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>$LOG_DIR/stdout.log</string>
    <key>StandardErrorPath</key>
    <string>$LOG_DIR/stderr.log</string>
</dict>
</plist>
EOF
fi

chmod 644 "$PLIST_FILE"
chown root:wheel "$PLIST_FILE"

echo "[5/5] Registering and starting launchd service..."
launchctl unload "$PLIST_FILE" 2>/dev/null || true
launchctl load -w "$PLIST_FILE"

echo ""
echo "=========================================================="
echo "  MACOS INSTALLATION COMPLETE!"
echo "  Application: $TARGET_BIN"
echo "  Launchd Svc: $PLIST_FILE"
echo "  Web UI URL:  http://localhost:8080"
echo "  Logs:        $LOG_DIR/stdout.log"
echo "  Note: Gatekeeper may require approval in System Settings"
echo "        > Privacy & Security if binary is not notarized."
echo "=========================================================="
