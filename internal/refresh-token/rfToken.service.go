package refreshtoken

import (
	"errors"
	"fmt"
	"go-app/global"
	dto "go-app/internal/dto"
	"go-app/internal/schema"
	"go-app/pkg/utils"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

var accTokenOption utils.SecretKey
var rfTokenOption utils.SecretKey

type IRefreshTokenService interface {
	CreateAccessToken(createDto dto.CreateTokenDto) dto.AccTokenResponseDto
	CreateRefreshToken(createDto dto.CreateTokenDto) dto.TokenResponseDto
	VerifyAccessToken(accessToken string) (*utils.MyCustomClaims, error)
	VerifyRefreshToken(refreshToken string) (*utils.MyCustomClaims, error)
	GetRefreshTokens(userId string) []schema.DbRefreshToken
	RemoveRefreshToken(userId string, rfToken string) bool
}

type refreshTokenService struct {
	rfTokenRepo IRefreshTokenRepo
}

func (rfs *refreshTokenService) CreateAccessToken(createDto dto.CreateTokenDto) dto.AccTokenResponseDto {
	user := utils.UserInfo{
		UserId: createDto.UserId,
		Email:  createDto.UserEmail,
		Role:   createDto.Role,
	}
	accToken, acErr := utils.GenerateJWT(user, accTokenOption)

	if acErr != nil {
		return dto.AccTokenResponseDto{}
	}
	return dto.AccTokenResponseDto{
		UserId:   createDto.UserId,
		AccToken: accToken,
	}
}

// CreateRefreshToken implements [IRefreshTokenService].
func (rfs *refreshTokenService) CreateRefreshToken(createDto dto.CreateTokenDto) dto.TokenResponseDto {
	user := utils.UserInfo{
		UserId: createDto.UserId,
		Email:  createDto.UserEmail,
		Role:   createDto.Role,
	}
	rfToken, rfErr := utils.GenerateJWT(user, rfTokenOption)
	accToken, acErr := utils.GenerateJWT(user, accTokenOption)

	if rfErr != nil || acErr != nil {
		return dto.TokenResponseDto{}
	}

	objUserId := utils.ObjectIDFromHex(createDto.UserId)
	if objUserId == primitive.NilObjectID {
		return dto.TokenResponseDto{}
	}

	rfDto := dto.CreateFreshTokenDto{
		Token:  rfToken,
		UserId: objUserId,
	}

	if !rfs.rfTokenRepo.CreateRefreshToken(rfDto) {
		return dto.TokenResponseDto{}
	}

	return dto.TokenResponseDto{
		UserId:   createDto.UserId,
		AccToken: accToken,
		RfToken:  rfToken,
	}
}

// VerifyAccessToken checks signature and expiry for the access token.
func (r *refreshTokenService) VerifyAccessToken(accessToken string) (*utils.MyCustomClaims, error) {
	token := normalizeToken(accessToken)
	if token == "" {
		return nil, errors.New("access token is empty")
	}

	return utils.VerifyJWT(token, accTokenOption)
}

// VerifyRefreshToken checks signature, expiry, and that the token still exists in storage.
func (r *refreshTokenService) VerifyRefreshToken(refreshToken string) (*utils.MyCustomClaims, error) {
	fmt.Println("normalize-refresh-token")
	token := normalizeToken(refreshToken)
	if token == "" {
		return nil, errors.New("refresh token is empty")
	}

	fmt.Println("verify-refresh-token-2")
	claims, err := utils.VerifyJWT(token, rfTokenOption)
	if err != nil {
		return nil, err
	}

	fmt.Println("check-refresh-token-in-db")
	storedTokens := r.rfTokenRepo.GetRefreshTokens(claims.UserId)

	// Kiểm tra xem token gửi lên có nằm trong danh sách các token hợp lệ của user trong DB không
	isValid := false
	for _, storedToken := range storedTokens {
		if storedToken.Token == token {
			isValid = true
			break
		}
	}

	if !isValid {
		return nil, errors.New("refresh token is invalid or has been revoked")
	}
	return claims, nil
}

// GetRefreshTokens implements [IRefreshTokenService].
func (r *refreshTokenService) GetRefreshTokens(userId string) []schema.DbRefreshToken {
	return r.rfTokenRepo.GetRefreshTokens(userId)
}

// RemoveRefreshToken implements [IRefreshTokenService].
func (r *refreshTokenService) RemoveRefreshToken(userId string, rfToken string) bool {
	return r.rfTokenRepo.RemoveRefreshToken(userId, rfToken)
}

func NewRefreshTokenService(rfTokenRepo IRefreshTokenRepo) IRefreshTokenService {
	// Khởi tạo ở đây: global.Config đã được load trước khi Wire gọi constructor này
	accDuration, err := time.ParseDuration(global.Config.Security.JWT.AccessTokenExpiration)
	if err != nil {
		panic("invalid access token expiration: " + err.Error())
	}
	rfDuration, err := time.ParseDuration(global.Config.Security.JWT.RefreshTokenExpiration)
	if err != nil {
		panic("invalid refresh token expiration: " + err.Error())
	}
	accTokenOption = utils.SecretKey{
		Key:        global.Config.Security.JWT.AccessTokenSecret,
		ExpireTime: int(accDuration.Seconds()),
	}
	rfTokenOption = utils.SecretKey{
		Key:        global.Config.Security.JWT.RefreshTokenSecret,
		ExpireTime: int(rfDuration.Seconds()),
	}
	global.Cache.Set("access_option", accTokenOption, -1)
	global.Cache.Set("refresh_option", rfTokenOption, -1)
	return &refreshTokenService{
		rfTokenRepo: rfTokenRepo,
	}
}

func normalizeToken(token string) string {
	trimmedToken := strings.TrimSpace(token)
	if strings.HasPrefix(strings.ToLower(trimmedToken), "bearer ") {
		return strings.TrimSpace(trimmedToken[len("Bearer "):])
	}
	return trimmedToken
}
