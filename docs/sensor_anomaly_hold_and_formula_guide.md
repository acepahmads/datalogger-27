# Sensor Anomaly Debounce Protection & Formula Evaluation Guide

## 1. Overview & Problem Statement

In industrial edge IoT environments and field operations, sensors often encounter transient instability:
1. **Communication Glitches & Timeouts**: RS485 bus noise, temporary adapter stalls, or sensor firmware micro-reboots cause sudden read errors or timeouts.
2. **Weird / Glitch Values ("Nilai Aneh")**: Incomplete ADC conversions, electrical interference, or floating pins may yield abnormal values (e.g. `9999`, `-32768`, or values outside physical limits) for a few polling cycles before recovering.
3. **Complex Engineering Transformations**: Raw Modbus registers often require mathematical evaluation beyond simple linear scale/offset—for example, converting 4–20 mA current to percentage (`(raw - 4) * 6.25`), Celsius to Fahrenheit (`x * 1.8 + 32`), multi-parameter power conversions (`round(x / 1000, 2)`), or trigonometric/exponential scaling.

To solve these issues reliably, the system provides two native capabilities:
- **Anomaly Debounce Buffer (Hold Last Good Value)**: Holds the last known valid sensor reading when an anomaly occurs. If the sensor recovers within 2 minutes (120 seconds, configurable), no bad data is written to the database. Only if the anomaly persists continuously for more than 2 minutes does the system commit the real sensor state/error.
- **Custom Formula Evaluation Engine (`pkg/formula`)**: A secure, high-performance recursive-descent math expression evaluator that functions identically to Python's `eval()` for arithmetic and mathematical expressions.

---

## 2. Anomaly Debounce Protection Buffer

### 2.1 Mechanics & Flowchart

```
                 [Sensor Read Cycle]
                         │
        ┌────────────────┴────────────────┐
        ▼                                 ▼
   [Read Success]                   [Read Anomaly]
(Value within Min/Max)           (Timeout, Error, or Out-of-Bounds)
        │                                 │
        ▼                                 ▼
Reset Grace Timer               Is Feature Enabled?
Update LastGoodValue                      │
        │                        ┌────────┴────────┐
        ▼                        ▼ (No)            ▼ (Yes)
Commit Real Sensor Val      Commit Bad Val   Is Anomaly Timer Active?
Quality: GOOD               Quality: BAD           │
                                           ┌───────┴───────┐
                                           ▼ (No)          ▼ (Yes)
                                      Start Timer      Elapsed Time < 120s?
                                           │               │
                                           │       ┌───────┴───────┐
                                           │       ▼ (Yes)         ▼ (No)
                                           │   Hold Last Val   Grace Expired!
                                           │   Quality:        Commit Real Val
                                           │   UNCERTAIN       Quality: BAD
                                           │       │               │
                                           └───────┼───────────────┘
                                                   ▼
                                           Persist to Ingestion
```

### 2.2 Anomaly Conditions
A reading is marked anomalous when any of the following occur:
- Communication timeout or RS485 CRC error.
- Value is `NaN` or `Inf`.
- Value is strictly less than the parameter's configured `min_value`.
- Value is strictly greater than the parameter's configured `max_value`.

### 2.3 Grace Window Configuration
- **Toggle**: `hold_last_value_enabled` (boolean, per parameter).
- **Duration**: `hold_last_value_seconds` (integer, defaults to `120` seconds / 2 minutes).
- **State Transition**:
  - `0s <= elapsed < 120s`: The polling service substitutes the last valid value. Ingestion records `is_held_value = true` and `quality = UNCERTAIN`. The error is suppressed from breaking the data pipeline.
  - `elapsed >= 120s`: The grace window expires. The actual sensor error or out-of-bounds reading is committed with `is_held_value = false` and `quality = BAD`.
  - **Recovery**: As soon as the sensor produces a valid reading within bounds, the anomaly timer is cleared and normal `GOOD` quality data resumes immediately.

---

## 3. Mathematical Formula Evaluation Engine (`pkg/formula`)

### 3.1 Architecture
The formula engine is built from scratch in pure Go without external dependencies or dangerous runtime shells. It provides:
- **Lexical Tokenizer**: Scans numbers, operators, identifiers, and nested parentheses.
- **Recursive-Descent Parser**: Builds an AST supporting standard operator precedence and right-associative power operations.
- **Strict Evaluator**: Zero-allocation numeric operations with divide-by-zero safety checks.

### 3.2 Supported Syntax

| Feature | Syntax / Operators | Examples |
|---|---|---|
| **Basic Arithmetic** | `+`, `-`, `*`, `/`, `%` | `x * 10 + 5`, `x % 100` |
| **Exponentiation** | `^`, `**` | `x ^ 2`, `2 ** 8` |
| **Grouping** | `(`, `)` | `(raw - 4.0) * (100.0 / 16.0)` |
| **Variables** | `x`, `val`, `value`, `raw` | Both scaled value (`x`) and raw ADC/register (`raw`) |
| **Constants** | `pi`, `e` | `x * 2 * pi` |
| **Rounding** | `round(x, n)`, `floor(x)`, `ceil(x)` | `round(x, 2)`, `floor(x)` |
| **Math Functions** | `abs`, `sqrt`, `min`, `max`, `pow`, `log`, `exp` | `sqrt(abs(x))`, `max(x, 0)` |
| **Trigonometry** | `sin`, `cos`, `tan` | `sin(x * pi / 180)` |

### 3.3 Quick Presets
In the parameter configuration UI, engineers can click quick presets:
- **`°C ➔ °F`**: `x * 1.8 + 32`
- **`4-20mA ➔ %`**: `(raw - 4.0) * (100.0 / 16.0)`
- **`Round (2 Decimals)`**: `round(x, 2)`
- **`W ➔ kW`**: `x / 1000`

---

## 4. Database Schema Changes

### 4.1 `parameters` Table
```sql
ALTER TABLE parameters
  ADD COLUMN formula VARCHAR(255) NULL AFTER unit,
  ADD COLUMN hold_last_value_enabled BOOLEAN NOT NULL DEFAULT FALSE AFTER formula,
  ADD COLUMN hold_last_value_seconds INT NOT NULL DEFAULT 120 AFTER hold_last_value_enabled,
  ADD COLUMN current_formula_value DOUBLE NULL AFTER current_value,
  ADD COLUMN is_current_held BOOLEAN NOT NULL DEFAULT FALSE AFTER current_quality;
```

### 4.2 `raw_data` Table
```sql
ALTER TABLE raw_data
  ADD COLUMN formula_value DOUBLE NULL AFTER value_text,
  ADD COLUMN is_held_value BOOLEAN NOT NULL DEFAULT FALSE AFTER quality;
```

---

## 5. REST API Endpoints

### 5.1 Formula Validation
- **Method**: `POST`
- **Path**: `/api/devices/parameters/validate-formula` (or `/api/devices/:id/parameters/validate-formula`)
- **Request Body**:
  ```json
  {
    "formula": "x * 1.8 + 32",
    "sample_x": 25.0,
    "sample_raw": 25.0
  }
  ```
- **Response**:
  ```json
  {
    "success": true,
    "message": "Formula valid",
    "data": {
      "valid": true,
      "formula": "x * 1.8 + 32",
      "result": 77.0
    }
  }
  ```

### 5.2 Parameter Configuration
Parameters can be created or updated with formula and hold fields:
```json
{
  "parameter_code": "TEMP_CABINET",
  "parameter_name": "Cabinet Temperature",
  "data_type": "FLOAT32",
  "register_type": "HOLDING_REGISTER",
  "register_address": 40001,
  "scale": 0.1,
  "offset": 0.0,
  "unit": "°F",
  "min_value": -40.0,
  "max_value": 120.0,
  "formula": "x * 1.8 + 32",
  "hold_last_value_enabled": true,
  "hold_last_value_seconds": 120
}
```

---

## 6. User Interface & Monitoring

1. **Parameter Modal**:
   - Integrated formula input with syntax help.
   - Real-time client-side preview computing outputs as you type with test inputs.
   - Quick preset buttons for common industrial conversions.
   - Hold Last Good Value switch with configurable hold duration (seconds).
2. **Device Detail View**:
   - Parameters table shows configured formulas and active hold shield badges (`🛡️ 120s`).
   - Live telemetry cards display primary reading and `ƒ(x)` calculated formula values.
   - Animated pulsing `[DITAHAN / HELD]` badge alerts operators when data is currently being protected by the grace window.
3. **Telemetry Monitor View**:
   - Card grid view shows live telemetry with `ƒ(x)` formula summary banner and held status indicators.
   - Table view features a dedicated **Formula** column and clear held status tags.
   - Search filter supports searching parameters by formula expression.
