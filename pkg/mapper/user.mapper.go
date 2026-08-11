package mapper

import (
	dto "go-app/internal/dto"
	"go-app/internal/schema"
)

// ToUserResponseDto maps DbUser database schema model to UserResponseDto
func ToUserResponseDto(u *schema.DbUser) *dto.UserResponseDto {
	if u == nil {
		return nil
	}
	return &dto.UserResponseDto{
		UserId:    u.ID.Hex(),
		UserName:  u.UserName,
		Email:     u.Email,
		AvatarUrl: u.AvatarUrl,
		IsActive:  &u.IsActive,
		Role:      u.Role.Hex(),
	}
}

// ToUserResponseDtoList maps a slice of DbUser database schema models to a slice of UserResponseDto
func ToUserResponseDtoList(users []schema.DbUser) []dto.UserResponseDto {
	if users == nil {
		return []dto.UserResponseDto{}
	}
	result := make([]dto.UserResponseDto, len(users))
	for i, u := range users {
		if res := ToUserResponseDto(&u); res != nil {
			result[i] = *res
		}
	}
	return result
}

// ToUserSearchResponseDto maps a DbUser database schema model to UserSearchResponseDto with relation status
func ToUserSearchResponseDto(u *schema.DbUser, relationStatus string, connectionId string) dto.UserSearchResponseDto {
	var userRes dto.UserResponseDto
	if res := ToUserResponseDto(u); res != nil {
		userRes = *res
	}
	return dto.UserSearchResponseDto{
		UserResponseDto: userRes,
		RelationStatus: relationStatus,
		ConnectionId:   connectionId,
	}
}
