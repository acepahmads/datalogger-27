package service_test

import (
	"fmt"
	"testing"
	"time"

	"datalogger/internal/database"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupReconciledDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:memreconcile_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test DB: %v", err)
	}

	// 1. Migrate all tables
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
		t.Fatalf("Failed to automigrate models: %v", err)
	}

	// 2. Initial seed (Phase 1 + 10 Phases)
	if err := database.Seed(db); err != nil {
		t.Fatalf("Seed failed: %v", err)
	}

	// 3. Run reconciliation migration
	if err := database.RunMigrations(db); err != nil {
		t.Fatalf("RunMigrations failed: %v", err)
	}

	return db
}

// 1. Test Phase 2.1 remains DONE 100% backed by 22/22 criteria
func TestPhase2_1Reconciliation(t *testing.T) {
	db := setupReconciledDB(t)

	var sub2_1 model.DevelopmentSubphase
	if err := db.Where("name LIKE ?", "%Phase 2.1%").First(&sub2_1).Error; err != nil {
		t.Fatalf("Subphase 2.1 not found: %v", err)
	}

	if sub2_1.Progress != 100.0 {
		t.Errorf("Expected Subphase 2.1 progress 100.0%%, got %.2f%%", sub2_1.Progress)
	}
	if sub2_1.Status != "DONE" {
		t.Errorf("Expected Subphase 2.1 status DONE, got %s", sub2_1.Status)
	}
	if sub2_1.AcceptanceCriteria != "22/22 PASS" {
		t.Errorf("Expected Subphase 2.1 acceptance 22/22 PASS, got %s", sub2_1.AcceptanceCriteria)
	}

	var tasks []model.DevelopmentTask
	db.Where("subphase_id = ?", sub2_1.ID).Find(&tasks)
	if len(tasks) == 0 {
		t.Fatalf("Expected child tasks for Subphase 2.1, got 0")
	}
	for _, task := range tasks {
		if task.Status != model.StatusDone {
			t.Errorf("Expected task %s to be DONE, got %s", task.TaskName, task.Status)
		}
		if task.Progress != 100.0 {
			t.Errorf("Expected task %s progress 100%%, got %.0f%%", task.TaskName, task.Progress)
		}
	}
}

// 2. Test Phase 2.2 remains DONE 100% backed by 27/27 criteria
func TestPhase2_2Reconciliation(t *testing.T) {
	db := setupReconciledDB(t)

	var sub2_2 model.DevelopmentSubphase
	if err := db.Where("name LIKE ?", "%Phase 2.2%").First(&sub2_2).Error; err != nil {
		t.Fatalf("Subphase 2.2 not found: %v", err)
	}

	if sub2_2.Progress != 100.0 {
		t.Errorf("Expected Subphase 2.2 progress 100.0%%, got %.2f%%", sub2_2.Progress)
	}
	if sub2_2.Status != "DONE" {
		t.Errorf("Expected Subphase 2.2 status DONE, got %s", sub2_2.Status)
	}
	if sub2_2.AcceptanceCriteria != "27/27 PASS" {
		t.Errorf("Expected Subphase 2.2 acceptance 27/27 PASS, got %s", sub2_2.AcceptanceCriteria)
	}

	// Verify all 14 granular subtasks are DONE 100%
	var activeTasks []model.DevelopmentTask
	db.Where("subphase_id = ? AND status = ?", sub2_2.ID, model.StatusDone).Find(&activeTasks)
	if len(activeTasks) < 14 {
		t.Errorf("Expected at least 14 active DONE subtasks in Phase 2.2, got %d", len(activeTasks))
	}
}

// 3. Test Phase 2.3 remains DONE 100% backed by 32/32 criteria with real hardware DEFERRED
func TestPhase2_3Reconciliation(t *testing.T) {
	db := setupReconciledDB(t)

	var sub2_3 model.DevelopmentSubphase
	if err := db.Where("name LIKE ?", "%Phase 2.3%").First(&sub2_3).Error; err != nil {
		t.Fatalf("Subphase 2.3 not found: %v", err)
	}

	if sub2_3.Progress != 100.0 {
		t.Errorf("Expected Subphase 2.3 progress 100.0%%, got %.2f%%", sub2_3.Progress)
	}
	if sub2_3.Status != "DONE" {
		t.Errorf("Expected Subphase 2.3 status DONE, got %s", sub2_3.Status)
	}
	if sub2_3.AcceptanceCriteria != "32/32 PASS" {
		t.Errorf("Expected Subphase 2.3 acceptance 32/32 PASS, got %s", sub2_3.AcceptanceCriteria)
	}

	// Verify physical hardware validation remains DEFERRED (Simulator PASS)
	var rtuTask, tcpTask model.DevelopmentTask
	db.Where("task_name = ?", "2.3.2 Real Modbus RTU Validation").First(&rtuTask)
	db.Where("task_name = ?", "2.3.3 Real Modbus TCP Validation").First(&tcpTask)

	if rtuTask.ID == 0 || tcpTask.ID == 0 {
		t.Fatalf("Expected 2.3.2 and 2.3.3 tasks to exist")
	}

	if rtuTask.TestResult != "SIMULATOR PASS: FC 01-04 and CRC-16 checks verified | REAL HARDWARE: DEFERRED (Hardware Not Available)" {
		t.Errorf("Unexpected RTU task result: %s", rtuTask.TestResult)
	}
	if tcpTask.TestResult != "SIMULATOR PASS: MBAP framing and socket reuse verified | REAL HARDWARE: DEFERRED (Hardware Not Available)" {
		t.Errorf("Unexpected TCP task result: %s", tcpTask.TestResult)
	}
}

// 4. Test Phase 2 overall progress is data-driven and genuinely 100%
func TestPhase2OverallProgress(t *testing.T) {
	db := setupReconciledDB(t)
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)

	phase, err := phaseService.GetPhaseByID(2)
	if err != nil {
		t.Fatalf("Failed to fetch Phase 2: %v", err)
	}

	if phase.Progress != 100.0 {
		t.Errorf("Expected Phase 2 overall progress 100.0%%, got %.2f%%", phase.Progress)
	}
	if phase.Status != model.PhaseCompleted {
		t.Errorf("Expected Phase 2 status COMPLETED, got %s", phase.Status)
	}
}

// 5. Test Legacy & Future tasks do NOT reduce completed Phase 2 progress
func TestLegacyAndFutureTasksIsolation(t *testing.T) {
	db := setupReconciledDB(t)

	// Verify future protocols are PLANNED and not DONE
	futureProtocols := []string{"Device Discovery", "TCP", "UDP", "MQTT"}
	for _, name := range futureProtocols {
		var task model.DevelopmentTask
		db.Where("phase_id = 2 AND task_name LIKE ?", name+"%").First(&task)
		if task.ID == 0 {
			t.Errorf("Expected future task %s to exist", name)
			continue
		}
		if task.Status != model.StatusPlanned {
			t.Errorf("Expected task %s to have status PLANNED, got %s", name, task.Status)
		}
		if task.Progress != 0.0 {
			t.Errorf("Expected task %s progress 0%%, got %.0f%%", name, task.Progress)
		}
	}

	// Verify superseded tasks are SUPERSEDED with 100% progress
	supersededProtocols := []string{"Modbus RTU", "Modbus TCP", "REST API", "Serial Communication", "Protocol Adapter"}
	for _, name := range supersededProtocols {
		var task model.DevelopmentTask
		db.Where("phase_id = 2 AND task_name = ?", name).First(&task)
		if task.ID == 0 {
			t.Errorf("Expected superseded task %s to exist", name)
			continue
		}
		if task.Status != model.StatusSuperseded {
			t.Errorf("Expected task %s to have status SUPERSEDED, got %s", name, task.Status)
		}
	}
}

// 6. Test Migration Idempotency: Running migrations multiple times does not duplicate tasks
func TestMigrationIdempotency(t *testing.T) {
	db := setupReconciledDB(t)

	var countBefore int64
	db.Model(&model.DevelopmentTask{}).Where("phase_id = 2").Count(&countBefore)

	// Run migration again 2 more times
	for i := 0; i < 2; i++ {
		if err := database.RunMigrations(db); err != nil {
			t.Fatalf("Idempotent RunMigrations iteration %d failed: %v", i+1, err)
		}
	}

	var countAfter int64
	db.Model(&model.DevelopmentTask{}).Where("phase_id = 2").Count(&countAfter)

	if countBefore != countAfter {
		t.Errorf("Duplicate tasks detected! Before: %d, After: %d", countBefore, countAfter)
	}

	// Verify progress remains 100%
	var phase model.DevelopmentPhase
	db.Where("phase_number = 2").First(&phase)
	if phase.Progress != 100.0 {
		t.Errorf("Expected Phase 2 progress 100.0%% after repeated migrations, got %.2f%%", phase.Progress)
	}
}

// 7. Test Progress Summary DTO reflects accepted milestones
func TestProgressSummaryDTO(t *testing.T) {
	db := setupReconciledDB(t)
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)

	summary, err := phaseService.GetProgressSummary()
	if err != nil {
		t.Fatalf("GetProgressSummary failed: %v", err)
	}

	if summary.CurrentPhaseNumber != 2 {
		t.Errorf("Expected CurrentPhaseNumber 2, got %d", summary.CurrentPhaseNumber)
	}
	if summary.CurrentSubphase != "Phase 2.3 — Communication Hardening & Real Device Validation" {
		t.Errorf("Unexpected CurrentSubphase: %s", summary.CurrentSubphase)
	}
	if summary.CompletedTasksCount == 0 {
		t.Errorf("Expected CompletedTasksCount > 0, got 0")
	}
}
