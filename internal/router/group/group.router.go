package group

import (
	"go-app/internal/middleware"
	"go-app/internal/wire"

	"github.com/gin-gonic/gin"
)

type GroupRouter struct{}

func (gr *GroupRouter) InitGroupRouter(router *gin.RouterGroup) {
	// wired (dependency injection DI)
	groupController, _ := wire.InitGroupRouterHandler()

	groupRouter := router.Group("/groups")
	groupRouter.Use(middleware.AuthNMiddleware()) // Áp dụng AuthenMiddleware
	{
		groupRouter.POST("", groupController.CreateNewGroup)
		groupRouter.PUT("/:group-id", groupController.UpdateGroupInfo)
		groupRouter.DELETE("/:group-id", groupController.DeleteGroup)
	}
}
