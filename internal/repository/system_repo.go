package repository

import (
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
	err := r.db.Preload("DeviceType").Preload("Connection").Preload("Parameters").Find(&devices).Error
	return devices, err
}

func (r *SystemRepository) GetDeviceByID(id uint) (*model.Device, error) {
	var dev model.Device
	err := r.db.Preload("DeviceType").Preload("Connection").Preload("Parameters").First(&dev, id).Error
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

// Logs & Audit
func (r *SystemRepository) GetSystemLogs(limit int) ([]model.SystemLog, error) {
	var logs []model.SystemLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
}

func (r *SystemRepository) AddSystemLog(log *model.SystemLog) error {
	return r.db.Create(log).Error
}

func (r *SystemRepository) GetAuditTrails(limit int) ([]model.AuditTrail, error) {
	var trails []model.AuditTrail
	err := r.db.Order("created_at DESC").Limit(limit).Find(&trails).Error
	return trails, err
}

func (r *SystemRepository) AddAuditTrail(trail *model.AuditTrail) error {
	return r.db.Create(trail).Error
}

func (r *SystemRepository) GetCommunicationLogs(limit int) ([]model.CommunicationLog, error) {
	var logs []model.CommunicationLog
	err := r.db.Order("created_at DESC").Limit(limit).Find(&logs).Error
	return logs, err
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
