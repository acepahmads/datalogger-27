# Phase 2.1 — Device Management: Technical Architecture & Implementation Documentation

**Project:** Datalogger Analysis Application  
**Phase:** Phase 2 — Device & Communication  
**Sub-Phase / Task:** Phase 2.1 — Device Management  
**Status:** **COMPLETED & VERIFIED (100%)**  
**Database:** MariaDB (Edge connection pooling, utf8mb4)  
**Backend:** Go (Gin, GORM, Zero CGO, REST, WebSocket)  
**Frontend:** Vue 2 (Vue 2.7, Vuex 3, Vue Router 3, Tailwind CSS)  
**Security:** JWT Authentication, Granular Role-Based Access Control (RBAC), Immutable Audit Trail  

---

## 1. Overview & Objectives

Phase 2.1 establishes the foundational data architecture, REST API services, configuration layer, and user interface for industrial equipment management. It is designed to scale across heterogeneous industrial edge deployments, including Modbus RTU, Modbus TCP, TCP/IP, UDP, HTTP/REST, MQTT, Serial RS232/RS485, WebSocket, and custom proprietary protocols.

In strict adherence to architectural directives:
- Communication protocol **configuration foundations** are established without running fake background telemetry.
- Administrative operational states (`ACTIVE`, `INACTIVE`, `MAINTENANCE`, `DISABLED`) are rigorously separated from dynamic communication link states (`ONLINE`, `OFFLINE`, `CONNECTING`, `ERROR`, `UNKNOWN`).
- Soft-delete semantics (`deleted_at`) ensure that historical acquisition readings and operational audit trails remain intact even when equipment is decommissioned.

---

## 2. Device Data Model

The device entity represents physical or logical equipment installed at an edge location.

```
+-----------------------------------------------------------------------------------+
|                                      devices                                      |
+-----------------------------------------------------------------------------------+
| id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY               |
| device_code             VARCHAR(64) UNIQUE NOT NULL (Indexed)                     |
| device_name             VARCHAR(128) NOT NULL                                     |
| device_type             VARCHAR(64) NOT NULL (e.g. MODBUS_TCP, MODBUS_RTU, MQTT)  |
| manufacturer            VARCHAR(128)                                              |
| model                   VARCHAR(128)                                              |
| serial_number           VARCHAR(128) (Indexed)                                    |
| firmware_version         VARCHAR(64)                                               |
| description             VARCHAR(255)                                              |
| location                VARCHAR(128)                                              |
| latitude                DOUBLE                                                    |
| longitude               DOUBLE                                                    |
| timezone                VARCHAR(64) DEFAULT 'UTC'                                 |
| status                  VARCHAR(32) DEFAULT 'ACTIVE' (Indexed)                    |
| enabled                 BOOLEAN DEFAULT TRUE (Indexed)                            |
| connection_status       VARCHAR(32) DEFAULT 'UNKNOWN'                             |
| last_seen_at            DATETIME(3)                                               |
| last_data_at            DATETIME(3)                                               |
| deleted_at              DATETIME(3) (Indexed - Soft Delete)                       |
| created_at              DATETIME(3)                                               |
| updated_at              DATETIME(3)                                               |
+-----------------------------------------------------------------------------------+
```

### Administrative vs. Communication Status

| Concept | States | Purpose |
| :--- | :--- | :--- |
| **Administrative Status** | `ACTIVE`, `INACTIVE`, `MAINTENANCE`, `DISABLED` | Set by system operators or maintenance schedules to govern whether data collection should occur. |
| **Communication Status** | `ONLINE`, `OFFLINE`, `CONNECTING`, `ERROR`, `UNKNOWN` | Set dynamically by protocol adapters and health supervisors based on live socket/serial link state. |

---

## 3. Device Connection Configuration Model

Each device maintains an extensible connection profile that specifies how communication adapters interface with hardware.

```
+-----------------------------------------------------------------------------------+
|                                device_connections                                 |
+-----------------------------------------------------------------------------------+
| id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY               |
| device_id               BIGINT UNSIGNED UNIQUE NOT NULL (Foreign Key -> devices)  |
| protocol                VARCHAR(32) NOT NULL                                      |
| connection_type         VARCHAR(64) DEFAULT 'ETHERNET'                            |
| host                    VARCHAR(255)                                              |
| port                    INT DEFAULT 502                                           |
| serial_port             VARCHAR(64)                                               |
| baud_rate               INT DEFAULT 9600                                          |
| data_bits               INT DEFAULT 8                                             |
| parity                  VARCHAR(16) DEFAULT 'N'                                   |
| stop_bits               INT DEFAULT 1                                             |
| timeout                 INT DEFAULT 1000 (ms)                                     |
| retry_count             INT DEFAULT 3                                             |
| polling_interval        INT DEFAULT 1000 (ms)                                     |
| enabled                 BOOLEAN DEFAULT TRUE                                      |
| extra_config            TEXT (JSON for MQTT topics, credentials, or custom)       |
| created_at              DATETIME(3)                                               |
| updated_at              DATETIME(3)                                               |
+-----------------------------------------------------------------------------------+
```

### Supported Protocols & Configuration Profiles

1. **MODBUS_TCP**: `host`, `port` (default 502), `timeout`, `retry_count`, `polling_interval`.
2. **MODBUS_RTU**: `serial_port` (e.g. `COM1` / `/dev/ttyUSB0`), `baud_rate` (9600..115200), `data_bits` (8), `parity` (`N`/`E`/`O`), `stop_bits` (1/2), `timeout`, `retry_count`, `polling_interval`.
3. **TCP**: `host`, `port`, `timeout`, `retry_count`, `polling_interval`.
4. **UDP**: `host`, `port`, `timeout`, `retry_count`.
5. **HTTP / REST**: `host` / URL, `port`, `timeout`, `retry_count`, `polling_interval`.
6. **MQTT**: Broker `host`, `port` (1883), `extra_config` (Client ID, Topic, Username, Password), `polling_interval`.
7. **SERIAL**: `serial_port`, `baud_rate`, `data_bits`, `parity`, `stop_bits`, `timeout`.
8. **WEBSOCKET**: `host`, `port`, `polling_interval`.
9. **CUSTOM**: Extensible parameters and proprietary payloads.

---

## 4. Device Parameters Architecture

Parameters define individual measurement registers and sensory channels associated with a parent device.

```
+-----------------------------------------------------------------------------------+
|                                    parameters                                     |
+-----------------------------------------------------------------------------------+
| id                      BIGINT UNSIGNED AUTO_INCREMENT PRIMARY KEY               |
| device_id               BIGINT UNSIGNED NOT NULL (Indexed)                        |
| parameter_code          VARCHAR(64) NOT NULL                                      |
| parameter_name          VARCHAR(128) NOT NULL                                     |
| data_type               VARCHAR(32) DEFAULT 'FLOAT32'                             |
| unit                    VARCHAR(32)                                               |
| description             VARCHAR(255)                                              |
| min_value               DOUBLE                                                    |
| max_value               DOUBLE                                                    |
| precision               INT DEFAULT 2                                             |
| scale                   DOUBLE DEFAULT 1.0                                        |
| offset                  DOUBLE DEFAULT 0.0                                        |
| register_address        INT                                                       |
| register_type           VARCHAR(32) (HOLDING_REGISTER, INPUT_REGISTER, COIL, ...)  |
| enabled                 BOOLEAN DEFAULT TRUE (Indexed)                            |
| deleted_at              DATETIME(3) (Indexed - Soft Delete)                       |
| created_at              DATETIME(3)                                               |
| updated_at              DATETIME(3)                                               |
+-----------------------------------------------------------------------------------+
UNIQUE INDEX: idx_device_param_code (device_id, parameter_code)
```

Constraint: A device cannot have duplicate parameter codes, but distinct devices may define the same parameter code (e.g., both Device 1 and Device 2 can have `TEMP_01`).

---

## 5. Device Health & Communication Foundation

The repository and service expose standard primitives for future protocol communication workers:

- `UpdateLastSeen(deviceID uint) error`: Updates the timestamp when the edge node acknowledged a ping or query.
- `UpdateLastData(deviceID uint) error`: Updates the timestamp when valid telemetry was parsed into engineering registers.
- `SetOnline(deviceID uint) error`: Marks communication status as `ONLINE`.
- `SetOffline(deviceID uint) error`: Marks communication status as `OFFLINE`.
- `SetConnectionError(deviceID uint, errMsg string) error`: Marks communication status as `ERROR`.

---

## 6. Security, Authorization & Audit Trail

### Role-Based Access Control (RBAC)

The system protects all device APIs behind JWT authentication and granular permission checks:

| Permission | Description | Allowed Roles |
| :--- | :--- | :--- |
| `device.view` | Query and inspect devices, connection settings, and telemetry readouts | Administrator, Engineer, Operator |
| `device.create` | Register new equipment profiles and connection configurations | Administrator, Engineer |
| `device.update` | Update equipment parameters, location, and communication parameters | Administrator, Engineer |
| `device.delete` | Soft delete equipment records | Administrator |
| `device.manage` | Add, update, enable/disable, and delete telemetry parameters | Administrator, Engineer |

### Immutable Audit Trail Events

All mutations record an audit trail event referencing user, action, resource, entity IDs, IP address, user agent, and before/after states:
- `CREATE_DEVICE`
- `UPDATE_DEVICE`
- `DELETE_DEVICE`
- `ENABLE_DEVICE`
- `DISABLE_DEVICE`
- `CHANGE_CONNECTION`
- `CREATE_PARAMETER`
- `UPDATE_PARAMETER`
- `DELETE_PARAMETER`

---

## 7. REST API Reference

| Method | Endpoint | Permission | Description |
| :--- | :--- | :--- | :--- |
| `GET` | `/api/devices` | `device.view` | List devices with search, filtering, pagination, and sorting |
| `GET` | `/api/devices/:id` | `device.view` | Get device profile, connection settings, and parameter list |
| `POST` | `/api/devices` | `device.create` | Register new device profile and connection settings |
| `PUT` | `/api/devices/:id` | `device.update` | Update existing device metadata and connection settings |
| `DELETE` | `/api/devices/:id` | `device.delete` | Soft delete device (preserves historical data) |
| `PUT` | `/api/devices/:id/enable` | `device.update` | Enable or disable device acquisition |
| `PUT` | `/api/devices/:id/status` | `device.update` | Update administrative status (`ACTIVE`, `MAINTENANCE`, etc.) |
| `PUT` | `/api/devices/:id/connection` | `device.update` | Reconfigure connection parameters |
| `GET` | `/api/devices/:id/activity` | `device.view` | Retrieve device-specific audit trail |
| `GET` | `/api/devices/:id/parameters` | `device.view` | List all telemetry parameters for device |
| `GET` | `/api/devices/:id/parameters/:paramId` | `device.view` | Get parameter details |
| `POST` | `/api/devices/:id/parameters` | `device.manage` | Create parameter under device |
| `PUT` | `/api/devices/:id/parameters/:paramId` | `device.manage` | Update parameter properties, scaling, or limits |
| `DELETE` | `/api/devices/:id/parameters/:paramId` | `device.manage` | Soft delete parameter |
| `PUT` | `/api/devices/:id/parameters/:paramId/enable` | `device.manage` | Toggle parameter enabled status |

---

## 8. Frontend User Interface

The Vue 2 frontend provides:
1. **Device List (`/monitoring/devices`)**:
   - High-density industrial table and responsive card grid.
   - Live KPI overview: Total, Online/Healthy, Offline/Error, Maintenance/Inactive.
   - Comprehensive search and filters (Admin Status, Comm Status, Protocol).
   - Multi-column sorting and pagination.
   - Quick action toggles (Enable/Disable, Edit, View Detail, Delete).
2. **Device Detail (`/monitoring/devices/:id`)**:
   - Header with status badges and quick action buttons.
   - 5 structured tabs:
     - **Overview**: Hardware specifications, physical location, coordinates, timestamps.
     - **Connection Config**: Protocol parameters, baud rates, host, port, timeout, retries.
     - **Parameters**: Register list with add, edit, toggle enable, and delete actions.
     - **Activity**: Live device-specific audit trail stream.
     - **Health Diagnostics**: Observed latency, packet counters, comm status (prepared for Phase 2.2/2.3 live adapter stream).
3. **Reusable Modals**:
   - `DeviceModal.vue`: Multi-tab form with dynamic protocol-specific fields (Modbus TCP, Modbus RTU, MQTT, HTTP, etc.).
   - `ParameterModal.vue`: Full parameter engineering form with data type, register offset, scaling factor, and boundary limits.
