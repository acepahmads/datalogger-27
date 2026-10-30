package handler_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"datalogger/internal/model"
)

func TestAuditTrail_FullNameAndHumanReadableDetails(t *testing.T) {
	router, db, _, _, _, _ := setupPackageBTestEnv(t)

	// Create a real named operator user in users table
	namedUser := model.User{
		Username: "teknisi_lapangan",
		Email:    "teknisi@cbi.id",
		Password: "hashed_password",
		FullName: "Budi Santoso",
		IsActive: true,
	}
	if err := db.Create(&namedUser).Error; err != nil {
		t.Fatalf("failed seeding named user: %v", err)
	}

	now := time.Now().UTC()

	// 1. Double-encoded UPDATE_PARAMETER JSON (exact production shape)
	doubleEncodedParamUpdate := `{"after":"{\"enabled\":true,\"offset\":0,\"parameter_code\":\"pm2.5\",\"parameter_name\":\"PM 2.5\",\"quality_validation_enabled\":true,\"scale\":1,\"stale_timeout_seconds\":60,\"warning_high\":null,\"warning_low\":1}","before":"{\"enabled\":true,\"offset\":0,\"parameter_code\":\"pm2.5\",\"parameter_name\":\"PM 2.5\",\"quality_validation_enabled\":true,\"scale\":1,\"stale_timeout_seconds\":60,\"warning_high\":null,\"warning_low\":null}","device_id":5,"parameter_id":12}`

	// 2. Toggle parameter JSON
	toggleParamJSON := `{"device_id":5,"enabled":false,"parameter_code":"hc","parameter_id":14}`

	// 3. Alarm cleared & triggered strings from production
	alarmClearedStr := `Alarm 24855 (THRESHOLD_WARNING_LOW) cleared`
	alarmTriggeredStr := `Alarm triggered: Code=THRESHOLD_WARNING_LOW, Severity=WARNING, Message=PM 2.5 (com4:pm2.5) dropped below warning low limit: 0.00 ug/m3 (Threshold: 1.00 ug/m3) [Quality: UNCERTAIN - ZERO_VALUE]`

	// 4. Generic JSON with secret key
	genericSecretJSON := `{"api_key":"ultra_secret_key_999","host":"192.168.1.53"}`

	entries := []model.AuditTrail{
		{
			UserID:    &namedUser.ID,
			Username:  namedUser.Username,
			Action:    "UPDATE_PARAMETER",
			Resource:  "PARAMETER",
			Details:   doubleEncodedParamUpdate,
			IPAddress: "192.168.1.60",
			CreatedAt: now.Add(-4 * time.Minute),
		},
		{
			Username:  namedUser.Username, // Without UserID -> resolved via username lookup
			Action:    "UPDATE_PARAMETER",
			Resource:  "PARAMETER",
			Details:   toggleParamJSON,
			IPAddress: "192.168.1.60",
			CreatedAt: now.Add(-3 * time.Minute),
		},
		{
			Username:  "system",
			Action:    "ALARM_CLEAR",
			Resource:  "ALARM",
			Details:   alarmClearedStr,
			CreatedAt: now.Add(-2 * time.Minute),
		},
		{
			Username:  "system",
			Action:    "ALARM_CREATE",
			Resource:  "ALARM",
			Details:   alarmTriggeredStr,
			CreatedAt: now.Add(-1 * time.Minute),
		},
		{
			UserID:    &namedUser.ID,
			Username:  namedUser.Username,
			Action:    "UPDATE_OUTBOUND",
			Resource:  "INTEGRATION",
			Details:   genericSecretJSON,
			CreatedAt: now,
		},
	}
	for i := range entries {
		if err := db.Create(&entries[i]).Error; err != nil {
			t.Fatalf("failed seeding audit entry %d: %v", i, err)
		}
	}

	req, _ := http.NewRequest("GET", "/api/audit-trails?page=1&page_size=20", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Data struct {
			Items []model.AuditTrail `json:"items"`
			Total int64              `json:"total"`
		} `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed unmarshaling response: %v", err)
	}

	if len(resp.Data.Items) < 5 {
		t.Fatalf("expected at least 5 audit trail items, got %d", len(resp.Data.Items))
	}

	for _, item := range resp.Data.Items {
		// Verify zero raw JSON braces leak in Details
		if strings.Contains(item.Details, "{") || strings.Contains(item.Details, "}") {
			t.Errorf("expected plain human-readable details (no JSON braces), got: %s", item.Details)
		}
		// Verify secret never leaks
		if strings.Contains(item.Details, "ultra_secret_key_999") {
			t.Errorf("secret leaked in formatted details: %s", item.Details)
		}

		switch {
		case item.Username == "teknisi_lapangan" && item.Action == "UPDATE_PARAMETER" && strings.Contains(item.Details, "PM 2.5"):
			if item.FullName != "Budi Santoso" {
				t.Errorf("expected full_name 'Budi Santoso', got '%s'", item.FullName)
			}
			if !strings.Contains(item.Details, "Batas Peringatan Minimum diatur menjadi 1") {
				t.Errorf("expected readable diff on warning_low, got: %s", item.Details)
			}

		case item.Username == "teknisi_lapangan" && item.Action == "UPDATE_PARAMETER" && strings.Contains(item.Details, "hc"):
			if item.FullName != "Budi Santoso" {
				t.Errorf("expected full_name 'Budi Santoso' via username lookup, got '%s'", item.FullName)
			}
			if !strings.Contains(item.Details, "Menonaktifkan pembacaan parameter sensor hc") {
				t.Errorf("expected human toggle message, got: %s", item.Details)
			}

		case item.Username == "system" && item.Action == "ALARM_CLEAR":
			if item.FullName != "Sistem Otomatis" {
				t.Errorf("expected full_name 'Sistem Otomatis' for system actor, got '%s'", item.FullName)
			}
			if !strings.Contains(item.Details, "Kondisi normal kembali: Alarm #24855") {
				t.Errorf("expected human alarm clear message, got: %s", item.Details)
			}

		case item.Username == "system" && item.Action == "ALARM_CREATE":
			if !strings.Contains(item.Details, "Muncul alarm Peringatan") || !strings.Contains(item.Details, "turun di bawah batas peringatan minimum") {
				t.Errorf("expected human alarm create message, got: %s", item.Details)
			}

		case item.Action == "UPDATE_OUTBOUND":
			if !strings.Contains(item.Details, "[REDACTED]") {
				t.Errorf("expected [REDACTED] for sensitive JSON key, got: %s", item.Details)
			}
		}
	}

	// Verify filtering by Nama User ("Budi Santoso") works on Actor filter
	reqByFullName, _ := http.NewRequest("GET", "/api/audit-trails?username=Budi+Santoso&page=1&page_size=20", nil)
	wByFullName := httptest.NewRecorder()
	router.ServeHTTP(wByFullName, reqByFullName)

	var fullNameResp struct {
		Data struct {
			Items []model.AuditTrail `json:"items"`
			Total int64              `json:"total"`
		} `json:"data"`
	}
	_ = json.Unmarshal(wByFullName.Body.Bytes(), &fullNameResp)
	if fullNameResp.Data.Total != 3 {
		t.Errorf("expected 3 audit records when filtering actor by full name 'Budi Santoso', got %d", fullNameResp.Data.Total)
	}
}
