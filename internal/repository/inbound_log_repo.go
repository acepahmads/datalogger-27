package repository

import (
	"context"
	"strings"
	"sync/atomic"
	"time"

	"datalogger/internal/model"

	"gorm.io/gorm"
)

const (
	MaxInboundRawPayloadBytes = 256 * 1024 // 256 KB per entry cap
	MaxRetainedInboundLogs    = 10000
)

type InboundLogFilter struct {
	TenantID       uint
	SiteCode       string
	DeviceCode     string
	Status         string
	Source         string
	LoggerType     string
	DetectedFormat string
	Endpoint       string
	Search         string
	StartTime      *time.Time
	EndTime        *time.Time
	Page           int
	PageSize       int
}

type InboundLogSummary struct {
	TotalRequests  int64      `json:"total_requests"`
	AcceptedCount  int64      `json:"accepted_count"`
	RejectedCount  int64      `json:"rejected_count"`
	TotalRecords   int64      `json:"total_records"`
	TotalParams    int64      `json:"total_params"`
	ActiveTenants  int64      `json:"active_tenants"`
	LastReceivedAt *time.Time `json:"last_received_at,omitempty"`
}

type InboundLogRepository struct {
	db         *gorm.DB
	writeCount atomic.Int64
}

func NewInboundLogRepository(db *gorm.DB) *InboundLogRepository {
	if db != nil {
		_ = db.AutoMigrate(&model.InboundIngestLog{})
	}
	return &InboundLogRepository{db: db}
}

func (r *InboundLogRepository) Create(ctx context.Context, entry *model.InboundIngestLog) error {
	if r == nil || r.db == nil || entry == nil {
		return nil
	}
	if entry.ReceivedAt.IsZero() {
		entry.ReceivedAt = time.Now().UTC()
	}
	if len(entry.RawPayload) > MaxInboundRawPayloadBytes {
		entry.RawPayload = entry.RawPayload[:MaxInboundRawPayloadBytes] + "\n/* ... truncated at 256 KB ... */"
	}
	if entry.TenantID > 0 && strings.TrimSpace(entry.TenantName) == "" {
		var t model.Tenant
		if err := r.db.WithContext(ctx).Select("id, name").Where("id = ?", entry.TenantID).First(&t).Error; err == nil && t.Name != "" {
			entry.TenantName = t.Name
		}
	}
	if err := r.db.WithContext(ctx).Create(entry).Error; err != nil {
		return err
	}

	// Periodically prune old entries beyond retention limit (every 200 writes) to keep Cloud DB fast
	if r.writeCount.Add(1)%200 == 0 {
		go r.pruneExcessRows()
	}
	return nil
}

func (r *InboundLogRepository) pruneExcessRows() {
	if r == nil || r.db == nil {
		return
	}
	var total int64
	if err := r.db.Model(&model.InboundIngestLog{}).Count(&total).Error; err != nil || total <= MaxRetainedInboundLogs {
		return
	}
	var cutoffID uint64
	err := r.db.Model(&model.InboundIngestLog{}).
		Select("id").
		Order("id DESC").
		Offset(MaxRetainedInboundLogs).
		Limit(1).
		Scan(&cutoffID).Error
	if err == nil && cutoffID > 0 {
		_ = r.db.Where("id <= ?", cutoffID).Delete(&model.InboundIngestLog{}).Error
	}
}

func (r *InboundLogRepository) applyFilter(q *gorm.DB, f InboundLogFilter) *gorm.DB {
	if f.TenantID > 0 {
		q = q.Where("tenant_id = ?", f.TenantID)
	}
	if sc := strings.TrimSpace(f.SiteCode); sc != "" {
		q = q.Where("LOWER(site_code) LIKE ?", "%"+strings.ToLower(sc)+"%")
	}
	if dc := strings.TrimSpace(f.DeviceCode); dc != "" {
		q = q.Where("LOWER(device_code) LIKE ?", "%"+strings.ToLower(dc)+"%")
	}
	if st := strings.ToUpper(strings.TrimSpace(f.Status)); st != "" {
		q = q.Where("UPPER(status) = ?", st)
	}
	if src := strings.ToUpper(strings.TrimSpace(f.Source)); src != "" {
		q = q.Where("UPPER(source) = ?", src)
	}
	if lt := strings.ToUpper(strings.TrimSpace(f.LoggerType)); lt != "" {
		q = q.Where("UPPER(logger_type) = ?", lt)
	}
	if fm := strings.ToUpper(strings.TrimSpace(f.DetectedFormat)); fm != "" {
		q = q.Where("UPPER(detected_format) LIKE ?", "%"+strings.ToLower(fm)+"%")
	}
	if ep := strings.TrimSpace(f.Endpoint); ep != "" {
		q = q.Where("LOWER(endpoint) LIKE ?", "%"+strings.ToLower(ep)+"%")
	}
	if f.StartTime != nil && !f.StartTime.IsZero() {
		q = q.Where("received_at >= ?", f.StartTime.UTC())
	}
	if f.EndTime != nil && !f.EndTime.IsZero() {
		q = q.Where("received_at <= ?", f.EndTime.UTC())
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		like := "%" + strings.ToLower(s) + "%"
		q = q.Where(
			"LOWER(tenant_name) LIKE ? OR LOWER(site_code) LIKE ? OR LOWER(device_code) LIKE ? OR LOWER(endpoint) LIKE ? OR LOWER(client_ip) LIKE ? OR LOWER(error_message) LIKE ? OR LOWER(raw_payload) LIKE ?",
			like, like, like, like, like, like, like,
		)
	}
	return q
}

func (r *InboundLogRepository) List(ctx context.Context, f InboundLogFilter) ([]model.InboundIngestLog, int64, error) {
	if r == nil || r.db == nil {
		return nil, 0, nil
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	pageSize := f.PageSize
	if pageSize < 1 {
		pageSize = 25
	}
	if pageSize > 100 {
		pageSize = 100
	}

	base := r.applyFilter(r.db.WithContext(ctx).Model(&model.InboundIngestLog{}), f)
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []model.InboundIngestLog
	err := r.applyFilter(r.db.WithContext(ctx).Model(&model.InboundIngestLog{}), f).
		Order("received_at DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error
	return items, total, err
}

func (r *InboundLogRepository) GetByID(ctx context.Context, id uint64) (*model.InboundIngestLog, error) {
	if r == nil || r.db == nil {
		return nil, gorm.ErrRecordNotFound
	}
	var item model.InboundIngestLog
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error; err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *InboundLogRepository) GetSummary(ctx context.Context, f InboundLogFilter) (*InboundLogSummary, error) {
	sum := &InboundLogSummary{}
	if r == nil || r.db == nil {
		return sum, nil
	}

	type aggRow struct {
		TotalRequests int64
		AcceptedCount int64
		RejectedCount int64
		TotalRecords  int64
		TotalParams   int64
		ActiveTenants int64
	}
	var row aggRow
	err := r.applyFilter(r.db.WithContext(ctx).Model(&model.InboundIngestLog{}), f).
		Select(
			"COUNT(*) AS total_requests, " +
				"COALESCE(SUM(CASE WHEN UPPER(status) = 'ACCEPTED' THEN 1 ELSE 0 END), 0) AS accepted_count, " +
				"COALESCE(SUM(CASE WHEN UPPER(status) = 'REJECTED' THEN 1 ELSE 0 END), 0) AS rejected_count, " +
				"COALESCE(SUM(records_processed), 0) AS total_records, " +
				"COALESCE(SUM(parameters_ingested), 0) AS total_params, " +
				"COUNT(DISTINCT CASE WHEN tenant_id > 0 THEN tenant_id END) AS active_tenants",
		).
		Scan(&row).Error
	if err != nil {
		return nil, err
	}

	sum.TotalRequests = row.TotalRequests
	sum.AcceptedCount = row.AcceptedCount
	sum.RejectedCount = row.RejectedCount
	sum.TotalRecords = row.TotalRecords
	sum.TotalParams = row.TotalParams
	sum.ActiveTenants = row.ActiveTenants

	var latest model.InboundIngestLog
	if err := r.applyFilter(r.db.WithContext(ctx).Model(&model.InboundIngestLog{}), f).
		Select("received_at").
		Order("received_at DESC, id DESC").
		Limit(1).
		First(&latest).Error; err == nil && !latest.ReceivedAt.IsZero() {
		t := latest.ReceivedAt
		sum.LastReceivedAt = &t
	}
	return sum, nil
}

func (r *InboundLogRepository) Purge(ctx context.Context, olderThanDays int, clearAll bool) (int64, error) {
	if r == nil || r.db == nil {
		return 0, nil
	}
	if clearAll {
		tx := r.db.WithContext(ctx).Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&model.InboundIngestLog{})
		return tx.RowsAffected, tx.Error
	}
	if olderThanDays < 1 {
		olderThanDays = 7
	}
	cutoff := time.Now().UTC().Add(-time.Duration(olderThanDays) * 24 * time.Hour)
	tx := r.db.WithContext(ctx).Where("received_at < ?", cutoff).Delete(&model.InboundIngestLog{})
	return tx.RowsAffected, tx.Error
}
