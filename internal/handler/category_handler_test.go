package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

type CategoryTestEnv struct {
	Router     *gin.Engine
	DB         *gorm.DB
	AdminToken string
	EngToken   string
	OpToken    string
}

func setupCategoryTestRouter(t *testing.T) *CategoryTestEnv {
	gin.SetMode(gin.TestMode)

	dsn := fmt.Sprintf("file:mem_cat_api_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = database.RunMigrations(db)

	adminRole := model.Role{Name: "Administrator", Description: "Full access"}
	engRole := model.Role{Name: "Engineer", Description: "Engineering access"}
	opRole := model.Role{Name: "Operator", Description: "Read access"}
	db.Create(&adminRole)
	db.Create(&engRole)
	db.Create(&opRole)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	adminUser := model.User{Username: "cat_admin", Email: "admin@cat.local", Password: string(hash), RoleID: adminRole.ID, IsActive: true}
	engUser := model.User{Username: "cat_eng", Email: "eng@cat.local", Password: string(hash), RoleID: engRole.ID, IsActive: true}
	opUser := model.User{Username: "cat_op", Email: "op@cat.local", Password: string(hash), RoleID: opRole.ID, IsActive: true}
	db.Create(&adminUser)
	db.Create(&engUser)
	db.Create(&opUser)

	systemRepo := repository.NewSystemRepository(db)
	categoryRepo := repository.NewCategoryRepository(db)
	categoryService := service.NewCategoryService(categoryRepo, systemRepo)
	categoryHandler := handler.NewCategoryHandler(categoryService)

	cfg := config.DefaultConfig()
	authService := service.NewAuthService(systemRepo, cfg)

	adminToken, _, _ := authService.Login("cat_admin", "password123", "127.0.0.1", "")
	engToken, _, _ := authService.Login("cat_eng", "password123", "127.0.0.1", "")
	opToken, _, _ := authService.Login("cat_op", "password123", "127.0.0.1", "")

	r := gin.New()
	r.Use(gin.Recovery())

	api := r.Group("/api")
	{
		categoriesGroup := api.Group("/categories")
		categoriesGroup.Use(middleware.JWTAuth(authService))
		{
			categoriesGroup.GET("", middleware.RequirePermission("category.view"), categoryHandler.ListCategories)
			categoriesGroup.GET("/:id", middleware.RequirePermission("category.view"), categoryHandler.GetCategory)
			categoriesGroup.POST("", middleware.RequirePermission("category.manage"), categoryHandler.CreateCategory)
			categoriesGroup.PUT("/reorder", middleware.RequirePermission("category.manage"), categoryHandler.ReorderCategories)
			categoriesGroup.PUT("/:id", middleware.RequirePermission("category.manage"), categoryHandler.UpdateCategory)
			categoriesGroup.PUT("/:id/toggle", middleware.RequirePermission("category.manage"), categoryHandler.ToggleCategory)
			categoriesGroup.DELETE("/:id", middleware.RequirePermission("category.manage"), categoryHandler.DeleteCategory)
			categoriesGroup.POST("/:id/parameters", middleware.RequirePermission("category.assign"), categoryHandler.AssignParameters)
		}
	}

	return &CategoryTestEnv{
		Router:     r,
		DB:         db,
		AdminToken: adminToken,
		EngToken:   engToken,
		OpToken:    opToken,
	}
}

func TestCategoryHandler_ListCategories(t *testing.T) {
	env := setupCategoryTestRouter(t)

	// Operator should be able to view categories
	req, _ := http.NewRequest("GET", "/api/categories", nil)
	req.Header.Set("Authorization", "Bearer "+env.OpToken)
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success bool                      `json:"success"`
		Data    []model.ParameterCategory `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if !resp.Success || len(resp.Data) < 3 {
		t.Errorf("Expected at least 3 seeded categories (AQMS/WQMS), found %d", len(resp.Data))
	}

	foundStatus := false
	for _, cat := range resp.Data {
		if cat.Code == "status" && cat.Name == "Status" && cat.Domain == model.DomainAll && cat.IconKey == "power" {
			foundStatus = true
			break
		}
	}
	if !foundStatus {
		t.Errorf("Expected built-in 'Status' category (code=status, domain=ALL, icon_key=power) to be seeded in ListCategories response")
	}
}

func TestCategoryHandler_CRUDAndRBAC(t *testing.T) {
	env := setupCategoryTestRouter(t)

	// 1. Operator attempts to create category -> Forbidden 403
	createPayload := service.CreateCategoryRequest{
		Code:              "custom_chem",
		Name:              "Custom Chemical",
		Description:       "Test custom chemical category",
		IconKey:           "gauge",
		ColorKey:          "teal",
		DefaultColumnSpan: 2,
	}
	body, _ := json.Marshal(createPayload)

	req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+env.OpToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected 403 Forbidden for operator creating category, got %d", w.Code)
	}

	// 2. Engineer creates category -> 201 Created
	req, _ = http.NewRequest("POST", "/api/categories", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+env.EngToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected 201 Created for engineer creating category, got %d: %s", w.Code, w.Body.String())
	}

	var createdResp struct {
		Success bool                    `json:"success"`
		Data    model.ParameterCategory `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &createdResp)
	createdID := createdResp.Data.ID

	// 3. Update category
	newDesc := "Updated chemical domain description"
	updatePayload := service.UpdateCategoryRequest{
		Description: &newDesc,
	}
	uBody, _ := json.Marshal(updatePayload)
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/categories/%d", createdID), bytes.NewBuffer(uBody))
	req.Header.Set("Authorization", "Bearer "+env.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for updating category, got %d: %s", w.Code, w.Body.String())
	}

	// 4. Toggle disabled
	toggleBody, _ := json.Marshal(map[string]bool{"enabled": false})
	req, _ = http.NewRequest("PUT", fmt.Sprintf("/api/categories/%d/toggle", createdID), bytes.NewBuffer(toggleBody))
	req.Header.Set("Authorization", "Bearer "+env.EngToken)
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for toggle category, got %d: %s", w.Code, w.Body.String())
	}

	// 5. Delete category safely
	req, _ = http.NewRequest("DELETE", fmt.Sprintf("/api/categories/%d", createdID), nil)
	req.Header.Set("Authorization", "Bearer "+env.AdminToken)
	w = httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for deleting category, got %d: %s", w.Code, w.Body.String())
	}
}

func TestCategoryHandler_AssignParameters(t *testing.T) {
	env := setupCategoryTestRouter(t)

	var gasCat model.ParameterCategory
	env.DB.Where("code = ?", "gas_pollutants").First(&gasCat)

	dev := model.Device{DeviceCode: "AQMS-01", DeviceName: "Ambient AQMS", DeviceType: "MODBUS_TCP"}
	env.DB.Create(&dev)

	p1 := model.Parameter{DeviceID: dev.ID, ParameterCode: "CO", ParameterName: "Carbon Monoxide"}
	env.DB.Create(&p1)

	// Assign CO to gasCat
	assignPayload := map[string]interface{}{
		"parameter_ids": []uint{p1.ID},
	}
	aBody, _ := json.Marshal(assignPayload)

	req, _ := http.NewRequest("POST", fmt.Sprintf("/api/categories/%d/parameters", gasCat.ID), bytes.NewBuffer(aBody))
	req.Header.Set("Authorization", "Bearer "+env.EngToken)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	env.Router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected 200 OK for assigning parameter, got %d: %s", w.Code, w.Body.String())
	}

	var rechecked model.Parameter
	env.DB.First(&rechecked, p1.ID)
	if rechecked.CategoryID == nil || *rechecked.CategoryID != gasCat.ID {
		t.Errorf("Expected parameter CO to have category_id %d, got: %v", gasCat.ID, rechecked.CategoryID)
	}
}

func TestCategoryHandler_AllNonOperatorRolesCanCRUDCategories(t *testing.T) {
	env := setupCategoryTestRouter(t)

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	systemRepo := repository.NewSystemRepository(env.DB)
	authService := service.NewAuthService(systemRepo, config.DefaultConfig())

	// Fetch or create ADMIN, SUPERADMIN, and custom non-Operator ("Supervisor") roles
	var superRole, adminRole, customRole model.Role
	if err := env.DB.Where("LOWER(name) = ?", "superadmin").First(&superRole).Error; err != nil {
		superRole = model.Role{Name: "SUPERADMIN", Description: "Superadmin"}
		if err := env.DB.Create(&superRole).Error; err != nil {
			t.Fatalf("failed to create SUPERADMIN role: %v", err)
		}
	}
	if err := env.DB.Where("LOWER(name) = ?", "admin").First(&adminRole).Error; err != nil {
		adminRole = model.Role{Name: "ADMIN", Description: "Admin"}
		if err := env.DB.Create(&adminRole).Error; err != nil {
			t.Fatalf("failed to create ADMIN role: %v", err)
		}
	}
	if err := env.DB.Where("LOWER(name) = ?", "supervisor").First(&customRole).Error; err != nil {
		customRole = model.Role{Name: "Supervisor", Description: "Custom non-operator role"}
		if err := env.DB.Create(&customRole).Error; err != nil {
			t.Fatalf("failed to create Supervisor role: %v", err)
		}
	}

	env.DB.Create(&model.User{Username: "u_super", Email: "s@cat.local", Password: string(hash), RoleID: superRole.ID, IsActive: true})
	env.DB.Create(&model.User{Username: "u_admin", Email: "a@cat.local", Password: string(hash), RoleID: adminRole.ID, IsActive: true})
	env.DB.Create(&model.User{Username: "u_supv", Email: "supv@cat.local", Password: string(hash), RoleID: customRole.ID, IsActive: true})

	superTok, _, _ := authService.Login("u_super", "password123", "127.0.0.1", "")
	adminTok, _, _ := authService.Login("u_admin", "password123", "127.0.0.1", "")
	supvTok, _, _ := authService.Login("u_supv", "password123", "127.0.0.1", "")

	tokens := map[string]string{
		"SUPERADMIN": superTok,
		"ADMIN":      adminTok,
		"ENGINEER":   env.EngToken,
		"Supervisor": supvTok,
	}

	idx := 0
	for roleLabel, tok := range tokens {
		idx++
		code := fmt.Sprintf("cat_role_%d", idx)
		b, _ := json.Marshal(service.CreateCategoryRequest{
			Code:   code,
			Name:   "Category " + roleLabel,
			Domain: "SPARING",
		})
		req, _ := http.NewRequest("POST", "/api/categories", bytes.NewBuffer(b))
		req.Header.Set("Authorization", "Bearer "+tok)
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		env.Router.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected non-Operator role %s to create category (201), got %d: %s", roleLabel, w.Code, w.Body.String())
		}

		var createdResp struct {
			Data model.ParameterCategory `json:"data"`
		}
		_ = json.Unmarshal(w.Body.Bytes(), &createdResp)

		// Operator must still be blocked from deleting or updating
		opDelReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/categories/%d", createdResp.Data.ID), nil)
		opDelReq.Header.Set("Authorization", "Bearer "+env.OpToken)
		opDelW := httptest.NewRecorder()
		env.Router.ServeHTTP(opDelW, opDelReq)
		if opDelW.Code != http.StatusForbidden {
			t.Fatalf("expected OPERATOR to receive 403 Forbidden on DELETE category, got %d", opDelW.Code)
		}

		// Non-Operator role can update and delete
		updName := "Updated " + roleLabel
		ub, _ := json.Marshal(service.UpdateCategoryRequest{Name: &updName})
		uReq, _ := http.NewRequest("PUT", fmt.Sprintf("/api/categories/%d", createdResp.Data.ID), bytes.NewBuffer(ub))
		uReq.Header.Set("Authorization", "Bearer "+tok)
		uReq.Header.Set("Content-Type", "application/json")
		uW := httptest.NewRecorder()
		env.Router.ServeHTTP(uW, uReq)
		if uW.Code != http.StatusOK {
			t.Fatalf("expected non-Operator role %s to update category (200), got %d: %s", roleLabel, uW.Code, uW.Body.String())
		}

		dReq, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/categories/%d", createdResp.Data.ID), nil)
		dReq.Header.Set("Authorization", "Bearer "+tok)
		dW := httptest.NewRecorder()
		env.Router.ServeHTTP(dW, dReq)
		if dW.Code != http.StatusOK {
			t.Fatalf("expected non-Operator role %s to delete category (200), got %d: %s", roleLabel, dW.Code, dW.Body.String())
		}
	}
}
