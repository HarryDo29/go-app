package upload

import (
	"go-app/internal/middleware"
	"go-app/internal/wire"

	"github.com/gin-gonic/gin"
)

type UploadRouter struct{}

func (r *UploadRouter) InitUploadRouter(Router *gin.RouterGroup) {
	uploadController, _ := wire.InitUploadRouterHandler()

	uploadRouter := Router.Group("/upload")
	uploadRouter.Use(middleware.AuthNMiddleware())
	{
		uploadRouter.POST("/presigned", uploadController.GeneratePresignURL)
		uploadRouter.POST("/multipart/init", uploadController.InitiateUploadPart)
		uploadRouter.POST("/multipart/presign-parts", uploadController.GetPresignedPartURLs)
		uploadRouter.POST("/multipart/complete", uploadController.CompleteUploadPart)
		uploadRouter.POST("/multipart/abort", uploadController.AbortUpload)
		uploadRouter.GET("/download", uploadController.GetDownloadURL)
	}
}
