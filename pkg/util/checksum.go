package util

func ComputeChecksum(data []byte) uint16 {
	var sum uint32
	for i := 0; i < len(data)-1; i += 2 {
		word := uint16(data[i])<<8 | uint16(data[i+1])
		sum += uint32(word)
	}
	if len(data)%2 != 0 {
		sum += uint32(uint16(data[len(data)-1]) << 8)
	}
	for (sum >> 16) != 0 {
		sum = (sum & 0xFFFF) + (sum >> 16)
	}
	return ^uint16(sum)
}
