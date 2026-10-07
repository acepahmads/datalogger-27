package communication

import (
	"context"
	"fmt"
	"sync"
	"time"

	"datalogger/internal/logger"
	"datalogger/internal/model"
	"datalogger/internal/service"
)

// AdapterFactory creates ProtocolAdapters based on connection configuration
type AdapterFactory func(cfg *model.DeviceConnection) (ProtocolAdapter, error)

// ConnectionManager manages all protocol connections and their lifecycles per device
type ConnectionManager struct {
	mu            sync.RWMutex
	adapters      map[uint]ProtocolAdapter
	deviceService *service.DeviceService
	factory       AdapterFactory
}

// NewConnectionManager initializes the communication connection manager
func NewConnectionManager(deviceService *service.DeviceService, factory AdapterFactory) *ConnectionManager {
	return &ConnectionManager{
		adapters:      make(map[uint]ProtocolAdapter),
		deviceService: deviceService,
		factory:       factory,
	}
}

// GetOrCreateAdapter retrieves an active adapter or instantiates one for the device
func (m *ConnectionManager) GetOrCreateAdapter(device *model.Device) (ProtocolAdapter, error) {
	if device == nil {
		return nil, fmt.Errorf("device cannot be nil")
	}
	if device.Connection == nil {
		return nil, fmt.Errorf("device %s (%d) has no connection configuration", device.DeviceCode, device.ID)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if adapter, exists := m.adapters[device.ID]; exists {
		// Check if protocol or target has changed
		existingCfg := adapter.GetConfig()
		if existingCfg != nil &&
			existingCfg.Protocol == device.Connection.Protocol &&
			existingCfg.Host == device.Connection.Host &&
			existingCfg.Port == device.Connection.Port &&
			existingCfg.SerialPort == device.Connection.SerialPort &&
			existingCfg.BaudRate == device.Connection.BaudRate &&
			existingCfg.SlaveID == device.Connection.SlaveID {
			return adapter, nil
		}
		// Config changed, close existing
		_ = adapter.Disconnect()
		delete(m.adapters, device.ID)
	}

	adapter, err := m.factory(device.Connection)
	if err != nil {
		return nil, fmt.Errorf("failed to create protocol adapter for device %s: %w", device.DeviceCode, err)
	}

	m.adapters[device.ID] = adapter
	return adapter, nil
}

// ConnectDevice connects a specific device adapter
func (m *ConnectionManager) ConnectDevice(ctx context.Context, device *model.Device) error {
	adapter, err := m.GetOrCreateAdapter(device)
	if err != nil {
		if m.deviceService != nil {
			_ = m.deviceService.SetConnectionError(device.ID, err.Error())
		}
		return err
	}

	timeoutMs := device.Connection.Timeout
	if timeoutMs <= 0 {
		timeoutMs = 1000
	}
	connCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	if err := adapter.Connect(connCtx); err != nil {
		if m.deviceService != nil {
			_ = m.deviceService.SetConnectionError(device.ID, err.Error())
		}
		return err
	}

	if m.deviceService != nil {
		_ = m.deviceService.SetOnline(device.ID)
	}
	return nil
}

// DisconnectDevice disconnects the device adapter gracefully
func (m *ConnectionManager) DisconnectDevice(deviceID uint) error {
	m.mu.Lock()
	adapter, exists := m.adapters[deviceID]
	m.mu.Unlock()

	if !exists {
		if m.deviceService != nil {
			_ = m.deviceService.SetOffline(deviceID)
		}
		return nil
	}

	err := adapter.Disconnect()
	if m.deviceService != nil {
		_ = m.deviceService.SetOffline(deviceID)
	}
	return err
}

// ReconnectDevice executes a clean disconnection and reconnection cycle
func (m *ConnectionManager) ReconnectDevice(ctx context.Context, device *model.Device) error {
	_ = m.DisconnectDevice(device.ID)
	time.Sleep(100 * time.Millisecond)
	return m.ConnectDevice(ctx, device)
}

// ExecuteWithRetry reads registers with configurable retry count and backoff
func (m *ConnectionManager) ExecuteWithRetry(
	ctx context.Context,
	device *model.Device,
	req ModbusReadRequest,
) (*ModbusReadResponse, error) {
	adapter, err := m.GetOrCreateAdapter(device)
	if err != nil {
		return nil, err
	}

	retryLimit := device.Connection.RetryCount
	if retryLimit < 0 {
		retryLimit = 0
	}
	if retryLimit > 5 {
		retryLimit = 5
	}

	var lastErr error
	for attempt := 0; attempt <= retryLimit; attempt++ {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		resp, err := adapter.ReadRegisters(ctx, req)
		if err == nil {
			// Successful read
			if m.deviceService != nil {
				_ = m.deviceService.SetOnline(device.ID)
				_ = m.deviceService.UpdateLastSeen(device.ID)
				_ = m.deviceService.RecordCommunicationResult(device.ID, true, int(resp.ResponseTime.Milliseconds()))
			}
			return resp, nil
		}

		lastErr = err
		if attempt < retryLimit {
			logger.Debug("Device %s read failed (attempt %d/%d): %v. Retrying...",
				device.DeviceCode, attempt+1, retryLimit+1, err)
			time.Sleep(time.Duration(50*(attempt+1)) * time.Millisecond)
		}
	}

	// All retries failed
	if m.deviceService != nil {
		_ = m.deviceService.SetConnectionError(device.ID, lastErr.Error())
		_ = m.deviceService.RecordCommunicationResult(device.ID, false, 0)
	}
	return nil, fmt.Errorf("device %s communication failed after %d retries: %w", device.DeviceCode, retryLimit, lastErr)
}

// TestConnection performs an isolated connection test
func (m *ConnectionManager) TestConnection(ctx context.Context, device *model.Device) (int, error) {
	adapter, err := m.GetOrCreateAdapter(device)
	if err != nil {
		return 0, err
	}

	start := time.Now()
	timeoutMs := device.Connection.Timeout
	if timeoutMs <= 0 {
		timeoutMs = 1500
	}
	testCtx, cancel := context.WithTimeout(ctx, time.Duration(timeoutMs)*time.Millisecond)
	defer cancel()

	if !adapter.IsConnected() {
		if err := adapter.Connect(testCtx); err != nil {
			return 0, err
		}
	}

	if err := adapter.HealthCheck(testCtx); err != nil {
		return int(time.Since(start).Milliseconds()), fmt.Errorf("health check failed: %w", err)
	}

	elapsed := int(time.Since(start).Milliseconds())
	return elapsed, nil
}

// GetAdapterStatus returns live status of the adapter
func (m *ConnectionManager) GetAdapterStatus(deviceID uint) (AdapterStatus, bool) {
	m.mu.RLock()
	adapter, exists := m.adapters[deviceID]
	m.mu.RUnlock()

	if !exists {
		return AdapterStatus{State: StateDisconnected}, false
	}
	return adapter.GetStatus(), true
}

// CloseAll closes all active connections
func (m *ConnectionManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	for id, adapter := range m.adapters {
		_ = adapter.Disconnect()
		delete(m.adapters, id)
	}
}
