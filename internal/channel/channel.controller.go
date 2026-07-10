package channel

import (
	"fmt"
	dto "go-app/internal/dto"
	"go-app/internal/websocket"
	"go-app/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
)

type IWsHub interface {
	Notify(userId string, event string, payload interface{}) bool
}

type ChannelController struct {
	channelService IChannelService
	wsHub          IWsHub
}

func NewChannelController(channelService IChannelService, wsHub IWsHub) *ChannelController {
	return &ChannelController{
		channelService: channelService,
		wsHub:          wsHub,
	}
}

// ================= Channel Endpoints =================
// GetChannels godoc
// @Summary      Get channels
// @Description  Get a list of channels for the current user
// @Tags         channel
// @Produce      json
// @Param        type query string false "Channel Type"
// @Param        updated_at query string false "Updated At (RFC3339)"
// @Success      200 {object} map[string]interface{}
// @Router       /channel [get]
func (cc *ChannelController) GetChannels(c *gin.Context) {
	userId := c.GetString("user-id")
	if userId == "" {
		response.ErrorResponse(c, response.ErrCodeAuthFailed)
		return
	}
	channelType := c.Query("type")
	updatedAtStr := c.Query("updated_at")
	limit := 10

	updatedAt := time.Now().UTC()
	if updatedAtStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, updatedAtStr)
		if err != nil {
			response.ErrorResponse(c, response.ErrCodeParamInvalid)
			return
		}
		updatedAt = parsedTime
	}

	queryDto := dto.ChannelQueryDto{
		ChannelType: channelType,
		UpdatedAt:   updatedAt,
		Limit:       limit,
	}
	result := cc.channelService.GetChannels(userId, queryDto)
	if result == nil {
		response.ErrorResponse(c, response.ErrCodeNotFound)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

// UpdateChannel godoc
// @Summary      Update channel
// @Description  Update channel info
// @Tags         channel
// @Accept       json
// @Produce      json
// @Param        channel-id path string true "Channel ID"
// @Param        req body dto.UpdateChannelDto true "Update Channel Info"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/{channel-id} [put]
func (cc *ChannelController) UpdateChannel(c *gin.Context) {
	channelId := c.Param("channel-id")
	if channelId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	var updateDto dto.UpdateChannelDto
	if err := c.ShouldBindJSON(&updateDto); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}

	result := cc.channelService.UpdateChannel(channelId, updateDto)
	if result == nil {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

// DeleteChannel godoc
// @Summary      Delete channel
// @Description  Delete a channel by ID
// @Tags         channel
// @Produce      json
// @Param        channel-id path string true "Channel ID"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/{channel-id} [delete]
func (cc *ChannelController) DeleteChannel(c *gin.Context) {
	channelId := c.Param("channel-id")
	if channelId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	ok := cc.channelService.DeleteChannel(channelId)
	if !ok {
		response.ErrorResponse(c, response.ErrCodeDeleteFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{"deleted": true})
}

// ================= Member Endpoints =================
// AddMemberToChannel godoc
// @Summary      Add member to channel
// @Description  Add one or more members to a channel
// @Tags         channel
// @Accept       json
// @Produce      json
// @Param        req body dto.CreateChannelMemberDto true "Member Info"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/members [post]
func (cc *ChannelController) AddMemberToChannel(c *gin.Context) {
	var memberDto dto.CreateChannelMemberDto
	if err := c.ShouldBindJSON(&memberDto); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	// thêm tv
	members := cc.channelService.AddMemberToGroupChannel(memberDto)
	if members == nil {
		response.ErrorResponse(c, response.ErrCodeCreateFailed)
		return
	}
	// tạo channel unread
	unreads := cc.channelService.CreateChannelUnreads(dto.CreateChannelUnreadDto{
		UserIds:   memberDto.UserIds,
		ChannelId: memberDto.ChannelId,
		Unread:    0,
		Version:   0,
	})
	if !unreads {
		response.ErrorResponse(c, response.ErrCodeCreateFailed)
		return
	}
	// broadcast new channel to new member
	channelDetail := cc.channelService.GetChannel(memberDto.ChannelId)
	if channelDetail != nil {
		for _, newMemberId := range memberDto.UserIds {
			cc.wsHub.Notify(newMemberId, websocket.EventNewChannel, channelDetail)
		}
	}

	// brodcast update channel member to existing members
	// Sử dụng trực tiếp memberDto.UserIds vì đây là danh sách những user MỚI thực sự được add
	newMemberMap := make(map[string]bool)
	for _, id := range memberDto.UserIds {
		newMemberMap[id] = true
	}

	channelMembers := cc.channelService.GetChannelMembers(memberDto.ChannelId)
	if channelMembers != nil {
		for _, channelMember := range *channelMembers {
			// Chỉ gửi sự kiện cho những ai KHÔNG nằm trong danh sách add mới (tức là thành viên cũ)
			if !newMemberMap[channelMember.User.UserId] {
				cc.wsHub.Notify(channelMember.User.UserId, websocket.EventUpdatedChannel, channelDetail)
			}
			fmt.Println("channel member", channelMember)
		}
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{
		"members": members,
		"unreads": unreads,
	})
}

// RemoveMemberFromChannel godoc
// @Summary      Remove member from channel
// @Description  Remove a specific member from a channel
// @Tags         channel
// @Produce      json
// @Param        member-id path string true "Member ID"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/members/{member-id} [delete]
func (cc *ChannelController) RemoveMemberFromChannel(c *gin.Context) {
	memberId := c.Param("member-id")
	if memberId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	ok := cc.channelService.RemoveMemberFromChannel(memberId)
	if !ok {
		response.ErrorResponse(c, response.ErrCodeDeleteFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{"removed": true})
}

// GetChannelMembers godoc
// @Summary      Get channel members
// @Description  Get a list of members in a specific channel
// @Tags         channel
// @Produce      json
// @Param        channel-id path string true "Channel ID"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/{channel-id}/members [get]
func (cc *ChannelController) GetChannelMembers(c *gin.Context) {
	channelId := c.Param("channel-id")
	if channelId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	result := cc.channelService.GetChannelMembers(channelId)
	if result == nil {
		response.ErrorResponse(c, response.ErrCodeNotFound)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

// GetChannelMemberCount godoc
// @Summary      Get channel member count
// @Description  Get the number of members in a specific channel
// @Tags         channel
// @Produce      json
// @Param        channel-id path string true "Channel ID"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/{channel-id}/members/count [get]
func (cc *ChannelController) GetChannelMemberCount(c *gin.Context) {
	channelId := c.Param("channel-id")
	if channelId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	count := cc.channelService.GetChannelMemberCount(channelId)
	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{"member_count": count})
}

// ================= Unread Endpoints =================
// GetChannelUnreads godoc
// @Summary      Get channel unreads
// @Description  Get unread message counts for a user
// @Tags         channel
// @Produce      json
// @Param        user-id path string true "User ID"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/unreads/user/{user-id} [get]
func (cc *ChannelController) GetChannelUnreads(c *gin.Context) {
	userId := c.Param("user-id")
	if userId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	result := cc.channelService.GetChannelUnreads(userId)
	if result == nil {
		response.ErrorResponse(c, response.ErrCodeNotFound)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

// UpdateChannelUnread godoc
// @Summary      Update channel unread
// @Description  Update unread message count status
// @Tags         channel
// @Accept       json
// @Produce      json
// @Param        unread-id path string true "Unread ID"
// @Param        req body dto.UpdateChannelUnreadDto true "Update Info"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/unreads/{unread-id} [put]
func (cc *ChannelController) UpdateChannelUnread(c *gin.Context) {
	unreadId := c.Param("unread-id")
	if unreadId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	var updateDto dto.UpdateChannelUnreadDto
	if err := c.ShouldBindJSON(&updateDto); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}

	result := cc.channelService.UpdateChannelUnread(unreadId, updateDto)
	if result == nil {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

// DeleteChannelUnread godoc
// @Summary      Delete channel unread
// @Description  Delete a channel unread record
// @Tags         channel
// @Produce      json
// @Param        unread-id path string true "Unread ID"
// @Success      200 {object} map[string]interface{}
// @Router       /channel/unreads/{unread-id} [delete]
func (cc *ChannelController) DeleteChannelUnread(c *gin.Context) {
	unreadId := c.Param("unread-id")
	if unreadId == "" {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	ok := cc.channelService.DeleteChannelUnread(unreadId)
	if !ok {
		response.ErrorResponse(c, response.ErrCodeDeleteFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{"deleted": true})
}
