package message

import (
	"go-app/global"
	"go-app/internal/message"
	"go-app/internal/middleware"
	"go-app/internal/wire"

	"github.com/gin-gonic/gin"
)

type MessageRouter struct{}

func (mr *MessageRouter) InitMessageRouter(router *gin.RouterGroup) {
	// Khởi tạo service qua wire, truyền thêm global.WsHub (singleton)
	messageService, _ := wire.InitMessageService()
	messageController := message.NewMessageController(messageService, global.WsHub)

	messageRouter := router.Group("")
	messageRouter.Use(middleware.AuthNMiddleware()) // Áp dụng AuthenMiddleware
	{
		// /api/messages/...
		messagesGroup := messageRouter.Group("/messages")
		{
			messagesGroup.POST("", messageController.CreateMessage)
			messagesGroup.PUT("/:message-id", messageController.UpdateMessage)
			messagesGroup.DELETE("/:message-id/recall", messageController.RecallMessage)
			messagesGroup.POST("/:message-id/hide", messageController.HideMessageForMe) // yêu cầu truyền ?channel_id=...
		}

		// /api/channels/...
		channelsGroup := messageRouter.Group("/channels")
		{
			channelsGroup.GET("/:channel-id/messages", messageController.GetMessagesByChannel)
			channelsGroup.DELETE("/:channel-id/history", messageController.DeleteChatHistory)
		}
	}
}
