package service_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
