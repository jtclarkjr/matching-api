package image

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"matching-api/internal/database"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// ListUserImages lists all images for the authenticated user
func (h *Handler) ListUserImages(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// Query user images
	query := `
		SELECT id, user_id, s3_key, url, thumbnail_url, position, 
		       content_type, size, etag, is_active, created_at, updated_at
		FROM images
		WHERE user_id = $1 AND is_active = true
		ORDER BY position ASC
	`

	rows, err := database.DB.Query(query, user.UserID)
	if err != nil {
		utils.LogError("Error querying user images", err)
		utils.WriteInternalError(w, err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.LogError("Error closing rows", err)
		}
	}()

	var images []models.Image
	for rows.Next() {
		var image models.Image

		err := rows.Scan(
			&image.ID, &image.UserID, &image.S3Key, &image.URL,
			&image.ThumbnailURL, &image.Position, &image.ContentType,
			&image.Size, &image.ETag, &image.IsActive,
			&image.CreatedAt, &image.UpdatedAt,
		)
		if err != nil {
			utils.LogError("Error scanning image row", err)
			continue
		}

		images = append(images, image)
	}

	utils.WriteSuccessResponse(w, "Images retrieved successfully", map[string]interface{}{
		"images": images,
		"total":  len(images),
	})
}

// Helper functions
func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/jpg",
		"image/png",
		"image/webp",
		"image/gif",
	}

	for _, validType := range validTypes {
		if contentType == validType {
			return true
		}
	}
	return false
}

func decodeBase64Image(data string) ([]byte, error) {
	// Remove data URL prefix if present (e.g., "data:image/jpeg;base64,")
	if strings.Contains(data, ",") {
		parts := strings.Split(data, ",")
		if len(parts) >= 2 {
			data = parts[1]
		}
	}

	return base64.StdEncoding.DecodeString(data)
}

func detectContentType(data []byte) string {
	if len(data) < 12 {
		return "application/octet-stream"
	}

	// Check for common image file signatures
	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "image/gif"
	}
	if string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP" {
		return "image/webp"
	}

	return "image/jpeg" // default
}

// UploadImage handles image upload
func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Parse multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		utils.WriteErrorResponse(w, "Failed to parse multipart form", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("image")
	if err != nil {
		utils.WriteErrorResponse(w, "Image file is required", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			utils.LogError("Error closing file", err)
		}
	}()

	// Get position from form (optional, defaults to 1)
	position := 1
	if posStr := r.FormValue("position"); posStr != "" {
		if pos, err := utils.GetQueryParamInt(r, "position", 1); err == nil && pos >= 1 && pos <= 9 {
			position = pos
		}
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg" // default
	}

	if !isValidImageType(contentType) {
		utils.WriteErrorResponse(w, "Invalid image type. Only JPEG, PNG, and WebP are allowed", http.StatusBadRequest)
		return
	}

	// Check if S3 service is available
	if h.S3Service == nil {
		utils.WriteErrorResponse(w, "S3 service not configured", http.StatusInternalServerError)
		return
	}

	// Upload to S3
	ctx := context.Background()
	uploadResult, err := h.S3Service.UploadImage(ctx, user.UserID, file, contentType, header.Size)
	if err != nil {
		utils.LogError("Error uploading file to S3", err)
		utils.WriteInternalError(w, err)
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// Remove existing image at this position (if any)
	_, err = database.DB.Exec(`
		UPDATE images SET is_active = false 
		WHERE user_id = $1 AND position = $2 AND is_active = true
	`, user.UserID, position)
	if err != nil {
		utils.LogError("Error deactivating existing image", err)
	}

	// Insert new image record
	var imageID string
	insertQuery := `
		INSERT INTO images (id, user_id, s3_key, url, thumbnail_url, position, 
		                   content_type, size, etag, is_active, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, true, NOW(), NOW())
		RETURNING id
	`

	err = database.DB.QueryRow(insertQuery,
		user.UserID, uploadResult.Key, uploadResult.URL, uploadResult.ThumbnailURL, position,
		contentType, header.Size, uploadResult.ETag,
	).Scan(&imageID)

	if err != nil {
		utils.LogError("Error inserting image record", err)
		utils.WriteInternalError(w, err)
		return
	}

	response := models.ImageUploadResponse{
		ImageID:      imageID,
		URL:          uploadResult.URL,
		ThumbnailURL: uploadResult.ThumbnailURL,
		Position:     position,
		S3Key:        uploadResult.Key,
		Size:         header.Size,
		ContentType:  contentType,
	}

	utils.WriteCreated(w, "Image uploaded successfully", response)
}

// UploadImageBase64 handles base64 image upload
func (h *Handler) UploadImageBase64(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Parse and validate request body
	var req models.ImageUploadBase64Request
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Check if S3 service is available
	if h.S3Service == nil {
		utils.WriteErrorResponse(w, "S3 service not configured", http.StatusInternalServerError)
		return
	}

	// Decode base64 image data
	imageData, err := decodeBase64Image(req.ImageData)
	if err != nil {
		utils.WriteErrorResponse(w, "Invalid base64 image data", http.StatusBadRequest)
		return
	}

	// Detect content type from image data
	contentType := detectContentType(imageData)
	if !isValidImageType(contentType) {
		utils.WriteErrorResponse(w, "Invalid image type", http.StatusBadRequest)
		return
	}

	// Upload to S3
	ctx := context.Background()
	uploadResult, err := h.S3Service.UploadImage(ctx, user.UserID, strings.NewReader(string(imageData)), contentType, int64(len(imageData)))
	if err != nil {
		utils.LogError("Error uploading base64 image to S3", err)
		utils.WriteInternalError(w, err)
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// Remove existing image at this position (if any)
	_, err = database.DB.Exec(`
		UPDATE images SET is_active = false 
		WHERE user_id = $1 AND position = $2 AND is_active = true
	`, user.UserID, req.Position)
	if err != nil {
		utils.LogError("Error deactivating existing image", err)
	}

	// Insert new image record
	var imageID string
	insertQuery := `
		INSERT INTO images (id, user_id, s3_key, url, thumbnail_url, position, 
		                   content_type, size, etag, is_active, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, true, NOW(), NOW())
		RETURNING id
	`

	err = database.DB.QueryRow(insertQuery,
		user.UserID, uploadResult.Key, uploadResult.URL, uploadResult.ThumbnailURL, req.Position,
		contentType, int64(len(imageData)), uploadResult.ETag,
	).Scan(&imageID)

	if err != nil {
		utils.LogError("Error inserting image record", err)
		utils.WriteInternalError(w, err)
		return
	}

	response := models.ImageUploadResponse{
		ImageID:      imageID,
		URL:          uploadResult.URL,
		ThumbnailURL: uploadResult.ThumbnailURL,
		Position:     req.Position,
		S3Key:        uploadResult.Key,
		Size:         int64(len(imageData)),
		ContentType:  contentType,
	}

	utils.WriteCreated(w, "Base64 image uploaded successfully", response)
}

// GeneratePresignedUploadURL generates a presigned URL for image upload
func (h *Handler) GeneratePresignedUploadURL(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Parse and validate request body
	var req models.PresignedUploadRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Validate content type
	if !isValidImageType(req.ContentType) {
		utils.WriteErrorResponse(w, "Invalid content type", http.StatusBadRequest)
		return
	}

	// Check if S3 service is available
	if h.S3Service == nil {
		utils.WriteErrorResponse(w, "S3 service not configured", http.StatusInternalServerError)
		return
	}

	// Generate presigned URL (15 minutes expiration)
	ctx := context.Background()
	duration := 15 * time.Minute
	presignedUpload, err := h.S3Service.GeneratePresignedURL(ctx, user.UserID, req.ContentType, duration)
	if err != nil {
		utils.LogError("Error generating presigned URL", err)
		utils.WriteInternalError(w, err)
		return
	}

	// Prepare database record for when upload completes
	// Note: In a real implementation, you might want to create a pending upload record

	response := models.PresignedUploadResponse{
		UploadURL:    presignedUpload.URL,
		ImageID:      presignedUpload.ImageID,
		S3Key:        presignedUpload.Key,
		PublicURL:    presignedUpload.PublicURL,
		ThumbnailURL: presignedUpload.ThumbnailURL,
		ExpiresAt:    presignedUpload.ExpiresAt,
	}

	utils.WriteSuccessResponse(w, "Presigned URL generated successfully", response)
}

// DownloadImage handles image download
func (h *Handler) DownloadImage(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get image key from URL parameter
	imageKey := r.URL.Query().Get("imageKey")
	if imageKey == "" {
		utils.WriteErrorResponse(w, "Image key is required", http.StatusBadRequest)
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// Verify user owns this image
	var imageExists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM images 
			WHERE s3_key = $1 AND user_id = $2 AND is_active = true
		)
	`
	err := database.DB.QueryRow(checkQuery, imageKey, user.UserID).Scan(&imageExists)
	if err != nil {
		utils.LogError("Error checking image ownership", err)
		utils.WriteInternalError(w, err)
		return
	}

	if !imageExists {
		utils.WriteErrorResponse(w, "Image not found or access denied", http.StatusNotFound)
		return
	}

	// Check if S3 service is available
	if h.S3Service == nil {
		utils.WriteErrorResponse(w, "S3 service not configured", http.StatusInternalServerError)
		return
	}

	// Download from S3
	ctx := context.Background()
	imageData, metadata, err := h.S3Service.DownloadImage(ctx, imageKey)
	if err != nil {
		utils.LogError("Error downloading image from S3", err)
		utils.WriteInternalError(w, err)
		return
	}
	defer func() {
		if err := imageData.Close(); err != nil {
			utils.LogError("Error closing image data", err)
		}
	}()

	// Set response headers
	w.Header().Set("Content-Type", *metadata.ContentType)
	w.Header().Set("Content-Length", fmt.Sprintf("%d", *metadata.ContentLength))
	w.Header().Set("Cache-Control", "max-age=31536000") // 1 year

	// Stream the image data
	_, err = fmt.Fprintf(w, "%s", imageData)
	if err != nil {
		utils.LogError("Error streaming image data", err)
	}
}

// DeleteImage handles image deletion
func (h *Handler) DeleteImage(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get image key from URL parameter
	imageKey := r.URL.Query().Get("imageKey")
	if imageKey == "" {
		utils.WriteErrorResponse(w, "Image key is required", http.StatusBadRequest)
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// Verify user owns this image and get S3 key
	var s3Key string
	var imageID string
	checkQuery := `
		SELECT id, s3_key FROM images 
		WHERE s3_key = $1 AND user_id = $2 AND is_active = true
	`
	err := database.DB.QueryRow(checkQuery, imageKey, user.UserID).Scan(&imageID, &s3Key)
	if err != nil {
		utils.WriteErrorResponse(w, "Image not found or access denied", http.StatusNotFound)
		return
	}

	// Mark image as inactive in database
	updateQuery := `
		UPDATE images 
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`
	_, err = database.DB.Exec(updateQuery, imageID)
	if err != nil {
		utils.LogError("Error marking image as inactive", err)
		utils.WriteInternalError(w, err)
		return
	}

	// Delete from S3 (optional - you might want to keep files for recovery)
	if h.S3Service != nil {
		ctx := context.Background()
		err = h.S3Service.DeleteImage(ctx, s3Key)
		if err != nil {
			utils.LogError("Error deleting image from S3", err)
			// Don't fail the request if S3 deletion fails
		}
	}

	utils.WriteSuccessResponse(w, "Image deleted successfully", nil)
}
