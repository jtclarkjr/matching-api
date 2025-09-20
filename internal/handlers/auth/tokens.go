package auth

import (
	"net/http"
	"os"

	"matching-api/internal/models"
	"matching-api/pkg/auth"
	"matching-api/pkg/utils"
)

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Generate new access token using refresh token
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token"
// @Success 200 {object} models.APIResponse{data=models.AuthResponse} "Token refreshed successfully"
// @Failure 401 {object} models.ErrorResponse "Invalid refresh token"
// @Failure 400 {object} models.ErrorResponse "Bad request - validation failed"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/refresh [post]
func (h *Handler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Verify refresh token and get user
	user := h.verifyRefreshToken(req.RefreshToken)
	if user == nil {
		utils.WriteUnauthorized(w, "Invalid refresh token")
		return
	}

	// Generate new tokens
	accessToken, newRefreshToken, err := h.generateTokenPair(user.ID, user.Email)
	if err != nil {
		utils.WriteInternalError(w, err)
		return
	}

	// Update session with new refresh token
	if err := h.updateUserSession(req.RefreshToken, user, newRefreshToken); err != nil {
		utils.LogError("Failed to update user session during token refresh", err)
	}

	// Prepare response
	authResponse := models.AuthResponse{
		User:         user.Public(),
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    int64(auth.AccessTokenExpiry.Seconds()),
	}

	utils.WriteSuccessResponse(w, "Token refreshed successfully", authResponse)
}

// Logout handles user logout by invalidating the refresh token
// @Summary User logout
// @Description Logout user by invalidating refresh token and clearing session
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "Refresh token to invalidate"
// @Success 200 {object} models.APIResponse "Logout successful"
// @Failure 400 {object} models.ErrorResponse "Bad request - validation failed"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/logout [post]
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	var req models.RefreshTokenRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Delete session from Redis
	if h.RedisService != nil {
		if err := h.RedisService.DeleteSession(req.RefreshToken); err != nil {
			utils.LogError("Failed to delete session during logout", err)
		}
	}

	utils.WriteSuccessResponse(w, "Logout successful", nil)
}

// generateTokenPair generates both access and refresh tokens
func (h *Handler) generateTokenPair(userID, email string) (string, string, error) {
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-super-secret-key-change-this-in-production"
	}
	
	jwtService := auth.NewJWTService(jwtSecret)
	
	accessToken, err := jwtService.GenerateAccessToken(userID, email)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := jwtService.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}