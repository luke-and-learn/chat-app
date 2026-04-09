type Message struct {
	Username	string	`json:"username"`
	Text		string	`json:"text"`
	Time		string	`json:"time"`
	Type		string	`json:"type"`
	UserCount	int		`json:"user_count"`
}

Type Client struct {
	hub		 *Hub
	conn	 *websocket.Conn 
	send	 chan Message
	username string
}

func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	for {
		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			break
		}
		msg.Username = c.username
		msg.Type = "message"
		c.hub.broadcast <- msg
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteJSON(msg); err != nil {
			break
		}
	}
}
