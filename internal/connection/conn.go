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
	mu       sync.Mutex
	SocketFD int

	SendSeq uint32
	RecvSeq uint32

	PeerAddr *syscall.SockaddrInet4

	SendBuffer map[uint32]*SendPacket    // used to keep track of sent packets for retransmission
	RecvBuffer map[uint32]*packet.Packet // used to keep track of received packets for in-order delivery
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
		SendBuffer: make(map[uint32]*SendPacket),
		RecvBuffer: make(map[uint32]*packet.Packet),
	}
}

func (c *Connection) Send(flags uint8, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	pkt := packet.NewPacket(c.SendSeq, c.RecvSeq, flags, 0, payload)
	data := pkt.Marshall()
	err := syscall.Sendto(c.SocketFD, data, 0, c.PeerAddr)
	c.SendSeq += uint32(len(payload))
	needsRetransmit :=
		len(payload) > 0 ||
			flags&packet.FLAG_SYN != 0 ||
			flags&packet.FLAG_FIN != 0
	if needsRetransmit {
		c.SendBuffer[pkt.SEQ] = &SendPacket{
			Packet:   pkt,
			Retries:  0,
			SendTime: time.Now(),
		}
	}
	return err
}

func (c *Connection) Recv() (*packet.Packet, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
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

	if pkt.Flags&packet.FLAG_ACK == 0 {
		//not an ACK
		c.RecvSeq += uint32(len(pkt.Payload))
		c.SendAck()
	} else {
		// ack recieved
		for seq, sent := range c.SendBuffer {
			end := seq + uint32(len(sent.Packet.Payload))
			if end <= pkt.ACK {
				delete(c.SendBuffer, seq)
			}
		}
	}

	return pkt, nil
}

func (c *Connection) SendAck() error {
	return c.Send(packet.FLAG_ACK, nil)
}
