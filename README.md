# Datalogger Analysis Application
> **Local-First Industrial Telemetry & Development Progress Dashboard**

[![Phase 1 Foundation](https://img.shields.io/badge/Phase%201-100%25%20Completed-success.svg)](docs/phase1_foundation_report.md)
[![Architecture](https://img.shields.io/badge/Architecture-Go%20%2B%20MariaDB%20Edge-blue.svg)](docs/architecture.md)
[![UI](https://img.shields.io/badge/Frontend-Vue%202%20%2B%20Tailwind%20SaaS-indigo.svg)](web/)
[![Zero CGO](https://img.shields.io/badge/CGO-Disabled%20(Cross--Platform)-teal.svg)](#cross-platform-build)

A production-ready, **local-first** industrial telemetry and analysis platform designed to run on low-spec hardware (Raspberry Pi 3/4/5, Linux ARM SBCs, Industrial PCs, and Windows). The system operates completely autonomously without requiring cloud or internet connectivity.

---

## Key Features

- **Local-First Architecture:** Complete offline autonomy for acquisition, processing, alarms, and visualization.
- **Embedded Web UI:** Modern industrial SaaS dashboard built with Vue 2 and Tailwind CSS.
- **Pure-Go MariaDB Engine:** High-performance relational database with zero CGO dependency, connection pooling, and automated schema migrations.
- **Dual-Purpose Dashboard:**
  - **Development Progress:** Project single source of truth tracking all 10 engineering phases and 110+ tasks.
  - **System Monitoring:** Real-time hardware health (CPU, RAM, Disk, MariaDB status, WebSocket) and telemetry metrics.
- **Real-Time Push Updates:** Concurrency-safe WebSocket engine delivering live telemetry and task updates without aggressive polling.
- **Multi-Platform Installer:** Automated installation scripts for Windows background launch and Linux/Raspberry Pi `systemd` service.

---

## 10 Development Phases

```text
Phase 1  ━━━━━━━ 100% DONE (Foundation & Architecture)
Phase 2  ━━━━━━━ 0%   PENDING (Device & Communication)
Phase 3  ━━━━━━━ 0%   PENDING (Data Engine)
Phase 4  ━━━━━━━ 0%   PENDING (Reliability)
Phase 5  ━━━━━━━ 0%   PENDING (Storage)
Phase 6  ━━━━━━━ 0%   PENDING (Alarm & Notification)
Phase 7  ━━━━━━━ 35%  WORKING (Web Application)
Phase 8  ━━━━━━━ 10%  PENDING (Output & Integration)
Phase 9  ━━━━━━━ 85%  WORKING (Cross Platform & Installer)
Phase 10 ━━━━━━━ 25%  WORKING (Optimization & Production)
```

---

## Quick Start

### 1. Run with Go (Windows / Linux / macOS)

```bash
# Clone the repository
git clone https://github.com/acepahmads/datalogger-27.git
cd datalogger-27

# Run the backend (serves prebuilt Web UI on port 8080)
go run cmd/main.go
```

Access the dashboard at: **`http://localhost:8080`**  
Default credentials: **`admin`** / **`admin123`**

---

### 2. Windows Local Installation (Automated)

Run the PowerShell installer to set up directories, build binaries, create a desktop shortcut, and launch the browser:

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\install-windows.ps1
```

---

### 3. Linux / Raspberry Pi `systemd` Service Installation

Deploy as a self-restarting systemd background service:

```bash
sudo bash ./scripts/install-linux.sh
```

---

## Cross-Platform Build

Because pure-Go MariaDB connector is used with `CGO_ENABLED=0`, cross-compilation requires no external C compilers:

```bash
# Windows 64-bit
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o bin/datalogger-win-x64.exe cmd/main.go

# Linux 64-bit
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w" -o bin/datalogger-linux-x64 cmd/main.go

# Raspberry Pi 4/5 (Linux ARM64)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags "-s -w" -o bin/datalogger-linux-arm64 cmd/main.go

# Raspberry Pi Zero / 2 / 3 (Linux ARM32)
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -ldflags "-s -w" -o bin/datalogger-linux-armv7 cmd/main.go
```

Or run the automated multi-architecture build script:

```bash
.\scripts\build-cross-platform.bat
```

---

## Running Automated Tests

```bash
go test -v ./...
```

---

## Documentation

- [System Architecture Specification](docs/architecture.md)
- [Phase 1 Milestone Completion Report](docs/phase1_foundation_report.md)
- [Industrial USB Serial Resilience & Multiple Devices Guide (Solusi 3 Pilar)](docs/usb_serial_resilience_and_multiple_devices_guide.md)
- [Phase 2.3 — Communication Hardening & Field Validation](docs/phase2_3_communication_hardening.md)
- [Phase 3.1 — Telemetry Data Pipeline & High-Throughput Ingestion](docs/phase3_1_data_pipeline_ingestion.md)
