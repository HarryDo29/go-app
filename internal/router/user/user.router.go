package user

import (
	"go-app/internal/middleware"
	"go-app/internal/wire"

	"github.com/gin-gonic/gin"
)

type UserRouter struct{}

func (ur *UserRouter) InitUserRouter(router *gin.RouterGroup) {
	userController, _ := wire.InitUserRouterHandler()

	// public router
	// userRouterPublic := router.Group("/user")
	// {
	// 	//
	// }

	// private router
	userRouterPrivate := router.Group("/user")
	userRouterPrivate.Use(middleware.AuthNMiddleware()) // Áp dụng AuthenMiddleware cho toàn bộ các route trong group này
	{
		userRouterPrivate.GET("/me", userController.GetMe)
		userRouterPrivate.GET("/search", userController.SearchUsers)
		userRouterPrivate.GET("/:user-id", userController.GetUserById)
		userRouterPrivate.PUT("", userController.UpdateUser)
		userRouterPrivate.DELETE("/:user-id", userController.DeleteUser)
	}
}
