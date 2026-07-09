package dto

type UserDto struct {
	UserName string `json:"user_name" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6"`
	Email    string `json:"email" binding:"required,email"`
	Role     string `json:"role" binding:"required,dive"`
}

type UpdateUserDto struct {
	UserName  string `json:"user_name"`
	AvatarUrl string `json:"avatar_url"`
	Password  string `json:"password"`
	IsActive  bool   `json:"is_active"`
}

type UserResponseDto struct {
	UserId    string `json:"user_id"`
	UserName  string `json:"user_name"`
	Email     string `json:"email"`
	AvatarUrl string `json:"avatar_url"`
	IsActive  *bool  `json:"is_active"`
	Role      string `json:"role"`
}

type UserSearchResponseDto struct {
	UserResponseDto
	RelationStatus string `json:"relation_status"` // NONE, PENDING, ACCEPTED, REJECTED
	ConnectionId   string `json:"connection_id,omitempty"`
}
