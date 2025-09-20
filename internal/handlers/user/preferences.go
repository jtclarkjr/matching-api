package user

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// GetPreferences retrieves user's matching preferences
// @Summary Get user preferences
// @Description Retrieve the current user's matching preferences
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=models.UserPrefs} "Preferences retrieved successfully"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /users/preferences [get]
func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get user preferences from database
	preferences := h.getUserPreferences(userClaims.UserID)
	
	utils.WriteSuccessResponse(w, "Preferences retrieved successfully", preferences)
}

// UpdatePreferences updates user's matching preferences
// @Summary Update user preferences
// @Description Update the current user's matching preferences
// @Tags User Management
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UpdatePreferencesRequest true "Preferences update data"
// @Success 200 {object} models.APIResponse{data=models.UserPrefs} "Preferences updated successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request - validation failed"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /users/preferences [put]
func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	var req models.UpdatePreferencesRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Validate age range
	if req.AgeMin != nil && req.AgeMax != nil && *req.AgeMin > *req.AgeMax {
		utils.WriteErrorResponse(w, "Minimum age cannot be greater than maximum age", http.StatusBadRequest)
		return
	}

	// Get current preferences from database
	preferences := h.getUserPreferences(userClaims.UserID)
	if preferences == nil {
		// Create new preferences if they don't exist
		preferences = &models.UserPrefs{
			ID:        uuid.New().String(),
			UserID:    userClaims.UserID,
			AgeMin:    18,
			AgeMax:    99,
			MaxDistance: 50,
			InterestedIn: []string{"female"},
			ShowMe:    "everyone",
			CreatedAt: time.Now(),
		}
	}

	// Update fields that were provided
	if req.AgeMin != nil {
		preferences.AgeMin = *req.AgeMin
	}
	if req.AgeMax != nil {
		preferences.AgeMax = *req.AgeMax
	}
	if req.MaxDistance != nil {
		preferences.MaxDistance = *req.MaxDistance
	}
	if req.InterestedIn != nil {
		preferences.InterestedIn = req.InterestedIn
	}
	if req.ShowMe != nil {
		preferences.ShowMe = *req.ShowMe
	}
	if req.OnlyVerified != nil {
		preferences.OnlyVerified = *req.OnlyVerified
	}
	if req.HideDistance != nil {
		preferences.HideDistance = *req.HideDistance
	}
	if req.HideAge != nil {
		preferences.HideAge = *req.HideAge
	}

	preferences.UpdatedAt = time.Now()

	// TODO: Save updated preferences to database
	
	utils.WriteSuccessResponse(w, "Preferences updated successfully", preferences)
}