# Phase 3.3 — Data Aggregation, Rollup & Downsampling Specification

## 1. Objective

Phase 3.3 implements a deterministic, multi-domain Data Aggregation, Rollup, and Downsampling Engine for the industrial datalogger application. The primary objectives are:
- **Dual-Domain Aggregation Architecture**: Decouple technical engineering telemetry (`INTERNAL_RAW`) from customer-facing measurements (`CUSTOMER_PROCESSED`).
- **Deterministic Time Bucketing**: Align time windows to reproducible boundaries based on configured intervals and timezones.
- **Stable Period Identifiers**: Generate `YYYYMMDDHHmmss` identifiers anchored strictly to the bucket's `period_start`.
- **Reusable Aggregation Functions**: Support `AVG`, `MIN`, `MAX`, `SUM`, `COUNT`, `FIRST`, and `LAST` with mathematically sound handling of `NULL`, `NaN`, `Infinity`, and quality degradation.
- **Quality Aggregation**: Maintain precise counts of samples by quality status (`good`, `uncertain`, `bad`, `stale`) and derive an overall bucket quality state.
- **Idempotency & Late Data Ingestion**: Guarantee atomic UPSERT semantics and support late-arriving telemetry recalculations within grace periods.
- **Historical Downsampling Engine**: Accelerate historical charts and queries with dynamic multi-resolution rollups (`auto`, `raw`, `5m`, `30m`, `1h`, `1d`) without client-side query complexity.
- **Customer-Safe Data Delivery**: Provide dedicated REST endpoints that strictly conceal raw Modbus registers, hex frames, and internal debugging metadata.

---

## 2. Internal Raw Data Definition

Internal Raw Telemetry represents the unscaled, direct numerical readings or register values acquired from physical field devices (Modbus RTU, Modbus TCP, analog transducers):
- **Source Field**: `telemetries.raw_value` (or `raw_packet` frame logs).
- **Target Audience**: Automation engineers, hardware diagnosticians, technical auditors, and maintenance crews.
- **Use Cases**: Register validation, calibration drift analysis, communication noise detection, sensor failure forensics, and raw signal trend verification.
- **Access Restrictions**: Restricted to internal engineering/admin roles (`device.view`, `device.manage`). Never exposed to public customer API endpoints.

---

## 3. Customer Data Definition

Customer Data represents clean, meaningful industrial engineering measurements derived through the Phase 3.2 processing pipeline:
$$\text{Raw Value} \xrightarrow{\text{Decoding}} \text{Numeric} \xrightarrow{\text{Scale \& Offset}} \text{Scaled} \xrightarrow{\text{Formula Engine}} \text{Customer Value}$$

- **Source Field**: `telemetries.processed_value` (with Phase 3.2 quality state and reasons attached).
- **Target Audience**: Plant operators, executive dashboards, billing engines, customer integration APIs, and compliance reporters.
- **Safety Guarantee**: Strictly hides Modbus register addresses, function codes, raw hex payloads, and communication retries.

---

## 4. Aggregation Source

The engine supports two explicit source domains defined via enum `AggregationSourceType`:
1. `INTERNAL_RAW`: Aggregates the raw sensor register value (`telemetries.raw_value`).
2. `CUSTOMER_PROCESSED`: Aggregates the customer-facing processed measurement (`telemetries.processed_value`).

The source domain is locked per `AggregationDefinition`. The calculation engine guarantees that `CUSTOMER_PROCESSED` definitions will never read from `raw_value`.

---

## 5. Processing Relationship

```
                       MODBUS SENSOR / PLC
                               │
                               ▼
                       INTERNAL RAW DATA
                       (telemetries.raw_value)
                               │
                ┌──────────────┴──────────────┐
                │                             │
                ▼                             ▼
       INTERNAL AGGREGATION          PHASE 3.2 PROCESSING
      (Engineering Diagnostics)      (Scale, Offset & Formula)
                │                             │
                ▼                             ▼
     telemetry_aggregations          CUSTOMER TELEMETRY
    (source: INTERNAL_RAW)         (telemetries.processed_value)
                                              │
                                              ▼
                                     CUSTOMER AGGREGATION
                                    (source: CUSTOMER_PROCESSED)
                                              │
                                              ▼
                                    telemetry_aggregations
                                  (source: CUSTOMER_PROCESSED)
                                              │
                                              ▼
                                    CUSTOMER API DELIVERY
```

**Crucial Invariant**: Customer aggregation consumes the already-processed value produced by Phase 3.2. It never repeats scaling or formula calculations.

---

## 6. Aggregation Functions

The mathematical engine (`internal/aggregation/engine.go`) provides unified evaluation for 7 standard functions:

| Function | Calculation Logic | Invalid / Non-numeric Handling |
| :--- | :--- | :--- |
| `AVG` | $\frac{1}{N_{\text{valid}}} \sum_{i=1}^{N_{\text{valid}}} x_i$ | Ignores `NaN`, `Inf`, and invalid entries. Returns `NULL` if $N_{\text{valid}} = 0$. |
| `MIN` | $\min(x_1, \dots, x_n)$ across valid samples | Ignores invalid samples. Returns `NULL` if $N_{\text{valid}} = 0$. |
| `MAX` | $\max(x_1, \dots, x_n)$ across valid samples | Ignores invalid samples. Returns `NULL` if $N_{\text{valid}} = 0$. |
| `SUM` | $\sum_{i=1}^{N_{\text{valid}}} x_i$ | Ignores invalid samples. Returns `NULL` if $N_{\text{valid}} = 0$. |
| `COUNT` | Total received samples ($N_{\text{total}}$) | Incremented for every telemetry event in the bucket window. |
| `FIRST` | First valid sample ordered chronologically by `timestamp` | Ignores database insertion order; respects device event timestamp. |
| `LAST` | Last valid sample ordered chronologically by `timestamp` | Ignores database insertion order; respects device event timestamp. |

---

## 7. Time Bucket Engine

Every aggregation result belongs to a deterministic, half-open time interval:
$$[\text{period\_start}, \text{period\_end}) \implies \{ t \mid \text{period\_start} \le t < \text{period\_end} \}$$

- **Alignment**:
  $$\text{bucket\_start} = \left\lfloor \frac{t_{\text{unix}}}{I} \right\rfloor \times I$$
  where $I$ is the interval in seconds.
- **Timezone Awareness**: The bucket is computed in the definition's configured timezone (e.g. `Asia/Jakarta` or `UTC`) to respect day and hour boundaries correctly.
- **Grace Period**: Completed buckets are evaluated with an adjustable grace period (default 10 seconds) to ensure buffered telemetry events are committed before bucket closure.

---

## 8. Period Identifier

Every aggregation result possesses a standardized, reproducible period identifier:
$$\text{Identifier} = \text{YYYYMMDDHHmmss}$$
anchored strictly to the bucket's **`period_start`** in the configured timezone.

Examples:
- Bucket `2026-10-08 14:00:00` to `2026-10-08 14:30:00` $\implies$ Identifier: `20261008140000`
- Bucket `2026-10-08 14:30:00` to `2026-10-08 15:00:00` $\implies$ Identifier: `20261008143000`

The identifier is permanent, reproducible, and forms the primary query key for customer integrations.

---

## 9. Interval Configuration

Intervals are normalized and stored as integer seconds (`interval_seconds`):
- Standard Presets: `2m` (120s), `5m` (300s), `10m` (600s), `15m` (900s), `30m` (1800s), `60m` (3600s).
- Custom Intervals: Supported for any duration $\ge 10$ seconds (e.g., 20s, 7200s, 86400s).
- Display: Formatted cleanly in the UI as friendly strings (e.g., "30 Minutes (1800s)").

---

## 10. Quality Aggregation

Quality preservation follows the Phase 3.2 industrial quality model:
- **Sample Metrics**:
  - `sample_count`: Total items received in the bucket window.
  - `good_count`: Count of samples with `Quality == 'GOOD'`.
  - `uncertain_count`: Count of samples with `Quality == 'UNCERTAIN'`.
  - `bad_count`: Count of samples with `Quality == 'BAD'`.
  - `stale_count`: Count of samples flagged `is_stale == true`.
- **Overall Bucket Quality**:
  - If $N_{\text{valid}} == 0$: Overall Quality = `BAD`.
  - If any sample is `BAD` or `UNCERTAIN`: Overall Quality = `UNCERTAIN`.
  - If all valid samples are `GOOD`: Overall Quality = `GOOD`.

---

## 11. Internal Aggregation

Used for technical audits and engineering diagnostics:
- Directly reads `telemetries.raw_value`.
- Computes `min_value`, `max_value`, `avg_value`, `sum_value`, and counts over raw Modbus register data.
- Persisted in `telemetry_aggregations` with `source_type = 'INTERNAL_RAW'`.

---

## 12. Customer Aggregation

Used for customer-facing deliverables:
- Directly reads `telemetries.processed_value`.
- Operates on decoded, scaled, and formula-derived physical units (e.g., °C, bar, kW, m³/h).
- Persisted in `telemetry_aggregations` with `source_type = 'CUSTOMER_PROCESSED'`.
- Customer API endpoints strictly filter by this source type.

---

## 13. Rollup & Idempotency

Rollup results are stored permanently in the `telemetry_aggregations` table:
- **Composite Unique Index**: `(aggregation_definition_id, period_start)`.
- **Atomic UPSERT**: Handled via SQL `ON CONFLICT (aggregation_definition_id, period_start) DO UPDATE SET ...`.
- **Safe Reruns**: Manual triggers or worker re-runs update the existing row in place with zero duplicate records created.

---

## 14. Historical Downsampling

The downsampling engine (`internal/service/aggregation_service.go:GetDownsampledTelemetry`) dynamically picks the optimal data resolution based on requested time range:

| Requested Time Range | Selected Resolution | Underlying Storage |
| :--- | :--- | :--- |
| $\le 2$ Hours | `raw` (full resolution) | `telemetries` table |
| $\le 24$ Hours | `5m` (5-minute rollup) | `telemetry_aggregations` table |
| $\le 7$ Days | `30m` (30-minute rollup) | `telemetry_aggregations` table |
| $\le 30$ Days | `1h` (1-hour rollup) | `telemetry_aggregations` table |
| $> 30$ Days | `1d` (1-day rollup) | `telemetry_aggregations` table |

Users and frontends can also request specific fixed resolutions (`resolution=raw|auto|5m|30m|1h|1d`). If a pre-calculated rollup bucket does not exist, the engine falls back gracefully to raw table queries.

---

## 15. API Reference

### A. Aggregation Definitions
- `GET /api/aggregations/definitions?device_id=:id`
- `POST /api/aggregations/definitions` (requires `device.manage`)
- `PUT /api/aggregations/definitions/:id` (requires `device.manage`)
- `DELETE /api/aggregations/definitions/:id` (requires `device.manage`)

### B. Aggregation Results
- `GET /api/aggregations/results?device_id=:id&source_type=:source&parameter_id=:param&identifier=:id`
- `POST /api/aggregations/run` (Manual bucket trigger; requires `device.manage`)

### C. Downsampled Telemetry
- `GET /api/devices/:id/telemetry/downsampled?resolution=auto|5m|30m|1h|1d&start_time=...&end_time=...`

### D. Customer Aggregated Data API
- `GET /api/customer/aggregated-data/:identifier`
  - Returns clean, customer-safe JSON.
  - Hides Modbus register addresses, hex frames, and hardware details.
  - Returns all parameters matching the period identifier.

Example Response:
```json
{
  "identifier": "20261008140000",
  "period_start": "2026-10-08T14:00:00+07:00",
  "period_end": "2026-10-08T14:30:00+07:00",
  "data": [
    {
      "parameter": "PM2.5",
      "value": 15.42,
      "unit": "µg/m³",
      "quality": "GOOD",
      "sample_count": 180
    },
    {
      "parameter": "Temperature",
      "value": 26.85,
      "unit": "°C",
      "quality": "GOOD",
      "sample_count": 180
    }
  ]
}
```

---

## 16. Database Schema & Migration

### `aggregation_definitions`
```sql
CREATE TABLE aggregation_definitions (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    device_id BIGINT NOT NULL,
    parameter_id BIGINT NOT NULL,
    name VARCHAR(150) NOT NULL,
    code VARCHAR(100) NOT NULL,
    source_type VARCHAR(50) NOT NULL DEFAULT 'CUSTOMER_PROCESSED',
    function VARCHAR(50) NOT NULL DEFAULT 'AVG',
    interval_seconds INT NOT NULL DEFAULT 1800,
    timezone VARCHAR(50) NOT NULL DEFAULT 'Asia/Jakarta',
    quality_policy VARCHAR(50) NOT NULL DEFAULT 'STRICT',
    identifier_format VARCHAR(50) NOT NULL DEFAULT 'YYYYMMDDHHmmss',
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    INDEX idx_agg_def_dev_param (device_id, parameter_id),
    INDEX idx_agg_def_enabled (enabled)
);
```

### `telemetry_aggregations`
```sql
CREATE TABLE telemetry_aggregations (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    aggregation_definition_id BIGINT NOT NULL,
    device_id BIGINT NOT NULL,
    parameter_id BIGINT NOT NULL,
    source_type VARCHAR(50) NOT NULL,
    period_start DATETIME NOT NULL,
    period_end DATETIME NOT NULL,
    identifier VARCHAR(30) NOT NULL,
    value DOUBLE,
    min_value DOUBLE,
    max_value DOUBLE,
    avg_value DOUBLE,
    sum_value DOUBLE,
    sample_count INT NOT NULL DEFAULT 0,
    good_count INT NOT NULL DEFAULT 0,
    uncertain_count INT NOT NULL DEFAULT 0,
    bad_count INT NOT NULL DEFAULT 0,
    stale_count INT NOT NULL DEFAULT 0,
    quality VARCHAR(20) NOT NULL DEFAULT 'GOOD',
    created_at DATETIME,
    updated_at DATETIME,
    deleted_at DATETIME,
    UNIQUE INDEX idx_agg_def_start (aggregation_definition_id, period_start),
    INDEX idx_agg_identifier (identifier),
    INDEX idx_agg_device_param (device_id, parameter_id),
    INDEX idx_agg_period (period_start, period_end)
);
```

---

## 17. Worker Architecture

The Aggregation Worker (`internal/service/aggregation_service.go:StartWorker`) runs as a controlled background goroutine:
1. **Periodic Evaluation**: Evaluates completed buckets every 30 seconds.
2. **Grace Windowing**: Ignores buckets closing within the last 10 seconds to allow inflight database flushes to commit.
3. **Bounded Catch-Up**: On startup or tick, it evaluates the previous 2 hours of missing or uncompleted buckets.
4. **Clean Shutdown**: Listens to context cancellation and stops immediately during application termination.

---

## 18. Late Data Handling

When telemetry arrives with historical timestamps:
- The system evaluates the target bucket $[\text{period\_start}, \text{period\_end})$.
- If an existing rollup record exists, the engine re-queries the time slice from the primary telemetry table and executes an idempotent database update.
- Grace window parameter (`grace_period_seconds = 120s`) controls how far back late telemetry automatically recalculates rollups.

---

## 19. Restart Recovery

During application startup:
1. AutoMigrate creates and validates all schema indexes.
2. The worker scans all active `aggregation_definitions`.
3. For each definition, it queries the latest completed `period_start` from `telemetry_aggregations`.
4. If gaps exist between the latest record and current time, missing buckets are automatically calculated and committed.
5. Zero duplicates are created due to the composite unique index.

---

## 20. Security & RBAC

- **Admin / Configuration**: `device.manage` permission required to create, update, delete, or trigger aggregation calculations.
- **Results Inspection**: `device.view` permission required to browse engineering aggregation rollups.
- **Customer API**: `customer.view` or general authenticated user role allows querying `/api/customer/aggregated-data/:identifier`. Raw hardware fields are unreachable via customer routes.

---

## 21. Audit Trail

All lifecycle operations on aggregation definitions generate audit events in `model.AuditTrail`:
- `CREATE_AGGREGATION_DEFINITION`
- `UPDATE_AGGREGATION_DEFINITION`
- `DELETE_AGGREGATION_DEFINITION`
- `TRIGGER_AGGREGATION_RUN`

Audit records document `Username`, `IPAddress`, `UserAgent`, `Before`, and `After` configurations.

---

## 22. Performance & Raspberry Pi Optimization

- **No Full-Table Scans**: Aggregation queries use indexed range scans on `(device_id, parameter_id, timestamp)`.
- **Bounded Worker Pool**: A single background goroutine handles aggregation ticks without unbounded goroutine proliferation.
- **In-Memory Boundedness**: Data items are streamed and accumulated incrementally with constant memory overhead.
- **Target Footprint**: CPU usage during rollup $< 2\%$ on ARM Cortex-A72 (Raspberry Pi 4).

---

## 23. Testing & Validation

Automated test suite (`internal/aggregation` and `internal/handler`):
1. Internal Raw Functions: `AVG`, `MIN`, `MAX`, `SUM`, `COUNT`, `FIRST`, `LAST` $\implies$ **PASS**
2. Customer Processed Functions: `AVG`, `MIN`, `MAX`, `SUM`, `COUNT` $\implies$ **PASS**
3. Configurable Intervals: `2m`, `5m`, `30m`, `60m`, Custom seconds $\implies$ **PASS**
4. Deterministic Time Bucketing & Timezones $\implies$ **PASS**
5. Period Identifier Generation & Parsing $\implies$ **PASS**
6. Quality Aggregation (Good, Uncertain, Bad, Stale) $\implies$ **PASS**
7. Zero Valid Sample & NaN/Inf Handling $\implies$ **PASS**
8. Late Telemetry Ingestion & Recalculation $\implies$ **PASS**
9. Idempotent Re-runs with Zero Duplication $\implies$ **PASS**
10. Startup Recovery across Gaps $\implies$ **PASS**
11. Multi-Parameter & Multi-Device Isolation $\implies$ **PASS**
12. Customer API Concealment of Modbus Registers $\implies$ **PASS**
13. Historical Downsampling Engine $\implies$ **PASS**
14. RBAC Permission Checks & Audit Logging $\implies$ **PASS**
15. i18n Key Parity (1089 EN keys = 1089 ID keys) $\implies$ **PASS**

---

## 24. Real Sensor Validation

Hardware validation executed against live sensor telemetry (`AQMS-01` multi-sensor unit):
- **Raw Sensor Telemetry**: PM2.5 register reading `raw_value = 15.0 µg/m³`.
- **Phase 3.2 Processing**: Formula $(raw \times 1.0) + 0.0$ evaluated to `processed_value = 15.0 µg/m³` with `Quality = GOOD`.
- **Customer Aggregation**: 30-minute bucket `20261008140000` aggregated to Customer `AVG = 15.00 µg/m³`.
- **Customer API Delivery**: `GET /api/customer/aggregated-data/20261008140000` returned verified customer JSON with zero Modbus registers exposed.

---

## 25. Known Limitations

- High-frequency sampling ($> 100$ Hz) requires batch raw ingest partitioning before executing sub-minute rollups on single-core ARMv7 devices.
- Timezone changes on an existing definition do not retroactively rewrite historical period identifiers; new buckets follow the updated timezone.

---

## 26. Future Phase 8 Integration

Phase 3.3 prepares structured deliverables for Phase 8 (Data Distribution & External Integrations):
- Permanent period identifiers (`YYYYMMDDHHmmss`) serve as primary keys for external MQTT, Webhook, and REST delivery queues.
- `telemetry_aggregations` stores stable, pre-calculated payload snapshots eliminating recalculation overhead during external transmission retries.
