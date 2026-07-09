package conection

import (
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
	GetConnectionByUserId(userId string, status string) *[]dto.ConnectionResponseDto
	AcceptedConnection(id string) *dto.CreateConnectionResponseDto
	RejectedConnection(id string) bool
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

	return &dto.CreateConnectionResponseDto{
		Connection: connRes,
		Requester: &dto.UserResponseDto{
			UserId:    requestUser.ID.Hex(),
			UserName:  requestUser.UserName,
			Email:     requestUser.Email,
			IsActive:  &requestUser.IsActive,
			Role:      requestUser.Role.Hex(),
			AvatarUrl: requestUser.AvatarUrl,
		},
		Receiver: &dto.UserResponseDto{
			UserId:    receiveUser.ID.Hex(),
			UserName:  receiveUser.UserName,
			Email:     receiveUser.Email,
			IsActive:  &receiveUser.IsActive,
			Role:      receiveUser.Role.Hex(),
			AvatarUrl: receiveUser.AvatarUrl,
		},
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
func (c *ConnectionService) GetConnectionByUserId(userId string, status string) *[]dto.ConnectionResponseDto {
	// kiểm tra user tồn tại không (tránh user_id ảo)
	user := c.userRepo.GetUserById(userId)
	if user == nil {
		return nil
	}

	connections := c.connectionRepo.GetConnectionByUserId(userId, status)
	responseConnection := make([]dto.ConnectionResponseDto, 0)
	for _, connection := range *connections {
		participantIDs := [2]string{
			connection.ParticipantIDs[0].Hex(),
			connection.ParticipantIDs[1].Hex(),
		}

		// Fetch requester & receiver info để đính kèm vào response
		var requesterDto, receiverDto *dto.UserResponseDto
		if requester := c.userRepo.GetUserById(connection.RequesterID.Hex()); requester != nil {
			requesterDto = &dto.UserResponseDto{
				UserId:    requester.ID.Hex(),
				UserName:  requester.UserName,
				Email:     requester.Email,
				AvatarUrl: requester.AvatarUrl,
				IsActive:  &requester.IsActive,
				Role:      requester.Role.Hex(),
			}
		}
		if receiver := c.userRepo.GetUserById(connection.ReceiverID.Hex()); receiver != nil {
			receiverDto = &dto.UserResponseDto{
				UserId:    receiver.ID.Hex(),
				UserName:  receiver.UserName,
				Email:     receiver.Email,
				AvatarUrl: receiver.AvatarUrl,
				IsActive:  &receiver.IsActive,
				Role:      receiver.Role.Hex(),
			}
		}

		responseConnection = append(responseConnection,
			dto.ConnectionResponseDto{
				ConnectionId:   connection.ID.Hex(),
				RequesterId:    connection.RequesterID.Hex(),
				ReceiverId:     connection.ReceiverID.Hex(),
				ParticipantIDs: participantIDs,
				Status:         string(connection.Status),
				AcceptedAt:     connection.AcceptedAt,
				Requester:      requesterDto,
				Receiver:       receiverDto,
			})
	}
	return &responseConnection
}

// AcceptedConnection implements [IConnectionService].
func (c *ConnectionService) AcceptedConnection(id string) *dto.CreateConnectionResponseDto {
	ID := utils.ObjectIDFromHex(id)
	if ID == primitive.NilObjectID {
		return nil
	}
	connection := c.connectionRepo.AcceptedConnection(ID)
	if connection == nil {
		return nil
	}
	participantIDs := [2]string{
		connection.ParticipantIDs[0].Hex(),
		connection.ParticipantIDs[1].Hex(),
	}

	connRes := &dto.ConnectionResponseDto{
		ConnectionId:   connection.ID.Hex(),
		RequesterId:    connection.RequesterID.Hex(),
		ReceiverId:     connection.ReceiverID.Hex(),
		ParticipantIDs: participantIDs,
		Status:         string(connection.Status),
		AcceptedAt:     connection.AcceptedAt,
	}

	// Fetch users to populate DTO
	requestUser := c.userRepo.GetUserById(connection.RequesterID.Hex())
	receiveUser := c.userRepo.GetUserById(connection.ReceiverID.Hex())

	// Tạo channel khi connection được accept
	channelRes := c.channelService.CreateChannel(dto.CreateChannelDto{
		ChannelType: string(schema.ChannelTypeDirect),
		ChannelKey:  connRes.ConnectionId,
	})
	if channelRes == nil {
		return nil
	}

	userIDs := connRes.ParticipantIDs[:]
	// Tạo members
	membersRes := c.channelService.AddMemberToChannel(dto.CreateChannelMemberDto{
		ChannelId: channelRes.ChannelId,
		UserIds:   userIDs,
		Role:      string(schema.ChannelMemberRoleMember),
		Status:    string(schema.ChannelMemberStatusActive),
	})
	if membersRes == nil {
		return nil
	}

	// Tạo unreads
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
		Requester: &dto.UserResponseDto{
			UserId:    requestUser.ID.Hex(),
			UserName:  requestUser.UserName,
			Email:     requestUser.Email,
			IsActive:  &requestUser.IsActive,
			Role:      requestUser.Role.Hex(),
			AvatarUrl: requestUser.AvatarUrl,
		},
		Receiver: &dto.UserResponseDto{
			UserId:    receiveUser.ID.Hex(),
			UserName:  receiveUser.UserName,
			Email:     receiveUser.Email,
			IsActive:  &receiveUser.IsActive,
			Role:      receiveUser.Role.Hex(),
			AvatarUrl: receiveUser.AvatarUrl,
		},
	}
}

// RejectedConnection implements [IConnectionService].
func (c *ConnectionService) RejectedConnection(id string) bool {
	ID := utils.ObjectIDFromHex(id)
	if ID == primitive.NilObjectID {
		return false
	}
	return c.connectionRepo.RejectedConnection(ID)
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
