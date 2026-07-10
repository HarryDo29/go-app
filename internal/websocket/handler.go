package websocket

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	gorilla "github.com/gorilla/websocket"
)

type Handler struct {
	Hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{
		Hub: hub,
	}
}

// config để upgrade kết nối websocket
var upgrader = gorilla.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Development: cho phép tất cả origin
		// Production: nên check domain frontend
		return true
	},
}

// HandleWebSocket godoc
// @Summary      WebSocket connection
// @Description  Upgrade to WebSocket connection
// @Tags         websocket
// @Produce      json
// @Param        connectionId query string true "Connection ID"
// @Success      101 {string} string "Switching Protocols"
// @Router       /ws [get]
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// get user id from context (AuthMiddleware)
	userId := c.GetString("user-id")
	connectionId := c.Query("connectionId")

	if userId == "" || connectionId == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "unauthorized",
		})
		return
	}

	// upgrade kết nối websocket
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "failed to upgrade websocket",
		})
		return
	}

	client := &Client{
		UserId:       userId,
		ConnectionId: connectionId,
		Conn:         conn,
		Hub:          h.Hub,
		Send:         make(chan WsResponse, 256),
	}

	h.Hub.Register(client)
	fmt.Println("client connected:", client.UserId, client.ConnectionId)

	go client.WritePump()
	go client.ReadPump()
}
