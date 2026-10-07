package modbus

// CalculateCRC16 computes the 16-bit Modbus CRC for the given byte slice using polynomial 0xA001
func CalculateCRC16(data []byte) uint16 {
	var crc uint16 = 0xFFFF

	for _, b := range data {
		crc ^= uint16(b)
		for j := 0; j < 8; j++ {
			if (crc & 0x0001) != 0 {
				crc = (crc >> 1) ^ 0xA001
			} else {
				crc >>= 1
			}
		}
	}

	return crc
}

// AppendCRC16 appends the 2-byte CRC16 (Low byte first, High byte second) to the data buffer
func AppendCRC16(data []byte) []byte {
	crc := CalculateCRC16(data)
	return append(data, byte(crc&0xFF), byte((crc>>8)&0xFF))
}

// ValidateCRC16 checks if the last 2 bytes of the message match the calculated CRC16
func ValidateCRC16(data []byte) bool {
	if len(data) < 3 {
		return false
	}
	expectedCRC := CalculateCRC16(data[:len(data)-2])
	actualLow := data[len(data)-2]
	actualHigh := data[len(data)-1]
	actualCRC := uint16(actualLow) | (uint16(actualHigh) << 8)

	return expectedCRC == actualCRC
}
