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

// CreateRefreshToken godoc
// @Summary      Create refresh token
// @Description  Create a new refresh token
// @Tags         refresh-token
// @Accept       json
// @Produce      json
// @Param        req body dto.CreateTokenDto true "Token Info"
// @Success      200 {object} map[string]interface{}
// @Router       /rf/refresh [post]
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

// func (rfc *RefreshTokenController) GetRefreshToken(c *gin.Context) {
// 	var dto dto.GetTokenHeaderDto
// 	if err := c.ShouldBindHeader(&dto); err != nil {
// 		response.ErrorResponse(c, response.ErrCodeParamInvalid)
// 		return
// 	}

// 	result := rfc.rfTokenService.GetRefreshTokens(dto.UserId)
// 	if result.ID.IsZero() {
// 		response.ErrorResponse(c, response.ErrCodeTokenNotFound) // token ko tồn tại
// 		return
// 	}
// 	response.SuccessResponse(c, response.ErrCodeSuccess, result)
// }
