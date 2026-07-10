package upload

import (
	"fmt"
	"go-app/internal/dto"
	"go-app/pkg/response"

	"github.com/gin-gonic/gin"
)

type UploadController struct {
	uploadService IUploadService
}

func NewUploadController(uploadService IUploadService) *UploadController {
	return &UploadController{
		uploadService: uploadService,
	}
}

// GeneratePresignedURL godoc
// @Summary      Generate presigned URL
// @Description  Generate a presigned URL for file upload
// @Tags         upload
// @Accept       json
// @Produce      json
// @Param        req body dto.GeneratePresignedURLReq true "Presigned URL Info"
// @Success      200 {object} map[string]interface{}
// @Router       /upload/presigned [post]
func (c *UploadController) GeneratePresignedURL(ctx *gin.Context) {
	var opts dto.GeneratePresignedURLReq
	if err := ctx.ShouldBindJSON(&opts); err != nil {
		response.ErrorResponse(ctx, response.ErrCodeBodyInvalid)
		return
	}

	url, err := c.uploadService.GeneratePresignedURL(opts)
	if err != nil {
		fmt.Println("err:", err.Error())
		response.ErrorResponse(ctx, response.ErrCodeServer)
		return
	}

	res := dto.GeneratePresignedURLRes{
		URL: url,
	}

	response.SuccessResponse(ctx, response.ErrCodeSuccess, res)
}
