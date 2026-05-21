package connection

import (
	"errors"
	"syscall"

	"github.com/jsndz/tcp/internal/packet"
)

func (c *Connection) Handshake() error {
	err := c.Send(packet.FLAG_SYN, nil)
	if err != nil {
		return err
	}
	buf := make([]byte, 65535)

	n, from, err := syscall.Recvfrom(
		c.SocketFD,
		buf,
		0,
	)

	if err != nil {
		return err
	}

	addr, ok := from.(*syscall.SockaddrInet4)

	if !ok {
		return errors.New("invalid sockaddr")
	}

	if addr.Addr != c.PeerAddr.Addr {
		return errors.New("unexpected packet source")
	}

	ipHeaderLen := (buf[0] & 0x0F) * 4

	pkt, err := packet.Unmarshall(
		buf[ipHeaderLen:n],
	)
	if err != nil {
		return err
	}
	if pkt.Flags&packet.FLAG_ACK == 0 {
		return errors.New("not ack")
	}
	if pkt.Flags&packet.FLAG_SYN == 0 {
		return errors.New("not SYN")
	}
	c.mu.Lock()
	c.RecvSeq = pkt.SEQ + 1
	c.SendWindow = uint32(pkt.Window)
	c.mu.Unlock()
	err = c.SendAck()
	if err != nil {
		return err
	}
	return nil
}
