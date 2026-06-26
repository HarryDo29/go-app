package websocket

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	UserId string
	ConnectionId string
	Conn         *websocket.Conn
	Hub          *Hub
	Send         chan MessagePayload
}

// lắng nghe dữ liệu từ Client --> Server
func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister(c)
		c.Conn.Close()
	}()

	for {
		var payload ClientMessagePayload

		err := c.Conn.ReadJSON(&payload)
		if err != nil {
			log.Println("websocket read error:", err)
			break
		}

		// xử lý message từ client
		switch payload.Event {
		case "UNREGISTER":
			c.Hub.Unregister(c)

		case "JOIN_CHANNEL":
			c.Hub.JoinChannel(payload.ChannelId, c.UserId)

		case "LEAVE_CHANNEL":
			c.Hub.LeaveChannel(payload.ChannelId, c.UserId)

		// case "SEND_MESSAGE":
		// 	message := MessagePayload{
		// 		ChannelId: payload.ChannelId,
		// 		SenderId:  c.UserID,
		// 		Content:   payload.Content,
		// 		Type:      payload.Type,
		// 	}
		// 	c.Hub.Broadcast(message)
		}

	}
}

// đưa dữ liệu từ Server --> Client
func (c *Client) WritePump() {
	defer c.Conn.Close()

	for message := range c.Send {
		err := c.Conn.WriteJSON(message)
		if err != nil {
			log.Println("websocket write error:", err)
			break
		}
	}
}
