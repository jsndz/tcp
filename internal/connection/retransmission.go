package connection

import (
	"syscall"
	"time"
)

const MAX_RETRIES = 10

func (conn *Connection) RetransmissionLoop() {

	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {
		for seq, sent := range conn.SendBuffer {
			timeout := time.Second * time.Duration(1<<sent.Retries)

			if time.Since(sent.SendTime) < timeout {
				continue
			}
			data := sent.Packet.Marshall()
			err := syscall.Sendto(conn.SocketFD, data, 0, conn.PeerAddr)
			if err != nil {
				continue
			}
			sent.Retries++
			sent.SendTime = time.Now()
			if sent.Retries > MAX_RETRIES {
				delete(conn.SendBuffer, seq)
			}
		}
	}
}
