package group

import (
	"fmt"
	"go-app/internal/channel"
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/internal/user"
	"go-app/pkg/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IChannelService định nghĩa local interface để tránh import cycle với package channel.
type IChannelService interface {
	CreateChannel(channelDto dto.CreateChannelDto) *dto.ChannelResponseDto
	AddMemberToChannel(memberDto dto.CreateChannelMemberDto) *[]dto.ChannelMemberResponseDto
	CreateChannelUnreads(unreadDto dto.CreateChannelUnreadDto) bool
}

type IGroupService interface {
	CreateGroup(createDto dto.CreateGroupDto) *dto.GroupResponseDto
	UpdateGroup(groupId string, updateDto dto.UpdateGroupDto) *dto.GroupResponseDto
	DeleteGroup(groupId string) bool
}

type GroupService struct {
	userService    user.IUserService
	groupRepo      IGroupRepo
	channelService IChannelService
}

// CreateGroup implements [IGroupService].
func (gc *GroupService) CreateGroup(groupDto dto.CreateGroupDto) *dto.GroupResponseDto {
	// tạo group
	group := gc.groupRepo.CreateGroup(groupDto)
	if group == nil {
		fmt.Println("group không tạo được")
		return nil
	}

	groupRes := &dto.GroupResponseDto{
		GroupId:     group.ID.Hex(),
		GroupName:   group.Name,
		OwnerId:     group.OwnerID.Hex(),
		MemberCount: group.MemberCount,
		Status:      group.Status,
	}

	// tạo channel ứng với group_id
	channelRes := gc.channelService.CreateChannel(dto.CreateChannelDto{
		ChannelType: string(schema.ChannelTypeGroup),
		ChannelKey:  groupRes.GroupId,
	})
	if channelRes == nil {
		// Rollback group
		return nil
	}

	// thêm owner group vào channel với role admin
	membersRes := gc.channelService.AddMemberToChannel(dto.CreateChannelMemberDto{
		ChannelId: channelRes.ChannelId,
		UserIds:   []string{groupRes.OwnerId},
		Role:      string(schema.ChannelMemberRoleAdmin),
		Status:    string(schema.ChannelMemberStatusActive),
	})
	if membersRes == nil {
		// Rollback channel và group
		return nil
	}

	// tạo unread cho member (mặc định unread: 0 và version: 0)
	unreadsRes := gc.channelService.CreateChannelUnreads(dto.CreateChannelUnreadDto{
		ChannelId: channelRes.ChannelId,
		UserIds:   []string{groupRes.OwnerId},
		Unread:    0,
		Version:   0,
	})
	if !unreadsRes {
		// Rollback members, channel và group
		return nil
	}

	return groupRes
}

// DeleteGroup implements [IGroupService].
func (gc *GroupService) DeleteGroup(groupId string) bool {
	id := utils.ObjectIDFromHex(groupId)
	if id == primitive.NilObjectID {
		return false
	}
	return gc.groupRepo.DeleteGroup(id)
}

// UpdateGroup implements [IGroupService].
func (gc *GroupService) UpdateGroup(groupId string, updateDto dto.UpdateGroupDto) *dto.GroupResponseDto {
	id := utils.ObjectIDFromHex(groupId)
	if id == primitive.NilObjectID {
		return nil
	}
	group := gc.groupRepo.UpdateGroup(id, updateDto)
	return &dto.GroupResponseDto{
		GroupId:     group.ID.Hex(),
		GroupName:   group.Name,
		OwnerId:     group.OwnerID.Hex(),
		MemberCount: group.MemberCount,
		Status:      group.Status,
	}
}

func NewGroupService(
	userService user.IUserService,
	groupRepo IGroupRepo,
	channelService channel.IChannelService,
) IGroupService {
	return &GroupService{
		userService:    userService,
		groupRepo:      groupRepo,
		channelService: channelService,
	}
}
