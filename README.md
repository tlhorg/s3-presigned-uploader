# S3 Presigned Uploader

A Go library for generating presigned URLs for secure file uploads to Amazon S3.

## Features

- Generate secure, time-limited presigned URLs for S3 uploads
- Built on AWS SDK v2 for Go
- Simple and clean interface
- Context-aware operations
- Configurable expiration times

## Installation

```bash
go get github.com/tlhorg/s3-presigned-uploader
```

## Usage

### Basic Setup

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"

    s3preup "github.com/tlhorg/s3-presigned-uploader"
)

func main() {
    ctx := context.Background()
    
    // Initialize the S3 provider
    provider, err := s3preup.New(ctx, "my-bucket", "us-west-2")
    if err != nil {
        log.Fatal(err)
    }
    
    // Generate a presigned URL valid for 15 minutes
    url, err := provider.PresignedUploadURL(ctx, "uploads/my-file.jpg", 15*time.Minute)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("Upload URL:", url)
}
```

### Using the Interface

The library provides an `UploadProvider` interface that can be used for dependency injection:

```go
func handleUpload(provider s3preup.UploadProvider) error {
    ctx := context.Background()
    
    url, err := provider.PresignedUploadURL(ctx, "documents/report.pdf", 30*time.Minute)
    if err != nil {
        return err
    }
    
    // Send URL to client for upload
    return sendURLToClient(url)
}
```

### Client-side Upload

Once you have the presigned URL, clients can upload files directly to S3 using a PUT request:

```bash
curl -X PUT \
  -H "Content-Type: application/octet-stream" \
  --data-binary @/path/to/file.jpg \
  "PRESIGNED_URL_HERE"
```

Or in JavaScript:

```javascript
async function uploadFile(presignedUrl, file) {
    const response = await fetch(presignedUrl, {
        method: 'PUT',
        body: file,
        headers: {
            'Content-Type': file.type
        }
    });
    
    if (response.ok) {
        console.log('Upload successful');
    } else {
        console.error('Upload failed');
    }
}
```

## API Reference

### Types

#### `UploadProvider` Interface

```go
type UploadProvider interface {
    PresignedUploadURL(ctx context.Context, destination string, expires time.Duration) (string, error)
}
```

#### `S3Provider` Struct

The main implementation of `UploadProvider` for AWS S3.

### Functions

#### `New(ctx context.Context, bucket string, region string) (*S3Provider, error)`

Creates a new S3Provider instance.

**Parameters:**
- `ctx`: Context for AWS configuration loading
- `bucket`: S3 bucket name
- `region`: AWS region (e.g., "us-west-2")

**Returns:**
- `*S3Provider`: Configured S3 provider instance
- `error`: Error if configuration fails

#### `PresignedUploadURL(ctx context.Context, destination string, expires time.Duration) (string, error)`

Generates a presigned URL for uploading to S3.

**Parameters:**
- `ctx`: Request context
- `destination`: S3 object key (e.g., "uploads/file.jpg")
- `expires`: Duration the URL will be valid

**Returns:**
- `string`: Presigned upload URL
- `error`: Error if URL generation fails

## Configuration

The library uses AWS SDK v2's default configuration loading, which supports:

- Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`, `AWS_REGION`)
- AWS credentials file (`~/.aws/credentials`)
- IAM roles (for EC2/ECS/Lambda)
- AWS SSO

## Required AWS Permissions

The AWS credentials used must have the following S3 permissions:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "s3:PutObject"
            ],
            "Resource": "arn:aws:s3:::your-bucket-name/*"
        }
    ]
}
```

## Examples

### Web Handler Example

```go
func uploadHandler(provider s3preup.UploadProvider) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        filename := r.URL.Query().Get("filename")
        if filename == "" {
            http.Error(w, "filename required", http.StatusBadRequest)
            return
        }
        
        destination := fmt.Sprintf("uploads/%s", filename)
        url, err := provider.PresignedUploadURL(r.Context(), destination, 15*time.Minute)
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        
        json.NewEncoder(w).Encode(map[string]string{"upload_url": url})
    }
}
```

## License

MIT License - see LICENSE file for details.

