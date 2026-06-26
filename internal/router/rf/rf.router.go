package rf

import (
	"go-app/internal/middleware"
	"go-app/internal/wire"

	"github.com/gin-gonic/gin"
)

type RfRouter struct{}

func (ur *RfRouter) InitRfRouter(router *gin.RouterGroup) {
	// wired (dependency injection DI)
	rfController, _ := wire.InitRefreshTokenRouterHandler()

	// rfRouterPublic := router.Group("/refresh-token")
	// {
	// 	//
	// }

	// private router
	rfRouterPrivate := router.Group("/refresh-token")
	{
		rfRouterPrivate.POST("/refresh",
			middleware.AuthNMiddleware(),
			rfController.CreateRefreshToken,
		)
	}
}
