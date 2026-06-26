package role

import (
	"context"
	"go-app/global"
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"time"

	"github.com/patrickmn/go-cache"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IRoleRepo interface {
	CreateRole(createDto dto.CreateRoleDto) *schema.DbRole
	GetAllRole() *map[string]dto.RoleDto
	GetRoleById(id primitive.ObjectID) *schema.DbRole
}

type roleRepo struct{}

// CreateRole implements [IRoleRepo].
func (r *roleRepo) CreateRole(createDto dto.CreateRoleDto) *schema.DbRole {
	role := schema.DbRole{
		ID:        primitive.NewObjectID(),
		RoleName:  createDto.RoleName,
		RoleNote:  createDto.RoleNote,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameRole)
	_, err := collection.InsertOne(ctx, &role)
	if err != nil {
		return nil
	}

	// Invalidate cache để lần sau fetch lại từ DB
	global.Cache.Delete("roles")

	return &role
}

// GetAllRole implements [IRoleRepo].
func (r *roleRepo) GetAllRole() *map[string]dto.RoleDto {
	var roles []schema.DbRole

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameRole)
	cursor, err := collection.Find(ctx, bson.M{}) // lấy tất cả
	if err != nil {
		return nil
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &roles); err != nil {
		return nil
	}

	// cache role in map
	rolesMap := make(map[string]dto.RoleDto)
	for _, role := range roles {
		dto := dto.RoleDto{
			RoleId:   role.ID.Hex(),
			RoleName: role.RoleName,
		}
		rolesMap[dto.RoleName] = dto
	}

	global.Cache.Add("roles", rolesMap, cache.DefaultExpiration)

	return &rolesMap
}

// GetRoleById implements [IRoleRepo].
// Ưu tiên đọc từ cache trước, nếu không có thì query DB.
func (r *roleRepo) GetRoleById(id primitive.ObjectID) *schema.DbRole {
	// Thử tìm trong cache trước
	if cached, ok := global.Cache.Get("roles"); ok {
		if rolesMap, ok := cached.(map[string]interface{}); ok {
			for _, v := range rolesMap {
				if dbRole, ok := v.(schema.DbRole); ok && dbRole.ID == id {
					return &dbRole
				}
			}
		}
	}

	// Cache miss → query trực tiếp DB
	var role schema.DbRole
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	collection := global.Mgo.Database.Collection(schema.CollectionNameRole)
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&role)
	if err != nil {
		return nil
	}
	return &role
}

func NewRoleRepo() IRoleRepo {
	return &roleRepo{}
}
