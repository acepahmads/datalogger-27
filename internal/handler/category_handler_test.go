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
