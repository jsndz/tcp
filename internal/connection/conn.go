package connection

import (
	"errors"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/jsndz/tcp/internal/packet"
)

type SendPacket struct {
	Packet   *packet.Packet
	Retries  int
	SendTime time.Time
}

type Connection struct {
	mu sync.Mutex

	SocketFD int

	SendSeq    uint32
	SendBase   uint32 // oldest unacknowledged SEQ
	SendWindow uint32 // max number of bytes that can be sent without ACK
	RecvSeq    uint32
	RecvWindow uint32 // max number of bytes that can be received without ACK -> your window size

	PeerAddr *syscall.SockaddrInet4

	SendBuffer map[uint32]*SendPacket
	RecvBuffer map[uint32]*packet.Packet

	DeliveryChan chan *packet.Packet
}

func NewConnection(socketFD int, peerAddr string) *Connection {
	ip := net.ParseIP(peerAddr).To4()

	var addr [4]byte
	copy(addr[:], ip)

	return &Connection{
		SocketFD: socketFD,

		SendSeq: 0,
		RecvSeq: 0,

		PeerAddr: &syscall.SockaddrInet4{
			Addr: addr,
		},

		SendBuffer: make(map[uint32]*SendPacket),
		RecvBuffer: make(map[uint32]*packet.Packet),

		DeliveryChan: make(chan *packet.Packet, 1024),
	}
}

func (c *Connection) Send(flags uint8, payload []byte) error {

	c.mu.Lock()

	seq := c.SendSeq
	ack := c.RecvSeq

	advance := uint32(len(payload))
	if flags&packet.FLAG_SYN != 0 {
		advance++
	}
	if flags&packet.FLAG_FIN != 0 {
		advance++
	}
	c.SendSeq += advance
	c.mu.Unlock()

	pkt := packet.NewPacket(
		seq,
		ack,
		flags,
		uint16(c.RecvWindow),
		payload,
	)

	data := pkt.Marshall()

	err := syscall.Sendto(
		c.SocketFD,
		data,
		0,
		c.PeerAddr,
	)

	if err != nil {
		return err
	}

	needsRetransmit :=
		len(payload) > 0 ||
			flags&packet.FLAG_SYN != 0 ||
			flags&packet.FLAG_FIN != 0

	if needsRetransmit {

		c.mu.Lock()
		c.SendBuffer[pkt.SEQ] = &SendPacket{
			Packet:   pkt,
			Retries:  0,
			SendTime: time.Now(),
		}
		c.mu.Unlock()
	}

	return nil
}

func (c *Connection) Recv() error {

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
	c.mu.Lock()
	c.SendWindow = uint32(pkt.Window)
	c.mu.Unlock()
	// ack packet
	if pkt.Flags&packet.FLAG_ACK != 0 {

		c.mu.Lock()
		for seq, sent := range c.SendBuffer {
			end := seq + uint32(len(sent.Packet.Payload))
			if sent.Packet.Flags&packet.FLAG_SYN != 0 {
				end++
			}
			if sent.Packet.Flags&packet.FLAG_FIN != 0 {
				end++
			}
			if end <= pkt.ACK {
				delete(c.SendBuffer, seq)
			}
		}
		if pkt.ACK > c.SendBase {
			c.SendBase = pkt.ACK
		}
		c.mu.Unlock()
	}

	if len(pkt.Payload) == 0 {
		return nil
	}

	c.mu.Lock()
	expected := c.RecvSeq
	c.mu.Unlock()

	if pkt.SEQ < expected {
		c.SendAck()
		return nil
	}

	if pkt.SEQ > expected {

		c.mu.Lock()
		c.RecvBuffer[pkt.SEQ] = pkt
		c.mu.Unlock()

		return nil
	}

	c.DeliveryChan <- pkt

	c.mu.Lock()
	c.RecvSeq += uint32(len(pkt.Payload))
	c.mu.Unlock()

	for {

		c.mu.Lock()
		next, ok := c.RecvBuffer[c.RecvSeq]
		if !ok {
			c.mu.Unlock()
			break
		}
		delete(c.RecvBuffer, c.RecvSeq)
		c.RecvSeq += uint32(len(next.Payload))
		c.mu.Unlock()

		c.DeliveryChan <- next
	}

	c.SendAck()

	return nil
}

func (c *Connection) SendAck() error {
	return c.Send(packet.FLAG_ACK, nil)
}
