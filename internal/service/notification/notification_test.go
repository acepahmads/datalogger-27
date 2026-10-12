package notification_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/smtp"
	"strings"
	"testing"

	"datalogger/internal/model"
	"datalogger/internal/service/notification"
)

func TestSanitizeConfigAndRedaction(t *testing.T) {
	rawJSON := `{
		"smtp_host": "smtp.gmail.com",
		"smtp_port": 587,
		"username": "alert@example.com",
		"password": "super-secret-password",
		"bot_token": "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11",
		"api_key": "my-api-key-9999",
		"secret_token": "topsecret",
		"from_email": "sender@example.com"
	}`

	sanitized := notification.SanitizeConfig(rawJSON)

	// Verify non-sensitive fields are preserved
	if sanitized["smtp_host"] != "smtp.gmail.com" {
		t.Errorf("expected smtp_host 'smtp.gmail.com', got '%v'", sanitized["smtp_host"])
	}
	if sanitized["from_email"] != "sender@example.com" {
		t.Errorf("expected from_email 'sender@example.com', got '%v'", sanitized["from_email"])
	}

	// Verify sensitive fields are masked
	sensitiveFields := []string{"password", "bot_token", "api_key", "secret_token"}
	for _, field := range sensitiveFields {
		val, ok := sanitized[field]
		if !ok || val != notification.MaskedSecret {
			t.Errorf("field %s was not redacted properly: %v", field, val)
		}
	}
}

func TestMergeSecrets(t *testing.T) {
	existingJSON := `{"smtp_host":"smtp.mail.com","password":"original-password-123","bot_token":"token-xyz"}`

	// User sends updated host and retains masked secrets
	incomingJSON := `{"smtp_host":"smtp.newmail.com","password":"********","bot_token":"********"}`
	merged := notification.MergeSecrets(incomingJSON, existingJSON)

	if !strings.Contains(merged, "original-password-123") {
		t.Errorf("MergeSecrets failed to preserve original password: %s", merged)
	}
	if !strings.Contains(merged, "token-xyz") {
		t.Errorf("MergeSecrets failed to preserve original token: %s", merged)
	}
	if !strings.Contains(merged, "smtp.newmail.com") {
		t.Errorf("MergeSecrets did not update host: %s", merged)
	}

	// User intentionally rotates password
	incomingNewSecret := `{"smtp_host":"smtp.newmail.com","password":"brand-new-secret"}`
	mergedNew := notification.MergeSecrets(incomingNewSecret, existingJSON)
	if !strings.Contains(mergedNew, "brand-new-secret") {
		t.Errorf("MergeSecrets failed to update rotated secret: %s", mergedNew)
	}
}

func TestEmailProvider_SuccessAndFailure(t *testing.T) {
	provider := notification.NewEmailProvider()

	// Validation
	if err := provider.ValidateConfig(`{}`); err == nil {
		t.Errorf("expected validation error for empty smtp_host")
	}

	validCfg := `{"smtp_host":"smtp.example.com","smtp_port":587,"username":"user","password":"pw","from_email":"no-reply@test.com","recipients":["ops@test.com"]}`
	if err := provider.ValidateConfig(validCfg); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}

	// Test unconfigured send
	unconfiguredRes := provider.Send(context.Background(), notification.NotificationMessage{Subject: "Test"}, "")
	if unconfiguredRes.Success || unconfiguredRes.Retryable {
		t.Errorf("expected unconfigured send to fail non-retryable")
	}

	// Mock successful SMTP send
	var sentAddr string
	var sentTo []string
	provider.SetCustomSender(func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		sentAddr = addr
		sentTo = to
		return nil
	})

	msg := notification.NotificationMessage{
		AlarmID:   10,
		Severity:  model.SeverityCritical,
		AlarmCode: "ALM_CRIT_HIGH",
		Subject:   "[CRITICAL] Temperature High",
		Body:      "Sensor temperature exceeds 90C",
		Recipient: "ops1@test.com, ops2@test.com",
	}

	res := provider.Send(context.Background(), msg, validCfg)
	if !res.Success {
		t.Fatalf("expected successful send, got error: %s", res.ErrorMsg)
	}
	if sentAddr != "smtp.example.com:587" {
		t.Errorf("unexpected sentAddr: %s", sentAddr)
	}
	if len(sentTo) != 2 || sentTo[0] != "ops1@test.com" || sentTo[1] != "ops2@test.com" {
		t.Errorf("unexpected recipients: %v", sentTo)
	}

	// Mock SMTP transient network failure -> should be retryable
	provider.SetCustomSender(func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		return errors.New("connection timeout to mail server")
	})

	failRes := provider.Send(context.Background(), msg, validCfg)
	if failRes.Success {
		t.Fatalf("expected failure, got success")
	}
	if !failRes.Retryable {
		t.Errorf("expected network failure to be retryable")
	}
}

func TestTelegramProvider_SuccessAndFailure(t *testing.T) {
	provider := notification.NewTelegramProvider()

	// Validation
	if err := provider.ValidateConfig(`{}`); err == nil {
		t.Errorf("expected validation error for empty bot_token")
	}
	validCfg := `{"bot_token":"12345:TOKEN","chat_id":"-100123456"}`
	if err := provider.ValidateConfig(validCfg); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}

	// Test unconfigured
	unconfiguredRes := provider.Send(context.Background(), notification.NotificationMessage{Subject: "Test"}, "")
	if unconfiguredRes.Success {
		t.Errorf("expected unconfigured send to fail")
	}

	// Test successful Telegram Send
	provider.SetCustomHTTPClient(func(req *http.Request) (*http.Response, error) {
		if !strings.Contains(req.URL.Path, "sendMessage") {
			t.Errorf("unexpected request path: %s", req.URL.Path)
		}
		body := `{"ok":true,"result":{"message_id":99}}`
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}, nil
	})

	msg := notification.NotificationMessage{
		AlarmID:   12,
		Severity:  model.SeverityWarning,
		AlarmCode: "WARN_LOW",
		Subject:   "[WARNING] Pressure Low",
		Body:      "Tank pressure below threshold",
	}

	res := provider.Send(context.Background(), msg, validCfg)
	if !res.Success {
		t.Fatalf("expected telegram success, got: %s", res.ErrorMsg)
	}

	// Test Telegram permanent client error (HTTP 400 Bad Request) -> Not retryable
	provider.SetCustomHTTPClient(func(req *http.Request) (*http.Response, error) {
		body := `{"ok":false,"error_code":400,"description":"Chat not found"}`
		return &http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(bytes.NewBufferString(body)),
			Header:     make(http.Header),
		}, nil
	})

	res400 := provider.Send(context.Background(), msg, validCfg)
	if res400.Success {
		t.Fatalf("expected telegram failure on 400")
	}
	if res400.Retryable {
		t.Errorf("400 error should not be retryable")
	}

	// Test Telegram transient server error (HTTP 502 Bad Gateway) -> Retryable
	provider.SetCustomHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusBadGateway,
			Body:       io.NopCloser(bytes.NewBufferString("Bad Gateway")),
			Header:     make(http.Header),
		}, nil
	})

	res502 := provider.Send(context.Background(), msg, validCfg)
	if !res502.Retryable {
		t.Errorf("502 error should be retryable")
	}
}

func TestWebhookProvider_SuccessAndFailure(t *testing.T) {
	provider := notification.NewWebhookProvider()

	// Validation
	if err := provider.ValidateConfig(`{}`); err == nil {
		t.Errorf("expected validation error for empty webhook_url")
	}
	validCfg := `{"webhook_url":"https://hooks.slack.com/services/xxx/yyy","method":"POST","headers":{"X-App":"Datalogger"}}`
	if err := provider.ValidateConfig(validCfg); err != nil {
		t.Errorf("unexpected validation error: %v", err)
	}

	// Mock Webhook HTTP 200 Success
	provider.SetCustomHTTPClient(func(req *http.Request) (*http.Response, error) {
		if req.Header.Get("X-App") != "Datalogger" {
			t.Errorf("missing custom header")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(bytes.NewBufferString(`{"status":"received"}`)),
			Header:     make(http.Header),
		}, nil
	})

	msg := notification.NotificationMessage{
		AlarmID:   15,
		Severity:  model.SeverityCritical,
		AlarmCode: "DEV_OFFLINE",
		Subject:   "[CRITICAL] Device Offline",
		Body:      "Generator 1 offline for > 15s",
	}

	res := provider.Send(context.Background(), msg, validCfg)
	if !res.Success {
		t.Fatalf("expected webhook success, got: %s", res.ErrorMsg)
	}

	// Mock Webhook HTTP 503 Service Unavailable -> Retryable
	provider.SetCustomHTTPClient(func(req *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode: http.StatusServiceUnavailable,
			Body:       io.NopCloser(bytes.NewBufferString(`Service Unavailable`)),
			Header:     make(http.Header),
		}, nil
	})

	res503 := provider.Send(context.Background(), msg, validCfg)
	if !res503.Retryable {
		t.Errorf("expected 503 to be retryable")
	}
}

func TestWebProvider(t *testing.T) {
	provider := notification.NewWebProvider(nil)

	if provider.ChannelType() != model.ChannelWeb {
		t.Errorf("unexpected channel type: %v", provider.ChannelType())
	}
	if err := provider.ValidateConfig("{}"); err != nil {
		t.Errorf("web provider config validation should succeed: %v", err)
	}

	msg := notification.NotificationMessage{
		AlarmID:   1,
		Severity:  model.SeverityWarning,
		AlarmCode: "WARN_LOW",
		Subject:   "Warning",
		Body:      "Body",
	}
	res := provider.Send(context.Background(), msg, "{}")
	if !res.Success {
		t.Errorf("expected web provider send to succeed, got: %s", res.ErrorMsg)
	}
}
