package services

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisService handles all Redis operations
type RedisService struct {
	client *redis.Client
	ctx    context.Context
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// NewRedisService creates a new Redis service instance
func NewRedisService() (*RedisService, error) {
	config := RedisConfig{
		Host:     getEnvOrDefault("REDIS_HOST", "localhost"),
		Port:     getEnvOrDefault("REDIS_PORT", "6379"),
		Password: os.Getenv("REDIS_PASSWORD"), // Empty string if not set
		DB:       getEnvAsIntOrDefault("REDIS_DB", 0),
	}

	addr := fmt.Sprintf("%s:%s", config.Host, config.Port)

	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx := context.Background()

	// Test connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisService{
		client: client,
		ctx:    ctx,
	}, nil
}

// Session Management Methods

// StoreSession stores a session with expiration
func (r *RedisService) StoreSession(sessionID string, data any, expiration time.Duration) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.Set(r.ctx, key, jsonData, expiration).Err()
}

// GetSession retrieves a session
func (r *RedisService) GetSession(sessionID string, dest any) error {
	key := fmt.Sprintf("session:%s", sessionID)
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("session not found")
		}
		return fmt.Errorf("failed to get session: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

// DeleteSession removes a session
func (r *RedisService) DeleteSession(sessionID string) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.Del(r.ctx, key).Err()
}

// ExtendSession extends session expiration
func (r *RedisService) ExtendSession(sessionID string, expiration time.Duration) error {
	key := fmt.Sprintf("session:%s", sessionID)
	return r.client.Expire(r.ctx, key, expiration).Err()
}

// Caching Methods

// CacheUser stores user data with TTL
func (r *RedisService) CacheUser(userID string, userData any, ttl time.Duration) error {
	jsonData, err := json.Marshal(userData)
	if err != nil {
		return fmt.Errorf("failed to marshal user data: %w", err)
	}

	key := fmt.Sprintf("user:%s", userID)
	return r.client.Set(r.ctx, key, jsonData, ttl).Err()
}

// GetCachedUser retrieves cached user data
func (r *RedisService) GetCachedUser(userID string, dest any) error {
	key := fmt.Sprintf("user:%s", userID)
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("user not found in cache")
		}
		return fmt.Errorf("failed to get cached user: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

// CacheMatches stores user matches with TTL
func (r *RedisService) CacheMatches(userID string, matches any, ttl time.Duration) error {
	jsonData, err := json.Marshal(matches)
	if err != nil {
		return fmt.Errorf("failed to marshal matches data: %w", err)
	}

	key := fmt.Sprintf("matches:%s", userID)
	return r.client.Set(r.ctx, key, jsonData, ttl).Err()
}

// GetCachedMatches retrieves cached matches
func (r *RedisService) GetCachedMatches(userID string, dest any) error {
	key := fmt.Sprintf("matches:%s", userID)
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("matches not found in cache")
		}
		return fmt.Errorf("failed to get cached matches: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

// CachePotentialMatches stores potential matches with TTL
func (r *RedisService) CachePotentialMatches(userID string, matches any, ttl time.Duration) error {
	jsonData, err := json.Marshal(matches)
	if err != nil {
		return fmt.Errorf("failed to marshal potential matches data: %w", err)
	}

	key := fmt.Sprintf("potential_matches:%s", userID)
	return r.client.Set(r.ctx, key, jsonData, ttl).Err()
}

// GetCachedPotentialMatches retrieves cached potential matches
func (r *RedisService) GetCachedPotentialMatches(userID string, dest any) error {
	key := fmt.Sprintf("potential_matches:%s", userID)
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("potential matches not found in cache")
		}
		return fmt.Errorf("failed to get cached potential matches: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

// Generic Cache Methods

// Set stores any data with TTL
func (r *RedisService) Set(key string, value any, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	return r.client.Set(r.ctx, key, jsonData, ttl).Err()
}

// Get retrieves any cached data
func (r *RedisService) Get(key string, dest any) error {
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return fmt.Errorf("key not found in cache")
		}
		return fmt.Errorf("failed to get cached data: %w", err)
	}

	return json.Unmarshal([]byte(val), dest)
}

// Delete removes a key from cache
func (r *RedisService) Delete(key string) error {
	return r.client.Del(r.ctx, key).Err()
}

// Exists checks if a key exists in cache
func (r *RedisService) Exists(key string) (bool, error) {
	result, err := r.client.Exists(r.ctx, key).Result()
	return result > 0, err
}

// Invalidate User Cache removes all cached data for a user
func (r *RedisService) InvalidateUserCache(userID string) error {
	keys := []string{
		fmt.Sprintf("user:%s", userID),
		fmt.Sprintf("matches:%s", userID),
		fmt.Sprintf("potential_matches:%s", userID),
	}

	for _, key := range keys {
		if err := r.client.Del(r.ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to delete key %s: %w", key, err)
		}
	}

	return nil
}

// Rate Limiting Methods

// IncrementRateLimit increments rate limit counter
func (r *RedisService) IncrementRateLimit(identifier string, window time.Duration) (int64, error) {
	key := fmt.Sprintf("rate_limit:%s", identifier)

	pipe := r.client.Pipeline()
	incr := pipe.Incr(r.ctx, key)
	pipe.Expire(r.ctx, key, window)

	_, err := pipe.Exec(r.ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to increment rate limit: %w", err)
	}

	return incr.Val(), nil
}

// GetRateLimit gets current rate limit count
func (r *RedisService) GetRateLimit(identifier string) (int64, error) {
	key := fmt.Sprintf("rate_limit:%s", identifier)
	val, err := r.client.Get(r.ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to get rate limit: %w", err)
	}

	count, err := strconv.ParseInt(val, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse rate limit value: %w", err)
	}

	return count, nil
}

// Health Check
func (r *RedisService) HealthCheck() error {
	_, err := r.client.Ping(r.ctx).Result()
	return err
}

// Close closes the Redis connection
func (r *RedisService) Close() error {
	return r.client.Close()
}

// Helper functions
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsIntOrDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
