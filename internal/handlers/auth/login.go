package auth

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"matching-api/internal/models"
	"matching-api/pkg/auth"
	"matching-api/pkg/utils"
)

// Login handles user login
// @Summary User login
// @Description Authenticate user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "User login credentials"
// @Success 200 {object} models.APIResponse{data=models.AuthResponse} "Login successful"
// @Failure 401 {object} models.ErrorResponse "Invalid credentials"
// @Failure 400 {object} models.ErrorResponse "Bad request - validation failed"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/login [post]
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Find user by email from database
	user := h.getUserByEmail(req.Email)
	if user == nil {
		utils.WriteUnauthorized(w, "Invalid email or password")
		return
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		utils.WriteUnauthorized(w, "Invalid email or password")
		return
	}

	// Generate tokens
	accessToken, refreshToken, err := h.generateTokenPair(user.ID, user.Email)
	if err != nil {
		utils.WriteInternalError(w, err)
		return
	}

	// Update last seen and store session
	user.LastSeen = &[]time.Time{time.Now()}[0]
	
	if err := h.storeUserSession(user, refreshToken); err != nil {
		utils.LogError("Failed to store user session during login", err)
	}

	// Prepare response
	authResponse := models.AuthResponse{
		User:         user.Public(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(auth.AccessTokenExpiry.Seconds()),
	}

	utils.WriteSuccessResponse(w, "Login successful", authResponse)
}