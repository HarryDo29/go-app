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
)

// INTERFACE
type IUserRepo interface {
	CreateUser(userDto dto.UserDto) *schema.DbUser
	GetUser(email string) *schema.DbUser
	GetUserById(id string) *schema.DbUser
	UpdateUser(id string, updateDto dto.UpdateUserDto) *schema.DbUser
	DeleteUser(id string) bool
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

func NewUserRepo() IUserRepo {
	return &userRepo{}
}
