package connection

import "fmt"

func (c *Connection) Read() {
	go c.RecvLoop()
	for data := range c.DeliveryChan {
		fmt.Println(string(data))
	}
}
