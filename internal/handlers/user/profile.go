package user

import (
	"log"
	"net/http"
	"time"

	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// GetProfile retrieves the current user's profile
// @Summary Get user profile
// @Description Retrieve the current user's profile information
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.User} "Profile retrieved successfully"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 404 {object} models.ErrorResponse "User not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /users/profile [get]
func (h *Handler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Try to get user from Redis cache first
	var user *models.User
	if h.RedisService != nil {
		var cachedUser models.User
		if err := h.RedisService.GetCachedUser(userClaims.UserID, &cachedUser); err == nil {
			user = &cachedUser
			log.Printf("User profile served from cache for user: %s", userClaims.UserID)
		}
	}
	
	// If not in cache, get from database
	if user == nil {
		user = h.getUserProfile(userClaims.UserID)
		if user == nil {
			utils.WriteNotFound(w, "User not found")
			return
		}
		
		// Cache the user data for future requests
		if h.RedisService != nil {
			if err := h.RedisService.CacheUser(user.ID, user, 30*time.Minute); err != nil {
				log.Printf("Warning: Failed to cache user profile: %v", err)
			}
		}
	}

	utils.WriteSuccessResponse(w, "Profile retrieved successfully", user)
}

// UpdateProfile updates the current user's profile
// @Summary Update user profile
// @Description Update the current user's profile information
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UpdateProfileRequest true "Profile update data"
// @Success 200 {object} models.APIResponse{data=models.User} "Profile updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request - validation failed"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /users/profile [put]
func (h *Handler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	var req models.UpdateProfileRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Get current user from database
	user := h.getUserProfile(userClaims.UserID)
	if user == nil {
		utils.WriteNotFound(w, "User not found")
		return
	}

	// Update fields that were provided
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Age != nil {
		user.Age = *req.Age
	}
	if req.Bio != nil {
		user.Bio = *req.Bio
	}
	if req.Location != nil {
		user.Location = req.Location
	}

	user.UpdatedAt = time.Now()

	// TODO: Save updated user to database
	
	// Invalidate cache and update with new data
	if h.RedisService != nil {
		// Clear old cache
		if err := h.RedisService.InvalidateUserCache(userClaims.UserID); err != nil {
			log.Printf("Warning: Failed to invalidate user cache: %v", err)
		}
		
		// Cache updated user data
		if err := h.RedisService.CacheUser(user.ID, user.Public(), 30*time.Minute); err != nil {
			log.Printf("Warning: Failed to cache updated user data: %v", err)
		}
	}
	
	utils.WriteSuccessResponse(w, "Profile updated successfully", user.Public())
}