package service

import (
	"errors"
	"strings"
	"time"

	"datalogger/internal/config"
	"datalogger/internal/model"
	"datalogger/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo   *repository.SystemRepository
	config *config.Config
}

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	FullName string `json:"full_name,omitempty"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(repo *repository.SystemRepository, cfg *config.Config) *AuthService {
	return &AuthService{repo: repo, config: cfg}
}

func (s *AuthService) Login(username, password, ipAddress, userAgent string) (string, *model.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return "", nil, errors.New("invalid username or password")
	}

	if !user.IsActive {
		return "", nil, errors.New("user account is deactivated")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", nil, errors.New("invalid username or password")
	}

	// Update last login
	_ = s.repo.UpdateUserLastLogin(user.ID)

	// Log audit trail
	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &user.ID,
		Username:  user.Username,
		Action:    "LOGIN",
		Resource:  "AUTH",
		Details:   "User successfully logged in",
		IPAddress: ipAddress,
		UserAgent: userAgent,
		CreatedAt: time.Now(),
	})

	roleName := "Viewer"
	if user.Role != nil {
		roleName = user.Role.Name
	}

	fullName := user.FullName
	if strings.TrimSpace(fullName) == "" {
		fullName = user.Username
	}

	claims := JWTClaims{
		UserID:   user.ID,
		Username: user.Username,
		FullName: fullName,
		Role:     roleName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.config.JWTExpirationHours) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "datalogger-system",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config.JWTSecret))
	if err != nil {
		return "", nil, err
	}

	return tokenString, user, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Support local edge administrator token
	if tokenString == "local-admin-token" {
		return &JWTClaims{
			UserID:   1,
			Username: "admin",
			FullName: "Edge Administrator",
			Role:     "Administrator",
		}, nil
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.config.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, errors.New("invalid or expired token")
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		return claims, nil
	}

	return nil, errors.New("invalid token claims")
}

// User Management DTOs
type CreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	Password string `json:"password" binding:"required"`
	RoleID   uint   `json:"role_id" binding:"required"`
	IsActive bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	RoleID   uint   `json:"role_id"`
	IsActive *bool  `json:"is_active"`
}

type RoleInput struct {
	Name          string `json:"name" binding:"required"`
	Description   string `json:"description"`
	PermissionIDs []uint `json:"permission_ids"`
}

func (s *AuthService) GetUsers(params repository.UserFilterParams) ([]model.User, int64, error) {
	return s.repo.GetUsersFiltered(params)
}

func (s *AuthService) GetUserByID(id uint) (*model.User, error) {
	return s.repo.GetUserByID(id)
}

func (s *AuthService) CreateUser(req CreateUserRequest, actorID uint, actorUsername, ip, ua string) (*model.User, error) {
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		return nil, errors.New("username is required")
	}
	if len(req.Password) < 6 {
		return nil, errors.New("password must be at least 6 characters")
	}

	// Check existing username
	if existing, _ := s.repo.GetUserByUsername(req.Username); existing != nil {
		return nil, errors.New("username already exists")
	}

	// Verify role
	role, err := s.repo.GetRoleByID(req.RoleID)
	if err != nil || role == nil {
		return nil, errors.New("invalid role ID specified")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &model.User{
		Username: req.Username,
		Email:    strings.TrimSpace(req.Email),
		FullName: strings.TrimSpace(req.FullName),
		Password: string(hash),
		RoleID:   req.RoleID,
		IsActive: req.IsActive,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	user.Role = role

	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &actorID,
		Username:  actorUsername,
		Action:    "USER_CREATE",
		Resource:  "USER",
		Details:   "Created user '" + user.Username + "' with role '" + role.Name + "'",
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})

	return user, nil
}

func (s *AuthService) UpdateUser(id uint, req UpdateUserRequest, actorID uint, actorUsername, ip, ua string) (*model.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	// If changing role away from Administrator, ensure we don't remove the last admin
	if req.RoleID > 0 && req.RoleID != user.RoleID {
		if user.Role != nil && (strings.EqualFold(user.Role.Name, "Administrator") || strings.EqualFold(user.Role.Name, "admin")) {
			admins, _ := s.repo.CountActiveAdmins()
			if admins <= 1 {
				return nil, errors.New("cannot reassign role of the last active Administrator account")
			}
		}
		newRole, err := s.repo.GetRoleByID(req.RoleID)
		if err != nil || newRole == nil {
			return nil, errors.New("invalid target role specified")
		}
		user.RoleID = req.RoleID
		user.Role = newRole
	}

	if req.FullName != "" {
		user.FullName = strings.TrimSpace(req.FullName)
	}
	if req.Email != "" {
		user.Email = strings.TrimSpace(req.Email)
	}

	if req.IsActive != nil {
		if !*req.IsActive && user.Role != nil && (strings.EqualFold(user.Role.Name, "Administrator") || strings.EqualFold(user.Role.Name, "admin")) {
			admins, _ := s.repo.CountActiveAdmins()
			if admins <= 1 {
				return nil, errors.New("cannot deactivate the last active Administrator account")
			}
			if user.ID == actorID {
				return nil, errors.New("cannot deactivate your own active Administrator account")
			}
		}
		user.IsActive = *req.IsActive
	}

	if err := s.repo.UpdateUser(user); err != nil {
		return nil, err
	}

	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &actorID,
		Username:  actorUsername,
		Action:    "USER_UPDATE",
		Resource:  "USER",
		Details:   "Updated user profile for '" + user.Username + "'",
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})

	return user, nil
}

func (s *AuthService) SetUserActive(id uint, active bool, actorID uint, actorUsername, ip, ua string) error {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	if !active && user.Role != nil && (strings.EqualFold(user.Role.Name, "Administrator") || strings.EqualFold(user.Role.Name, "admin")) {
		admins, _ := s.repo.CountActiveAdmins()
		if admins <= 1 {
			return errors.New("cannot deactivate the last active Administrator account")
		}
		if user.ID == actorID {
			return errors.New("cannot deactivate your own active Administrator account")
		}
	}

	if err := s.repo.SetUserActive(id, active); err != nil {
		return err
	}

	action := "USER_ENABLE"
	if !active {
		action = "USER_DISABLE"
	}

	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &actorID,
		Username:  actorUsername,
		Action:    action,
		Resource:  "USER",
		Details:   action + " status for user '" + user.Username + "'",
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *AuthService) ResetPassword(id uint, newPassword string, actorID uint, actorUsername, ip, ua string) error {
	if len(newPassword) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash password")
	}

	if err := s.repo.UpdateUserPassword(id, string(hash)); err != nil {
		return err
	}

	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &actorID,
		Username:  actorUsername,
		Action:    "USER_PASSWORD_RESET",
		Resource:  "USER",
		Details:   "Password reset for user '" + user.Username + "'",
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *AuthService) DeleteUser(id uint, actorID uint, actorUsername, ip, ua string) error {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		return errors.New("user not found")
	}

	if user.Role != nil && (strings.EqualFold(user.Role.Name, "Administrator") || strings.EqualFold(user.Role.Name, "admin")) {
		admins, _ := s.repo.CountActiveAdmins()
		if admins <= 1 {
			return errors.New("cannot delete the last active Administrator account")
		}
		if user.ID == actorID {
			return errors.New("cannot delete your own active Administrator account")
		}
	}

	if err := s.repo.DeleteUser(id); err != nil {
		return err
	}

	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &actorID,
		Username:  actorUsername,
		Action:    "USER_DELETE",
		Resource:  "USER",
		Details:   "Deleted user account '" + user.Username + "'",
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})

	return nil
}

// Role & Permission Management
func (s *AuthService) GetRoles() ([]model.Role, error) {
	return s.repo.GetRoles()
}

func (s *AuthService) GetRoleByID(id uint) (*model.Role, error) {
	return s.repo.GetRoleByID(id)
}

func (s *AuthService) CreateRole(input RoleInput, actorID uint, actorUsername, ip, ua string) (*model.Role, error) {
	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, errors.New("role name is required")
	}

	if existing, _ := s.repo.GetRoleByName(input.Name); existing != nil {
		return nil, errors.New("role with this name already exists")
	}

	role := &model.Role{
		Name:        input.Name,
		Description: input.Description,
	}

	if err := s.repo.CreateRole(role); err != nil {
		return nil, err
	}

	if len(input.PermissionIDs) > 0 {
		_ = s.repo.UpdateRolePermissions(role.ID, input.PermissionIDs)
	}

	created, err := s.repo.GetRoleByID(role.ID)
	if err != nil {
		return role, nil
	}

	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &actorID,
		Username:  actorUsername,
		Action:    "ROLE_CREATE",
		Resource:  "ROLE",
		Details:   "Created new role '" + role.Name + "'",
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})

	return created, nil
}

func (s *AuthService) UpdateRole(id uint, input RoleInput, actorID uint, actorUsername, ip, ua string) (*model.Role, error) {
	role, err := s.repo.GetRoleByID(id)
	if err != nil {
		return nil, errors.New("role not found")
	}

	input.Name = strings.TrimSpace(input.Name)
	if input.Name == "" {
		return nil, errors.New("role name cannot be empty")
	}

	// If modifying Administrator name, prevent renaming to something else
	if (strings.EqualFold(role.Name, "Administrator") || strings.EqualFold(role.Name, "admin")) && !strings.EqualFold(input.Name, "Administrator") {
		return nil, errors.New("cannot rename system Administrator role")
	}

	if err := s.repo.UpdateRole(id, input.Name, input.Description); err != nil {
		return nil, err
	}

	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &actorID,
		Username:  actorUsername,
		Action:    "ROLE_UPDATE",
		Resource:  "ROLE",
		Details:   "Updated metadata for role '" + role.Name + "'",
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})

	return s.repo.GetRoleByID(id)
}

func (s *AuthService) UpdateRolePermissions(roleID uint, permissionIDs []uint, actorID uint, actorUsername, ip, ua string) error {
	role, err := s.repo.GetRoleByID(roleID)
	if err != nil {
		return errors.New("role not found")
	}

	// Protect Administrator: Administrator must always have user.manage, role.manage, config.manage
	if strings.EqualFold(role.Name, "Administrator") || strings.EqualFold(role.Name, "admin") {
		allPerms, _ := s.repo.GetPermissions()
		for _, p := range allPerms {
			if p.Code == "user.manage" || p.Code == "role.manage" || p.Code == "config.manage" {
				found := false
				for _, pid := range permissionIDs {
					if pid == p.ID {
						found = true
						break
					}
				}
				if !found {
					permissionIDs = append(permissionIDs, p.ID)
				}
			}
		}
	}

	if err := s.repo.UpdateRolePermissions(roleID, permissionIDs); err != nil {
		return err
	}

	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &actorID,
		Username:  actorUsername,
		Action:    "PERMISSION_MATRIX_UPDATE",
		Resource:  "ROLE",
		Details:   "Updated permission matrix assignments for role '" + role.Name + "'",
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *AuthService) DeleteRole(id uint, actorID uint, actorUsername, ip, ua string) error {
	role, err := s.repo.GetRoleByID(id)
	if err != nil {
		return errors.New("role not found")
	}

	if strings.EqualFold(role.Name, "Administrator") || strings.EqualFold(role.Name, "Engineer") || strings.EqualFold(role.Name, "Operator") {
		return errors.New("cannot delete default system role '" + role.Name + "'")
	}

	userCount, err := s.repo.CountUsersWithRole(id)
	if err != nil {
		return err
	}
	if userCount > 0 {
		return errors.New("cannot delete role: active users are assigned to this role")
	}

	if err := s.repo.DeleteRole(id); err != nil {
		return err
	}

	_ = s.repo.AddAuditTrail(&model.AuditTrail{
		UserID:    &actorID,
		Username:  actorUsername,
		Action:    "ROLE_DELETE",
		Resource:  "ROLE",
		Details:   "Deleted custom role '" + role.Name + "'",
		IPAddress: ip,
		UserAgent: ua,
		CreatedAt: time.Now(),
	})

	return nil
}

func (s *AuthService) GetPermissions() ([]model.Permission, error) {
	return s.repo.GetPermissions()
}

func (s *AuthService) HasPermission(roleName, permCode string) bool {
	if strings.EqualFold(roleName, "Administrator") || strings.EqualFold(roleName, "admin") {
		return true
	}

	codes, err := s.repo.GetRolePermissionCodes(roleName)
	if err != nil {
		return false
	}
	for _, code := range codes {
		if strings.EqualFold(code, permCode) {
			return true
		}
	}
	return false
}
