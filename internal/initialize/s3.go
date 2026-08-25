package initialize

import (
	"context"
	"errors"
	"log"

	"go-app/global"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"go.uber.org/zap"
)

func InitS3() {
	s3Cfg := global.Config.S3
	ctx := context.TODO()

	// set up region
	region := s3Cfg.Region
	if region == "" {
		region = "us-east-1"
	}
	optFns := []func(*awsConfig.LoadOptions) error{
		awsConfig.WithRegion(region),
	}

	// If static credentials are provided (MinIO or explicit AWS credentials)
	if s3Cfg.AccessKey != "" && s3Cfg.SecretKey != "" {
		optFns = append(optFns, awsConfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				s3Cfg.AccessKey,
				s3Cfg.SecretKey,
				"",
			),
		))
	}

	cfg, err := awsConfig.LoadDefaultConfig(ctx, optFns...)
	if err != nil {
		log.Fatalf("Unable to load S3 SDK config: %v", err)
	}

	// Create S3 client
	s3Client := s3.NewFromConfig(cfg, func(opt *s3.Options) {
		opt.UsePathStyle = s3Cfg.UsePathStyle
		if s3Cfg.Endpoint != "" {
			opt.BaseEndpoint = aws.String(s3Cfg.Endpoint)
		}
	})

	// Create Presign client
	presignClient := s3.NewPresignClient(s3Client)

	// Ensure bucket exists (especially useful for local MinIO)
	if s3Cfg.BucketName != "" {
		_, err := s3Client.HeadBucket(ctx, &s3.HeadBucketInput{
			Bucket: aws.String(s3Cfg.BucketName),
		})
		if err != nil {
			var notFound *types.NotFound
			var noSuchBucket *types.NoSuchBucket
			if errors.As(err, &notFound) || errors.As(err, &noSuchBucket) {
				_, createErr := s3Client.CreateBucket(ctx, &s3.CreateBucketInput{
					Bucket: aws.String(s3Cfg.BucketName),
				})
				if createErr != nil {
					global.Logger.Warn("Could not create S3 bucket automatically",
						zap.Error(createErr),
					)
				} else {
					global.Logger.Info("Created S3 bucket successfully",
						zap.String("bucket", s3Cfg.BucketName),
					)
				}
			}
		}
	}

	global.S3Client = s3Client
	global.S3PresignClient = presignClient
	global.Logger.Info("Initialize S3 client successfully!",
		zap.String("provider", s3Cfg.Provider),
		zap.String("endpoint", s3Cfg.Endpoint),
		zap.String("region", region),
		zap.String("bucket", s3Cfg.BucketName),
	)
}
