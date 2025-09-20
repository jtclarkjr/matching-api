package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/httprate"
	"matching-api/pkg/services"
)

// RedisStore implements httprate.LimitCounter interface using our Redis service
type RedisStore struct {
	redis         *services.RedisService
	requestLimit  int
	windowLength  time.Duration
}

// NewRedisStore creates a new Redis store for httprate
func NewRedisStore(redisService *services.RedisService) *RedisStore {
	return &RedisStore{
		redis:        redisService,
		requestLimit: 100,
		windowLength: time.Minute,
	}
}

// Config sets the request limit and window length
func (s *RedisStore) Config(requestLimit int, windowLength time.Duration) {
	s.requestLimit = requestLimit
	s.windowLength = windowLength
}

// Increment increments the counter for a key in the current window
func (s *RedisStore) Increment(key string, currentWindow time.Time) error {
	if s.redis == nil {
		return fmt.Errorf("redis service not available")
	}

	windowKey := fmt.Sprintf("%s:%d", key, currentWindow.Unix()/int64(s.windowLength.Seconds()))
	_, err := s.redis.IncrementRateLimit(windowKey, s.windowLength)
	return err
}

// IncrementBy increments the counter for a key by the specified amount
func (s *RedisStore) IncrementBy(key string, currentWindow time.Time, amount int) error {
	if s.redis == nil {
		return fmt.Errorf("redis service not available")
	}

	windowKey := fmt.Sprintf("%s:%d", key, currentWindow.Unix()/int64(s.windowLength.Seconds()))
	
	// Increment multiple times (not ideal, but works with our existing API)
	for i := 0; i < amount; i++ {
		_, err := s.redis.IncrementRateLimit(windowKey, s.windowLength)
		if err != nil {
			return err
		}
	}
	return nil
}

// Get retrieves the current and previous window counts
func (s *RedisStore) Get(key string, currentWindow, previousWindow time.Time) (int, int, error) {
	if s.redis == nil {
		return 0, 0, fmt.Errorf("redis service not available")
	}

	currentWindowKey := fmt.Sprintf("%s:%d", key, currentWindow.Unix()/int64(s.windowLength.Seconds()))
	previousWindowKey := fmt.Sprintf("%s:%d", key, previousWindow.Unix()/int64(s.windowLength.Seconds()))

	currentCount, err := s.redis.GetRateLimit(currentWindowKey)
	if err != nil {
		currentCount = 0 // Default to 0 if not found
	}

	previousCount, err := s.redis.GetRateLimit(previousWindowKey)
	if err != nil {
		previousCount = 0 // Default to 0 if not found
	}

	return int(currentCount), int(previousCount), nil
}

// Rate limiting configurations
func APIRateLimit(redisService *services.RedisService) func(next http.Handler) http.Handler {
	if redisService != nil {
		// Use Redis-backed rate limiting
		store := NewRedisStore(redisService)
		return httprate.Limit(
			100,               // 100 requests
			time.Minute,       // per minute
			httprate.WithLimitCounter(store),
			httprate.WithKeyFuncs(httprate.KeyByIP),
		)
	}
	// Fallback to in-memory rate limiting
	return httprate.LimitByIP(100, time.Minute)
}

func AuthRateLimit(redisService *services.RedisService) func(next http.Handler) http.Handler {
	if redisService != nil {
		// Use Redis-backed rate limiting with stricter limits for auth
		store := NewRedisStore(redisService)
		return httprate.Limit(
			10,                // 10 requests
			time.Minute,       // per minute
			httprate.WithLimitCounter(store),
			httprate.WithKeyFuncs(httprate.KeyByIP),
		)
	}
	// Fallback to in-memory rate limiting
	return httprate.LimitByIP(10, time.Minute)
}

func StrictRateLimit(redisService *services.RedisService) func(next http.Handler) http.Handler {
	if redisService != nil {
		// Use Redis-backed rate limiting with user-based keys when possible
		store := NewRedisStore(redisService)
		return httprate.Limit(
			200,               // 200 requests
			time.Minute,       // per minute
			httprate.WithLimitCounter(store),
			httprate.WithKeyFuncs(KeyByUserOrIP),
		)
	}
	// Fallback to in-memory rate limiting
	return httprate.LimitByIP(200, time.Minute)
}

// KeyByUserOrIP creates a rate limit key using user ID if available, otherwise IP
func KeyByUserOrIP(r *http.Request) (string, error) {
	// Try to get user from context first
	if userClaims, ok := GetUserFromContext(r.Context()); ok && userClaims.UserID != "" {
		return fmt.Sprintf("user:%s", userClaims.UserID), nil
	}
	
	// Fall back to IP-based rate limiting
	ip, err := httprate.KeyByIP(r)
	if err != nil {
		return "", err
	}
	
	return fmt.Sprintf("ip:%s", ip), nil
}