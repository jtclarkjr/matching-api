package match

import (
	"matching-api/internal/handlers/shared"
	"matching-api/pkg/services"
)

// Handler handles matching-related requests
type Handler struct {
	shared.BaseHandler
}

// NewHandler creates a new match handler
func NewHandler(redisService *services.RedisService) *Handler {
	return &Handler{
		BaseHandler: shared.NewBaseHandler(redisService),
	}
}