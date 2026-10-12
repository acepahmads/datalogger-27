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

// TelegramConfig models Telegram Bot API credentials
type TelegramConfig struct {
	BotToken    string `json:"bot_token"`
	ChatID      string `json:"chat_id"`
	APIEndpoint string `json:"api_endpoint,omitempty"`
}

// HTTPClientDo defines the interface for HTTP requests
type HTTPClientDo func(req *http.Request) (*http.Response, error)

// TelegramProvider manages Telegram Bot message dispatch
type TelegramProvider struct {
	httpClient HTTPClientDo
}

// NewTelegramProvider constructs a TelegramProvider
func NewTelegramProvider() *TelegramProvider {
	client := &http.Client{Timeout: 10 * time.Second}
	return &TelegramProvider{
		httpClient: client.Do,
	}
}

// SetCustomHTTPClient allows injecting mock HTTP handlers for testing
func (p *TelegramProvider) SetCustomHTTPClient(do HTTPClientDo) {
	p.httpClient = do
}

func (p *TelegramProvider) ChannelType() model.NotificationChannelType {
	return model.ChannelTelegram
}

func (p *TelegramProvider) parseConfig(configJSON string) (*TelegramConfig, error) {
	if strings.TrimSpace(configJSON) == "" {
		return nil, fmt.Errorf("telegram configuration is empty")
	}

	var cfg TelegramConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid telegram configuration JSON: %w", err)
	}

	if cfg.APIEndpoint == "" {
		cfg.APIEndpoint = "https://api.telegram.org"
	}
	cfg.APIEndpoint = strings.TrimRight(cfg.APIEndpoint, "/")

	return &cfg, nil
}

func (p *TelegramProvider) ValidateConfig(configJSON string) error {
	cfg, err := p.parseConfig(configJSON)
	if err != nil {
		return err
	}
	if strings.TrimSpace(cfg.BotToken) == "" {
		return fmt.Errorf("bot_token is required")
	}
	return nil
}

func (p *TelegramProvider) Send(ctx context.Context, msg NotificationMessage, configJSON string) DeliveryResult {
	cfg, err := p.parseConfig(configJSON)
	if err != nil || strings.TrimSpace(cfg.BotToken) == "" {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  "Telegram bot token not configured",
			Retryable: false,
		}
	}

	chatID := msg.Recipient
	if strings.TrimSpace(chatID) == "" {
		chatID = cfg.ChatID
	}
	if strings.TrimSpace(chatID) == "" {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  "No recipient chat ID specified",
			Retryable: false,
		}
	}

	// Mock / Test loopback
	if cfg.BotToken == "mock" || cfg.BotToken == "test" {
		return DeliveryResult{
			Success:          true,
			ProviderResponse: fmt.Sprintf("Mock Telegram accepted message for chat %s", chatID),
		}
	}

	// Format Telegram message
	emoji := "⚠️"
	if msg.Severity == model.SeverityCritical {
		emoji = "🚨"
	}
	text := fmt.Sprintf("%s *[%s ALARM]* %s\n\n*Device:* `%s`\n*Parameter:* `%s`\n*Message:* %s",
		emoji, msg.Severity, msg.AlarmCode, msg.Device, msg.Parameter, msg.Body)

	payload := map[string]interface{}{
		"chat_id":    chatID,
		"text":       text,
		"parse_mode": "Markdown",
	}
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  fmt.Sprintf("Failed to marshal telegram payload: %v", err),
			Retryable: false,
		}
	}

	apiURL := fmt.Sprintf("%s/bot%s/sendMessage", cfg.APIEndpoint, cfg.BotToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  fmt.Sprintf("Failed to create telegram request: %v", err),
			Retryable: false,
		}
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := p.httpClient(req)
	if err != nil {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  fmt.Sprintf("Telegram HTTP error: %v", err),
			Retryable: true,
		}
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		isRetryable := (resp.StatusCode == 429 || resp.StatusCode >= 500)
		return DeliveryResult{
			Success:          false,
			ErrorMsg:         fmt.Sprintf("Telegram API returned HTTP %d: %s", resp.StatusCode, string(respBytes)),
			ProviderResponse: string(respBytes),
			Retryable:        isRetryable,
		}
	}

	return DeliveryResult{
		Success:          true,
		ProviderResponse: string(respBytes),
	}
}
