package match

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// GetPotentialMatches retrieves potential matches for the user based on preferences
// @Summary Get potential matches
// @Description Retrieve potential matches based on user preferences and compatibility
// @Tags Matching
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Number of potential matches (max 50)" default(10)
// @Success 200 {object} models.APIResponse{data=[]models.User} "Potential matches found"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid parameters"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /matches/potential [get]
func (h *Handler) GetPotentialMatches(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get limit parameter
	limit, err := utils.GetQueryParamInt(r, "limit", 10)
	if err != nil {
		utils.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	if limit < 1 || limit > 50 {
		limit = 10
	}

	// Get user's current location and preferences
	user := h.simulateGetUser(userClaims.UserID)
	if user == nil {
		utils.WriteNotFound(w, "User not found")
		return
	}

	preferences := h.simulateGetUserPreferences(userClaims.UserID)
	if preferences == nil {
		utils.WriteErrorResponse(w, "User preferences not found", http.StatusBadRequest)
		return
	}

	// Try to get potential matches from cache first
	cacheKey := fmt.Sprintf("potential_matches_%s_%d", userClaims.UserID, limit)
	var potentialMatches []models.User
	
	if h.RedisService != nil {
		if err := h.RedisService.Get(cacheKey, &potentialMatches); err == nil {
			log.Printf("Potential matches served from cache for user: %s", userClaims.UserID)
		} else {
			// Cache miss, compute potential matches
			potentialMatches = h.findPotentialMatches(user, preferences, limit)
			
			// Cache the results for 5 minutes (potential matches change more frequently)
			if err := h.RedisService.Set(cacheKey, potentialMatches, 5*time.Minute); err != nil {
				log.Printf("Warning: Failed to cache potential matches: %v", err)
			}
		}
	} else {
		// Fallback when Redis is not available
		potentialMatches = h.findPotentialMatches(user, preferences, limit)
	}

	utils.WriteSuccessResponse(w, "Potential matches found", potentialMatches)
}