package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/database"
	"datalogger/internal/handler"
	"datalogger/internal/middleware"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

type PackageCTestEnv struct {
	Router        *gin.Engine
	DB            *gorm.DB
	AdminToken    string
	OperatorToken string
	EngineerToken string
	AdminUser     *model.User
	OperatorUser  *model.User
	EngineerUser  *model.User
	SystemService *service.SystemService
	ReportService *service.ReportService
	AuthService   *service.AuthService
}

func setupPackageCTestEnv(t *testing.T) *PackageCTestEnv {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_pkg_c_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open SQLite: %v", err)
	}

	_ = database.RunMigrations(db)
	_ = database.Seed(db)

	cfg := config.DefaultConfig()

	// Repositories
	systemRepo := repository.NewSystemRepository(db)
	alarmRepo := repository.NewAlarmRepository(db)
	telemetryRepo := repository.NewTelemetryRepository(db)

	// Services
	alarmService := service.NewAlarmService(alarmRepo, systemRepo)
	systemService := service.NewSystemService(systemRepo, nil, cfg)
	systemService.SetAlarmService(alarmService)
	authService := service.NewAuthService(systemRepo, cfg)
	middleware.SetPermissionChecker(authService.HasPermission)
	reportService := service.NewReportService(db, systemRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authService)
	systemHandler := handler.NewSystemHandler(systemService)
	reportHandler := handler.NewReportHandler(reportService)
	userHandler := handler.NewUserHandler(authService)
	roleHandler := handler.NewRoleHandler(authService)

	// Setup Gin Router with identical routes to main.go
	router := gin.New()
	router.Use(middleware.CORS())

	api := router.Group("/api")
	{
		api.POST("/auth/login", authHandler.Login)

		// Reports (Task #79)
		reportsGroup := api.Group("/reports")
		reportsGroup.Use(middleware.JWTAuth(authService))
		{
			reportsGroup.GET("/preview", middleware.RequirePermission("report.view"), reportHandler.PreviewReport)
			reportsGroup.POST("/preview", middleware.RequirePermission("report.view"), reportHandler.PreviewReport)
			reportsGroup.GET("/generate", middleware.RequirePermission("report.generate"), reportHandler.GenerateReport)
			reportsGroup.POST("/generate", middleware.RequirePermission("report.generate"), reportHandler.GenerateReport)
		}

		// Users (Task #80)
		usersGroup := api.Group("/users")
		usersGroup.Use(middleware.JWTAuth(authService))
		{
			usersGroup.GET("", middleware.RequirePermission("user.view"), userHandler.ListUsers)
			usersGroup.GET("/:id", middleware.RequirePermission("user.view"), userHandler.GetUser)
			usersGroup.POST("", middleware.RequirePermission("user.manage"), userHandler.CreateUser)
			usersGroup.PUT("/:id", middleware.RequirePermission("user.manage"), userHandler.UpdateUser)
			usersGroup.PUT("/:id/enable", middleware.RequirePermission("user.manage"), userHandler.ToggleUserActive)
			usersGroup.PUT("/:id/password", middleware.RequirePermission("user.manage"), userHandler.ResetPassword)
			usersGroup.DELETE("/:id", middleware.RequirePermission("user.manage"), userHandler.DeleteUser)
		}

		// Roles & Permissions (Tasks #81, #82)
		rolesGroup := api.Group("/roles")
		rolesGroup.Use(middleware.JWTAuth(authService))
		{
			rolesGroup.GET("", middleware.RequirePermission("role.view"), roleHandler.ListRoles)
			rolesGroup.GET("/:id", middleware.RequirePermission("role.view"), roleHandler.GetRole)
			rolesGroup.POST("", middleware.RequirePermission("role.manage"), roleHandler.CreateRole)
			rolesGroup.PUT("/:id", middleware.RequirePermission("role.manage"), roleHandler.UpdateRole)
			rolesGroup.PUT("/:id/permissions", middleware.RequirePermission("role.manage"), roleHandler.UpdateRolePermissions)
			rolesGroup.DELETE("/:id", middleware.RequirePermission("role.manage"), roleHandler.DeleteRole)
		}

		permissionsGroup := api.Group("/permissions")
		permissionsGroup.Use(middleware.JWTAuth(authService))
		{
			permissionsGroup.GET("", middleware.RequirePermission("role.view"), roleHandler.ListPermissions)
		}

		// System Configuration (Task #83)
		configGroup := api.Group("/system/config")
		configGroup.Use(middleware.JWTAuth(authService))
		{
			configGroup.GET("", middleware.RequirePermission("config.view"), systemHandler.GetConfig)
			configGroup.PUT("", middleware.RequirePermission("config.manage"), systemHandler.SaveConfig)
			configGroup.POST("", middleware.RequirePermission("config.manage"), systemHandler.SaveConfig)
		}
	}

	// Create test devices & parameters & telemetry
	var dev model.Device
	if err := db.Where("code = ?", "DEV-01").First(&dev).Error; err != nil {
		dev = model.Device{
			Code:             "DEV-01",
			Name:             "Main Edge Ingestor",
			Status:           model.DeviceStatusActive,
			ConnectionStatus: model.DeviceConnOnline,
			SuccessCount:     150,
			FailedCount:      2,
			LatencyMs:        12,
		}
		db.Create(&dev)
	}

	var param model.Parameter
	if err := db.Where("code = ? AND device_id = ?", "TEMP_01", dev.ID).First(&param).Error; err != nil {
		currVal := 25.4
		param = model.Parameter{
			DeviceID:     dev.ID,
			Code:         "TEMP_01",
			Name:         "Temperature Sensor",
			Unit:         "°C",
			CurrentValue: &currVal,
		}
		db.Create(&param)
	}

	// Insert telemetry data into raw_data
	now := time.Now()
	rawList := []*model.RawData{
		{DeviceID: dev.ID, ParameterID: param.ID, Value: 24.0, ValueNumeric: func() *float64 { v := 24.0; return &v }(), Quality: "GOOD", ReceivedAt: now.Add(-10 * time.Minute)},
		{DeviceID: dev.ID, ParameterID: param.ID, Value: 26.0, ValueNumeric: func() *float64 { v := 26.0; return &v }(), Quality: "GOOD", ReceivedAt: now.Add(-5 * time.Minute)},
		{DeviceID: dev.ID, ParameterID: param.ID, Value: 30.0, ValueNumeric: func() *float64 { v := 30.0; return &v }(), Quality: "UNCERTAIN", ReceivedAt: now.Add(-1 * time.Minute)},
	}
	_ = telemetryRepo.SaveBatch(context.Background(), rawList)

	// Insert test alarm
	devID := dev.ID
	paramID := param.ID
	testAlarm := model.Alarm{
		DeviceID:    &devID,
		ParameterID: &paramID,
		AlarmCode:   "HIGH_TEMP",
		Severity:    model.SeverityCritical,
		Status:      model.AlarmActive,
		Message:     "Temperature exceeded 28°C",
		TriggeredAt: now.Add(-2 * time.Minute),
	}
	db.Create(&testAlarm)

	// Create test users
	var adminRole, engRole, opRole model.Role
	db.Where("name = ? OR name = ?", "Administrator", "SUPERADMIN").First(&adminRole)
	db.Where("name = ? OR name = ?", "Engineer", "ENGINEER").First(&engRole)
	db.Where("name = ? OR name = ?", "Operator", "OPERATOR").First(&opRole)

	passHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	var adminUser model.User
	if err := db.Where("username = ?", "admin").First(&adminUser).Error; err != nil {
		adminUser = model.User{
			Username: "admin",
			Password: string(passHash),
			FullName: "Admin User",
			RoleID:   adminRole.ID,
			IsActive: true,
		}
		db.Create(&adminUser)
	}

	engUser := model.User{
		Username: "engineer_c",
		Email:    "engineer_c@plant.local",
		Password: string(passHash),
		FullName: "Engineer C",
		RoleID:   engRole.ID,
		IsActive: true,
	}
	if err := db.Create(&engUser).Error; err != nil {
		t.Fatalf("Failed to create engUser: %v", err)
	}

	opUser := model.User{
		Username: "operator_c",
		Email:    "operator_c@plant.local",
		Password: string(passHash),
		FullName: "Operator C",
		RoleID:   opRole.ID,
		IsActive: true,
	}
	if err := db.Create(&opUser).Error; err != nil {
		t.Fatalf("Failed to create opUser: %v", err)
	}

	// Generate tokens
	adminToken, _, err := authService.Login("admin", "admin123", "127.0.0.1", "test-agent")
	if adminToken == "" {
		adminToken, _, _ = authService.Login("admin", "password123", "127.0.0.1", "test-agent")
	}
	engToken, _, err := authService.Login("engineer_c", "password123", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to login engineer_c: %v", err)
	}
	opToken, _, err := authService.Login("operator_c", "password123", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Failed to login operator_c: %v", err)
	}

	return &PackageCTestEnv{
		Router:        router,
		DB:            db,
		AdminToken:    adminToken,
		OperatorToken: opToken,
		EngineerToken: engToken,
		AdminUser:     &adminUser,
		OperatorUser:  &opUser,
		EngineerUser:  &engUser,
		SystemService: systemService,
		ReportService: reportService,
		AuthService:   authService,
	}
}

// -------------------------------------------------------------
// PACKAGE C1: REPORTING (#79) TESTS
// -------------------------------------------------------------

func TestPackageC1_ReportPreviewJSON(t *testing.T) {
	env := setupPackageCTestEnv(t)

	req, _ := http.NewRequest(http.MethodGet, "/api/reports/preview?report_type=COMPREHENSIVE", nil)
	req.Header.Set("Authorization", "Bearer "+env.AdminToken)
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for report preview, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool               `json:"success"`
		Data    service.ReportData `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("Failed to parse report preview JSON: %v", err)
	}

	if !resp.Success {
		t.Errorf("Expected success = true")
	}
	if resp.Data.Title == "" {
		t.Errorf("Expected non-empty report title")
	}
	if len(resp.Data.DeviceSummaries) == 0 {
		t.Errorf("Expected at least 1 device summary, got 0")
	}
	if len(resp.Data.ParameterSummaries) == 0 {
		t.Errorf("Expected parameter summaries, got 0")
	}

	// Verify calculated statistics
	foundParam := false
	for _, p := range resp.Data.ParameterSummaries {
		if p.ParameterCode == "TEMP_01" {
			foundParam = true
			if p.SampleCount < 3 {
				t.Errorf("Expected at least 3 samples for TEMP_01, got %d", p.SampleCount)
			}
			if p.MinValue == nil || *p.MinValue != 24.0 {
				t.Errorf("Expected min value 24.0, got %v", p.MinValue)
			}
			if p.MaxValue == nil || *p.MaxValue != 30.0 {
				t.Errorf("Expected max value 30.0, got %v", p.MaxValue)
			}
			if p.AvgValue == nil || *p.AvgValue < 25.0 || *p.AvgValue > 28.0 {
				t.Errorf("Expected average value around 26.67, got %v", p.AvgValue)
			}
			if p.GoodCount < 2 {
				t.Errorf("Expected at least 2 good samples, got %d", p.GoodCount)
			}
		}
	}
	if !foundParam {
		t.Errorf("Expected to find TEMP_01 parameter in summary")
	}

	// Verify alarm summary
	if resp.Data.AlarmSummary.TotalAlarms == 0 {
		t.Errorf("Expected at least 1 alarm in summary, got 0")
	}
	if resp.Data.AlarmSummary.CriticalCount == 0 {
		t.Errorf("Expected critical alarm count > 0")
	}
}

func TestPackageC1_ReportGenerateHTMLAndCSV(t *testing.T) {
	env := setupPackageCTestEnv(t)

	// 1. HTML Output
	htmlReq, _ := http.NewRequest(http.MethodGet, "/api/reports/generate?format=html", nil)
	htmlReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wHTML := httptest.NewRecorder()
	env.Router.ServeHTTP(wHTML, htmlReq)

	if wHTML.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for HTML report, got %d: %s", wHTML.Code, wHTML.Body.String())
	}
	ct := wHTML.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Expected text/html content type, got %s", ct)
	}
	htmlBody := wHTML.Body.String()
	if !strings.Contains(htmlBody, "<!DOCTYPE html>") || !strings.Contains(htmlBody, "Operational Telemetry &amp; Performance Report") && !strings.Contains(htmlBody, "Operational Telemetry & Performance Report") {
		t.Errorf("HTML report does not contain expected header")
	}
	if !strings.Contains(htmlBody, "@media print") {
		t.Errorf("HTML report missing @media print styles for PDF printing")
	}

	// 2. CSV Output
	csvReq, _ := http.NewRequest(http.MethodGet, "/api/reports/generate?format=csv", nil)
	csvReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wCSV := httptest.NewRecorder()
	env.Router.ServeHTTP(wCSV, csvReq)

	if wCSV.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for CSV report, got %d: %s", wCSV.Code, wCSV.Body.String())
	}
	csvCt := wCSV.Header().Get("Content-Type")
	if !strings.Contains(csvCt, "text/csv") {
		t.Errorf("Expected text/csv content type, got %s", csvCt)
	}
	csvBody := wCSV.Body.String()
	if !strings.Contains(csvBody, "# REPORT METADATA") || !strings.Contains(csvBody, "# PARAMETER TELEMETRY STATISTICS") {
		t.Errorf("CSV report missing standard section headers: %s", csvBody[:200])
	}
}

// -------------------------------------------------------------
// PACKAGE C2: USER MANAGEMENT (#80) TESTS
// -------------------------------------------------------------

func TestPackageC2_UserListingPaginationAndFilter(t *testing.T) {
	env := setupPackageCTestEnv(t)

	// List users
	req, _ := http.NewRequest(http.MethodGet, "/api/users?page=1&page_size=10", nil)
	req.Header.Set("Authorization", "Bearer "+env.AdminToken)
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for list users, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool `json:"success"`
		Data    struct {
			Users []model.User `json:"users"`
			Total int64        `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if resp.Data.Total < 3 {
		t.Errorf("Expected at least 3 users, got %d", resp.Data.Total)
	}

	// Verify no passwords are exposed in the JSON response
	for _, u := range resp.Data.Users {
		if u.Password != "" {
			t.Errorf("Security violation: password hash leaked in User list JSON for user %s!", u.Username)
		}
	}
}

func TestPackageC2_UserCreateUpdateTogglePasswordAndAudit(t *testing.T) {
	env := setupPackageCTestEnv(t)

	var opRole model.Role
	env.DB.Where("name = ? OR name = ?", "Operator", "OPERATOR").First(&opRole)

	// 1. Create User
	createPayload := map[string]interface{}{
		"username":  "new_operator",
		"email":     "new_op@plant.local",
		"full_name": "New Plant Operator",
		"password":  "securepass123",
		"role_id":   opRole.ID,
		"is_active": true,
	}
	bodyBytes, _ := json.Marshal(createPayload)
	createReq, _ := http.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(bodyBytes))
	createReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	createReq.Header.Set("Content-Type", "application/json")
	wCreate := httptest.NewRecorder()
	env.Router.ServeHTTP(wCreate, createReq)

	if wCreate.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for user create, got %d: %s", wCreate.Code, wCreate.Body.String())
	}

	var createdUser model.User
	var createResp struct {
		Success bool       `json:"success"`
		Data    model.User `json:"data"`
	}
	_ = json.Unmarshal(wCreate.Body.Bytes(), &createResp)
	createdUser = createResp.Data

	if createdUser.ID == 0 || createdUser.Username != "new_operator" {
		t.Fatalf("User was not created correctly: %+v", createdUser)
	}
	if createdUser.Password != "" {
		t.Errorf("Password must not be returned in API response!")
	}

	// 2. Update User Profile
	updatePayload := map[string]interface{}{
		"full_name": "Updated Plant Operator Name",
		"email":     "updated_op@plant.local",
	}
	upBytes, _ := json.Marshal(updatePayload)
	upReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d", createdUser.ID), bytes.NewBuffer(upBytes))
	upReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	upReq.Header.Set("Content-Type", "application/json")
	wUp := httptest.NewRecorder()
	env.Router.ServeHTTP(wUp, upReq)

	if wUp.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for user update, got %d: %s", wUp.Code, wUp.Body.String())
	}

	// 3. Toggle Status (Disable user)
	disPayload := map[string]interface{}{"is_active": false}
	disBytes, _ := json.Marshal(disPayload)
	disReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/enable", createdUser.ID), bytes.NewBuffer(disBytes))
	disReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	disReq.Header.Set("Content-Type", "application/json")
	wDis := httptest.NewRecorder()
	env.Router.ServeHTTP(wDis, disReq)

	if wDis.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for user toggle, got %d: %s", wDis.Code, wDis.Body.String())
	}

	// Verify user cannot log in when deactivated
	_, _, err := env.AuthService.Login("new_operator", "securepass123", "127.0.0.1", "test")
	if err == nil || !strings.Contains(err.Error(), "deactivated") {
		t.Errorf("Expected login failure for deactivated user, got: %v", err)
	}

	// 4. Reset Password
	pwdPayload := map[string]interface{}{"password": "newpassword456"}
	pwdBytes, _ := json.Marshal(pwdPayload)
	pwdReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/password", createdUser.ID), bytes.NewBuffer(pwdBytes))
	pwdReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	pwdReq.Header.Set("Content-Type", "application/json")
	wPwd := httptest.NewRecorder()
	env.Router.ServeHTTP(wPwd, pwdReq)

	if wPwd.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for password reset, got %d: %s", wPwd.Code, wPwd.Body.String())
	}

	// 5. Verify Audit Trail Entries
	var trails []model.AuditTrail
	env.DB.Where("resource = ?", "USER").Find(&trails)
	if len(trails) < 4 {
		t.Errorf("Expected at least 4 audit trail logs for USER operations, got %d", len(trails))
	}
}

func TestPackageC2_LastAdministratorProtection(t *testing.T) {
	env := setupPackageCTestEnv(t)

	// Count admins
	var adminRole model.Role
	env.DB.Where("name = ? OR name = ?", "Administrator", "SUPERADMIN").First(&adminRole)

	// Ensure there is only 1 admin
	env.DB.Where("role_id = ? AND username != ?", adminRole.ID, "admin").Delete(&model.User{})

	var soleAdmin model.User
	env.DB.Where("username = ?", "admin").First(&soleAdmin)

	// 1. Try to deactivate the sole administrator
	disPayload := map[string]interface{}{"is_active": false}
	disBytes, _ := json.Marshal(disPayload)
	disReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/users/%d/enable", soleAdmin.ID), bytes.NewBuffer(disBytes))
	disReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	disReq.Header.Set("Content-Type", "application/json")
	wDis := httptest.NewRecorder()
	env.Router.ServeHTTP(wDis, disReq)

	if wDis.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request when trying to deactivate sole admin, got %d: %s", wDis.Code, wDis.Body.String())
	}

	// 2. Try to delete the sole administrator
	delReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/users/%d", soleAdmin.ID), nil)
	delReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wDel := httptest.NewRecorder()
	env.Router.ServeHTTP(wDel, delReq)

	if wDel.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request when trying to delete sole admin, got %d: %s", wDel.Code, wDel.Body.String())
	}
}

// -------------------------------------------------------------
// PACKAGE C3: ROLE & PERMISSION MANAGEMENT (#81, #82) TESTS
// -------------------------------------------------------------

func TestPackageC3_RoleAndPermissionMatrixManagement(t *testing.T) {
	env := setupPackageCTestEnv(t)

	// 1. List permissions
	permReq, _ := http.NewRequest(http.MethodGet, "/api/permissions", nil)
	permReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wPerm := httptest.NewRecorder()
	env.Router.ServeHTTP(wPerm, permReq)

	if wPerm.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for permissions, got %d: %s", wPerm.Code, wPerm.Body.String())
	}

	var permResp struct {
		Success bool `json:"success"`
		Data    struct {
			Permissions []model.Permission `json:"permissions"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wPerm.Body.Bytes(), &permResp)

	if len(permResp.Data.Permissions) < 10 {
		t.Errorf("Expected at least 10 permissions, got %d", len(permResp.Data.Permissions))
	}

	// 2. Create Custom Role (Auditor)
	var reportPerm model.Permission
	env.DB.Where("code = ?", "report.view").First(&reportPerm)

	rolePayload := map[string]interface{}{
		"name":           "Auditor",
		"description":    "Quality auditor with read-only access to operational reports",
		"permission_ids": []uint{reportPerm.ID},
	}
	roleBytes, _ := json.Marshal(rolePayload)
	roleReq, _ := http.NewRequest(http.MethodPost, "/api/roles", bytes.NewBuffer(roleBytes))
	roleReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	roleReq.Header.Set("Content-Type", "application/json")
	wRole := httptest.NewRecorder()
	env.Router.ServeHTTP(wRole, roleReq)

	if wRole.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for role create, got %d: %s", wRole.Code, wRole.Body.String())
	}

	var createdRole model.Role
	var roleResp struct {
		Success bool       `json:"success"`
		Data    model.Role `json:"data"`
	}
	_ = json.Unmarshal(wRole.Body.Bytes(), &roleResp)
	createdRole = roleResp.Data

	if createdRole.ID == 0 || createdRole.Name != "Auditor" {
		t.Fatalf("Role not created as expected: %+v", createdRole)
	}

	// 3. Update Role Permissions (Add config.view)
	var configPerm model.Permission
	env.DB.Where("code = ?", "config.view").First(&configPerm)

	updatePermPayload := map[string]interface{}{
		"permission_ids": []uint{reportPerm.ID, configPerm.ID},
	}
	upPermBytes, _ := json.Marshal(updatePermPayload)
	upPermReq, _ := http.NewRequest(http.MethodPut, fmt.Sprintf("/api/roles/%d/permissions", createdRole.ID), bytes.NewBuffer(upPermBytes))
	upPermReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	upPermReq.Header.Set("Content-Type", "application/json")
	wUpPerm := httptest.NewRecorder()
	env.Router.ServeHTTP(wUpPerm, upPermReq)

	if wUpPerm.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for role permissions update, got %d: %s", wUpPerm.Code, wUpPerm.Body.String())
	}

	// 4. Verify RBAC checking works for the new role
	if !env.AuthService.HasPermission("Auditor", "report.view") {
		t.Errorf("Expected Auditor to have report.view")
	}
	if !env.AuthService.HasPermission("Auditor", "config.view") {
		t.Errorf("Expected Auditor to have config.view")
	}
	if env.AuthService.HasPermission("Auditor", "user.manage") {
		t.Errorf("Auditor must NOT have user.manage")
	}

	// 5. Delete Custom Role
	delRoleReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/roles/%d", createdRole.ID), nil)
	delRoleReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wDelRole := httptest.NewRecorder()
	env.Router.ServeHTTP(wDelRole, delRoleReq)

	if wDelRole.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for role delete, got %d: %s", wDelRole.Code, wDelRole.Body.String())
	}

	// 6. Protection: Cannot delete system Administrator role
	var adminRole model.Role
	env.DB.Where("name = ? OR name = ?", "Administrator", "SUPERADMIN").First(&adminRole)
	delAdminReq, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/api/roles/%d", adminRole.ID), nil)
	delAdminReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wDelAdmin := httptest.NewRecorder()
	env.Router.ServeHTTP(wDelAdmin, delAdminReq)

	if wDelAdmin.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request when deleting system Administrator role, got %d: %s", wDelAdmin.Code, wDelAdmin.Body.String())
	}
}

func TestPackageC3_RBACEnforcementAccessAllowedAndDenied(t *testing.T) {
	env := setupPackageCTestEnv(t)

	// Operator tries to create a user (requires user.manage) -> MUST BE 403 FORBIDDEN
	createPayload := map[string]interface{}{
		"username": "unauthorized_user",
		"password": "password123",
		"role_id":  1,
	}
	b, _ := json.Marshal(createPayload)
	reqDenied, _ := http.NewRequest(http.MethodPost, "/api/users", bytes.NewBuffer(b))
	reqDenied.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	reqDenied.Header.Set("Content-Type", "application/json")
	wDenied := httptest.NewRecorder()
	env.Router.ServeHTTP(wDenied, reqDenied)

	if wDenied.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for operator creating user, got %d: %s", wDenied.Code, wDenied.Body.String())
	}

	// Operator tries to view reports (has report.view) -> MUST BE 200 OK
	reqAllowed, _ := http.NewRequest(http.MethodGet, "/api/reports/preview", nil)
	reqAllowed.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	wAllowed := httptest.NewRecorder()
	env.Router.ServeHTTP(wAllowed, reqAllowed)

	if wAllowed.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for operator viewing reports, got %d: %s", wAllowed.Code, wAllowed.Body.String())
	}

	// Operator tries to generate reports (report.generate) -> MUST BE 403 Forbidden
	reqGenDenied, _ := http.NewRequest(http.MethodGet, "/api/reports/generate?format=html", nil)
	reqGenDenied.Header.Set("Authorization", "Bearer "+env.OperatorToken)
	wGenDenied := httptest.NewRecorder()
	env.Router.ServeHTTP(wGenDenied, reqGenDenied)

	if wGenDenied.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for operator generating reports, got %d: %s", wGenDenied.Code, wGenDenied.Body.String())
	}

	// Admin generates reports using token query parameter -> MUST BE 200 OK
	reqGenQuery, _ := http.NewRequest(http.MethodGet, "/api/reports/generate?format=html&token="+env.AdminToken, nil)
	wGenQuery := httptest.NewRecorder()
	env.Router.ServeHTTP(wGenQuery, reqGenQuery)

	if wGenQuery.Code != http.StatusOK {
		t.Errorf("Expected 200 OK for admin generating reports via token query, got %d: %s", wGenQuery.Code, wGenQuery.Body.String())
	}
}

// -------------------------------------------------------------
// PACKAGE C4: SYSTEM CONFIGURATION (#83) TESTS
// -------------------------------------------------------------

func TestPackageC4_SystemConfigValidationAndRestartIndicator(t *testing.T) {
	env := setupPackageCTestEnv(t)

	originalCfg := *config.Get()
	t.Cleanup(func() {
		cfg := config.Get()
		*cfg = originalCfg
		_ = os.Remove("config.json")
	})

	// 1. Get masked configuration (config.view)
	getReq, _ := http.NewRequest(http.MethodGet, "/api/system/config", nil)
	getReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	wGet := httptest.NewRecorder()
	env.Router.ServeHTTP(wGet, getReq)

	if wGet.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for get config, got %d: %s", wGet.Code, wGet.Body.String())
	}

	var getResp struct {
		Success bool                    `json:"success"`
		Data    service.SystemConfigDTO `json:"data"`
	}
	_ = json.Unmarshal(wGet.Body.Bytes(), &getResp)

	if getResp.Data.DBPassword != "••••••••" {
		t.Errorf("Expected DB password to be masked, got %s", getResp.Data.DBPassword)
	}

	// 2. Reject invalid port (> 65535)
	badPortPayload := map[string]interface{}{
		"port": "999999",
	}
	bpBytes, _ := json.Marshal(badPortPayload)
	bpReq, _ := http.NewRequest(http.MethodPost, "/api/system/config", bytes.NewBuffer(bpBytes))
	bpReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	bpReq.Header.Set("Content-Type", "application/json")
	wBP := httptest.NewRecorder()
	env.Router.ServeHTTP(wBP, bpReq)

	if wBP.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for invalid port, got %d: %s", wBP.Code, wBP.Body.String())
	}

	// 3. Reject invalid baud rate
	badBaudPayload := map[string]interface{}{
		"serial_baud_rate": 1337,
	}
	bbBytes, _ := json.Marshal(badBaudPayload)
	bbReq, _ := http.NewRequest(http.MethodPost, "/api/system/config", bytes.NewBuffer(bbBytes))
	bbReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	bbReq.Header.Set("Content-Type", "application/json")
	wBB := httptest.NewRecorder()
	env.Router.ServeHTTP(wBB, bbReq)

	if wBB.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for invalid baud rate, got %d: %s", wBB.Code, wBB.Body.String())
	}

	// 4. Valid update with restart and acquisition disruption flags
	validPayload := map[string]interface{}{
		"db_port":          "3307",
		"serial_baud_rate": 115200,
		"log_level":        "debug",
	}
	vBytes, _ := json.Marshal(validPayload)
	vReq, _ := http.NewRequest(http.MethodPost, "/api/system/config", bytes.NewBuffer(vBytes))
	vReq.Header.Set("Authorization", "Bearer "+env.AdminToken)
	vReq.Header.Set("Content-Type", "application/json")
	wV := httptest.NewRecorder()
	env.Router.ServeHTTP(wV, vReq)

	if wV.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for valid config save, got %d: %s", wV.Code, wV.Body.String())
	}

	var saveResp struct {
		Success bool                     `json:"success"`
		Data    service.ConfigSaveResult `json:"data"`
	}
	_ = json.Unmarshal(wV.Body.Bytes(), &saveResp)

	if !saveResp.Data.Saved {
		t.Errorf("Expected saved = true")
	}
	if !saveResp.Data.RestartRequired {
		t.Errorf("Expected RestartRequired = true due to DB port change")
	}
	if !saveResp.Data.AcquisitionDisruption {
		t.Errorf("Expected AcquisitionDisruption = true due to Serial baud rate change")
	}

	// 5. Verify Audit Trail logged for CONFIG
	var cfgTrails []model.AuditTrail
	env.DB.Where("resource = ? AND action = ?", "CONFIG", "CONFIG_UPDATE").Find(&cfgTrails)
	if len(cfgTrails) == 0 {
		t.Errorf("Expected audit trail record for CONFIG_UPDATE")
	}
}
