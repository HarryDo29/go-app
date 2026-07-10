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

// Login godoc
// @Summary      User login
// @Description  Authenticate user and return tokens
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        req body appDto.LoginDto true "Login Info"
// @Success      200 {object} map[string]interface{}
// @Router       /auth/login [post]
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

// Register godoc
// @Summary      User registration
// @Description  Register a new user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        req body appDto.RegisterDto true "Register Info"
// @Success      200 {object} map[string]interface{}
// @Router       /auth/register [post]
func (ac *AuthController) Register(c *gin.Context) {
	var req appDto.RegisterDto
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(c, response.ErrCodeBodyInvalid)
		return
	}
	result := ac.authService.Register(&req)
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

// Logout godoc
// @Summary      User logout
// @Description  Invalidate user session/token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        req body appDto.Logout true "Logout Info"
// @Success      200 {object} map[string]interface{}
// @Router       /auth/logout [post]
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

// RefreshToken godoc
// @Summary      Refresh token
// @Description  Get new access token using refresh token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        req body appDto.RefreshTokenDto true "Refresh Token Info"
// @Success      200 {object} map[string]interface{}
// @Router       /auth/refresh [post]
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

// ChangePassword godoc
// @Summary      Change password
// @Description  Change user password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        req body appDto.ChangePasswordDto true "Change Password Info"
// @Success      200 {object} map[string]interface{}
// @Router       /auth/change-password [post]
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

// ForgetPassword godoc
// @Summary      Forget password
// @Description  Send OTP to email for password reset
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        req body appDto.ForgetPasswordDto true "Forget Password Info"
// @Success      200 {object} map[string]interface{}
// @Router       /auth/forget [post]
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

// VerifyForgetPassword godoc
// @Summary      Verify forget password OTP
// @Description  Verify OTP and return reset token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        req body appDto.VerifyOtpResetDto true "Verify OTP Info"
// @Success      200 {object} map[string]interface{}
// @Router       /auth/forget/verify [post]
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

// ResetPassword godoc
// @Summary      Reset password
// @Description  Reset password using reset token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        req body appDto.ResetPasswordDto true "Reset Password Info"
// @Success      200 {object} map[string]interface{}
// @Router       /auth/reset [post]
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

// SendOtp godoc
// @Summary      Send OTP
// @Description  Send OTP to user email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Success      200 {object} map[string]interface{}
// @Router       /auth/otp [post]
func (ac *AuthController) SendOtp(c *gin.Context) {
	email := c.GetString("email")
	if result := ac.authService.SendOtp(email); result {
		response.SuccessResponse(c, response.ErrCodeSuccess, map[string]string{"message": "OTP sent successfully"})
	} else {
		response.ErrorResponse(c, response.ErrCodeUpdateFailed)
	}
}

// ValidateOtp godoc
// @Summary      Validate OTP
// @Description  Validate OTP sent to user email
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        req body appDto.VerifyOtpDto true "Verify OTP Info"
// @Success      200 {object} map[string]interface{}
// @Router       /auth/otp/verify [post]
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
