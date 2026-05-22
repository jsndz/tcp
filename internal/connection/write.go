package connection

func (c *Connection) Write() error {
	go c.RetransmissionLoop()
	go c.SendLoop()
	return nil
}