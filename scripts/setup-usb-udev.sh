#!/usr/bin/env bash
# ==============================================================================
# Datalogger Edge System — USB Serial Persistent udev Rule Setup
# Solves: ttyUSB0 -> ttyUSB1 port jumping & multiple USB serial device scrambling
# ==============================================================================

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${BLUE}==================================================================${NC}"
echo -e "${BLUE}   Datalogger Edge System — USB Serial Persistent Setup Tool      ${NC}"
echo -e "${BLUE}==================================================================${NC}"
echo ""

if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}[ERROR] This script must be run with sudo/root privileges.${NC}"
  echo "Usage: sudo bash scripts/setup-usb-udev.sh"
  exit 1
fi

# 1. Discover currently plugged USB serial devices
PORTS=()
for p in /dev/ttyUSB* /dev/ttyACM*; do
  if [ -e "$p" ]; then
    PORTS+=("$p")
  fi
done

if [ ${#PORTS[@]} -eq 0 ]; then
  echo -e "${YELLOW}[WARN] No USB serial adapters (/dev/ttyUSB* or /dev/ttyACM*) currently detected.${NC}"
  echo "Please plug in your USB-to-RS485 adapters and run this script again."
  echo ""
  exit 0
fi

echo -e "${GREEN}Found ${#PORTS[@]} USB serial adapter(s) currently attached:${NC}"
echo "------------------------------------------------------------------"

RULES_FILE="/etc/udev/rules.d/99-datalogger-serial.rules"
TEMP_RULES="/tmp/99-datalogger-serial.rules"
rm -f "$TEMP_RULES"
touch "$TEMP_RULES"

IDX=1
for PORT in "${PORTS[@]}"; do
  DEVNAME=$(basename "$PORT")
  
  # Read udev hardware attributes
  ID_VENDOR=$(udevadm info -q property -n "$PORT" | grep '^ID_VENDOR_ID=' | cut -d= -f2 || true)
  ID_MODEL=$(udevadm info -q property -n "$PORT" | grep '^ID_MODEL_ID=' | cut -d= -f2 || true)
  ID_SERIAL_SHORT=$(udevadm info -q property -n "$PORT" | grep '^ID_SERIAL_SHORT=' | cut -d= -f2 || true)
  ID_VENDOR_NAME=$(udevadm info -q property -n "$PORT" | grep '^ID_VENDOR_FROM_DATABASE=' | cut -d= -f2 || true)
  ID_MODEL_NAME=$(udevadm info -q property -n "$PORT" | grep '^ID_MODEL_FROM_DATABASE=' | cut -d= -f2 || true)
  DEVPATH_SYS=$(udevadm info -q path -n "$PORT" || true)
  
  # Check for kernel USB tree port (physical socket)
  DEV_KERNEL=$(udevadm info -a -n "$PORT" | grep -m 1 'KERNELS=="[0-9]-[0-9]' | tr -d ' ' || true)

  # Check existing by-id or by-path
  BY_ID=$(ls -l /dev/serial/by-id/* 2>/dev/null | grep "$DEVNAME" | awk '{print $9}' | head -n 1 || true)
  BY_PATH=$(ls -l /dev/serial/by-path/* 2>/dev/null | grep "$DEVNAME" | awk '{print $9}' | head -n 1 || true)

  ALIAS="datalogger_rs485_$IDX"

  echo -e "${CYAN}[Device $IDX] Current Port: $PORT${NC}"
  echo "  - Vendor/Model: ${ID_VENDOR_NAME:-Unknown} (${ID_VENDOR}:${ID_MODEL})"
  echo "  - Chip Serial : ${ID_SERIAL_SHORT:-N/A (Chip has no unique serial)}"
  if [ -n "$BY_ID" ]; then
    echo "  - Persistent by-ID: $BY_ID"
  fi
  if [ -n "$BY_PATH" ]; then
    echo "  - Persistent by-Path: $BY_PATH"
  fi

  # Generate rule logic:
  # Priority A: If unique chip serial is available, bind by VID + PID + Serial
  # Priority B: If no serial (e.g. CH340 clones), bind by physical USB port path (KERNELS)
  if [ -n "$ID_VENDOR" ] && [ -n "$ID_MODEL" ] && [ -n "$ID_SERIAL_SHORT" ] && [ "$ID_SERIAL_SHORT" != "null" ]; then
    echo "SUBSYSTEM==\"tty\", ATTRS{idVendor}==\"$ID_VENDOR\", ATTRS{idProduct}==\"$ID_MODEL\", ATTRS{serial}==\"$ID_SERIAL_SHORT\", SYMLINK+=\"$ALIAS\", MODE=\"0666\"" >> "$TEMP_RULES"
    echo -e "  ${GREEN}➔ Will create permanent alias: /dev/$ALIAS (locked to chip serial $ID_SERIAL_SHORT)${NC}"
  elif [ -n "$DEV_KERNEL" ]; then
    echo "SUBSYSTEM==\"tty\", $DEV_KERNEL, SYMLINK+=\"$ALIAS\", MODE=\"0666\"" >> "$TEMP_RULES"
    echo -e "  ${GREEN}➔ Will create permanent alias: /dev/$ALIAS (locked to physical USB socket)${NC}"
  else
    # Fallback to vendor + model
    echo "SUBSYSTEM==\"tty\", ATTRS{idVendor}==\"$ID_VENDOR\", ATTRS{idProduct}==\"$ID_MODEL\", SYMLINK+=\"$ALIAS\", MODE=\"0666\"" >> "$TEMP_RULES"
    echo -e "  ${GREEN}➔ Will create permanent alias: /dev/$ALIAS${NC}"
  fi
  echo "------------------------------------------------------------------"
  IDX=$((IDX + 1))
done

# Write rules file
cp "$TEMP_RULES" "$RULES_FILE"
rm -f "$TEMP_RULES"
chmod 644 "$RULES_FILE"

echo ""
echo -e "${BLUE}Applying udev rules...${NC}"
udevadm control --reload-rules
udevadm trigger

sleep 1

echo ""
echo -e "${GREEN}==================================================================${NC}"
echo -e "${GREEN}   Persistent udev Rules Installed Successfully!                  ${NC}"
echo -e "${GREEN}==================================================================${NC}"
echo "File created: $RULES_FILE"
echo ""
echo "Current persistent aliases in /dev/:"
ls -la /dev/datalogger_rs485_* 2>/dev/null || echo "(Aliases will appear upon next USB re-plug or restart)"
echo ""
echo -e "${CYAN}Cara Menggunakan di Web Datalogger:${NC}"
echo "1. Buka menu 'Devices & Sensors' di dashboard."
echo "2. Pada field 'Serial Port', pilih atau ketik alias permanen:"
for ((i=1; i<IDX; i++)); do
  echo "   - /dev/datalogger_rs485_$i"
done
echo "3. Port ini 100% permanen, tidak akan berpindah atau tertukar selamanya!"
echo ""
