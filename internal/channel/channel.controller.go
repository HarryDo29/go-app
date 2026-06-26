package channel

import (
	dto "go-app/internal/dto"
	"go-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type ChannelController struct {
	channelService IChannelService
}

func NewChannelController(channelService IChannelService) *ChannelController {
	return &ChannelController{
		channelService: channelService,
	}
}

// ================= Channel Endpoints =================
func (cc *ChannelController) GetChannels(c *gin.Context) {
	userId := c.GetString("user-id")
	if userId == "" {
		response.ErrorResponse(c, response.ErrCodeAuthFailed)
		return
	}
	channelType := c.Query("type")

	result := cc.channelService.GetChannels(userId, channelType)
	if result == nil {
		response.ErrorResponse(c, response.ErrCodeNotFound)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

// func (cc *ChannelController) GetChannel(c *gin.Context) {
// 	channelId := c.Param("channelId")
// 	if channelId == "" {
// 		response.ErrorResponse(c, response.ErrCodeParamInvalid)
// 		return
// 	}
// 	result := cc.channelService.GetChannel(channelId)
// 	if result == nil {
// 		response.ErrorResponse(c, response.ErrCodeNotFound)
// 		return
// 	}
// 	response.SuccessResponse(c, response.ErrCodeSuccess, result)
// }

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
func (cc *ChannelController) AddMemberToChannel(c *gin.Context) {
	var memberDto dto.CreateChannelMemberDto
	if err := c.ShouldBindJSON(&memberDto); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	// thêm tv
	members := cc.channelService.AddMemberToChannel(memberDto)
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
	response.SuccessResponse(c, response.ErrCodeSuccess, gin.H{
		"members": members,
		"unreads": unreads,
	})
}

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
