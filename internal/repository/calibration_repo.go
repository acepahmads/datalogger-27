package repository

import (
	"context"
	"strings"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

type CalibrationLogFilter struct {
	DeviceID    uint
	ParameterID uint
	Domain      string
	Action      string
	Page        int
	PageSize    int
}

type CalibrationRepository struct {
	db *gorm.DB
}

func NewCalibrationRepository(db *gorm.DB) *CalibrationRepository {
	return &CalibrationRepository{db: db}
}

// ListCalibratableParameters returns active analog/numeric parameters and their parent active devices
func (r *CalibrationRepository) ListCalibratableParameters(ctx context.Context, deviceID uint) ([]model.Parameter, map[uint]model.Device, error) {
	var devices []model.Device
	devQuery := r.db.WithContext(ctx).Preload("Connection").Where("deleted_at IS NULL")
	if deviceID > 0 {
		devQuery = devQuery.Where("id = ?", deviceID)
	}
	if err := devQuery.Order("id ASC").Find(&devices).Error; err != nil {
		return nil, nil, err
	}

	devMap := make(map[uint]model.Device, len(devices))
	devIDs := make([]uint, 0, len(devices))
	for _, d := range devices {
		devMap[d.ID] = d
		devIDs = append(devIDs, d.ID)
	}
	if len(devIDs) == 0 {
		return []model.Parameter{}, devMap, nil
	}

	var params []model.Parameter
	if err := r.db.WithContext(ctx).
		Preload("Category").
		Where("device_id IN (?) AND deleted_at IS NULL", devIDs).
		Order("device_id ASC, id ASC").
		Find(&params).Error; err != nil {
		return nil, nil, err
	}

	filtered := make([]model.Parameter, 0, len(params))
	for _, p := range params {
		if p.IsDiscreteStatus() {
			continue
		}
		if strings.EqualFold(string(p.DataType), string(model.DataTypeString)) {
			continue
		}
		filtered = append(filtered, p)
	}
	return filtered, devMap, nil
}

func (r *CalibrationRepository) GetParameterWithDevice(ctx context.Context, paramID uint) (*model.Parameter, *model.Device, error) {
	var param model.Parameter
	if err := r.db.WithContext(ctx).Preload("Category").First(&param, paramID).Error; err != nil {
		return nil, nil, err
	}
	var dev model.Device
	if err := r.db.WithContext(ctx).Preload("Connection").First(&dev, param.DeviceID).Error; err != nil {
		return &param, nil, err
	}
	return &param, &dev, nil
}

func (r *CalibrationRepository) UpdateParameterCalibrationFields(ctx context.Context, paramID uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&model.Parameter{}).Where("id = ?", paramID).Updates(updates).Error
}

func (r *CalibrationRepository) CreateCalibrationLog(ctx context.Context, entry *model.SensorCalibrationLog) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *CalibrationRepository) GetCalibrationLogByID(ctx context.Context, id uint) (*model.SensorCalibrationLog, error) {
	var entry model.SensorCalibrationLog
	if err := r.db.WithContext(ctx).First(&entry, id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (r *CalibrationRepository) ListCalibrationLogs(ctx context.Context, filter CalibrationLogFilter) ([]model.SensorCalibrationLog, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.SensorCalibrationLog{})
	if filter.DeviceID > 0 {
		q = q.Where("device_id = ?", filter.DeviceID)
	}
	if filter.ParameterID > 0 {
		q = q.Where("parameter_id = ?", filter.ParameterID)
	}
	if dom := model.NormalizeCalibrationDomain(filter.Domain); dom != "" {
		q = q.Where("domain = ?", dom)
	}
	if act := strings.ToUpper(strings.TrimSpace(filter.Action)); act != "" {
		q = q.Where("action = ?", act)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := filter.Page
	if page <= 0 {
		page = 1
	}
	pageSize := filter.PageSize
	if pageSize <= 0 {
		pageSize = 50
	}
	if pageSize > 500 {
		pageSize = 500
	}
	offset := (page - 1) * pageSize

	var logs []model.SensorCalibrationLog
	if err := q.Order("created_at DESC, id DESC").Limit(pageSize).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

