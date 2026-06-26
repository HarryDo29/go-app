package group

import (
	dto "go-app/internal/dto"
	"go-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type GroupController struct {
	groupService IGroupService
}

func NewGroupController(
	groupService IGroupService,
) *GroupController {
	return &GroupController{
		groupService: groupService,
	}
}

func (gc *GroupController) CreateNewGroup(c *gin.Context) {
	var groupDto dto.CreateGroupDto
	err := c.ShouldBindJSON(&groupDto)
	if err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	// tạo group cùng với channel, members, unreads
	result := gc.groupService.CreateGroup(groupDto)
	if result == nil {
		response.ErrorResponse(c, response.ErrCodeCreateFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

func (gc *GroupController) UpdateGroupInfo(c *gin.Context) {
	var groupId = c.Param("group-id")
	var updateDto dto.UpdateGroupDto
	err := c.ShouldBindJSON(&updateDto)
	if err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	group := gc.groupService.UpdateGroup(groupId, updateDto)
	if group == nil {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, group)
}

func (gc *GroupController) DeleteGroup(c *gin.Context) {
	var groupId = c.Param("group-id")
	result := gc.groupService.DeleteGroup(groupId)
	if !result {
		response.ErrorResponse(c, response.ErrCodeDeleteFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{
		"message": "Delete group success",
	})
}
