package upload

import (
	"context"
	"fmt"
	"go-app/global"
	"go-app/internal/dto"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
)

type IUploadService interface {
	GeneratePresignedURL(opts dto.GeneratePresignedURLReq) (string, error)
}

type UploadService struct{}

func (s *UploadService) GeneratePresignedURL(opts dto.GeneratePresignedURLReq) (string, error) {
	// Thời gian hết hạn của URL, ví dụ 15 phút
	expires := time.Second * 15 * 60

	// Các tuỳ chọn cho pre-signed URL (nếu bạn muốn gán Header cố định)
	// reqParams := make(map[string]string)
	// if opts.ContentType != "" {
	// 	reqParams["response-content-type"] = opts.ContentType
	// }

	bucketName := "go-chat-app"
	// Kiểm tra bucket đã tồn tại chưa, nếu chưa thì tạo
	exists, errBucketExists := global.Minio.BucketExists(context.Background(), bucketName)
	if errBucketExists == nil && !exists {
		err := global.Minio.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{})
		if err == nil {
			// Cấu hình policy cho phép đọc (GetObject) công khai (Public Read-Only)
			policy := fmt.Sprintf(`{
				"Version": "2012-10-17",
				"Statement": [
					{
						"Effect": "Allow",
						"Principal": {
							"AWS": ["*"]
						},
						"Action": [
							"s3:GetObject"
						],
						"Resource": [
							"arn:aws:s3:::%s/*"
						]
					}
				]
			}`, bucketName)
			_ = global.Minio.SetBucketPolicy(context.Background(), bucketName, policy)
		}
	}

	fullPath := opts.ObjectName
	if opts.Folder != "" {
		// Xóa dấu gạch chéo dư thừa nếu có và ghép lại
		fullPath = fmt.Sprintf("%s/%s", strings.Trim(opts.Folder, "/"), opts.ObjectName)
	}
	presignedURL, err := global.Minio.PresignedPutObject(
		context.Background(),
		bucketName,
		fullPath,
		expires,
	)

	if err != nil {
		return "", err
	}

	return presignedURL.String(), nil
}

func NewUploadService() IUploadService {
	return &UploadService{}
}
