package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func setupAutoRetryTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	if err := db.AutoMigrate(
		&model.Device{},
		&model.OutputDestination{},
		&model.DestinationPayloadConfig{},
		&model.OutboundDeliveryQueue{},
		&model.DeliveryLog{},
	); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	return db
}

func TestOutboundAutoRetry_TransientRecoveryAndPermanentSkip(t *testing.T) {
	db := setupAutoRetryTestDB(t)
	repo := repository.NewOutboundRepository(db)
	svc := NewOutboundService(repo, nil, nil).(*outboundService)

	var serverHealthy atomic.Bool
	var hitCount atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hitCount.Add(1)
		if !serverHealthy.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error":"upstream down"}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":1,"success":true,"message":"received"}`))
	}))
	defer ts.Close()

	ctx := context.Background()
	dest := &model.OutputDestination{
		TenantID:                 1,
		Name:                     "SPARING Auto-Retry Dest",
		Type:                     model.DestREST,
		Endpoint:                 ts.URL,
		Enabled:                  true,
		RetryLimit:               2,
		AutoRetryIntervalMinutes: 5,
		AutoRetryRetentionHours:  24,
	}
	if err := db.Create(dest).Error; err != nil {
		t.Fatalf("create dest: %v", err)
	}

	now := time.Now()
	pastDue := now.Add(-1 * time.Minute)
	expiredAt := now.Add(-48 * time.Hour)

	// 1. Two transient FAILED items (HTTP 503 & network error 0), both due now
	item1 := &model.OutboundDeliveryQueue{
		TenantID:        1,
		BatchID:         "BATCH-TRANSIENT-1",
		DestinationID:   dest.ID,
		PeriodStart:     now.Add(-2 * time.Hour),
		PeriodEnd:       now.Add(-1 * time.Hour),
		IntervalSeconds: 3600,
		ParameterCount:  150,
		PayloadSnapshot: `{"uid":"test","data":[]}`,
		IdempotencyKey:  "KEY-T1",
		Status:          model.DeliveryFailed,
		Attempts:        2,
		MaxRetries:      2,
		LastHTTPCode:    0,
		LastError:       "network error: dial tcp timeout",
		NextRetryAt:     &pastDue,
		CreatedAt:       now.Add(-20 * time.Minute),
	}
	item2 := &model.OutboundDeliveryQueue{
		TenantID:        1,
		BatchID:         "BATCH-TRANSIENT-2",
		DestinationID:   dest.ID,
		PeriodStart:     now.Add(-1 * time.Hour),
		PeriodEnd:       now,
		IntervalSeconds: 3600,
		ParameterCount:  150,
		PayloadSnapshot: `{"uid":"test","data":[]}`,
		IdempotencyKey:  "KEY-T2",
		Status:          model.DeliveryFailed,
		Attempts:        2,
		MaxRetries:      2,
		LastHTTPCode:    503,
		LastError:       "HTTP error 503: upstream down",
		NextRetryAt:     &pastDue,
		CreatedAt:       now.Add(-10 * time.Minute),
	}
	// 2. Permanent HTTP 400 item -> must NEVER be auto-retried
	itemPerm := &model.OutboundDeliveryQueue{
		TenantID:        1,
		BatchID:         "BATCH-PERM-400",
		DestinationID:   dest.ID,
		PeriodStart:     now.Add(-3 * time.Hour),
		PeriodEnd:       now.Add(-2 * time.Hour),
		IntervalSeconds: 3600,
		ParameterCount:  150,
		PayloadSnapshot: `{"uid":"test","data":[]}`,
		IdempotencyKey:  "KEY-PERM400",
		Status:          model.DeliveryFailed,
		Attempts:        1,
		MaxRetries:      2,
		LastHTTPCode:    400,
		LastError:       "HTTP error 400: invalid schema",
		NextRetryAt:     nil,
		CreatedAt:       now.Add(-15 * time.Minute),
	}
	// 3. Expired transient item (older than 24h retention) -> must be skipped
	itemExpired := &model.OutboundDeliveryQueue{
		TenantID:        1,
		BatchID:         "BATCH-EXPIRED",
		DestinationID:   dest.ID,
		PeriodStart:     expiredAt.Add(-time.Hour),
		PeriodEnd:       expiredAt,
		IntervalSeconds: 3600,
		ParameterCount:  150,
		PayloadSnapshot: `{"uid":"test","data":[]}`,
		IdempotencyKey:  "KEY-EXP",
		Status:          model.DeliveryFailed,
		Attempts:        2,
		MaxRetries:      2,
		LastHTTPCode:    0,
		LastError:       "network error: offline",
		NextRetryAt:     &pastDue,
		CreatedAt:       expiredAt,
	}

	for _, it := range []*model.OutboundDeliveryQueue{item1, item2, itemPerm, itemExpired} {
		if err := db.Create(it).Error; err != nil {
			t.Fatalf("create queue item %s: %v", it.BatchID, err)
		}
	}

	// Pass 1: Upstream STILL down -> only 1 probe hit, item2 deferred to ~5m later, itemPerm untouched, itemExpired cleared
	serverHealthy.Store(false)
	svc.ProcessAutoRetryOnce()

	if got := hitCount.Load(); got != 1 {
		t.Fatalf("expected exactly 1 probe while offline, got %d hits", got)
	}

	var refreshed1, refreshed2, refreshedPerm, refreshedExp model.OutboundDeliveryQueue
	_ = db.First(&refreshed1, item1.ID).Error
	_ = db.First(&refreshed2, item2.ID).Error
	_ = db.First(&refreshedPerm, itemPerm.ID).Error
	_ = db.First(&refreshedExp, itemExpired.ID).Error

	if refreshed1.Status != model.DeliveryFailed || refreshed1.NextRetryAt == nil || !refreshed1.NextRetryAt.After(time.Now().Add(4*time.Minute)) {
		t.Fatalf("expected item1 FAILED with NextRetryAt ~5m ahead, got status=%s next=%v", refreshed1.Status, refreshed1.NextRetryAt)
	}
	if refreshed2.Status != model.DeliveryFailed || refreshed2.NextRetryAt == nil || !refreshed2.NextRetryAt.After(time.Now().Add(4*time.Minute)) {
		t.Fatalf("expected item2 deferred ~5m ahead without hammering, got status=%s next=%v", refreshed2.Status, refreshed2.NextRetryAt)
	}
	if refreshedPerm.Attempts != 1 || refreshedPerm.NextRetryAt != nil {
		t.Fatalf("permanent HTTP 400 must never auto-retry: attempts=%d next=%v", refreshedPerm.Attempts, refreshedPerm.NextRetryAt)
	}
	if refreshedExp.Attempts != 2 || refreshedExp.NextRetryAt != nil {
		t.Fatalf("expired batch (>24h retention) must not auto-retry and should clear NextRetryAt: attempts=%d next=%v", refreshedExp.Attempts, refreshedExp.NextRetryAt)
	}

	// Pass 2: Simulate 5 minutes elapsed AND upstream comes back online -> BOTH item1 and item2 flush to ACKNOWLEDGED in one sweep!
	serverHealthy.Store(true)
	dueAgain := time.Now().Add(-10 * time.Second)
	_ = db.Model(&model.OutboundDeliveryQueue{}).Where("id IN ?", []uint64{item1.ID, item2.ID}).Update("next_retry_at", &dueAgain).Error

	svc.ProcessAutoRetryOnce()

	_ = db.First(&refreshed1, item1.ID).Error
	_ = db.First(&refreshed2, item2.ID).Error
	_ = db.First(&refreshedPerm, itemPerm.ID).Error

	if refreshed1.Status != model.DeliveryAcknowledged {
		t.Fatalf("expected item1 ACKNOWLEDGED after recovery, got %s", refreshed1.Status)
	}
	if refreshed2.Status != model.DeliveryAcknowledged {
		t.Fatalf("expected item2 ACKNOWLEDGED in same recovery sweep, got %s", refreshed2.Status)
	}
	if refreshedPerm.Status != model.DeliveryFailed || refreshedPerm.Attempts != 1 {
		t.Fatalf("permanent HTTP 400 must remain FAILED (attempts=1), got status=%s attempts=%d", refreshedPerm.Status, refreshedPerm.Attempts)
	}
	_ = ctx
}

func TestIsTransientDeliveryFailure(t *testing.T) {
	if !IsTransientDeliveryFailure(0, "network error: Post https://cloud.example: dial tcp timeout") {
		t.Fatal("network error (HTTP 0) must be transient")
	}
	if !IsTransientDeliveryFailure(502, "HTTP error 502: Bad Gateway") {
		t.Fatal("HTTP 502 must be transient")
	}
	if !IsTransientDeliveryFailure(429, "HTTP error 429: Too Many Requests") {
		t.Fatal("HTTP 429 must be transient")
	}
	if IsTransientDeliveryFailure(400, "HTTP error 400: bad payload") {
		t.Fatal("HTTP 400 must be permanent")
	}
	if IsTransientDeliveryFailure(401, "HTTP error 401: unauthorized") {
		t.Fatal("HTTP 401 must be permanent")
	}
	if IsTransientDeliveryFailure(422, "HTTP error 422: validation failed") {
		t.Fatal("HTTP 422 must be permanent")
	}
}
