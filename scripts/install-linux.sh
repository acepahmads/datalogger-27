#!/usr/bin/env bash
# Datalogger Analysis Application - Linux & Raspberry Pi systemd Installer
# Prepares directories, installs MariaDB if needed, installs systemd service, auto-starts on boot

set -e

echo "=========================================================="
echo "  DATALOGGER ANALYSIS APPLICATION - LINUX / RASPBERRY PI"
echo "  Edge Local-First Service Installer (MariaDB Edge)"
echo "=========================================================="

INSTALL_DIR="/opt/datalogger"
DATA_DIR="/var/lib/datalogger"
LOG_DIR="/var/log/datalogger"
SERVICE_FILE="/etc/systemd/system/datalogger.service"

if [ "$EUID" -ne 0 ]; then
  echo "Error: Please run as root (sudo bash install-linux.sh)"
  exit 1
fi

echo "[1/6] Creating application directories..."
mkdir -p "$INSTALL_DIR"
mkdir -p "$INSTALL_DIR/web"
mkdir -p "$DATA_DIR"
mkdir -p "$LOG_DIR"

echo "[2/6] Ensuring MariaDB server is installed and running..."
if ! command -v mariadb &> /dev/null && ! command -v mysql &> /dev/null; then
  echo "MariaDB server not found. Installing via apt..."
  apt-get update && apt-get install -y mariadb-server
fi

# Ensure service is enabled and started
systemctl enable mariadb 2>/dev/null || systemctl enable mysql 2>/dev/null || true
systemctl start mariadb 2>/dev/null || systemctl start mysql 2>/dev/null || true

# Initialize datalogger database
echo "Creating MariaDB 'datalogger' database if not exists..."
mariadb -e "CREATE DATABASE IF NOT EXISTS datalogger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || \
mysql -e "CREATE DATABASE IF NOT EXISTS datalogger CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null || true

echo "[3/6] Selecting and copying binary for this CPU architecture..."
ARCH=$(uname -m)
echo "Detected architecture: $ARCH"

if [ -f "./bin/datalogger-linux-arm64" ] && [ "$ARCH" = "aarch64" -o "$ARCH" = "arm64" ]; then
  echo "Using pre-built Raspberry Pi 64-bit ARM binary (bin/datalogger-linux-arm64)"
  cp "./bin/datalogger-linux-arm64" "$INSTALL_DIR/datalogger"
elif [ -f "./bin/datalogger-linux-armv7" ] && [ "$ARCH" = "armv7l" -o "$ARCH" = "armhf" ]; then
  echo "Using pre-built Raspberry Pi 32-bit ARM binary (bin/datalogger-linux-armv7)"
  cp "./bin/datalogger-linux-armv7" "$INSTALL_DIR/datalogger"
elif [ -f "./bin/datalogger-linux-amd64" ] && [ "$ARCH" = "x86_64" ]; then
  echo "Using pre-built Linux x86_64 binary (bin/datalogger-linux-amd64)"
  cp "./bin/datalogger-linux-amd64" "$INSTALL_DIR/datalogger"
elif [ -f "./bin/datalogger" ]; then
  cp "./bin/datalogger" "$INSTALL_DIR/datalogger"
elif [ -f "./datalogger" ]; then
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
fi

# Copy config if not existing in target
if [ ! -f "$INSTALL_DIR/config.json" ]; then
  if [ -f "./config.json" ]; then
    cp ./config.json "$INSTALL_DIR/config.json"
  fi
fi

echo "[4/6] Configuring systemd service unit..."
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
RestartSec=5s
Environment="DATALOGGER_PORT=8080"
Environment="DB_DRIVER=mariadb"
Environment="DB_HOST=127.0.0.1"
Environment="DB_PORT=3306"
Environment="DB_NAME=datalogger"
Environment="DB_USER=root"
Environment="DB_PASSWORD="
Environment="DATALOGGER_LOG_DIR=$LOG_DIR"
LimitNOFILE=65536

[Install]
WantedBy=multi-user.target
EOF

echo "[5/6] Reloading systemd daemon and enabling service..."
systemctl daemon-reload
systemctl enable datalogger.service
systemctl restart datalogger.service

echo "[6/6] Checking service status..."
sleep 2
systemctl is-active --quiet datalogger.service && echo "Service is RUNNING!" || echo "Check systemctl status datalogger"

IP_ADDR=$(hostname -I | awk '{print $1}')

echo "=========================================================="
echo "  INSTALLATION SUCCESSFUL!"
echo "  Target Host:    $ARCH ($IP_ADDR)"
echo "  Database:       MariaDB Edge (127.0.0.1:3306/datalogger)"
echo "  Web UI Access:  http://$IP_ADDR:8080"
echo "  Local Access:   http://localhost:8080"
echo "  Default Admin:  admin / admin123"
echo "=========================================================="
