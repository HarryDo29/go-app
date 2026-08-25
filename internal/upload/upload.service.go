package upload

import (
	"context"
	"fmt"
	"math"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"go-app/global"
	"go-app/internal/dto"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var ctx = context.TODO()

type IUploadService interface {
	InitiateSingleUpload(opts dto.InitiateUploadReq) (dto.InitiateSingleUploadRes, error)
	InitiateUploadPart(req dto.InitiateUploadReq) (*dto.InitiateUploadPartRes, error)
	GetPresignedPartURLs(req dto.PresignPartReq) ([]dto.PartPresignURL, error)
	CompleteUpload(req dto.CompleteUploadReq) (dto.CompleteUploadRes, error)
	AbortUpload(req dto.AbortUpload) error
	GetDownloadURL(fileKey string) (string, error)
	GetObjectStream(fileKey string, rangeHeader *string) (*s3.GetObjectOutput, error)
}

type UploadService struct{}

func buildObjectKey(folder, objectName string) string {
	ext := filepath.Ext(objectName)                           // e.g. ".png"
	fileName := fmt.Sprintf("%s%s", uuid.New().String(), ext) // e.g. "9b1deb4d-3b7d-4bad-9bdd-2b0d7b3dcb6d.png"
	if folder != "" {
		// path.Clean and path.Join safely handle leading/trailing and duplicate slashes
		return path.Clean(path.Join(strings.Trim(folder, "/"), fileName))
	}
	return fileName
}

// Helper to calculate partSize and totalParts based on S3 constraints
func calculatePartInfo(fileSize int64) (partSize int64, totalParts int) {
	partSize = global.Config.S3.MinPartSize

	// S3 allows a maximum of 10,000 parts
	// If file is huge (e.g. > 50GB), increase partSize so totalParts <= 10,000
	const maxS3Parts = 10000
	if fileSize/partSize > maxS3Parts {
		partSize = int64(math.Ceil(float64(fileSize) / float64(maxS3Parts)))
	}

	totalParts = int(math.Ceil(float64(fileSize) / float64(partSize)))
	return partSize, totalParts
}

// save image under 5MB
func (s *UploadService) InitiateSingleUpload(opts dto.InitiateUploadReq) (dto.InitiateSingleUploadRes, error) {
	objectKey := buildObjectKey(opts.Folder, opts.ObjectName)

	input := &s3.PutObjectInput{
		Bucket:        aws.String(global.Config.S3.BucketName),
		Key:           aws.String(objectKey),
		ContentLength: aws.Int64(opts.Size),
	}
	if opts.ContentType != "" {
		input.ContentType = aws.String(opts.ContentType)
	}

	presignedReq, err := global.S3PresignClient.PresignPutObject(
		ctx,
		input,
		s3.WithPresignExpires(15*time.Minute),
	)
	if err != nil {
		return dto.InitiateSingleUploadRes{}, err
	}

	return dto.InitiateSingleUploadRes{
		ObjectKey:  objectKey,
		PresignURL: presignedReq.URL,
	}, nil
}

// init upload use presigned url, must
func (s *UploadService) InitiateUploadPart(req dto.InitiateUploadReq) (*dto.InitiateUploadPartRes, error) {
	objectKey := buildObjectKey(req.Folder, req.ObjectName)

	input := &s3.CreateMultipartUploadInput{
		Bucket:      aws.String(global.Config.S3.BucketName),
		Key:         aws.String(objectKey),
		ContentType: aws.String(req.ContentType),
	}

	output, err := global.S3Client.CreateMultipartUpload(ctx, input)
	if err != nil {
		return nil, err
	}

	partSize, totalParts := calculatePartInfo(req.Size)
	res := &dto.InitiateUploadPartRes{
		UploadId:   *output.UploadId,
		ObjectKey:  objectKey,
		PartSize:   partSize,
		TotalParts: totalParts,
	}
	return res, nil
}

// get presigned url for uploading each part
func (s *UploadService) GetPresignedPartURLs(req dto.PresignPartReq) ([]dto.PartPresignURL, error) {
	partURLs := make([]dto.PartPresignURL, 0, req.TotalParts)

	for i := 1; i <= req.TotalParts; i++ {
		input := &s3.UploadPartInput{
			Bucket:     aws.String(global.Config.S3.BucketName),
			Key:        aws.String(req.ObjectKey),
			UploadId:   aws.String(req.UploadId),
			PartNumber: aws.Int32(int32(i)),
		}

		presignedReq, err := global.S3PresignClient.PresignUploadPart(
			ctx, input,
			s3.WithPresignExpires(1*time.Hour),
		)
		if err != nil {
			return nil, fmt.Errorf("failed to presign upload part %d: %w", i, err)
		}

		partURLs = append(partURLs, dto.PartPresignURL{
			PartNumber: int32(i),
			PresignURL: presignedReq.URL,
		})
	}
	return partURLs, nil
}

// complete upload
func (s *UploadService) CompleteUpload(req dto.CompleteUploadReq) (dto.CompleteUploadRes, error) {
	// S3 strictly requires parts sorted by PartNumber in ascending order
	sort.Slice(req.Parts, func(i, j int) bool {
		return req.Parts[i].PartNumber < req.Parts[j].PartNumber
	})

	completedParts := make([]types.CompletedPart, len(req.Parts))
	for i, p := range req.Parts {
		completedParts[i] = types.CompletedPart{
			PartNumber: aws.Int32(p.PartNumber),
			ETag:       aws.String(p.ETag),
		}
	}

	input := &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(global.Config.S3.BucketName),
		Key:      aws.String(req.ObjectKey),
		UploadId: aws.String(req.UploadId),
		MultipartUpload: &types.CompletedMultipartUpload{
			Parts: completedParts,
		},
	}

	_, err := global.S3Client.CompleteMultipartUpload(ctx, input)
	if err != nil {
		return dto.CompleteUploadRes{}, fmt.Errorf("failed to complete multipart upload: %w", err)
	}

	downloadURL, err := s.GetDownloadURL(req.ObjectKey)
	if err != nil {
		return dto.CompleteUploadRes{}, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return dto.CompleteUploadRes{
		ObjectKey:   req.ObjectKey,
		DownloadURL: downloadURL,
	}, nil
}

// abort (destroy) Upload
func (s *UploadService) AbortUpload(req dto.AbortUpload) error {
	input := &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(global.Config.S3.BucketName),
		Key:      aws.String(req.ObjectKey),
		UploadId: aws.String(req.UploadId),
	}
	_, err := global.S3Client.AbortMultipartUpload(ctx, input)
	return err
}

// generate presigned url for downloading
func (s *UploadService) GetDownloadURL(fileKey string) (string, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(global.Config.S3.BucketName),
		Key:    aws.String(fileKey),
	}

	presignedReq, err := global.S3PresignClient.PresignGetObject(
		ctx, input,
		s3.WithPresignExpires(24*time.Hour))
	if err != nil {
		return "", fmt.Errorf("failed to generate download URL: %w", err)
	}

	return presignedReq.URL, nil
}

// stream object to client, use for download directly from S3
// streaming video/audio (voice) to client through server
func (s *UploadService) GetObjectStream(fileKey string, rangeHeader *string) (*s3.GetObjectOutput, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(global.Config.S3.BucketName),
		Key:    aws.String(fileKey),
		Range:  rangeHeader,
	}

	return global.S3Client.GetObject(ctx, input)
}

func NewUploadService() IUploadService {
	return &UploadService{}
}
