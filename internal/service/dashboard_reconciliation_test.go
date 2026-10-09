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

// 7. Test Progress Summary DTO reflects accepted Phase 3 milestones
func TestProgressSummaryDTO(t *testing.T) {
	db := setupReconciledDB(t)
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)

	summary, err := phaseService.GetProgressSummary()
	if err != nil {
		t.Fatalf("GetProgressSummary failed: %v", err)
	}

	if summary.CurrentPhaseNumber != 4 && summary.CurrentPhaseNumber != 3 {
		t.Errorf("Expected CurrentPhaseNumber 4 or 3, got %d", summary.CurrentPhaseNumber)
	}
	if summary.CompletedTasksCount == 0 {
		t.Errorf("Expected CompletedTasksCount > 0, got 0")
	}
	if summary.CurrentPhaseNumber == 4 {
		validSubphases := map[string]bool{
			"Phase 4.1 — Reliability Foundation & Auto-Recovery": true,
			"Phase 4.2 — Persistent Queue & Data Integrity":      true,
			"Phase 4.3 — Backup & Disaster Recovery":             true,
			"Phase 4.4 — Retention & Storage Management":         true,
		}
		if !validSubphases[summary.CurrentSubphase] {
			t.Errorf("Unexpected CurrentSubphase: %s", summary.CurrentSubphase)
		}
	} else if summary.CurrentPhaseNumber == 3 {
		if summary.CurrentSubphase != "Phase 3.3 — Aggregation, Rollup & Downsampling" {
			t.Errorf("Unexpected CurrentSubphase: %s", summary.CurrentSubphase)
		}
	}
}

// 8. Test Phase 3 progress calculation is data-driven and genuinely 100%
func TestPhase3ProgressCalculation(t *testing.T) {
	db := setupReconciledDB(t)
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)

	phase, err := phaseService.GetPhaseByID(3)
	if err != nil {
		t.Fatalf("Failed to fetch Phase 3: %v", err)
	}

	if phase.Progress != 100.0 {
		t.Errorf("Expected Phase 3 progress 100.0%%, got %.2f%%", phase.Progress)
	}
	if phase.Status != model.PhaseCompleted {
		t.Errorf("Expected Phase 3 status COMPLETED, got %s", phase.Status)
	}
	if phase.CompletedDate == nil {
		t.Errorf("Expected Phase 3 CompletedDate to be set")
	}
}

// 9. Test Phase 3.1, 3.2, and 3.3 completion preservation
func TestPhase3SubphaseCompletionPreservation(t *testing.T) {
	db := setupReconciledDB(t)

	// Subphase 3.1
	var sub3_1 model.DevelopmentSubphase
	if err := db.Where("name LIKE ?", "%Phase 3.1%").First(&sub3_1).Error; err != nil {
		t.Fatalf("Subphase 3.1 not found: %v", err)
	}
	if sub3_1.Progress != 100.0 || sub3_1.Status != "DONE" || sub3_1.AcceptanceCriteria != "16/16 PASS" {
		t.Errorf("Subphase 3.1 criteria mismatch: progress=%.1f, status=%s, criteria=%s",
			sub3_1.Progress, sub3_1.Status, sub3_1.AcceptanceCriteria)
	}
	var active3_1 []model.DevelopmentTask
	db.Where("subphase_id = ? AND status = ?", sub3_1.ID, model.StatusDone).Find(&active3_1)
	if len(active3_1) != 16 {
		t.Errorf("Expected 16 active DONE tasks in Subphase 3.1, got %d", len(active3_1))
	}

	// Subphase 3.2
	var sub3_2 model.DevelopmentSubphase
	if err := db.Where("name LIKE ?", "%Phase 3.2%").First(&sub3_2).Error; err != nil {
		t.Fatalf("Subphase 3.2 not found: %v", err)
	}
	if sub3_2.Progress != 100.0 || sub3_2.Status != "DONE" || sub3_2.AcceptanceCriteria != "22/22 PASS" {
		t.Errorf("Subphase 3.2 criteria mismatch: progress=%.1f, status=%s, criteria=%s",
			sub3_2.Progress, sub3_2.Status, sub3_2.AcceptanceCriteria)
	}
	var active3_2 []model.DevelopmentTask
	db.Where("subphase_id = ? AND status = ?", sub3_2.ID, model.StatusDone).Find(&active3_2)
	if len(active3_2) != 22 {
		t.Errorf("Expected 22 active DONE tasks in Subphase 3.2, got %d", len(active3_2))
	}

	// Subphase 3.3
	var sub3_3 model.DevelopmentSubphase
	if err := db.Where("name LIKE ?", "%Phase 3.3%").First(&sub3_3).Error; err != nil {
		t.Fatalf("Subphase 3.3 not found: %v", err)
	}
	if sub3_3.Progress != 100.0 || sub3_3.Status != "DONE" || sub3_3.AcceptanceCriteria != "27/27 PASS" {
		t.Errorf("Subphase 3.3 criteria mismatch: progress=%.1f, status=%s, criteria=%s",
			sub3_3.Progress, sub3_3.Status, sub3_3.AcceptanceCriteria)
	}
	var active3_3 []model.DevelopmentTask
	db.Where("subphase_id = ? AND status = ?", sub3_3.ID, model.StatusDone).Find(&active3_3)
	if len(active3_3) != 27 {
		t.Errorf("Expected 27 active DONE tasks in Subphase 3.3, got %d", len(active3_3))
	}
}

// 10. Test Active vs Superseded task counting in Phase 3
func TestPhase3ActiveVsSupersededCounting(t *testing.T) {
	db := setupReconciledDB(t)

	var totalTasks int64
	db.Model(&model.DevelopmentTask{}).Where("phase_id = 3").Count(&totalTasks)
	if totalTasks != 83 {
		t.Errorf("Expected 83 total tasks in Phase 3, got %d", totalTasks)
	}

	var activeTasks int64
	db.Model(&model.DevelopmentTask{}).Where("phase_id = 3 AND status != ? AND status != ?",
		model.StatusSuperseded, model.StatusPlanned).Count(&activeTasks)
	if activeTasks != 65 {
		t.Errorf("Expected 65 active deliverables in Phase 3, got %d", activeTasks)
	}

	var completedActiveTasks int64
	db.Model(&model.DevelopmentTask{}).Where("phase_id = 3 AND status = ?", model.StatusDone).Count(&completedActiveTasks)
	if completedActiveTasks != 65 {
		t.Errorf("Expected 65 completed active deliverables in Phase 3, got %d", completedActiveTasks)
	}

	var supersededTasks int64
	db.Model(&model.DevelopmentTask{}).Where("phase_id = 3 AND status = ?", model.StatusSuperseded).Count(&supersededTasks)
	if supersededTasks != 18 {
		t.Errorf("Expected 18 superseded legacy tasks in Phase 3, got %d", supersededTasks)
	}
}

// 11. Test Legacy task mapping (#22 to #39)
func TestPhase3LegacyTaskMapping(t *testing.T) {
	db := setupReconciledDB(t)

	legacyTasks := []struct {
		ID             uint
		Name           string
		ExpectedSubIdx int // 1 for 3.1, 2 for 3.2, 3 for 3.3
	}{
		{22, "Raw Data Acquisition", 1},
		{23, "Protocol Parsing", 1},
		{24, "Data Mapping", 1},
		{25, "Data Validation", 2},
		{26, "Scaling", 2},
		{27, "Conversion", 2},
		{28, "Formula", 2},
		{29, "Scheduler", 1},
		{30, "Polling", 1},
		{31, "Average", 3},
		{32, "Min", 3},
		{33, "Max", 3},
		{34, "Aggregation", 3},
		{35, "Spike Detection", 2},
		{36, "Outlier Detection", 2},
		{37, "Data Quality", 2},
		{38, "Buffer", 1},
		{39, "Queue", 1},
	}

	for _, lt := range legacyTasks {
		var task model.DevelopmentTask
		if err := db.Where("id = ? AND phase_id = 3", lt.ID).First(&task).Error; err != nil {
			t.Errorf("Legacy task #%d (%s) not found: %v", lt.ID, lt.Name, err)
			continue
		}

		if task.SubphaseID == nil {
			t.Errorf("Legacy task #%d (%s) has nil SubphaseID", lt.ID, lt.Name)
		}
		if task.Status != model.StatusSuperseded {
			t.Errorf("Expected legacy task #%d (%s) status SUPERSEDED, got %s", lt.ID, lt.Name, task.Status)
		}
		if task.Progress != 100.0 {
			t.Errorf("Expected legacy task #%d (%s) progress 100%%, got %.0f%%", lt.ID, lt.Name, task.Progress)
		}
		if task.TestResult == "" || len(task.TestResult) < 20 {
			t.Errorf("Expected legacy task #%d (%s) to have descriptive TestResult, got %q", lt.ID, lt.Name, task.TestResult)
		}
	}
}

// 12. Test Orphaned task detection: Zero orphaned tasks in Phase 3
func TestPhase3NoOrphanedTasks(t *testing.T) {
	db := setupReconciledDB(t)

	var orphanedCount int64
	db.Model(&model.DevelopmentTask{}).Where("phase_id = 3 AND subphase_id IS NULL").Count(&orphanedCount)
	if orphanedCount != 0 {
		t.Errorf("Found %d orphaned tasks with subphase_id IS NULL in Phase 3", orphanedCount)
	}
}

// 13. Test Duplicate task detection: Exactly 83 tasks, unique IDs
func TestPhase3NoDuplicateTasks(t *testing.T) {
	db := setupReconciledDB(t)

	var tasks []model.DevelopmentTask
	db.Where("phase_id = 3").Find(&tasks)

	if len(tasks) != 83 {
		t.Errorf("Expected exactly 83 tasks in Phase 3, got %d", len(tasks))
	}

	seenIDs := make(map[uint]bool)
	for _, task := range tasks {
		if seenIDs[task.ID] {
			t.Errorf("Duplicate task ID %d detected in Phase 3", task.ID)
		}
		seenIDs[task.ID] = true
	}
}

// 14. Test Migration Idempotency for Phase 3
func TestPhase3MigrationIdempotency(t *testing.T) {
	db := setupReconciledDB(t)

	var countBefore int64
	db.Model(&model.DevelopmentTask{}).Where("phase_id = 3").Count(&countBefore)

	// Run migration 3 more times
	for i := 0; i < 3; i++ {
		if err := database.RunMigrations(db); err != nil {
			t.Fatalf("RunMigrations failed at iteration %d: %v", i+1, err)
		}
	}

	var countAfter int64
	db.Model(&model.DevelopmentTask{}).Where("phase_id = 3").Count(&countAfter)

	if countBefore != countAfter {
		t.Errorf("Task count changed after repeated migrations! Before: %d, After: %d", countBefore, countAfter)
	}

	var phase3 model.DevelopmentPhase
	db.Where("phase_number = 3").First(&phase3)
	if phase3.Progress != 100.0 {
		t.Errorf("Expected Phase 3 progress 100.0%% after repeated migrations, got %.2f%%", phase3.Progress)
	}
	if phase3.Status != model.PhaseCompleted {
		t.Errorf("Expected Phase 3 status COMPLETED after repeated migrations, got %s", phase3.Status)
	}
}

// 15. Test Reconciliation task tracked under Phase 10
func TestPhase10ReconciliationAuditTask(t *testing.T) {
	db := setupReconciledDB(t)

	var auditTask model.DevelopmentTask
	if err := db.Where("task_name = ?", "Phase 3 Progress Audit & Reconciliation").First(&auditTask).Error; err != nil {
		t.Fatalf("Audit task not found in database: %v", err)
	}

	if auditTask.Status != model.StatusDone {
		t.Errorf("Expected audit task status DONE, got %s", auditTask.Status)
	}
	if auditTask.Progress != 100.0 {
		t.Errorf("Expected audit task progress 100%%, got %.0f%%", auditTask.Progress)
	}
	if auditTask.Notes == "" {
		t.Errorf("Expected audit task to contain detailed notes")
	}

	var auditTrail model.AuditTrail
	if err := db.Where("action = ?", "RECONCILE_PHASE_3").First(&auditTrail).Error; err != nil {
		t.Errorf("Expected AuditTrail record for RECONCILE_PHASE_3: %v", err)
	}
}

// 16. Test Hotfix task tracked under Phase 10
func TestPhase10HotfixTask(t *testing.T) {
	db := setupReconciledDB(t)

	var hotfixTask model.DevelopmentTask
	if err := db.Where("task_name = ?", "Phase 3 Legacy Task Rendering Hotfix").First(&hotfixTask).Error; err != nil {
		t.Fatalf("Hotfix task not found in database: %v", err)
	}

	if hotfixTask.Status != model.StatusDone {
		t.Errorf("Expected hotfix task status DONE, got %s", hotfixTask.Status)
	}
	if hotfixTask.Progress != 100.0 {
		t.Errorf("Expected hotfix task progress 100%%, got %.0f%%", hotfixTask.Progress)
	}
	if hotfixTask.Notes == "" {
		t.Errorf("Expected hotfix task to contain detailed notes")
	}
	if hotfixTask.TestResult == "" {
		t.Errorf("Expected hotfix task to contain test result")
	}
}

// 17. Test PhaseService UpdateTask with SubphaseID
func TestPhaseServiceUpdateTaskSubphaseID(t *testing.T) {
	db := setupReconciledDB(t)
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)

	// Fetch a task
	task, err := phaseService.GetTaskByID(22)
	if err != nil {
		t.Fatalf("Failed to fetch task 22: %v", err)
	}

	targetSubphaseID := float64(5)
	updatedTask, err := phaseService.UpdateTask(task.ID, map[string]interface{}{
		"subphase_id": targetSubphaseID,
		"status":      "SUPERSEDED",
		"progress":    float64(100),
	}, "test-admin")
	if err != nil {
		t.Fatalf("Failed to update task 22: %v", err)
	}

	if updatedTask.SubphaseID == nil || *updatedTask.SubphaseID != uint(targetSubphaseID) {
		t.Errorf("Expected SubphaseID %d, got %v", uint(targetSubphaseID), updatedTask.SubphaseID)
	}
}


