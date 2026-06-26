package websocket

import (
	"go-app/internal/middleware"
	ws "go-app/internal/websocket"
	"go-app/internal/wire"

	"github.com/gin-gonic/gin"
)

type WebSocketRouter struct {
	Handler *ws.Handler
}

func NewWebSocketRouter(handler *ws.Handler) *WebSocketRouter {
	return &WebSocketRouter{
		Handler: handler,
	}
}

func (r *WebSocketRouter) InitWebSocketRouter(router *gin.RouterGroup) {
	wsHandler, _ := wire.InitWebSocketHandler()

	go wsHandler.Hub.Run()

	wsRouter := router.Group("/ws")
	wsRouter.Use(middleware.AuthNMiddleware())
	{
		wsRouter.GET("", wsHandler.HandleWebSocket)
	}
}
