package packet

type PacketType uint

const (
	SYN PacketType = iota + 1
	SYN_ACK
	ACK
	DATA
	FIN
	HEARTBEAT
)

const (
	FLAG_SYN uint8 = 1 << 0
	FLAG_ACK uint8 = 1 << 1
	FLAG_FIN uint8 = 1 << 2
	FLAG_RST uint8 = 1 << 3
)

type Packet struct {
	Version    uint8
	Type       PacketType
	SEQ        uint
	ACK        uint
	Flags      uint8
	Window     uint16
	PayloadLen uint16
	Checksum   uint16
	Payload    []byte
}
