package models

import (
	"time"
)

// Image represents an uploaded image with S3 metadata
type Image struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	S3Key        string    `json:"s3_key"`
	URL          string    `json:"url"`
	ThumbnailURL string    `json:"thumbnail_url,omitempty"`
	Position     int       `json:"position"` // Position in user's photo gallery (1-9)
	ContentType  string    `json:"content_type"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// ImageProcessingJob represents an image processing task
type ImageProcessingJob struct {
	ID           string    `json:"id"`
	UserID       string    `json:"user_id"`
	ImageID      string    `json:"image_id"`
	OriginalKey  string    `json:"original_key"`
	ProcessedKey string    `json:"processed_key"`
	OriginalURL  string    `json:"original_url"`
	ProcessedURL string    `json:"processed_url"`
	Status       string    `json:"status"` // pending, processing, completed, failed
	JobType      string    `json:"job_type"` // resize, crop, filter, compress, thumbnail
	Parameters   map[string]interface{} `json:"parameters"`
	CreatedAt    time.Time `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// ImageUploadRequest represents multipart image upload request
type ImageUploadRequest struct {
	Position int `json:"position" validate:"min=1,max=9"`
	// File will be handled via multipart form, not JSON
}

// ImageUploadBase64Request represents base64 image upload request
type ImageUploadBase64Request struct {
	Position  int    `json:"position" validate:"min=1,max=9"`
	ImageData string `json:"image_data" validate:"required"` // base64 encoded
	Filename  string `json:"filename,omitempty"`
}

// ImageUploadResponse represents upload response  
type ImageUploadResponse struct {
	ImageID      string `json:"image_id"`
	URL          string `json:"url"`
	ThumbnailURL string `json:"thumbnail_url"`
	Position     int    `json:"position"`
	S3Key        string `json:"s3_key"`
	Size         int64  `json:"size"`
	ContentType  string `json:"content_type"`
}

// PresignedUploadRequest represents presigned URL generation request
type PresignedUploadRequest struct {
	ContentType string `json:"content_type" validate:"required"`
	Position    int    `json:"position" validate:"min=1,max=9"`
}

// PresignedUploadResponse represents presigned URL response
type PresignedUploadResponse struct {
	UploadURL    string    `json:"upload_url"`
	ImageID      string    `json:"image_id"`
	S3Key        string    `json:"s3_key"`
	PublicURL    string    `json:"public_url"`
	ThumbnailURL string    `json:"thumbnail_url"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// ImageProcessingOptions contains processing parameters
type ImageProcessingOptions struct {
	MaxWidth     int      `json:"max_width"`
	MaxHeight    int      `json:"max_height"`
	Quality      int      `json:"quality"`
	Format       string   `json:"format"`
	AutoOrient   bool     `json:"auto_orient"`
	RemoveEXIF   bool     `json:"remove_exif"`
	Watermark    bool     `json:"watermark"`
	Blur         *int     `json:"blur,omitempty"`
	Brightness   *float64 `json:"brightness,omitempty"`
	Contrast     *float64 `json:"contrast,omitempty"`
	Saturation   *float64 `json:"saturation,omitempty"`
}

// ImageListResponse represents a list of user images
type ImageListResponse struct {
	Images []Image `json:"images"`
	Total  int     `json:"total"`
}
