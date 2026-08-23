package admin

import (
	"go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/pkg/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IRoleRepo interface {
	GetRole(roleName string) *schema.DbRole
}

type IUserRepo interface {
	GetUserById(id string) *schema.DbUser
	GetUsers(limit int, index primitive.ObjectID) *[]schema.DbUser
	UpdateUserRole(id string, roleId string) *schema.DbUser
	UpdateUser(id string, updateDto dto.UpdateUserDto) *schema.DbUser
	SearchUsers(keyword string, userId string, limit int) *[]schema.DbUser
	GetOverallStats() *dto.OverallStats
}

type IAdminService interface {
	ListUsers(limit int, index string) *[]schema.DbUser
	GetUserDetail(id string) *schema.DbUser
	UpdateUserRole(id string, role string) *schema.DbUser
	UpdateUserStatus(id string) *schema.DbUser
	SearchUsers(keyword string, userId string, limit int) *[]schema.DbUser
	GetOverall() *dto.OverallStats
}

type adminService struct {
	userRepo IUserRepo
	roleRepo IRoleRepo
}

// list all users (indexing)
func (as *adminService) ListUsers(limit int, index string) *[]schema.DbUser {
	cursor, err := utils.DecodeCursor(index)
	if err != nil {
		return nil
	}
	users := as.userRepo.GetUsers(limit, cursor)
	return users
}

// get user's detail
func (ac *adminService) GetUserDetail(id string) *schema.DbUser {
	user := ac.userRepo.GetUserById(id)
	return user
}

// update user's role
func (ac *adminService) UpdateUserRole(id string, roleName string) *schema.DbUser {
	// get role
	role := ac.roleRepo.GetRole(roleName)
	if role == nil {
		return nil
	}

	// update role
	updateUser := ac.userRepo.UpdateUserRole(id, role.ID.Hex())
	return updateUser
}

// delete user (ban)
func (ac *adminService) UpdateUserStatus(id string) *schema.DbUser {
	user := ac.userRepo.GetUserById(id)
	if user != nil {
		return nil
	}

	updateUserDto := dto.UpdateUserDto{
		IsActive: !user.IsActive,
	}
	updateUser := ac.userRepo.UpdateUser(id, updateUserDto)
	return updateUser
}

// search users
func (ac *adminService) SearchUsers(keyword string, userId string, limit int) *[]schema.DbUser {
	if keyword == "" {
		return nil
	}
	//
	users := ac.userRepo.SearchUsers(keyword, userId, limit)
	return users
}

// total users (details)
func (ac *adminService) GetOverall() *dto.OverallStats {
	return ac.userRepo.GetOverallStats()
}

func NewAdminService(userRepo IUserRepo, roleRepo IRoleRepo) IAdminService {
	return &adminService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}
