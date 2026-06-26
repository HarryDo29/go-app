package channel

import (
	"go-app/internal/middleware"
	"go-app/internal/wire"

	"github.com/gin-gonic/gin"
)

type ChannelRouter struct{}

func (cr *ChannelRouter) InitChannelRouter(router *gin.RouterGroup) {
	// wired (dependency injection DI)
	channelController, _ := wire.InitChannelRouterHandler()

	channelRouter := router.Group("/channels")
	channelRouter.Use(middleware.AuthNMiddleware()) // Áp dụng AuthenMiddleware
	{
		// Channel endpoints
		channelRouter.GET("", channelController.GetChannels)
		channelRouter.PUT("/:channel-id", channelController.UpdateChannel)
		channelRouter.DELETE("/:channel-id", channelController.DeleteChannel)

		// Member endpoints
		channelRouter.POST("/members", channelController.AddMemberToChannel)
		channelRouter.DELETE("/members/:member-id", channelController.RemoveMemberFromChannel)
		channelRouter.GET("/:channel-id/members", channelController.GetChannelMembers)
		channelRouter.GET("/:channel-id/members/count", channelController.GetChannelMemberCount)

		// Unread endpoints
		channelRouter.GET("/unreads/user/:user-id", channelController.GetChannelUnreads)
		channelRouter.PUT("/unreads/:unread-id", channelController.UpdateChannelUnread)
		channelRouter.DELETE("/unreads/:unread-id", channelController.DeleteChannelUnread)
	}
}
