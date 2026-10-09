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

	// 9. Check legacy tasks 45 & 46 are SUPERSEDED
	var legacyBackup model.DevelopmentTask
	if err := db.Where("phase_id = ? AND (task_name = ? OR id = 45)", phase4.ID, "Backup").First(&legacyBackup).Error; err == nil {
		if legacyBackup.Status != model.StatusSuperseded {
			t.Errorf("Legacy task 45 should be SUPERSEDED, got %s", legacyBackup.Status)
		}
	}

	// 10. Check backup permissions exist
	var backupPermCount int64
	db.Model(&model.Permission{}).Where("category = ?", "BACKUP").Count(&backupPermCount)
	if backupPermCount != 5 {
		t.Errorf("Expected 5 BACKUP permissions, got %d", backupPermCount)
	}
}

