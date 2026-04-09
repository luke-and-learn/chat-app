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

Type Hub struct {
	clients		map[*Client]bool
	broadcast	chan Message
	register	chan *Client
	unregister	chan *Client
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			h.broadcastAll(Message{
				Username:	client.username,
				Text:		client.username + " joined the chat",
				Type: 		"join",
				UserCount:	len(h.clients),
			})
		}

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.broadcastAll(Message{
					username:	client.username,
					Text: 		client.username + " left the chat",
					Type: 		"leave",
					UserCount: 	len(h.clients),
				})
			}
		case msg := <-h.broadcast:
			h.broadcastAll(msg)
	}
}

func (h *Hub) broadcastAll(msg Message) {
	for client := range h.clients {
		select {
		case client.send <- msg:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}
