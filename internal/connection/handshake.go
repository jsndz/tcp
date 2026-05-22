package connection

import (
	"errors"
	"fmt"

	"github.com/jsndz/tcp/internal/packet"
)

func (c *Connection) Connect() error {
	err := c.Send(packet.FLAG_SYN, nil)
	if err != nil {
		return err
	}
	pkt, err := c.ReadPacket()
	if err != nil {
		return err
	}
	if pkt.Flags&packet.FLAG_ACK == 0 {
		fmt.Println("flag:", pkt.Flags)
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

func (c *Connection) Accept() error {
	pkt, err := c.ReadPacket()
	if err != nil {
		return err
	}
	if pkt.Flags&packet.FLAG_SYN == 0 {
		return errors.New("not SYN")
	}

	c.mu.Lock()
	c.RecvSeq = pkt.SEQ + 1
	c.SendWindow = uint32(pkt.Window)
	c.mu.Unlock()

	return c.Send(
		packet.FLAG_SYN|packet.FLAG_ACK,
		nil,
	)
}
