package notification

import (
	"context"

	"datalogger/internal/model"
)

// NotificationMessage encapsulates the normalized payload dispatched to providers
type NotificationMessage struct {
	ID        uint
	AlarmID   uint
	Severity  model.AlarmSeverity
	AlarmCode string
	Device    string
	Parameter string
	Subject   string
	Body      string
	Recipient string
}

// DeliveryResult records provider response and failure metadata
type DeliveryResult struct {
	Success          bool
	ProviderResponse string
	ErrorMsg         string
	Retryable        bool
}

// ChannelProvider defines the unified interface implemented by delivery adapters
type ChannelProvider interface {
	ChannelType() model.NotificationChannelType
	ValidateConfig(configJSON string) error
	Send(ctx context.Context, msg NotificationMessage, configJSON string) DeliveryResult
}
