package auth

import (
	"fmt"
	appDto "go-app/internal/dto"
	"go-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	authService IAuthService
}

func NewAuthController(authService IAuthService) *AuthController {
	return &AuthController{
		authService: authService,
	}
}

// GetResetUserId dùng làm callback cho ResetTokenMiddleware
func (ac *AuthController) GetResetUserId(token string) string {
	return ac.authService.GetResetUserId(token)
}

func (ac *AuthController) Login(c *gin.Context) {
	var req appDto.LoginDto
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	result := ac.authService.Login(&req)
	if result == (appDto.LoginResponseDto{}) {
		response.ErrorResponse(c, response.ErrCodeAuthFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

func (ac *AuthController) Register(c *gin.Context) {
	var req appDto.RegisterDto
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	result := ac.authService.Register(&req)
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

func (ac *AuthController) Logout(c *gin.Context) {
	userId := c.GetString("user-id")
	var req appDto.Logout
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	if result := ac.authService.Logout(userId, req); result {
		response.SuccessResponse(c, response.ErrCodeSuccess, map[string]string{"message": "Logged out successfully"})
	} else {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
	}
}

func (ac *AuthController) RefreshToken(c *gin.Context) {
	var req appDto.RefreshTokenDto
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	fmt.Println("verify-refresh-token")
	accToken := ac.authService.RefreshToken(req.RefreshToken)
	if accToken == "" {
		response.ErrorResponse(c, response.ErrCodeTokenInvalid)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, map[string]interface{}{
		"accessToken": accToken,
	})
}

func (ac *AuthController) ChangePassword(c *gin.Context) {
	userId := c.GetString("user-id")
	var req appDto.ChangePasswordDto
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	if result := ac.authService.ChangePassword(userId, &req); result {
		response.SuccessResponse(c, response.ErrCodeSuccess, map[string]string{"message": "Password changed successfully"})
	} else {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
	}
}

func (ac *AuthController) ForgetPassword(c *gin.Context) {
	var req appDto.ForgetPasswordDto
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	if result := ac.authService.ForgetPassword(req); result {
		response.SuccessResponse(c, response.ErrCodeSuccess, map[string]string{"message": "OTP sent to your email"})
	} else {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
	}
}

func (ac *AuthController) VerifyForgetPassword(c *gin.Context) {
	var req appDto.VerifyOtpResetDto
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	resetToken := ac.authService.VerifyOtpForReset(req)
	if resetToken == "" {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, map[string]string{"reset_token": resetToken})
}

func (ac *AuthController) ResetPassword(c *gin.Context) {
	userId := c.GetString("user-id")
	var req appDto.ResetPasswordDto
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	if result := ac.authService.ResetPassword(userId, &req); result {
		response.SuccessResponse(c, response.ErrCodeSuccess, map[string]string{"message": "Password reset successfully"})
	} else {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
	}
}

func (ac *AuthController) SendOtp(c *gin.Context) {
	email := c.GetString("email")
	if result := ac.authService.SendOtp(email); result {
		response.SuccessResponse(c, response.ErrCodeSuccess, map[string]string{"message": "OTP sent successfully"})
	} else {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
	}
}

func (ac *AuthController) ValidateOtp(c *gin.Context) {
	userId := c.GetString("user-id")
	email := c.GetString("email")
	var req appDto.VerifyOtpDto
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	if result := ac.authService.ValidateOtp(userId, email, req); result {
		response.SuccessResponse(c, response.ErrCodeSuccess, map[string]string{"message": "OTP verified successfully"})
	} else {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
	}
}
