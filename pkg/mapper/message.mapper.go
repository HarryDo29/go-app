package mapper

import (
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ToMessageResponseDto maps Message database schema model to MessageResponseDto
func ToMessageResponseDto(msg *schema.DbMessage) *dto.MessageResponseDto {
	if msg == nil {
		return nil
	}
	res := &dto.MessageResponseDto{
		MsgId:     msg.ID.Hex(),
		ChannelId: msg.ChannelID.Hex(),
		FromId:    msg.FromID.Hex(),
		Content:   msg.Content,
		MsgType:   string(msg.MsgType),
		MsgSeq:    msg.MsgSeq,
		Status:    string(msg.Status),
		IsDelete:  msg.IsDelete,
		CreatedAt: msg.CreatedAt.Format(time.RFC3339),
	}
	if msg.RepliedToMsgID != primitive.NilObjectID {
		res.RepliedToMsgId = msg.RepliedToMsgID.Hex()
	}
	return res
}

// ToMessageResponseDtoList maps a slice of Message database schema models to a slice of MessageResponseDto
func ToMessageResponseDtoList(messages []schema.DbMessage) []dto.MessageResponseDto {
	if messages == nil {
		return []dto.MessageResponseDto{}
	}
	result := make([]dto.MessageResponseDto, len(messages))
	for i, msg := range messages {
		if res := ToMessageResponseDto(&msg); res != nil {
			result[i] = *res
		}
	}
	return result
}
