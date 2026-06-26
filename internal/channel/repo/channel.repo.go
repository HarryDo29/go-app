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

type IChannelRepo interface {
	CreateChannel(channelDto dto.CreateChannelDto) *schema.DbChannel
	GetChannelById(id primitive.ObjectID) *schema.DbChannel
	GetChannels(channelIDs []primitive.ObjectID) *[]schema.DbChannel
	UpdateChannel(channelId primitive.ObjectID, updateDto dto.UpdateChannelDto) *schema.DbChannel
	DeleteChannel(channelId primitive.ObjectID) bool
}

type ChannelRepo struct{}

// CreateChannel implements [IChannelRepo].
func (c *ChannelRepo) CreateChannel(channelDto dto.CreateChannelDto) *schema.DbChannel {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channelKey := utils.ObjectIDFromHex(channelDto.ChannelKey)
	if channelKey == primitive.NilObjectID {
		return nil
	}
	channel := &schema.DbChannel{
		ID:          primitive.NewObjectID(),
		ChannelType: schema.ChannelType(channelDto.ChannelType),
		ChannelKey:  channelKey,
		LastMsgID:   primitive.NilObjectID,
		LastMsgSeq:  0,
		LastMsgTime: time.Time{},
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	collection := global.Mgo.Database.Collection(schema.CollectionNameChannel)
	_, err := collection.InsertOne(ctx, channel)
	if err != nil {
		return nil
	}
	return c.GetChannelById(channel.ID)
}

func (c *ChannelRepo) GetChannels(channelIDs []primitive.ObjectID) *[]schema.DbChannel {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var channels []schema.DbChannel
	collection := global.Mgo.Database.Collection(schema.CollectionNameChannel)

	opts := options.Find().SetSort(
		bson.D{{Key: "last_msg_time", Value: -1}},
	)
	findResult, err := collection.Find(
		ctx,
		bson.M{
			"_id":       bson.M{"$in": channelIDs},
			"is_active": true,
		},
		opts,
	)
	if err != nil {
		fmt.Println("err: ", err.Error())
		return nil
	}
	defer findResult.Close(ctx)

	err = findResult.All(ctx, &channels)
	if err != nil {
		return nil
	}
	if len(channels) == 0 {
		return nil
	}
	return &channels
}

// GetChannelById implements [IChannelRepo].
func (c *ChannelRepo) GetChannelById(id primitive.ObjectID) *schema.DbChannel {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var channel *schema.DbChannel
	collection := global.Mgo.Database.Collection(schema.CollectionNameChannel)
	collection.FindOne(ctx, bson.M{"_id": id}).Decode(&channel)
	if channel == nil {
		return nil
	}
	return channel
}

// UpdateChannel implements [IChannelRepo].
func (c *ChannelRepo) UpdateChannel(channelId primitive.ObjectID, updateDto dto.UpdateChannelDto) *schema.DbChannel {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{}
	if updateDto.LastMsgId != "" {
		lastMsgId := utils.ObjectIDFromHex(updateDto.LastMsgId)
		if lastMsgId == primitive.NilObjectID {
			return nil
		}
		update["last_msg_id"] = lastMsgId
	}
	if updateDto.LastMsgSeq != 0 {
		update["last_msg_seq"] = updateDto.LastMsgSeq
	}
	if !updateDto.LastMsgTime.IsZero() {
		update["last_msg_time"] = updateDto.LastMsgTime
	}
	if len(update) == 0 {
		return nil
	}

	var channel *schema.DbChannel
	collection := global.Mgo.Database.Collection(schema.CollectionNameChannel)
	collection.FindOneAndUpdate(ctx, bson.M{"_id": channelId}, bson.M{"$set": update}).Decode(&channel)
	if channel == nil {
		return nil
	}
	return channel
}

// DeleteChannel implements [IChannelRepo].
func (c *ChannelRepo) DeleteChannel(channelId primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameChannel)
	result, err := collection.UpdateOne(ctx,
		bson.M{"_id": channelId},
		bson.M{
			"$set": bson.M{
				"is_active":  false,
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

func NewChannelRepo() IChannelRepo {
	return &ChannelRepo{}
}
