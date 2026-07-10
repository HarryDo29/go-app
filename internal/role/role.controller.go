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

// AddNewRole godoc
// @Summary      Add new role
// @Description  Create a new role
// @Tags         role
// @Accept       json
// @Produce      json
// @Param        req body dto.CreateRoleDto true "Create Role Info"
// @Success      200 {object} map[string]interface{}
// @Router       /role [post]
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

// GetRoles godoc
// @Summary      Get all roles
// @Description  Get a list of all roles
// @Tags         role
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /role [get]
func (rc *RoleController) GetRoles(c *gin.Context) {
	result := rc.roleService.GetAllRole()
	if len(*result) == 0 {
		response.ErrorResponse(c, response.ErrCodeServer)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}
