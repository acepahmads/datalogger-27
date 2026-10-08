# Phase 3.2 — Data Quality & Processing Documentation

## 1. Objective

Phase 3.2 implements an industrial-grade, deterministic, and lightweight **Data Quality & Processing Layer** directly integrated after the telemetry ingestion engine established in Phase 3.1.

The primary objectives achieved are:
1. **Preserve Phase 3.1 Telemetry Pipeline**: Unbroken ingestion from field sensors through the Modbus communication engine to WebSockets and MariaDB.
2. **Standardized Quality State Machine**: Introduction of strict canonical quality states (`GOOD`, `BAD`, `UNCERTAIN`, `STALE`).
3. **Machine-Readable Traceability**: Explicit quality reasons (`ReasonOutOfHardRange`, `ReasonOutOfWarningRange`, `ReasonSpikeDetected`, `ReasonStaleData`, etc.) and multi-flag bitfields.
4. **Range & Anomaly Validation**: Dual-threshold evaluation (Hard Limits $\rightarrow$ `BAD`, Soft Warning Limits $\rightarrow$ `UNCERTAIN`, Anomaly/Spike $\rightarrow$ `UNCERTAIN`).
5. **Lightweight Determinism**: Bounded circular buffer history (maximum 10 points per parameter) with zero statistical overhead or heavy ML dependencies, running smoothly on low-power ARM devices (Raspberry Pi 3/4/5).
6. **Non-Blocking Fault Isolation**: Quality validation failures, database latency, or malformed sensor data never block or crash device workers or the polling engine.
7. **End-to-End Hardware Validation**: Verified with live connected hardware (`AQMS-01` Modbus RTU sensor streaming at 9600 baud, 8-N-1, Slave ID 6).

---

## 2. Architecture & Pipeline

The pipeline extends the existing ingestion flow seamlessly:

```
SENSOR (Modbus RTU / TCP)
    ↓
DEVICE COMMUNICATION WORKER
    ↓
POLLING ENGINE
    ↓
TELEMETRY INGESTION (TelemetryService.Ingest)
    ↓
RAW TELEMETRY (Immutable raw_value, raw_hex, device_timestamp)
    ↓
QUALITY VALIDATION (QualityProcessor.Validate)
    ├─ Type & Numerical Normalization (NaN, Inf, Null checks)
    ├─ Scaling & Offset Preservation (without double scaling)
    ├─ Hard Limit & Soft Warning Limit Checks
    ├─ Timestamp Sanity & Drift Validation
    ├─ Bounded Spike Delta Anomaly Detection
    └─ Duplicate Detection
    ↓
QUALITY RESULT (State: GOOD / BAD / UNCERTAIN / STALE, Reasons, Flags)
    ↓
PROCESSED TELEMETRY (ProcessedValue, ProcessedAt, QualityReason)
    ├─ Latest In-Memory Cache (Thread-safe parameter state)
    ├─ WebSocket Broadcast (`device.telemetry.received`)
    └─ Asynchronous Non-blocking Batch Worker (MariaDB batch insert)
```

### Stale Evaluation Architecture (Zero DB Write Storms)
Stale parameters are detected by an in-memory background worker (`staleDetectionWorker`, ticking every 5 seconds) that inspects timestamps in `latestCache`. When a parameter exceeds its configured `stale_timeout_seconds`, an in-memory state transition to `STALE` is triggered and emitted over WebSocket, avoiding costly continuous database updates when communication drops.

---

## 3. Standardized Quality States

| Quality State | Description | Typical Conditions |
|---|---|---|
| **`GOOD`** | Data is valid, healthy, and passes all rules | Value inside normal warning bounds, timestamp valid, no communication error. |
| **`UNCERTAIN`** | Value received, but confidence is reduced | Soft warning limit exceeded, sudden spike detected, future timestamp, or holding last good value during anomaly grace window. |
| **`BAD`** | Value fundamentally invalid or rejected | Outside hard minimum/maximum limits, NaN, Infinity, communication read timeout, Modbus CRC error, or decoder failure. |
| **`STALE`** | No recent telemetry received | $T_{current} - T_{last\_packet} > T_{stale\_timeout}$. Polling may still run, but sensor is unresponsive. |

---

## 4. Machine-Readable Quality Reasons & Flags

### Canonical Reasons
- `NONE`: No violation; normal healthy telemetry.
- `OUT_OF_HARD_RANGE`: Value is lower than `min_value` or higher than `max_value`.
- `OUT_OF_WARNING_RANGE`: Value is inside hard limits, but outside `warning_low` or `warning_high`.
- `NULL_VALUE`: Reading is null or unparseable.
- `NAN_VALUE`: Value is `NaN` (Not-a-Number).
- `INFINITE_VALUE`: Value is positive or negative infinity.
- `INVALID_TIMESTAMP`: Device or packet timestamp is unparseable or zero.
- `FUTURE_TIMESTAMP`: Packet timestamp is more than 5 minutes ahead of server time.
- `EXCESSIVELY_OLD_TIMESTAMP`: Device timestamp is older than 24 hours.
- `STALE_DATA`: No packet received within `stale_timeout_seconds`.
- `SPIKE_DETECTED`: Absolute change between consecutive readings exceeds `spike_threshold`.
- `DUPLICATE_DATA`: Consecutive readings have identical timestamps and identical values.
- `DECODING_ERROR`: Word swap, byte order, or register type conversion failure.
- `COMMUNICATION_ERROR`: RS-485 serial offline, I/O timeout, or socket connection reset.
- `HOLD_ANOMALY_GRACE`: Active protection grace window holding last valid reading.

### Traceability Flags
- `range_violation`: Hard limit breached.
- `range_warning`: Warning limit breached.
- `spike`: Sudden anomalous step-change detected.
- `stale`: Parameter data is timed out.
- `timestamp_warning`: Future or anomalous timestamp detected.
- `duplicate`: Duplicate sample.
- `comm_error`: Communication error from Modbus driver.
- `held_anomaly`: Shielding database from sensor glitch.

---

## 5. Parameter Quality Configuration

Each parameter in the system can be individually configured via the web UI or REST API:

```json
{
  "quality_validation_enabled": true,
  "processing_enabled": true,
  "min_value": 0.0,
  "max_value": 100.0,
  "warning_low": 10.0,
  "warning_high": 40.0,
  "stale_timeout_seconds": 120,
  "spike_detection_enabled": true,
  "spike_threshold": 15.0,
  "spike_window_size": 3
}
```

- **Sensible Defaults**: Existing parameters default to `stale_timeout_seconds = 120` and `spike_window_size = 3`.
- **Audit Logging**: Any modification to quality configuration records an audit trail entry identifying WHO changed WHAT, the timestamp, and the exact BEFORE and AFTER JSON diff.

---

## 6. Database Changes & Migrations

Database tables were safely extended using idempotent schema migrations with zero data destruction:

### Table `parameters`
- `quality_validation_enabled` (TINYINT/BOOLEAN, default `TRUE`)
- `processing_enabled` (TINYINT/BOOLEAN, default `TRUE`)
- `warning_low` (DOUBLE, nullable)
- `warning_high` (DOUBLE, nullable)
- `stale_timeout_seconds` (INT, default `120`)
- `spike_detection_enabled` (TINYINT/BOOLEAN, default `FALSE`)
- `spike_threshold` (DOUBLE, default `0.0`)
- `spike_window_size` (INT, default `3`)
- `current_processed_value` (DOUBLE, nullable)
- `current_quality_reason` (VARCHAR(64), nullable)
- `current_quality_flags` (VARCHAR(255), nullable)

### Table `raw_data`
- `processed_value` (DOUBLE, nullable)
- `quality_reason` (VARCHAR(64), nullable)
- `quality_flags` (VARCHAR(255), nullable)
- `processed_at` (DATETIME(3), nullable)
- **Composite Index**: `idx_raw_device_param_ts_quality` on `(device_id, parameter_id, timestamp, quality)` for sub-millisecond historical analytics.

---

## 7. REST API Endpoints

### 1. Global Quality Summary
- **Endpoint**: `GET /api/telemetry/quality-summary`
- **RBAC**: Authenticated
- **Response**:
```json
{
  "success": true,
  "data": {
    "total_count": 1,
    "good_count": 1,
    "uncertain_count": 0,
    "bad_count": 0,
    "stale_count": 0,
    "health_percent": 100.0
  }
}
```

### 2. Device Quality Summary
- **Endpoint**: `GET /api/devices/:id/telemetry/quality-summary`
- **RBAC**: `device.view`

### 3. Get Parameter Quality Configuration
- **Endpoint**: `GET /api/devices/:id/parameters/:paramId/quality`
- **RBAC**: `device.view`

### 4. Update Parameter Quality Configuration
- **Endpoint**: `PUT /api/devices/:id/parameters/:paramId/quality`
- **RBAC**: `device.manage`
- **Audit**: Logged to `audit_trails` as `UPDATE_PARAMETER_QUALITY`.

---

## 8. WebSocket Integration

The existing `device.telemetry.received` event has been augmented with quality properties without breaking legacy frontend consumers:

```json
{
  "device_id": 4,
  "parameter_id": 6,
  "parameter_code": "Temperature",
  "value": 26.12,
  "raw_value": 26.115644,
  "quality": "GOOD",
  "quality_reason": "NONE",
  "quality_flags": "",
  "processed_value": 26.12,
  "received_at": "2026-10-08T06:23:42.145Z"
}
```

---

## 9. User Interface Enhancements

1. **Telemetry Monitor View (`TelemetryMonitorView.vue`)**:
   - **Quality KPI Summary Header**: 5 reactive stat cards displaying Connected Devices, Overall Health Index %, Good Sensors, Uncertain/Bad Sensors, and Stale Sensors.
   - **Quality Filter Dropdown**: Supports filtering by `ALL`, `GOOD`, `BAD`, `UNCERTAIN`, and `STALE`.
   - **Traceability Reasons**: Sensor cards and table view display quality badges alongside machine-readable reasons (e.g. `OUT_OF_WARNING_RANGE`, `SPIKE_DETECTED`).
2. **Device Detail View (`DeviceDetailView.vue`)**:
   - **Live Telemetry Quality Bar**: Real-time breakdown of Good, Uncertain, Bad, and Stale parameters for the active device.
   - **Trend Curve Points**: SVG chart dots dynamically colored according to quality state (Blue for Good, Amber for Uncertain, Red for Bad, Purple for Stale).
   - **Floating Crosshair Tooltip**: Displays the exact quality and quality reason for historical and live samples.
   - **Parameter Limits Column**: Displays both Hard Limits (`min..max`) and Soft Warning Limits (`warn_low..warn_high`).
3. **Parameter Modal (`ParameterModal.vue`)**:
   - Dedicated "Data Quality & Processing" configuration section for warning bounds, stale timeout, and spike detection thresholds.
4. **Theme & Internationalization**:
   - 100% key parity between English (`en.js`) and Indonesian (`id.js`) (998 keys each).
   - Full support for Dark, Light, and System themes.

---

## 10. Automated Acceptance Test Results

An automated acceptance suite in `internal/quality/phase3_2_acceptance_test.go` verifies all 25 criteria:

| # | Acceptance Criterion | Test Result |
|---|---|---|
| 1 | GOOD Value Validation | **PASS** |
| 2 | Below Hard Minimum ($\rightarrow$ `BAD`) | **PASS** |
| 3 | Above Hard Maximum ($\rightarrow$ `BAD`) | **PASS** |
| 4 | Below Warning Minimum ($\rightarrow$ `UNCERTAIN`) | **PASS** |
| 5 | Above Warning Maximum ($\rightarrow$ `UNCERTAIN`) | **PASS** |
| 6 | NULL / Read Error Handling ($\rightarrow$ `BAD`) | **PASS** |
| 7 | Invalid Numeric Handling | **PASS** |
| 8 | NaN Value Handling ($\rightarrow$ `BAD`, `ReasonNaNValue`) | **PASS** |
| 9 | Infinity Value Handling ($\rightarrow$ `BAD`, `ReasonInfiniteValue`) | **PASS** |
| 10 | Valid Timestamp Verification | **PASS** |
| 11 | Future Timestamp Drift ($\rightarrow$ `UNCERTAIN`, `ReasonFutureTimestamp`) | **PASS** |
| 12 | Stale Data Detection ($\rightarrow$ `STALE`, `ReasonStaleData`) | **PASS** |
| 13 | Bounded Spike Anomaly Detection ($\rightarrow$ `UNCERTAIN`, `ReasonSpikeDetected`) | **PASS** |
| 14 | Normal Gradual Change (No false positive spike) | **PASS** |
| 15 | Duplicate Telemetry Detection | **PASS** |
| 16 | Scaling Multiplier Preservation | **PASS** |
| 17 | Offset Adder Preservation | **PASS** |
| 18 | Quality Reason & Flags Traceability | **PASS** |
| 19 | WebSocket Quality Payload Fields | **PASS** |
| 20 | API Quality Summary Response Structure | **PASS** |
| 21 | Multiple Devices Isolation (Spike memory isolated per device) | **PASS** |
| 22 | Multiple Parameters Isolation (Failure of one param does not affect another) | **PASS** |
| 23 | Processor Failure Isolation (Panic resilience on nil inputs) | **PASS** |
| 24 | Non-Blocking Database Delay Simulation (100 concurrent evaluations < 100ms) | **PASS** |
| 25 | Restart & Recovery State Reset | **PASS** |

**Acceptance Test Result**: **25/25 PASS (100%)**

---

## 11. Real Hardware Validation Evidence

- **Target Device**: `AQMS-01` (Ambient Air Quality Monitoring Sensor)
- **Protocol**: Modbus RTU over RS-485 Serial (9600 Baud, 8-N-1, Slave ID 6)
- **Acquired Parameter**: `Temperature` (`FLOAT32`, Register 4, Byte Order `CDAB`)
- **Observed Live Telemetry**:
  - `Raw Hex`: `EC D7 41 D0`
  - `Raw Value`: `26.115644454956055`
  - `Processed Value`: `26.12`
  - `Engineering Unit`: `°C`
  - `Quality`: `GOOD`
  - `Quality Reason`: `NONE`
  - `Received At`: `2026-10-08T06:23:42.145742669Z`
  - `Polling Interval`: 1000ms
  - `Latency`: 12ms

Real sensor communication remains continuous, stable, and correctly processed through the quality pipeline.

---

## 12. Regression Test Results

- **Phase 1 (Foundation)**: 17/17 PASS (100%)
- **Phase 2.1 (Device Management)**: 22/22 PASS (100%)
- **Phase 2.2 (Modbus RTU/TCP)**: 27/27 PASS (100%)
- **Phase 2.3 (Reliability & Hardening)**: 32/32 PASS (100%)
- **Phase 3.1 (Pipeline & Ingestion)**: PASS (100%)
- **Go Test Suite (`go test ./...`)**: PASS (0 failed packages)
- **Go Vet (`go vet ./...`)**: PASS (0 warnings)
- **Frontend Build (`npm run build`)**: PASS (2.08s, 0 errors)
- **Locale Parity (`en.js` vs `id.js`)**: PASS (998 keys each, 0 missing keys)

---

## 13. Progress Tracking Dashboard

In accordance with project requirements, the Development Progress Dashboard records Phase 3.2 as **DONE (100%)**:
- **Phase 3.1**: `DONE 100%` (16 subtasks)
- **Phase 3.2**: `DONE 100%` (22 subtasks)
- **Phase 3.3**: `PLANNED` (Downsampling & Aggregation)
- **Phase 3 Overall**: Progress recalculated automatically based on subtask weights.

---

## 14. Next Recommended Phase

Proceed to **Phase 3.3 — Aggregation, Rollup & Downsampling**:
- Periodic rollup cron workers (1-minute, 5-minute, 1-hour, 1-day rollups: Min, Max, Avg, Count, Sum).
- Automatic downsampling queries for long-range historical charting to maintain 60 FPS performance on resource-constrained devices.
- Telemetry table partitioning and retention policies.
