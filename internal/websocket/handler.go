package websocket

import (
	"go-app/internal/channel"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	gorilla "github.com/gorilla/websocket"
)

type Handler struct {
	Hub            *Hub
	ChannelService channel.IChannelService
}

func NewHandler(hub *Hub, channelService channel.IChannelService) *Handler {
	return &Handler{
		Hub:            hub,
		ChannelService: channelService,
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

// HandleWebSocket
func (h *Handler) HandleWebSocket(c *gin.Context) {
	// get user id from context (AuthMiddleware)
	userId := c.GetString("user-id")

	if userId == "" {
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
		ConnectionId: uuid.New().String(),
		Conn:         conn,
		Hub:          h.Hub,
		Send:         make(chan MessagePayload, 256),
	}

	h.Hub.Register(client)

	channels := h.ChannelService.GetChannelsByUserId(userId)
	if channels != nil {
		for _, ch := range *channels {
			h.Hub.JoinChannel(ch.ChannelId, userId)
		}
	}

	go client.WritePump()
	go client.ReadPump()
}
