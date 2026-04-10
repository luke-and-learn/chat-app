package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Message struct {
	Username	string	`json:"username"`
	Text		string	`json:"text"`
	Time		string	`json:"time"`
	Type		string	`json:"type"`
	UserCount	int		`json:"user_count"`
}

type Client struct {
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

type Hub struct {
	clients		map[*Client]bool
	broadcast	chan Message
	register	chan *Client
	unregister	chan *Client
}

func newHub() *Hub {
	return &Hub{
		clients: 	make(map[*Client]bool),
		broadcast:	make(chan Message, 256),
		register: 	make(chan *Client),
		unregister: make(chan *Client),
	}
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

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				h.broadcastAll(Message{
					Username:	client.username,
					Text: 		client.username + " left the chat",
					Type: 		"leave",
					UserCount: 	len(h.clients),
				})
			}
		case msg := <-h.broadcast:
			h.broadcastAll(msg)
		}
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

var upgrader = websocket.Upgrader{
	ReadBufferSize:		1024,
	WriteBufferSize:	1024,
	CheckOrigin:		func(r *http.Request) bool { return true },
}

func main() {
	hub := newHub()
	go hub.run()

	r := gin.Default()

	r.Static("/static", "./static")
	r.GET("/", func(c *gin.Context) {
		c.File(".static/index.html")
	})

	r.GET("/ws", func(c *gin.Context) {
		username := c.Query("username")
		if username == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "username required"})
			return
		}
		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil { return }

		client := &Client{
			hub: hub, conn: conn,
			send: make(chan Message, 256),
			username: username,
		}
		hub.register <- client
		go client.writePump()
		go client.readPump()
	})

	r.Run(":8080")
}
