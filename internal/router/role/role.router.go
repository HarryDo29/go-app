package role

import (
	"go-app/internal/wire"

	"github.com/gin-gonic/gin"
)

type RoleRouter struct{}

func (rr *RoleRouter) InitRoleRouter(router *gin.RouterGroup) {
	roleController, _ := wire.InitRoleRouterHandler()

	// public — chỉ admin mới dùng nhưng tạm để public để test
	roleRouterPublic := router.Group("/role")
	{
		roleRouterPublic.POST("/", roleController.AddNewRole)
		roleRouterPublic.GET("/", roleController.GetRoles)
	}
}
