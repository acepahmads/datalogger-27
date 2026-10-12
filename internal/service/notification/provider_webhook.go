package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"datalogger/internal/model"
)

// WebhookConfig models generic HTTP webhook integration settings
type WebhookConfig struct {
	WebhookURL string            `json:"webhook_url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	AuthToken  string            `json:"auth_token,omitempty"`
}

// WebhookProvider manages HTTP webhook notification dispatch
type WebhookProvider struct {
	httpClient HTTPClientDo
}

// NewWebhookProvider constructs a WebhookProvider
func NewWebhookProvider() *WebhookProvider {
	client := &http.Client{Timeout: 10 * time.Second}
	return &WebhookProvider{
		httpClient: client.Do,
	}
}

// SetCustomHTTPClient allows injecting mock HTTP handlers for testing
func (p *WebhookProvider) SetCustomHTTPClient(do HTTPClientDo) {
	p.httpClient = do
}

func (p *WebhookProvider) ChannelType() model.NotificationChannelType {
	return model.ChannelWebhook
}

func (p *WebhookProvider) parseConfig(configJSON string) (*WebhookConfig, error) {
	if strings.TrimSpace(configJSON) == "" {
		return nil, fmt.Errorf("webhook configuration is empty")
	}

	var cfg WebhookConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid webhook configuration JSON: %w", err)
	}

	if cfg.Method == "" {
		cfg.Method = http.MethodPost
	}

	return &cfg, nil
}

func (p *WebhookProvider) ValidateConfig(configJSON string) error {
	cfg, err := p.parseConfig(configJSON)
	if err != nil {
		return err
	}
	if strings.TrimSpace(cfg.WebhookURL) == "" {
		return fmt.Errorf("webhook_url is required")
	}
	return nil
}

func (p *WebhookProvider) Send(ctx context.Context, msg NotificationMessage, configJSON string) DeliveryResult {
	cfg, err := p.parseConfig(configJSON)
	if err != nil || strings.TrimSpace(cfg.WebhookURL) == "" {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  "Webhook URL not configured",
			Retryable: false,
		}
	}

	targetURL := cfg.WebhookURL
	if strings.TrimSpace(msg.Recipient) != "" && (strings.HasPrefix(msg.Recipient, "http://") || strings.HasPrefix(msg.Recipient, "https://")) {
		targetURL = msg.Recipient
	}

	// Mock / Test loopback
	if targetURL == "mock" || targetURL == "test" || strings.HasPrefix(targetURL, "http://mock") {
		return DeliveryResult{
			Success:          true,
			ProviderResponse: "Mock Webhook accepted notification",
		}
	}

	payload := map[string]interface{}{
		"notification_id": msg.ID,
		"alarm_id":        msg.AlarmID,
		"severity":        msg.Severity,
		"alarm_code":      msg.AlarmCode,
		"device":          msg.Device,
		"parameter":       msg.Parameter,
		"subject":         msg.Subject,
		"message":         msg.Body,
		"timestamp":       time.Now().UTC().Format(time.RFC3339),
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  fmt.Sprintf("Failed to marshal webhook payload: %v", err),
			Retryable: false,
		}
	}

	req, err := http.NewRequestWithContext(ctx, cfg.Method, targetURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  fmt.Sprintf("Failed to construct webhook request: %v", err),
			Retryable: false,
		}
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Datalogger-Notification-Engine/1.0")

	for k, v := range cfg.Headers {
		req.Header.Set(k, v)
	}
	if cfg.AuthToken != "" {
		req.Header.Set("Authorization", "Bearer "+cfg.AuthToken)
	}

	resp, err := p.httpClient(req)
	if err != nil {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  fmt.Sprintf("Webhook HTTP request failed: %v", err),
			Retryable: true,
		}
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		isRetryable := (resp.StatusCode == 429 || resp.StatusCode >= 500)
		return DeliveryResult{
			Success:          false,
			ErrorMsg:         fmt.Sprintf("Webhook returned HTTP %d: %s", resp.StatusCode, string(respBytes)),
			ProviderResponse: string(respBytes),
			Retryable:        isRetryable,
		}
	}

	return DeliveryResult{
		Success:          true,
		ProviderResponse: fmt.Sprintf("HTTP %d: %s", resp.StatusCode, string(respBytes)),
	}
}
