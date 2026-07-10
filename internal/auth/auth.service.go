package auth

import (
	"fmt"
	"go-app/global"
	dto "go-app/internal/dto"
	rf "go-app/internal/refresh-token"
	roleRepo "go-app/internal/role"
	user "go-app/internal/user"
	"go-app/pkg/utils"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IAuthService interface {
	Login(dto *dto.LoginDto) dto.LoginResponseDto
	Register(dto *dto.RegisterDto) dto.RegisterResponseDto
	Logout(userId string, logoutDto dto.Logout) bool
	ChangePassword(userId string, dto *dto.ChangePasswordDto) bool
	ForgetPassword(forgetDto dto.ForgetPasswordDto) bool
	VerifyOtpForReset(verifyDto dto.VerifyOtpResetDto) string
	ResetPassword(userId string, dto *dto.ResetPasswordDto) bool
	RefreshToken(refreshToken string) string
	GetResetUserId(resetToken string) string // dùng cho ResetTokenMiddleware
	SendOtp(email string) bool
	ValidateOtp(userId string, email string, dto dto.VerifyOtpDto) bool
}

type authService struct {
	userRepo        user.IUserRepo
	rfService       rf.IRefreshTokenService
	roleRepo        roleRepo.IRoleRepo
	otpStore        sync.Map // key=userId, value=otpCode (dùng cho cả 2FA và forget-password)
	resetTokenStore sync.Map // key=resetToken, value=userId  (dùng cho forget-password flow)
}

// Login implements [IAuthService].
func (a *authService) Login(loginDto *dto.LoginDto) dto.LoginResponseDto {
	// 1. check user exist
	user := a.userRepo.GetUser(loginDto.Email)
	if user == nil {
		return dto.LoginResponseDto{}
	}
	// 2. check password
	if !utils.VerifyPassword(user.Password, loginDto.Password) {
		return dto.LoginResponseDto{}
	}
	//3. get role name
	var roleName string
	if user.Role != primitive.NilObjectID {
		if dbRole := a.roleRepo.GetRoleById(user.Role); dbRole != nil {
			roleName = dbRole.RoleName
		}
	}

	userIDStr := user.ID.Hex()

	createDto := dto.CreateTokenDto{
		UserId:    userIDStr,
		UserEmail: user.Email,
		Role:      roleName,
	}
	result := a.rfService.CreateRefreshToken(createDto)

	return dto.LoginResponseDto{
		AccessToken:  result.AccToken,
		RefreshToken: result.RfToken,
		User: dto.UserResponseDto{
			UserId:   userIDStr,
			UserName: user.UserName,
			Email:    user.Email,
			Role:     roleName,
			IsActive: &user.IsActive,
		},
	}
}

// Register implements [IAuthService].
func (a *authService) Register(registerDto *dto.RegisterDto) dto.RegisterResponseDto {
	if cache, err := global.Cache.Get("roles"); err == false || cache == nil {
		a.roleRepo.GetAllRole()
	}

	// 1. check user exist
	check := a.userRepo.GetUser(registerDto.Email)
	if check != nil {
		return dto.RegisterResponseDto{}
	}
	// 2. hash password
	hash, err := utils.HashPassword(registerDto.Password)
	if err != nil {
		return dto.RegisterResponseDto{}
	}
	// 3. create user
	nUser := dto.UserDto{
		UserName: registerDto.UserName,
		Email:    registerDto.Email,
		Password: hash,
		Role:     "user",
	}
	model := a.userRepo.CreateUser(nUser)

	userIDStr := model.ID.Hex()

	// 4. generate token
	createDto := dto.CreateTokenDto{
		UserId:    userIDStr,
		UserEmail: model.Email,
		Role:      "user",
	}
	result := a.rfService.CreateRefreshToken(createDto)

	return dto.RegisterResponseDto{
		AccessToken:  result.AccToken,
		RefreshToken: result.RfToken,
		User: dto.UserResponseDto{
			UserId:   userIDStr,
			UserName: model.UserName,
			Email:    model.Email,
			Role:     "user",
		},
	}
}

// Logout implements [IAuthService].
func (a *authService) Logout(userId string, logoutDto dto.Logout) bool {
	return a.rfService.RemoveRefreshToken(userId, logoutDto.RefreshToken)
}

// RefreshToken implements [IAuthService].
func (a *authService) RefreshToken(refreshToken string) string {
	// 1. Verify the refresh token
	fmt.Println("verify-refresh-token-1")
	claims, err := a.rfService.VerifyRefreshToken(refreshToken)
	if err != nil {
		fmt.Println("err: ", err)
		return ""
	}

	// 2. Generate new access token
	createDto := dto.CreateTokenDto{
		UserId:    claims.UserInfo.UserId,
		UserEmail: claims.UserInfo.Email,
		Role:      claims.UserInfo.Role,
	}
	result := a.rfService.CreateAccessToken(createDto)
	if result.AccToken == "" {
		return ""
	}
	return result.AccToken
}

// ChangePassword implements [IAuthService].
func (a *authService) ChangePassword(userId string, changeDto *dto.ChangePasswordDto) bool {
	// 1. Check if new password matches confirm password
	if changeDto.NewPassword != changeDto.ConfPassword {
		return false
	}

	// 2. Find user
	user := a.userRepo.GetUserById(userId)
	if user == nil {
		return false
	}

	// 3. Verify old password
	if !utils.VerifyPassword(user.Password, changeDto.OldPassword) {
		return false
	}

	// 4. Hash new password
	hash, err := utils.HashPassword(changeDto.NewPassword)
	if err != nil {
		return false
	}

	// 5. Update user password in repo
	updatedUser := a.userRepo.UpdateUser(user.ID.Hex(), dto.UpdateUserDto{
		Password: hash,
	})

	return updatedUser != nil
}

// ForgetPassword implements [IAuthService] — public, không cần JWT.
func (a *authService) ForgetPassword(forgetDto dto.ForgetPasswordDto) bool {
	user := a.userRepo.GetUser(forgetDto.Email)
	if user == nil {
		return false
	}

	otpCode, err := utils.GenerateUniqueOTP()
	if err != nil {
		return false
	}

	// Lưu OTP theo userId
	a.resetTokenStore.Store(user.ID.Hex(), otpCode)

	// TODO: gửi email thật — hiện tại log ra console
	fmt.Printf("\n[FORGET PASSWORD] OTP [%s] gửi đến email [%s]\n\n", otpCode, forgetDto.Email)
	return true
}

// VerifyOtpForReset implements [IAuthService] — validate OTP, trả về reset_token (JWT).
func (a *authService) VerifyOtpForReset(verifyDto dto.VerifyOtpResetDto) string {
	user := a.userRepo.GetUser(verifyDto.Email)
	if user == nil {
		return ""
	}

	value, ok := a.resetTokenStore.Load(user.ID.Hex())
	if !ok {
		return "" // OTP chưa gửi hoặc đã hết hạn
	}

	storedOtp, ok := value.(string)
	if !ok || storedOtp != verifyDto.Otp {
		return ""
	}

	// OTP hợp lệ — xóa OTP
	a.resetTokenStore.Delete(user.ID.Hex())

	// Lấy reset_option từ cache (đã khởi tạo trong NewRefreshTokenService)
	resetCache, ok := global.Cache.Get("reset_option")
	if !ok {
		return ""
	}
	resetOption := resetCache.(utils.SecretKey)
	if resetOption == (utils.SecretKey{}) {
		resetExpire, _ := time.ParseDuration(global.Config.Security.JWT.ResetPasswordExpiration)
		reset := utils.SecretKey{
			Key:        global.Config.Security.JWT.ResetPasswordSecret,
			ExpireTime: int(resetExpire.Seconds()),
		}
		global.Cache.Set("reset_option", reset, 10*time.Minute)
	}

	userInfo := utils.UserInfo{
		UserId: user.ID.Hex(),
		Email:  user.Email,
	}

	// Tạo JWT ngắn hạn làm reset_token — middleware sẽ verify bằng VerifyJWT
	resetToken, err := utils.GenerateJWT(userInfo, resetOption)
	if err != nil {
		return ""
	}

	return resetToken
}

// ResetPassword implements [IAuthService] — dùng reset_token từ Authorization header.
func (a *authService) ResetPassword(userId string, resetDto *dto.ResetPasswordDto) bool {
	// 1. Tìm user theo userId đã decode từ reset_token
	user := a.userRepo.GetUserById(userId)
	if user == nil {
		return false
	}

	// 2. Validate mật khẩu
	if resetDto.NewPassword != resetDto.ConfPassword {
		return false
	}

	// 3. Hash mật khẩu mới
	hash, err := utils.HashPassword(resetDto.NewPassword)
	if err != nil {
		return false
	}

	// 4. Cập nhật password
	updatedUser := a.userRepo.UpdateUser(userId, dto.UpdateUserDto{
		Password: hash,
	})
	return updatedUser != nil
}

// SendOtp implements [IAuthService].
func (a *authService) SendOtp(email string) bool {
	// check user exist (email lấy từ JWT context)
	user := a.userRepo.GetUser(email)
	if user == nil {
		return false
	}

	// 1. Generate securely a 6-digit OTP
	otpCode, err := utils.GenerateUniqueOTP()
	if err != nil {
		return false
	}

	// 2. Store in thread-safe memory store
	a.otpStore.Store(user.ID.Hex(), otpCode)

	// 3. Mock sending OTP by printing to logger/console
	fmt.Printf("\n[OTP SERVICE] ========> SUCCESS: Sent OTP [%s] to email [%s]\n\n", otpCode, email)

	return true
}

// ValidateOtp implements [IAuthService].
func (a *authService) ValidateOtp(userId string, email string, dto dto.VerifyOtpDto) bool {
	// 1. Lấy user theo email
	user := a.userRepo.GetUser(email)
	if user == nil {
		return false
	}

	if user.ID.Hex() != userId {
		return false
	}

	// 2. Lấy OTP đã lưu trong otpStore theo userId
	value, ok := a.otpStore.Load(userId)
	if !ok {
		return false // OTP chưa được gửi hoặc đã hết hạn
	}

	storedOtp, ok := value.(string)
	if !ok {
		return false
	}

	// 3. So sánh OTP
	if storedOtp != dto.Otp {
		return false
	}

	// 4. Xóa OTP sau khi xác thực thành công (mỗi OTP chỉ dùng 1 lần)
	a.otpStore.Delete(userId)
	return true
}

// GetResetUserId implements [IAuthService] — dùng cho ResetTokenMiddleware.
func (a *authService) GetResetUserId(resetToken string) string {
	value, ok := a.resetTokenStore.Load(resetToken)
	if !ok {
		return ""
	}
	userId, ok := value.(string)
	if !ok {
		return ""
	}
	return userId
}

func NewAuthService(userRepo user.IUserRepo, rfService rf.IRefreshTokenService, rRepo roleRepo.IRoleRepo) IAuthService {
	return &authService{
		userRepo:  userRepo,
		rfService: rfService,
		roleRepo:  rRepo,
	}
}
