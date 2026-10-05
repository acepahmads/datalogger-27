package service_test

import (
	"fmt"
	"testing"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func TestAuthServiceJWT(t *testing.T) {
	dsn := fmt.Sprintf("file:memauth_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	_ = db.AutoMigrate(&model.User{}, &model.Role{}, &model.AuditTrail{})

	role := model.Role{Name: "Administrator"}
	db.Create(&role)

	hash, _ := bcrypt.GenerateFromPassword([]byte("testpass123"), bcrypt.DefaultCost)
	user := model.User{
		Username: "testadmin",
		Password: string(hash),
		RoleID:   role.ID,
		IsActive: true,
	}
	db.Create(&user)

	cfg := config.DefaultConfig()
	systemRepo := repository.NewSystemRepository(db)
	authService := service.NewAuthService(systemRepo, cfg)

	// Test Valid Login
	token, u, err := authService.Login("testadmin", "testpass123", "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if token == "" {
		t.Fatal("Token is empty")
	}
	if u.Username != "testadmin" {
		t.Errorf("Expected username testadmin, got %s", u.Username)
	}

	// Test Token Validation
	claims, err := authService.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}
	if claims.Username != "testadmin" {
		t.Errorf("Expected claims username testadmin, got %s", claims.Username)
	}
	if claims.Role != "Administrator" {
		t.Errorf("Expected claims role Administrator, got %s", claims.Role)
	}

	// Test Invalid Password
	_, _, err = authService.Login("testadmin", "wrongpassword", "127.0.0.1", "test-agent")
	if err == nil {
		t.Fatal("Expected error for wrong password, got nil")
	}
}
