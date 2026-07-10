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

// CreateNewGroup godoc
// @Summary      Create a new group
// @Description  Create a new group along with channel, members, and unreads
// @Tags         group
// @Accept       json
// @Produce      json
// @Param        req body dto.CreateGroupDto true "Create Group Info"
// @Success      200 {object} map[string]interface{}
// @Router       /group [post]
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

// UpdateGroupInfo godoc
// @Summary      Update group info
// @Description  Update information of an existing group
// @Tags         group
// @Accept       json
// @Produce      json
// @Param        group-id path string true "Group ID"
// @Param        req body dto.UpdateGroupDto true "Update Group Info"
// @Success      200 {object} map[string]interface{}
// @Router       /group/{group-id} [put]
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

// DeleteGroup godoc
// @Summary      Delete a group
// @Description  Delete a group by its ID
// @Tags         group
// @Produce      json
// @Param        group-id path string true "Group ID"
// @Success      200 {object} map[string]interface{}
// @Router       /group/{group-id} [delete]
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
