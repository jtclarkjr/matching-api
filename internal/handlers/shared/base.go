package shared

import (
	"matching-api/pkg/services"
)

// BaseHandler contains common dependencies used by all handlers
type BaseHandler struct {
	RedisService *services.RedisService
}

// NewBaseHandler creates a new base handler with common dependencies
func NewBaseHandler(redisService *services.RedisService) BaseHandler {
	return BaseHandler{
		RedisService: redisService,
	}
}