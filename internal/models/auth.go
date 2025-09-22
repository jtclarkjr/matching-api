package models

import "time"

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=2,max=50"`
	LastName  string `json:"last_name" validate:"required,min=2,max=50"`
	Age       int    `json:"age" validate:"required,min=18,max=100"`
	Gender    string `json:"gender" validate:"required,oneof=male female non-binary"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User         *User  `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // seconds
}

// RefreshTokenRequest represents the token refresh request
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// JWTClaims represents the JWT token claims
type JWTClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
}

// RefreshToken represents stored refresh tokens
type RefreshToken struct {
	ID        string    `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Token     string    `json:"token" db:"token"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IsRevoked bool      `json:"is_revoked" db:"is_revoked"`
}

// UpdateProfileRequest represents profile update request
type UpdateProfileRequest struct {
	FirstName *string   `json:"first_name,omitempty" validate:"omitempty,min=2,max=50"`
	LastName  *string   `json:"last_name,omitempty" validate:"omitempty,min=2,max=50"`
	Age       *int      `json:"age,omitempty" validate:"omitempty,min=18,max=100"`
	Bio       *string   `json:"bio,omitempty" validate:"omitempty,max=500"`
	Location  *Location `json:"location,omitempty"`
}

// UpdatePreferencesRequest represents preferences update request
type UpdatePreferencesRequest struct {
	AgeMin       *int     `json:"age_min,omitempty" validate:"omitempty,min=18,max=100"`
	AgeMax       *int     `json:"age_max,omitempty" validate:"omitempty,min=18,max=100"`
	MaxDistance  *int     `json:"max_distance,omitempty" validate:"omitempty,min=1,max=100"`
	InterestedIn []string `json:"interested_in,omitempty" validate:"omitempty,dive,oneof=male female non-binary"`
	ShowMe       *string  `json:"show_me,omitempty" validate:"omitempty,oneof=male female non-binary everyone"`
	OnlyVerified *bool    `json:"only_verified,omitempty"`
	HideDistance *bool    `json:"hide_distance,omitempty"`
	HideAge      *bool    `json:"hide_age,omitempty"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Code    int    `json:"code,omitempty"`
}
