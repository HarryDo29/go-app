package user

import (
	"go-app/global"
	dto "go-app/internal/dto"
	"go-app/internal/role"
)

// INTERFACE
type IUserService interface {
	CreateUser(dto dto.UserDto) *dto.UserResponseDto
	GetUser(email string) *dto.UserResponseDto
	GetUserById(id string) *dto.UserResponseDto
	UpdateUser(id string, dto dto.UpdateUserDto) *dto.UserResponseDto
	DeleteUser(id string) bool
}

type userService struct {
	userRepo IUserRepo
	roleRepo role.IRoleRepo
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

func NewUserService(userRepo IUserRepo, roleRepo role.IRoleRepo) IUserService {
	return &userService{
		userRepo: userRepo,
		roleRepo: roleRepo,
	}
}
