package notification

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"

	"datalogger/internal/model"
)

// EmailConfig models SMTP connection credentials and sender details
type EmailConfig struct {
	SMTPHost   string   `json:"smtp_host"`
	SMTPPort   int      `json:"smtp_port"`
	Username   string   `json:"username"`
	Password   string   `json:"password"`
	FromEmail  string   `json:"from_email"`
	UseTLS     bool     `json:"use_tls"`
	Recipients []string `json:"recipients"`
}

// SMTPSender is a pluggable interface for sending mail, allowing test fixtures to mock network I/O
type SMTPSender func(addr string, a smtp.Auth, from string, to []string, msg []byte) error

// EmailProvider manages SMTP email notification dispatch
type EmailProvider struct {
	sender SMTPSender
}

// NewEmailProvider constructs an EmailProvider
func NewEmailProvider() *EmailProvider {
	return &EmailProvider{
		sender: smtp.SendMail,
	}
}

// SetCustomSender allows injecting test mocks
func (p *EmailProvider) SetCustomSender(sender SMTPSender) {
	p.sender = sender
}

func (p *EmailProvider) ChannelType() model.NotificationChannelType {
	return model.ChannelEmail
}

func (p *EmailProvider) parseConfig(configJSON string) (*EmailConfig, error) {
	if strings.TrimSpace(configJSON) == "" {
		return nil, fmt.Errorf("email configuration is empty")
	}

	var cfg EmailConfig
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return nil, fmt.Errorf("invalid email configuration JSON: %w", err)
	}

	if cfg.SMTPPort <= 0 {
		cfg.SMTPPort = 587
	}
	if cfg.FromEmail == "" {
		cfg.FromEmail = "datalogger@local"
	}

	return &cfg, nil
}

func (p *EmailProvider) ValidateConfig(configJSON string) error {
	cfg, err := p.parseConfig(configJSON)
	if err != nil {
		return err
	}
	if strings.TrimSpace(cfg.SMTPHost) == "" {
		return fmt.Errorf("smtp_host is required")
	}
	return nil
}

func (p *EmailProvider) Send(ctx context.Context, msg NotificationMessage, configJSON string) DeliveryResult {
	cfg, err := p.parseConfig(configJSON)
	if err != nil || strings.TrimSpace(cfg.SMTPHost) == "" {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  "SMTP server not configured",
			Retryable: false,
		}
	}

	// Resolve recipient list
	var toList []string
	if strings.TrimSpace(msg.Recipient) != "" {
		parts := strings.Split(msg.Recipient, ",")
		for _, part := range parts {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				toList = append(toList, trimmed)
			}
		}
	}
	if len(toList) == 0 {
		toList = cfg.Recipients
	}
	if len(toList) == 0 {
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  "No recipient email specified",
			Retryable: false,
		}
	}

	// Mock / Test loopback
	if cfg.SMTPHost == "mock" || cfg.SMTPHost == "test" || cfg.SMTPHost == "localhost.mock" {
		return DeliveryResult{
			Success:          true,
			ProviderResponse: fmt.Sprintf("Mock SMTP accepted email for %d recipient(s)", len(toList)),
		}
	}

	// Build MIME Message
	subject := msg.Subject
	if subject == "" {
		subject = fmt.Sprintf("[%s] Alarm Alert: %s", msg.Severity, msg.AlarmCode)
	}

	header := make(map[string]string)
	header["From"] = cfg.FromEmail
	header["To"] = strings.Join(toList, ", ")
	header["Subject"] = subject
	header["MIME-Version"] = "1.0"
	header["Content-Type"] = "text/plain; charset=\"utf-8\""

	var messageStr string
	for k, v := range header {
		messageStr += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	messageStr += "\r\n" + msg.Body

	addr := fmt.Sprintf("%s:%d", cfg.SMTPHost, cfg.SMTPPort)

	var auth smtp.Auth
	if cfg.Username != "" && cfg.Password != "" {
		auth = smtp.PlainAuth("", cfg.Username, cfg.Password, cfg.SMTPHost)
	}

	// Channel for synchronous wait with context timeout
	done := make(chan error, 1)
	go func() {
		if cfg.UseTLS && cfg.SMTPPort == 465 {
			// Direct TLS connection
			tlsConfig := &tls.Config{
				ServerName: cfg.SMTPHost,
			}
			dialer := &net.Dialer{Timeout: 10 * time.Second}
			conn, dErr := tls.DialWithDialer(dialer, "tcp", addr, tlsConfig)
			if dErr != nil {
				done <- dErr
				return
			}
			client, cErr := smtp.NewClient(conn, cfg.SMTPHost)
			if cErr != nil {
				_ = conn.Close()
				done <- cErr
				return
			}
			defer client.Quit()

			if auth != nil {
				if aErr := client.Auth(auth); aErr != nil {
					done <- aErr
					return
				}
			}
			if fErr := client.Mail(cfg.FromEmail); fErr != nil {
				done <- fErr
				return
			}
			for _, recipient := range toList {
				if rErr := client.Rcpt(recipient); rErr != nil {
					done <- rErr
					return
				}
			}
			w, wErr := client.Data()
			if wErr != nil {
				done <- wErr
				return
			}
			_, _ = w.Write([]byte(messageStr))
			done <- w.Close()
		} else {
			// Standard SMTP / STARTTLS
			done <- p.sender(addr, auth, cfg.FromEmail, toList, []byte(messageStr))
		}
	}()

	select {
	case <-ctx.Done():
		return DeliveryResult{
			Success:   false,
			ErrorMsg:  fmt.Sprintf("SMTP delivery timed out: %v", ctx.Err()),
			Retryable: true,
		}
	case err := <-done:
		if err != nil {
			isRetryable := true
			errStr := strings.ToLower(err.Error())
			if strings.Contains(errStr, "authentication") || strings.Contains(errStr, "invalid recipient") {
				isRetryable = false
			}
			return DeliveryResult{
				Success:   false,
				ErrorMsg:  fmt.Sprintf("SMTP delivery error: %v", err),
				Retryable: isRetryable,
			}
		}
		return DeliveryResult{
			Success:          true,
			ProviderResponse: fmt.Sprintf("250 OK: Message accepted for delivery to %s", strings.Join(toList, ", ")),
		}
	}
}
