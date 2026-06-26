package connection

import (
	"go-app/internal/middleware"
	"go-app/internal/wire"

	"github.com/gin-gonic/gin"
)

type ConnectionRouter struct{}

func (cr *ConnectionRouter) InitConnectionRouter(router *gin.RouterGroup) {
	// wired (dependency injection DI)
	connectionController, _ := wire.InitConnectionRouterHandler()

	connectionRouter := router.Group("/connections")
	connectionRouter.Use(middleware.AuthNMiddleware()) // Áp dụng AuthenMiddleware
	{
		connectionRouter.POST("", connectionController.CreateConnection)
		connectionRouter.POST("/detail", connectionController.GetConnection) // lấy detail bằng participants
		connectionRouter.GET("/user", connectionController.GetConnectionByUserId)
		connectionRouter.PUT("/:connection-id/accept", connectionController.AcceptedConnection)
		connectionRouter.DELETE("/:connection-id", connectionController.DeleteConnection)
	}
}
