package repository

import (
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
		query = query.Where("status = ?", status)
	}
	err := query.Limit(100).Find(&alarms).Error
	return alarms, err
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
