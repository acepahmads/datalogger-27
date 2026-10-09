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

// setupReconciliationTestDB initializes an isolated in-memory SQLite DB for testing status rules
func setupReconciliationTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:mem_reconcile_test_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	err = db.AutoMigrate(
		&model.DevelopmentPhase{},
		&model.DevelopmentSubphase{},
		&model.DevelopmentTask{},
		&model.DevelopmentTaskLog{},
		&model.AuditTrail{},
		&model.SystemLog{},
	)
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

// 1. All subphases DONE and progress 100% -> parent phase COMPLETED
func TestReconciliation_AllSubphasesDoneAndProgress100_ParentCompleted(t *testing.T) {
	db := setupReconciliationTestDB(t)
	repo := repository.NewPhaseRepository(db)

	phase := model.DevelopmentPhase{
		PhaseNumber: 4,
		Name:        "Phase 4 — Reliability & Storage",
		Status:      model.PhaseWorking,
		Progress:    75.0,
		OrderIndex:  4,
	}
	db.Create(&phase)

	sub1 := model.DevelopmentSubphase{
		PhaseID:    phase.ID,
		Name:       "Phase 4.1 — Foundation",
		Status:     "PENDING",
		OrderIndex: 1,
	}
	sub2 := model.DevelopmentSubphase{
		PhaseID:    phase.ID,
		Name:       "Phase 4.2 — Queue",
		Status:     "PENDING",
		OrderIndex: 2,
	}
	db.Create(&sub1)
	db.Create(&sub2)

	// Create completed tasks for subphase 1
	task1 := model.DevelopmentTask{
		PhaseID:    phase.ID,
		SubphaseID: &sub1.ID,
		TaskName:   "Task 1",
		Status:     model.StatusDone,
		Progress:   100.0,
	}
	task2 := model.DevelopmentTask{
		PhaseID:    phase.ID,
		SubphaseID: &sub1.ID,
		TaskName:   "Task 2",
		Status:     model.StatusDone,
		Progress:   100.0,
	}
	// Create completed tasks for subphase 2
	task3 := model.DevelopmentTask{
		PhaseID:    phase.ID,
		SubphaseID: &sub2.ID,
		TaskName:   "Task 3",
		Status:     model.StatusDone,
		Progress:   100.0,
	}
	db.Create(&task1)
	db.Create(&task2)
	db.Create(&task3)

	// Trigger reconciliation
	err := repo.ReconcileAllPhases()
	if err != nil {
		t.Fatalf("ReconcileAllPhases failed: %v", err)
	}

	// Verify subphases are both DONE at 100%
	var fetchedSub1, fetchedSub2 model.DevelopmentSubphase
	db.First(&fetchedSub1, sub1.ID)
	db.First(&fetchedSub2, sub2.ID)
	if fetchedSub1.Status != "DONE" || fetchedSub1.Progress != 100.0 {
		t.Errorf("Expected sub1 DONE at 100%%, got %s at %.1f%%", fetchedSub1.Status, fetchedSub1.Progress)
	}
	if fetchedSub2.Status != "DONE" || fetchedSub2.Progress != 100.0 {
		t.Errorf("Expected sub2 DONE at 100%%, got %s at %.1f%%", fetchedSub2.Status, fetchedSub2.Progress)
	}

	// Verify parent phase is COMPLETED with 100% progress and non-nil completed_date
	var fetchedPhase model.DevelopmentPhase
	db.First(&fetchedPhase, phase.ID)
	if fetchedPhase.Status != model.PhaseCompleted {
		t.Errorf("Expected parent phase status COMPLETED, got %s", fetchedPhase.Status)
	}
	if fetchedPhase.Progress != 100.0 {
		t.Errorf("Expected parent phase progress 100%%, got %.1f%%", fetchedPhase.Progress)
	}
	if fetchedPhase.CompletedDate == nil {
		t.Errorf("Expected CompletedDate to be set for COMPLETED phase")
	}
}

// 2. One required task still WORKING -> parent remains WORKING
func TestReconciliation_OneRequiredTaskWorking_ParentRemainsWorking(t *testing.T) {
	db := setupReconciliationTestDB(t)
	repo := repository.NewPhaseRepository(db)

	phase := model.DevelopmentPhase{
		PhaseNumber: 4,
		Name:        "Phase 4 — Reliability & Storage",
		Status:      model.PhaseWorking,
		Progress:    50.0,
		OrderIndex:  4,
	}
	db.Create(&phase)

	task1 := model.DevelopmentTask{
		PhaseID:  phase.ID,
		TaskName: "Task 1 Done",
		Status:   model.StatusDone,
		Progress: 100.0,
	}
	task2 := model.DevelopmentTask{
		PhaseID:  phase.ID,
		TaskName: "Task 2 Working",
		Status:   model.StatusWorking,
		Progress: 60.0,
	}
	db.Create(&task1)
	db.Create(&task2)

	err := repo.ReconcileAllPhases()
	if err != nil {
		t.Fatalf("ReconcileAllPhases failed: %v", err)
	}

	var fetchedPhase model.DevelopmentPhase
	db.First(&fetchedPhase, phase.ID)
	if fetchedPhase.Status != model.PhaseWorking {
		t.Errorf("Expected phase status WORKING, got %s", fetchedPhase.Status)
	}
	expectedAvg := 80.0 // (100 + 60) / 2
	if fetchedPhase.Progress != expectedAvg {
		t.Errorf("Expected progress %.1f%%, got %.1f%%", expectedAvg, fetchedPhase.Progress)
	}
	if fetchedPhase.CompletedDate != nil {
		t.Errorf("Expected CompletedDate to be nil while working")
	}
}

// 3. A required task BLOCKED or FAILED -> parent does not become DONE
func TestReconciliation_RequiredTaskBlockedOrFailed_ParentDoesNotBecomeDone(t *testing.T) {
	db := setupReconciliationTestDB(t)
	repo := repository.NewPhaseRepository(db)

	phase := model.DevelopmentPhase{
		PhaseNumber: 5,
		Name:        "Phase 5 — Storage",
		Status:      model.PhasePending,
		Progress:    0.0,
		OrderIndex:  5,
	}
	db.Create(&phase)

	taskDone := model.DevelopmentTask{
		PhaseID:  phase.ID,
		TaskName: "Task 1 Done",
		Status:   model.StatusDone,
		Progress: 100.0,
	}
	taskBlocked := model.DevelopmentTask{
		PhaseID:  phase.ID,
		TaskName: "Task 2 Blocked",
		Status:   model.StatusBlocked,
		Progress: 0.0,
	}
	db.Create(&taskDone)
	db.Create(&taskBlocked)

	err := repo.ReconcileAllPhases()
	if err != nil {
		t.Fatalf("ReconcileAllPhases failed: %v", err)
	}

	var fetchedPhase model.DevelopmentPhase
	db.First(&fetchedPhase, phase.ID)
	if fetchedPhase.Status == model.PhaseCompleted {
		t.Errorf("Phase must NOT become COMPLETED when a required task is BLOCKED")
	}
	if fetchedPhase.Status != model.PhaseWorking {
		t.Errorf("Expected phase status WORKING due to blocker, got %s", fetchedPhase.Status)
	}

	// Now test with FAILED status
	db.Model(&taskBlocked).Update("status", model.StatusFailed)
	_ = repo.ReconcileAllPhases()
	db.First(&fetchedPhase, phase.ID)
	if fetchedPhase.Status == model.PhaseCompleted {
		t.Errorf("Phase must NOT become COMPLETED when a required task is FAILED")
	}

	// Now test with WAITING_APPROVAL status
	db.Model(&taskBlocked).Update("status", model.StatusWaitingApproval)
	_ = repo.ReconcileAllPhases()
	db.First(&fetchedPhase, phase.ID)
	if fetchedPhase.Status == model.PhaseCompleted {
		t.Errorf("Phase must NOT become COMPLETED when a required task is WAITING_APPROVAL")
	}
}

// 4. PLANNED and SUPERSEDED tasks do not block completion
func TestReconciliation_PlannedAndSupersededTasks_DoNotBlockCompletion(t *testing.T) {
	db := setupReconciliationTestDB(t)
	repo := repository.NewPhaseRepository(db)

	phase := model.DevelopmentPhase{
		PhaseNumber: 4,
		Name:        "Phase 4 — Reliability & Storage",
		Status:      model.PhaseWorking,
		OrderIndex:  4,
	}
	db.Create(&phase)

	subActive := model.DevelopmentSubphase{
		PhaseID:    phase.ID,
		Name:       "Phase 4.1 — Active Subphase",
		Status:     "WORKING",
		OrderIndex: 1,
	}
	subPlanned := model.DevelopmentSubphase{
		PhaseID:    phase.ID,
		Name:       "Phase 4.X — Future Planned Subphase",
		Status:     "PLANNED",
		OrderIndex: 2,
	}
	db.Create(&subActive)
	db.Create(&subPlanned)

	// Active required task DONE (100%)
	tActive := model.DevelopmentTask{
		PhaseID:    phase.ID,
		SubphaseID: &subActive.ID,
		TaskName:   "Active Delivered Task",
		Status:     model.StatusDone,
		Progress:   100.0,
	}
	// Superseded legacy task (must be excluded from active denominator and completion checks)
	tSuperseded := model.DevelopmentTask{
		PhaseID:    phase.ID,
		SubphaseID: &subActive.ID,
		TaskName:   "Legacy Deprecated Task",
		Status:     model.StatusSuperseded,
		Progress:   0.0,
	}
	// Planned future task (must be excluded from blocking active completion)
	tPlanned := model.DevelopmentTask{
		PhaseID:    phase.ID,
		SubphaseID: &subPlanned.ID,
		TaskName:   "Planned Future Task",
		Status:     model.StatusPlanned,
		Progress:   0.0,
	}
	db.Create(&tActive)
	db.Create(&tSuperseded)
	db.Create(&tPlanned)

	err := repo.ReconcileAllPhases()
	if err != nil {
		t.Fatalf("ReconcileAllPhases failed: %v", err)
	}

	var fetchedPhase model.DevelopmentPhase
	db.First(&fetchedPhase, phase.ID)
	if fetchedPhase.Status != model.PhaseCompleted {
		t.Errorf("Expected parent phase COMPLETED when all active tasks are DONE (planned/superseded ignored), got %s", fetchedPhase.Status)
	}
	if fetchedPhase.Progress != 100.0 {
		t.Errorf("Expected progress 100%%, got %.1f%%", fetchedPhase.Progress)
	}

	var fetchedActiveSub model.DevelopmentSubphase
	db.First(&fetchedActiveSub, subActive.ID)
	if fetchedActiveSub.Status != "DONE" || fetchedActiveSub.Progress != 100.0 {
		t.Errorf("Expected active subphase DONE at 100%%, got %s (%.1f%%)", fetchedActiveSub.Status, fetchedActiveSub.Progress)
	}
}

// 5. Repeated reconciliation produces no duplicate tasks or unintended changes (idempotency)
func TestReconciliation_RepeatedReconciliation_IsIdempotent(t *testing.T) {
	db := setupReconciledDB(t)
	repo := repository.NewPhaseRepository(db)

	var initialTaskCount, initialPhaseCount, initialSubphaseCount int64
	db.Model(&model.DevelopmentTask{}).Count(&initialTaskCount)
	db.Model(&model.DevelopmentPhase{}).Count(&initialPhaseCount)
	db.Model(&model.DevelopmentSubphase{}).Count(&initialSubphaseCount)

	// Run reconciliation 5 times consecutively
	for i := 0; i < 5; i++ {
		if err := repo.ReconcileAllPhases(); err != nil {
			t.Fatalf("ReconcileAllPhases iteration %d failed: %v", i+1, err)
		}
	}

	var postTaskCount, postPhaseCount, postSubphaseCount int64
	db.Model(&model.DevelopmentTask{}).Count(&postTaskCount)
	db.Model(&model.DevelopmentPhase{}).Count(&postPhaseCount)
	db.Model(&model.DevelopmentSubphase{}).Count(&postSubphaseCount)

	if postTaskCount != initialTaskCount {
		t.Errorf("Task count changed after repeated reconciliation! Initial: %d, Post: %d", initialTaskCount, postTaskCount)
	}
	if postPhaseCount != initialPhaseCount {
		t.Errorf("Phase count changed after repeated reconciliation! Initial: %d, Post: %d", initialPhaseCount, postPhaseCount)
	}
	if postSubphaseCount != initialSubphaseCount {
		t.Errorf("Subphase count changed after repeated reconciliation! Initial: %d, Post: %d", initialSubphaseCount, postSubphaseCount)
	}

	// Verify Phase 4 remains COMPLETED at 100%
	var phase4 model.DevelopmentPhase
	db.Where("phase_number = 4").First(&phase4)
	if phase4.Status != model.PhaseCompleted || phase4.Progress != 100.0 {
		t.Errorf("Phase 4 state drifted after repeated reconciliation: status=%s, progress=%.1f", phase4.Status, phase4.Progress)
	}
}

// 6. Existing Phase 1–3 status and progress remain correct
func TestReconciliation_ExistingPhase1To3_StatusAndProgressRemainCorrect(t *testing.T) {
	db := setupReconciledDB(t)
	phaseRepo := repository.NewPhaseRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	phaseService := service.NewPhaseService(phaseRepo, systemRepo)

	phases, err := phaseService.GetAllPhases()
	if err != nil {
		t.Fatalf("GetAllPhases failed: %v", err)
	}

	phaseMap := make(map[int]model.DevelopmentPhase)
	for _, p := range phases {
		phaseMap[p.PhaseNumber] = p
	}

	// Phase 1: Foundation (COMPLETED, 100%)
	p1, ok := phaseMap[1]
	if !ok {
		t.Fatalf("Phase 1 not found")
	}
	if p1.Status != model.PhaseCompleted || p1.Progress != 100.0 {
		t.Errorf("Phase 1 incorrect: status=%s, progress=%.1f%% (expected COMPLETED, 100%%)", p1.Status, p1.Progress)
	}

	// Phase 2: Device & Communication (COMPLETED, 100%)
	p2, ok := phaseMap[2]
	if !ok {
		t.Fatalf("Phase 2 not found")
	}
	if p2.Status != model.PhaseCompleted || p2.Progress != 100.0 {
		t.Errorf("Phase 2 incorrect: status=%s, progress=%.1f%% (expected COMPLETED, 100%%)", p2.Status, p2.Progress)
	}

	// Phase 3: Data Engine (COMPLETED, 100%)
	p3, ok := phaseMap[3]
	if !ok {
		t.Fatalf("Phase 3 not found")
	}
	if p3.Status != model.PhaseCompleted || p3.Progress != 100.0 {
		t.Errorf("Phase 3 incorrect: status=%s, progress=%.1f%% (expected COMPLETED, 100%%)", p3.Status, p3.Progress)
	}

	// Phase 4: Reliability & Storage (COMPLETED, 100%)
	p4, ok := phaseMap[4]
	if !ok {
		t.Fatalf("Phase 4 not found")
	}
	if p4.Status != model.PhaseCompleted || p4.Progress != 100.0 {
		t.Errorf("Phase 4 incorrect: status=%s, progress=%.1f%% (expected COMPLETED, 100%%)", p4.Status, p4.Progress)
	}

	// Verify Phase 4 has exactly 60 active DONE tasks across subphases 4.1, 4.2, 4.3, 4.4
	var activePhase4DoneCount int64
	db.Model(&model.DevelopmentTask{}).
		Where("phase_id = ? AND status = ? AND status != ? AND status != ?", p4.ID, model.StatusDone, model.StatusSuperseded, model.StatusPlanned).
		Count(&activePhase4DoneCount)
	if activePhase4DoneCount != 60 {
		t.Errorf("Expected 60 active DONE tasks for Phase 4, got %d", activePhase4DoneCount)
	}
}
