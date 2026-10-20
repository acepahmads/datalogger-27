#!/usr/bin/env bash
# ==============================================================================
# Datalogger Analysis Application - Linux & Raspberry Pi systemd Installer
# Prepares directories, verifies MariaDB, installs systemd service, auto-starts on boot
# Supports safe upgrade, pre-flight checks, and non-destructive uninstall.
# ==============================================================================

set -e

INSTALL_DIR="/opt/datalogger"
DATA_DIR="/var/lib/datalogger"
LOG_DIR="/var/log/datalogger"
SERVICE_FILE="/etc/systemd/system/datalogger.service"

echo "=========================================================="
echo "  DATALOGGER ANALYSIS APPLICATION - LINUX / RASPBERRY PI"
echo "  Edge Local-First Service Installer (MariaDB Edge)"
echo "=========================================================="

if [ "$EUID" -ne 0 ]; then
  echo "Error: Please run as root (sudo bash install-linux.sh)"
  exit 1
fi

# ------------------------------------------------------------------------------
# Handle Uninstallation
# ------------------------------------------------------------------------------
if [ "$1" = "--uninstall" ] || [ "$1" = "-u" ] || [ "$1" = "--remove" ]; then
  echo "Uninstalling Datalogger Application from this Linux system..."
  if command -v systemctl &> /dev/null; then
    if systemctl is-active --quiet datalogger.service 2>/dev/null; then
      echo "Stopping datalogger.service..."
      systemctl stop datalogger.service || true
    fi
    if [ -f "$SERVICE_FILE" ]; then
      echo "Disabling and removing systemd service..."
      systemctl disable datalogger.service 2>/dev/null || true
      rm -f "$SERVICE_FILE"
      systemctl daemon-reload
    fi
  fi

  if [ -d "$INSTALL_DIR" ]; then
    echo "Removing application binaries and Web assets from $INSTALL_DIR..."
    rm -rf "$INSTALL_DIR"
  fi

  echo ""
  echo "Data Preservation Confirmation:"
  echo "Persistent telemetry data is stored in $DATA_DIR and MariaDB database 'datalogger'."
  read -p "Do you want to PERMANENTLY DELETE persistent telemetry data and logs? [y/N]: " CONFIRM_PURGE
  if [ "$CONFIRM_PURGE" = "y" ] || [ "$CONFIRM_PURGE" = "Y" ]; then
    rm -rf "$DATA_DIR" "$LOG_DIR"
    echo "Persistent telemetry directories removed."
    echo "Note: To drop MariaDB database, run: mariadb -e 'DROP DATABASE datalogger;'"
  else
    echo "Persistent telemetry data and logs have been PRESERVED in $DATA_DIR and $LOG_DIR."
  fi
  echo "Uninstallation completed successfully."
  exit 0
fi

# ------------------------------------------------------------------------------
# [1/6] Pre-flight verification & directory creation
# ------------------------------------------------------------------------------
echo "[1/6] Running pre-flight checks and creating application directories..."

# Check available disk space (at least 150MB recommended)
AVAIL_KB=$(df -P "$INSTALL_DIR" 2>/dev/null | tail -1 | awk '{print $4}' || df -P /opt 2>/dev/null | tail -1 | awk '{print $4}' || echo "999999")
if [ "$AVAIL_KB" -lt 150000 ]; then
  echo "Warning: Low disk space detected ($((AVAIL_KB / 1024)) MB available). Minimum 150 MB recommended."
fi

mkdir -p "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR/web"
mkdir -p "$DATA_DIR"
mkdir -p "$LOG_DIR"

# ------------------------------------------------------------------------------
# [2/6] Ensure MariaDB server is installed and running
# ------------------------------------------------------------------------------
echo "[2/6] Ensuring MariaDB server is installed and running..."
if ! command -v mariadb &> /dev/null && ! command -v mysql &> /dev/null; then
  echo "MariaDB server not found. Installing via apt..."
  apt-get update && apt-get install -y mariadb-server || true
fi

# Ensure service is enabled and started
if command -v systemctl &> /dev/null; then
  systemctl enable mariadb 2>/dev/null || systemctl enable mysql 2>/dev/null || true
  systemctl start mariadb 2>/dev/null || systemctl start mysql 2>/dev/null || true
fi

# Initialize datalogger database and dedicated edge credentials
echo "Configuring MariaDB database & permissions for Datalogger..."
mariadb -e "CREATE DATABASE IF NOT EXISTS datalogger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || mysql -e "CREATE DATABASE IF NOT EXISTS datalogger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || true
mariadb -e "CREATE USER IF NOT EXISTS 'datalogger'@'localhost' IDENTIFIED BY 'datalogger';" 2>/dev/null || true
mariadb -e "CREATE USER IF NOT EXISTS 'datalogger'@'127.0.0.1' IDENTIFIED BY 'datalogger';" 2>/dev/null || true
mariadb -e "GRANT ALL PRIVILEGES ON datalogger.* TO 'datalogger'@'localhost' IDENTIFIED BY 'datalogger';" 2>/dev/null || true
mariadb -e "GRANT ALL PRIVILEGES ON datalogger.* TO 'datalogger'@'127.0.0.1' IDENTIFIED BY 'datalogger';" 2>/dev/null || true
mariadb -e "GRANT ALL PRIVILEGES ON datalogger.* TO 'root'@'localhost';" 2>/dev/null || true
mariadb -e "GRANT ALL PRIVILEGES ON datalogger.* TO 'root'@'127.0.0.1';" 2>/dev/null || true
mariadb -e "FLUSH PRIVILEGES;" 2>/dev/null || mysql -e "FLUSH PRIVILEGES;" 2>/dev/null || true

# ------------------------------------------------------------------------------
# [3/6] Binary deployment with rollback protection
# ------------------------------------------------------------------------------
echo "[3/6] Selecting and copying binary for this CPU architecture..."
ARCH=$(uname -m)
echo "Detected architecture: $ARCH"

# Create backup of existing binary for rollback safety
if [ -f "$INSTALL_DIR/datalogger" ]; then
  cp "$INSTALL_DIR/datalogger" "$INSTALL_DIR/datalogger.bak"
  echo "Previous binary backed up to $INSTALL_DIR/datalogger.bak"
fi

if [ -f "./bin/datalogger-linux-arm64" ] && [ "$ARCH" = "aarch64" -o "$ARCH" = "arm64" ]; then
  echo "Using pre-built Raspberry Pi 64-bit ARM binary (bin/datalogger-linux-arm64)"
  cp "./bin/datalogger-linux-arm64" "$INSTALL_DIR/datalogger"
elif [ -f "./bin/datalogger-linux-armv7" ] && [ "$ARCH" = "armv7l" -o "$ARCH" = "armhf" ]; then
  echo "Using pre-built Raspberry Pi 32-bit ARMv7 binary (bin/datalogger-linux-armv7)"
  cp "./bin/datalogger-linux-armv7" "$INSTALL_DIR/datalogger"
elif [ -f "./bin/datalogger-linux-armv6" ] && [ "$ARCH" = "armv6l" ]; then
  echo "Using pre-built Raspberry Pi 32-bit ARMv6 binary (bin/datalogger-linux-armv6)"
  cp "./bin/datalogger-linux-armv6" "$INSTALL_DIR/datalogger"
elif [ -f "./bin/datalogger-linux-amd64" ] && [ "$ARCH" = "x86_64" ]; then
  echo "Using pre-built Linux x86_64 binary (bin/datalogger-linux-amd64)"
  cp "./bin/datalogger-linux-amd64" "$INSTALL_DIR/datalogger"
elif [ -f "./bin/datalogger" ]; then
  cp "./bin/datalogger" "$INSTALL_DIR/datalogger"
elif [ -f "./datalogger" ]; then
  echo "Using standalone datalogger binary in current directory..."
  cp "./datalogger" "$INSTALL_DIR/datalogger"
else
  echo "No pre-built binary found for $ARCH."
  if ! command -v go &> /dev/null; then
    echo "Go compiler not found on this machine. Installing golang-go via apt..."
    apt-get update && apt-get install -y golang-go
  fi
  echo "Compiling native binary on target machine using Go..."
  CGO_ENABLED=0 go build -ldflags "-s -w" -o "$INSTALL_DIR/datalogger" ./cmd/main.go
fi

chmod +x "$INSTALL_DIR/datalogger"

# Copy web assets if present
if [ -d "./web/dist" ]; then
  cp -r ./web/dist "$INSTALL_DIR/web/"
elif [ -d "./dist" ]; then
  cp -r ./dist "$INSTALL_DIR/web/"
fi

# Copy config if not existing in target (never overwrite during upgrades!)
if [ ! -f "$INSTALL_DIR/config.json" ]; then
  if [ -f "./config.json" ]; then
    cp ./config.json "$INSTALL_DIR/config.json"
    echo "Default configuration copied to $INSTALL_DIR/config.json"
  fi
else
  echo "Existing configuration preserved at $INSTALL_DIR/config.json"
fi

# Ensure config.json uses datalogger credentials
if [ -f "$INSTALL_DIR/config.json" ]; then
  sed -i 's/"db_user": "root"/"db_user": "datalogger"/g' "$INSTALL_DIR/config.json" 2>/dev/null || true
  sed -i 's/"db_password": ""/"db_password": "datalogger"/g' "$INSTALL_DIR/config.json" 2>/dev/null || true
fi

# ------------------------------------------------------------------------------
# [4/6] Configuring systemd service unit
# ------------------------------------------------------------------------------
echo "[4/6] Configuring systemd service unit..."
if command -v systemctl &> /dev/null; then
  cat <<EOF > "$SERVICE_FILE"
[Unit]
Description=Datalogger Analysis Application Service
After=network.target mariadb.service mysql.service
Wants=mariadb.service

[Service]
Type=simple
User=root
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/datalogger
Restart=always
RestartSec=3s
Environment="DATALOGGER_PORT=8080"
Environment="DB_DRIVER=mariadb"
Environment="DB_HOST=127.0.0.1"
Environment="DB_PORT=3306"
Environment="DB_NAME=datalogger"
Environment="DB_USER=datalogger"
Environment="DB_PASSWORD=datalogger"
Environment="DATALOGGER_LOG_DIR=$LOG_DIR"
Environment="DATALOGGER_DATA_DIR=$DATA_DIR"
Environment="DATALOGGER_WEB_DIR=$INSTALL_DIR/web/dist"
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF
  chmod 644 "$SERVICE_FILE"

  echo "[5/6] Reloading systemd daemon and enabling service..."
  systemctl daemon-reload
  systemctl enable datalogger.service
  systemctl restart datalogger.service

  echo "[6/6] Checking service status..."
  sleep 3
  if systemctl is-active --quiet datalogger.service; then
    echo "Service is RUNNING!"
  else
    echo "Notice: Service starting up... checking logs:"
    journalctl -u datalogger.service -n 15 --no-pager || true
  fi
else
  echo "[5/6] systemctl not found on this system (container or non-systemd environment)."
  echo "You can launch the application directly with:"
  echo "  cd $INSTALL_DIR && ./datalogger"
  echo "[6/6] Non-systemd background launcher helper created at $INSTALL_DIR/start.sh"
  cat <<EOF > "$INSTALL_DIR/start.sh"
#!/bin/sh
cd "$INSTALL_DIR"
export DATALOGGER_PORT=8080
export DB_DRIVER=mariadb
exec "$INSTALL_DIR/datalogger"
EOF
  chmod +x "$INSTALL_DIR/start.sh"
fi

IP_ADDR=$(hostname -I 2>/dev/null | awk '{print $1}' || echo "127.0.0.1")

echo "=========================================================="
echo "  INSTALLATION / UPGRADE SUCCESSFUL!"
echo "  Target Host:    $ARCH ($IP_ADDR)"
echo "  Installation:   $INSTALL_DIR"
echo "  Database:       MariaDB Edge (127.0.0.1:3306/datalogger)"
echo "  Web UI Access:  http://$IP_ADDR:8080"
echo "  Local Access:   http://localhost:8080"
echo "  Default Admin:  admin / admin123"
echo "=========================================================="
