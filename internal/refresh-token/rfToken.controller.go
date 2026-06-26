package refreshtoken

import (
	dto "go-app/internal/dto"
	"go-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type RefreshTokenController struct {
	rfTokenService IRefreshTokenService
}

func NewRefreshTokenController(rfTokenService IRefreshTokenService) *RefreshTokenController {
	return &RefreshTokenController{
		rfTokenService: rfTokenService,
	}
}

func (rfc *RefreshTokenController) CreateRefreshToken(c *gin.Context) {
	var dto dto.CreateTokenDto
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}
	result := rfc.rfTokenService.CreateRefreshToken(dto)
	if result.UserId == "" || result.AccToken == "" || result.RfToken == "" {
		response.ErrorResponse(c, response.ErrCodeServer) // ko tao thanh cong
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}

func (rfc *RefreshTokenController) GetRefreshToken(c *gin.Context) {
	var dto dto.GetTokenHeaderDto
	if err := c.ShouldBindHeader(&dto); err != nil {
		response.ErrorResponse(c, response.ErrCodeParamInvalid)
		return
	}

	result := rfc.rfTokenService.GetRefreshToken(dto.UserId)
	if result.ID.IsZero() {
		response.ErrorResponse(c, response.ErrCodeTokenNotFound) // token ko tồn tại
		return
	}
	response.SuccessResponse(c, response.ErrCodeSuccess, result)
}
