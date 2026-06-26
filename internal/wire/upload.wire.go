//go:build wireinject

package wire

import (
	"go-app/internal/upload"

	"github.com/google/wire"
)

func InitUploadRouterHandler() (*upload.UploadController, error) {
	wire.Build(
		upload.NewUploadService,
		upload.NewUploadController,
	)
	return new(upload.UploadController), nil
}
