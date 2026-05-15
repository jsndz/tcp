package connection

import (
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
	return &Connection{
		SocketFD: socketFD,
		SendSeq:  0,
		RecvSeq:  0,
		PeerAddr: &syscall.SockaddrInet4{Addr: [4]byte(ip)},
	}
}

func (c *Connection) Send(flags uint8, payload []byte) error {
	packet := packet.NewPacket(c.SendSeq, 0, flags, 0, payload)
	data := packet.Marshall()
	err := syscall.Sendto(c.SocketFD, data, 0, c.PeerAddr)
	c.SendSeq += uint32(len(payload))
	return err
}

func (c *Connection) Recv() error {
	buf := make([]byte, 65535)
	n, from, err := syscall.Recvfrom(c.SocketFD, buf, 0)
	if err != nil {
		return err
	}
	if from.(*syscall.SockaddrInet4).Addr != c.PeerAddr.Addr {
		return nil
	}

	packet, err := packet.Unmarshall(buf[:n])
	if err != nil {
		return err
	}
	c.RecvSeq += uint32(len(packet.Payload))
	return nil
}
