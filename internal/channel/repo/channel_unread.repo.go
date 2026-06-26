package repo

import (
	"context"
	"fmt"
	"go-app/global"
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/pkg/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IChannelUnreadRepo interface {
	CreateChannelUnreads(unreadDto dto.CreateChannelUnreadDto) bool
	GetChannelUnread(unreadId primitive.ObjectID) *schema.DbChannelUnread
	GetChannelUnreads(userId primitive.ObjectID) *[]schema.DbChannelUnread
	UpdateChannelUnread(
		unreadId primitive.ObjectID,
		updateDto dto.UpdateChannelUnreadDto) *schema.DbChannelUnread
	DeleteChannelUnread(unreadId primitive.ObjectID) bool
}

type ChannelUnreadRepo struct{}

// CreateChannelUnread implements [IChannelUnreadRepo].
func (c *ChannelUnreadRepo) CreateChannelUnreads(unreadDto dto.CreateChannelUnreadDto) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channelId := utils.ObjectIDFromHex(unreadDto.ChannelId)
	if channelId == primitive.NilObjectID {
		fmt.Println("channelId is nil")
		return false
	}
	channelUnreads := make([]interface{}, 0)
	for _, userId := range unreadDto.UserIds {
		userId := utils.ObjectIDFromHex(userId)
		if userId == primitive.NilObjectID {
			continue
		}
		channelUnreads = append(channelUnreads, schema.DbChannelUnread{
			ID:          primitive.NewObjectID(),
			UserID:      userId,
			ChannelID:   channelId,
			LastMsgID:   primitive.NilObjectID,
			LastMsgTime: time.Unix(0, 0),
			IsActive:    true,
			Unread:      unreadDto.Unread,
			Version:     unreadDto.Version,
		})
	}

	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelUnread)
	result, err := collection.InsertMany(ctx, channelUnreads)
	if err != nil {
		fmt.Println("err:", err)
		return false
	}
	if len(result.InsertedIDs) == 0 {
		return false
	}
	return true
}

func (c *ChannelUnreadRepo) GetChannelUnread(unreadId primitive.ObjectID) *schema.DbChannelUnread {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var unread *schema.DbChannelUnread
	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelUnread)
	collection.FindOne(ctx,
		bson.M{
			"_id":       unreadId,
			"is_active": true,
		}).Decode(&unread)

	if unread == nil {
		return nil
	}
	return unread
}

// GetChannelUnread implements [IChannelUnreadRepo].
func (c *ChannelUnreadRepo) GetChannelUnreads(userId primitive.ObjectID) *[]schema.DbChannelUnread {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var unreads []schema.DbChannelUnread
	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelUnread)

	filter := bson.M{
		"user_id":   userId,
		"is_active": true,
	}
	if len(filter) == 0 {
		return nil
	}

	opt := options.Find().SetSort(bson.D{
		{Key: "last_msg_time", Value: -1},
	})

	cursor, err := collection.Find(ctx, filter, opt)
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)
	cursor.All(ctx, &unreads)

	if unreads == nil {
		return nil
	}
	return &unreads
}

// UpdateChannelUnread implements [IChannelUnreadRepo].
func (c *ChannelUnreadRepo) UpdateChannelUnread(
	unreadId primitive.ObjectID,
	updateDto dto.UpdateChannelUnreadDto,
) *schema.DbChannelUnread {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{}
	if !updateDto.LastMsgID.IsZero() {
		update["last_msg_id"] = updateDto.LastMsgID
	}
	if !updateDto.LastMsgTime.IsZero() {
		update["last_msg_time"] = updateDto.LastMsgTime
	}
	if updateDto.Unread >= 0 {
		update["unread"] = updateDto.Unread
	}
	if updateDto.Version != 0 {
		update["version"] = updateDto.Version
	}
	if len(update) == 0 {
		return nil
	}

	var unread *schema.DbChannelUnread
	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelUnread)
	collection.FindOneAndUpdate(ctx, bson.M{"_id": unreadId}, bson.M{"$set": update}).Decode(&unread)
	if unread == nil {
		return nil
	}
	return unread
}

// DeleteChannelUnread implements [IChannelUnreadRepo].
func (c *ChannelUnreadRepo) DeleteChannelUnread(unreadId primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelUnread)
	result, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": unreadId},
		bson.M{"is_active": false},
	)
	if err != nil {
		return false
	}
	if result.ModifiedCount == 0 {
		return false
	}
	return true
}

func NewChannelUnreadRepo() IChannelUnreadRepo {
	return &ChannelUnreadRepo{}
}
