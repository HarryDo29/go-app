package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserId       string
	ConnectionId string
	Conn         *websocket.Conn
	Hub          *Hub
	Send         chan WsResponse
}

// lắng nghe dữ liệu từ Client --> Server
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	for {
		var payload ClientMessagePayload

		err := c.Conn.ReadJSON(&payload) // đọc message từ client
		if err != nil {
			log.Println("websocket read error:", err)
			break
		}

		switch payload.Event {
		case "UNREGISTER":
			c.Hub.Unregister(c)
		}
	}
}

// đưa dữ liệu từ Server --> Client
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for res := range c.Send {
		err := c.Conn.WriteJSON(res)
		if err != nil {
			log.Println("websocket write error:", err)
			break
		}
	}
}
