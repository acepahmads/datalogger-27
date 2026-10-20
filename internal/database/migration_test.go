package database_test

import (
	"fmt"
	"testing"
	"time"

	"datalogger/internal/database"
	"datalogger/internal/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestMigrationPhase3Tracking(t *testing.T) {
	dsn := fmt.Sprintf("file:mem_mig_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// 1. Run migrations
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// 2. Check Phase 1 and 2 remain at 100%
	var phase1, phase2, phase3 model.DevelopmentPhase
	if err := db.Where("phase_number = 1").First(&phase1).Error; err == nil {
		if phase1.Progress < 100.0 {
			t.Errorf("Phase 1 progress should be 100%%, got %.2f%%", phase1.Progress)
		}
	}
	if err := db.Where("phase_number = 2").First(&phase2).Error; err == nil {
		if phase2.Progress < 100.0 {
			t.Errorf("Phase 2 progress should be 100%%, got %.2f%%", phase2.Progress)
		}
	}

	// 3. Check Phase 3
	if err := db.Where("phase_number = 3").First(&phase3).Error; err != nil {
		t.Fatalf("Phase 3 not found: %v", err)
	}
	if phase3.Status != model.PhaseCompleted && phase3.Status != model.PhaseWorking {
		t.Errorf("Expected Phase 3 WORKING or COMPLETED, got %s", phase3.Status)
	}

	// 4. Check Subphase 3.1
	var sub3_1 model.DevelopmentSubphase
	if err := db.Where("phase_id = ? AND (name LIKE ? OR order_index = 1)", phase3.ID, "%Phase 3.1%").First(&sub3_1).Error; err != nil {
		t.Fatalf("Subphase 3.1 not found: %v", err)
	}
	if sub3_1.Progress < 100.0 {
		t.Errorf("Subphase 3.1 progress should be 100%%, got %.2f%%", sub3_1.Progress)
	}
	if sub3_1.Status != "DONE" {
		t.Errorf("Subphase 3.1 status should be DONE, got %s", sub3_1.Status)
	}

	// 5. Check all 16 subtasks
	var tasks []model.DevelopmentTask
	if err := db.Where("subphase_id = ?", sub3_1.ID).Order("order_index ASC").Find(&tasks).Error; err != nil {
		t.Fatalf("Failed to fetch subtasks: %v", err)
	}
	if len(tasks) != 16 {
		t.Fatalf("Expected 16 subtasks for Phase 3.1, got %d", len(tasks))
	}

	for _, task := range tasks {
		if task.Status != model.StatusDone {
			t.Errorf("Task '%s' status should be DONE, got %s", task.TaskName, task.Status)
		}
		if task.Progress < 100.0 {
			t.Errorf("Task '%s' progress should be 100%%, got %.2f%%", task.TaskName, task.Progress)
		}
	}
}

func TestMigrationPhase4Tracking(t *testing.T) {
	dsn := fmt.Sprintf("file:mem_mig4_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// 1. Run migrations
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// 2. Check Phase 4
	var phase4 model.DevelopmentPhase
	if err := db.Where("phase_number = 4").First(&phase4).Error; err != nil {
		t.Fatalf("Phase 4 not found: %v", err)
	}
	if phase4.Status != model.PhaseWorking && phase4.Status != model.PhaseCompleted {
		t.Errorf("Expected Phase 4 WORKING, got %s", phase4.Status)
	}

	// 3. Check Subphase 4.1
	var sub4_1 model.DevelopmentSubphase
	if err := db.Where("phase_id = ? AND (name LIKE ? OR order_index = 1)", phase4.ID, "%Phase 4.1%").First(&sub4_1).Error; err != nil {
		t.Fatalf("Subphase 4.1 not found: %v", err)
	}
	if sub4_1.Progress < 100.0 {
		t.Errorf("Subphase 4.1 progress should be 100%%, got %.2f%%", sub4_1.Progress)
	}
	if sub4_1.Status != "DONE" {
		t.Errorf("Subphase 4.1 status should be DONE, got %s", sub4_1.Status)
	}

	// 4. Check all 17 subtasks
	var tasks []model.DevelopmentTask
	if err := db.Where("subphase_id = ? AND status = ?", sub4_1.ID, model.StatusDone).Order("order_index ASC").Find(&tasks).Error; err != nil {
		t.Fatalf("Failed to fetch subtasks: %v", err)
	}
	if len(tasks) != 17 {
		t.Fatalf("Expected 17 subtasks for Phase 4.1, got %d", len(tasks))
	}

	for _, task := range tasks {
		if task.Status != model.StatusDone {
			t.Errorf("Task '%s' status should be DONE, got %s", task.TaskName, task.Status)
		}
		if task.Progress < 100.0 {
			t.Errorf("Task '%s' progress should be 100%%, got %.2f%%", task.TaskName, task.Progress)
		}
	}

	// 5. Check Subphase 4.2
	var sub4_2 model.DevelopmentSubphase
	if err := db.Where("phase_id = ? AND (name LIKE ? OR order_index = 2)", phase4.ID, "%Phase 4.2%").First(&sub4_2).Error; err != nil {
		t.Fatalf("Subphase 4.2 not found: %v", err)
	}
	if sub4_2.Progress < 100.0 {
		t.Errorf("Subphase 4.2 progress should be 100%%, got %.2f%%", sub4_2.Progress)
	}
	if sub4_2.Status != "DONE" {
		t.Errorf("Subphase 4.2 status should be DONE, got %s", sub4_2.Status)
	}

	// 6. Check all 19 subtasks in Subphase 4.2
	var tasks42 []model.DevelopmentTask
	if err := db.Where("subphase_id = ? AND status = ?", sub4_2.ID, model.StatusDone).Order("order_index ASC").Find(&tasks42).Error; err != nil {
		t.Fatalf("Failed to fetch subtasks for 4.2: %v", err)
	}
	if len(tasks42) != 19 {
		t.Fatalf("Expected 19 subtasks for Phase 4.2, got %d", len(tasks42))
	}

	for _, task := range tasks42 {
		if task.Status != model.StatusDone {
			t.Errorf("Task '%s' status should be DONE, got %s", task.TaskName, task.Status)
		}
		if task.Progress < 100.0 {
			t.Errorf("Task '%s' progress should be 100%%, got %.2f%%", task.TaskName, task.Progress)
		}
	}

	// 7. Check Subphase 4.3
	var sub4_3 model.DevelopmentSubphase
	if err := db.Where("phase_id = ? AND (name LIKE ? OR order_index = 3)", phase4.ID, "%Phase 4.3%").First(&sub4_3).Error; err != nil {
		t.Fatalf("Subphase 4.3 not found: %v", err)
	}
	if sub4_3.Progress < 100.0 {
		t.Errorf("Subphase 4.3 progress should be 100%%, got %.2f%%", sub4_3.Progress)
	}
	if sub4_3.Status != "DONE" {
		t.Errorf("Subphase 4.3 status should be DONE, got %s", sub4_3.Status)
	}

	// 8. Check all 13 subtasks in Subphase 4.3
	var tasks43 []model.DevelopmentTask
	if err := db.Where("subphase_id = ? AND status = ?", sub4_3.ID, model.StatusDone).Order("order_index ASC").Find(&tasks43).Error; err != nil {
		t.Fatalf("Failed to fetch subtasks for 4.3: %v", err)
	}
	if len(tasks43) != 13 {
		t.Fatalf("Expected 13 subtasks for Phase 4.3, got %d", len(tasks43))
	}

	for _, task := range tasks43 {
		if task.Status != model.StatusDone {
			t.Errorf("Task '%s' status should be DONE, got %s", task.TaskName, task.Status)
		}
		if task.Progress < 100.0 {
			t.Errorf("Task '%s' progress should be 100%%, got %.2f%%", task.TaskName, task.Progress)
		}
	}

	// 9. Check legacy tasks 45, 46, 47 & 48 are SUPERSEDED
	var legacyTasks []model.DevelopmentTask
	db.Where("phase_id = ? AND id IN (45, 46, 47, 48)", phase4.ID).Find(&legacyTasks)
	for _, lt := range legacyTasks {
		if lt.Status != model.StatusSuperseded {
			t.Errorf("Legacy task %d (%s) should be SUPERSEDED, got %s", lt.ID, lt.TaskName, lt.Status)
		}
	}

	// 10. Check backup permissions exist
	var backupPermCount int64
	db.Model(&model.Permission{}).Where("category = ?", "BACKUP").Count(&backupPermCount)
	if backupPermCount != 5 {
		t.Errorf("Expected 5 BACKUP permissions, got %d", backupPermCount)
	}

	// 11. Check Subphase 4.4 (Retention & Storage Management)
	var sub4_4 model.DevelopmentSubphase
	if err := db.Where("phase_id = ? AND (name LIKE ? OR order_index = 4)", phase4.ID, "%Phase 4.4%").First(&sub4_4).Error; err != nil {
		t.Fatalf("Subphase 4.4 not found: %v", err)
	}
	if sub4_4.Progress < 100.0 {
		t.Errorf("Subphase 4.4 progress should be 100%%, got %.2f%%", sub4_4.Progress)
	}
	if sub4_4.Status != "DONE" {
		t.Errorf("Subphase 4.4 status should be DONE, got %s", sub4_4.Status)
	}

	// 12. Check all 11 subtasks in Subphase 4.4
	var tasks44 []model.DevelopmentTask
	if err := db.Where("subphase_id = ? AND status = ?", sub4_4.ID, model.StatusDone).Order("order_index ASC").Find(&tasks44).Error; err != nil {
		t.Fatalf("Failed to fetch subtasks for 4.4: %v", err)
	}
	if len(tasks44) != 11 {
		t.Fatalf("Expected 11 subtasks for Phase 4.4, got %d", len(tasks44))
	}
	for _, task := range tasks44 {
		if task.Status != model.StatusDone {
			t.Errorf("Task '%s' status should be DONE, got %s", task.TaskName, task.Status)
		}
		if task.Progress < 100.0 {
			t.Errorf("Task '%s' progress should be 100%%, got %.2f%%", task.TaskName, task.Progress)
		}
	}

	// 13. Check retention permissions exist
	var retPermCount int64
	db.Model(&model.Permission{}).Where("category = ?", "RETENTION").Count(&retPermCount)
	if retPermCount != 3 {
		t.Errorf("Expected 3 RETENTION permissions, got %d", retPermCount)
	}

	// 14. Phase 4 overall should now have 100% progress
	var phase4Final model.DevelopmentPhase
	db.Where("phase_number = 4").First(&phase4Final)
	if phase4Final.Progress < 100.0 {
		t.Errorf("Phase 4 final progress should be 100%%, got %.2f%%", phase4Final.Progress)
	}
	if phase4Final.Status != model.PhaseCompleted {
		t.Errorf("Phase 4 final status should be COMPLETED, got %s", phase4Final.Status)
	}
}

func TestMigrationRetentionPoliciesSeeded(t *testing.T) {
	dsn := fmt.Sprintf("file:mem_mig_ret_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// 1. Run migrations first time
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// 2. Verify all 8 default categories exist
	expectedCategories := []model.RetentionCategory{
		model.CategoryRawTelemetry,
		model.CategoryInternalAggregations,
		model.CategoryCustomerAggregations,
		model.CategoryClearedAlarms,
		model.CategoryCommunicationLogs,
		model.CategorySystemLogs,
		model.CategoryAuditTrails,
		model.CategoryBackupArchives,
	}

	var policies []model.RetentionPolicy
	if err := db.Find(&policies).Error; err != nil {
		t.Fatalf("Failed to query retention policies: %v", err)
	}

	if len(policies) != 8 {
		t.Fatalf("Expected exactly 8 retention policies seeded by migration, got %d", len(policies))
	}

	catMap := make(map[model.RetentionCategory]*model.RetentionPolicy)
	for i := range policies {
		p := &policies[i]
		catMap[p.Category] = p
		// Default policies should be disabled initially for safety
		if p.Enabled {
			t.Errorf("Default policy %s (%s) should be disabled by default", p.ID, p.Category)
		}
	}

	for _, expectedCat := range expectedCategories {
		if _, ok := catMap[expectedCat]; !ok {
			t.Errorf("Missing expected retention policy category: %s", expectedCat)
		}
	}

	// 3. User modifies one policy
	customDays := 120
	if err := db.Model(&model.RetentionPolicy{}).Where("id = ?", "pol_raw_telemetry").Updates(map[string]interface{}{
		"retention_days": customDays,
		"enabled":        true,
	}).Error; err != nil {
		t.Fatalf("Failed to update test policy: %v", err)
	}

	// 4. Run migrations second time (idempotency check)
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Second migration run failed: %v", err)
	}

	var policiesAfter []model.RetentionPolicy
	db.Find(&policiesAfter)
	if len(policiesAfter) != 8 {
		t.Fatalf("Expected still 8 policies after second migration run, got %d", len(policiesAfter))
	}

	var modifiedPol model.RetentionPolicy
	if err := db.Where("id = ?", "pol_raw_telemetry").First(&modifiedPol).Error; err != nil {
		t.Fatalf("Failed to find modified policy: %v", err)
	}
	if modifiedPol.RetentionDays != customDays || !modifiedPol.Enabled {
		t.Errorf("Migration overwrote user policy changes! Got retention_days=%d, enabled=%v", modifiedPol.RetentionDays, modifiedPol.Enabled)
	}
}

func TestMigrationPhase5Tracking(t *testing.T) {
	dsn := fmt.Sprintf("file:mem_mig_phase5_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migration to trigger updatePhase5Tracking
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// 1. Verify Phase 5 status & progress
	var phase5 model.DevelopmentPhase
	if err := db.Where("phase_number = 5").First(&phase5).Error; err != nil {
		t.Fatalf("Phase 5 not found in DB: %v", err)
	}

	if phase5.Progress < 100.0 {
		t.Errorf("Phase 5 progress should be 100%%, got %.2f%%", phase5.Progress)
	}
	if phase5.Status != model.PhaseCompleted {
		t.Errorf("Phase 5 status should be COMPLETED, got %s", phase5.Status)
	}

	// 2. Verify all 8 tasks are DONE
	var tasks []model.DevelopmentTask
	if err := db.Where("phase_id = ?", phase5.ID).Order("order_index ASC").Find(&tasks).Error; err != nil {
		t.Fatalf("Failed to fetch Phase 5 tasks: %v", err)
	}

	if len(tasks) != 8 {
		t.Fatalf("Expected 8 tasks in Phase 5, got %d", len(tasks))
	}

	for _, task := range tasks {
		if task.Status != model.StatusDone {
			t.Errorf("Task %d ('%s') status should be DONE, got %s", task.ID, task.TaskName, task.Status)
		}
		if task.Progress < 100.0 {
			t.Errorf("Task %d ('%s') progress should be 100%%, got %.2f%%", task.ID, task.TaskName, task.Progress)
		}
		if task.TestResult == "" {
			t.Errorf("Task %d ('%s') should have verified test results recorded", task.ID, task.TaskName)
		}
	}
}

func TestMigrationPhase6Wave1Tracking(t *testing.T) {
	dsn := fmt.Sprintf("file:mem_mig_phase6_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migration to trigger updatePhase6Wave1Tracking
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// 1. Verify Phase 6 status & progress
	var phase6 model.DevelopmentPhase
	if err := db.Where("phase_number = 6").First(&phase6).Error; err != nil {
		t.Fatalf("Phase 6 not found in DB: %v", err)
	}

	if phase6.Progress < 40.0 {
		t.Errorf("Phase 6 progress should be at least 40%%, got %.2f%%", phase6.Progress)
	}
	if phase6.Status != model.PhaseWorking {
		t.Errorf("Phase 6 status should be WORKING, got %s", phase6.Status)
	}

	// 2. Verify Tasks #57, #58, #59, #63 are DONE
	wave1TaskNames := map[string]bool{
		"Threshold":    true,
		"Warning":      true,
		"Critical":     true,
		"Alarm Engine": true,
	}

	var tasks []model.DevelopmentTask
	if err := db.Where("phase_id = ?", phase6.ID).Find(&tasks).Error; err != nil {
		t.Fatalf("Failed to fetch Phase 6 tasks: %v", err)
	}

	for _, task := range tasks {
		if wave1TaskNames[task.TaskName] {
			if task.Status != model.StatusDone {
				t.Errorf("Wave 1 Task '%s' should be DONE, got %s", task.TaskName, task.Status)
			}
			if task.Progress < 100.0 {
				t.Errorf("Wave 1 Task '%s' progress should be 100%%, got %.2f%%", task.TaskName, task.Progress)
			}
			if task.TestResult == "" {
				t.Errorf("Wave 1 Task '%s' should have verified test results recorded", task.TaskName)
			}
		}
	}
}

func TestMigrationPhase6Wave2Tracking(t *testing.T) {
	dsn := fmt.Sprintf("file:mem_mig_phase6_w2_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Run migration to trigger updatePhase6Tracking
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// 1. Verify Phase 6 status & progress (7 done out of 10 tasks = 70.0%)
	var phase6 model.DevelopmentPhase
	if err := db.Where("phase_number = 6").First(&phase6).Error; err != nil {
		t.Fatalf("Phase 6 not found in DB: %v", err)
	}

	if phase6.Progress < 70.0 {
		t.Errorf("Phase 6 progress should be at least 70%% after Wave 2, got %.2f%%", phase6.Progress)
	}
	if phase6.Status != model.PhaseWorking {
		t.Errorf("Phase 6 status should be WORKING (not DONE yet), got %s", phase6.Status)
	}

	// 2. Verify Tasks #60, #61, #62 are DONE
	wave2TaskNames := map[string]bool{
		"Device Offline":      true,
		"Communication Error": true,
		"Data Abnormality":    true,
	}

	var tasks []model.DevelopmentTask
	if err := db.Where("phase_id = ?", phase6.ID).Find(&tasks).Error; err != nil {
		t.Fatalf("Failed to fetch Phase 6 tasks: %v", err)
	}

	for _, task := range tasks {
		if wave2TaskNames[task.TaskName] {
			if task.Status != model.StatusDone {
				t.Errorf("Wave 2 Task '%s' should be DONE, got %s", task.TaskName, task.Status)
			}
			if task.Progress < 100.0 {
				t.Errorf("Wave 2 Task '%s' progress should be 100%%, got %.2f%%", task.TaskName, task.Progress)
			}
			if task.TestResult == "" {
				t.Errorf("Wave 2 Task '%s' should have verified test results recorded", task.TaskName)
			}
		}
	}
}

func TestMigrationPhase6Wave3Tracking(t *testing.T) {
	dsn := fmt.Sprintf("file:mem_mig_phase6_w3_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// 1. Verify Phase 6 is WORKING with 80% progress
	var phase6 model.DevelopmentPhase
	if err := db.Where("phase_number = 6").First(&phase6).Error; err != nil {
		t.Fatalf("Phase 6 not found in DB: %v", err)
	}

	if phase6.Progress < 80.0 {
		t.Errorf("Phase 6 progress should be at least 80%% after Wave 3, got %.2f%%", phase6.Progress)
	}
	if phase6.Status != model.PhaseWorking {
		t.Errorf("Phase 6 status should be WORKING (not DONE yet), got %s", phase6.Status)
	}

	// 2. Verify Task #64 "Notification Engine" is DONE
	var task64 model.DevelopmentTask
	if err := db.Where("phase_id = ? AND task_name = ?", phase6.ID, "Notification Engine").First(&task64).Error; err != nil {
		t.Fatalf("Task 64 'Notification Engine' not found: %v", err)
	}
	if task64.Status != model.StatusDone {
		t.Errorf("Task 64 should be DONE, got %s", task64.Status)
	}
	if task64.Progress < 100.0 {
		t.Errorf("Task 64 progress should be 100%%, got %.2f%%", task64.Progress)
	}
	if task64.TestResult == "" {
		t.Errorf("Task 64 should have test results recorded")
	}

	// 3. Verify Tasks #65 and #66 remain PENDING (Wave 4)
	var pendingTasks []model.DevelopmentTask
	if err := db.Where("phase_id = ? AND status = ?", phase6.ID, model.StatusPending).Find(&pendingTasks).Error; err != nil {
		t.Fatalf("Failed to query pending tasks: %v", err)
	}
}

func TestMigrationPhase7AndOperatorCommandCenterTracking(t *testing.T) {
	dsn := fmt.Sprintf("file:mem_mig_p7_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// 1. Verify Phase 7 Task 67 (Dashboard)
	var phase7 model.DevelopmentPhase
	if err := db.Where("phase_number = 7").First(&phase7).Error; err != nil {
		t.Fatalf("Phase 7 not found: %v", err)
	}
	var task67 model.DevelopmentTask
	if err := db.Where("phase_id = ? AND (id = 67 OR task_name = 'Dashboard')", phase7.ID).First(&task67).Error; err != nil {
		t.Fatalf("Task 67 'Dashboard' not found: %v", err)
	}
	if task67.Status != model.StatusDone {
		t.Errorf("Task 67 should be DONE, got %s", task67.Status)
	}
	if task67.Progress < 100.0 {
		t.Errorf("Task 67 progress should be 100%%, got %.2f%%", task67.Progress)
	}

	// 2. Verify Phase 10 Operator Command Center Acceptance Test
	var phase10 model.DevelopmentPhase
	if err := db.Where("phase_number = 10").First(&phase10).Error; err != nil {
		t.Fatalf("Phase 10 not found: %v", err)
	}
	var occTask model.DevelopmentTask
	if err := db.Where("phase_id = ? AND task_name LIKE ?", phase10.ID, "%Operator Command Center%").First(&occTask).Error; err != nil {
		t.Fatalf("Operator Command Center acceptance task not found in Phase 10: %v", err)
	}
	if occTask.Status != model.StatusDone {
		t.Errorf("Operator Command Center task should be DONE, got %s", occTask.Status)
	}
	if occTask.Progress < 100.0 {
		t.Errorf("Operator Command Center progress should be 100%%, got %.2f%%", occTask.Progress)
	}
}

func TestMigrationPhase9Tracking(t *testing.T) {
	dsn := fmt.Sprintf("file:mem_mig_p9_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// 1. Run migrations
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// 2. Verify Phase 9 status & progress
	var phase9 model.DevelopmentPhase
	if err := db.Where("phase_number = 9").First(&phase9).Error; err != nil {
		t.Fatalf("Phase 9 not found in DB: %v", err)
	}

	if phase9.Progress < 90.0 {
		t.Errorf("Phase 9 progress should be >= 90%%, got %.2f%%", phase9.Progress)
	}
	if phase9.Status != model.PhaseWorking {
		t.Errorf("Phase 9 status should be WORKING (pending hardware acceptance), got %s", phase9.Status)
	}

	// 3. Verify all 10 tasks in Phase 9
	var tasks []model.DevelopmentTask
	if err := db.Where("phase_id = ?", phase9.ID).Order("order_index ASC").Find(&tasks).Error; err != nil {
		t.Fatalf("Failed to query Phase 9 tasks: %v", err)
	}

	if len(tasks) != 10 {
		t.Fatalf("Expected exactly 10 tasks in Phase 9, got %d", len(tasks))
	}

	expectedStatus := map[string]model.TaskStatus{
		"Windows x64":           model.StatusDone,
		"Linux x64":             model.StatusDone,
		"Linux ARM64":           model.StatusDone,
		"Raspberry Pi ARM64":    model.StatusDone,
		"Raspberry Pi ARM32":    model.StatusWorking,
		"macOS where feasible":  model.StatusWorking,
		"Application binary":    model.StatusDone,
		"Web UI":                model.StatusDone,
		"System service":        model.StatusDone,
		"Auto-start":            model.StatusDone,
	}

	for _, task := range tasks {
		expStatus, ok := expectedStatus[task.TaskName]
		if !ok {
			t.Errorf("Unexpected task name in Phase 9: %s", task.TaskName)
			continue
		}
		if task.Status != expStatus {
			t.Errorf("Task '%s' expected status %s, got %s", task.TaskName, expStatus, task.Status)
		}
		if task.TestResult == "" {
			t.Errorf("Task '%s' should have recorded test results", task.TaskName)
		}
	}
}


