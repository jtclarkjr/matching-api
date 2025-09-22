package chat

import (
	"matching-api/internal/handlers/shared"
	"matching-api/pkg/services"
)

// Handler handles chat-related requests
type Handler struct {
	shared.BaseHandler
}

// NewHandler creates a new chat handler
func NewHandler(redisService *services.RedisService) *Handler {
	return &Handler{
		BaseHandler: shared.NewBaseHandler(redisService),
	}
}
