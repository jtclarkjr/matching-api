package notification

import (
	"matching-api/internal/handlers/shared"
	"matching-api/pkg/services"
)

// Handler handles notification-related requests
type Handler struct {
	shared.BaseHandler
}

// NewHandler creates a new notification handler
func NewHandler(redisService *services.RedisService) *Handler {
	return &Handler{
		BaseHandler: shared.NewBaseHandler(redisService),
	}
}
