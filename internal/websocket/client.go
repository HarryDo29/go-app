package websocket

import (
	"encoding/json"
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
		var req WsRequest

		err := c.Conn.ReadJSON(&req) // đọc message từ client
		if err != nil {
			log.Println("websocket read error:", err)
			break
		}
		// call hub to handle event
		switch req.Event {
		case "TYPING_START":
			var p TypingPayload
			if err := json.Unmarshal(req.Payload, &p); err == nil {
				// c.Hub.HandleTyping(p.ChannelId, p.UserId)
			}
		case "TYPING_STOP":
			var p TypingPayload
			if err := json.Unmarshal(req.Payload, &p); err == nil {
				// c.Hub.StopTyping(p.ChannelId, p.UserId)
			}
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
