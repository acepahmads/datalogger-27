package service_test

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
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

func TestOutboundService_TwoStepHTTP200ApplicationRejection(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	formatter := service.NewConfigurablePayloadFormatter()
	svc := service.NewOutboundService(outboundRepo, deviceRepo, formatter)

	ctx := context.Background()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/auth":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":1,"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.validtoken"}`))
		case "/api/v1/aqmdata":
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if body["id_stasiun"] != "KABUPATEN_JOMBANG" {
				// KemenLH returns HTTP 200 with status:0 when vendor has no access to id_stasiun
				w.Write([]byte(`{"status":0,"message":"Vendor tidak memiliki akses ke stasiun ini"}`))
				return
			}
			w.Write([]byte(`{"status":1,"message":"Data berhasil disimpan"}`))
		}
	}))
	defer ts.Close()

	dest := &model.OutputDestination{
		Name:     "KLH",
		Type:     model.DestREST,
		Endpoint: ts.URL + "/api/v1/aqmdata",
		AuthType: "LOGIN_BEARER",
		AuthURL:  ts.URL + "/api/v1/auth",
		Username: "pt_cakrawala_bima_instrument",
		Password: "secret",
		Enabled:  true,
	}
	_ = svc.CreateDestination(ctx, dest)

	// 1. Draft config with unauthorized station -> Step 1 OK, Step 2 MUST fail despite HTTP 200!
	draftUnauthorized := &model.DestinationPayloadConfig{
		DestinationID:          dest.ID,
		TemplateType:           "KLHK_ISPU_AQMS",
		StationIDField:         "id_stasiun",
		StationIDValue:         "KABUPATEN_SUKABUMI",
		TimestampField:         "waktu",
		TimestampFormat:        "CUSTOM",
		TimestampTimezone:      "Asia/Jakarta",
		OmitIntervalField:      true,
		Structure:              "FLAT",
		ConflictResolution:     "PRIORITY_PRIMARY",
		IncludeSensorStatFlags: true,
	}
	resFail, err := svc.TestPayloadWithConfig(ctx, dest.ID, draftUnauthorized)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resFail.IsTwoStep || !resFail.AuthSuccess {
		t.Errorf("expected IsTwoStep=true and AuthSuccess=true, got %+v", resFail)
	}
	if resFail.Success {
		t.Errorf("expected Success=false when KemenLH responds HTTP 200 status=0, got Success=true")
	}
	if !strings.Contains(resFail.ErrorMessage, "Vendor tidak memiliki akses ke stasiun ini") {
		t.Errorf("expected rejection message in ErrorMessage, got %q", resFail.ErrorMessage)
	}

	// 2. Draft config with authorized station -> Step 1 OK, Step 2 OK!
	draftAuthorized := *draftUnauthorized
	draftAuthorized.StationIDValue = "KABUPATEN_JOMBANG"
	resOK, err := svc.TestPayloadWithConfig(ctx, dest.ID, &draftAuthorized)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resOK.Success || !resOK.AuthSuccess {
		t.Errorf("expected both Step 1 and Step 2 to succeed for authorized station, got %+v", resOK)
	}
}

func TestOutboundService_AutoRebuildAndRequeueOnConfigAndCredentialFix(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	svc := service.NewOutboundService(outboundRepo, deviceRepo, nil)

	ctx := context.Background()

	// Seed active device + parameter + 30m CUSTOMER_PROCESSED aggregation
	db.Create(&model.Device{
		ID:         1,
		DeviceCode: "com4",
		DeviceName: "AQMS Sensor",
		Enabled:    true,
		Status:     model.DeviceStatusActive,
	})
	db.Create(&model.Parameter{
		ID:            1,
		DeviceID:      1,
		ParameterCode: "pm2.5",
		ParameterName: "PM 2.5",
		Unit:          "ug/m3",
		Enabled:       true,
	})

	pmVal := 41.68
	bucketEnd := time.Date(2026, 10, 30, 1, 30, 0, 0, time.UTC) // 08:30:00 WIB
	db.Create(&model.AggregationResult{
		ID:                      501,
		AggregationDefinitionID: 1,
		DeviceID:                1,
		ParameterID:             1,
		SourceType:              model.SourceCustomerProcessed,
		PeriodStart:             bucketEnd.Add(-30 * time.Minute),
		PeriodEnd:               bucketEnd,
		IntervalSeconds:         1800,
		Value:                   &pmVal,
		Quality:                 "GOOD",
		Identifier:              bucketEnd.Format("20060102150405"),
	})

	// Mock KemenLH 2-step server checking username/password, custom header, and id_stasiun
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Vendor-Key") != "valid-header" {
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"status":0,"message":"Invalid X-Vendor-Key header"}`))
			return
		}
		switch r.URL.Path {
		case "/api/v1/auth":
			var creds map[string]string
			_ = json.NewDecoder(r.Body).Decode(&creds)
			if creds["username"] != "pt_cakrawala_bima_instrument" || creds["password"] != "hxvB61b1td1AjUuc" {
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"status":0,"message":"Invalid credentials"}`))
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"status":1,"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.validtoken"}`))
		case "/api/v1/aqmdata":
			var body map[string]interface{}
			_ = json.NewDecoder(r.Body).Decode(&body)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if body["id_stasiun"] != "KABUPATEN_JOMBANG" {
				w.Write([]byte(`{"status":0,"message":"Vendor tidak memiliki akses ke stasiun ini"}`))
				return
			}
			w.Write([]byte(`{"status":1,"message":"Data berhasil disimpan"}`))
		}
	}))
	defer ts.Close()

	// 1. Create destination initially with WRONG password & WRONG custom header
	dest := &model.OutputDestination{
		Name:            "KLH",
		Type:            model.DestREST,
		Endpoint:        ts.URL + "/api/v1/aqmdata",
		IntervalSeconds: 1800,
		AuthType:        "LOGIN_BEARER",
		AuthURL:         ts.URL + "/api/v1/auth",
		Username:        "pt_cakrawala_bima_instrument",
		Password:        "wrong_password",
		CustomHeaders:   `{"X-Vendor-Key":"wrong-header"}`,
		Enabled:         true,
		RetryLimit:      5,
	}
	if err := svc.CreateDestination(ctx, dest); err != nil {
		t.Fatalf("CreateDestination failed: %v", err)
	}

	// Set initial payload config with WRONG id_stasiun = KABUPATEN_SUKABUMI
	cfg, err := svc.GetPayloadConfig(ctx, dest.ID)
	if err != nil {
		t.Fatalf("GetPayloadConfig failed: %v", err)
	}
	cfg.StationIDValue = "KABUPATEN_SUKABUMI"
	if err := svc.UpdatePayloadConfig(ctx, cfg); err != nil {
		t.Fatalf("UpdatePayloadConfig initial failed: %v", err)
	}

	// 2. Bucket completes -> 1 batch enqueued with KABUPATEN_SUKABUMI
	if err := svc.HandleBucketCompleted(ctx, 1, bucketEnd); err != nil {
		t.Fatalf("HandleBucketCompleted failed: %v", err)
	}
	svc.ProcessPendingQueueOnce()

	items, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if len(items) != 1 || items[0].Status != model.DeliveryFailed {
		t.Fatalf("expected 1 FAILED item after bad header/password, got %+v", items)
	}
	if !strings.Contains(items[0].PayloadSnapshot, `"KABUPATEN_SUKABUMI"`) {
		t.Fatalf("expected initial snapshot to contain KABUPATEN_SUKABUMI, got %s", items[0].PayloadSnapshot)
	}

	// 3. Fix Destination header & password -> item MUST automatically re-queue to PENDING (attempts=0)!
	dest.Password = "hxvB61b1td1AjUuc"
	dest.CustomHeaders = `{"X-Vendor-Key":"valid-header"}`
	if err := svc.UpdateDestination(ctx, dest); err != nil {
		t.Fatalf("UpdateDestination failed: %v", err)
	}
	if dest.RequeuedBatches != 1 {
		t.Errorf("expected UpdateDestination to requeue 1 batch, got %d", dest.RequeuedBatches)
	}

	itemsAfterDestFix, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if itemsAfterDestFix[0].Status != model.DeliveryPending || itemsAfterDestFix[0].Attempts != 0 {
		t.Fatalf("expected item to be reset to PENDING with attempts=0 after destination fix, got status=%s attempts=%d",
			itemsAfterDestFix[0].Status, itemsAfterDestFix[0].Attempts)
	}

	// 4. Worker dispatches -> Step 1 Auth now succeeds, Step 2 fails with "Vendor tidak memiliki akses ke stasiun ini"
	svc.ProcessPendingQueueOnce()
	itemsAfterAuthOK, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if itemsAfterAuthOK[0].Status != model.DeliveryFailed || !strings.Contains(itemsAfterAuthOK[0].LastError, "Vendor tidak memiliki akses ke stasiun ini") {
		t.Fatalf("expected FAILED due to unauthorized id_stasiun, got status=%s err=%s",
			itemsAfterAuthOK[0].Status, itemsAfterAuthOK[0].LastError)
	}

	// 5. Fix Payload Config id_stasiun = KABUPATEN_JOMBANG -> PayloadSnapshot MUST auto-rebuild AND re-queue to PENDING!
	cfg.StationIDValue = "KABUPATEN_JOMBANG"
	if err := svc.UpdatePayloadConfig(ctx, cfg); err != nil {
		t.Fatalf("UpdatePayloadConfig fix failed: %v", err)
	}
	if cfg.RequeuedBatches != 1 {
		t.Errorf("expected UpdatePayloadConfig to rebuild & requeue 1 batch, got %d", cfg.RequeuedBatches)
	}

	itemsAfterPayloadFix, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	rebuilt := itemsAfterPayloadFix[0]
	if rebuilt.Status != model.DeliveryPending || rebuilt.Attempts != 0 {
		t.Fatalf("expected rebuilt item to be PENDING with attempts=0, got status=%s attempts=%d", rebuilt.Status, rebuilt.Attempts)
	}
	if strings.Contains(rebuilt.PayloadSnapshot, "KABUPATEN_SUKABUMI") || !strings.Contains(rebuilt.PayloadSnapshot, `"KABUPATEN_JOMBANG"`) {
		t.Fatalf("expected PayloadSnapshot to be updated to KABUPATEN_JOMBANG, got: %s", rebuilt.PayloadSnapshot)
	}
	if !strings.Contains(rebuilt.PayloadSnapshot, `"pm25":41.68`) {
		t.Errorf("expected persisted sensor value pm25:41.68 preserved in rebuilt payload, got: %s", rebuilt.PayloadSnapshot)
	}

	// 6. Worker dispatches -> Batch succeeds!
	svc.ProcessPendingQueueOnce()
	finalItems, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if finalItems[0].Status != model.DeliveryDelivered && finalItems[0].Status != model.DeliveryAcknowledged {
		t.Fatalf("expected batch to be DELIVERED/ACKNOWLEDGED after config fix, got status=%s err=%s",
			finalItems[0].Status, finalItems[0].LastError)
	}

	updatedDest, _ := svc.GetDestinationByID(ctx, dest.ID, true)
	if updatedDest.SuccessDeliveries != 1 || updatedDest.FailedDeliveries != 0 || updatedDest.Status != "ACTIVE" {
		t.Errorf("expected destination stats synced to Success=1 Failed=0 ACTIVE, got %+v", updatedDest)
	}
}

func TestOutboundService_SparingTwoStepJWT_HourlyAndTwoMinute(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	formatter := service.NewConfigurablePayloadFormatter()
	svc := service.NewOutboundService(outboundRepo, deviceRepo, formatter)

	ctx := context.Background()

	// Seed active SPARING water quality device & 5 parameters
	db.Create(&model.Device{
		ID:         1,
		DeviceCode: "WQ-01",
		DeviceName: "SPARING Water Quality",
		Enabled:    true,
		Status:     model.DeviceStatusActive,
	})
	sparingParams := []model.Parameter{
		{ID: 1, DeviceID: 1, ParameterCode: "ph", ParameterName: "pH", Unit: "", Enabled: true},
		{ID: 2, DeviceID: 1, ParameterCode: "cod", ParameterName: "COD", Unit: "mg/L", Enabled: true},
		{ID: 3, DeviceID: 1, ParameterCode: "tss", ParameterName: "TSS", Unit: "mg/L", Enabled: true},
		{ID: 4, DeviceID: 1, ParameterCode: "nh3n", ParameterName: "NH3-N", Unit: "mg/L", Enabled: true},
		{ID: 5, DeviceID: 1, ParameterCode: "debit", ParameterName: "Debit", Unit: "m3/s", Enabled: true},
	}
	for _, p := range sparingParams {
		pCopy := p
		db.Create(&pCopy)
	}

	// Seed 30 x 2-minute CUSTOMER_PROCESSED buckets: [1793588400 .. 1793591880], ending at 1793592000 (1793592000 % 3600 == 0)
	var aggID uint64 = 1000
	for i := 0; i < 30; i++ {
		pStart := time.Unix(1793588400+int64(i*120), 0).UTC()
		pEnd := pStart.Add(120 * time.Second)
		vals := []float64{6.93, 40.99, 27.95, 0.0, 0.24}
		if i == 1 {
			vals[1] = 41.00
			vals[2] = 27.94
		} else if i == 29 {
			vals[0] = 7.00
			vals[1] = 48.08
			vals[2] = 28.55
			vals[4] = 0.33
		}
		for pIdx, param := range sparingParams {
			aggID++
			v := vals[pIdx]
			db.Create(&model.AggregationResult{
				ID:                      aggID,
				AggregationDefinitionID: uint(param.ID),
				DeviceID:                1,
				ParameterID:             param.ID,
				SourceType:              model.SourceCustomerProcessed,
				PeriodStart:             pStart,
				PeriodEnd:               pEnd,
				IntervalSeconds:         120,
				Value:                   &v,
				Quality:                 "GOOD",
				Identifier:              fmt.Sprintf("%d_%d", param.ID, pEnd.Unix()),
			})
		}
	}

	var secretHits int32
	var sendHits int32
	var currentSecret atomic.Value
	currentSecret.Store("sparing-secret-key-2026")
	var lastDecodedPayload atomic.Value

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/secret-sensor":
			if r.Method != http.MethodGet {
				w.WriteHeader(http.StatusMethodNotAllowed)
				return
			}
			atomic.AddInt32(&secretHits, 1)
			w.Header().Set("Content-Type", "text/plain")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(currentSecret.Load().(string)))

		case "/api/send-redirect":
			// Simulate HTTP -> HTTPS 301 redirect on sparing.kemenlh.go.id
			http.Redirect(w, r, "/api/send", http.StatusMovedPermanently)

		case "/api/send":
			atomic.AddInt32(&sendHits, 1)
			if r.Method != http.MethodPost {
				w.WriteHeader(http.StatusMethodNotAllowed)
				_, _ = w.Write([]byte(`{"status":false,"desc":"Method must be POST"}`))
				return
			}
			var reqBody map[string]string
			if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil || reqBody["token"] == "" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"status":false,"desc":"Token kosong"}`))
				return
			}

			parts := strings.Split(reqBody["token"], ".")
			if len(parts) != 3 {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"status":false,"desc":"Format JWT tidak valid"}`))
				return
			}

			hdrBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			var hdr map[string]string
			_ = json.Unmarshal(hdrBytes, &hdr)
			if hdr["alg"] != "HS256" || hdr["typ"] != "JWT" {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"status":false,"desc":"Header JWT wajib HS256"}`))
				return
			}

			mac := hmac.New(sha256.New, []byte(currentSecret.Load().(string)))
			_, _ = mac.Write([]byte(parts[0] + "." + parts[1]))
			expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
			if parts[2] != expectedSig {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"status":false,"desc":"Signature JWT tidak valid"}`))
				return
			}

			payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			var decoded map[string]interface{}
			if err := json.Unmarshal(payloadBytes, &decoded); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if decoded["uid"] != "79456d9c51128700b78b6fd0" {
				_, _ = w.Write([]byte(`{"status":false,"desc":"UID tidak terdaftar pada sistem SPARING"}`))
				return
			}

			lastDecodedPayload.Store(string(payloadBytes))
			_, _ = w.Write([]byte(`{"status":true,"desc":null}`))
		}
	}))
	defer ts.Close()

	// 1. Create Hourly Bundled SPARING 2-Step destination (2m source -> 1h dispatch)
	dest := &model.OutputDestination{
		Name:                    "KLHK SPARING Hourly",
		Type:                    model.DestREST,
		Endpoint:                ts.URL + "/api/send-redirect",
		IntervalSeconds:         120,
		DispatchIntervalSeconds: 3600,
		AuthType:                "SPARING_JWT_2STEP",
		AuthURL:                 ts.URL + "/api/secret-sensor",
		AuthCacheSeconds:        600,
		Enabled:                 true,
		RetryLimit:              5,
	}
	if err := svc.CreateDestination(ctx, dest); err != nil {
		t.Fatalf("CreateDestination failed: %v", err)
	}

	cfg, err := svc.GetPayloadConfig(ctx, dest.ID)
	if err != nil {
		t.Fatalf("GetPayloadConfig failed: %v", err)
	}
	if cfg.TemplateType != "KLHK_SPARING_HOURLY" || cfg.Structure != "SPARING_DATA_ARRAY" {
		t.Fatalf("expected auto-initialized KLHK_SPARING_HOURLY / SPARING_DATA_ARRAY, got template=%s structure=%s",
			cfg.TemplateType, cfg.Structure)
	}

	// 2. Test rejection detection when UID is wrong ("status": false, "desc": "...")
	cfg.StationIDValue = "WRONG_UID"
	resReject, err := svc.TestPayloadWithConfig(ctx, dest.ID, cfg)
	if err != nil {
		t.Fatalf("TestPayloadWithConfig error: %v", err)
	}
	if !resReject.IsTwoStep || !resReject.AuthSuccess || resReject.Success {
		t.Fatalf("expected Step 1 OK and Step 2 rejected on wrong UID, got %+v", resReject)
	}
	if !strings.Contains(resReject.ErrorMessage, "UID tidak terdaftar") {
		t.Errorf("expected SPARING desc error message, got %q", resReject.ErrorMessage)
	}

	// Set valid UID and save config
	cfg.StationIDValue = "79456d9c51128700b78b6fd0"
	if err := svc.UpdatePayloadConfig(ctx, cfg); err != nil {
		t.Fatalf("UpdatePayloadConfig failed: %v", err)
	}

	// 3. Non-hourly 2-minute bucket completion (:02:00) MUST NOT enqueue anything yet
	if err := svc.HandleBucketCompleted(ctx, 1, time.Unix(1793588520, 0).UTC()); err != nil {
		t.Fatalf("HandleBucketCompleted non-hourly failed: %v", err)
	}
	qEarly, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if len(qEarly) != 0 {
		t.Fatalf("expected 0 queue items at non-hourly 2m mark, got %d", len(qEarly))
	}

	// 4. Hourly boundary completion (1793592000 % 3600 == 0) MUST bundle all 30 2-minute buckets into 1 queue item!
	if err := svc.HandleBucketCompleted(ctx, 1, time.Unix(1793592000, 0).UTC()); err != nil {
		t.Fatalf("HandleBucketCompleted hourly failed: %v", err)
	}
	qHourly, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if len(qHourly) != 1 {
		t.Fatalf("expected 1 bundled hourly queue item, got %d", len(qHourly))
	}
	if qHourly[0].ParameterCount != 150 {
		t.Errorf("expected ParameterCount=150 (30 buckets x 5 params), got %d", qHourly[0].ParameterCount)
	}

	// Dispatch hourly bundled item via worker (through 301 redirect!)
	svc.ProcessPendingQueueOnce()

	qAfterHourly, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if qAfterHourly[0].Status != model.DeliveryAcknowledged {
		t.Fatalf("expected ACKNOWLEDGED for hourly SPARING JWT delivery, got status=%s err=%s",
			qAfterHourly[0].Status, qAfterHourly[0].LastError)
	}

	decodedHourlyStr, _ := lastDecodedPayload.Load().(string)
	if !strings.HasPrefix(decodedHourlyStr, `{"uid":"79456d9c51128700b78b6fd0","data":[{"datetime":1793588400,"pH":6.93,"cod":40.99,"tss":27.95,"nh3n":0,"debit":0.24}`) {
		t.Errorf("unexpected decoded JWT payload prefix: %s", decodedHourlyStr)
	}

	// 5. Switch destination schedule to Immediate Per 2 Minutes (DispatchIntervalSeconds = 0) + rotate upstream secret
	currentSecret.Store("sparing-secret-rotated-v2")
	dest.DispatchIntervalSeconds = 0
	if err := svc.UpdateDestination(ctx, dest); err != nil {
		t.Fatalf("UpdateDestination to 2m immediate failed: %v", err)
	}

	cfgAfterSwitch, _ := svc.GetPayloadConfig(ctx, dest.ID)
	if cfgAfterSwitch.Structure != "FLAT" || cfgAfterSwitch.TemplateType != "KLHK_SPARING" {
		t.Fatalf("expected payload structure to auto-switch to FLAT / KLHK_SPARING, got structure=%s template=%s",
			cfgAfterSwitch.Structure, cfgAfterSwitch.TemplateType)
	}

	// Now 2-minute bucket completion (:02:00) immediately enqueues a single flat SPARING record!
	if err := svc.HandleBucketCompleted(ctx, 1, time.Unix(1793588520, 0).UTC()); err != nil {
		t.Fatalf("HandleBucketCompleted 2m immediate failed: %v", err)
	}
	svc.ProcessPendingQueueOnce()

	qAll, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if len(qAll) != 2 {
		t.Fatalf("expected 2 total queue items (1 hourly + 1 immediate 2m), got %d", len(qAll))
	}
	for _, item := range qAll {
		if item.Status != model.DeliveryAcknowledged {
			t.Errorf("expected queue item #%d to be ACKNOWLEDGED, got status=%s err=%s", item.ID, item.Status, item.LastError)
		}
	}

	decodedFlatStr, _ := lastDecodedPayload.Load().(string)
	expectedFlatDecoded := `{"uid":"79456d9c51128700b78b6fd0","datetime":1793588400,"pH":6.93,"cod":40.99,"tss":27.95,"nh3n":0,"debit":0.24}`
	if decodedFlatStr != expectedFlatDecoded {
		t.Errorf("decoded 2m flat JWT payload mismatch:\ngot:  %s\nwant: %s", decodedFlatStr, expectedFlatDecoded)
	}
}

func TestOutboundService_KLHKONLIMOBodyDeliveryAndRedaction(t *testing.T) {
	db := setupOutboundTestDB(t)
	outboundRepo := repository.NewOutboundRepository(db)
	deviceRepo := repository.NewDeviceRepository(db)
	formatter := service.NewConfigurablePayloadFormatter()
	svc := service.NewOutboundService(outboundRepo, deviceRepo, formatter)

	ctx := context.Background()

	// Seed WQMS station device with real-world channel names (including "Turbudity (NTU)", "Conductivity / DHL", "TOC")
	db.Create(&model.Device{
		ID:         1,
		DeviceCode: "WQMS-ONLIMO-01",
		DeviceName: "Stasiun WQMS ONLIMO",
		Enabled:    true,
		Status:     model.DeviceStatusActive,
	})
	wqmsParams := []model.Parameter{
		{ID: 1, DeviceID: 1, ParameterCode: "ch_1", ParameterName: "Temperature", Unit: "C", Enabled: true},
		{ID: 2, DeviceID: 1, ParameterCode: "ch_2", ParameterName: "Conductivity / DHL", Unit: "ms/cm", Enabled: true},
		{ID: 3, DeviceID: 1, ParameterCode: "ch_3", ParameterName: "TDS", Unit: "mg/L", Enabled: true},
		{ID: 4, DeviceID: 1, ParameterCode: "ch_4", ParameterName: "Salinity", Unit: "PSU", Enabled: true},
		{ID: 5, DeviceID: 1, ParameterCode: "ch_5", ParameterName: "DO", Unit: "mg/L", Enabled: true},
		{ID: 6, DeviceID: 1, ParameterCode: "ch_6", ParameterName: "PH", Unit: "", Enabled: true},
		{ID: 7, DeviceID: 1, ParameterCode: "ch_7", ParameterName: "Turbudity (NTU)", Unit: "ntu", Enabled: true},
		{ID: 8, DeviceID: 1, ParameterCode: "ch_8", ParameterName: "Depth", Unit: "m", Enabled: true},
		{ID: 9, DeviceID: 1, ParameterCode: "ch_9", ParameterName: "Nitrat", Unit: "mg/L", Enabled: true},
		{ID: 10, DeviceID: 1, ParameterCode: "ch_10", ParameterName: "Ammonia (NH3-N)", Unit: "mg/L", Enabled: true},
		{ID: 11, DeviceID: 1, ParameterCode: "ch_11", ParameterName: "COD", Unit: "mg/L", Enabled: true},
		{ID: 12, DeviceID: 1, ParameterCode: "ch_12", ParameterName: "BOD", Unit: "mg/L", Enabled: true},
		{ID: 13, DeviceID: 1, ParameterCode: "ch_13", ParameterName: "TSS", Unit: "mg/L", Enabled: true},
		{ID: 14, DeviceID: 1, ParameterCode: "ch_14", ParameterName: "TOC", Unit: "mg/L", Enabled: true},
	}
	for _, p := range wqmsParams {
		pCopy := p
		db.Create(&pCopy)
	}

	// PeriodEnd = 2026-11-04 06:00:00 UTC -> 2026-11-04 13:00:00 WIB
	pStart := time.Date(2026, 11, 4, 5, 0, 0, 0, time.UTC)
	pEnd := time.Date(2026, 11, 4, 6, 0, 0, 0, time.UTC)
	sensorVals := []float64{28.5, 385.88, 192.5, 3.88, 4.56, 6.97, 14.2, 1.85, 74.05, 84.03, 68.05, 56.04, 24.0, 3.22}
	for idx, param := range wqmsParams {
		v := sensorVals[idx]
		db.Create(&model.AggregationResult{
			ID:                      uint64(2000 + idx),
			AggregationDefinitionID: uint(param.ID),
			DeviceID:                1,
			ParameterID:             param.ID,
			SourceType:              model.SourceCustomerProcessed,
			PeriodStart:             pStart,
			PeriodEnd:               pEnd,
			IntervalSeconds:         3600,
			Value:                   &v,
			Quality:                 "GOOD",
			Identifier:              fmt.Sprintf("%d_%d", param.ID, pEnd.Unix()),
		})
	}

	var lastRawHTTPBody atomic.Value
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rawBytes, _ := io.ReadAll(r.Body)
		lastRawHTTPBody.Store(string(rawBytes))

		var body map[string]interface{}
		if err := json.Unmarshal(rawBytes, &body); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte(`{"status":false,"message":"Invalid JSON"}`))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		if body["apikey"] != "klhk-onlimo-apikey-999" || body["apisecret"] != "klhk-onlimo-secret-xyz" {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":{"statusCode":403,"statusDesc":"Forbidden"}}`))
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":{"statusCode":200,"statusDesc":"OK"}}`))
	}))
	defer ts.Close()

	// 1. Create destination with KLHK_ONLIMO_BODY (with surrounding whitespace to verify trim)
	dest := &model.OutputDestination{
		Name:            "KLHK WQMS ONLIMO",
		Type:            model.DestREST,
		Endpoint:        ts.URL + "/api/wqms/post",
		IntervalSeconds: 3600,
		AuthType:        "KLHK_ONLIMO_BODY",
		APIKeyValue:     "  klhk-onlimo-apikey-999  ",
		AuthToken:       "  klhk-onlimo-secret-xyz  ",
		Enabled:         true,
		RetryLimit:      5,
	}
	if err := svc.CreateDestination(ctx, dest); err != nil {
		t.Fatalf("CreateDestination failed: %v", err)
	}

	// Verify auto-initialized KLHK_WQMS_ONLIMO payload config
	cfg, err := svc.GetPayloadConfig(ctx, dest.ID)
	if err != nil {
		t.Fatalf("GetPayloadConfig failed: %v", err)
	}
	if cfg.TemplateType != "KLHK_WQMS_ONLIMO" || cfg.Structure != "ONLIMO_DATA_OBJECT" || !cfg.FillMissingPresetFields {
		t.Fatalf("expected auto-initialized KLHK_WQMS_ONLIMO + ONLIMO_DATA_OBJECT + FillMissingPresetFields=true, got %+v", cfg)
	}

	cfg.StationIDValue = "ONLIMO_ST_01"
	if err := svc.UpdatePayloadConfig(ctx, cfg); err != nil {
		t.Fatalf("UpdatePayloadConfig failed: %v", err)
	}

	// 2. Verify API redaction and safe round-trip UpdateDestination with ***REDACTED***
	redactedDest, err := svc.GetDestinationByID(ctx, dest.ID, true)
	if err != nil {
		t.Fatalf("GetDestinationByID failed: %v", err)
	}
	if redactedDest.APIKeyValue != "***REDACTED***" || redactedDest.AuthToken != "***REDACTED***" {
		t.Fatalf("expected both APIKeyValue and AuthToken redacted, got apikey=%q secret=%q",
			redactedDest.APIKeyValue, redactedDest.AuthToken)
	}
	redactedDest.Name = "KLHK WQMS ONLIMO Updated"
	if err := svc.UpdateDestination(ctx, redactedDest); err != nil {
		t.Fatalf("UpdateDestination with redacted secrets failed: %v", err)
	}

	// 3. Verify GeneratePreviewWithConfig masks secrets and never leaks raw credentials
	previewStr, err := svc.GeneratePreviewWithConfig(ctx, dest.ID, nil, cfg)
	if err != nil {
		t.Fatalf("GeneratePreviewWithConfig failed: %v", err)
	}
	if strings.Contains(previewStr, "klhk-onlimo-apikey-999") || strings.Contains(previewStr, "klhk-onlimo-secret-xyz") {
		t.Fatalf("SECURITY LEAK: GeneratePreviewWithConfig exposed raw ONLIMO credentials:\n%s", previewStr)
	}
	if !strings.Contains(previewStr, `"apikey": "***REDACTED***"`) || !strings.Contains(previewStr, `"apisecret": "***REDACTED***"`) {
		t.Errorf("expected masked ***REDACTED*** in preview JSON, got:\n%s", previewStr)
	}

	// 4. Verify TestPayloadWithConfig injects real credentials on wire while keeping PayloadSent redacted
	testRes, err := svc.TestPayloadWithConfig(ctx, dest.ID, cfg)
	if err != nil {
		t.Fatalf("TestPayloadWithConfig failed: %v", err)
	}
	if !testRes.Success {
		t.Fatalf("expected TestPayloadWithConfig to succeed, got error=%q body=%q", testRes.ErrorMessage, testRes.ResponseText)
	}
	if strings.Contains(testRes.PayloadSent, "klhk-onlimo-apikey-999") || strings.Contains(testRes.PayloadSent, "klhk-onlimo-secret-xyz") {
		t.Fatalf("SECURITY LEAK: TestPayloadResult.PayloadSent exposed raw credentials: %s", testRes.PayloadSent)
	}

	// 5. Enqueue bucket via HandleBucketCompleted & dispatch via ProcessPendingQueueOnce
	if err := svc.HandleBucketCompleted(ctx, 1, pEnd); err != nil {
		t.Fatalf("HandleBucketCompleted failed: %v", err)
	}
	qItems, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if len(qItems) != 1 {
		t.Fatalf("expected 1 queued ONLIMO item, got %d", len(qItems))
	}
	if strings.Contains(qItems[0].PayloadSnapshot, "klhk-onlimo-apikey-999") || strings.Contains(qItems[0].PayloadSnapshot, "klhk-onlimo-secret-xyz") {
		t.Fatalf("SECURITY LEAK: OutboundDeliveryQueue.PayloadSnapshot exposed raw credentials: %s", qItems[0].PayloadSnapshot)
	}

	svc.ProcessPendingQueueOnce()

	qAfter, _, _ := svc.GetQueueItems(ctx, repository.QueueFilter{DestinationID: &dest.ID})
	if qAfter[0].Status != model.DeliveryAcknowledged {
		t.Fatalf("expected ACKNOWLEDGED for ONLIMO statusCode=200 delivery, got status=%s err=%s", qAfter[0].Status, qAfter[0].LastError)
	}

	expectedWireJSON := `{"data":{"IDStasiun":"ONLIMO_ST_01","Tanggal":"2026-11-04","Jam":"13:00:00","Suhu":28.5,"DHL":385.88,"TDS":192.5,"Salinitas":3.88,"DO":4.56,"PH":6.97,"Turbidity":14.2,"Kedalaman":1.85,"Nitrat":74.05,"Amonia":84.03,"COD":68.05,"BOD":56.04,"TSS":24},"apikey":"klhk-onlimo-apikey-999","apisecret":"klhk-onlimo-secret-xyz"}`
	gotWireJSON, _ := lastRawHTTPBody.Load().(string)
	if gotWireJSON != expectedWireJSON {
		t.Fatalf("ONLIMO wire HTTP POST body mismatch:\ngot:  %s\nwant: %s", gotWireJSON, expectedWireJSON)
	}

	// 6. Verify application-level rejection when apisecret is wrong (HTTP 200 with {"status":{"statusCode":403,"statusDesc":"Forbidden"}})
	badDest := *redactedDest
	badDest.AuthToken = "wrong-secret"
	if err := svc.UpdateDestination(ctx, &badDest); err != nil {
		t.Fatalf("UpdateDestination bad secret failed: %v", err)
	}
	badTestRes, err := svc.TestPayloadWithConfig(ctx, dest.ID, cfg)
	if err != nil {
		t.Fatalf("TestPayloadWithConfig bad secret unexpected error: %v", err)
	}
	if badTestRes.Success || !strings.Contains(badTestRes.ErrorMessage, "403 Forbidden") {
		t.Fatalf("expected ONLIMO rejection '403 Forbidden' on wrong apisecret, got success=%v err=%q", badTestRes.Success, badTestRes.ErrorMessage)
	}
}


