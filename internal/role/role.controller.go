package role

import (
	dto "go-app/internal/dto"
	"go-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type RoleController struct {
	roleService IRoleService
}

func NewRoleController(roleService IRoleService) *RoleController {
	return &RoleController{
		roleService: roleService,
	}
}

func (rc *RoleController) AddNewRole(c *gin.Context) {
	var createDto dto.CreateRoleDto
	if err := c.ShouldBindJSON(&createDto); err != nil {
		response.ErrorResponse(c, response.ErrCodeServer)
		return
	}

	result := rc.roleService.CreateRole(createDto)
	if result == nil { // ← fix: nil = thất bại
		response.ErrorResponse(c, response.ErrCodeServer)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

func (rc *RoleController) GetRoles(c *gin.Context) {
	result := rc.roleService.GetAllRole()
	if len(*result) == 0 {
		response.ErrorResponse(c, response.ErrCodeServer)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}
