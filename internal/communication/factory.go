package communication

import (
	"fmt"

	"datalogger/internal/communication/modbus"
	"datalogger/internal/model"
)

// DefaultAdapterFactory constructs production ProtocolAdapters (Modbus TCP and Modbus RTU)
func DefaultAdapterFactory(cfg *model.DeviceConnection) (ProtocolAdapter, error) {
	if cfg == nil {
		return nil, fmt.Errorf("connection configuration cannot be nil")
	}

	switch cfg.Protocol {
	case model.ProtocolModbusTCP, model.ProtocolTCP:
		return modbus.NewModbusTCPAdapter(cfg), nil

	case model.ProtocolModbusRTU, model.ProtocolSerial:
		return modbus.NewModbusRTUAdapter(cfg), nil

	default:
		// Default fallback to Modbus TCP
		return modbus.NewModbusTCPAdapter(cfg), nil
	}
}
