package match

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// GetMatches retrieves user's current matches
// @Summary Get user matches
// @Description Retrieve user's current matches with pagination
// @Tags Matching
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of matches per page (max 100)" default(20)
// @Success 200 {object} models.APIResponse{data=object{matches=[]models.Match,page=int,limit=int,total=int}} "Matches retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid pagination"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /matches [get]
func (h *Handler) GetMatches(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get pagination parameters
	page, err := utils.GetQueryParamInt(r, "page", 1)
	if err != nil {
		utils.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	limit, err := utils.GetQueryParamInt(r, "limit", 20)
	if err != nil {
		utils.WriteErrorResponse(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate pagination
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	// Try to get matches from Redis cache first
	cacheKey := fmt.Sprintf("matches_page_%s_%d_%d", userClaims.UserID, page, limit)
	var matches []models.Match

	if h.RedisService != nil {
		if err := h.RedisService.Get(cacheKey, &matches); err == nil {
			log.Printf("Matches served from cache for user: %s", userClaims.UserID)
		} else {
			// Cache miss, get from database
			matches = h.getUserMatches(userClaims.UserID, page, limit)

			// Cache the results for 10 minutes
			if err := h.RedisService.Set(cacheKey, matches, 10*time.Minute); err != nil {
				log.Printf("Warning: Failed to cache matches: %v", err)
			}
		}
	} else {
		// Fallback when Redis is not available
		matches = h.getUserMatches(userClaims.UserID, page, limit)
	}

	response := map[string]interface{}{
		"matches": matches,
		"page":    page,
		"limit":   limit,
		"total":   len(matches), // TODO: Get actual total count from database
	}

	utils.WriteSuccessResponse(w, "Matches retrieved successfully", response)
}

// UnMatch removes a match between users
// @Summary Remove a match
// @Description Remove/unmatch with another user
// @Tags Matching
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param matchID path string true "Match ID"
// @Success 200 {object} models.APIResponse "Match removed successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid match ID"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 403 {object} models.ErrorResponse "Forbidden - not part of this match"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /matches/{matchID} [delete]
func (h *Handler) UnMatch(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	matchID := chi.URLParam(r, "matchID")
	if matchID == "" {
		utils.WriteErrorResponse(w, "Match ID is required", http.StatusBadRequest)
		return
	}

	// Verify user is part of this match and deactivate it
	if !h.isUserInMatch(userClaims.UserID, matchID) {
		utils.WriteForbidden(w, "You are not part of this match")
		return
	}

	// TODO: Deactivate match in database

	utils.WriteSuccessResponse(w, "Match removed successfully", nil)
}
