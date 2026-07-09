package channel

import (
	"fmt"
	channelRepo "go-app/internal/channel/repo"
	"go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/pkg/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IUserRepo, IConnectionRepo và IGroupRepo được định nghĩa local để tránh import cycle
type IUserRepo interface {
	GetUserById(id string) *schema.DbUser
}

type IConnectionRepo interface {
	GetConnectionById(id primitive.ObjectID) *schema.DbConnection
}

type IGroupRepo interface {
	GetGroupByID(ID primitive.ObjectID) *schema.DbGroup
	UpdateGroup(groupId primitive.ObjectID, dto dto.UpdateGroupDto) *schema.DbGroup
}

type IMessageRepo interface {
	GetMessageByID(id primitive.ObjectID) *schema.Message
}

type IChannelService interface {
	//channel
	CreateChannel(channelDto dto.CreateChannelDto) *dto.ChannelResponseDto // tạo channel + tạo member (owner, ...)
	GetChannels(userId string, queryDto dto.ChannelQueryDto) *[]dto.ChannelResponseDto
	GetChannelsByUserId(userId string) *[]dto.ChannelResponseDto
	GetChannel(channelId string) *dto.ChannelResponseDto // active
	UpdateChannel(
		channelId string,
		updateDto dto.UpdateChannelDto) *dto.ChannelResponseDto // update channel ()
	DeleteChannel(channelId string) bool // set channel status delete

	// channel-member
	AddMemberToChannel(channelMemberDto dto.CreateChannelMemberDto) *[]dto.ChannelMemberResponseDto
	AddMemberToGroupChannel(channelMemberDto dto.CreateChannelMemberDto) *[]dto.ChannelMemberResponseDto
	RemoveMemberFromChannel(channelMemberId string) bool // remove member (set stauts left or delete)
	// GetChannelIdsByUserId(userId string) *[]primitive.ObjectID          // danh sách channelId của user
	GetChannelMembers(channelId string) *[]dto.ChannelMemberResponseDto // view danh sách member
	GetChannelMemberCount(channelId string) int                         // lấy số lượng member để update vào channel
	CheckUserInChannel(channelId string, userId string) bool            // kiểm tra user có trong channel

	// channel-unread
	CreateChannelUnreads(unreadDto dto.CreateChannelUnreadDto) bool
	GetChannelUnreads(channelId string) *[]dto.ChannelUnreadResponseDto // lấy unread theo channel (danh sách)
	GetChannelUnread(unreadId string) *dto.ChannelUnreadResponseDto     // lấy unread theo userId (danh sách tn chưa đọc)
	UpdateChannelUnread(
		unreadId string,
		updateDto dto.UpdateChannelUnreadDto,
	) *dto.ChannelUnreadResponseDto // khi user đọc tin nhắn channel nào thì đánh dấu đã đọc
	DeleteChannelUnread(unreadId string) bool // Xóa unread khi user ko còn là member của channel
}

type ChannelService struct {
	channelRepo       channelRepo.IChannelRepo
	channelMemberRepo channelRepo.IChannelMemberRepo
	channelUnreadRepo channelRepo.IChannelUnreadRepo
	userRepo          IUserRepo
	connectionRepo    IConnectionRepo
	groupRepo         IGroupRepo
	msgRepo           IMessageRepo
}

// CreateChannel implements [IChannelService].
func (c *ChannelService) CreateChannel(channelDto dto.CreateChannelDto) *dto.ChannelResponseDto {
	// tạo channel
	channel := c.channelRepo.CreateChannel(channelDto)
	if channel == nil {
		fmt.Println("channel ko tạo được")
		return nil
	}
	resDto := dto.ChannelResponseDto{
		ChannelId:   channel.ID.Hex(),
		ChannelType: string(channel.ChannelType),
		ChannelKey:  channel.ChannelKey.Hex(),
		Subject:     nil,
		Group:       nil,
		LastMsg:     nil,
		UpdatedAt:   channel.UpdatedAt,
	}

	switch channel.ChannelType {
	case schema.ChannelTypeDirect:
		// TODO: get subject user
		// get connection (by channel key)
		// get user
	case schema.ChannelTypeGroup:
		// TODO: get group
		// get group (by channel key)
	}
	return &resDto
}

func (c *ChannelService) GetChannels(userId string, queryDto dto.ChannelQueryDto) *[]dto.ChannelResponseDto {
	userID := utils.ObjectIDFromHex(userId)
	if userID == primitive.NilObjectID {
		return nil
	}

	channels := c.channelRepo.GetChannelsByQuery(userID, queryDto)
	if channels == nil {
		return nil
	}

	channelRes := make([]dto.ChannelResponseDto, 0)
	for _, channel := range *channels {
		cType := string(channel.ChannelType)

		tmp := dto.ChannelResponseDto{
			ChannelId:   channel.ID.Hex(),
			ChannelType: string(channel.ChannelType),
			ChannelKey:  channel.ChannelKey.Hex(),
			Subject:     nil,
			Group:       nil,
			LastMsg:     nil,
			UpdatedAt:   channel.UpdatedAt,
		}

		switch cType {
		case string(schema.ChannelTypeDirect):
			// get connection (by channel key)
			connection := c.connectionRepo.GetConnectionById(channel.ChannelKey)
			if connection == nil {
				continue
			}
			// get user (friend) in connection
			for _, ID := range connection.ParticipantIDs {
				if ID != userID {
					user := c.userRepo.GetUserById(ID.Hex())
					if user == nil {
						continue
					}
					tmp.Subject = &dto.UserResponseDto{
						UserId:    user.ID.Hex(),
						UserName:  user.UserName,
						Email:     user.Email,
						IsActive:  &user.IsActive,
						Role:      user.Role.Hex(),
						AvatarUrl: user.AvatarUrl,
					}
					break
				}
			}
		case string(schema.ChannelTypeGroup):
			// get group (channel key)
			group := c.groupRepo.GetGroupByID(channel.ChannelKey)
			if group == nil {
				continue
			}
			tmp.Group = &dto.GroupResponseDto{
				GroupId:     group.ID.Hex(),
				GroupName:   group.Name,
				OwnerId:     group.OwnerID.Hex(),
				MemberCount: group.MemberCount,
				Status:      group.Status,
				Members:     nil,
			}

			// get members (channel_members)
			members := make([]dto.ChannelMemberResponseDto, 0)
			result := c.channelMemberRepo.GetChannelMembers(channel.ID)
			if result == nil {
				continue
			}
			for _, member := range *result {
				if member.LeftAt == nil &&
					(member.Status != schema.ChannelMemberStatusKicked && member.Status != schema.ChannelMemberStatusLeft) {
					user := c.userRepo.GetUserById(member.UserID.Hex())
					if user == nil {
						continue
					}
					userRes := &dto.UserResponseDto{
						UserId:    user.ID.Hex(),
						UserName:  user.UserName,
						Email:     user.Email,
						IsActive:  &user.IsActive,
						Role:      user.Role.Hex(),
						AvatarUrl: user.AvatarUrl,
					}
					members = append(members, dto.ChannelMemberResponseDto{
						MemberId:  member.ID.Hex(),
						ChannelId: member.ChannelID.Hex(),
						User:      userRes,
						Role:      string(member.Role),
						Status:    string(member.Status),
						JoinedAt:  member.JoinedAt,
					})
				}
			}
			tmp.Group.Members = &members
		}
		// get last message
		lastMsg := c.msgRepo.GetMessageByID(channel.LastMsgID)
		if lastMsg == nil {
			tmp.LastMsg = nil
		} else {
			tmp.LastMsg = &dto.MessageResponseDto{
				MsgId:          lastMsg.ID.Hex(),
				ChannelId:      lastMsg.ChannelID.Hex(),
				FromId:         lastMsg.FromID.Hex(),
				Content:        lastMsg.Content,
				MsgType:        string(lastMsg.MsgType),
				MsgSeq:         lastMsg.MsgSeq,
				Status:         string(lastMsg.Status),
				IsDelete:       lastMsg.IsDelete,
				RepliedToMsgId: lastMsg.RepliedToMsgID.Hex(),
			}
		}
		channelRes = append(channelRes, tmp)
	}
	return &channelRes
}

// GetChannelsByUserId implements [IChannelService].
func (c *ChannelService) GetChannelsByUserId(userId string) *[]dto.ChannelResponseDto {
	return c.GetChannels(userId, dto.ChannelQueryDto{})
}

// GetChannel implements [IChannelService].
func (c *ChannelService) GetChannel(channelId string) *dto.ChannelResponseDto {
	channelID := utils.ObjectIDFromHex(channelId)
	if channelID == primitive.NilObjectID {
		return nil
	}

	channel := c.channelRepo.GetChannelById(channelID)
	if channel == nil {
		return nil
	}

	res := dto.ChannelResponseDto{
		ChannelId:   channel.ID.Hex(),
		ChannelType: string(channel.ChannelType),
		ChannelKey:  channel.ChannelKey.Hex(),
		Subject:     nil,
		Group:       nil,
		LastMsg:     nil,
		UpdatedAt:   channel.UpdatedAt,
	}

	// get group (channel key)
	group := c.groupRepo.GetGroupByID(channel.ChannelKey)

	res.Group = &dto.GroupResponseDto{
		GroupId:     group.ID.Hex(),
		GroupName:   group.Name,
		OwnerId:     group.OwnerID.Hex(),
		MemberCount: group.MemberCount,
		Status:      group.Status,
		Members:     nil,
	}

	// get members (channel_members)
	members := make([]dto.ChannelMemberResponseDto, 0)
	result := c.channelMemberRepo.GetChannelMembers(channel.ID)

	for _, member := range *result {
		if member.LeftAt == nil &&
			(member.Status != schema.ChannelMemberStatusKicked && member.Status != schema.ChannelMemberStatusLeft) {
			user := c.userRepo.GetUserById(member.UserID.Hex())
			if user == nil {
				continue
			}
			userRes := &dto.UserResponseDto{
				UserId:    user.ID.Hex(),
				UserName:  user.UserName,
				Email:     user.Email,
				IsActive:  &user.IsActive,
				Role:      user.Role.Hex(),
				AvatarUrl: user.AvatarUrl,
			}
			members = append(members, dto.ChannelMemberResponseDto{
				MemberId:  member.ID.Hex(),
				ChannelId: member.ChannelID.Hex(),
				User:      userRes,
				Role:      string(member.Role),
				Status:    string(member.Status),
				JoinedAt:  member.JoinedAt,
			})
		}
	}
	res.Group.Members = &members

	// get last message
	lastMsg := c.msgRepo.GetMessageByID(channel.LastMsgID)
	if lastMsg == nil {
		res.LastMsg = nil
	} else {
		res.LastMsg = &dto.MessageResponseDto{
			MsgId:          lastMsg.ID.Hex(),
			ChannelId:      lastMsg.ChannelID.Hex(),
			FromId:         lastMsg.FromID.Hex(),
			Content:        lastMsg.Content,
			MsgType:        string(lastMsg.MsgType),
			MsgSeq:         lastMsg.MsgSeq,
			Status:         string(lastMsg.Status),
			IsDelete:       lastMsg.IsDelete,
			RepliedToMsgId: lastMsg.RepliedToMsgID.Hex(),
		}
	}

	return &res
}

// UpdateChannel implements [IChannelService].
func (c *ChannelService) UpdateChannel(
	channelId string,
	updateDto dto.UpdateChannelDto,
) *dto.ChannelResponseDto {
	channelID := utils.ObjectIDFromHex(channelId)
	if channelID == primitive.NilObjectID {
		return nil
	}

	channel := c.channelRepo.UpdateChannel(channelID, updateDto)
	if channel == nil {
		return nil
	}

	return &dto.ChannelResponseDto{
		ChannelId:   channel.ID.Hex(),
		ChannelType: string(channel.ChannelType),
		ChannelKey:  channel.ChannelKey.Hex(),
		Subject:     nil,
		Group:       nil,
		LastMsg:     nil,
		UpdatedAt:   channel.UpdatedAt,
	}
}

// DeleteChannel implements [IChannelService].
func (c *ChannelService) DeleteChannel(channelId string) bool {
	channelID := utils.ObjectIDFromHex(channelId)
	if channelID == primitive.NilObjectID {
		return false
	}
	return c.channelRepo.DeleteChannel(channelID)
}

func (c *ChannelService) AddMemberToChannel(channelMemberDto dto.CreateChannelMemberDto) *[]dto.ChannelMemberResponseDto {
	members := c.channelMemberRepo.CreateChannelMember(channelMemberDto)
	if members == nil {
		return nil
	}

	memberRes := make([]dto.ChannelMemberResponseDto, 0)
	for _, member := range *members {
		c.channelRepo.AddParticipant(member.ChannelID, member.UserID) // thêm participant_id vào channel
		memberRes = append(memberRes, dto.ChannelMemberResponseDto{
			MemberId:  member.ID.Hex(),
			ChannelId: member.ChannelID.Hex(),
			User:      nil,
			Role:      string(member.Role),
			Status:    string(member.Status),
			JoinedAt:  member.JoinedAt,
		})
	}
	return &memberRes
}

// AddMemberToGroupChannel implements [IChannelService].
func (c *ChannelService) AddMemberToGroupChannel(channelMemberDto dto.CreateChannelMemberDto) *[]dto.ChannelMemberResponseDto {
	channelID := utils.ObjectIDFromHex(channelMemberDto.ChannelId)
	if channelID == primitive.NilObjectID {
		return nil
	}

	channel := c.channelRepo.GetChannelById(channelID)
	if channel == nil || channel.ChannelType != schema.ChannelTypeGroup {
		return nil
	}

	// Gọi hàm AddMemberToChannel chung để xử lý thêm thành viên
	memberRes := c.AddMemberToChannel(channelMemberDto)
	if memberRes == nil {
		return nil
	}

	// Tính tổng số lượng thành viên thực tế của nhóm
	totalMembers := c.channelMemberRepo.GetChannelMembers(channelID)
	var memberCount int64 = 0
	if totalMembers != nil {
		memberCount = int64(len(*totalMembers))
	}

	updateDto := dto.UpdateGroupDto{
		MemberCount: memberCount,
	}

	group := c.groupRepo.UpdateGroup(channel.ChannelKey, updateDto)
	if group == nil {
		return nil
	}

	return memberRes
}

// RemoveMemberFromChannel implements [IChannelService].
func (c *ChannelService) RemoveMemberFromChannel(channelMemberId string) bool {
	memberId := utils.ObjectIDFromHex(channelMemberId)
	if memberId == primitive.NilObjectID {
		return false
	}
	member := c.channelMemberRepo.GetChannelMemberById(memberId)
	if member != nil {
		c.channelRepo.RemoveParticipant(member.ChannelID, member.UserID)
	}
	return c.channelMemberRepo.DeleteChannelMember(memberId)
}

// GetChannelMembers implements [IChannelService].
func (c *ChannelService) GetChannelMembers(channelId string) *[]dto.ChannelMemberResponseDto {
	channelID := utils.ObjectIDFromHex(channelId)
	if channelID == primitive.NilObjectID {
		return nil
	}

	members := c.channelMemberRepo.GetChannelMembers(channelID)
	if members == nil {
		return nil
	}

	memberRes := make([]dto.ChannelMemberResponseDto, 0)
	for _, member := range *members {
		var userRes *dto.UserResponseDto
		user := c.userRepo.GetUserById(member.UserID.Hex())
		if user != nil {
			userRes = &dto.UserResponseDto{
				UserId:    user.ID.Hex(),
				UserName:  user.UserName,
				Email:     user.Email,
				IsActive:  &user.IsActive,
				Role:      user.Role.Hex(),
				AvatarUrl: user.AvatarUrl,
			}
		}

		memberRes = append(memberRes, dto.ChannelMemberResponseDto{
			MemberId:  member.ID.Hex(),
			ChannelId: member.ChannelID.Hex(),
			User:      userRes,
			Role:      string(member.Role),
			Status:    string(member.Status),
			JoinedAt:  member.JoinedAt,
		})
	}
	return &memberRes
}

// GetChannelMemberCount implements [IChannelService].
func (c *ChannelService) GetChannelMemberCount(channelId string) int {
	channelID := utils.ObjectIDFromHex(channelId)
	if channelID == primitive.NilObjectID {
		return 0
	}
	members := c.channelMemberRepo.GetChannelMembers(channelID)
	if members == nil {
		return 0
	}
	return len(*members)
}

// CheckUserInChannel implements [IChannelService].
func (c *ChannelService) CheckUserInChannel(channelId string, userId string) bool {
	cID := utils.ObjectIDFromHex(channelId)
	uID := utils.ObjectIDFromHex(userId)
	if cID == primitive.NilObjectID || uID == primitive.NilObjectID {
		return false
	}
	return c.channelMemberRepo.CheckUserInChannel(cID, uID)
}

// CreateChannelUnread implements [IChannelService].
func (c *ChannelService) CreateChannelUnreads(unreadDto dto.CreateChannelUnreadDto) bool {
	return c.channelUnreadRepo.CreateChannelUnreads(unreadDto)
}

// GetChannelUnreads implements [IChannelService].
func (c *ChannelService) GetChannelUnreads(userId string) *[]dto.ChannelUnreadResponseDto {
	userID := utils.ObjectIDFromHex(userId)
	if userID == primitive.NilObjectID {
		return nil
	}

	unreads := c.channelUnreadRepo.GetChannelUnreads(userID)
	if unreads == nil {
		return nil
	}

	unreadDto := make([]dto.ChannelUnreadResponseDto, 0)
	for _, unread := range *unreads {
		unreadDto = append(unreadDto, dto.ChannelUnreadResponseDto{
			UnreadId:    unread.ID.Hex(),
			ChannelId:   unread.ChannelID.Hex(),
			UserId:      unread.UserID.Hex(),
			Unread:      unread.Unread,
			LastMsgId:   unread.LastMsgID.Hex(),
			LastMsgTime: unread.LastMsgTime,
		})
	}
	return &unreadDto
}

// GetChannelUnread implements [IChannelService].
func (c *ChannelService) GetChannelUnread(unreadId string) *dto.ChannelUnreadResponseDto {
	unreadID := utils.ObjectIDFromHex(unreadId)
	if unreadID == primitive.NilObjectID {
		return nil
	}
	unread := c.channelUnreadRepo.GetChannelUnread(unreadID)
	if unread == nil {
		return nil
	}
	return &dto.ChannelUnreadResponseDto{
		UnreadId:    unread.ID.Hex(),
		ChannelId:   unread.ChannelID.Hex(),
		UserId:      unread.UserID.Hex(),
		Unread:      unread.Unread,
		LastMsgId:   unread.LastMsgID.Hex(),
		LastMsgTime: unread.LastMsgTime,
	}
}

// UpdateChannelUnread implements [IChannelService].
func (c *ChannelService) UpdateChannelUnread(
	unreadId string,
	updateDto dto.UpdateChannelUnreadDto,
) *dto.ChannelUnreadResponseDto {
	unreadID := utils.ObjectIDFromHex(unreadId)
	if unreadID == primitive.NilObjectID {
		return nil
	}

	unread := c.channelUnreadRepo.UpdateChannelUnread(unreadID, updateDto)
	if unread == nil {
		return nil
	}

	return &dto.ChannelUnreadResponseDto{
		UnreadId:    unread.ID.Hex(),
		ChannelId:   unread.ChannelID.Hex(),
		UserId:      unread.UserID.Hex(),
		Unread:      unread.Unread,
		LastMsgId:   unread.LastMsgID.Hex(),
		LastMsgTime: unread.LastMsgTime,
	}
}

// DeleteChannelUnread implements [IChannelService].
func (c *ChannelService) DeleteChannelUnread(unreadId string) bool {
	unreadID := utils.ObjectIDFromHex(unreadId)
	if unreadID == primitive.NilObjectID {
		return false
	}
	return c.channelUnreadRepo.DeleteChannelUnread(unreadID)
}

func NewChannelService(
	channelRepo channelRepo.IChannelRepo,
	channelMemberRepo channelRepo.IChannelMemberRepo,
	channelUnreadRepo channelRepo.IChannelUnreadRepo,
	userRepo IUserRepo,
	connectionRepo IConnectionRepo,
	groupRepo IGroupRepo,
	msgRepo IMessageRepo,
) IChannelService {
	return &ChannelService{
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
		channelUnreadRepo: channelUnreadRepo,
		userRepo:          userRepo,
		connectionRepo:    connectionRepo,
		groupRepo:         groupRepo,
		msgRepo:           msgRepo,
	}
}
