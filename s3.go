package s3preup

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadProvider contract for generating a presigned URL for uploads.
type UploadProvider interface {
	PresignedUploadURL(ctx context.Context, destination string, expires time.Duration) (string, error)
}

// S3Provider implements the UploadProvider interface for AWS S3.
// It holds a pre-configured S3 presign client for efficient reuse.
type S3Provider struct {
	bucket        string
	presignClient *s3.PresignClient
}

// New creates and configures a new S3Provider.
// It initializes the AWS SDK configuration and S3 clients, returning an error if setup fails.
// The provided context is used for the initial AWS config loading.
func New(ctx context.Context, bucket string, region string) (*S3Provider, error) {
	cfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(region))

	if err != nil {
		return nil, err
	}

	s3Client := s3.NewFromConfig(cfg)
	presignClient := s3.NewPresignClient(s3Client)

	return &S3Provider{
		bucket:        bucket,
		presignClient: presignClient,
	}, nil
}

// PresignedUploadURL generates a temporary, secure URL that can be used to PUT an object into S3.
// The destination is the full object key (e.g., "uploads/my-file.zip").
// The expiry duration specifies how long the URL will be valid for.
func (p *S3Provider) PresignedUploadURL(ctx context.Context, destination string, expires time.Duration) (string, error) {
	presignedPutRequest, err := p.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(destination),
	}, func(po *s3.PresignOptions) {
		po.Expires = expires
	})

	if err != nil {
		return "", err
	}

	return presignedPutRequest.URL, nil
}
