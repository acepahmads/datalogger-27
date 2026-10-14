package service_test

import (
	"testing"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupCategoryTestDB(t *testing.T) (*gorm.DB, *service.CategoryService, *repository.CategoryRepository, *repository.SystemRepository) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	if err := db.AutoMigrate(
		&model.ParameterCategory{},
		&model.Device{},
		&model.Parameter{},
		&model.RawData{},
		&model.AuditTrail{},
		&model.User{},
		&model.Role{},
		&model.Permission{},
	); err != nil {
		t.Fatalf("AutoMigrate failed: %v", err)
	}

	catRepo := repository.NewCategoryRepository(db)
	sysRepo := repository.NewSystemRepository(db)
	catService := service.NewCategoryService(catRepo, sysRepo)

	return db, catService, catRepo, sysRepo
}

func TestCategoryService_CreateAndValidation(t *testing.T) {
	_, catService, _, sysRepo := setupCategoryTestDB(t)

	// Valid creation
	req := &service.CreateCategoryRequest{
		Code:              "gas_pollutants",
		Name:              "Gas & Pollutants",
		Description:       "Harmful air pollutants",
		IconKey:           "atom",
		ColorKey:          "purple",
		DisplayOrder:      1,
		DefaultColumnSpan: 2,
	}

	cat, err := catService.CreateCategory(req, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}
	if cat.ID == 0 || cat.Code != "gas_pollutants" || cat.ColorKey != "purple" || cat.DefaultColumnSpan != 2 {
		t.Errorf("Unexpected created category: %+v", cat)
	}

	// Verify Audit trail recorded
	trails, err := sysRepo.GetAuditTrails(10)
	if err != nil || len(trails) == 0 || trails[0].Action != "CATEGORY_CREATED" {
		t.Errorf("Expected CATEGORY_CREATED audit trail, got: %+v", trails)
	}

	// Duplicate code rejection
	_, err = catService.CreateCategory(req, "admin", "127.0.0.1", "TestAgent")
	if err == nil {
		t.Errorf("Expected error for duplicate category code, got nil")
	}

	// Invalid characters in code rejection
	badCodeReq := &service.CreateCategoryRequest{
		Code: "bad code with spaces!",
		Name: "Bad Code",
	}
	_, err = catService.CreateCategory(badCodeReq, "admin", "127.0.0.1", "TestAgent")
	if err == nil {
		t.Errorf("Expected error for invalid category code, got nil")
	}
}

func TestCategoryService_UpdateAndToggle(t *testing.T) {
	_, catService, _, sysRepo := setupCategoryTestDB(t)

	cat, err := catService.CreateCategory(&service.CreateCategoryRequest{
		Code: "meteorology",
		Name: "Weather",
	}, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	newName := "Weather & Meteorology"
	newColor := "cyan"
	newSpan := 3
	updated, err := catService.UpdateCategory(cat.ID, &service.UpdateCategoryRequest{
		Name:              &newName,
		ColorKey:          &newColor,
		DefaultColumnSpan: &newSpan,
	}, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to update category: %v", err)
	}
	if updated.Name != newName || updated.ColorKey != "cyan" || updated.DefaultColumnSpan != 3 {
		t.Errorf("Updated fields mismatch: %+v", updated)
	}

	// Toggle disable
	toggled, err := catService.ToggleCategory(cat.ID, false, "admin", "127.0.0.1", "TestAgent")
	if err != nil || toggled.Enabled != false {
		t.Errorf("Failed to toggle category: %v", err)
	}

	// Verify Audit trail
	trails, _ := sysRepo.GetAuditTrails(10)
	foundToggle := false
	for _, tr := range trails {
		if tr.Action == "CATEGORY_TOGGLED" {
			foundToggle = true
			break
		}
	}
	if !foundToggle {
		t.Errorf("Expected CATEGORY_TOGGLED audit log")
	}
}

func TestCategoryService_DeleteSafetyAndReassignment(t *testing.T) {
	db, catService, catRepo, _ := setupCategoryTestDB(t)

	// Create 2 categories
	cat1, _ := catService.CreateCategory(&service.CreateCategoryRequest{
		Code: "cat1",
		Name: "Category 1",
	}, "admin", "127.0.0.1", "TestAgent")

	cat2, _ := catService.CreateCategory(&service.CreateCategoryRequest{
		Code: "cat2",
		Name: "Category 2",
	}, "admin", "127.0.0.1", "TestAgent")

	// Create a dummy device and parameters assigned to cat1
	dev := model.Device{DeviceCode: "DEV-01", DeviceName: "Device 1", DeviceType: "MODBUS_RTU"}
	db.Create(&dev)

	param1 := model.Parameter{DeviceID: dev.ID, ParameterCode: "PM25", ParameterName: "PM2.5", CategoryID: &cat1.ID}
	param2 := model.Parameter{DeviceID: dev.ID, ParameterCode: "O3", ParameterName: "Ozone", CategoryID: &cat1.ID}
	db.Create(&param1)
	db.Create(&param2)

	// Create dummy telemetry for param1
	telemetry := model.RawData{DeviceID: dev.ID, ParameterID: param1.ID, Value: 35.5}
	db.Create(&telemetry)

	// Case A: Delete cat1 with reassignment to cat2
	err := catService.DeleteCategory(cat1.ID, &cat2.ID, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to delete category with reassignment: %v", err)
	}

	// Verify category 1 is deleted
	_, err = catRepo.GetByID(cat1.ID)
	if err == nil {
		t.Errorf("Expected category 1 to be deleted")
	}

	// Verify parameters are NOT deleted and are now in cat2!
	var recheckedParams []model.Parameter
	db.Find(&recheckedParams)
	if len(recheckedParams) != 2 {
		t.Fatalf("Expected 2 parameters to survive, got %d", len(recheckedParams))
	}
	for _, p := range recheckedParams {
		if p.CategoryID == nil || *p.CategoryID != cat2.ID {
			t.Errorf("Expected parameter %s to be reassigned to cat2 (%d), got: %v", p.ParameterCode, cat2.ID, p.CategoryID)
		}
	}

	// Verify telemetry was NOT deleted!
	var recheckedTelemetry []model.RawData
	db.Find(&recheckedTelemetry)
	if len(recheckedTelemetry) != 1 {
		t.Errorf("Expected telemetry to remain intact, found %d", len(recheckedTelemetry))
	}

	// Case B: Delete cat2 without reassignment (sets CategoryID to nil)
	err = catService.DeleteCategory(cat2.ID, nil, "admin", "127.0.0.1", "TestAgent")
	if err != nil {
		t.Fatalf("Failed to delete category 2: %v", err)
	}

	// Verify parameters survive as Uncategorized (nil)
	var survivingParams []model.Parameter
	db.Find(&survivingParams)
	if len(survivingParams) != 2 {
		t.Fatalf("Expected 2 parameters to survive, got %d", len(survivingParams))
	}
	for _, p := range survivingParams {
		if p.CategoryID != nil {
			t.Errorf("Expected parameter %s category to be nil (uncategorized), got: %v", p.ParameterCode, p.CategoryID)
		}
	}
}

func TestCategoryService_ReorderAndAssignParameters(t *testing.T) {
	db, catService, catRepo, _ := setupCategoryTestDB(t)

	c1, _ := catService.CreateCategory(&service.CreateCategoryRequest{Code: "c1", Name: "Cat 1", DisplayOrder: 1}, "admin", "127.0.0.1", "Test")
	c2, _ := catService.CreateCategory(&service.CreateCategoryRequest{Code: "c2", Name: "Cat 2", DisplayOrder: 2}, "admin", "127.0.0.1", "Test")

	// Reorder [c2, c1]
	err := catService.ReorderCategories([]uint{c2.ID, c1.ID}, "admin", "127.0.0.1", "Test")
	if err != nil {
		t.Fatalf("Reorder failed: %v", err)
	}

	list, _ := catRepo.List(false)
	if len(list) != 2 || list[0].ID != c2.ID || list[1].ID != c1.ID {
		t.Errorf("Reorder order incorrect: %+v", list)
	}

	// Assign parameters
	dev := model.Device{DeviceCode: "DEV-TEST", DeviceType: "MODBUS_TCP"}
	db.Create(&dev)
	p1 := model.Parameter{DeviceID: dev.ID, ParameterCode: "P1", ParameterName: "P1"}
	p2 := model.Parameter{DeviceID: dev.ID, ParameterCode: "P2", ParameterName: "P2"}
	db.Create(&p1)
	db.Create(&p2)

	err = catService.AssignParameters(&c1.ID, []uint{p1.ID, p2.ID}, "admin", "127.0.0.1", "Test")
	if err != nil {
		t.Fatalf("AssignParameters failed: %v", err)
	}

	var p1Check model.Parameter
	db.First(&p1Check, p1.ID)
	if p1Check.CategoryID == nil || *p1Check.CategoryID != c1.ID {
		t.Errorf("Expected p1 to be assigned to c1, got %v", p1Check.CategoryID)
	}
}
