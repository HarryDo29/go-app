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

type IMessageOffsetRepo interface {
	CreateMessageOffset(createDto dto.CreateMessageOffsetDto) *schema.MessageOffsets
	GetMessageOffsetByID(id primitive.ObjectID) *schema.MessageOffsets
	GetMessageOffsetByUserAndChannel(userId primitive.ObjectID, channelId primitive.ObjectID) *schema.MessageOffsets
	GetMessageOffsetsByUser(userId primitive.ObjectID) *[]schema.MessageOffsets
	UpdateMessageOffset(id primitive.ObjectID, updateDto dto.UpdateMessageOffsetDto) *schema.MessageOffsets
	DeleteMessageOffset(id primitive.ObjectID) bool
}

type MessageOffsetRepo struct{}

// CreateMessageOffset implements [IMessageOffsetRepo].
func (r *MessageOffsetRepo) CreateMessageOffset(createDto dto.CreateMessageOffsetDto) *schema.MessageOffsets {
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

	offset := &schema.MessageOffsets{
		ID:        primitive.NewObjectID(),
		UserID:    userId,
		ChannelID: channelId,
		Offset:    createDto.Offset,
		Version:   createDto.Version,
		Sync:      false,
	}

	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageOffsets)
	_, err := collection.InsertOne(ctx, offset)
	if err != nil {
		return nil
	}
	return r.GetMessageOffsetByID(offset.ID)
}

// GetMessageOffsetByID implements [IMessageOffsetRepo].
func (r *MessageOffsetRepo) GetMessageOffsetByID(id primitive.ObjectID) *schema.MessageOffsets {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var offset *schema.MessageOffsets
	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageOffsets)
	collection.FindOne(ctx, bson.M{"_id": id}).Decode(&offset)
	if offset == nil {
		return nil
	}
	return offset
}

// GetMessageOffsetByUserAndChannel implements [IMessageOffsetRepo].
// Lấy offset của một user trong một channel cụ thể.
func (r *MessageOffsetRepo) GetMessageOffsetByUserAndChannel(
	userId primitive.ObjectID,
	channelId primitive.ObjectID,
) *schema.MessageOffsets {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var offset *schema.MessageOffsets
	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageOffsets)
	collection.FindOne(ctx, bson.M{
		"uid":        userId,
		"channel_id": channelId,
	}).Decode(&offset)
	if offset == nil {
		return nil
	}
	return offset
}

// GetMessageOffsetsByUser implements [IMessageOffsetRepo].
// Lấy tất cả offsets của một user trên nhiều channel.
func (r *MessageOffsetRepo) GetMessageOffsetsByUser(userId primitive.ObjectID) *[]schema.MessageOffsets {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var offsets []schema.MessageOffsets
	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageOffsets)
	cursor, err := collection.Find(ctx, bson.M{"uid": userId})
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &offsets); err != nil {
		return nil
	}
	if len(offsets) == 0 {
		return nil
	}
	return &offsets
}

// UpdateMessageOffset implements [IMessageOffsetRepo].
func (r *MessageOffsetRepo) UpdateMessageOffset(
	id primitive.ObjectID,
	updateDto dto.UpdateMessageOffsetDto,
) *schema.MessageOffsets {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{}
	if updateDto.Offset != 0 {
		update["offset"] = updateDto.Offset
	}
	if updateDto.Version != 0 {
		update["version"] = updateDto.Version
	}
	update["sync"] = updateDto.Sync

	if len(update) == 0 {
		return nil
	}

	var offset *schema.MessageOffsets
	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageOffsets)
	collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": update},
	).Decode(&offset)
	if offset == nil {
		return nil
	}
	return offset
}

// DeleteMessageOffset implements [IMessageOffsetRepo].
func (r *MessageOffsetRepo) DeleteMessageOffset(id primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameMessageOffsets)
	result, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return false
	}
	if result.DeletedCount == 0 {
		return false
	}
	return true
}

func NewMessageOffsetRepo() IMessageOffsetRepo {
	return &MessageOffsetRepo{}
}
