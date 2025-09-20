package auth

import (
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/google/uuid"
	"matching-api/internal/models"
	"matching-api/pkg/auth"
	"matching-api/pkg/utils"
)

// Register handles user registration
// @Summary Register a new user
// @Description Register a new user account with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "User registration data"
// @Success 201 {object} models.APIResponse{data=models.AuthResponse} "User registered successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request - validation failed"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /auth/register [post]
func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req models.RegisterRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Check if user already exists (in real app, check database)
	// TODO: Implement database user existence check
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteInternalError(w, err)
		return
	}

	// Create user
	user := &models.User{
		ID:        uuid.New().String(),
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Age:       req.Age,
		Gender:    req.Gender,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// TODO: Save user to database here
	
	// Generate tokens
	accessToken, refreshToken, err := h.generateTokenPair(user.ID, user.Email)
	if err != nil {
		utils.WriteInternalError(w, err)
		return
	}

	// Store session and cache user data
	if err := h.storeUserSession(user, refreshToken); err != nil {
		// Log error but don't fail the request
		utils.LogError("Failed to store user session during registration", err)
	}
	
	// Prepare response
	authResponse := models.AuthResponse{
		User:         user.Public(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(auth.AccessTokenExpiry.Seconds()),
	}

	utils.WriteCreated(w, "User registered successfully", authResponse)
}