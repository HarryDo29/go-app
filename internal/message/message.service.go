package message

import (
	"errors"
	dto "go-app/internal/dto"
	messageRepo "go-app/internal/message/repo"
	"go-app/internal/schema"
	"go-app/pkg/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IChannelRepo interface {
	UpdateChannel(channelId primitive.ObjectID, updateDto dto.UpdateChannelDto) *schema.DbChannel
}

type IChannelMemberRepo interface {
	CheckUserInChannel(channelId primitive.ObjectID, userId primitive.ObjectID) bool
	GetChannelMembers(channelId primitive.ObjectID) *[]schema.DbChannelMember
}

type IMessageService interface {
	CheckUserInChannel(channelId string, userId string) bool
	// Lấy danh sách userId của tất cả member active trong channel (dùng để bắn realtime)
	GetMemberIds(channelId string) []string
	// Message CRUD
	CreateMessage(userId string, createDto dto.CreateMessageDto) (*schema.Message, error)
	UpdateMessage(msgId string, userId string, updateDto dto.UpdateMessageDto) (*schema.Message, error)

	// Thu hồi tin nhắn: chỉ người gửi được dùng, soft-delete message → mọi người đều thấy bị thu hồi
	RecallMessage(msgId string, userId string) *dto.MessageResponseDto

	// Ẩn tin nhắn chỉ với phía user hiện tại (dù là tin của mình hay người khác)
	// Ghi vào message_extra, không ảnh hưởng phía người khác
	HideMessageForMe(msgId string, userId string, channelId string) error

	// Xóa đoạn chat: ẩn tất cả messages có msg_seq <= currentSeq với user hiện tại
	DeleteChatHistory(userId string, channelId string, upToSeq int64) error

	// Query
	GetMessageByID(msgId string) (*dto.MessageResponseDto, error)
	GetMessagesByChannel(channelId string, userId string, limit int64, beforeSeq int64) (*[]dto.MessageResponseDto, error)
	GetDeletedMessageIDsByUser(userId string, channelId string) ([]string, error)
	GetChatOffset(userId string, channelId string) (*schema.MessageOffsets, error)
}

type MessageService struct {
	messageRepo       messageRepo.IMessageRepo
	messageOffsetRepo messageRepo.IMessageOffsetRepo
	messageExtraRepo  messageRepo.IMessageExtraRepo
	channelRepo       IChannelRepo
	channelMemberRepo IChannelMemberRepo
}

func NewMessageService(
	messageRepo messageRepo.IMessageRepo,
	messageOffsetRepo messageRepo.IMessageOffsetRepo,
	messageExtraRepo messageRepo.IMessageExtraRepo,
	channelRepo IChannelRepo,
	channelMemberRepo IChannelMemberRepo,
) IMessageService {
	return &MessageService{
		messageRepo:       messageRepo,
		messageOffsetRepo: messageOffsetRepo,
		messageExtraRepo:  messageExtraRepo,
		channelRepo:       channelRepo,
		channelMemberRepo: channelMemberRepo,
	}
}

func (s *MessageService) CheckUserInChannel(channelId string, userId string) bool {
	cID := utils.ObjectIDFromHex(channelId)
	uID := utils.ObjectIDFromHex(userId)
	if cID == primitive.NilObjectID || uID == primitive.NilObjectID {
		return false
	}
	return s.channelMemberRepo.CheckUserInChannel(cID, uID)
}

// GetMemberIds trả về danh sách userId (string) của tất cả member active trong channel
// Dùng để bắn realtime qua WebSocket sau khi tạo/cập nhật tin nhắn
func (s *MessageService) GetMemberIds(channelId string) []string {
	cID := utils.ObjectIDFromHex(channelId)
	if cID == primitive.NilObjectID {
		return nil
	}

	members := s.channelMemberRepo.GetChannelMembers(cID)
	if members == nil {
		return nil
	}

	ids := make([]string, 0, len(*members))
	for _, m := range *members {
		ids = append(ids, m.UserID.Hex())
	}
	return ids
}

// CreateMessage tạo một tin nhắn mới trong channel.
func (s *MessageService) CreateMessage(userId string, createDto dto.CreateMessageDto) (*schema.Message, error) {
	if !s.CheckUserInChannel(createDto.ChannelId, userId) {
		return nil, errors.New("permission denied: user is not a member of this channel")
	}
	// create message
	msg := s.messageRepo.CreateMessage(userId, createDto)
	if msg == nil {
		return nil, errors.New("failed to create message")
	}
	// update last msg in channel
	channel := s.channelRepo.UpdateChannel(msg.ChannelID, dto.UpdateChannelDto{
		LastMsgId:   msg.ID.Hex(),
		LastMsgSeq:  msg.MsgSeq,
		LastMsgTime: msg.CreatedAt,
	})
	if channel == nil {
		return nil, errors.New("failed to update channel")
	}
	return msg, nil
}

// UpdateMessage chỉnh sửa nội dung hoặc trạng thái của một tin nhắn.
// Chỉ cho phép người gửi (userId == msg.FromID) chỉnh sửa nội dung.
func (s *MessageService) UpdateMessage(
	msgId string,
	userId string,
	updateDto dto.UpdateMessageDto,
) (*schema.Message, error) {
	id := utils.ObjectIDFromHex(msgId)
	if id == primitive.NilObjectID {
		return nil, errors.New("invalid message id")
	}

	// Lấy message để kiểm tra quyền
	existing := s.messageRepo.GetMessageByID(id)
	if existing == nil {
		return nil, errors.New("message not found")
	}
	if existing.IsDelete {
		return nil, errors.New("message has been deleted")
	}
	// chỉ người gửi mới được chính sửa msg gốc
	if existing.FromID.Hex() != userId {
		return nil, errors.New("permission denied: only the sender can edit message")
	}

	// kiểm tra update dto có tồn tại k
	if updateDto.Content == "" && updateDto.Status == "" {
		return nil, errors.New("update data not exist")
	}

	updated := s.messageRepo.UpdateMessage(id, updateDto)
	if updated == nil {
		return nil, errors.New("failed to update message")
	}
	return updated, nil
}

// RecallMessage thu hồi một tin nhắn do chính user gửi.
// Chỉ người gửi (FromID == userId) mới được phép thu hồi.
// Thực hiện soft-delete trên message → tất cả mọi người trong channel đều thấy tin bị thu hồi.
func (s *MessageService) RecallMessage(msgId string, userId string) *dto.MessageResponseDto {
	id := utils.ObjectIDFromHex(msgId)
	if id == primitive.NilObjectID {
		return nil
	}

	existing := s.messageRepo.GetMessageByID(id)
	if existing == nil {
		return nil
	}
	if existing.IsDelete {
		return nil
	}

	// Chỉ người gửi mới được thu hồi
	if existing.FromID.Hex() != userId {
		return nil
	}

	ok := s.messageRepo.DeleteMessage(id)
	if !ok {
		return nil
	}

	// Trả về DTO của message đã bị xoá
	return &dto.MessageResponseDto{
		MsgId:          existing.ID.Hex(),
		ChannelId:      existing.ChannelID.Hex(),
		FromId:         existing.FromID.Hex(),
		Content:        existing.Content,
		MsgType:        string(existing.MsgType),
		MsgSeq:         existing.MsgSeq,
		Status:         string(existing.Status),
		IsDelete:       true,
		RepliedToMsgId: existing.RepliedToMsgID.Hex(),
	}
}

// HideMessageForMe ẩn một tin nhắn chỉ với phía user hiện tại.
// Áp dụng với mọi tin nhắn trong channel (của mình hoặc của người khác).
// Ghi vào message_extra để đánh dấu; không ảnh hưởng đến trải nghiệm của các user khác.
func (s *MessageService) HideMessageForMe(msgId string, userId string, channelId string) error {
	msgID := utils.ObjectIDFromHex(msgId)
	if msgID == primitive.NilObjectID {
		return errors.New("invalid message id")
	}

	// Kiểm tra message tồn tại
	existing := s.messageRepo.GetMessageByID(msgID)
	if existing == nil {
		return errors.New("message not found")
	}

	userID := utils.ObjectIDFromHex(userId)
	channelID := utils.ObjectIDFromHex(channelId)

	if userID == primitive.NilObjectID || channelID == primitive.NilObjectID {
		return errors.New("invalid user or channel id")
	}

	if !s.CheckUserInChannel(channelId, userId) {
		return errors.New("permission denied: user is not a member of this channel")
	}

	// Idempotent: kiểm tra đã ẩn chưa
	extras := s.messageExtraRepo.GetMessageExtrasByUser(userID, channelID)
	if extras != nil {
		for _, e := range *extras {
			if e.MsgID == msgID {
				// Đã ẩn rồi, bỏ qua
				return nil
			}
		}
	}

	// Tạo message_extra để đánh dấu message này bị ẩn với user
	extra := s.messageExtraRepo.CreateMessageExtra(dto.CreateMessageExtraDto{
		UserId:    userId,
		ChannelId: channelId,
		MsgId:     msgId,
		Version:   1,
	})
	if extra == nil {
		return errors.New("failed to hide message")
	}
	return nil
}

// DeleteChatHistory xóa đoạn chat bằng cách upsert message_offset.
// Tất cả messages có msg_seq <= upToSeq sẽ bị ẩn với user này.
// upToSeq = 0 sẽ dùng msg_seq của tin nhắn mới nhất trong channel.
func (s *MessageService) DeleteChatHistory(userId string, channelId string, upToSeq int64) error {
	channelID := utils.ObjectIDFromHex(channelId)
	if channelID == primitive.NilObjectID {
		return errors.New("invalid channel id")
	}
	userID := utils.ObjectIDFromHex(userId)
	if userID == primitive.NilObjectID {
		return errors.New("invalid user id")
	}

	if !s.CheckUserInChannel(channelId, userId) {
		return errors.New("permission denied: user is not a member of this channel")
	}

	// Nếu upToSeq không được truyền, lấy msg_seq của tin nhắn mới nhất
	if upToSeq <= 0 {
		latest := s.messageRepo.GetMessagesByChannel(channelID, 1, 0) // lấy msg mới nhất
		if latest == nil || len(*latest) == 0 {
			// Không có message nào, không cần làm gì
			return nil
		}
		upToSeq = (*latest)[0].MsgSeq
	}

	// Kiểm tra đã có offset cho cặp (userId, channelId) chưa
	existing := s.messageOffsetRepo.GetMessageOffsetByUserAndChannel(userID, channelID)

	if existing == nil {
		// Tạo mới
		offset := s.messageOffsetRepo.CreateMessageOffset(dto.CreateMessageOffsetDto{
			UserId:    userId,
			ChannelId: channelId,
			Offset:    upToSeq,
			Version:   1,
		})
		if offset == nil {
			return errors.New("failed to create chat offset")
		}
		return nil
	}

	// Chỉ cập nhật nếu upToSeq lớn hơn offset hiện tại (đẩy offset lên, không lùi)
	if upToSeq <= existing.Offset {
		return nil
	}

	updated := s.messageOffsetRepo.UpdateMessageOffset(
		existing.ID,
		dto.UpdateMessageOffsetDto{
			Offset:  upToSeq,
			Version: existing.Version + 1,
			Sync:    false,
		})
	if updated == nil {
		return errors.New("failed to update chat offset")
	}
	return nil
}

// GetMessageByID lấy thông tin một message bằng ID.
// Chỉ trả về nếu message còn tồn tại (IsDelete = false).
func (s *MessageService) GetMessageByID(msgId string) (*dto.MessageResponseDto, error) {
	id := utils.ObjectIDFromHex(msgId)
	if id == primitive.NilObjectID {
		return nil, errors.New("invalid message id")
	}
	msg := s.messageRepo.GetMessageByID(id)
	if msg == nil {
		return nil, errors.New("message not found")
	}
	if msg.IsDelete {
		return &dto.MessageResponseDto{
			MsgId:     msg.ID.Hex(),
			ChannelId: msg.ChannelID.Hex(),
			FromId:    msg.FromID.Hex(),
			Content:   "This message has been deleted",
			MsgType:   "",
			MsgSeq:    msg.MsgSeq,
			Status:    string(msg.Status),
			IsDelete:  msg.IsDelete,
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
		}, nil
	}
	return &dto.MessageResponseDto{
		MsgId:     msg.ID.Hex(),
		ChannelId: msg.ChannelID.Hex(),
		FromId:    msg.FromID.Hex(),
		Content:   msg.Content,
		MsgType:   string(msg.MsgType),
		MsgSeq:    msg.MsgSeq,
		Status:    string(msg.Status),
		IsDelete:  msg.IsDelete,
		CreatedAt: msg.CreatedAt.Format(time.RFC3339),
	}, nil
}

// GetMessagesByChannel lấy danh sách tin nhắn của channel, lọc ra những message
// đã bị user xóa rời rạc (message_extra) hoặc nằm trong vùng offset (message_offset).
func (s *MessageService) GetMessagesByChannel(
	channelId string,
	userId string,
	limit int64,
	beforeSeq int64,
) (*[]dto.MessageResponseDto, error) {
	channelID := utils.ObjectIDFromHex(channelId)
	if channelID == primitive.NilObjectID {
		return nil, errors.New("invalid channel id")
	}
	userID := utils.ObjectIDFromHex(userId)
	if userID == primitive.NilObjectID {
		return nil, errors.New("invalid user id")
	}

	if !s.CheckUserInChannel(channelId, userId) {
		return nil, errors.New("permission denied: user is not a member of this channel")
	}

	// Lấy offset của user trong channel (nếu có)
	var userOffset int64 = 0
	offset := s.messageOffsetRepo.GetMessageOffsetByUserAndChannel(userID, channelID)
	if offset != nil {
		userOffset = offset.Offset
	}

	// Lấy danh sách msgId bị xóa rời rạc bởi user
	deletedMsgIDs := map[primitive.ObjectID]bool{}
	extras := s.messageExtraRepo.GetMessageExtrasByUser(userID, channelID)
	if extras != nil {
		for _, e := range *extras {
			deletedMsgIDs[e.MsgID] = true
		}
	}

	// Lấy messages từ repo (lấy dư để bù cho các messages bị lọc)
	fetchLimit := limit
	if fetchLimit <= 0 {
		fetchLimit = 20
	}

	messages := s.messageRepo.GetMessagesByChannel(channelID, fetchLimit*2, beforeSeq)
	if messages == nil {
		return nil, nil
	}

	// Lọc messages theo offset và message_extra
	filtered := make([]dto.MessageResponseDto, 0, len(*messages))
	for _, msg := range *messages {
		// Bỏ qua nếu msg_seq nằm trong vùng đã bị xóa bởi offset
		if msg.MsgSeq <= userOffset {
			continue
		}
		// Bỏ qua nếu message đã bị user xóa rời rạc
		if deletedMsgIDs[msg.ID] {
			continue
		}
		filtered = append(filtered, dto.MessageResponseDto{
			MsgId:          msg.ID.Hex(),
			ChannelId:      msg.ChannelID.Hex(),
			FromId:         msg.FromID.Hex(),
			Content:        msg.Content,
			MsgType:        string(msg.MsgType),
			MsgSeq:         msg.MsgSeq,
			Status:         string(msg.Status),
			IsDelete:       msg.IsDelete,
			RepliedToMsgId: msg.RepliedToMsgID.Hex(),
			CreatedAt:      msg.CreatedAt.Format(time.RFC3339),
		})
		if int64(len(filtered)) >= fetchLimit {
			break
		}
	}

	if len(filtered) == 0 {
		return nil, nil
	}
	return &filtered, nil
}

// GetDeletedMessageIDsByUser trả về danh sách các msgId bị user xóa rời rạc trong channel.
func (s *MessageService) GetDeletedMessageIDsByUser(userId string, channelId string) ([]string, error) {
	userObjID := utils.ObjectIDFromHex(userId)
	if userObjID == primitive.NilObjectID {
		return nil, errors.New("invalid user id")
	}
	channelObjID := utils.ObjectIDFromHex(channelId)
	if channelObjID == primitive.NilObjectID {
		return nil, errors.New("invalid channel id")
	}

	extras := s.messageExtraRepo.GetMessageExtrasByUser(userObjID, channelObjID)
	if extras == nil {
		return []string{}, nil
	}

	ids := make([]string, 0, len(*extras))
	for _, e := range *extras {
		ids = append(ids, e.MsgID.Hex())
	}
	return ids, nil
}

// GetChatOffset trả về thông tin offset (điểm xóa đoạn chat) của user trong channel.
func (s *MessageService) GetChatOffset(userId string, channelId string) (*schema.MessageOffsets, error) {
	userObjID := utils.ObjectIDFromHex(userId)
	if userObjID == primitive.NilObjectID {
		return nil, errors.New("invalid user id")
	}
	channelObjID := utils.ObjectIDFromHex(channelId)
	if channelObjID == primitive.NilObjectID {
		return nil, errors.New("invalid channel id")
	}

	offset := s.messageOffsetRepo.GetMessageOffsetByUserAndChannel(userObjID, channelObjID)
	if offset == nil {
		return nil, nil
	}
	return offset, nil
}
