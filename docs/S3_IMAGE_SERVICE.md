# S3 Image Service Configuration

This document describes how to configure and use the S3-based image service for the Matching API.

## Overview

The image service provides:

- **Image Upload**: Multiple upload methods (multipart form, base64, presigned URLs)
- **Image Storage**: Secure S3 storage with proper organization
- **Image Retrieval**: Direct download and listing capabilities
- **Image Management**: Delete functionality with proper authorization
- **Fallback Support**: Continues to work without S3 configuration

## Environment Variables

### Required (for S3 functionality)

```bash
AWS_S3_BUCKET=your-bucket-name
```

### Optional

```bash
AWS_REGION=us-east-1                    # Default: us-east-1
AWS_S3_BASE_URL=https://cdn.example.com # Optional CDN URL
```

### AWS Credentials

The service uses the AWS SDK's default credential chain:

1. Environment variables (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`)
2. AWS credentials file (`~/.aws/credentials`)
3. IAM roles (for EC2 instances)

## S3 Bucket Setup

### 1. Create S3 Bucket

```bash
aws s3 mb s3://your-matching-api-images
```

### 2. Configure Bucket Policy (Optional - for public access)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "PublicReadGetObject",
      "Effect": "Allow",
      "Principal": "*",
      "Action": "s3:GetObject",
      "Resource": "arn:aws:s3:::your-matching-api-images/*"
    }
  ]
}
```

### 3. Enable CORS (for browser uploads)

```json
[
  {
    "AllowedHeaders": ["*"],
    "AllowedMethods": ["PUT", "POST", "DELETE"],
    "AllowedOrigins": ["https://yourdomain.com"],
    "ExposeHeaders": ["ETag"]
  }
]
```

## API Endpoints

### Upload Methods

#### 1. Multipart Form Upload

```http
POST /api/v1/images/upload
Content-Type: multipart/form-data
Authorization: Bearer <token>

Form fields:
- image: file (required)
- position: integer (optional, 1-9)
```

#### 2. Base64 Upload

```http
POST /api/v1/images/upload-base64
Content-Type: application/json
Authorization: Bearer <token>

{
  "image_data": "base64-encoded-image",
  "position": 1,
  "filename": "optional.jpg"
}
```

#### 3. Presigned URL (Client-side upload)

```http
POST /api/v1/images/presigned-upload
Content-Type: application/json
Authorization: Bearer <token>

{
  "content_type": "image/jpeg",
  "position": 1
}

Response:
{
  "upload_url": "https://...",
  "public_url": "https://...",
  "image_id": "uuid",
  "expires_at": "2024-..."
}
```

### Image Management

#### List User Images

```http
GET /api/v1/images
Authorization: Bearer <token>
```

#### Download Image

```http
GET /api/v1/images/download/{imageKey}
Authorization: Bearer <token>
```

#### Delete Image

```http
DELETE /api/v1/images/{imageKey}
Authorization: Bearer <token>
```

## Storage Structure

Images are organized in S3 with the following structure:

```
bucket-name/
  images/
    {user-id}/
      {image-id}.jpg
      {image-id}.png
      {image-id}_thumb.jpg  # Thumbnails (if processed)
```

## Features

### Security

- User-scoped access control
- Server-side encryption (AES256)
- File type validation
- Size limits (10MB)
- Authorization required for all operations

### Performance

- Direct S3 uploads via presigned URLs
- CDN support via base URL override
- Proper cache headers (1 year max-age)
- Thumbnail URL generation (ready for processing)

### Monitoring

- Structured metadata in S3
- ETag tracking for integrity
- Upload timestamps
- Error handling and logging

## Fallback Behavior

If S3 is not configured (`AWS_S3_BUCKET` not set), the service:

- Falls back to simulated photo uploads in user handlers
- Continues to accept requests without errors
- Returns placeholder URLs for development

## Example Configuration

### Docker Compose

```yaml
services:
  matching-api:
    image: matching-api
    environment:
      - AWS_S3_BUCKET=my-matching-images
      - AWS_REGION=us-west-2
      - AWS_ACCESS_KEY_ID=your-key
      - AWS_SECRET_ACCESS_KEY=your-secret
      - AWS_S3_BASE_URL=https://cdn.mydomain.com
```

### Kubernetes

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: matching-api-config
data:
  AWS_S3_BUCKET: "my-matching-images"
  AWS_REGION: "us-west-2"
---
apiVersion: v1
kind: Secret
metadata:
  name: aws-credentials
data:
  AWS_ACCESS_KEY_ID: <base64-encoded>
  AWS_SECRET_ACCESS_KEY: <base64-encoded>
```

## Development

For local development without S3:

```bash
# No S3 configuration needed
go run ./cmd/server
```

For local development with S3:

```bash
export AWS_S3_BUCKET=my-dev-bucket
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=your-key
export AWS_SECRET_ACCESS_KEY=your-secret
go run ./cmd/server
```

## Testing

The service includes comprehensive error handling for:

- Missing AWS credentials
- Invalid bucket access
- Network timeouts
- File validation failures
- Authorization errors

All operations return proper HTTP status codes and JSON error messages.
