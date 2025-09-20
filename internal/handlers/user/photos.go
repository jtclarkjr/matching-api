package user

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// UploadPhoto handles photo upload for user profile
// @Summary Upload profile photo
// @Description Upload a new photo to user profile
// @Tags User Management
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param photo formData file true "Photo file (JPEG/PNG, max 5MB)"
// @Param position query int false "Photo position (1-9)" default(1)
// @Success 201 {object} models.APIResponse{data=models.Photo} "Photo uploaded successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid file"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /users/photos [post]
func (h *Handler) UploadPhoto(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Parse multipart form (limit to 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		utils.WriteErrorResponse(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("photo")
	if err != nil {
		utils.WriteErrorResponse(w, "Photo file is required", http.StatusBadRequest)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Printf("Error closing file: %v", err)
		}
	}()

	// Validate file type and size
	if !h.isValidImageType(header.Header.Get("Content-Type")) {
		utils.WriteErrorResponse(w, "Invalid file type. Only JPEG and PNG are allowed", http.StatusBadRequest)
		return
	}

	if header.Size > 5<<20 { // 5MB limit
		utils.WriteErrorResponse(w, "File too large. Maximum size is 5MB", http.StatusBadRequest)
		return
	}

	// Get position for photo ordering
	position, err := utils.GetQueryParamInt(r, "position", 1)
	if err != nil {
		utils.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if position < 1 || position > 9 {
		position = 1
	}

	// Upload to S3 if service is available
	if h.S3Service != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		contentType := header.Header.Get("Content-Type")
		result, err := h.S3Service.UploadImage(ctx, userClaims.UserID, file, contentType, header.Size)
		if err != nil {
			utils.WriteErrorResponse(w, fmt.Sprintf("Failed to upload to S3: %v", err), http.StatusInternalServerError)
			return
		}

		// Create photo record with S3 data
		photo := &models.Photo{
			ID:        utils.ExtractImageIDFromKey(result.Key),
			UserID:    userClaims.UserID,
			URL:       result.URL,
			Position:  position,
			CreatedAt: time.Now(),
		}

		utils.WriteCreated(w, "Photo uploaded successfully", photo)
	} else {
		// Fallback to simulation if S3 service not available
		photoURL := h.simulatePhotoUpload(header.Filename)
		
		// Create photo record
		photo := &models.Photo{
			ID:        uuid.New().String(),
			UserID:    userClaims.UserID,
			URL:       photoURL,
			Position:  position,
			CreatedAt: time.Now(),
		}

		utils.WriteCreated(w, "Photo uploaded successfully", photo)
	}
}

// DeletePhoto deletes a user's photo
// @Summary Delete profile photo
// @Description Delete a specific photo from user profile
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param photoID path string true "Photo ID"
// @Success 200 {object} models.APIResponse "Photo deleted successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid photo ID"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 403 {object} models.ErrorResponse "Forbidden - photo doesn't belong to user"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /users/photos/{photoID} [delete]
func (h *Handler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	_, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	photoID := chi.URLParam(r, "photoID")
	if photoID == "" {
		utils.WriteErrorResponse(w, "Photo ID is required", http.StatusBadRequest)
		return
	}

	// TODO: Verify photo belongs to user and delete from database and storage
	
	utils.WriteSuccessResponse(w, "Photo deleted successfully", nil)
}