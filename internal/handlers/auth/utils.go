package auth

import (
	"time"

	"golang.org/x/crypto/bcrypt"

	"matching-api/internal/models"
	"matching-api/pkg/auth"
	"matching-api/pkg/utils"
)

// getUserByEmail retrieves user by email from database
func (h *Handler) getUserByEmail(email string) *models.User {
	// TODO: Implement database query to find user by email
	// SELECT * FROM users WHERE email = ? AND is_active = true
	
	// Placeholder mock data until database integration
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	
	return &models.User{
		ID:        "user-123",
		Email:     email,
		Password:  string(hashedPassword),
		FirstName: "John",
		LastName:  "Doe",
		Age:       28,
		Gender:    "male",
		Bio:       "Love hiking and good coffee",
		IsActive:  true,
		CreatedAt: time.Now().AddDate(0, -1, 0), // Created 1 month ago
		UpdatedAt: time.Now(),
	}
}

// verifyRefreshToken verifies refresh token and returns user if valid
func (h *Handler) verifyRefreshToken(token string) *models.User {
	var user *models.User
	
	if h.RedisService != nil {
		var sessionData map[string]interface{}
		if err := h.RedisService.GetSession(token, &sessionData); err == nil {
			// Session found in Redis, validate it
			userID, ok := sessionData["user_id"].(string)
			if !ok {
				return nil
			}
			
			// Try to get user from cache first
			var cachedUser models.User
			if err := h.RedisService.GetCachedUser(userID, &cachedUser); err == nil {
				user = &cachedUser
			} else {
				// Fallback to database query
				user = h.validateRefreshToken(token)
			}
		}
	} else {
		// Fallback to database query when Redis is not available
		user = h.validateRefreshToken(token)
	}
	
	return user
}

// validateRefreshToken validates refresh token against database
func (h *Handler) validateRefreshToken(token string) *models.User {
	// TODO: Implement database token validation
	// SELECT u.* FROM users u JOIN refresh_tokens rt ON u.id = rt.user_id 
	// WHERE rt.token = ? AND rt.expires_at > NOW() AND rt.is_active = true
	
	// Placeholder mock data until database integration
	return &models.User{
		ID:        "user-123",
		Email:     "john@example.com",
		FirstName: "John",
		LastName:  "Doe",
		Age:       28,
		Gender:    "male",
		IsActive:  true,
		CreatedAt: time.Now().AddDate(0, -1, 0),
		UpdatedAt: time.Now(),
	}
}

// storeUserSession stores user session in Redis and caches user data
func (h *Handler) storeUserSession(user *models.User, refreshToken string) error {
	if h.RedisService == nil {
		return nil // No Redis service available
	}

	// Store session data
	sessionData := map[string]interface{}{
		"user_id":       user.ID,
		"email":         user.Email,
		"refresh_token": refreshToken,
		"created_at":    time.Now(),
	}
	
	if err := h.RedisService.StoreSession(refreshToken, sessionData, auth.RefreshTokenExpiry); err != nil {
		utils.LogError("Failed to store session in Redis", err)
		return err
	}
	
	// Cache user data for faster profile lookups
	if err := h.RedisService.CacheUser(user.ID, user.Public(), 30*time.Minute); err != nil {
		utils.LogError("Failed to cache user data", err)
		return err
	}

	return nil
}

// updateUserSession updates existing session with new refresh token
func (h *Handler) updateUserSession(oldToken string, user *models.User, newToken string) error {
	if h.RedisService == nil {
		return nil // No Redis service available
	}

	// Delete old session
	if err := h.RedisService.DeleteSession(oldToken); err != nil {
		utils.LogError("Failed to delete old session", err)
	}
	
	// Store new session
	sessionData := map[string]interface{}{
		"user_id":       user.ID,
		"email":         user.Email,
		"refresh_token": newToken,
		"refreshed_at":  time.Now(),
	}
	
	if err := h.RedisService.StoreSession(newToken, sessionData, auth.RefreshTokenExpiry); err != nil {
		utils.LogError("Failed to store new session", err)
		return err
	}
	
	// Update cached user data
	if err := h.RedisService.CacheUser(user.ID, user.Public(), 30*time.Minute); err != nil {
		utils.LogError("Failed to update cached user data", err)
		return err
	}

	return nil
}