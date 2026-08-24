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
	// global.WsHub is declared as IWsHub to avoid a circular import, but the
	// concrete type is always *ws.Hub. We assert here since this package already
	// imports both sides and is the natural bridge point.
	hub, ok := global.WsHub.(*ws.Hub)
	if !ok {
		panic("global.WsHub is not a *websocket.Hub")
	}
	wsHandler := ws.NewHandler(hub)

	wsRouter := router.Group("/ws")
	wsRouter.Use(middleware.AuthNMiddleware())
	{
		wsRouter.GET("", wsHandler.HandleWebSocket)
	}
}
