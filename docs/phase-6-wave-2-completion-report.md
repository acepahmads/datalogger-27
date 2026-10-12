# Phase 6 — Wave 2 Completion Report: Device Health, Communication Errors & Data Abnormality

## Executive Summary

Phase 6 — Wave 2 delivers robust edge device reliability, communication fault detection, and telemetry quality anomaly alerting into the centralized `AlarmEvaluator` engine. Building directly upon the Wave 1 foundational alarm evaluation pipeline, Wave 2 implements **Task #60 (Device Offline)**, **Task #61 (Communication Error)**, and **Task #62 (Data Abnormality)**.

All detection rules, thresholds, deduplication, latching, and automatic recovery semantics have been implemented and verified via deterministic test suites, ensuring zero false alarms during administrative maintenance, reboots, or transient communication blips.

---

## 1. Implemented Detection Rules & Configurable Defaults

### Task #60: Device Offline (`DEVICE_OFFLINE`)
- **Severity**: `CRITICAL`
- **Scope**: Device-level (`DeviceID != nil`, `ParameterID == nil`).
- **Offline Timeout**: Configurable via `AlarmEvaluatorConfig.DeviceOfflineTimeout` (Default: `60s`).
- **Startup Grace Period**: Configurable via `AlarmEvaluatorConfig.StartupGracePeriod` (Default: `30s`).
  - Devices that have never connected yet do not trigger false alarms during system startup or initial discovery.
- **Administrative Disablement Suppression**:
  - Devices with `Enabled == false`, `Status != DeviceStatusActive`, `Connection.Enabled == false`, or in `StateDisabled` are suppressed from triggering offline alarms.
  - If a device is administratively disabled while an active offline alarm exists, the alarm is automatically cleared.
- **Context Recorded**:
  - Device code, outage duration (e.g. `1m15s`), timestamp of last successful communication, and failure category (e.g. `TIMEOUT`, `CONNECTION_REFUSED`).

### Task #61: Communication Error (`COMMUNICATION_ERROR`)
- **Severity**: `WARNING`
- **Scope**: Device-level (`DeviceID != nil`, `ParameterID == nil`).
- **Consecutive Failure Threshold**: Configurable via `AlarmEvaluatorConfig.CommErrorConsecutiveLimit` (Default: `5` consecutive failures).
  - Individual, transient failed requests (attempts 1–4) are suppressed to prevent noisy alarm flapping.
- **Deduplication & Anti-Storm Cooldown**:
  - One latched alarm entry is tracked in-memory per device episode. Failures 6, 7, ... do not create duplicate alarms.
- **Supported Error Categories**:
  - `TIMEOUT`: Request deadline exceeded / socket timeout.
  - `CONNECTION_REFUSED`: Target unreachable, socket reset, broken pipe.
  - `PROTOCOL_EXCEPTION`: Modbus exception responses, CRC checksum mismatches, framing errors.
  - `COMMUNICATION_FAILURE`: Adapter I/O failure or unexpected transport disconnect.
- **Recovery Condition**:
  - Resolves when consecutive successful communication reads $\ge$ `CommRecoverySuccessLimit` (Default: `2` consecutive successes) or when connection state reaches `CONNECTED`.

### Task #62: Data Abnormality (`DATA_ABNORMALITY_*`)
- **Severity**: `WARNING`
- **Scope**: Parameter-specific (`DeviceID != nil`, `ParameterID != nil`). Evaluates processed engineering telemetry values.
- **Sub-Types & Thresholds**:
  1. **Stuck Value (`DATA_ABNORMALITY_STUCK_VALUE`)**:
     - Unchanged readings over configurable count (`StuckValueCountThreshold`, default: `10` consecutive samples) **AND** duration (`StuckValueDurationThreshold`, default: `30s`).
     - Tolerance deadband: `StuckTolerance` (default: `1e-6`).
     - Legitimate constant readings below count/duration do not trigger false alarms.
     - Recovery: Value changes by $> \text{tolerance}$ -> alarm automatically clears.
     - Non-numeric inputs (`NaN`, `Inf`) and `BAD` quality samples are excluded from stuck comparison.
  2. **Repeated BAD Quality (`DATA_ABNORMALITY_BAD_QUALITY`)**:
     - Triggered when consecutive `QualityBad` samples reach `BadQualityCountThreshold` (default: `5` samples).
     - Context captures decoding errors, sensor disconnection, or hard range violations.
     - Recovery: Arrival of a `QualityGood` sample automatically clears the alarm.
  3. **Repeated Stale Data (`DATA_ABNORMALITY_STALE_DATA`)**:
     - Triggered when consecutive `QualityStale` samples reach `StaleDataCountThreshold` (default: `3` samples).
     - Recovery: Fresh `QualityGood` sample clears the alarm.
  4. **Spike / Outlier (`DATA_ABNORMALITY_SPIKE`)**:
     - Triggered immediately on detection of anomalous sudden jumps flagged by the quality processor (`FlagSpike` / `ReasonSpikeDetected`).
     - Recovery: Cleared after 3 consecutive normal (non-spike) readings.

---

## 2. Trigger and Recovery Semantics Summary

| Alarm Code | Severity | Trigger Criteria | Anti-Flapping / Cooldown | Recovery Condition |
| :--- | :--- | :--- | :--- | :--- |
| `DEVICE_OFFLINE` | `CRITICAL` | `time.Since(LastCommSuccess) >= 60s` OR `State == StateError` | Startup grace period (30s) + 1 alarm per episode | `ConsecutiveSuccesses >= 1` or `StateConnected` |
| `COMMUNICATION_ERROR` | `WARNING` | Consecutive failures $\ge 5$ | Individual failures 1–4 ignored; deduplicated per device | `ConsecutiveSuccesses >= 2` or `StateConnected` |
| `DATA_ABNORMALITY_STUCK_VALUE` | `WARNING` | Identical value for $\ge 10$ samples **AND** $\ge 30\text{s}$ | Count + Duration dual criteria; tolerance deadband | Telemetry value changes by $> 10^{-6}$ |
| `DATA_ABNORMALITY_BAD_QUALITY` | `WARNING` | Consecutive `QualityBad` samples $\ge 5$ | Single or transient bad samples ignored | Receipt of `QualityGood` sample |
| `DATA_ABNORMALITY_STALE_DATA` | `WARNING` | Consecutive `QualityStale` samples $\ge 3$ | Stale debounce count | Receipt of fresh `QualityGood` sample |
| `DATA_ABNORMALITY_SPIKE` | `WARNING` | Quality processor detects spike outlier | Rate-limited per spike episode | 3 consecutive normal readings |

---

## 3. Files Changed

| File | Change Description |
| :--- | :--- |
| [`internal/service/alarm_evaluator.go`](file:///d:/cbi-project-src/datalogger-27/internal/service/alarm_evaluator.go) | Added Wave 2 alarm codes, `DeviceHealthInput`, `paramAbnormalityState`, `EvaluateDeviceOffline`, `EvaluateDeviceCommFailure`, `EvaluateDeviceCommSuccess`, `evaluateDataAbnormality`, `ClearDeviceAlarms`, and multi-category `ReloadActiveAlarms` DB reconciliation. |
| [`internal/communication/polling.go`](file:///d:/cbi-project-src/datalogger-27/internal/communication/polling.go) | Injected `AlarmEvaluator` into `PollingEngine`; hooked comm failure and success evaluations in `pollDevice`; hooked device offline watchdog in `runWorker`; added immediate alarm cleanup in `StopDeviceWorker`. |
| [`cmd/main.go`](file:///d:/cbi-project-src/datalogger-27/cmd/main.go) | Wired `pollingEngine.SetAlarmEvaluator(alarmEvaluator)`. |
| [`internal/database/migration.go`](file:///d:/cbi-project-src/datalogger-27/internal/database/migration.go) | Added `verifiedTasksWave2` to `updatePhase6Tracking`, marking Tasks #60, #61, and #62 as `DONE` (100%) and advancing Phase 6 progress to 70%. |
| [`internal/database/migration_test.go`](file:///d:/cbi-project-src/datalogger-27/internal/database/migration_test.go) | Added `TestMigrationPhase6Wave2Tracking` asserting Phase 6 progress $\ge 70\%$ and Tasks #60, #61, #62 `DONE`. |
| [`internal/service/alarm_evaluator_test.go`](file:///d:/cbi-project-src/datalogger-27/internal/service/alarm_evaluator_test.go) | Added 12 deterministic unit tests for Wave 2 covering offline timeouts, disabled suppression, comm error thresholds, stuck values, bad quality, stale data, spikes, restart reconciliation, and concurrent execution. |
| `bin/datalogger*` | Rebuilt production edge binaries for Windows AMD64, Linux ARM64, Linux AMD64, and Linux ARMv7. |

---

## 4. Automated Test Evidence

### A. AlarmEvaluator Suite (`internal/service`)
```text
=== RUN   TestAlarmEvaluator_NormalValues
--- PASS: TestAlarmEvaluator_NormalValues (0.00s)
=== RUN   TestAlarmEvaluator_WarningThresholds
--- PASS: TestAlarmEvaluator_WarningThresholds (0.00s)
=== RUN   TestAlarmEvaluator_CriticalPrecedence
--- PASS: TestAlarmEvaluator_CriticalPrecedence (0.00s)
=== RUN   TestAlarmEvaluator_EscalationWarningToCritical
--- PASS: TestAlarmEvaluator_EscalationWarningToCritical (0.00s)
=== RUN   TestAlarmEvaluator_DeescalationCriticalToWarning
--- PASS: TestAlarmEvaluator_DeescalationCriticalToWarning (0.00s)
=== RUN   TestAlarmEvaluator_HysteresisDeadband
--- PASS: TestAlarmEvaluator_HysteresisDeadband (0.00s)
=== RUN   TestAlarmEvaluator_RepeatedSamplesAntiDuplication
--- PASS: TestAlarmEvaluator_RepeatedSamplesAntiDuplication (0.00s)
=== RUN   TestAlarmEvaluator_ConcurrentTelemetrySafety
--- PASS: TestAlarmEvaluator_ConcurrentTelemetrySafety (0.00s)
=== RUN   TestAlarmEvaluator_DataQualityRejection
--- PASS: TestAlarmEvaluator_DataQualityRejection (0.00s)
=== RUN   TestAlarmEvaluator_UncertainQualityAnnotation
--- PASS: TestAlarmEvaluator_UncertainQualityAnnotation (0.00s)
=== RUN   TestAlarmEvaluator_StartupReloadFromDB
--- PASS: TestAlarmEvaluator_StartupReloadFromDB (0.00s)
=== RUN   TestAlarmEvaluator_AcknowledgedPreservationOnRecovery
--- PASS: TestAlarmEvaluator_AcknowledgedPreservationOnRecovery (0.00s)
=== RUN   TestTelemetryService_IngestAlarmIntegration
--- PASS: TestTelemetryService_IngestAlarmIntegration (0.00s)
=== RUN   TestAlarmEvaluator_DeviceOfflineTimeoutAndRecovery
--- PASS: TestAlarmEvaluator_DeviceOfflineTimeoutAndRecovery (0.00s)
=== RUN   TestAlarmEvaluator_DeviceOfflineDisabledSuppression
--- PASS: TestAlarmEvaluator_DeviceOfflineDisabledSuppression (0.00s)
=== RUN   TestAlarmEvaluator_DeviceOfflineRepeatedChecksNoDuplicate
--- PASS: TestAlarmEvaluator_DeviceOfflineRepeatedChecksNoDuplicate (0.00s)
=== RUN   TestAlarmEvaluator_CommErrorThresholdAndCooldown
--- PASS: TestAlarmEvaluator_CommErrorThresholdAndCooldown (0.00s)
=== RUN   TestAlarmEvaluator_CommErrorCategories
--- PASS: TestAlarmEvaluator_CommErrorCategories (0.00s)
=== RUN   TestAlarmEvaluator_DataAbnormalityStuckValue
--- PASS: TestAlarmEvaluator_DataAbnormalityStuckValue (0.00s)
=== RUN   TestAlarmEvaluator_ConstantReadingSafeThresholds
--- PASS: TestAlarmEvaluator_ConstantReadingSafeThresholds (0.00s)
=== RUN   TestAlarmEvaluator_DataAbnormalityBadQuality
--- PASS: TestAlarmEvaluator_DataAbnormalityBadQuality (0.00s)
=== RUN   TestAlarmEvaluator_DataAbnormalityStaleData
--- PASS: TestAlarmEvaluator_DataAbnormalityStaleData (0.00s)
=== RUN   TestAlarmEvaluator_DataAbnormalitySpike
--- PASS: TestAlarmEvaluator_DataAbnormalitySpike (0.00s)
=== RUN   TestAlarmEvaluator_RestartReconciliationWave2
--- PASS: TestAlarmEvaluator_RestartReconciliationWave2 (0.00s)
=== RUN   TestAlarmEvaluator_ConcurrentWave2Safety
--- PASS: TestAlarmEvaluator_ConcurrentWave2Safety (0.01s)
PASS
ok  	datalogger/internal/service	3.657s
```

### B. Full Test Suite (`go test ./...`)
```text
ok  	datalogger/internal/aggregation	3.120s
ok  	datalogger/internal/communication	8.432s
ok  	datalogger/internal/communication/modbus	(cached)
ok  	datalogger/internal/database	2.201s
ok  	datalogger/internal/handler	12.292s
ok  	datalogger/internal/quality	(cached)
ok  	datalogger/internal/queue	(cached)
ok  	datalogger/internal/service	16.404s
ok  	datalogger/pkg/formula	(cached)
```
- **Go Vet**: `go vet ./...` completed with zero warnings or errors.
- **Frontend Build**: `npm run build` in `web/` completed in 3.35s with zero bundle errors.

---

## 5. Dashboard Status Before & After Wave 2

### Development Progress Dashboard (`http://192.168.1.53:8080`)

| Metric | Before Wave 2 | After Wave 2 |
| :--- | :--- | :--- |
| **Phase 6 Status** | `WORKING` | `WORKING` |
| **Phase 6 Progress** | 40.0% (4/10 tasks) | **70.0% (7/10 tasks)** |
| **Task #57 (Threshold)** | `DONE` (100%) | `DONE` (100%) |
| **Task #58 (Warning)** | `DONE` (100%) | `DONE` (100%) |
| **Task #59 (Critical)** | `DONE` (100%) | `DONE` (100%) |
| **Task #60 (Device Offline)** | `PENDING` (0%) | **`DONE` (100%)** |
| **Task #61 (Communication Error)** | `PENDING` (0%) | **`DONE` (100%)** |
| **Task #62 (Data Abnormality)** | `PENDING` (0%) | **`DONE` (100%)** |
| **Task #63 (Alarm Engine)** | `DONE` (100%) | `DONE` (100%) |
| **Task #64 (Notification Engine)** | `PENDING` (0%) | `PENDING` (0%) [Wave 3] |
| **Task #65 (Alarm Acknowledgement)**| `PENDING` (0%) | `PENDING` (0%) [Wave 4] |
| **Task #66 (Alarm History)** | `PENDING` (0%) | `PENDING` (0%) [Wave 4] |

---

## 6. Known Limitations & Physical-Device Verification Status

1. **Physical Serial Fault Injection**:
   - Unit tests use realistic mock transport errors (`read tcp: i/o timeout`, `connection reset by peer`, `modbus crc exception checksum error`). Full hardware-level physical RS485 bus cable disconnect testing should be validated in end-to-end site acceptance tests.
2. **Dynamic Abnormality Tuning per Parameter**:
   - Parameters currently inherit the engine defaults (`StuckValueCountThreshold = 10`, `StuckValueDurationThreshold = 30s`, `StuckTolerance = 1e-6`) unless customized in `AlarmEvaluatorConfig`. Per-parameter threshold overrides stored directly in `model.Parameter` (if required) can be extended in future revisions without breaking the evaluator API.
3. **Wave 3 & Wave 4 Scope Boundary**:
   - Outbound dispatching via email/SMS/webhooks/push (Tasks #64) and UI historical query viewers (Tasks #65, #66) remain in their scheduled waves and have not been modified.

---

## 7. Git Metadata

- **Commit**: `3297d50`
- **Commit Message**: `feat(phase-6): implement wave 2 device offline, communication errors, and data abnormality`
- **Branch**: `main`
- **Local Time**: 2026-10-12 10:14:31 +07:00
