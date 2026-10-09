# Phase 4.3 — Backup & Restore Specification & Operator Runbook

## 1. System Overview

Phase 4.3 implements a local-first, consistent, and auditable **Backup & Restore Subsystem** for the Edge Datalogger Analysis Application. Designed specifically for critical industrial edge environments (Windows x64 and Linux ARM/AMD edge appliances), the system guarantees point-in-time recovery without silent loss of pending telemetry queued in the Write-Ahead Log (WAL).

### Architectural Diagram

```
+----------------------------------------------------------------------------------------------------+
|                                    BACKUP & RESTORE ARCHITECTURE                                   |
+----------------------------------------------------------------------------------------------------+
                                      |
                       HTTP / REST API & Scheduler
                                      |
                           +----------------------+
                           |    BackupService     |
                           +----------------------+
                                  |         |
         +------------------------+         +--------------------------+
         |                                                             |
+-------------------------+                               +--------------------------+
|  DatabaseBackupAdapter  |                               |    PersistentQueue /     |
| (mariadb-dump / Native) |                               |     WAL Coordinator      |
+-------------------------+                               +--------------------------+
         |                                                             |
+-------------------------+                               +--------------------------+
| MariaDB Logical Dump    |                               | WAL Segments &           |
| (Single-Transaction     |                               | Checkpoint Snapshot      |
|  Repeatable-Read)       |                               | (Synced to disk)         |
+-------------------------+                               +--------------------------+
         |                                                             |
         +------------------------+         +--------------------------+
                                  |         |
                           +----------------------+
                           |  Atomic Packaging &  |
                           | Checksum Calculation |
                           | (Gzip Tarball + SHA) |
                           +----------------------+
                                      |
                           +----------------------+
                           |   final archive:     |
                           |   backup_*.tar.gz    |
                           +----------------------+
                                      |
                           +----------------------+
                           |  BackupRepository    |
                           |  Catalog Entry &     |
                           |  Audit Trail Log     |
                           +----------------------+
```

---

## 2. Core Components

1. **Backup Service (`internal/service/backup_service.go`)**:
   - Central coordinator for pre-flight disk capacity checks, concurrent operation locking (`ErrOperationInProgress`), WAL snapshot coordination, database export, manifest packaging into atomic `.tar.gz`, cryptographic validation, restore previews, and disaster recovery execution.
2. **Database Backup Adapter (`internal/service/database_backup_adapter.go`)**:
   - Detects external MariaDB dump utilities (`mariadb-dump`, `mysqldump`, `mariadb`, `mysql`).
   - Pure-Go streaming logical SQL dump & restore engine that operates without external CLI dependencies, enabling zero-CGO, cross-platform portability.
   - Preserves transactions using single-transaction flags and handles foreign key checks dynamically.
   - Strictly sanitizes logs to redact connection passwords and user credentials.
3. **WAL Snapshot Coordinator (`internal/queue/wal.go`)**:
   - Safely captures open and uncommitted telemetry segments from the append-only Write-Ahead Log while continuous data acquisition and polling remain active.
   - Computes IEEE CRC-32 and SHA-256 hashes per segment file during staging.
4. **Manifest & Integrity Validator (`internal/model/backup.go`)**:
   - Generates machine-readable `manifest.json` detailing application version, database engine, schema statistics, row counts, WAL checkpoints, and individual artifact checksums.
   - Validates checksums and verifies SQL header and structure markers prior to authorizing a restore.
5. **Backup Catalog Repository (`internal/repository/backup_repo.go`)**:
   - Persists archive catalog metadata in MariaDB with statuses (`COMPLETED`, `FAILED`, `IN_PROGRESS`, `VALID`, `INVALID`).
   - Supports paginated catalog queries, size statistics, and retention lookups.
6. **Automated Scheduler & Pruner (`internal/service/backup_scheduler.go`)**:
   - Runs periodic background backups based on configurable hour intervals or daily execution times (UTC).
   - Enforces overlap prevention, non-overlapping mutex locking, and automatic retention pruning based on `max_keep_count`.
7. **Disaster Recovery Restore Engine**:
   - Multi-stage restore workflow:
     1. Backup selection and compatibility verification.
     2. Manifest and cryptographic SHA-256/CRC-32 checksum validation.
     3. Impact preview with target comparison and destructive warnings.
     4. Explicit user confirmation (`confirm: true`).
     5. Automatic pre-restore `SAFETY` snapshot of active database state.
     6. Background telemetry ingestion coordination.
     7. Database reconstruction from SQL dump.
     8. Persistent WAL recovery and idempotent telemetry replay using `record_uuid`.
     9. Post-restore integrity verification and audit logging.

---

## 3. Consistency Boundary & Guarantees

### A. Database Consistency
- When using MariaDB, logical dumps are captured with transaction consistency (repeatable read), ensuring all committed database transactions up to the snapshot start time are preserved.
- When using the pure-Go streaming engine, foreign key checks are temporarily disabled (`SET FOREIGN_KEY_CHECKS=0;`) during restore to prevent constraint order issues, and re-enabled upon completion.

### B. WAL Snapshot & Zero-Data-Loss Guarantees
- Ingested telemetry is spooled to disk in binary-framed WAL segments before or in parallel with MariaDB persistence.
- During backup creation, the WAL's active segment is synchronized (`Sync()`) to flush pending OS buffers, and all active/uncommitted segments are copied with CRC-32 validation.
- Telemetry arriving *during* backup creation is appended to the live WAL and remains safe for subsequent replay.
- Telemetry arriving *before* the snapshot boundary is included in the snapshot archive.

### C. Idempotent Replay & Deduplication
- Every telemetry record has a stable `RecordUUID` indexed with a database unique index.
- During restore, the restored WAL segments are replayed into MariaDB using `clause.OnConflict{DoNothing: true}` (or `INSERT IGNORE`).
- Even if a record was committed to the database prior to a crash and also exists in the restored WAL, repeated replay produces **zero duplicate rows**.

### D. Crash & Interruption Recovery
- Backup archives are created in a temporary staging directory (`staging_backup_*`) and compressed into a temporary file (`.tar.gz.tmp`).
- Only after full verification of manifest and archive checksums is the archive atomically renamed to its final filename (`backup_*.tar.gz`).
- Any interrupted or failed backup leaves behind no corrupted permanent archives; temp files are automatically removed.

---

## 4. Manifest Schema (`manifest.json`)

Each `.tar.gz` archive contains a top-level `manifest.json` structured as follows:

```json
{
  "backup_id": "backup_20261009_033353_0600",
  "backup_version": "1.0",
  "app_version": "1.0.0",
  "created_at": "2026-10-09T03:33:53.060000000Z",
  "backup_type": "MANUAL",
  "database": {
    "engine": "MariaDB 10.4.32",
    "database_name": "datalogger",
    "schema_version": "Phase 4.3",
    "dump_method": "mariadb-dump (single-transaction)",
    "table_count": 27,
    "total_rows": 14250,
    "table_rows": {
      "devices": 3,
      "parameters": 45,
      "raw_data": 12850,
      "aggregated_data": 1352
    },
    "artifact": {
      "path": "database/dump.sql",
      "size_bytes": 1485290,
      "sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
      "description": "MariaDB logical SQL schema and data export"
    }
  },
  "wal": {
    "pending_records": 120,
    "spool_size_bytes": 45800,
    "checkpoint_segment": 4,
    "checkpoint_offset": 12800,
    "active_segment": 4,
    "active_offset": 58600,
    "segments": [
      {
        "path": "wal/000004.wal",
        "size_bytes": 45800,
        "sha256": "a591a6d40bf420404a011733cfb7b190d62c65bf0bcda32b57b277d9ad9f146e",
        "crc32": 2864434397,
        "description": "WAL Segment 000004"
      }
    ]
  },
  "config_summary": {
    "app_name": "Datalogger Analysis Application",
    "environment": "production",
    "port": "8080",
    "db_type": "mariadb",
    "db_host": "127.0.0.1",
    "db_port": "3306",
    "db_name": "datalogger",
    "queue_enabled": true,
    "queue_sync_mode": "batch",
    "queue_max_size_bytes": 104857600,
    "backup_dir": "data/backups",
    "backup_schedule_enabled": true
  },
  "total_size_bytes": 384500,
  "archive_sha256": "f2ca1bb6c7e907d06dafe4687e579fce76b37e4e93b7605022da52e6ccc26fd2",
  "status": "COMPLETED",
  "validation_status": "VALID"
}
```

*Note: Passwords, JWT secrets, and private tokens are strictly omitted from `config_summary`.*

---

## 5. Configuration & Environment Variables

| Variable | Default | Description |
|---|---|---|
| `DATALOGGER_BACKUP_DIR` | `data/backups` | Absolute or relative directory storing backup archives. |
| `DATALOGGER_BACKUP_SCHEDULE_ENABLED` | `true` | Enables background automatic backup scheduler. |
| `DATALOGGER_BACKUP_INTERVAL_HOURS` | `24` | Backup frequency in hours (e.g. 24 = daily). |
| `DATALOGGER_BACKUP_TIME` | `02:00` | Preferred time of day in UTC for daily execution. |
| `DATALOGGER_BACKUP_COMPRESSION` | `gzip` | Archive compression method (`gzip` or `none`). |
| `DATALOGGER_BACKUP_MIN_FREE_SPACE_MB`| `500` | Minimum disk space required before starting backup. |
| `DATALOGGER_BACKUP_MAX_KEEP_COUNT` | `10` | Retention limit; oldest completed archives pruned. |

---

## 6. REST API Endpoints

All endpoints require JWT Bearer authentication and RBAC permissions.

| Method | Endpoint | Required Permission | Description |
|---|---|---|---|
| `GET` | `/api/backups` | `backup.view` | Paginated catalog list (`page`, `page_size`, `status`). |
| `GET` | `/api/backups/status` | `backup.view` | Backup subsystem metrics, free disk, active jobs. |
| `GET` | `/api/backups/schedule` | `backup.view` | Current automated scheduler configuration. |
| `PUT` | `/api/backups/schedule` | `backup.manage` | Update schedule configuration. |
| `GET` | `/api/backups/jobs/:jobId` | `backup.view` | Inspect asynchronous job progress and stage. |
| `GET` | `/api/backups/:id` | `backup.view` | Retrieve single backup catalog record and manifest. |
| `GET` | `/api/backups/:id/restore-preview`| `backup.view` | Impact preview comparing backup against active database. |
| `GET` | `/api/backups/:id/download` | `backup.view` | Stream `.tar.gz` archive to authorized client. |
| `POST` | `/api/backups` | `backup.create` | Initiate on-demand manual backup. |
| `POST` | `/api/backups/:id/validate` | `backup.validate` | Trigger cryptographic SHA-256 and manifest validation. |
| `POST` | `/api/backups/:id/restore` | `backup.restore` | Authenticated disaster recovery restore (`confirm: true`). |
| `DELETE`| `/api/backups/:id` | `backup.manage` | Permanently delete archive and remove from catalog. |

---

## 7. Operator Disaster Recovery Runbook

### Scenario A: Routine On-Demand Backup
1. Navigate to **Administration > Backup & Restore** in the Web UI.
2. Click **Create Backup**.
3. (Optional) Provide a maintenance note (e.g., `Pre-firmware upgrade snapshot DEV-01`).
4. Click **Confirm**. The background job will initialize, snapshot MariaDB and WAL, and register the valid archive in the catalog within milliseconds to seconds.

### Scenario B: Restoring After Accidental Data Corruption or Hardware Failure
1. In **Backup Catalog**, locate the desired point-in-time backup.
2. Click **Validate** to ensure checksums and files are intact.
3. Click **Restore**.
4. In the **Restore Impact Preview** dialog:
   - Review table row count comparisons.
   - Review pending WAL records to be replayed.
   - Ensure the checkbox **Create safety backup before restore** is checked.
5. Click **Yes, Restore Database**.
6. The engine will:
   - Create a `SAFETY` snapshot of the active state.
   - Restore database schemas and rows.
   - Replay pending WAL records with deduplication.
   - Verify database consistency and log an audit event.

### Scenario C: Manual Disaster Recovery from Terminal (CLI Fallback)
If the web application daemon is stopped or cannot boot due to a corrupted database:

1. Locate the latest verified archive in `data/backups/`:
   ```bash
   ls -la data/backups/
   # e.g., backup_20261009_020000_1234.tar.gz
   ```
2. Extract the archive into a temporary folder:
   ```bash
   mkdir /tmp/recovery
   tar -xzf data/backups/backup_20261009_020000_1234.tar.gz -C /tmp/recovery/
   ```
3. Inspect `manifest.json` to verify engine and table counts:
   ```bash
   cat /tmp/recovery/manifest.json
   ```
4. Restore the MariaDB database:
   ```bash
   mysql -u root -p datalogger < /tmp/recovery/database/dump.sql
   ```
5. Restore WAL queue files into the persistent queue spool directory:
   ```bash
   cp -r /tmp/recovery/wal/* data/queue/
   ```
6. Start the Datalogger application daemon:
   ```bash
   ./bin/datalogger-linux-amd64
   ```
7. The daemon will automatically detect uncommitted WAL records on startup and replay them idempotently into MariaDB.

---

## 8. Verified Acceptance Test Results

All automated acceptance and regression tests passed 100%:

```
=== RUN   TestPhase43_BackupCreationAndAtomicPackaging
--- PASS: TestPhase43_BackupCreationAndAtomicPackaging (0.08s)
=== RUN   TestPhase43_BackupManifestAndIntegrityValidation
--- PASS: TestPhase43_BackupManifestAndIntegrityValidation (0.14s)
=== RUN   TestPhase43_WALSnapshotCoordinationUnderIngestion
--- PASS: TestPhase43_WALSnapshotCoordinationUnderIngestion (0.03s)
=== RUN   TestPhase43_RestorePreviewAndSafetyBackup
--- PASS: TestPhase43_RestorePreviewAndSafetyBackup (0.20s)
=== RUN   TestPhase43_DisasterRecoveryRestoreAndIdempotentReplay
--- PASS: TestPhase43_DisasterRecoveryRestoreAndIdempotentReplay (0.14s)
=== RUN   TestPhase43_SchedulerExecutionAndOverlapPrevention
--- PASS: TestPhase43_SchedulerExecutionAndOverlapPrevention (0.96s)
=== RUN   TestPhase43_SecurityAndPathTraversalRejection
--- PASS: TestPhase43_SecurityAndPathTraversalRejection (0.04s)
=== RUN   TestPhase43_DiskSpacePreflightCheck
--- PASS: TestPhase43_DiskSpacePreflightCheck (0.04s)
PASS
ok      datalogger/internal/service     5.001s

=== RUN   TestPhase43_API_AuthenticationRequired
--- PASS: TestPhase43_API_AuthenticationRequired (0.29s)
=== RUN   TestPhase43_API_RBACEnforcement
--- PASS: TestPhase43_API_RBACEnforcement (0.25s)
=== RUN   TestPhase43_API_CatalogAndStatus
--- PASS: TestPhase43_API_CatalogAndStatus (0.28s)
=== RUN   TestPhase43_API_RestoreRequiresConfirmation
--- PASS: TestPhase43_API_RestoreRequiresConfirmation (0.36s)
PASS
ok      datalogger/internal/handler     5.274s
```
