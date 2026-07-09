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
)

type IChannelMemberRepo interface {
	CreateChannelMember(channelMemberDto dto.CreateChannelMemberDto) *[]schema.DbChannelMember
	GetChannelMemberById(id primitive.ObjectID) *schema.DbChannelMember
	GetChannelMembers(channelId primitive.ObjectID) *[]schema.DbChannelMember
	GetChannelIds(userId primitive.ObjectID) *[]primitive.ObjectID
	UpdateChannelMember(channelMemberId primitive.ObjectID, updateDto dto.UpdateChannelMemberDto) *schema.DbChannelMember
	DeleteChannelMember(channelMemberId primitive.ObjectID) bool
	CheckUserInChannel(channelId primitive.ObjectID, userId primitive.ObjectID) bool
}

type ChannelMemberRepo struct{}

// CreateChannelMember implements [IChannelMemberRepo].
func (c *ChannelMemberRepo) CreateChannelMember(channelMemberDto dto.CreateChannelMemberDto) *[]schema.DbChannelMember {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	channelId := utils.ObjectIDFromHex(channelMemberDto.ChannelId)
	if channelId == primitive.NilObjectID {
		return nil
	}
	members := make([]interface{}, 0)
	for _, userId := range channelMemberDto.UserIds {
		userId := utils.ObjectIDFromHex(userId)
		if userId == primitive.NilObjectID {
			continue
		}
		if c.CheckUserInChannel(channelId, userId) {
			continue
		}
		role := schema.ChannelMemberRole(channelMemberDto.Role)
		if role == "" {
			role = schema.ChannelMemberRoleMember
		}
		status := schema.ChannelMemberStatus(channelMemberDto.Status)
		if status == "" {
			status = schema.ChannelMemberStatusActive
		}

		member := &schema.DbChannelMember{
			ID:        primitive.NewObjectID(),
			ChannelID: channelId,
			UserID:    userId,
			Role:      role,
			Status:    status,
			JoinedAt:  time.Now(),
		}
		members = append(members, member)
	}

	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelMembers)
	if len(members) == 0 {
		return c.GetChannelMembers(channelId)
	}
	result, err := collection.InsertMany(ctx, members)
	if err != nil {
		return nil
	}
	if len(result.InsertedIDs) == 0 {
		return nil
	}
	return c.GetChannelMembers(channelId)
}

func (c *ChannelMemberRepo) IsMemberInChannel(
	channelId primitive.ObjectID,
	userId primitive.ObjectID,
) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelMembers)
	var member schema.DbChannelMember
	err := collection.FindOne(ctx, bson.M{
		"channel_id": channelId,
		"user_id":    userId,
		"status":     schema.ChannelMemberStatusActive,
		"left_at":    nil,
	}).Decode(&member)

	if err != nil {
		return false
	}
	if member != (schema.DbChannelMember{}) {
		return false
	}
	return true
}

// GetChannelMemberById implements [IChannelMemberRepo].
func (c *ChannelMemberRepo) GetChannelMemberById(id primitive.ObjectID) *schema.DbChannelMember {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var member *schema.DbChannelMember
	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelMembers)
	collection.FindOne(ctx, bson.M{"_id": id}).Decode(&member)
	if member == nil {
		return nil
	}
	return member
}

func (c *ChannelMemberRepo) GetChannelIds(userId primitive.ObjectID) *[]primitive.ObjectID {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelMembers)

	cursor, err := collection.Find(ctx, bson.M{"user_id": userId})
	if err != nil {
		fmt.Println("err", err.Error())
		return nil
	}
	defer cursor.Close(ctx)

	var results []schema.DbChannelMember
	err = cursor.All(ctx, &results)
	if err != nil {
		return nil
	}

	channelIDs := make([]primitive.ObjectID, 0, len(results))
	for _, res := range results {
		channelIDs = append(channelIDs, res.ChannelID)
	}
	return &channelIDs
}

func (c *ChannelMemberRepo) GetChannelMembers(channelId primitive.ObjectID) *[]schema.DbChannelMember {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var members []schema.DbChannelMember
	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelMembers)

	cursor, err := collection.Find(ctx,
		bson.M{
			"channel_id": channelId,
			"status":     schema.ChannelMemberStatusActive,
			"left_at":    nil,
		})

	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	err = cursor.All(ctx, &members)
	if err != nil {
		fmt.Println("hello2")
		return nil
	}
	return &members
}

func (c *ChannelMemberRepo) CheckUserInChannel(channelId primitive.ObjectID, userId primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelMembers)
	var member schema.DbChannelMember
	err := collection.FindOne(ctx, bson.M{
		"channel_id": channelId,
		"user_id":    userId,
		"status":     schema.ChannelMemberStatusActive,
		"left_at":    nil,
	}).Decode(&member)

	if err != nil {
		return false
	}
	return true
}

// UpdateChannelMember implements [IChannelMemberRepo].
func (c *ChannelMemberRepo) UpdateChannelMember(channelMemberId primitive.ObjectID, updateDto dto.UpdateChannelMemberDto) *schema.DbChannelMember {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	update := bson.M{}
	if updateDto.Role != "" {
		update["role"] = schema.ChannelMemberRole(updateDto.Role)
	}
	if updateDto.Status != "" {
		update["status"] = schema.ChannelMemberStatus(updateDto.Status)
		if schema.ChannelMemberStatus(updateDto.Status) == schema.ChannelMemberStatusLeft || schema.ChannelMemberStatus(updateDto.Status) == schema.ChannelMemberStatusKicked {
			now := time.Now()
			update["left_at"] = &now
		}
	}
	if len(update) == 0 {
		return nil
	}

	var member *schema.DbChannelMember
	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelMembers)
	collection.FindOneAndUpdate(ctx, bson.M{"_id": channelMemberId},
		bson.M{
			"$set":       update,
			"updated_at": time.Now(),
		}).Decode(&member)
	if member == nil {
		return nil
	}
	return member
}

// DeleteChannelMember implements [IChannelMemberRepo].
func (c *ChannelMemberRepo) DeleteChannelMember(channelMemberId primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameChannelMembers)
	result, err := collection.UpdateOne(ctx, bson.M{"_id": channelMemberId},
		bson.M{
			"$set": bson.M{
				"status":  schema.ChannelMemberStatusLeft,
				"role":    "",
				"left_at": time.Now(),
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

func NewChannelMemberRepo() IChannelMemberRepo {
	return &ChannelMemberRepo{}
}
