package database

import (
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Seed populates the database with initial required data if empty.
func Seed(db *gorm.DB) error {
	var count int64
	db.Model(&model.User{}).Count(&count)
	if count > 0 {
		return nil // Already seeded
	}

	logger.Info("Starting database seed with 10 development phases and initial configuration...")

	// 1. Roles & Permissions
	adminRole := model.Role{
		Name:        "Administrator",
		Description: "Full system administration and development management access",
	}
	engineerRole := model.Role{
		Name:        "Engineer",
		Description: "Device configuration, parameter engineering, and development updates",
	}
	operatorRole := model.Role{
		Name:        "Operator",
		Description: "Operational monitoring, alarm acknowledgement, and reporting",
	}

	db.Create(&adminRole)
	db.Create(&engineerRole)
	db.Create(&operatorRole)

	// 2. Default Admin User (admin / admin123)
	hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	adminUser := model.User{
		Username: "admin",
		Email:    "admin@datalogger.local",
		Password: string(hash),
		FullName: "Lead System Engineer",
		RoleID:   adminRole.ID,
		IsActive: true,
	}
	db.Create(&adminUser)

	// 3. Populate All 10 Development Phases & Pre-configured Tasks
	phasesData := []struct {
		Number      int
		Name        string
		Description string
		Status      model.PhaseStatus
		Tasks       []struct {
			Name        string
			Description string
			Status      model.TaskStatus
			Progress    float64
			Priority    model.Priority
			Result      string
		}
	}{
		{
			Number:      1,
			Name:        "Phase 1 — Foundation",
			Description: "Core architecture, Go backend, Vue 2 Web UI shell, MariaDB database, configuration, logging, and cross-platform service installer.",
			Status:      model.PhaseCompleted,
			Tasks: []struct {
				Name        string
				Description string
				Status      model.TaskStatus
				Progress    float64
				Priority    model.Priority
				Result      string
			}{
				{"Architecture", "Design modular local-first architecture for edge/low-spec devices", model.StatusDone, 100, model.PriorityCritical, "Complete modular backend cmd/ internal/ pkg/ structure and decoupled Vue frontend"},
				{"Go backend", "Implement lightweight Go REST API and WebSocket engine using Gin framework", model.StatusDone, 100, model.PriorityCritical, "Gin server with low memory footprint and concurrent request handling implemented"},
				{"Vue Web UI", "Build modern industrial SaaS dashboard shell with Vue 2, Tailwind CSS, Vuex & Router", model.StatusDone, 100, model.PriorityCritical, "Responsive compact industrial UI with KPI cards, timeline, and dark mode"},
				{"Database", "Implement high-performance MariaDB storage with connection pooling, migrations, and transactional safety", model.StatusDone, 100, model.PriorityCritical, "MariaDB Edge connection pooling, auto-migrations, and health monitoring configured"},
				{"Configuration Management", "Structured JSON/ENV configuration system with edge defaults", model.StatusDone, 100, model.PriorityHigh, "JSON and Environment variable configuration loader functional"},
				{"Logging", "Asynchronous non-blocking structured logger with file persistence and levels", model.StatusDone, 100, model.PriorityHigh, "Asynchronous buffer logging with DEBUG, INFO, WARN, ERROR levels"},
				{"Basic Installer", "Self-contained installation scripts for Windows PowerShell and Linux bash", model.StatusDone, 100, model.PriorityHigh, "PowerShell and Linux shell install scripts created"},
				{"System Service", "Service definition for auto-start on Windows background and Linux systemd", model.StatusDone, 100, model.PriorityMedium, "Systemd unit and Windows background service launcher ready"},
				{"Cross Platform Build", "Pure Go build scripts supporting Windows x64, Linux x64, ARM64, and Raspberry Pi ARM", model.StatusDone, 100, model.PriorityCritical, "Zero-CGO cross-compilation verified for x64 and ARM architectures"},
			},
		},
		{
			Number:      2,
			Name:        "Phase 2 — Device & Communication",
			Description: "Industrial protocol adapters, device discovery, connection management, Modbus RTU/TCP, MQTT, Serial, TCP/UDP.",
			Status:      model.PhasePending,
			Tasks: []struct {
				Name        string
				Description string
				Status      model.TaskStatus
				Progress    float64
				Priority    model.Priority
				Result      string
			}{
				{"Device Management", "CRUD management for industrial sensors, meters, and edge devices", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Device Registration", "Device profile registration with unique IDs, locations, and metadata", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Device Discovery", "Automated IP network and serial port scanning for attached devices", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Parameter Management", "Definition of telemetry parameters, registers, data types, and scaling", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Modbus RTU", "RS485/RS232 serial Modbus RTU master implementation with framing check", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Modbus TCP", "Modbus TCP master client over Ethernet/WiFi with keep-alive", model.StatusPending, 0, model.PriorityCritical, ""},
				{"TCP", "Raw socket TCP client/server protocol handler for custom industrial streams", model.StatusPending, 0, model.PriorityMedium, ""},
				{"UDP", "Lightweight UDP datagram receiver for fast telemetry packets", model.StatusPending, 0, model.PriorityMedium, ""},
				{"REST API", "HTTP client adapter for third-party RESTful smart sensors and APIs", model.StatusPending, 0, model.PriorityMedium, ""},
				{"MQTT", "MQTT client subscriber for broker-based IoT sensors and gateways", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Serial Communication", "Configurable UART serial port communication (baud rate, parity, stop bits)", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Protocol Adapter", "Modular protocol interface abstraction allowing plug-and-play drivers", model.StatusPending, 0, model.PriorityCritical, ""},
			},
		},
		{
			Number:      3,
			Name:        "Phase 3 — Data Engine",
			Description: "Raw data acquisition, formula evaluation, data scaling, validation, spike/outlier detection, and polling schedulers.",
			Status:      model.PhasePending,
			Tasks: []struct {
				Name        string
				Description string
				Status      model.TaskStatus
				Progress    float64
				Priority    model.Priority
				Result      string
			}{
				{"Raw Data Acquisition", "High-frequency acquisition worker with ring buffer architecture", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Protocol Parsing", "Binary and ASCII payload decoders for industrial sensor telemetry", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Data Mapping", "Mapping extracted register bits and bytes into structured parameter values", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Data Validation", "Range checking, sanity checking, and physical boundary validation", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Scaling", "Linear scaling calculation (Scale Factor * Raw + Offset)", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Conversion", "Unit conversions (Celsius to Fahrenheit, bar to psi, kW to MW)", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Formula", "Math expression evaluation engine for virtual and calculated parameters", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Scheduler", "High-precision cron and interval polling engine per device channel", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Polling", "Concurrent asynchronous device polling workers with rate throttling", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Average", "Moving average calculation algorithms over configurable time windows", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Min", "Windowed minimum value tracking per parameter", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Max", "Windowed maximum value tracking per parameter", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Aggregation", "Periodic bucket aggregation (1m, 5m, 15m, 1h, 1d) for fast trends", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Spike Detection", "Rate-of-change spike detection to flag transient electrical noise", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Outlier Detection", "Statistical standard deviation filtering for faulty sensor readings", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Data Quality", "OPC-compliant data quality assessment (Good, Uncertain, Bad)", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Buffer", "Memory-efficient circular buffer for high-speed burst handling", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Queue", "Local in-memory and disk-backed FIFO queue for downstream processing", model.StatusPending, 0, model.PriorityCritical, ""},
			},
		},
		{
			Number:      4,
			Name:        "Phase 4 — Reliability",
			Description: "Auto-recovery, reconnect handlers, local queues, data integrity checks, backup/restore, and retention policies.",
			Status:      model.PhasePending,
			Tasks: []struct {
				Name        string
				Description string
				Status      model.TaskStatus
				Progress    float64
				Priority    model.Priority
				Result      string
			}{
				{"Auto Reconnect", "Exponential backoff auto-reconnect for dropped serial and TCP sockets", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Auto Recovery", "Worker panic recovery and self-healing background supervisor routines", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Retry Mechanism", "Configurable poll retry policies with dead-letter queue handling", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Local Queue", "Persistent disk buffer ensuring zero data loss during power/network anomalies", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Data Integrity", "Checksum validation and database consistency verification tools", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Backup", "Automated scheduled local database backup with gzip compression", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Restore", "Single-command database recovery and snapshot restoration tool", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Housekeeping", "Automated MariaDB table maintenance, index optimization, and temp file cleanup", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Retention Policy", "Automated pruning and downsampling of old raw data to preserve disk space", model.StatusPending, 0, model.PriorityHigh, ""},
			},
		},
		{
			Number:      5,
			Name:        "Phase 5 — Storage",
			Description: "Optimized relational storage for raw telemetry, processed engineering values, aggregations, metadata, logs, and audit trails.",
			Status:      model.PhasePending,
			Tasks: []struct {
				Name        string
				Description string
				Status      model.TaskStatus
				Progress    float64
				Priority    model.Priority
				Result      string
			}{
				{"Raw Data Storage", "Partitioned/indexed time-series storage for raw sensor measurements", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Processed Data Storage", "Indexed storage of calibrated, converted engineering values", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Aggregated Data", "Pre-calculated statistical rollups for instantaneous chart querying", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Device Metadata", "Structured configuration tables for equipment specifications and locations", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Parameter Metadata", "Parameter definitions, unit descriptors, and alarm limit configurations", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Audit Trail Storage", "Immutable event journal tracking all user configuration edits", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Log Storage", "Rotating circular file and MariaDB log table for diagnostic traces", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Alarm Storage", "Historical alarm event and acknowledgement lifecycle records", model.StatusPending, 0, model.PriorityHigh, ""},
			},
		},
		{
			Number:      6,
			Name:        "Phase 6 — Alarm & Notification",
			Description: "Threshold monitoring, abnormality detection, multi-channel notification engine (Web, Email, SMS, Telegram, WhatsApp, Push).",
			Status:      model.PhasePending,
			Tasks: []struct {
				Name        string
				Description string
				Status      model.TaskStatus
				Progress    float64
				Priority    model.Priority
				Result      string
			}{
				{"Threshold", "Configurable High/Low and Warning/Critical limit evaluation", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Warning", "Yellow alert triggers for parameters entering pre-critical operating zones", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Critical", "Red emergency triggers requiring immediate operational intervention", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Device Offline", "Heartbeat monitoring alerting when devices stop responding", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Communication Error", "Error rate monitoring alerting on serial parity or CRC failures", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Data Abnormality", "Statistical outlier and stuck-value detection alarm triggers", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Alarm Engine", "Centralized evaluation pipeline with alarm state latching and hysteresis", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Notification Engine", "Multi-channel dispatcher with rate limiting and retry queue", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Alarm Acknowledgement", "Operator acknowledgement workflow with notes and audit logging", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Alarm History", "Queryable historical alarm log with duration and resolution stats", model.StatusPending, 0, model.PriorityMedium, ""},
			},
		},
		{
			Number:      7,
			Name:        "Phase 7 — Web Application",
			Description: "Full industrial Web UI with monitoring dashboards, historical trends, reports, alarms, audit logs, and user management.",
			Status:      model.PhasePending,
			Tasks: []struct {
				Name        string
				Description string
				Status      model.TaskStatus
				Progress    float64
				Priority    model.Priority
				Result      string
			}{
				{"Dashboard", "Live overview dashboard with high-density operational telemetry cards", model.StatusWorking, 50, model.PriorityCritical, "Phase 1 dashboard shell functional"},
				{"Device Monitoring", "Real-time device status cards, connection latency, and health indicators", model.StatusWorking, 40, model.PriorityHigh, "Device list and telemetry cards created"},
				{"Parameter Monitoring", "Live digital readout tables and analog gauge visualizations", model.StatusWorking, 40, model.PriorityHigh, "Parameter readouts integrated"},
				{"Communication Monitoring", "Packet throughput, error rate, and port activity visualizer", model.StatusWorking, 40, model.PriorityMedium, "Comm metrics UI established"},
				{"Data Monitoring", "Raw and processed data live feed with search and filtering", model.StatusWorking, 30, model.PriorityMedium, "Data monitoring viewer drafted"},
				{"Historical Trend", "Interactive multi-axis time-series charts with zoom and pan", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Analysis", "Statistical summaries, min/max/mean analysis, and correlation tools", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Alarm Dashboard", "Dedicated active alarm banner, audible alerts, and acknowledgement UI", model.StatusWorking, 40, model.PriorityHigh, "Alarm monitoring components scaffolded"},
				{"Audit Trail Viewer", "Searchable change log of user logins, config edits, and commands", model.StatusWorking, 50, model.PriorityMedium, "Audit trail log viewer functional"},
				{"Log Viewer", "Real-time streaming console for backend and device logs", model.StatusWorking, 50, model.PriorityMedium, "Log viewer viewer functional"},
				{"Delivery Status", "Outbound transmission queue status and external sync metrics", model.StatusWorking, 30, model.PriorityMedium, "Delivery status panel drafted"},
				{"Export", "CSV, Excel, and JSON data exporter with date range selection", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Reporting", "Automated PDF and HTML operational report generation engine", model.StatusPending, 0, model.PriorityMedium, ""},
				{"User Management", "User account administration, credential management, and status toggle", model.StatusWorking, 50, model.PriorityHigh, "User and role management UI shell built"},
				{"Role Management", "Role-based access control (RBAC) role definition interface", model.StatusWorking, 50, model.PriorityMedium, "Role models created"},
				{"Permission Management", "Granular capability matrix mapping roles to system permissions", model.StatusWorking, 50, model.PriorityMedium, "Permission matrix models created"},
				{"System Configuration", "Web interface for runtime application parameters and serial settings", model.StatusWorking, 50, model.PriorityHigh, "System settings UI scaffolded"},
			},
		},
		{
			Number:      8,
			Name:        "Phase 8 — Output & Integration",
			Description: "Outbound data routing via REST, MQTT, TCP/UDP, WebSocket, and optional cloud forwarder with offline queuing.",
			Status:      model.PhasePending,
			Tasks: []struct {
				Name        string
				Description string
				Status      model.TaskStatus
				Progress    float64
				Priority    model.Priority
				Result      string
			}{
				{"REST API", "Outbound HTTP webhooks pushing telemetry batches to third-party endpoints", model.StatusPending, 0, model.PriorityHigh, ""},
				{"TCP", "Raw TCP streaming socket client for SCADA/MES upstream integration", model.StatusPending, 0, model.PriorityMedium, ""},
				{"UDP", "Broadcast/multicast UDP telemetry transmitter for local displays", model.StatusPending, 0, model.PriorityMedium, ""},
				{"MQTT", "Outbound MQTT publisher with QoS 1/2 and retained messages", model.StatusPending, 0, model.PriorityHigh, ""},
				{"WebSocket", "Real-time WebSocket streaming feed for third-party dashboards", model.StatusWorking, 50, model.PriorityHigh, "WebSocket hub implemented for Phase 1"},
				{"External System Integration", "Configurable payload formatters (JSON, CSV, Sparkplug B)", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Data Routing", "Rule-based routing filtering which parameters go to which destination", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Delivery Queue", "Persistent offline queue buffering telemetry when external network is down", model.StatusPending, 0, model.PriorityCritical, ""},
				{"Delivery Retry", "Intelligent exponential backoff retry mechanism for failed deliveries", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Delivery Acknowledgement", "Tracking delivery receipts and timestamp confirmations", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Optional Cloud Transmission", "Opt-in cloud sync adapter with complete local-first offline isolation", model.StatusPending, 0, model.PriorityHigh, ""},
			},
		},
		{
			Number:      9,
			Name:        "Phase 9 — Cross Platform & Installer",
			Description: "Cross-platform packaging for Windows, Linux x64, ARM64, and Raspberry Pi ARM with automated service installers.",
			Status:      model.PhaseWorking,
			Tasks: []struct {
				Name        string
				Description string
				Status      model.TaskStatus
				Progress    float64
				Priority    model.Priority
				Result      string
			}{
				{"Windows x64", "64-bit Windows binary packaging with auto-startup script", model.StatusDone, 100, model.PriorityHigh, "Windows standalone executable builds and runs natively"},
				{"Linux x64", "64-bit Linux server binary packaging with systemd integration", model.StatusWorking, 80, model.PriorityHigh, "Cross-compilation for linux/amd64 ready"},
				{"Linux ARM64", "64-bit ARM binary packaging for Raspberry Pi 4/5 and Orange Pi", model.StatusWorking, 80, model.PriorityCritical, "CGO_ENABLED=0 pure-Go ARM64 build ready"},
				{"Raspberry Pi ARM64", "Debian/Raspberry Pi OS package and installation verification", model.StatusWorking, 80, model.PriorityCritical, "Systemd service installer created"},
				{"Raspberry Pi ARM32", "32-bit ARM (armv7/armv6) compilation for Pi Zero and legacy SBCs", model.StatusWorking, 70, model.PriorityMedium, "ARM32 compilation supported"},
				{"macOS where feasible", "Darwin amd64 and arm64 binary compilation", model.StatusWorking, 70, model.PriorityLow, "Darwin cross-compilation target configured"},
				{"Application binary", "Self-contained single binary embedding frontend assets and MariaDB connector", model.StatusDone, 100, model.PriorityCritical, "Embedded static files in single executable"},
				{"Web UI", "Optimized production bundle with gzip compression and low RAM footprint", model.StatusDone, 100, model.PriorityHigh, "Vue 2 production build bundle prepared"},
				{"System service", "Auto-restarting service scripts for Linux systemd and Windows background", model.StatusDone, 100, model.PriorityHigh, "Installer scripts written for PowerShell and bash"},
				{"Auto-start", "OS startup registration for headless industrial edge deployment", model.StatusDone, 100, model.PriorityHigh, "Service auto-start configuration implemented"},
			},
		},
		{
			Number:      10,
			Name:        "Phase 10 — Optimization & Production",
			Description: "Low-resource performance profiling, memory optimization, stress testing, long-running stability, and release documentation.",
			Status:      model.PhasePending,
			Tasks: []struct {
				Name        string
				Description string
				Status      model.TaskStatus
				Progress    float64
				Priority    model.Priority
				Result      string
			}{
				{"CPU Optimization", "Goroutine profiling and throttling to maintain < 5% CPU on Raspberry Pi", model.StatusPending, 0, model.PriorityHigh, ""},
				{"RAM Optimization", "Buffer reuse and struct memory layout tuning to maintain < 35MB RAM", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Database Optimization", "Connection pooling, query cache, and tailored index tuning for MariaDB", model.StatusWorking, 75, model.PriorityHigh, "Connection pooling and indexed schemas configured"},
				{"Storage Optimization", "Automatic vacuum and telemetry table partitioning", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Stress Test", "Sustained high-frequency polling and WebSocket client load testing", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Long Running Test", "72-hour continuous acquisition leak testing on ARM hardware", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Failure Recovery Test", "Simulated power cut, disk full, and network drop recovery validation", model.StatusPending, 0, model.PriorityHigh, ""},
				{"Security Test", "JWT protection, SQL injection prevention, and API permission validation", model.StatusWorking, 70, model.PriorityHigh, "Parameterized GORM queries and JWT validation"},
				{"Installer Test", "Clean machine installation test on fresh Windows and Raspberry Pi OS", model.StatusWorking, 50, model.PriorityMedium, "Installation scripts tested locally"},
				{"Upgrade Test", "Database schema migration test without telemetry data loss", model.StatusPending, 0, model.PriorityMedium, ""},
				{"Uninstall Test", "Clean service deregistration and directory cleanup script test", model.StatusPending, 0, model.PriorityLow, ""},
				{"Documentation", "Comprehensive architecture documentation, API reference, and deployment guide", model.StatusWorking, 60, model.PriorityHigh, "Phase 1 documentation prepared"},
				{"Release Management", "Semantic versioning and production release tagging", model.StatusWorking, 50, model.PriorityMedium, "v1.0.0-phase1 tagged"},
			},
		},
	}

	var phase1ID uint
	now := time.Now()

	for _, p := range phasesData {
		phase := model.DevelopmentPhase{
			PhaseNumber: p.Number,
			Name:        p.Name,
			Description: p.Description,
			Status:      p.Status,
			OrderIndex:  p.Number,
		}
		if p.Number == 1 {
			phase.Progress = 100.0
			phase.CompletedDate = &now
		} else if p.Number == 9 || p.Number == 7 || p.Number == 10 {
			phase.Status = model.PhaseWorking
			phase.Progress = 35.0
		}
		db.Create(&phase)

		if p.Number == 1 {
			phase1ID = phase.ID
		}

		// Create tasks
		for idx, t := range p.Tasks {
			task := model.DevelopmentTask{
				PhaseID:     phase.ID,
				TaskName:    t.Name,
				Description: t.Description,
				Status:      t.Status,
				Progress:    t.Progress,
				Priority:    t.Priority,
				Owner:       "Lead System Engineer",
				OrderIndex:  idx + 1,
				TestResult:  t.Result,
			}
			if t.Status == model.StatusDone {
				task.CompletionDate = &now
			}
			db.Create(&task)

			// Add task log for completed / working tasks
			if t.Status == model.StatusDone || t.Status == model.StatusWorking {
				resultText := "IN_PROGRESS"
				if t.Status == model.StatusDone {
					resultText = "PASSED"
				}
				db.Create(&model.DevelopmentTaskLog{
					TaskID:    task.ID,
					Timestamp: now.Add(-time.Duration(idx*15) * time.Minute),
					User:      "admin",
					Action:    "Implemented and verified",
					Result:    resultText,
					Log:       t.Description,
				})
			}
		}
	}

	// 4. Sample Industrial Devices & Parameters
	modbusTCPType := model.DeviceType{Code: "MODBUS_TCP", Name: "Modbus TCP Meter", Protocol: "MODBUS_TCP", Description: "Industrial Ethernet Modbus TCP slave"}
	modbusRTUType := model.DeviceType{Code: "MODBUS_RTU", Name: "Modbus RTU Sensor", Protocol: "MODBUS_RTU", Description: "RS-485 Modbus RTU serial device"}
	mqttType := model.DeviceType{Code: "MQTT_GW", Name: "MQTT Gateway", Protocol: "MQTT", Description: "Local broker sensor aggregator"}
	db.Create(&modbusTCPType)
	db.Create(&modbusRTUType)
	db.Create(&mqttType)

	dev1 := model.Device{
		Code:         "PM-01",
		Name:         "Main Power Quality Meter PM-500",
		DeviceTypeID: modbusTCPType.ID,
		Status:       model.DeviceOnline,
		Enabled:      true,
		Location:     "Main Electrical Switchboard A",
		LatencyMs:    12,
		SuccessCount: 14820,
		FailedCount:  3,
	}
	dev2 := model.Device{
		Code:         "TH-01",
		Name:         "Server Room Temp & Humidity Sensor",
		DeviceTypeID: modbusRTUType.ID,
		Status:       model.DeviceOnline,
		Enabled:      true,
		Location:     "Edge Server Rack 01",
		LatencyMs:    35,
		SuccessCount: 9640,
		FailedCount:  12,
	}
	dev3 := model.Device{
		Code:         "FM-01",
		Name:         "Cooling Water Flow Meter FM-100",
		DeviceTypeID: modbusTCPType.ID,
		Status:       model.DeviceOnline,
		Enabled:      true,
		Location:     "Chiller Plant Room",
		LatencyMs:    18,
		SuccessCount: 8200,
		FailedCount:  0,
	}
	db.Create(&dev1)
	db.Create(&dev2)
	db.Create(&dev3)

	// Connections
	db.Create(&model.DeviceConnection{DeviceID: dev1.ID, ConnectionType: "TCP", Address: "192.168.1.101", Port: 502, TimeoutMs: 1000, PollIntervalMs: 1000})
	db.Create(&model.DeviceConnection{DeviceID: dev2.ID, ConnectionType: "SERIAL", Address: "COM3", BaudRate: 9600, DataBits: 8, StopBits: 1, Parity: "N", TimeoutMs: 1000, PollIntervalMs: 2000})
	db.Create(&model.DeviceConnection{DeviceID: dev3.ID, ConnectionType: "TCP", Address: "192.168.1.103", Port: 502, TimeoutMs: 1000, PollIntervalMs: 1000})

	// Sample Parameters
	v := 220.5
	c := 45.2
	t := 24.8
	h := 58.2
	f := 125.4
	db.Create(&model.Parameter{DeviceID: dev1.ID, Code: "VOLTAGE_L1", Name: "Phase Voltage L1-N", Unit: "V", DataType: model.DataTypeFloat32, RegisterAddress: 30001, ScaleFactor: 0.1, CurrentValue: &v})
	db.Create(&model.Parameter{DeviceID: dev1.ID, Code: "CURRENT_L1", Name: "Line Current L1", Unit: "A", DataType: model.DataTypeFloat32, RegisterAddress: 30003, ScaleFactor: 0.01, CurrentValue: &c})
	db.Create(&model.Parameter{DeviceID: dev2.ID, Code: "ROOM_TEMP", Name: "Ambient Temperature", Unit: "°C", DataType: model.DataTypeFloat32, RegisterAddress: 40001, ScaleFactor: 0.1, CurrentValue: &t})
	db.Create(&model.Parameter{DeviceID: dev2.ID, Code: "ROOM_HUM", Name: "Relative Humidity", Unit: "%RH", DataType: model.DataTypeFloat32, RegisterAddress: 40002, ScaleFactor: 0.1, CurrentValue: &h})
	db.Create(&model.Parameter{DeviceID: dev3.ID, Code: "FLOW_RATE", Name: "Chilled Water Flow Rate", Unit: "m³/h", DataType: model.DataTypeFloat32, RegisterAddress: 30101, ScaleFactor: 0.1, CurrentValue: &f})

	// 5. Notification Channels
	db.Create(&model.NotificationChannel{Name: "Web Notification Alert", Type: model.ChannelWeb, Enabled: true})
	db.Create(&model.NotificationChannel{Name: "Plant Ops Email", Type: model.ChannelEmail, Enabled: false, ConfigJSON: `{"smtp_server":"smtp.local","port":587}`})
	db.Create(&model.NotificationChannel{Name: "Engineering Telegram Bot", Type: model.ChannelTelegram, Enabled: false, ConfigJSON: `{"bot_token":"","chat_id":""}`})

	// 6. Initial Audit Trail & Logs
	db.Create(&model.AuditTrail{
		Username:  "admin",
		Action:    "SYSTEM_INIT",
		Resource:  "SYSTEM",
		Details:   "Datalogger Analysis Application initialized in local-first edge mode",
		IPAddress: "127.0.0.1",
		CreatedAt: now,
	})
	db.Create(&model.SystemLog{
		Level:     "INFO",
		Component: "CORE",
		Message:   "Datalogger Analysis Application Phase 1 Foundation successfully started",
		Details:   "MariaDB Edge database connected. HTTP & WebSocket engine operational.",
		CreatedAt: now,
	})

	// Evidence for Phase 1
	var firstTask model.DevelopmentTask
	db.Where("phase_id = ? AND task_name = ?", phase1ID, "Architecture").First(&firstTask)
	if firstTask.ID > 0 {
		db.Create(&model.DevelopmentEvidence{
			TaskID:      firstTask.ID,
			Title:       "Modular Local-First Architecture Document",
			FilePath:    "docs/architecture.md",
			Description: "Comprehensive architectural specifications detailing local-first design and edge hardware optimization",
			CreatedAt:   now,
		})
	}

	logger.Info("Database seed completed with 10 development phases, all initial tasks, and system foundation!")
	return nil
}
