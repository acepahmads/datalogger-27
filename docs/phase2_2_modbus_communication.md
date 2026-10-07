# Phase 2.2 — Modbus RTU / TCP Communication Engine

**Project:** Datalogger Analysis Application  
**Version:** v1.2.0-phase2.2  
**Phase:** Phase 2 — Device & Communication  
**Task:** Phase 2.2 — Modbus RTU / TCP Communication Engine  
**Status:** **COMPLETED (100% — 27/27 Acceptance Criteria Verified)**  
**Author:** Antigravity AI Pair Programmer & Lead Engineering Team  
**Date:** 2026-10-07  

---

## 1. Executive Summary

Phase 2.2 implements the first real industrial communication engine for the Datalogger Analysis Application. Built from the ground up to operate reliably on edge devices (such as Raspberry Pi 4/5, industrial IPCs, and embedded ARM64/AMD64 platforms), the engine introduces:
- **Pure-Go Modbus RTU Master**: Complete RS485/RS232 serial framing, CRC-16 (polynomial 0xA001) calculation, silent timeout detection, inter-frame delay handling, and exception parsing without requiring external C libraries (zero CGO).
- **High-Performance Modbus TCP Client**: Full MBAP (Modbus Application Protocol) transport client over Ethernet and Wi-Fi supporting persistent socket reuse, thread-safe transaction identification, slave/unit addressing, and TCP keepalive.
- **Unified Protocol Adapter Abstraction (`ProtocolAdapter`)**: Clean interface enabling interchangeable protocol adapters (RTU, TCP, and future protocols such as MQTT, OPC-UA, or CANopen).
- **Communication Connection Manager (`ConnectionManager`)**: Manages device connection lifecycle (`DISCONNECTED`, `CONNECTING`, `CONNECTED`, `RECONNECTING`, `ERROR`), exponential retry backoffs, thread-safe pooling, and strict per-device isolation.
- **Flexible Register Addressing Resolution**: Normalizes Modicon 5-digit / 6-digit conventional addresses (4x Holding Registers, 3x Input Registers, 1x Discrete Inputs, 0x Coils) into standard 0-based PDU offsets.
- **Multiformat Register Decoder & Endian Transposition**: Supports `BOOL`, `INT16`, `UINT16`, `INT32`, `UINT32`, `FLOAT32`, `FLOAT64`, and `STRING` with full byte/word swapping orders (`ABCD` Big Endian, `CDAB` Word Swap, `BADC` Byte Swap, `DCBA` Little Endian).
- **Parameter Mapping & Scaling**: Linear transformation `engineering_value = (raw_value * scale) + offset` with decimal precision rounding.
- **Isolated Device Polling Workers (`PollingEngine`)**: Goroutine-isolated scheduling per active device preventing slow or unresponsive devices from blocking or starving other channels.
- **Realtime WebSocket Synchronization & Logging**: Event broadcasting (`device.connection.changed`, `device.communication.success`, `device.communication.error`) and structured diagnostic logging.
- **Diagnostic REST APIs & Interactive UI**: Live manual connection controls (Connect, Disconnect, Reconnect, Test Link) and on-demand parameter register test reads with hex/decimal payload inspector.

---

## 2. Architecture Overview

```
                     +---------------------------------------+
                     |           Device Management           |
                     |         (Device & Parameters)         |
                     +---------------------------------------+
                                         |
                                         v
                     +---------------------------------------+
                     |       ConnectionManager (Pool)        |
                     | - State Machine: CONNECTED/RECONNECT  |
                     | - Device Isolation & Mutexes          |
                     | - Backoff & Reconnect Strategy        |
                     +---------------------------------------+
                                         |
                                         v
                     +---------------------------------------+
                     |            ProtocolAdapter            |
                     |  <<Connect, Disconnect, Read, Check>> |
                     +---------------------------------------+
                                    /         \
                                   /           \
                                  v             v
            +---------------------------+   +---------------------------+
            |     ModbusTCPAdapter      |   |     ModbusRTUAdapter      |
            | - MBAP Framing            |   | - Serial RS485 / RS232    |
            | - Transaction ID Counter  |   | - CRC-16 Generation/Check |
            | - Socket Reuse & Timeout  |   | - Framing & Silence Delays|
            +---------------------------+   +---------------------------+
                                  \             /
                                   \           /
                                    v         v
                     +---------------------------------------+
                     |          Modbus Frame Parser          |
                     | - FC 01: Read Coils                   |
                     | - FC 02: Read Discrete Inputs         |
                     | - FC 03: Read Holding Registers       |
                     | - FC 04: Read Input Registers         |
                     | - Modbus Exception Code Decoder       |
                     +---------------------------------------+
                                         |
                                         v
                     +---------------------------------------+
                     |      Register Decoder & Endianness    |
                     | - ABCD / CDAB / BADC / DCBA           |
                     | - BOOL, INT16, UINT16, INT32,         |
                     |   UINT32, FLOAT32, FLOAT64, STRING    |
                     +---------------------------------------+
                                         |
                                         v
                     +---------------------------------------+
                     |     Parameter Mapping & Scaling       |
                     | engineering_value = (raw * scale) + offset |
                     +---------------------------------------+
                                         |
                                         v
                     +---------------------------------------+
                     |        Polling & Telemetry Cache      |
                     | - Updates LastSeenAt & LastDataAt     |
                     | - Publishes WebSocket Events          |
                     | - Ready for Phase 3 Data Pipeline     |
                     +---------------------------------------+
```

---

## 3. Protocol Adapter Interface

Located in `internal/communication/modbus/types.go` and exported via `internal/communication/adapter.go`:

```go
type ProtocolAdapter interface {
    Connect(ctx context.Context) error
    Disconnect() error
    IsConnected() bool
    Read(ctx context.Context, req ModbusReadRequest) (*ModbusReadResponse, error)
    HealthCheck(ctx context.Context) (time.Duration, error)
    GetStatus() AdapterStatus
}
```

### Supported Function Codes (Read Operations)
| Function Code | Modbus Name | Register Type | Data Unit |
| :--- | :--- | :--- | :--- |
| **0x01** | Read Coils | `COIL` (0x) | Single Bit (0 or 1) |
| **0x02** | Read Discrete Inputs | `DISCRETE_INPUT` (1x) | Single Bit (0 or 1) |
| **0x03** | Read Holding Registers | `HOLDING_REGISTER` (4x) | 16-bit Word(s) |
| **0x04** | Read Input Registers | `INPUT_REGISTER` (3x) | 16-bit Word(s) |

---

## 4. Modbus RTU Implementation

The RTU master is implemented in `internal/communication/modbus/rtu.go`:
- **Serial Port Framing**: Configurable Baud Rate (2400 to 115200), Data Bits (7 or 8), Parity (None `N`, Even `E`, Odd `O`), Stop Bits (1 or 2).
- **CRC-16 Calculation & Validation**: Pure-Go lookup algorithm using polynomial `0xA001` (`internal/communication/modbus/crc.go`). All incoming responses must pass CRC validation before byte parsing. Frames failing CRC are rejected immediately with a descriptive error.
- **Serial Transport Abstraction (`SerialTransport`)**: Decouples the physical serial hardware (`github.com/goburrow/serial`) from the frame engine. Enables deterministic, zero-hardware automated unit testing via in-memory mock serial pipes (`MockSerialPipe`).
- **Timing & Deadlines**: Implements per-request read/write deadlines calculated from the connection timeout, preventing hangs on disconnected or broken RS485 lines.

---

## 5. Modbus TCP Implementation

The TCP client is implemented in `internal/communication/modbus/tcp.go`:
- **MBAP Header Formatting**: 7-byte MBAP header (Transaction ID: 2 bytes, Protocol ID: 2 bytes `0x0000`, Length: 2 bytes, Unit ID / Slave ID: 1 byte).
- **Atomic Transaction Identification**: Atomic rolling uint16 counter ensures matched request/response pairs even in high-throughput or interleaved scenarios.
- **Connection Keepalive & Reuse**: Persistent TCP socket is maintained across read iterations to avoid socket exhaustion on edge Linux / Windows gateways.
- **Graceful Error Recovery**: Sockets encountering I/O errors or network timeouts are safely closed and automatically flagged for reconnect.

---

## 6. Connection Manager & Device Isolation

The Connection Manager is implemented in `internal/communication/manager.go`:
- **State Machine**:
  - `DISCONNECTED`: Adapter idle, no network socket or serial COM port open.
  - `CONNECTING`: Physical handshake or dial underway.
  - `CONNECTED`: Adapter verified healthy and ready for transactions.
  - `RECONNECTING`: Connection lost; background backoff attempt active.
  - `ERROR`: Maximum retries exhausted or fatal configuration fault.
- **Exponential Backoff**: When errors occur, retry delays scale from 500ms up to 30s to prevent tight polling storms.
- **Thread-Safe Pooling**: Device connections are stored in a synchronized map protected by `sync.RWMutex`.
- **Strict Device Isolation**: Each device maintains its own mutex and adapter instance. If Device A is offline or timing out on a 3000ms delay, Device B and Device C continue executing at full speed without any latency penalty or lock contention.

---

## 7. Register Addressing & Mapping Convention

Industrial PLCs and sensors commonly use Modicon 1-based convention. The resolver in `internal/communication/modbus/address.go` safely converts addresses to 0-based PDU protocol offsets:

| Input Addressing Format | Detected Register Type | PDU Function Code | Calculated 0-Based Offset |
| :--- | :--- | :--- | :--- |
| `40001` (5-digit) | Holding Register | FC 03 | `0` |
| `40010` (5-digit) | Holding Register | FC 03 | `9` |
| `400001` (6-digit) | Holding Register | FC 03 | `0` |
| `30001` (5-digit) | Input Register | FC 04 | `0` |
| `10001` (5-digit) | Discrete Input | FC 02 | `0` |
| `00001` (5-digit) | Coil | FC 01 | `0` |
| `100` (Direct offset) | Configured Type | User-specified | `100` |

---

## 8. Multi-Format Register Decoder & Endianness

Located in `internal/communication/modbus/decoder.go`:
Supports transposition across 4 byte orders:
- **`ABCD`**: Big Endian (Standard IEEE-754 / Network Byte Order)
- **`CDAB`**: Word Swap (Common in Modicon, ABB, Schneider PLCs)
- **`BADC`**: Byte Swap (Mid-Little Endian)
- **`DCBA`**: Little Endian (True Little Endian / Intel architecture)

Supported Data Types:
- **`BOOL` / `BOOLEAN`**: Reads coil or discrete bit state (1 = true, 0 = false).
- **`INT16`**: Signed 16-bit integer (-32,768 to 32,767).
- **`UINT16`**: Unsigned 16-bit integer (0 to 65,535).
- **`INT32`**: Signed 32-bit integer (-2,147,483,648 to 2,147,483,647) spanning 2 registers.
- **`UINT32`**: Unsigned 32-bit integer (0 to 4,294,967,295) spanning 2 registers.
- **`FLOAT32`**: Single-precision IEEE-754 floating point spanning 2 registers.
- **`FLOAT64`**: Double-precision IEEE-754 floating point spanning 4 registers.
- **`STRING`**: ASCII byte sequence decoded up to null terminator or register boundary.

---

## 9. Parameter Mapping & Linear Scaling

Located in `internal/communication/modbus/mapping.go`:
$$\text{engineering\_value} = (\text{raw\_value} \times \text{scale}) + \text{offset}$$

- Preserves numeric precision and rounds to configured decimal places (`precision`).
- Non-numeric or boolean values bypass scaling safely.
- Bound checking validates telemetry against `min_value` and `max_value`.

---

## 10. Polling Engine Architecture

Located in `internal/communication/polling.go`:
- **Goroutine per Device**: Background worker routine scheduled by `time.NewTicker(pollingInterval)`.
- **Cancellation Context**: Gracefully starts and stops workers using `context.WithCancel`.
- **Status Propagation**: Automatically invokes `DeviceService.RecordCommunicationResult` and `DeviceService.UpdateParameterCurrentValue` on success.
- **WebSocket Broadcast**: Emits `device.communication.success` or `device.communication.error` events in real-time.

---

## 11. REST API Endpoints

All endpoints are mounted under `/api/devices/:id/communication` and protected by JWT authentication and RBAC permissions:

| Method | Endpoint | Permission | Description |
| :--- | :--- | :--- | :--- |
| `POST` | `/api/devices/:id/communication/connect` | `device.communication.manage` | Establishes connection and starts polling |
| `POST` | `/api/devices/:id/communication/disconnect` | `device.communication.manage` | Stops polling and closes transport |
| `POST` | `/api/devices/:id/communication/reconnect` | `device.communication.manage` | Closes and immediately re-establishes link |
| `POST` | `/api/devices/:id/communication/test` | `device.communication.test` | Non-disruptive physical link latency probe |
| `GET` | `/api/devices/:id/communication/status` | `device.communication.view` | Retrieves current engine metrics & state |
| `POST` | `/api/devices/:id/communication/parameters/:paramId/test-read` | `device.communication.test` | Single-shot diagnostic register read & decode |

---

## 12. Frontend User Interface Integration

- **Device Detail View (`DeviceDetailView.vue`)**:
  - Live communication control strip: Connect, Disconnect, Reconnect, Test Link.
  - Real-time diagnostic cards: Engine State, Latency (ms), Retries count, Last Error.
  - Modbus Slave ID (#1-#247) and Byte Order (`ABCD`, `CDAB`, `BADC`, `DCBA`) indicators.
  - Real-time WebSocket event listener dynamically updating parameter values and connection health without browser reload.
- **Parameter Diagnostic Inspector Modal**:
  - Triggered via "Test Read" button on any parameter row.
  - Displays Modbus Function Code, PDU register offset, raw byte payload in hex, decoded engineering value with units, and roundtrip latency.
  - "Read Again" button allows instant re-triggering for live debugging.
- **Device & Parameter Configuration Modals**:
  - `DeviceModal.vue`: Configures Slave ID and default Byte Order.
  - `ParameterModal.vue`: Configures parameter-level Byte Order overrides.

---

## 13. Testing Strategy & Simulation

The test suite runs 100% locally with zero external network or hardware dependencies:
- **`MockModbusServer`**: High-fidelity in-memory TCP server supporting all 4 read function codes, custom register tables, simulated Modbus exceptions (FC + 0x80), network timeouts, and dropped connections.
- **`MockSerialPipe`**: Thread-safe in-memory serial pipe testing RTU CRC validation, frame truncation, and serial port emulation.
- **Automated Test Results**:
  - `internal/communication/modbus`: PASS (CRC, Decoders, Endianness, Address resolution, TCP, RTU)
  - `internal/communication`: PASS (ConnectionManager, Lifecycle, Device isolation, Polling engine)
  - `internal/handler`: PASS (Communication REST APIs, Diagnostic test-read, RBAC permissions)
  - `internal/service`: PASS (Device service, Activity audit logging, Progress recalculation)

---

## 14. Scope Control & Phase 3 Boundaries

In strict compliance with project guidelines, the following advanced data processing features are intentionally deferred to Phase 3:
- Moving average, windowed averaging, and statistical smoothing.
- Anomaly detection, spike filtering, and rate-of-change clamping.
- Telemetry interpolation and data imputation.
- Alarm thresholds and notification dispatchers.
- Cloud telemetry streaming.

Phase 2.2 strictly establishes the communication foundation, connection lifecycle, and register decoding pipeline.
