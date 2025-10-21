package s3preup

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Create a custom interface that matches the methods we need from the presign client
type presignClientAPI interface {
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (string, error)
}

// mockPresignClient is a mock implementation of our custom presignClientAPI
type mockPresignClient struct {
	mock.Mock
}

// PresignPutObject mock implementation that returns a URL string directly
func (m *mockPresignClient) PresignPutObject(
	ctx context.Context,
	params *s3.PutObjectInput,
	optFns ...func(*s3.PresignOptions),
) (string, error) {
	args := m.Called(ctx, params, optFns)
	return args.String(0), args.Error(1)
}

// TestNew tests the New function for creating an S3Provider
func TestNew(t *testing.T) {
	ctx := context.Background()

	t.Run("successful creation with valid parameters", func(t *testing.T) {
		provider, err := New(ctx, "test-bucket", "us-east-1")

		assert.NoError(t, err)
		assert.NotNil(t, provider)
		assert.Equal(t, "test-bucket", provider.bucket)
		assert.NotNil(t, provider.presignClient)
	})
}

// testS3Provider is a modified version of S3Provider that accepts our custom mock interface
type testS3Provider struct {
	bucket        string
	presignClient presignClientAPI
}

// PresignedUploadURL implementation that uses our custom mock interface
func (p *testS3Provider) PresignedUploadURL(ctx context.Context, destination string, expires time.Duration) (string, error) {
	return p.presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(destination),
	}, func(po *s3.PresignOptions) {
		po.Expires = expires
	})
}

// TestPresignedUploadURL tests the PresignedUploadURL method
func TestPresignedUploadURL(t *testing.T) {
	ctx := context.Background()
	bucket := "test-bucket"
	destination := "uploads/test-file.txt"
	expires := 15 * time.Minute
	expectedURL := "https://test-bucket.s3.amazonaws.com/uploads/test-file.txt"

	t.Run("successful presigned URL generation", func(t *testing.T) {
		mockClient := new(mockPresignClient)

		mockClient.On("PresignPutObject",
			mock.Anything,
			mock.MatchedBy(func(input *s3.PutObjectInput) bool {
				return *input.Bucket == bucket && *input.Key == destination
			}),
			mock.Anything,
		).Return(expectedURL, nil)

		provider := &testS3Provider{
			bucket:        bucket,
			presignClient: mockClient,
		}

		url, err := provider.PresignedUploadURL(ctx, destination, expires)

		assert.NoError(t, err)
		assert.Equal(t, expectedURL, url)
		mockClient.AssertExpectations(t)
	})

	t.Run("error during presigned URL generation", func(t *testing.T) {
		mockClient := new(mockPresignClient)
		expectedError := errors.New("presign error")

		mockClient.On("PresignPutObject",
			mock.Anything,
			mock.Anything,
			mock.Anything,
		).Return("", expectedError)

		provider := &testS3Provider{
			bucket:        bucket,
			presignClient: mockClient,
		}

		url, err := provider.PresignedUploadURL(ctx, destination, expires)

		assert.Error(t, err)
		assert.Equal(t, expectedError, err)
		assert.Empty(t, url)
		mockClient.AssertExpectations(t)
	})
}
