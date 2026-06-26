package repo

import (
	"context"
	"go-app/global"
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/pkg/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type IMessageExtraRepo interface {
	CreateMessageExtra(createDto dto.CreateMessageExtraDto) *schema.MessageExtras
	GetMessageExtraByID(id primitive.ObjectID) *schema.MessageExtras
	GetMessageExtrasByMsg(msgId primitive.ObjectID) *[]schema.MessageExtras
	GetMessageExtrasByUser(userId primitive.ObjectID, channelId primitive.ObjectID) *[]schema.MessageExtras
	UpdateMessageExtra(id primitive.ObjectID, updateDto dto.UpdateMessageExtraDto) *schema.MessageExtras
	DeleteMessageExtra(id primitive.ObjectID) bool
}

type MessageExtraRepo struct{}

// CreateMessageExtra implements [IMessageExtraRepo].
func (r *MessageExtraRepo) CreateMessageExtra(createDto dto.CreateMessageExtraDto) *schema.MessageExtras {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userId := utils.ObjectIDFromHex(createDto.UserId)
	if userId == primitive.NilObjectID {
		return nil
	}
	channelId := utils.ObjectIDFromHex(createDto.ChannelId)
	if channelId == primitive.NilObjectID {
		return nil
	}
	msgId := utils.ObjectIDFromHex(createDto.MsgId)
	if msgId == primitive.NilObjectID {
		return nil
	}

	extra := &schema.MessageExtras{
		ID:        primitive.NewObjectID(),
		UserID:    userId,
		ChannelID: channelId,
		MsgID:     msgId,
		Version:   createDto.Version,
		Sync:      false,
	}

	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageExtras)
	_, err := collection.InsertOne(ctx, extra)
	if err != nil {
		return nil
	}
	return r.GetMessageExtraByID(extra.ID)
}

// GetMessageExtraByID implements [IMessageExtraRepo].
func (r *MessageExtraRepo) GetMessageExtraByID(id primitive.ObjectID) *schema.MessageExtras {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var extra *schema.MessageExtras
	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageExtras)
	collection.FindOne(ctx, bson.M{"_id": id}).Decode(&extra)
	if extra == nil {
		return nil
	}
	return extra
}

// GetMessageExtrasByMsg implements [IMessageExtraRepo].
// Lấy tất cả trạng thái đọc/nhận của một tin nhắn.
func (r *MessageExtraRepo) GetMessageExtrasByMsg(msgId primitive.ObjectID) *[]schema.MessageExtras {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var extras []schema.MessageExtras
	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageExtras)
	cursor, err := collection.Find(ctx, bson.M{"msg_id": msgId})
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &extras); err != nil {
		return nil
	}
	if len(extras) == 0 {
		return nil
	}
	return &extras
}

// GetMessageExtrasByUser implements [IMessageExtraRepo].
// Lấy tất cả message extras của một user trong một channel.
func (r *MessageExtraRepo) GetMessageExtrasByUser(userId primitive.ObjectID, channelId primitive.ObjectID) *[]schema.MessageExtras {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"uid":        userId,
		"channel_id": channelId,
	}

	var extras []schema.MessageExtras
	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageExtras)
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &extras); err != nil {
		return nil
	}
	if len(extras) == 0 {
		return nil
	}
	return &extras
}

// UpdateMessageExtra implements [IMessageExtraRepo].
func (r *MessageExtraRepo) UpdateMessageExtra(id primitive.ObjectID, updateDto dto.UpdateMessageExtraDto) *schema.MessageExtras {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{}
	if updateDto.Version != 0 {
		update["version"] = updateDto.Version
	}
	update["sync"] = updateDto.Sync

	if len(update) == 0 {
		return nil
	}

	var extra *schema.MessageExtras
	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageExtras)
	collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	).Decode(&extra)
	if extra == nil {
		return nil
	}
	return extra
}

// DeleteMessageExtra implements [IMessageExtraRepo].
func (r *MessageExtraRepo) DeleteMessageExtra(id primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageExtras)
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false
	}
	if result.DeletedCount == 0 {
		return false
	}
	return true
}

func NewMessageExtraRepo() IMessageExtraRepo {
	return &MessageExtraRepo{}
}
