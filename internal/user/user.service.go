package user

import (
	"go-app/global"
	dto "go-app/internal/dto"
	"go-app/internal/schema"
)

// INTERFACE
type IRoleRepo interface {
	GetAllRole() *map[string]dto.RoleDto
}

type IConnectionService interface {
	GetConnectionByUserId(userId string, status string) *[]schema.DbConnection
}

type IUserService interface {
	CreateUser(dto dto.UserDto) *dto.UserResponseDto
	GetUser(email string) *dto.UserResponseDto
	GetUserById(id string) *dto.UserResponseDto
	UpdateUser(id string, dto dto.UpdateUserDto) *dto.UserResponseDto
	DeleteUser(id string) bool
	SearchUsers(keyword string, userId string) *[]dto.UserSearchResponseDto
}

type userService struct {
	userRepo          IUserRepo
	roleRepo          IRoleRepo
	connectionService IConnectionService
}

// CreateUser implements [IUserService].
func (us *userService) CreateUser(userDto dto.UserDto) *dto.UserResponseDto {
	if cache, err := global.Cache.Get("roles"); err == false || cache == nil {
		us.roleRepo.GetAllRole() // call GetAllRole => return map in cache
	}

	result := us.userRepo.CreateUser(userDto)
	if result == nil {
		return nil
	}

	userRes := dto.UserResponseDto{
		UserId:    result.ID.Hex(),
		UserName:  result.UserName,
		Email:     result.Email,
		AvatarUrl: result.AvatarUrl,
		IsActive:  &result.IsActive,
		Role:      result.Role.Hex(),
	}
	return &userRes
}

// GetUser implements [IUserService].
func (us *userService) GetUser(email string) *dto.UserResponseDto {
	result := us.userRepo.GetUser(email)
	if result == nil {
		return nil
	}

	userRes := dto.UserResponseDto{
		UserId:    result.ID.Hex(),
		UserName:  result.UserName,
		Email:     result.Email,
		AvatarUrl: result.AvatarUrl,
		IsActive:  &result.IsActive,
		Role:      result.Role.Hex(),
	}
	return &userRes
}

// GetUserById implements [IUserService].
func (us *userService) GetUserById(id string) *dto.UserResponseDto {
	result := us.userRepo.GetUserById(id)
	if result == nil {
		return nil
	}

	userRes := dto.UserResponseDto{
		UserId:    result.ID.Hex(),
		UserName:  result.UserName,
		Email:     result.Email,
		AvatarUrl: result.AvatarUrl,
		IsActive:  &result.IsActive,
		Role:      result.Role.Hex(),
	}
	return &userRes
}

// UpdateUser implements [IUserService].
func (us *userService) UpdateUser(id string, updateDto dto.UpdateUserDto) *dto.UserResponseDto {
	result := us.userRepo.UpdateUser(id, updateDto)
	if result == nil {
		return nil
	}

	userRes := dto.UserResponseDto{
		UserId:    result.ID.Hex(),
		UserName:  result.UserName,
		Email:     result.Email,
		AvatarUrl: result.AvatarUrl,
		IsActive:  &result.IsActive,
		Role:      result.Role.Hex(),
	}
	return &userRes
}

// DeleteUser implements [IUserService].
func (us *userService) DeleteUser(id string) bool {
	return us.userRepo.DeleteUser(id)
}

// SearchUsers implements [IUserService].
func (us *userService) SearchUsers(keyword string, userId string) *[]dto.UserSearchResponseDto {
	// 1. Fetch users from repo (limit 10 for search)
	users := us.userRepo.SearchUsers(keyword, userId, 10)
	if users == nil {
		return nil
	}

	// 2. Fetch connections of current user
	connections := us.connectionService.GetConnectionByUserId(userId, "ACCEPTED")

	// Create a map for quick lookup of relation status
	relationMap := make(map[string]struct {
		Status string
		ID     string
	})

	if connections != nil {
		for _, conn := range *connections {
			// Determine the other user's ID in this connection
			otherId := ""
			for _, id := range conn.ParticipantIDs {
				if id.Hex() != userId {
					otherId = id.Hex()
					break
				}
			}
			relationMap[otherId] = struct {
				Status string
				ID     string
			}{
				Status: string(conn.Status),
				ID:     conn.ID.Hex(),
			}
		}
	}

	// 3. Map users to UserSearchResponseDto
	var result []dto.UserSearchResponseDto
	for _, u := range *users {
		status := "NONE"
		connId := ""
		if relation, exists := relationMap[u.ID.Hex()]; exists {
			status = relation.Status
			connId = relation.ID
		}

		result = append(result, dto.UserSearchResponseDto{
			UserResponseDto: dto.UserResponseDto{
				UserId:    u.ID.Hex(),
				UserName:  u.UserName,
				Email:     u.Email,
				AvatarUrl: u.AvatarUrl,
				IsActive:  &u.IsActive,
				Role:      u.Role.Hex(),
			},
			RelationStatus: status,
			ConnectionId:   connId,
		})
	}

	return &result
}

func NewUserService(
	userRepo IUserRepo,
	role IRoleRepo,
	connectionService IConnectionService,
) IUserService {
	return &userService{
		userRepo:          userRepo,
		roleRepo:          role,
		connectionService: connectionService,
	}
}
