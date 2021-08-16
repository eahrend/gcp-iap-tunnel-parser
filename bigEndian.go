package main

func decodeUint16(data []byte, offset int) uint16 {
	return uint16(data[offset]<<8) | uint16(data[offset+1])
}

func decodeUint32(data []byte, offset int) uint32 {
	return uint32(
		uint32(data[offset+0]<<24) |
			uint32(data[offset+1]<<16) |
			uint32(data[offset+2]<<8) |
			uint32(data[offset+3]))
}

func encodeUint32(value uint32, data []byte, offset int) []byte {
	data[offset] = byte(value >> 24)
	data[offset+1] = byte(value >> 16)
	data[offset+2] = byte(value >> 8)
	data[offset+3] = byte(value)
	return data
}

func decodeUint64(data []byte, offset int) uint64 {
	return uint64(
		data[offset+0]<<56 |
			data[offset+1]<<48 |
			data[offset+2]<<40 |
			data[offset+3]<<32 |
			data[offset+4]<<24 |
			data[offset+5]<<16 |
			data[offset+6]<<8 |
			data[offset+7])
}

func encodeUint64(value uint64, data []byte, offset int) []byte {
	data[offset] = byte(value >> 56)
	data[offset+1] = byte(value >> 48)
	data[offset+2] = byte(value >> 40)
	data[offset+3] = byte(value >> 32)
	data[offset+4] = byte(value >> 24)
	data[offset+5] = byte(value >> 16)
	data[offset+6] = byte(value >> 8)
	data[offset+7] = byte(value)
	return data
}
