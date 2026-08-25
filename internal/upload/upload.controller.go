package upload

import (
	"fmt"
	"go-app/global"
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
// @Success      200 {object} dto.InitiateSingleUploadRes
// @Router       /upload/presigned [post]
func (c *UploadController) GeneratePresignURL(ctx *gin.Context) {
	var opts dto.InitiateUploadReq
	if err := ctx.ShouldBindJSON(&opts); err != nil {
		response.ErrorResponse(ctx, response.ErrCodeBodyInvalid)
		return
	}

	// Verify file size for single upload (must be > 0 and <= MaxSingleSize)
	if opts.Size <= 0 ||
		(global.Config.S3.MaxSingleSize > 0 && opts.Size > global.Config.S3.MaxSingleSize) {
		response.ErrorResponse(ctx, response.ErrCodeBodyInvalid)
		return
	}

	res, err := c.uploadService.InitiateSingleUpload(opts)
	if err != nil {
		fmt.Println("err:", err.Error())
		response.ErrorResponse(ctx, response.ErrCodeServer)
		return
	}

	response.SuccessResponse(ctx, response.ErrCodeSuccess, res)
}

// InitiateUploadPart godoc
// @Summary      Initiate multipart upload
// @Description  Initiate multipart upload and get part configuration (part_size, total_parts)
// @Tags         upload
// @Accept       json
// @Produce      json
// @Param        req body dto.InitiateUploadReq true "Multipart Upload Init Info"
// @Success      200 {object} dto.InitiateUploadPartRes
// @Router       /upload/multipart/init [post]
func (c *UploadController) InitiateUploadPart(ctx *gin.Context) {
	var req dto.InitiateUploadReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, response.ErrCodeBodyInvalid)
		return
	}

	// Verify file size for multipart upload (must be > 0 and <= MaxFileSize)
	if req.Size <= 0 || (global.Config.S3.MaxFileSize > 0 && req.Size > global.Config.S3.MaxFileSize) {
		response.ErrorResponse(ctx, response.ErrCodeBodyInvalid)
		return
	}

	res, err := c.uploadService.InitiateUploadPart(req)
	if err != nil {
		fmt.Println("err:", err.Error())
		response.ErrorResponse(ctx, response.ErrCodeServer)
		return
	}

	response.SuccessResponse(ctx, response.ErrCodeSuccess, res)
}

// GetPresignedPartURLs godoc
// @Summary      Get presigned part URLs
// @Description  Get presigned URLs for each part in multipart upload
// @Tags         upload
// @Accept       json
// @Produce      json
// @Param        req body dto.PresignPartReq true "Presigned Part Info"
// @Success      200 {array} dto.PartPresignURL
// @Router       /upload/multipart/presign-parts [post]
func (c *UploadController) GetPresignedPartURLs(ctx *gin.Context) {
	var req dto.PresignPartReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, response.ErrCodeBodyInvalid)
		return
	}

	res, err := c.uploadService.GetPresignedPartURLs(req)
	if err != nil {
		fmt.Println("err:", err.Error())
		response.ErrorResponse(ctx, response.ErrCodeServer)
		return
	}

	response.SuccessResponse(ctx, response.ErrCodeSuccess, res)
}

// CompleteUploadPart godoc
// @Summary      Complete multipart upload
// @Description  Complete multipart upload and return a presigned download URL
// @Tags         upload
// @Accept       json
// @Produce      json
// @Param        req body dto.CompleteUploadReq true "Complete Upload Info"
// @Success      200 {object} dto.CompleteUploadRes
// @Router       /upload/multipart/complete [post]
func (c *UploadController) CompleteUploadPart(ctx *gin.Context) {
	var req dto.CompleteUploadReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, response.ErrCodeBodyInvalid)
		return
	}

	res, err := c.uploadService.CompleteUpload(req)
	if err != nil {
		fmt.Println("err:", err.Error())
		response.ErrorResponse(ctx, response.ErrCodeServer)
		return
	}

	response.SuccessResponse(ctx, response.ErrCodeSuccess, res)
}

// AbortUpload godoc
// @Summary      Abort multipart upload
// @Description  Abort and clean up an unfinished multipart upload
// @Tags         upload
// @Accept       json
// @Produce      json
// @Param        req body dto.AbortUpload true "Abort Upload Info"
// @Success      200 {object} map[string]interface{}
// @Router       /upload/multipart/abort [post]
func (c *UploadController) AbortUpload(ctx *gin.Context) {
	var req dto.AbortUpload
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, response.ErrCodeBodyInvalid)
		return
	}

	if err := c.uploadService.AbortUpload(req); err != nil {
		fmt.Println("err:", err.Error())
		response.ErrorResponse(ctx, response.ErrCodeServer)
		return
	}

	response.SuccessResponse(ctx, response.ErrCodeSuccess, nil)
}

// GetDownloadURL godoc
// @Summary      Get download URL
// @Description  Get a presigned download URL for an object key
// @Tags         upload
// @Produce      json
// @Param        key query string true "Object Key"
// @Success      200 {object} map[string]interface{}
// @Router       /upload/download [get]
func (c *UploadController) GetDownloadURL(ctx *gin.Context) {
	fileKey := ctx.Query("key")
	if fileKey == "" {
		response.ErrorResponse(ctx, response.ErrCodeParamInvalid)
		return
	}

	url, err := c.uploadService.GetDownloadURL(fileKey)
	if err != nil {
		fmt.Println("err:", err.Error())
		response.ErrorResponse(ctx, response.ErrCodeServer)
		return
	}

	response.SuccessResponse(ctx, response.ErrCodeSuccess, gin.H{
		"object_key":   fileKey,
		"download_url": url,
	})
}
