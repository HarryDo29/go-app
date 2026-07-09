package websocket

import (
	"go-app/global"
	"go-app/internal/middleware"
	ws "go-app/internal/websocket"

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
	// Dùng global.WsHub (singleton) - Hub đã được khởi tạo và chạy trong run.go
	wsHandler := ws.NewHandler(global.WsHub)

	wsRouter := router.Group("/ws")
	wsRouter.Use(middleware.AuthNMiddleware())
	{
		wsRouter.GET("", wsHandler.HandleWebSocket)
	}
}
