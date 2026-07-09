package group

import (
	dto "go-app/internal/dto"
	"go-app/internal/websocket"
	"go-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type IWsHub interface {
	Notify(userId string, event string, payload interface{}) bool
}

type GroupController struct {
	groupService IGroupService
	hub          IWsHub
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

	// response channel realtime
	gc.hub.Notify(groupDto.OwnerId, websocket.EventNewChannel, result)
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

func NewGroupController(
	groupService IGroupService,
	hub IWsHub,
) *GroupController {
	return &GroupController{
		groupService: groupService,
		hub:          hub,
	}
}
