package repo

import (
	"context"
	"go-app/global"
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/pkg/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IMessageRepo interface {
	CreateMessage(userId string, createDto dto.CreateMessageDto) *schema.Message
	GetMessageByID(id primitive.ObjectID) *schema.Message
	GetMessagesByChannel(channelId primitive.ObjectID, limit int64, beforeSeq int64) *[]schema.Message
	UpdateMessage(id primitive.ObjectID, updateDto dto.UpdateMessageDto) *schema.Message
	DeleteMessage(id primitive.ObjectID) bool
}

type MessageRepo struct{}

// CreateMessage implements [IMessageRepo].
func (r *MessageRepo) CreateMessage(userId string, createDto dto.CreateMessageDto) *schema.Message {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channelId := utils.ObjectIDFromHex(createDto.ChannelId)
	if channelId == primitive.NilObjectID {
		return nil
	}
	fromId := utils.ObjectIDFromHex(userId)
	if fromId == primitive.NilObjectID {
		return nil
	}

	msg := &schema.Message{
		ID:        primitive.NewObjectID(),
		ChannelID: channelId,
		FromID:    fromId,
		Content:   createDto.Content,
		MsgType:   schema.MessageType(createDto.MsgType),
		MsgSeq:    time.Now().UnixMilli(), // auto-increment via timestamp
		Status:    schema.MessageStatusSent,
		IsDelete:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if createDto.RepliedToMsgId != "" {
		repliedID := utils.ObjectIDFromHex(createDto.RepliedToMsgId)
		if repliedID != primitive.NilObjectID {
			msg.RepliedToMsgID = repliedID
		}
	} else {
		msg.RepliedToMsgID = primitive.NilObjectID
	}

	collection := global.Mgo.Database.Collection(schema.CollectionNameMessage)
	_, err := collection.InsertOne(ctx, msg)
	if err != nil {
		return nil
	}
	return r.GetMessageByID(msg.ID)
}

// GetMessageByID implements [IMessageRepo].
func (r *MessageRepo) GetMessageByID(id primitive.ObjectID) *schema.Message {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var msg *schema.Message
	collection := global.Mgo.Database.Collection(schema.CollectionNameMessage)
	collection.FindOne(ctx, bson.M{"_id": id}).Decode(&msg)
	if msg == nil {
		return nil
	}
	return msg
}

// GetMessagesByChannel implements [IMessageRepo].
// Returns messages for a channel, ordered by msg_seq descending.
// If beforeSeq > 0, fetches messages with msg_seq < beforeSeq (pagination).
func (r *MessageRepo) GetMessagesByChannel(
	channelId primitive.ObjectID,
	limit int64,
	beforeSeq int64,
) *[]schema.Message {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"channel_id": channelId,
		// "is_delete":  false,
	} // set filter

	if beforeSeq > 0 {
		filter["msg_seq"] = bson.M{"$lt": beforeSeq}
		// '$lt': less than --> lấy các msg có msg_seq bé hơn dto.msg_seq
	}

	if limit <= 0 {
		limit = 20
	}

	opts := options.Find().
		SetSort(bson.D{
			{Key: "msg_seq", Value: -1},
		}).             // set sort tăng dần
		SetLimit(limit) // set limit số lượng tin nhắn

	collection := global.Mgo.Database.Collection(schema.CollectionNameMessage)
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	var messages []schema.Message
	if err = cursor.All(ctx, &messages); err != nil {
		return nil
	}
	if len(messages) == 0 {
		return nil
	}
	return &messages
}

// UpdateMessage implements [IMessageRepo].
func (r *MessageRepo) UpdateMessage(
	id primitive.ObjectID,
	updateDto dto.UpdateMessageDto,
) *schema.Message {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{}
	if updateDto.Content != "" {
		update["content"] = updateDto.Content
	}
	if updateDto.Status != "" {
		update["status"] = schema.MessageStatus(updateDto.Status)
	}
	if len(update) == 0 {
		return nil
	}
	update["updated_at"] = time.Now() // cập nhật lại

	var msg *schema.Message
	collection := global.Mgo.Database.Collection(schema.CollectionNameMessage)
	collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update}, // cập nhật
		options.FindOneAndUpdate().SetReturnDocument(options.After), // trả về bản ghi sau khi update
	).Decode(&msg)
	if msg == nil {
		return nil
	}
	return msg
}

// DeleteMessage implements [IMessageRepo].
// Soft delete: sets is_delete = true.
func (r *MessageRepo) DeleteMessage(id primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameMessage)
	result, err := collection.UpdateOne(ctx,
		bson.M{"_id": id},
		bson.M{
			"$set": bson.M{
				"is_delete":  true,
				"updated_at": time.Now(),
			},
		})
	if err != nil {
		return false
	}
	if result.ModifiedCount == 0 {
		return false
	}
	return true
}

func NewMessageRepo() IMessageRepo {
	return &MessageRepo{}
}
