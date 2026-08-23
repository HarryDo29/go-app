package user

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
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// INTERFACE
type IUserRepo interface {
	CreateUser(userDto dto.UserDto) *schema.DbUser
	GetUser(email string) *schema.DbUser
	GetUserById(id string) *schema.DbUser
	UpdateUser(id string, updateDto dto.UpdateUserDto) *schema.DbUser
	UpdateUserRole(id string, roleId string) *schema.DbUser
	DeleteUser(id string) bool
	SearchUsers(keyword string, userId string, limit int) *[]schema.DbUser
	GetUsers(limit int, index primitive.ObjectID) *[]schema.DbUser
	GetOverallStats() *dto.OverallStats
}

type userRepo struct{}

// CreateUser implements [IUserRepo].
func (ur *userRepo) CreateUser(userDto dto.UserDto) *schema.DbUser {
	var cacheRole map[string]dto.RoleDto

	// 1. Đọc danh sách vai trò từ in-memory Cache
	cached, ok := global.Cache.Get("roles")
	if ok {
		if mapData, ok := cached.(map[string]dto.RoleDto); ok {
			cacheRole = mapData
		}
	}

	// 2. Lookup role theo tên — value trong cache là dto.RoleDto
	roleData, exists := cacheRole[userDto.Role]
	if !exists {
		fmt.Println("Role not found:", userDto.Role)
		return nil
	}
	roleId := utils.ObjectIDFromHex(roleData.RoleId)

	user := schema.DbUser{
		ID:        primitive.NewObjectID(),
		UserName:  userDto.UserName,
		Password:  userDto.Password,
		Email:     userDto.Email,
		IsActive:  true,
		Role:      roleId,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameUser)
	_, err := collection.InsertOne(ctx, &user)
	if err != nil {
		fmt.Println("User create failed: ", err)
		return nil
	}
	return &user
}

// SearchUsers implements [IUserRepo].
func (ur *userRepo) SearchUsers(keyword string, userId string, limit int) *[]schema.DbUser {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameUser)
	userID := utils.ObjectIDFromHex(userId)

	// Create regex pattern for case-insensitive search
	filter := bson.M{
		// $or: MongoDB chỉ cần tìm 1 trong các điều kiện đúng
		"$or": []bson.M{
			{"user_name": bson.M{"$regex": keyword, "$options": "i"}},
			{"email": bson.M{"$regex": keyword, "$options": "i"}},
			// "$options": "i": i viết tắt là insensitive (không tính hoa thường)
		},
	}

	if userID != primitive.NilObjectID {
		filter = bson.M{
			"$and": []bson.M{
				filter,
				{"_id": bson.M{"$ne": userID}},
				// $ne: not equal (khác với userID)
			},
		}
	}

	opts := options.Find().SetLimit(int64(limit))
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		fmt.Println("Search users error: ", err)
		return nil
	}
	defer cursor.Close(ctx)

	var users []schema.DbUser
	if err := cursor.All(ctx, &users); err != nil {
		fmt.Println("Decode search users error: ", err)
		return nil
	}

	return &users
}

// GetUser implements [IUserRepo].
func (ur *userRepo) GetUser(email string) *schema.DbUser {
	var user schema.DbUser
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameUser)
	err := collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == nil {
		return &user
	}
	return nil
}

// GetUserById implements [IUserRepo].
func (ur *userRepo) GetUserById(id string) *schema.DbUser {
	objID := utils.ObjectIDFromHex(id)
	if objID == primitive.NilObjectID {
		return nil
	}

	var user schema.DbUser
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameUser)
	err := collection.FindOne(ctx, bson.M{"_id": objID}).Decode(&user)
	if err == nil {
		return &user
	}
	return nil
}

// UpdateUser implements [IUserRepo].
func (ur *userRepo) UpdateUser(id string, updateDto dto.UpdateUserDto) *schema.DbUser {
	objID := utils.ObjectIDFromHex(id)
	if objID == primitive.NilObjectID {
		return nil
	}

	updateData := bson.M{}

	if updateDto.UserName != "" {
		updateData["user_name"] = updateDto.UserName
	}
	if updateDto.AvatarUrl != "" {
		updateData["avatar_url"] = updateDto.AvatarUrl
	}
	if updateDto.Password != "" {
		updateData["password"] = updateDto.Password
	}

	if len(updateData) == 0 {
		return nil // không thay đổi gì hết
	}

	updateData["updated_at"] = time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameUser)
	_, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{"$set": updateData})
	if err != nil {
		return nil // update thất bại
	}

	return ur.GetUserById(id) // update thành công
}

// UpdateUserRole implements [IUserRepo]
func (ur *userRepo) UpdateUserRole(id string, roleId string) *schema.DbUser {
	objID := utils.ObjectIDFromHex(id)
	if objID == primitive.NilObjectID {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameUser)

	// update user role
	updateData := bson.M{
		"role":       roleId,
		"updated_at": time.Now(),
	}
	_, err := collection.UpdateOne(ctx,
		bson.M{"_id": objID},
		bson.M{"$set": updateData},
	)
	if err != nil {
		return nil // update thất bại
	}
	return ur.GetUserById(id) // update thành công
}

// DeleteUser implements [IUserRepo].
func (ur *userRepo) DeleteUser(id string) bool {
	objID := utils.ObjectIDFromHex(id)
	if objID == primitive.NilObjectID {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameUser)
	// soft delete, chỉ vô hiệu hoá tài khoản
	_, err := collection.UpdateOne(ctx, bson.M{"_id": objID}, bson.M{
		"$set": bson.M{
			"is_active":  false,
			"updated_at": time.Now(),
		},
	})
	return err == nil
}

func (ur *userRepo) GetUsers(limit int, index primitive.ObjectID) *[]schema.DbUser {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameUser)
	filter := bson.M{}
	if index != primitive.NilObjectID {
		filter["_id"] = bson.M{"$gt": index}
	}
	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{
			{Key: "_id", Value: 1},
		})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	var users []schema.DbUser
	if err := cursor.All(ctx, &users); err != nil {
		return nil
	}
	return &users
}

func (ur *userRepo) GetOverallStats() *dto.OverallStats {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	weekAgo := time.Now().AddDate(0, 0, -7)

	collection := global.Mgo.Database.Collection(schema.CollectionNameUser)
	findOptions := options.Find().
		SetSort(bson.D{
			{Key: "created_at", Value: -1},
		}).
		SetLimit(10)

	findCursor, err := collection.Find(ctx, bson.D{}, findOptions)
	if err != nil {
		return nil
	}
	defer findCursor.Close(ctx)

	// get last registrations
	var lastRegistrations []schema.DbUser
	if err := findCursor.All(ctx, &lastRegistrations); err != nil {
		return nil
	}

	pipeline := mongo.Pipeline{
		{{Key: "$facet", Value: bson.M{
			"total":    []bson.M{{"$count": "count"}},
			"active":   []bson.M{{"$match": bson.M{"is_active": true}}, {"$count": "count"}},
			"banned":   []bson.M{{"$match": bson.M{"is_active": false}}, {"$count": "count"}},
			"new_week": []bson.M{{"$match": bson.M{"created_at": bson.M{"$gte": weekAgo}}}, {"$count": "count"}},
		}}},
	}

	aggCursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil
	}
	defer aggCursor.Close(ctx)

	var result []bson.M
	if err := aggCursor.All(ctx, &result); err != nil {
		return nil
	}
	if len(result) == 0 {
		return nil
	}

	data := result[0]
	return &dto.OverallStats{
		TotalUsers:        utils.ExtractCount(data, "total"),
		ActiveUsers:       utils.ExtractCount(data, "active"),
		BannedUsers:       utils.ExtractCount(data, "banned"),
		NewThisWeek:       utils.ExtractCount(data, "new_week"),
		LastRegistrations: lastRegistrations,
	}
}

func NewUserRepo() IUserRepo {
	return &userRepo{}
}
