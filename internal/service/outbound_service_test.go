package service_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/repository"
	"datalogger/internal/service"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func setupOutboundTestDB(t *testing.T) *gorm.DB {
	dsn := fmt.Sprintf("file:memoutbound_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
	})
	if err != nil {
		t.Fatalf("failed to connect memory sqlite: %v", err)
	}

	err = db.AutoMigrate(
		&model.DeviceType{},
		&model.Device{},
		&model.DeviceConnection{},
		&model.ParameterCategory{},
		&model.Parameter{},
		&model.AggregationDefinition{},
		&model.AggregationResult{},
		&model.OutputDestination{},
		&model.DestinationPayloadConfig{},
		&model.OutboundDeliveryQueue{},
		&model.DeliveryLog{},
	)
	if err != nil {
		t.Fatalf("failed to migrate models: %v", err)
	}

	return db
}

func TestOutboundService_AggregationSourcedOnly(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	formatter := service.NewConfigurablePayloadFormatter()
	svc := service.NewOutboundService(outboundRepo, deviceRepo, formatter)

	ctx := context.Background()

	// 1. Create Device
	dev := &model.Device{
		ID:         1,
		DeviceCode: "AQMS-001",
		DeviceName: "Station Central",
	}
	db.Create(dev)

	// 2. Create Parameters
	param1 := &model.Parameter{
		ID:            1,
		DeviceID:      1,
		ParameterCode: "pm25",
		ParameterName: "Particulate PM2.5",
		Unit:          "µg/m3",
	}
	param2 := &model.Parameter{
		ID:            2,
		DeviceID:      1,
		ParameterCode: "so2",
		ParameterName: "Sulfur Dioxide",
		Unit:          "ppm",
	}
	db.Create(param1)
	db.Create(param2)

	// 3. Create Aggregation Results:
	// One is CUSTOMER_PROCESSED (MUST be picked up)
	// One is INTERNAL_RAW (MUST NEVER be picked up by outbound pipeline)
	val1 := 35.8
	valRaw := 999.9
	now := time.Now().Truncate(time.Minute)

	db.Create(&model.AggregationResult{
		ID:                      101,
		AggregationDefinitionID: 1,
		DeviceID:                1,
		ParameterID:             1,
		SourceType:              model.SourceCustomerProcessed,
		PeriodStart:             now.Add(-5 * time.Minute),
		PeriodEnd:               now,
		IntervalSeconds:         300,
		Value:                   &val1,
		Quality:                 "GOOD",
		Identifier:              now.Format("20060102150405"),
	})

	db.Create(&model.AggregationResult{
		ID:                      102,
		AggregationDefinitionID: 2,
		DeviceID:                1,
		ParameterID:             2,
		SourceType:              model.SourceInternalRaw, // INTERNAL RAW!
		PeriodStart:             now.Add(-5 * time.Minute),
		PeriodEnd:               now,
		IntervalSeconds:         300,
		Value:                   &valRaw,
		Quality:                 "GOOD",
		Identifier:              now.Format("20060102150405"),
	})

	// 4. Create an enabled destination
	dest := &model.OutputDestination{
		Name:            "Central Monitoring Cloud",
		Type:            model.DestREST,
		Endpoint:        "http://example.com/webhook",
		Enabled:         true,
		IntervalSeconds: 300,
		RetryLimit:      3,
	}
	if err := svc.CreateDestination(ctx, dest); err != nil {
		t.Fatalf("failed to create destination: %v", err)
	}

	// 5. Trigger bucket completion
	err := svc.HandleBucketCompleted(ctx, 1, now)
	if err != nil {
		t.Fatalf("HandleBucketCompleted error: %v", err)
	}

	// 6. Inspect enqueued items
	items, total, err := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if err != nil || total != 1 {
		t.Fatalf("expected 1 enqueued item, got %d, err: %v", total, err)
	}

	item := items[0]
	// Verify that the payload snapshot contains PM25 but strictly DOES NOT contain internal raw value (999.9 or so2)
	if !strings.Contains(item.PayloadSnapshot, "35.8") {
		t.Errorf("expected customer processed value 35.8 in payload snapshot, got: %s", item.PayloadSnapshot)
	}
	if strings.Contains(item.PayloadSnapshot, "999.9") {
		t.Fatalf("CRITICAL SECURITY VIOLATION: Internal raw value 999.9 was leaked into outbound payload!")
	}

	// 7. Verify Idempotency: Triggering the same bucket again must NOT create duplicate queue entries
	err = svc.HandleBucketCompleted(ctx, 1, now)
	if err != nil {
		t.Fatalf("second HandleBucketCompleted error: %v", err)
	}
	_, totalAfter, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if totalAfter != 1 {
		t.Errorf("expected still 1 queue item due to idempotency, got %d", totalAfter)
	}

	// 8. Test Replay creates an explicit new delivery with is_replay = true
	replayItem, err := svc.ReplayQueueItem(ctx, item.ID)
	if err != nil {
		t.Fatalf("failed to replay queue item: %v", err)
	}
	if !replayItem.IsReplay {
		t.Errorf("expected replayItem.IsReplay to be true")
	}
	if replayItem.ReplayOfID == nil || *replayItem.ReplayOfID != item.ID {
		t.Errorf("expected replayOfID to equal original item ID")
	}
}

func TestOutboundService_RetryAndPermanentError(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	formatter := service.NewConfigurablePayloadFormatter()
	svc := service.NewOutboundService(outboundRepo, deviceRepo, formatter)

	ctx := context.Background()

	// Setup mock HTTP server that simulates responses
	var mockStatus int
	var mockBody string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(mockStatus)
		w.Write([]byte(mockBody))
	}))
	defer ts.Close()

	dest := &model.OutputDestination{
		Name:       "Test Webhook",
		Type:       model.DestREST,
		Endpoint:   ts.URL,
		Enabled:    true,
		RetryLimit: 3,
	}
	_ = svc.CreateDestination(ctx, dest)

	// 1. Test Permanent 400 Bad Request
	mockStatus = 400
	mockBody = `{"error": "Invalid schema"}`

	testRes, err := svc.TestPayload(ctx, dest.ID)
	if err != nil {
		t.Fatalf("TestPayload returned unexpected error: %v", err)
	}
	if testRes.Success {
		t.Errorf("expected test delivery to fail on HTTP 400")
	}
	if testRes.HTTPCode != 400 {
		t.Errorf("expected HTTP 400, got %d", testRes.HTTPCode)
	}

	// 2. Test Success with Acknowledgement
	mockStatus = 200
	mockBody = `{"status": "success", "ack": true, "receipt": "RC-9988"}`

	testResSuccess, err := svc.TestPayload(ctx, dest.ID)
	if err != nil || !testResSuccess.Success {
		t.Fatalf("expected test delivery to succeed on HTTP 200, got err: %v, success: %v", err, testResSuccess.Success)
	}
}

func TestOutboundService_CredentialRedaction(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	svc := service.NewOutboundService(outboundRepo, deviceRepo, nil)

	ctx := context.Background()

	dest := &model.OutputDestination{
		Name:        "Secure Cloud Endpoint",
		Type:        model.DestREST,
		Endpoint:    "https://api.cloud.example/v1",
		AuthType:    "BEARER_TOKEN",
		AuthToken:   "super-secret-token-12345",
		APIKeyValue: "secret-api-key-999",
		Password:    "secret-password",
		Enabled:     true,
	}
	_ = svc.CreateDestination(ctx, dest)

	// Non-superadmin view: credentials MUST be redacted
	redactedDests, err := svc.GetDestinations(ctx, false, true)
	if err != nil || len(redactedDests) == 0 {
		t.Fatalf("failed to retrieve redacted destinations: %v", err)
	}

	if redactedDests[0].AuthToken != "***REDACTED***" {
		t.Errorf("expected AuthToken to be redacted, got: %s", redactedDests[0].AuthToken)
	}
	if redactedDests[0].APIKeyValue != "***REDACTED***" {
		t.Errorf("expected APIKeyValue to be redacted, got: %s", redactedDests[0].APIKeyValue)
	}
	if redactedDests[0].Password != "***REDACTED***" {
		t.Errorf("expected Password to be redacted, got: %s", redactedDests[0].Password)
	}

	// Superadmin view: actual secrets available for worker/superadmin configuration
	rawDests, _ := svc.GetDestinations(ctx, false, false)
	if rawDests[0].AuthToken != "super-secret-token-12345" {
		t.Errorf("expected raw AuthToken for internal engine/superadmin, got: %s", rawDests[0].AuthToken)
	}
}

func TestOutboundService_PurgeQueue_AllConditions(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	svc := service.NewOutboundService(outboundRepo, deviceRepo, nil)

	ctx := context.Background()

	// Seed queue items and logs
	for i := 1; i <= 5; i++ {
		q := &model.OutboundDeliveryQueue{
			ID:              uint64(i),
			BatchID:         fmt.Sprintf("BATCH-%d", i),
			DestinationID:   1,
			PeriodStart:     time.Now(),
			PeriodEnd:       time.Now(),
			IntervalSeconds: 300,
			PayloadSnapshot: "{}",
			IdempotencyKey:  fmt.Sprintf("KEY-%d", i),
			Status:          model.DeliveryFailed,
		}
		if err := db.Create(q).Error; err != nil {
			t.Fatalf("failed to seed queue: %v", err)
		}
		log := &model.DeliveryLog{
			ID:            uint(i),
			DestinationID: 1,
			Status:        "FAILED",
			DeliveredAt:   time.Now(),
		}
		if err := db.Create(log).Error; err != nil {
			t.Fatalf("failed to seed log: %v", err)
		}
	}

	// Purge ALL without any destination or status filter (this must NOT trigger 'WHERE conditions required')
	deletedQueue, deletedLogs, err := svc.PurgeQueue(ctx, repository.QueuePurgeFilter{
		Status:    "ALL",
		PurgeLogs: true,
	})
	if err != nil {
		t.Fatalf("PurgeQueue failed: %v", err)
	}
	if deletedQueue != 5 {
		t.Errorf("expected 5 queue items deleted, got %d", deletedQueue)
	}
	if deletedLogs != 5 {
		t.Errorf("expected 5 logs deleted, got %d", deletedLogs)
	}

	// Verify database is empty
	var remainingQueue int64
	db.Model(&model.OutboundDeliveryQueue{}).Count(&remainingQueue)
	if remainingQueue != 0 {
		t.Errorf("expected 0 remaining queue items, got %d", remainingQueue)
	}

	var remainingLogs int64
	db.Model(&model.DeliveryLog{}).Count(&remainingLogs)
	if remainingLogs != 0 {
		t.Errorf("expected 0 remaining logs, got %d", remainingLogs)
	}
}

func TestOutboundService_TwoStepLoginBearer(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	formatter := service.NewConfigurablePayloadFormatter()
	svc := service.NewOutboundService(outboundRepo, deviceRepo, formatter)

	ctx := context.Background()

	var authHits int32
	var dataHits int32
	var currentToken atomic.Value
	currentToken.Store("jwt-token-v1")

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth":
			atomic.AddInt32(&authHits, 1)
			var creds map[string]string
			if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if creds["username"] != "pt_cakrawala_bima_instrument" || creds["password"] != "secret123" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message":"invalid credentials"}`))
				return
			}
			tok := currentToken.Load().(string)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf(`{"status":"ok","data":{"token":%q}}`, tok)))

		case "/api/v1/aqmdata":
			atomic.AddInt32(&dataHits, 1)
			expectedAuth := "Bearer " + currentToken.Load().(string)
			if r.Header.Get("Authorization") != expectedAuth {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"message":"expired or invalid token"}`))
				return
			}
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			if body["id_stasiun"] == "" || body["waktu"] == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`{"message":"missing id_stasiun or waktu"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":"success","ack":true}`))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer ts.Close()

	// 1. Mode strict 2x send on every batch (AuthCacheSeconds = 0)
	dest := &model.OutputDestination{
		Name:             "KLHK ISPU AQMS",
		Type:             model.DestREST,
		Endpoint:         ts.URL + "/api/v1/aqmdata",
		AuthType:         "LOGIN_BEARER",
		AuthURL:          ts.URL + "/api/v1/auth",
		Username:         "pt_cakrawala_bima_instrument",
		Password:         "secret123",
		AuthCacheSeconds: 0,
		Enabled:          true,
	}
	if err := svc.CreateDestination(ctx, dest); err != nil {
		t.Fatalf("CreateDestination failed: %v", err)
	}

	// First send -> 1 auth + 1 data
	res1, err := svc.TestPayload(ctx, dest.ID)
	if err != nil || !res1.Success {
		t.Fatalf("first 2-step TestPayload failed: err=%v, res=%+v", err, res1)
	}
	// Second send -> because AuthCacheSeconds == 0, performs 2nd auth + 2nd data
	res2, err := svc.TestPayload(ctx, dest.ID)
	if err != nil || !res2.Success {
		t.Fatalf("second 2-step TestPayload failed: err=%v, res=%+v", err, res2)
	}
	if atomic.LoadInt32(&authHits) != 2 || atomic.LoadInt32(&dataHits) != 2 {
		t.Errorf("expected 2 auth hits and 2 data hits in uncached 2-step mode, got auth=%d data=%d", authHits, dataHits)
	}

	// 2. Mode cached token (AuthCacheSeconds = 900) + automatic 401 re-auth retry
	dest.AuthCacheSeconds = 900
	dest.Password = "***REDACTED***" // verify redacted password preservation on UpdateDestination too!
	if err := svc.UpdateDestination(ctx, dest); err != nil {
		t.Fatalf("UpdateDestination failed: %v", err)
	}

	atomic.StoreInt32(&authHits, 0)
	atomic.StoreInt32(&dataHits, 0)

	// Send #1 -> caches jwt-token-v1 (auth=1, data=1)
	if r, err := svc.TestPayload(ctx, dest.ID); err != nil || !r.Success {
		t.Fatalf("cached send #1 failed: err=%v, res=%+v", err, r)
	}
	// Send #2 -> reuses cached jwt-token-v1 without hitting /auth (auth=1, data=2)
	if r, err := svc.TestPayload(ctx, dest.ID); err != nil || !r.Success {
		t.Fatalf("cached send #2 failed: err=%v, res=%+v", err, r)
	}
	if atomic.LoadInt32(&authHits) != 1 || atomic.LoadInt32(&dataHits) != 2 {
		t.Errorf("expected 1 auth hit and 2 data hits with caching, got auth=%d data=%d", authHits, dataHits)
	}

	// Simulate upstream token rotation / expiration -> /aqmdata returns 401 on stale token -> worker auto re-logins and retries!
	currentToken.Store("jwt-token-v2")
	if r, err := svc.TestPayload(ctx, dest.ID); err != nil || !r.Success {
		t.Fatalf("expected automatic re-auth retry on 401 to succeed, got err=%v, res=%+v", err, r)
	}
	if atomic.LoadInt32(&authHits) != 2 || atomic.LoadInt32(&dataHits) != 4 {
		t.Errorf("expected auth=2 and data=4 after 401 re-login retry, got auth=%d data=%d", authHits, dataHits)
	}
}
