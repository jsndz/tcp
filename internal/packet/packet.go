package packet

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/jsndz/tcp/pkg/util"
)

type PacketType uint

const (
	FLAG_SYN  uint8 = 1 << 0
	FLAG_ACK  uint8 = 1 << 1
	FLAG_FIN  uint8 = 1 << 2
	FLAG_RST  uint8 = 1 << 3
	FLAG_DATA uint8 = 1 << 4
)

type Packet struct {
	Version    uint8
	Flags      uint8
	SEQ        uint32
	ACK        uint32
	Window     uint16
	PayloadLen uint16
	Checksum   uint16
	Payload    []byte
}

func NewPacket(seq uint32, ack uint32, flags uint8, window uint16, payload []byte) *Packet {
	pkt := &Packet{
		Version:    1,
		Flags:      flags,
		SEQ:        seq,
		ACK:        ack,
		Window:     window,
		PayloadLen: uint16(len(payload)),
		Payload:    payload,
		Checksum:   0,
	}
	pkt.Checksum = util.ComputeChecksum(pkt.Marshall())
	return pkt
}

func (p *Packet) Marshall() []byte {
	var buf bytes.Buffer
	buf.WriteByte(p.Version)
	buf.WriteByte(p.Flags)
	binary.Write(&buf, binary.BigEndian, p.SEQ)
	binary.Write(&buf, binary.BigEndian, p.ACK)
	binary.Write(&buf, binary.BigEndian, p.Window)
	binary.Write(&buf, binary.BigEndian, p.PayloadLen)
	binary.Write(&buf, binary.BigEndian, p.Checksum)
	buf.Write(p.Payload)

	return buf.Bytes()
}

func (p *Packet) ValidateChecksum() bool {
	data := p.Marshall()
	checksum := util.ComputeChecksum(data)
	if checksum == 0 {
		return true
	}
	return false
}

func Unmarshall(data []byte) (*Packet, error) {
	if len(data) < 16 {
		return nil, errors.New("Invalid packet")
	}
	pkt := &Packet{
		Version: data[0],
		Flags:   data[1],
	}
	buf := bytes.NewReader(data[2:])
	binary.Read(buf, binary.BigEndian, &pkt.SEQ)
	binary.Read(buf, binary.BigEndian, &pkt.ACK)
	binary.Read(buf, binary.BigEndian, &pkt.Window)
	binary.Read(buf, binary.BigEndian, &pkt.PayloadLen)
	binary.Read(buf, binary.BigEndian, &pkt.Checksum)

	if int(pkt.PayloadLen) != len(data)-16 {
		return nil, errors.New("Invalid packet")
	}
	pkt.Payload = data[16:]

	if util.ComputeChecksum(data) != 0 {
		return nil, errors.New("Invalid packet")
	}

	return pkt, nil
}
