package services

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/uuid"
)

// S3Service handles S3 operations for image storage
type S3Service struct {
	client     *s3.Client
	uploader   *manager.Uploader
	downloader *manager.Downloader
	bucket     string
	region     string
	baseURL    string
}

// S3Config holds S3 configuration
type S3Config struct {
	Bucket  string
	Region  string
	BaseURL string // Optional CDN URL
}

// UploadResult represents the result of an S3 upload
type UploadResult struct {
	URL          string
	Key          string
	ETag         string
	Size         int64
	ContentType  string
	ThumbnailURL string // Optional thumbnail URL
}

// NewS3Service creates a new S3 service instance
func NewS3Service(cfg S3Config) (*S3Service, error) {
	// Load AWS config
	awsCfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(cfg.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client
	client := s3.NewFromConfig(awsCfg)

	// Create upload and download managers
	uploader := manager.NewUploader(client)
	downloader := manager.NewDownloader(client)

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = fmt.Sprintf("https://%s.s3.%s.amazonaws.com", cfg.Bucket, cfg.Region)
	}

	return &S3Service{
		client:     client,
		uploader:   uploader,
		downloader: downloader,
		bucket:     cfg.Bucket,
		region:     cfg.Region,
		baseURL:    baseURL,
	}, nil
}

// UploadImage uploads an image to S3
func (s *S3Service) UploadImage(ctx context.Context, userID string, imageData io.Reader, contentType string, size int64) (*UploadResult, error) {
	// Generate unique key
	imageID := uuid.New().String()
	extension := getFileExtension(contentType)
	key := fmt.Sprintf("images/%s/%s%s", userID, imageID, extension)

	// Read image data into buffer for multiple operations
	buffer := &bytes.Buffer{}
	writtenSize, err := io.Copy(buffer, imageData)
	if err != nil {
		return nil, fmt.Errorf("failed to read image data: %w", err)
	}

	// Validate size
	if writtenSize != size {
		return nil, fmt.Errorf("size mismatch: expected %d, got %d", size, writtenSize)
	}

	// Upload to S3
	uploadInput := &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(buffer.Bytes()),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
		Metadata: map[string]string{
			"user-id":    userID,
			"image-id":   imageID,
			"uploaded":   time.Now().UTC().Format(time.RFC3339),
			"file-size":  fmt.Sprintf("%d", size),
		},
		ServerSideEncryption: types.ServerSideEncryptionAes256,
		CacheControl:         aws.String("max-age=31536000"), // 1 year
	}

	result, err := s.uploader.Upload(ctx, uploadInput)
	if err != nil {
		return nil, fmt.Errorf("failed to upload image: %w", err)
	}

	// Generate thumbnail (optional - would require image processing)
	thumbnailURL := s.generateThumbnailURL(key)

	return &UploadResult{
		URL:          s.getPublicURL(key),
		Key:          key,
		ETag:         strings.Trim(*result.ETag, "\""),
		Size:         size,
		ContentType:  contentType,
		ThumbnailURL: thumbnailURL,
	}, nil
}

// DownloadImage downloads an image from S3
func (s *S3Service) DownloadImage(ctx context.Context, key string) (io.ReadCloser, *s3.GetObjectOutput, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to download image: %w", err)
	}

	return result.Body, result, nil
}

// DeleteImage deletes an image from S3
func (s *S3Service) DeleteImage(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete image: %w", err)
	}

	return nil
}

// DeleteImages deletes multiple images from S3
func (s *S3Service) DeleteImages(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	// Convert keys to ObjectIdentifier slice
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{
			Key: aws.String(key),
		}
	}

	input := &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
		Delete: &types.Delete{
			Objects: objects,
		},
	}

	_, err := s.client.DeleteObjects(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete images: %w", err)
	}

	return nil
}

// GeneratePresignedURL generates a presigned URL for direct upload
func (s *S3Service) GeneratePresignedURL(ctx context.Context, userID string, contentType string, duration time.Duration) (*PresignedUpload, error) {
	imageID := uuid.New().String()
	extension := getFileExtension(contentType)
	key := fmt.Sprintf("images/%s/%s%s", userID, imageID, extension)

	// Create presign client
	presignClient := s3.NewPresignClient(s.client)

	// Generate presigned PUT URL
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
		Metadata: map[string]string{
			"user-id":  userID,
			"image-id": imageID,
		},
		ServerSideEncryption: types.ServerSideEncryptionAes256,
		CacheControl:         aws.String("max-age=31536000"),
	}

	presignedReq, err := presignClient.PresignPutObject(ctx, input, func(opts *s3.PresignOptions) {
		opts.Expires = duration
	})
	if err != nil {
		return nil, fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return &PresignedUpload{
		URL:          presignedReq.URL,
		Key:          key,
		ImageID:      imageID,
		ExpiresAt:    time.Now().Add(duration),
		PublicURL:    s.getPublicURL(key),
		ThumbnailURL: s.generateThumbnailURL(key),
	}, nil
}

// GetImageMetadata retrieves metadata for an image
func (s *S3Service) GetImageMetadata(ctx context.Context, key string) (*ImageMetadata, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get image metadata: %w", err)
	}

	metadata := &ImageMetadata{
		Key:          key,
		Size:         *result.ContentLength,
		ContentType:  *result.ContentType,
		ETag:         strings.Trim(*result.ETag, "\""),
		LastModified: *result.LastModified,
		URL:          s.getPublicURL(key),
	}

	// Extract custom metadata
	if result.Metadata != nil {
		if userID, ok := result.Metadata["user-id"]; ok {
			metadata.UserID = userID
		}
		if imageID, ok := result.Metadata["image-id"]; ok {
			metadata.ImageID = imageID
		}
	}

	return metadata, nil
}

// ListUserImages lists all images for a user
func (s *S3Service) ListUserImages(ctx context.Context, userID string) ([]ImageMetadata, error) {
	prefix := fmt.Sprintf("images/%s/", userID)
	
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	}

	var images []ImageMetadata
	paginator := s3.NewListObjectsV2Paginator(s.client, input)

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to list images: %w", err)
		}

		for _, obj := range page.Contents {
			images = append(images, ImageMetadata{
				Key:          *obj.Key,
				Size:         *obj.Size,
				ETag:         strings.Trim(*obj.ETag, "\""),
				LastModified: *obj.LastModified,
				URL:          s.getPublicURL(*obj.Key),
			})
		}
	}

	return images, nil
}

// Helper functions

func (s *S3Service) getPublicURL(key string) string {
	return fmt.Sprintf("%s/%s", s.baseURL, key)
}

func (s *S3Service) generateThumbnailURL(key string) string {
	// This would integrate with image processing service or Lambda
	// For now, return a placeholder or the same URL
	parts := strings.Split(key, ".")
	if len(parts) > 1 {
		thumbnailKey := strings.Join(parts[:len(parts)-1], ".") + "_thumb." + parts[len(parts)-1]
		return s.getPublicURL(thumbnailKey)
	}
	return s.getPublicURL(key)
}

func getFileExtension(contentType string) string {
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil || len(extensions) == 0 {
		// Default extensions based on common types
		switch contentType {
		case "image/jpeg", "image/jpg":
			return ".jpg"
		case "image/png":
			return ".png"
		case "image/webp":
			return ".webp"
		case "image/gif":
			return ".gif"
		default:
			return ".jpg" // Default
		}
	}
	return extensions[0]
}

// Support types

// PresignedUpload represents a presigned upload URL
type PresignedUpload struct {
	URL          string    `json:"url"`
	Key          string    `json:"key"`
	ImageID      string    `json:"image_id"`
	ExpiresAt    time.Time `json:"expires_at"`
	PublicURL    string    `json:"public_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
}

// ImageMetadata represents S3 image metadata
type ImageMetadata struct {
	Key          string    `json:"key"`
	ImageID      string    `json:"image_id,omitempty"`
	UserID       string    `json:"user_id,omitempty"`
	Size         int64     `json:"size"`
	ContentType  string    `json:"content_type,omitempty"`
	ETag         string    `json:"etag"`
	LastModified time.Time `json:"last_modified"`
	URL          string    `json:"url"`
}