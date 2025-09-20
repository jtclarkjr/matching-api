package user

import (
	"matching-api/internal/handlers/shared"
	"matching-api/pkg/services"
)

// Handler handles user profile-related requests
type Handler struct {
	shared.BaseHandler
	S3Service *services.S3Service
}

// NewHandler creates a new user handler
func NewHandler(s3Service *services.S3Service, redisService *services.RedisService) *Handler {
	return &Handler{
		BaseHandler: shared.NewBaseHandler(redisService),
		S3Service:   s3Service,
	}
}