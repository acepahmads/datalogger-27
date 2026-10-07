package modbus

import (
	"testing"
)

func TestModbusCRC16(t *testing.T) {
	// Standard Modbus RTU example: Read Holding Registers
	// Slave: 0x01, FC: 0x03, Start: 0x0000, Qty: 0x000A
	// PDU = [0x01, 0x03, 0x00, 0x00, 0x00, 0x0A]
	// Expected CRC: 0xC5CD (in RTU frame: 0xC5 Low, 0xCD High)
	data := []byte{0x01, 0x03, 0x00, 0x00, 0x00, 0x0A}
	crc := CalculateCRC16(data)
	expectedCRC := uint16(0xCDC5) // Low: 0xC5, High: 0xCD -> uint16 is 0xCDC5

	if crc != expectedCRC {
		t.Fatalf("expected CRC 0x%04X, got 0x%04X", expectedCRC, crc)
	}

	// Test AppendCRC16
	adu := AppendCRC16(data)
	if len(adu) != 8 {
		t.Fatalf("expected ADU length 8, got %d", len(adu))
	}
	if adu[6] != 0xC5 || adu[7] != 0xCD {
		t.Fatalf("expected CRC bytes [0xC5, 0xCD], got [0x%02X, 0x%02X]", adu[6], adu[7])
	}

	// Test ValidateCRC16
	if !ValidateCRC16(adu) {
		t.Fatalf("ValidateCRC16 failed for valid frame")
	}

	// Corrupted frame
	corrupted := make([]byte, len(adu))
	copy(corrupted, adu)
	corrupted[3] ^= 0xFF
	if ValidateCRC16(corrupted) {
		t.Fatalf("ValidateCRC16 should fail for corrupted frame")
	}

	// Corrupted CRC byte
	corruptedCRC := make([]byte, len(adu))
	copy(corruptedCRC, adu)
	corruptedCRC[7] ^= 0x01
	if ValidateCRC16(corruptedCRC) {
		t.Fatalf("ValidateCRC16 should fail when CRC byte altered")
	}
}
