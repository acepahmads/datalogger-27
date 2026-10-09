# Phase 4.4 — Retention & Storage Management Specification & Operator Runbook

## 1. System Overview

Phase 4.4 delivers a robust, safe, and fully auditable **Retention & Storage Management Subsystem** for the Edge Datalogger Analysis Application. Operating on industrial edge gateways (Windows x64 and Linux ARM/AMD devices), this subsystem guarantees long-term storage sustainability while rigorously preventing data loss or accidental deletion of unbacked, unaggregated, or operational records.

### Architectural Diagram

```
+----------------------------------------------------------------------------------------------------+
|                             RETENTION & STORAGE MANAGEMENT ARCHITECTURE                            |
+----------------------------------------------------------------------------------------------------+
                                               |
                                HTTP / REST API & Housekeeping Worker
                                               |
                                    +--------------------+
                                    |  RetentionService  |
                                    +--------------------+
                                               |
         +--------------------+----------------+--------------------+--------------------+
         |                    |                                     |                    |
+-----------------+  +-----------------+                   +-----------------+  +-----------------+
|  Policy Engine  |  | Dry-Run Engine  |                   |  Backup Gateway |  | Rollup Gateway  |
| & Safety Checks |  |  (Side-effect   |                   |  (Coverage Gate |  | (Aggregate Gate |
| (Disabled Default) |      Free)      |                   |  via Phase 4.3) |  | via Phase 3.3) |
+-----------------+  +-----------------+                   +-----------------+  +-----------------+
         |                    |                                     |                    |
         +--------------------+----------------+--------------------+--------------------+
                                               |
                                    +--------------------+
                                    |  Batch Deletion    |
                                    |     Executor       |
                                    | (Bounded & Yields) |
                                    +--------------------+
                                               |
                  +----------------------------+----------------------------+
                  |                                                         |
         +------------------+                                      +------------------+
         | MariaDB Tables   |                                      | Disk & Spool     |
         | (raw_data,       |                                      | Monitor          |
         |  aggregation_*,  |                                      | (information_    |
         |  alarms, logs,   |                                      |  schema, WAL,    |
         |  audit_trails)   |                                      |  gopsutil/disk)  |
         +------------------+                                      +------------------+
```

---

## 2. Critical Safety Invariants

The subsystem strictly enforces 10 non-negotiable safety rules:

1. **Master/Configuration Protection**: Operational tables (`devices`, `parameters`, `system_configs`, `roles`, `users`) are strictly excluded from retention cleanup.
2. **Alarm Integrity**: Active or unacknowledged alarms (`status != 'CLEARED'`) are never pruned. Only historical alarms marked `CLEARED` that exceed the retention boundary are candidate for deletion.
3. **Internal vs Customer Isolation**: Internal raw telemetry aggregations (`source_type = 'INTERNAL_RAW'`) and customer processed aggregations (`source_type = 'CUSTOMER_PROCESSED'`) are managed by separate policies. Cleanup on internal aggregates never touches customer aggregations.
4. **Customer Data Preservation**: Customer processed aggregations cannot be deleted by raw telemetry cleanup logic.
5. **WAL Protection**: Persistent WAL queue files (`data/queue/*.wal`) and checkpoints (`checkpoint.json`) are never touched or deleted by retention cleanup.
6. **Disabled by Default**: Newly installed or migrated retention policies are seeded with `enabled = false`. Destructive actions require deliberate administrator activation and explicit execution authorization.
7. **No Raw Deletion Merely on Aggregate Existence**: Raw telemetry is preserved through its own dedicated retention period regardless of whether aggregations have already been computed.
8. **Audit Trail Longevity**: Audit records (`audit_trails`) default to a long-term retention period (365 days) and require verified backup coverage before cleanup.
9. **Backup Protection Gate**: When `require_backup = true` is configured for a policy, cleanup is unconditionally blocked if the target records are not covered by a verified, completed backup archive whose snapshot boundary spans past the deletion cutoff.
10. **Rollup Protection Gate**: When `require_rollup = true` is configured, raw telemetry cleanup is unconditionally blocked if aggregation jobs have not successfully computed rollups up to the deletion cutoff timestamp.

---

## 3. Data Categories & Default Baseline Policies

| Policy ID | Category | Target Table / Scope | Default Days | Min Age (Hrs) | Protected Period | Require Backup | Require Rollup | Batch Size | Default State |
|:---|:---|:---|:---:|:---:|:---:|:---:|:---:|:---:|:---:|
| `pol-raw-telemetry` | `RAW_TELEMETRY` | `raw_data` | 30 | 24 | 0 | Yes | Yes | 500 | Disabled |
| `pol-internal-agg` | `INTERNAL_AGGREGATIONS` | `aggregation_results` (`source_type='INTERNAL_RAW'`) | 90 | 24 | 0 | Yes | No | 500 | Disabled |
| `pol-customer-agg` | `CUSTOMER_AGGREGATIONS` | `aggregation_results` (`source_type='CUSTOMER_PROCESSED'`) | 365 | 24 | 0 | Yes | No | 500 | Disabled |
| `pol-cleared-alarms`| `CLEARED_ALARMS` | `alarms` (`status='CLEARED'`) | 180 | 24 | 0 | No | No | 200 | Disabled |
| `pol-system-logs` | `SYSTEM_LOGS` | `system_logs` | 60 | 24 | 0 | No | No | 500 | Disabled |
| `pol-comm-logs` | `COMMUNICATION_LOGS` | `communication_logs` | 30 | 24 | 0 | No | No | 500 | Disabled |
| `pol-audit-trails` | `AUDIT_TRAILS` | `audit_trails` | 365 | 24 | 0 | Yes | No | 200 | Disabled |
| `pol-backup-catalog`| `BACKUP_ARCHIVES` | `backup_records` (Pruned via BackupService) | 90 | 24 | 0 | No | No | 50 | Disabled |

---

## 4. Subsystem Components

### A. Retention Policy Engine (`internal/service/retention_service.go`)
- Validates all policy adjustments:
  - `retention_days` must be >= 1.
  - `min_age_hours` must be >= 1.
  - `batch_size` must be bounded between 10 and 5,000.
  - Category must match an authorized `DataCategory`.
- Manages policy configuration updates and operational toggling (`EnablePolicy`).

### B. Dry-Run Engine
- Computes pre-execution previews without any database modifications.
- Performs:
  - Calculation of exact cutoff timestamp: `cutoff = now - (RetentionDays + ProtectedPeriodDays)`.
  - Application of safety boundary: `cutoff <= now - MinAgeHours`.
  - Query of candidate records using indexed date ranges.
  - Evaluation of backup coverage status and snapshot boundaries.
  - Evaluation of rollup status (verifies whether aggregations are caught up).
  - Estimation of reclaimable table storage (using average row lengths from `information_schema.tables`).
  - Reporting of dependencies and explicit blocking reasons.

### C. Bounded Batch Deletion Engine
- Executes deletions in small, configurable batches (chunk size 10 to 5,000).
- Utilizes indexed primary keys (`id`) to prevent table locks and long-running lock escalations.
- Enforces a 5ms sleep between batches to yield edge CPU and disk I/O to real-time Modbus ingestion.
- Idempotent and resumable: if interrupted, the next cycle resumes from remaining records.

### D. Backup-Aware Protection Gate
- Interrogates `BackupService.GetLatestCompletedBackup()`.
- If `require_backup` is enabled:
  - Checks if a valid completed backup exists.
  - Verifies that the backup snapshot timestamp is at or after the deletion cutoff timestamp.
  - If no backup covers the cutoff, execution is blocked with status `BLOCKED` and audit logged.

### E. Rollup-Aware Downsampling Gate
- If `require_rollup` is enabled:
  - Queries `aggregation_results` for the most recent aggregate record.
  - Verifies that downsampling has processed up to or beyond the deletion cutoff.
  - If rollup lags behind, raw deletion is blocked to prevent calculation holes.

### F. Storage Monitoring Subsystem
- Interrogates `information_schema.tables` for live MariaDB table sizes (`data_length`, `index_length`).
- Queries `TelemetryService` for active Write-Ahead Log queue spool size and pending files.
- Queries `BackupService` for total backup storage utilization and archive count.
- Interrogates host filesystem using `gopsutil/v3/disk` to report total, used, and free disk space.
- Detects shared filesystems across application paths to prevent double-counting capacity.
- Evaluates warning (80%) and critical (90%) storage thresholds.

### G. Housekeeping Scheduler (`internal/service/retention_scheduler.go`)
- Background worker running on a configurable interval (default: every 6 hours).
- Enforces an execution mutex to strictly prevent overlapping cycles.
- Gracefully handles application shutdown and recovers safely on startup.

---

## 5. REST API Reference

All retention endpoints require JWT authentication and specific RBAC permissions.

### Endpoints

| Method | Endpoint | Permission | Description |
|:---|:---|:---|:---|
| `GET` | `/api/retention/storage` | `retention.view` | Retrieves storage overview, table sizes, spool usage, and disk thresholds |
| `GET` | `/api/retention/policies` | `retention.view` | Lists all retention policies and their execution status |
| `GET` | `/api/retention/policies/:id` | `retention.view` | Retrieves configuration and execution history for a single policy |
| `PUT` | `/api/retention/policies/:id` | `retention.manage` | Updates policy retention parameters and safety requirements |
| `POST`| `/api/retention/policies/:id/toggle` | `retention.manage` | Enables or disables a retention policy |
| `POST`| `/api/retention/policies/:id/dry-run` | `retention.view` | Runs a side-effect-free dry-run simulation |
| `POST`| `/api/retention/policies/:id/execute` | `retention.execute` | Triggers immediate bounded cleanup (requires `{"confirm": true}`) |
| `GET` | `/api/retention/history` | `retention.view` | Retrieves paginated execution and cleanup audit logs |

---

## 6. Vue 2 Administration Interface

Located at `/administration/retention` in the web application:

1. **Storage Health & Metrics Header**:
   - Live disk capacity bar with warning (80%) and critical (90%) indicators.
   - MariaDB storage usage, active WAL spool size, and backup storage breakdown.
   - Housekeeping status badge (Idle / Running / Disabled).
2. **Retention Policy Management**:
   - Tabular view of all 8 policy categories.
   - Color-coded category tags and enablement switches.
   - Safety requirement badges (Backup Required, Rollup Required).
   - Policy configuration modal with input validation.
3. **Dry-Run Preview Modal**:
   - Pre-flight simulation display with candidate record counts and estimated storage recovery.
   - Safety checks indicator:
     - Backup coverage verification status.
     - Rollup aggregation readiness status.
   - Explicit blocking reasons display if safety requirements fail.
4. **Destructive Execution Confirmation**:
   - Requires explicit checkbox confirmation (`confirm: true`) and confirmation dialog.
   - Real-time progress feedback and audit trail creation.
5. **Database Table Breakdown**:
   - Per-table breakdown of data size, index size, total size, and estimated row count.
6. **Execution History Log**:
   - Paginated historical log displaying trigger types (SCHEDULED / MANUAL), records evaluated, records deleted, duration, status (COMPLETED / BLOCKED / FAILED), and blocking reasons.

---

## 7. Testing & Verification

Automated test suites verify all safety constraints and operational requirements:

- `internal/service/retention_test.go`:
  - `TestRetentionPolicyValidation`: Validates retention days, minimum age, and batch size limits.
  - `TestDefaultPoliciesDisabled`: Confirms newly seeded policies default to `enabled = false`.
  - `TestDryRunSideEffects`: Proves dry-run preview generates candidate counts without altering data.
  - `TestCalculateCutoff`: Verifies cutoff boundary formulas across various retention and protected day configurations.
  - `TestAlarmRetentionPreservesActive`: Proves active and unacknowledged alarms are preserved while cleared alarms are pruned.
  - `TestInternalVsCustomerAggregationSeparation`: Proves customer aggregations are untouched by internal aggregation cleanup.
  - `TestBackupCoverageGate`: Confirms deletion is blocked when required backup coverage is absent.
  - `TestRollupCoverageGate`: Confirms raw telemetry deletion is blocked when aggregate calculation lags behind.
  - `TestStorageOverviewCapacity`: Confirms disk and table storage calculation logic.
- `internal/handler/retention_handler_test.go`:
  - Validates unauthenticated request rejection (401).
  - Validates operator role rejection for policy management and execution (403).
  - Validates engineer view permissions.
  - Validates administrator execution with required `confirm: true` parameter.
  - Validates dry-run simulation and storage overview endpoints.
- `internal/database/migration_test.go`:
  - Verifies Subphase 4.4 migration, 11 subtask registrations, legacy task superseding, and retention permissions seeding.
