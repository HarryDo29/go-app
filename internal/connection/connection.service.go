package conection

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

type IConnectionService interface {
	CreateConnection(conDto dto.ConnectionDto) *dto.CreateConnectionResponseDto
	GetConnection(participant_ids [2]string) *dto.ConnectionResponseDto
	GetConnectionByUserId(userId string) *[]dto.ConnectionResponseDto
	AcceptedConnection(id string) *dto.ConnectionResponseDto
	DeleteConnection(id string) bool
}

type ConnectionService struct {
	userRepo       user.IUserRepo
	connectionRepo IConnectionRepo
	channelService IChannelService
}

// CreateConnection implements [IConnectionService].
func (c *ConnectionService) CreateConnection(conDto dto.ConnectionDto) *dto.CreateConnectionResponseDto {
	var newConnection *schema.DbConnection
	// validate user_id trước khi tạo connection
	requestUser := c.userRepo.GetUserById(conDto.RequesterId)
	receiveUser := c.userRepo.GetUserById(conDto.ReceiverId)
	if receiveUser != nil && requestUser != nil {
		newConnection = c.connectionRepo.CreateConnection(conDto)
	}
	if newConnection == nil {
		return nil
	}
	participantIDs := [2]string{
		newConnection.ParticipantIDs[0].Hex(),
		newConnection.ParticipantIDs[1].Hex(),
	}
	connRes := &dto.ConnectionResponseDto{
		ConnectionId:   newConnection.ID.Hex(),
		RequesterId:    newConnection.RequesterID.Hex(),
		ReceiverId:     newConnection.ReceiverID.Hex(),
		ParticipantIDs: participantIDs,
		Status:         string(newConnection.Status),
		AcceptedAt:     newConnection.AcceptedAt,
	}

	// tạo channel ứng với connection
	channelRes := c.channelService.CreateChannel(dto.CreateChannelDto{
		ChannelType: string(schema.ChannelTypeDirect),
		ChannelKey:  connRes.ConnectionId,
	})
	if channelRes == nil {
		// Rollback connection
		return nil
	}
	userIDs := connRes.ParticipantIDs[:]
	// tạo members ứng với 2 userId trong connection
	membersRes := c.channelService.AddMemberToChannel(dto.CreateChannelMemberDto{
		ChannelId: channelRes.ChannelId,
		UserIds:   userIDs,
		Role:      string(schema.ChannelMemberRoleMember),
		Status:    string(schema.ChannelMemberStatusActive),
	})
	if membersRes == nil {
		fmt.Println("MemberRes is nil")
		return nil
	}

	// tạo unreads ứng với 2 userId trong connection
	unreadsRes := c.channelService.CreateChannelUnreads(dto.CreateChannelUnreadDto{
		ChannelId: channelRes.ChannelId,
		UserIds:   userIDs,
		Unread:    0,
		Version:   0,
	})
	if !unreadsRes {
		return nil
	}

	return &dto.CreateConnectionResponseDto{
		Connection: connRes,
		Channel:    channelRes,
		Members:    membersRes,
		Unreads:    unreadsRes,
	}
}

// GetConnection implements [IConnectionService].
func (c *ConnectionService) GetConnection(participant_ids [2]string) *dto.ConnectionResponseDto {
	participantIDs := [2]primitive.ObjectID{}
	// kiểm tra user tồn tại không (tránh user_id ảo)
	for i, userId := range participant_ids {
		user := c.userRepo.GetUserById(userId)
		if user == nil {
			return nil
		}
		participantIDs[i] = user.ID
	}
	// lấy connection theo participantIDs
	connection := c.connectionRepo.GetConnection(participantIDs)
	if connection == nil {
		return nil
	}
	// convert sang response dto
	return &dto.ConnectionResponseDto{
		ConnectionId:   connection.ID.Hex(),
		RequesterId:    connection.RequesterID.Hex(),
		ReceiverId:     connection.ReceiverID.Hex(),
		ParticipantIDs: participant_ids,
		Status:         string(connection.Status),
		AcceptedAt:     connection.AcceptedAt,
	}

}

// GetConnectionByUserId implements [IConnectionService].
func (c *ConnectionService) GetConnectionByUserId(userId string) *[]dto.ConnectionResponseDto {
	// kiểm tra user tồn tại không (tránh user_id ảo)
	user := c.userRepo.GetUserById(userId)
	if user == nil {
		return nil
	}
	connections := c.connectionRepo.GetConnectionByUserId(userId)
	fmt.Println("connections: ", *connections)
	responseConnection := make([]dto.ConnectionResponseDto, 0)
	for _, connection := range *connections {
		participantIDs := [2]string{
			connection.ParticipantIDs[0].Hex(),
			connection.ParticipantIDs[1].Hex(),
		}
		responseConnection = append(responseConnection,
			dto.ConnectionResponseDto{
				ConnectionId:   connection.ID.Hex(),
				RequesterId:    connection.RequesterID.Hex(),
				ReceiverId:     connection.ReceiverID.Hex(),
				ParticipantIDs: participantIDs,
				Status:         string(connection.Status),
				AcceptedAt:     connection.AcceptedAt,
			})
	}
	return &responseConnection
}

// AcceptedConnection implements [IConnectionService].
func (c *ConnectionService) AcceptedConnection(id string) *dto.ConnectionResponseDto {
	ID := utils.ObjectIDFromHex(id)
	if ID == primitive.NilObjectID {
		return nil
	}
	// tạo channel luôn ?
	connection := c.connectionRepo.AcceptedConnection(ID)
	participantIDs := [2]string{
		connection.ParticipantIDs[0].Hex(),
		connection.ParticipantIDs[1].Hex(),
	}
	return &dto.ConnectionResponseDto{
		ConnectionId:   connection.ID.Hex(),
		RequesterId:    connection.RequesterID.Hex(),
		ReceiverId:     connection.ReceiverID.Hex(),
		ParticipantIDs: participantIDs,
		Status:         string(connection.Status),
		AcceptedAt:     connection.AcceptedAt,
	}
}

// DeleteConnection implements [IConnectionService].
func (c *ConnectionService) DeleteConnection(id string) bool {
	ID := utils.ObjectIDFromHex(id)
	if ID == primitive.NilObjectID {
		return false
	}
	return c.connectionRepo.DeleteConnection(ID)
}

func NewConnectionService(
	userRepo user.IUserRepo,
	connectionRepo IConnectionRepo,
	channelService channel.IChannelService,
) IConnectionService {
	return &ConnectionService{
		userRepo:       userRepo,
		connectionRepo: connectionRepo,
		channelService: channelService,
	}
}
