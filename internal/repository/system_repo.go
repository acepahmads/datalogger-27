package repository

import (
	"errors"
	"strings"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

type SystemRepository struct {
	db *gorm.DB
}

func NewSystemRepository(db *gorm.DB) *SystemRepository {
	return &SystemRepository{db: db}
}

func (r *SystemRepository) GetDB() *gorm.DB {
	return r.db
}

// Devices
func (r *SystemRepository) GetAllDevices() ([]model.Device, error) {
	var devices []model.Device
	err := r.db.Preload("TypeRel").Preload("Connection").Preload("Parameters").Find(&devices).Error
	return devices, err
}

func (r *SystemRepository) GetDeviceByID(id uint) (*model.Device, error) {
	var dev model.Device
	err := r.db.Preload("TypeRel").Preload("Connection").Preload("Parameters").First(&dev, id).Error
	if err != nil {
		return nil, err
	}
	return &dev, nil
}

// Alarms
func (r *SystemRepository) GetAllAlarms(status string) ([]model.Alarm, error) {
	var alarms []model.Alarm
	query := r.db.Preload("Device").Preload("Parameter").Order("triggered_at DESC")
	if status != "" {
		st := strings.ToUpper(strings.TrimSpace(status))
		query = query.Where("status = ?", st)
		if st == "ACTIVE" || st == "ACKNOWLEDGED" {
			query = query.Where("(device_id IS NULL OR device_id IN (SELECT id FROM devices WHERE (enabled = 1 OR enabled = true) AND status != 'DISABLED')) AND (parameter_id IS NULL OR parameter_id IN (SELECT id FROM parameters WHERE (enabled = 1 OR enabled = true)))")
		}
	}
	err := query.Limit(100).Find(&alarms).Error
	return alarms, err
}

func (r *SystemRepository) ClearActiveAlarmsByDevice(deviceID uint) error {
	now := time.Now().UTC()
	return r.db.Model(&model.Alarm{}).
		Where("device_id = ? AND status IN ?", deviceID, []model.AlarmStatus{model.AlarmActive, model.AlarmAcked}).
		Updates(map[string]interface{}{
			"status":     model.AlarmCleared,
			"cleared_at": &now,
		}).Error
}

func (r *SystemRepository) AcknowledgeAlarm(id uint, username string) error {
	now := time.Now()
	return r.db.Model(&model.Alarm{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":          model.AlarmAcked,
		"acknowledged_at": &now,
		"acknowledged_by": username,
	}).Error
}

// Filter Parameter Definitions
type SystemLogFilterParams struct {
	Level     string
	Component string
	Search    string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

type AuditTrailFilterParams struct {
	Username  string
	Action    string
	Resource  string
	Search    string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

type CommunicationLogFilterParams struct {
	DeviceID  *uint
	Protocol  string
	Direction string
	Status    string
	Search    string
	StartTime *time.Time
	EndTime   *time.Time
	Page      int
	PageSize  int
}

// Logs & Audit
func (r *SystemRepository) GetSystemLogs(limit int) ([]model.SystemLog, error) {
	var logs []model.SystemLog
	err := r.db.Order("created_at DESC, id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *SystemRepository) GetSystemLogsFiltered(params SystemLogFilterParams) ([]model.SystemLog, int64, error) {
	query := r.db.Model(&model.SystemLog{})
	if params.Level != "" && strings.ToUpper(params.Level) != "ALL" {
		query = query.Where("level = ?", strings.ToUpper(params.Level))
	}
	if params.Component != "" {
		query = query.Where("component = ?", params.Component)
	}
	if params.StartTime != nil {
		query = query.Where("created_at >= ?", *params.StartTime)
	}
	if params.EndTime != nil {
		query = query.Where("created_at <= ?", *params.EndTime)
	}
	if params.Search != "" {
		like := "%" + params.Search + "%"
		query = query.Where("message LIKE ? OR details LIKE ? OR component LIKE ?", like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 || pageSize > 500 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	var logs []model.SystemLog
	err := query.Order("created_at DESC, id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

func (r *SystemRepository) AddSystemLog(log *model.SystemLog) error {
	return r.db.Create(log).Error
}

func (r *SystemRepository) GetAuditTrails(limit int) ([]model.AuditTrail, error) {
	var trails []model.AuditTrail
	err := r.db.Order("created_at DESC, id DESC").Limit(limit).Find(&trails).Error
	return trails, err
}

func (r *SystemRepository) GetAuditTrailsFiltered(params AuditTrailFilterParams) ([]model.AuditTrail, int64, error) {
	query := r.db.Model(&model.AuditTrail{})
	if params.Username != "" {
		query = query.Where("username = ?", params.Username)
	}
	if params.Action != "" {
		query = query.Where("action = ?", params.Action)
	}
	if params.Resource != "" {
		query = query.Where("resource = ?", params.Resource)
	}
	if params.StartTime != nil {
		query = query.Where("created_at >= ?", *params.StartTime)
	}
	if params.EndTime != nil {
		query = query.Where("created_at <= ?", *params.EndTime)
	}
	if params.Search != "" {
		like := "%" + params.Search + "%"
		query = query.Where("details LIKE ? OR action LIKE ? OR resource LIKE ? OR username LIKE ?", like, like, like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 || pageSize > 500 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	var trails []model.AuditTrail
	err := query.Order("created_at DESC, id DESC").Offset(offset).Limit(pageSize).Find(&trails).Error
	return trails, total, err
}

func (r *SystemRepository) AddAuditTrail(trail *model.AuditTrail) error {
	return r.db.Create(trail).Error
}

func (r *SystemRepository) GetCommunicationLogs(limit int) ([]model.CommunicationLog, error) {
	var logs []model.CommunicationLog
	err := r.db.Order("created_at DESC, id DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *SystemRepository) GetCommunicationLogsFiltered(params CommunicationLogFilterParams) ([]model.CommunicationLog, int64, error) {
	query := r.db.Model(&model.CommunicationLog{})
	if params.DeviceID != nil && *params.DeviceID > 0 {
		query = query.Where("device_id = ?", *params.DeviceID)
	}
	if params.Protocol != "" && strings.ToUpper(params.Protocol) != "ALL" {
		query = query.Where("protocol = ?", params.Protocol)
	}
	if params.Direction != "" && strings.ToUpper(params.Direction) != "ALL" {
		query = query.Where("direction = ?", strings.ToUpper(params.Direction))
	}
	if params.Status != "" && strings.ToUpper(params.Status) != "ALL" {
		query = query.Where("status = ?", strings.ToUpper(params.Status))
	}
	if params.StartTime != nil {
		query = query.Where("created_at >= ?", *params.StartTime)
	}
	if params.EndTime != nil {
		query = query.Where("created_at <= ?", *params.EndTime)
	}
	if params.Search != "" {
		like := "%" + params.Search + "%"
		query = query.Where("payload_hex LIKE ? OR error_message LIKE ?", like, like)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 || pageSize > 500 {
		pageSize = 50
	}
	offset := (page - 1) * pageSize

	var logs []model.CommunicationLog
	err := query.Order("created_at DESC, id DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	return logs, total, err
}

func (r *SystemRepository) AddCommunicationLog(log *model.CommunicationLog) error {
	return r.db.Create(log).Error
}

// System Health
func (r *SystemRepository) RecordSystemHealth(health *model.SystemHealth) error {
	return r.db.Create(health).Error
}

func (r *SystemRepository) GetRecentHealth(limit int) ([]model.SystemHealth, error) {
	var list []model.SystemHealth
	err := r.db.Order("timestamp DESC").Limit(limit).Find(&list).Error
	return list, err
}

// User & Auth
func (r *SystemRepository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SystemRepository) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.Preload("Role").First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *SystemRepository) UpdateUserLastLogin(id uint) error {
	now := time.Now()
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("last_login", &now).Error
}

type UserFilterParams struct {
	Search   string
	Status   string // "ACTIVE", "INACTIVE", "ALL"
	RoleID   uint
	Page     int
	PageSize int
}

func (r *SystemRepository) GetUsersFiltered(params UserFilterParams) ([]model.User, int64, error) {
	query := r.db.Model(&model.User{}).Preload("Role")

	if params.Search != "" {
		like := "%" + params.Search + "%"
		query = query.Where("username LIKE ? OR full_name LIKE ? OR email LIKE ?", like, like, like)
	}

	if strings.EqualFold(params.Status, "ACTIVE") {
		query = query.Where("is_active = ?", true)
	} else if strings.EqualFold(params.Status, "INACTIVE") {
		query = query.Where("is_active = ?", false)
	}

	if params.RoleID > 0 {
		query = query.Where("role_id = ?", params.RoleID)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 || pageSize > 500 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var users []model.User
	err := query.Order("id ASC").Offset(offset).Limit(pageSize).Find(&users).Error
	return users, total, err
}

func (r *SystemRepository) CreateUser(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *SystemRepository) UpdateUser(user *model.User) error {
	return r.db.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"full_name": user.FullName,
		"email":     user.Email,
		"role_id":   user.RoleID,
		"is_active": user.IsActive,
	}).Error
}

func (r *SystemRepository) SetUserActive(id uint, isActive bool) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("is_active", isActive).Error
}

func (r *SystemRepository) UpdateUserPassword(id uint, hashedPassword string) error {
	return r.db.Model(&model.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}

func (r *SystemRepository) DeleteUser(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

func (r *SystemRepository) CountActiveAdmins() (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("users.is_active = ? AND (roles.name = ? OR roles.name = ?)", true, "Administrator", "admin").
		Count(&count).Error
	return count, err
}

// Role & Permission Management
func (r *SystemRepository) GetRoles() ([]model.Role, error) {
	var roles []model.Role
	err := r.db.Preload("Permissions").Order("id ASC").Find(&roles).Error
	return roles, err
}

func (r *SystemRepository) GetRoleByID(id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *SystemRepository) GetRoleByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.db.Preload("Permissions").Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *SystemRepository) CreateRole(role *model.Role) error {
	return r.db.Create(role).Error
}

func (r *SystemRepository) UpdateRole(id uint, name, description string) error {
	return r.db.Model(&model.Role{}).Where("id = ?", id).Updates(map[string]interface{}{
		"name":        name,
		"description": description,
	}).Error
}

func (r *SystemRepository) UpdateRolePermissions(roleID uint, permissionIDs []uint) error {
	var role model.Role
	if err := r.db.First(&role, roleID).Error; err != nil {
		return err
	}

	var perms []model.Permission
	if len(permissionIDs) > 0 {
		if err := r.db.Where("id IN ?", permissionIDs).Find(&perms).Error; err != nil {
			return err
		}
	}

	return r.db.Model(&role).Association("Permissions").Replace(perms)
}

func (r *SystemRepository) DeleteRole(id uint) error {
	var role model.Role
	if err := r.db.First(&role, id).Error; err != nil {
		return err
	}
	// Clear permission associations first
	_ = r.db.Model(&role).Association("Permissions").Clear()
	return r.db.Delete(&role).Error
}

func (r *SystemRepository) CountUsersWithRole(roleID uint) (int64, error) {
	var count int64
	err := r.db.Model(&model.User{}).Where("role_id = ?", roleID).Count(&count).Error
	return count, err
}

func (r *SystemRepository) GetPermissions() ([]model.Permission, error) {
	var perms []model.Permission
	err := r.db.Order("category ASC, code ASC").Find(&perms).Error
	return perms, err
}

func (r *SystemRepository) GetRolePermissionCodes(roleName string) ([]string, error) {
	var role model.Role
	err := r.db.Preload("Permissions").Where("name = ?", roleName).First(&role).Error
	if err != nil {
		return nil, err
	}
	codes := make([]string, 0, len(role.Permissions))
	for _, p := range role.Permissions {
		codes = append(codes, p.Code)
	}
	return codes, nil
}

// Branding & Organization Persistence
func (r *SystemRepository) GetBranding() (*model.SystemBranding, error) {
	var branding model.SystemBranding
	err := r.db.First(&branding).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaultBranding := model.SystemBranding{
				CustomerName:     "PT. Mitra Industri Enviro",
				CustomerSubtitle: "AQMS",
				CustomerHasLogo:  false,
				MakerName:        "PT. Centra Biotech Indonesia",
				MakerSubtitle:    "CBI Industrial IoT",
				MakerHasLogo:     false,
				CreatedAt:        time.Now(),
				UpdatedAt:        time.Now(),
			}
			_ = r.db.Create(&defaultBranding)
			return &defaultBranding, nil
		}
		return nil, err
	}
	return &branding, nil
}

func (r *SystemRepository) SaveBranding(b *model.SystemBranding) error {
	var existing model.SystemBranding
	err := r.db.First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			b.CreatedAt = time.Now()
			b.UpdatedAt = time.Now()
			return r.db.Create(b).Error
		}
		return err
	}
	b.ID = existing.ID
	b.CreatedAt = existing.CreatedAt
	b.UpdatedAt = time.Now()
	return r.db.Save(b).Error
}

// Station Location Persistence
func (r *SystemRepository) GetStationLocation() (*model.StationLocation, error) {
	var loc model.StationLocation
	err := r.db.First(&loc).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			defaultLoc := model.StationLocation{
				StationName:    "Ambient Monitoring Station",
				Latitude:       nil,
				Longitude:      nil,
				LocationSource: model.LocationSourceUnconfigured,
				IsConfigured:   false,
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
			}
			_ = r.db.Create(&defaultLoc)
			return &defaultLoc, nil
		}
		return nil, err
	}
	return &loc, nil
}

func (r *SystemRepository) SaveStationLocation(loc *model.StationLocation) error {
	var existing model.StationLocation
	err := r.db.First(&existing).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			loc.CreatedAt = time.Now()
			loc.UpdatedAt = time.Now()
			return r.db.Create(loc).Error
		}
		return err
	}
	loc.ID = existing.ID
	loc.CreatedAt = existing.CreatedAt
	loc.UpdatedAt = time.Now()
	return r.db.Save(loc).Error
}


