package connection

import (
	"errors"
	"net"
	"syscall"

	"github.com/jsndz/tcp/internal/packet"
)

type Connection struct {
	SocketFD int

	SendSeq uint32
	RecvSeq uint32

	PeerAddr *syscall.SockaddrInet4
}

func NewConnection(socketFD int, peerAddr string) *Connection {
	ip := net.ParseIP(peerAddr).To4()
	var addr [4]byte
	copy(addr[:], ip)
	return &Connection{
		SocketFD: socketFD,
		SendSeq:  0,
		RecvSeq:  0,
		PeerAddr: &syscall.SockaddrInet4{
			Addr: addr,
		},
	}
}

func (c *Connection) Send(flags uint8, payload []byte) error {
	packet := packet.NewPacket(c.SendSeq, c.RecvSeq, flags, 0, payload)
	data := packet.Marshall()
	err := syscall.Sendto(c.SocketFD, data, 0, c.PeerAddr)
	c.SendSeq += uint32(len(payload))
	return err
}

func (c *Connection) Recv(data []byte) (*packet.Packet, error) {
	buf := make([]byte, 65535)
	n, from, err := syscall.Recvfrom(c.SocketFD, buf, 0)
	if err != nil {
		return nil, err
	}

	if from.(*syscall.SockaddrInet4).Addr != c.PeerAddr.Addr {
		return nil, errors.New("unexpected packet source")
	}
	ipHeaderLen := (buf[0] & 0x0F) * 4

	pkt, err := packet.Unmarshall(buf[ipHeaderLen:n])
	if err != nil {
		return nil, err
	}
	if pkt.SEQ < c.RecvSeq {
		return nil, errors.New("duplicate packet")
	}

	if pkt.SEQ > c.RecvSeq {
		return nil, errors.New("out of order packet")
	}
	c.RecvSeq += uint32(len(pkt.Payload))
	if pkt.Flags&packet.FLAG_ACK != 0 {
		return pkt, nil
	}
	flag := packet.FLAG_ACK
	if len(data) > 0 {
		flag |= packet.FLAG_DATA
	}
	c.Send(flag, data)
	return pkt, nil
}
