package match

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// Swipe handles user swipe actions (like/pass/super like)
// @Summary Swipe on a user
// @Description Perform a swipe action (like/pass/super like) on another user
// @Tags Matching
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.SwipeRequest true "Swipe action data"
// @Success 200 {object} models.APIResponse{data=object{swipe=models.Swipe,is_match=bool,match_id=string}} "Swipe recorded successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request - validation failed or already swiped"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /matches/swipe [post]
func (h *Handler) Swipe(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	userClaims, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	var req models.SwipeRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Validate that user isn't swiping on themselves
	if req.TargetID == userClaims.UserID {
		utils.WriteErrorResponse(w, "Cannot swipe on yourself", http.StatusBadRequest)
		return
	}

	// Check if user has already swiped on this person
	if h.hasAlreadySwiped(userClaims.UserID, req.TargetID) {
		utils.WriteErrorResponse(w, "Already swiped on this user", http.StatusBadRequest)
		return
	}

	// Create swipe record
	swipe := &models.Swipe{
		ID:        uuid.New().String(),
		UserID:    userClaims.UserID,
		TargetID:  req.TargetID,
		Action:    req.Action,
		CreatedAt: time.Now(),
	}

	// TODO: Save swipe to database

	response := map[string]interface{}{
		"swipe":    swipe,
		"is_match": false,
		"match_id": nil,
	}

	// Check for mutual like (match)
	if req.Action == models.SwipeRight || req.Action == models.SuperLike {
		if h.hasUserLikedBack(req.TargetID, userClaims.UserID) {
			// It's a match! Create match record
			match := h.createMatch(userClaims.UserID, req.TargetID)
			response["is_match"] = true
			response["match_id"] = match.ID
			response["match"] = match

			// Invalidate match caches for both users
			h.invalidateMatchCaches(userClaims.UserID, req.TargetID)

			// TODO: Send notification to both users
		}
	}

	utils.WriteSuccessResponse(w, "Swipe recorded successfully", response)
}

// invalidateMatchCaches invalidates match-related caches for both users
func (h *Handler) invalidateMatchCaches(userID1, userID2 string) {
	if h.RedisService == nil {
		return
	}

	// Clear match cache for current user
	if err := h.RedisService.Delete(fmt.Sprintf("matches_*_%s_*", userID1)); err != nil {
		log.Printf("Warning: Failed to invalidate match cache for user %s: %v", userID1, err)
	}
	
	// Clear match cache for target user
	if err := h.RedisService.Delete(fmt.Sprintf("matches_*_%s_*", userID2)); err != nil {
		log.Printf("Warning: Failed to invalidate match cache for user %s: %v", userID2, err)
	}
	
	// Clear potential matches cache for both users (they shouldn't see each other again)
	if err := h.RedisService.Delete(fmt.Sprintf("potential_matches_%s_*", userID1)); err != nil {
		log.Printf("Warning: Failed to invalidate potential matches cache: %v", err)
	}
	if err := h.RedisService.Delete(fmt.Sprintf("potential_matches_%s_*", userID2)); err != nil {
		log.Printf("Warning: Failed to invalidate potential matches cache: %v", err)
	}
}