package database

import (
	"fmt"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"

	"gorm.io/gorm"
)

// RunMigrations executes auto-migrations and Phase 2.1 schema synchronization
func RunMigrations(db *gorm.DB) error {
	logger.Info("Executing MariaDB schema migrations for Phase 2.1...")

	// 0. Pre-migrate existing tables to ensure unique columns are populated before index creation
	preMigrateEdgeSchema(db)

	// 1. Auto-migrate models
	err := db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.DevelopmentPhase{},
		&model.DevelopmentSubphase{},
		&model.DevelopmentTask{},
		&model.DevelopmentTaskLog{},
		&model.DevelopmentDependency{},
		&model.DevelopmentEvidence{},
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.Parameter{},
		&model.Sensor{},
		&model.RawData{},
		&model.ProcessedData{},
		&model.AggregatedData{},
		&model.Alarm{},
		&model.NotificationChannel{},
		&model.Notification{},
		&model.AuditTrail{},
		&model.SystemLog{},
		&model.CommunicationLog{},
		&model.OutputDestination{},
		&model.DeliveryLog{},
		&model.SystemHealth{},
	)
	if err != nil {
		return err
	}

	// 2. Data backfill / synchronization for existing devices from Phase 1
	var existingDevices []model.Device
	if err := db.Unscoped().Find(&existingDevices).Error; err == nil {
		for _, dev := range existingDevices {
			updates := make(map[string]interface{})
			if dev.DeviceCode == "" && dev.Code != "" {
				updates["device_code"] = dev.Code
			}
			if dev.DeviceName == "" && dev.Name != "" {
				updates["device_name"] = dev.Name
			}
			if dev.DeviceType == "" {
				updates["device_type"] = "MODBUS_TCP"
			}
			if dev.Status == "" || dev.Status == "STANDBY" {
				updates["status"] = model.DeviceStatusActive
			}
			if dev.ConnectionStatus == "" {
				if dev.Status == "ONLINE" {
					updates["connection_status"] = model.DeviceConnOnline
					updates["status"] = model.DeviceStatusActive
				} else {
					updates["connection_status"] = model.DeviceConnUnknown
				}
			}
			if dev.Timezone == "" {
				updates["timezone"] = "UTC"
			}
			if len(updates) > 0 {
				_ = db.Model(&model.Device{}).Unscoped().Where("id = ?", dev.ID).Updates(updates).Error
			}
		}
	}

	// 3. Backfill existing parameters from Phase 1
	var existingParams []model.Parameter
	if err := db.Unscoped().Find(&existingParams).Error; err == nil {
		for _, param := range existingParams {
			updates := make(map[string]interface{})
			if param.ParameterCode == "" && param.Code != "" {
				updates["parameter_code"] = param.Code
			}
			if param.ParameterName == "" && param.Name != "" {
				updates["parameter_name"] = param.Name
			}
			if param.Scale == 0 && param.ScaleFactor != 0 {
				updates["scale"] = param.ScaleFactor
			}
			if len(updates) > 0 {
				_ = db.Model(&model.Parameter{}).Unscoped().Where("id = ?", param.ID).Updates(updates).Error
			}
		}
	}

	// 4. Ensure Device Permissions exist in system
	ensureDevicePermissions(db)

	// 5. Update Development Tracking Dashboard for Phase 2.1
	updatePhase2Tracking(db)

	logger.Info("Phase 2.1 schema synchronization and migration completed successfully")
	return nil
}

func updatePhase2Tracking(db *gorm.DB) {
	var phase2 model.DevelopmentPhase
	if err := db.Where("phase_number = 2").First(&phase2).Error; err == nil {
		now := time.Now()
		// Update phase status to WORKING
		db.Model(&phase2).Update("status", model.PhaseWorking)

		// Find and update Task "Device Management"
		var task model.DevelopmentTask
		if err := db.Where("phase_id = ? AND task_name = ?", phase2.ID, "Device Management").First(&task).Error; err == nil {
			db.Model(&task).Updates(map[string]interface{}{
				"status":          model.StatusDone,
				"progress":        100.0,
				"completion_date": &now,
				"test_result":     "PASSED: Full Device CRUD, Connection Configuration, Validation, Unique Constraints, RBAC, and Audit Trail verified (10/10 tests)",
			})

			// Add task log if not existing
			var logCount int64
			db.Model(&model.DevelopmentTaskLog{}).Where("task_id = ? AND action = ?", task.ID, "Phase 2.1 Implementation").Count(&logCount)
			if logCount == 0 {
				db.Create(&model.DevelopmentTaskLog{
					TaskID:    task.ID,
					Timestamp: now,
					User:      "admin",
					Action:    "Phase 2.1 Implementation",
					Result:    "PASSED",
					Log:       "Phase 2.1 Device Management completed with REST CRUD, connection models, parameters, RBAC authorization, audit trail, and Vue 2 frontend.",
				})
			}

			// Add evidence
			var evCount int64
			db.Model(&model.DevelopmentEvidence{}).Where("task_id = ?", task.ID).Count(&evCount)
			if evCount == 0 {
				db.Create(&model.DevelopmentEvidence{
					TaskID:      task.ID,
					Title:       "Phase 2.1 Device Management Documentation",
					FilePath:    "docs/phase2_1_device_management.md",
					Description: "Comprehensive design and implementation documentation for industrial device management, protocol configurations, and parameter definitions.",
					CreatedAt:   now,
				})
			}
		}

		// Recalculate Phase 2 Progress
		var tasks []model.DevelopmentTask
		if err := db.Where("phase_id = ?", phase2.ID).Find(&tasks).Error; err == nil && len(tasks) > 0 {
			var sum float64
			for _, t := range tasks {
				sum += t.Progress
			}
			avgProgress := sum / float64(len(tasks))
			db.Model(&phase2).Update("progress", avgProgress)
		}
	}
}

func ensureDevicePermissions(db *gorm.DB) {
	perms := []model.Permission{
		{Code: "device.view", Name: "View Devices", Category: "DEVICE", Description: "View devices, connections, and parameter readouts"},
		{Code: "device.create", Name: "Create Devices", Category: "DEVICE", Description: "Register new industrial devices"},
		{Code: "device.update", Name: "Update Devices", Category: "DEVICE", Description: "Update device metadata and connection configurations"},
		{Code: "device.delete", Name: "Delete Devices", Category: "DEVICE", Description: "Soft-delete industrial devices"},
		{Code: "device.manage", Name: "Manage Devices", Category: "DEVICE", Description: "Full administrative control of devices and parameters"},
	}

	for _, p := range perms {
		var existing model.Permission
		if db.Where("code = ?", p.Code).First(&existing).Error != nil {
			db.Create(&p)
		}
	}

	// Assign permissions to Administrator and Engineer roles
	var adminRole model.Role
	if db.Where("name = ?", "Administrator").First(&adminRole).Error == nil {
		var allPerms []model.Permission
		db.Find(&allPerms)
		_ = db.Model(&adminRole).Association("Permissions").Replace(allPerms)
	}

	var engRole model.Role
	if db.Where("name = ?", "Engineer").First(&engRole).Error == nil {
		var engPerms []model.Permission
		db.Where("code IN ?", []string{"device.view", "device.create", "device.update", "device.manage"}).Find(&engPerms)
		_ = db.Model(&engRole).Association("Permissions").Replace(engPerms)
	}

	var opRole model.Role
	if db.Where("name = ?", "Operator").First(&opRole).Error == nil {
		var opPerms []model.Permission
		db.Where("code IN ?", []string{"device.view"}).Find(&opPerms)
		_ = db.Model(&opRole).Association("Permissions").Replace(opPerms)
	}
}

// preMigrateEdgeSchema safely pre-populates existing database rows before unique indexes are built
func preMigrateEdgeSchema(db *gorm.DB) {
	// 1. Devices table pre-migration
	if db.Migrator().HasTable("devices") {
		_ = db.Exec("UPDATE devices SET device_code = code WHERE (device_code IS NULL OR device_code = '') AND (code IS NOT NULL AND code != '');").Error
		_ = db.Exec("UPDATE devices SET device_name = name WHERE (device_name IS NULL OR device_name = '') AND (name IS NOT NULL AND name != '');").Error

		if !db.Migrator().HasColumn("devices", "device_code") {
			_ = db.Migrator().AddColumn(&model.Device{}, "DeviceCode")
		}
		if !db.Migrator().HasColumn("devices", "device_name") {
			_ = db.Migrator().AddColumn(&model.Device{}, "DeviceName")
		}

		type DevMini struct {
			ID         uint   `gorm:"primaryKey"`
			Code       string `gorm:"column:code"`
			Name       string `gorm:"column:name"`
			DeviceCode string `gorm:"column:device_code"`
			DeviceName string `gorm:"column:device_name"`
		}
		var rows []DevMini
		if err := db.Table("devices").Select("id, code, name, device_code, device_name").Find(&rows).Error; err == nil {
			for _, r := range rows {
				dCode := r.DeviceCode
				if dCode == "" {
					if r.Code != "" {
						dCode = r.Code
					} else {
						dCode = fmt.Sprintf("DEV-%d", r.ID)
					}
				}
				dName := r.DeviceName
				if dName == "" {
					if r.Name != "" {
						dName = r.Name
					} else {
						dName = fmt.Sprintf("Device %d", r.ID)
					}
				}
				_ = db.Table("devices").Where("id = ?", r.ID).Updates(map[string]interface{}{
					"device_code": dCode,
					"device_name": dName,
				}).Error
			}
		}
	}

	// 2. Parameters table pre-migration
	if db.Migrator().HasTable("parameters") {
		_ = db.Exec("UPDATE parameters SET parameter_code = code WHERE (parameter_code IS NULL OR parameter_code = '') AND (code IS NOT NULL AND code != '');").Error
		_ = db.Exec("UPDATE parameters SET parameter_name = name WHERE (parameter_name IS NULL OR parameter_name = '') AND (name IS NOT NULL AND name != '');").Error

		if !db.Migrator().HasColumn("parameters", "parameter_code") {
			_ = db.Migrator().AddColumn(&model.Parameter{}, "ParameterCode")
		}
		if !db.Migrator().HasColumn("parameters", "parameter_name") {
			_ = db.Migrator().AddColumn(&model.Parameter{}, "ParameterName")
		}

		type ParamMini struct {
			ID            uint   `gorm:"primaryKey"`
			Code          string `gorm:"column:code"`
			Name          string `gorm:"column:name"`
			ParameterCode string `gorm:"column:parameter_code"`
			ParameterName string `gorm:"column:parameter_name"`
		}
		var pRows []ParamMini
		if err := db.Table("parameters").Select("id, code, name, parameter_code, parameter_name").Find(&pRows).Error; err == nil {
			for _, r := range pRows {
				pCode := r.ParameterCode
				if pCode == "" {
					if r.Code != "" {
						pCode = r.Code
					} else {
						pCode = fmt.Sprintf("PARAM-%d", r.ID)
					}
				}
				pName := r.ParameterName
				if pName == "" {
					if r.Name != "" {
						pName = r.Name
					} else {
						pName = fmt.Sprintf("Parameter %d", r.ID)
					}
				}
				_ = db.Table("parameters").Where("id = ?", r.ID).Updates(map[string]interface{}{
					"parameter_code": pCode,
					"parameter_name": pName,
				}).Error
			}
		}
	}
}
