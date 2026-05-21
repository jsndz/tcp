package connection

func (c *Connection) Write() error {
	err := c.Handshake()
	if err != nil {
		return err
	}
	go c.RetransmissionLoop()
	go c.SendLoop()
	return nil
}
