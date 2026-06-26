package dto

import "go.mongodb.org/mongo-driver/bson/primitive"

type CreateTokenDto struct {
	UserId    string
	UserEmail string
	Role      string
}

type CreateFreshTokenDto struct {
	UserId primitive.ObjectID
	Token  string
}

type AccTokenResponseDto struct {
	UserId   string
	AccToken string
}

type TokenResponseDto struct {
	UserId   string
	AccToken string
	RfToken  string
}

type GetTokenHeaderDto struct {
	UserId string `header:"User-Id" binding:"required"`
}

type RefreshTokenDto struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

type LoginDto struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponseDto struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	User         UserResponseDto `json:"user"`
}

type Logout struct {
	RefreshToken string `json:"refresh_token"`
}

type RegisterDto struct {
	Email    string `json:"email"`
	UserName string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponseDto struct {
	AccessToken  string          `json:"access_token"`
	RefreshToken string          `json:"refresh_token"`
	User         UserResponseDto `json:"user"`
}

type SendOtpDto struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOtpDto struct {
	Otp string `json:"otp" binding:"required,len=6"`
}

type ChangePasswordDto struct {
	OldPassword  string `json:"old_password"`
	NewPassword  string `json:"new_password"`
	ConfPassword string `json:"conf_password"`
}

type ForgetPasswordDto struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOtpResetDto struct {
	Email string `json:"email" binding:"required,email"`
	Otp   string `json:"otp" binding:"required,len=6"`
}

type ResetPasswordDto struct {
	NewPassword  string `json:"new_password" binding:"required,min=6"`
	ConfPassword string `json:"confirm_password" binding:"required,eqfield=NewPassword"`
}
