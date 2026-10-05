package service_test

import (
	"fmt"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:memdb_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(
		&model.User{},
		&model.Role{},
		&model.Permission{},
		&model.DevelopmentPhase{},
		&model.DevelopmentSubphase{},
		&model.DevelopmentTask{},
		&model.DevelopmentTaskLog{},
		&model.DevelopmentDependency{},
		&model.DevelopmentEvidence{},
		&model.AuditTrail{},
		&model.SystemLog{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestPhaseProgressCalculation(t *testing.T) {
	db := setupTestDB(t)
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)

	// Create Phase
	phase := model.DevelopmentPhase{
		PhaseNumber: 1,
		Name:        "Phase 1 — Foundation",
		Description: "Architecture and Foundation",
		Status:      model.PhasePending,
		Progress:    0,
	}
	db.Create(&phase)

	// Add 2 Tasks: 1 Done (100%), 1 In Progress (50%)
	task1 := model.DevelopmentTask{
		PhaseID:     phase.ID,
		TaskName:    "Architecture",
		Status:      model.StatusDone,
		Progress:    100,
		OrderIndex:  1,
	}
	task2 := model.DevelopmentTask{
		PhaseID:     phase.ID,
		TaskName:    "Go Backend",
		Status:      model.StatusWorking,
		Progress:    50,
		OrderIndex:  2,
	}
	db.Create(&task1)
	db.Create(&task2)

	// Recalculate
	avg, err := phaseRepo.RecalculateProgress(phase.ID)
	if err != nil {
		t.Fatalf("RecalculateProgress failed: %v", err)
	}

	expected := 75.0 // (100 + 50) / 2
	if avg != expected {
		t.Errorf("Expected average progress %f, got %f", expected, avg)
	}

	// Verify phase status
	p, err := phaseService.GetPhaseByID(phase.ID)
	if err != nil {
		t.Fatalf("GetPhaseByID failed: %v", err)
	}
	if p.Status != model.PhaseWorking {
		t.Errorf("Expected phase status WORKING, got %s", p.Status)
	}

	// Complete second task
	_, err = phaseService.UpdateTask(task2.ID, map[string]interface{}{
		"status":   "DONE",
		"progress": float64(100),
	}, "test-admin")
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	pUpdated, _ := phaseService.GetPhaseByID(phase.ID)
	if pUpdated.Progress != 100.0 {
		t.Errorf("Expected phase progress 100.0, got %f", pUpdated.Progress)
	}
	if pUpdated.Status != model.PhaseCompleted {
		t.Errorf("Expected phase status COMPLETED, got %s", pUpdated.Status)
	}
}

func TestTaskUpdateAndActivityLog(t *testing.T) {
	db := setupTestDB(t)
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)

	phase := model.DevelopmentPhase{
		PhaseNumber: 2,
		Name:        "Phase 2 — Device & Communication",
		Status:      model.PhasePending,
	}
	db.Create(&phase)

	task := model.DevelopmentTask{
		PhaseID:  phase.ID,
		TaskName: "Modbus RTU",
		Status:   model.StatusPending,
		Progress: 0,
	}
	db.Create(&task)

	// Update task status
	_, err := phaseService.UpdateTask(task.ID, map[string]interface{}{
		"status":      "WORKING",
		"progress":    float64(45),
		"test_result": "Initial serial frame verified",
	}, "engineer-1")
	if err != nil {
		t.Fatalf("UpdateTask failed: %v", err)
	}

	// Verify activity log was automatically generated
	activities, err := phaseService.GetActivityTimeline(10)
	if err != nil {
		t.Fatalf("GetActivityTimeline failed: %v", err)
	}
	if len(activities) == 0 {
		t.Fatal("Expected at least one activity log entry, got 0")
	}

	latest := activities[0]
	if latest.User != "engineer-1" {
		t.Errorf("Expected user engineer-1, got %s", latest.User)
	}
	if latest.Task != "Modbus RTU" {
		t.Errorf("Expected task Modbus RTU, got %s", latest.Task)
	}
}
