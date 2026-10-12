package notification

import (
	"context"
	"fmt"
	"time"

	"datalogger/internal/model"
	"datalogger/internal/websocket"
)

// WebProvider handles web, in-app, and WebSocket notification delivery
type WebProvider struct {
	hub *websocket.Hub
}

// NewWebProvider constructs a web delivery provider
func NewWebProvider(hub *websocket.Hub) *WebProvider {
	return &WebProvider{hub: hub}
}

func (p *WebProvider) ChannelType() model.NotificationChannelType {
	return model.ChannelWeb
}

func (p *WebProvider) ValidateConfig(configJSON string) error {
	// Web provider requires no mandatory external connection credentials
	return nil
}

func (p *WebProvider) Send(ctx context.Context, msg NotificationMessage, configJSON string) DeliveryResult {
	if p.hub != nil {
		p.hub.Broadcast(websocket.WSMessage{
			Type:      "notification.alert",
			Timestamp: time.Now().UTC(),
			Data: map[string]interface{}{
				"id":         msg.ID,
				"alarm_id":   msg.AlarmID,
				"severity":   msg.Severity,
				"alarm_code": msg.AlarmCode,
				"device":     msg.Device,
				"parameter":  msg.Parameter,
				"subject":    msg.Subject,
				"message":    msg.Body,
				"recipient":  msg.Recipient,
			},
		})
	}

	return DeliveryResult{
		Success:          true,
		ProviderResponse: fmt.Sprintf("Delivered to in-app WebSocket hub (clients notified)"),
	}
}
