package auth

import (
	"matching-api/internal/handlers/shared"
	"matching-api/pkg/services"
)

// Handler handles authentication-related requests
type Handler struct {
	shared.BaseHandler
}

// NewHandler creates a new auth handler
func NewHandler(redisService *services.RedisService) *Handler {
	return &Handler{
		BaseHandler: shared.NewBaseHandler(redisService),
	}
}