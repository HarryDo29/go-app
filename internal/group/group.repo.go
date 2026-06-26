package group

import (
	"context"
	"fmt"
	"go-app/global"
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/pkg/utils"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type IGroupRepo interface {
	CreateGroup(groupDto dto.CreateGroupDto) *schema.DbGroup
	GetGroupByID(ID primitive.ObjectID) *schema.DbGroup
	UpdateGroup(groupId primitive.ObjectID, dto dto.UpdateGroupDto) *schema.DbGroup
	DeleteGroup(groupId primitive.ObjectID) bool
}

type GroupRepo struct{}

// CreateGroup implements [IGroupRepo].
func (g *GroupRepo) CreateGroup(groupDto dto.CreateGroupDto) *schema.DbGroup {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ownerId := utils.ObjectIDFromHex(groupDto.OwnerId)
	if ownerId == primitive.NilObjectID {
		return nil
	}
	group := schema.DbGroup{
		ID:          primitive.NewObjectID(),
		Name:        groupDto.GroupName,
		OwnerID:     ownerId,
		MemberCount: 1,
		Status:      schema.GroupStatusActive,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	collection := global.Mgo.Database.Collection(schema.CollectionNameGroup)
	_, err := collection.InsertOne(ctx, &group)
	if err != nil {
		fmt.Println("lỗi ở phần insert group")
		return nil
	}
	return g.GetGroupByID(group.ID)
}

func (g *GroupRepo) GetGroupByID(ID primitive.ObjectID) *schema.DbGroup {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var group *schema.DbGroup
	collection := global.Mgo.Database.Collection(schema.CollectionNameGroup)
	err := collection.FindOne(ctx, bson.M{"_id": ID}).Decode(&group)
	if err != nil {
		fmt.Println("lỗi ở phần tìm kiếm group", err)
		return nil
	}

	return group
}

// UpdateGroup implements [IGroupRepo].
func (g *GroupRepo) UpdateGroup(groupId primitive.ObjectID, dto dto.UpdateGroupDto) *schema.DbGroup {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	updateData := bson.M{}
	if dto.GroupName != "" {
		updateData["name"] = dto.GroupName
	}
	if dto.MemberCount != 0 {
		updateData["member_count"] = dto.MemberCount
	}
	if dto.Status != "" {
		updateData["status"] = dto.Status
	}
	if len(updateData) <= 0 {
		return nil
	}
	updateData["updated_at"] = time.Now()
	collection := global.Mgo.Database.Collection(schema.CollectionNameGroup)
	result, err := collection.UpdateOne(ctx, bson.M{"_id": groupId}, bson.M{"$set": updateData})
	if err != nil {
		fmt.Println("err in group repo: ", err)
		return nil
	}
	if result.MatchedCount == 0 {
		return nil
	}
	return g.GetGroupByID(groupId)
}

// DeleteGroup implements [IGroupRepo].
func (g *GroupRepo) DeleteGroup(groupId primitive.ObjectID) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameGroup)
	result, err := collection.UpdateOne(ctx, bson.M{"_id": groupId}, bson.M{
		"$set": bson.M{
			"status":     schema.GroupStatusDeleted,
			"deleted_at": time.Now(),
		},
	})
	if err != nil {
		return false
	}
	if result.MatchedCount == 0 {
		return false
	}
	return true
}

func NewGroupRepo() IGroupRepo {
	return &GroupRepo{}
}
