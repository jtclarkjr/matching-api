package middleware

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"matching-api/pkg/services"

	"gitea.com/go-chi/session"
	_ "gitea.com/go-chi/session/redis"
)

// Redis provider registration is handled automatically by the redis package

// SessionConfig holds session configuration
type SessionConfig struct {
	SessionName string
	MaxAge      int64
	Secure      bool
}

// NewSessionMiddleware creates a Chi session middleware with Redis support
func NewSessionMiddleware(redisService *services.RedisService) func(next http.Handler) http.Handler {
	config := SessionConfig{
		SessionName: getEnvOrDefault("SESSION_NAME", "session_name"),
		MaxAge:      getEnvAsInt64OrDefault("SESSION_MAX_AGE", 7*24*3600), // 7 days
		Secure:      getEnvAsBoolOrDefault("SESSION_SECURE", false),
	}

	var sessionOptions session.Options

	if redisService != nil {
		// Configure Redis-backed sessions
		redisHost := getEnvOrDefault("REDIS_HOST", "localhost")
		redisPort := getEnvOrDefault("REDIS_PORT", "6379")
		redisPassword := os.Getenv("REDIS_PASSWORD")
		redisDB := getEnvAsIntOrDefault("REDIS_DB", 0)

		sessionOptions = session.Options{
			Provider:       "redis",
			ProviderConfig: fmt.Sprintf("addr=%s:%s,password=%s,db=%d,pool_size=100,idle_timeout=180", redisHost, redisPort, redisPassword, redisDB),
			CookieName:     config.SessionName,
			CookiePath:     "/",
			Gclifetime:     config.MaxAge,
			Maxlifetime:    config.MaxAge,
			Secure:         config.Secure,
			CookieLifeTime: int(config.MaxAge),
			Domain:         getEnvOrDefault("SESSION_DOMAIN", ""),
		}

		// Redis provider is registered in init() function
	} else {
		// Fallback to memory-based sessions
		sessionOptions = session.Options{
			Provider:       "memory",
			CookieName:     config.SessionName,
			CookiePath:     "/",
			Gclifetime:     config.MaxAge,
			Maxlifetime:    config.MaxAge,
			Secure:         config.Secure,
			CookieLifeTime: int(config.MaxAge),
			Domain:         getEnvOrDefault("SESSION_DOMAIN", ""),
		}
	}

	return session.Sessioner(sessionOptions)
}

// SessionHelper provides helper functions for working with Chi sessions
type SessionHelper struct{}

// NewSessionHelper creates a new session helper
func NewSessionHelper() *SessionHelper {
	return &SessionHelper{}
}

// SetUserSession stores user information in session
func (h *SessionHelper) SetUserSession(store session.Store, userID, email string) error {
	if err := store.Set("user_id", userID); err != nil {
		return fmt.Errorf("failed to set user_id in session: %w", err)
	}
	if err := store.Set("email", email); err != nil {
		return fmt.Errorf("failed to set email in session: %w", err)
	}
	if err := store.Set("login_time", time.Now().Unix()); err != nil {
		return fmt.Errorf("failed to set login_time in session: %w", err)
	}
	return nil
}

// GetUserSession retrieves user information from session
func (h *SessionHelper) GetUserSession(store session.Store) (userID, email string, exists bool) {
	userIDRaw := store.Get("user_id")
	emailRaw := store.Get("email")

	if userIDRaw == nil || emailRaw == nil {
		return "", "", false
	}

	userID, userOk := userIDRaw.(string)
	email, emailOk := emailRaw.(string)

	if !userOk || !emailOk {
		return "", "", false
	}

	return userID, email, true
}

// ClearUserSession removes user information from session
func (h *SessionHelper) ClearUserSession(store session.Store) error {
	if err := store.Delete("user_id"); err != nil {
		return fmt.Errorf("failed to delete user_id from session: %w", err)
	}
	if err := store.Delete("email"); err != nil {
		return fmt.Errorf("failed to delete email from session: %w", err)
	}
	if err := store.Delete("login_time"); err != nil {
		return fmt.Errorf("failed to delete login_time from session: %w", err)
	}
	return store.Flush()
}

// IsLoggedIn checks if user is logged in
func (h *SessionHelper) IsLoggedIn(store session.Store) bool {
	userID := store.Get("user_id")
	return userID != nil
}

// GetLoginTime retrieves login time from session
func (h *SessionHelper) GetLoginTime(store session.Store) (time.Time, bool) {
	loginTimeRaw := store.Get("login_time")
	if loginTimeRaw == nil {
		return time.Time{}, false
	}

	loginTimestamp, ok := loginTimeRaw.(int64)
	if !ok {
		return time.Time{}, false
	}

	return time.Unix(loginTimestamp, 0), true
}

// UpdateSessionActivity updates last activity timestamp
func (h *SessionHelper) UpdateSessionActivity(store session.Store) error {
	if err := store.Set("last_activity", time.Now().Unix()); err != nil {
		return fmt.Errorf("failed to update session activity: %w", err)
	}
	return nil
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

func getEnvAsInt64OrDefault(key string, defaultValue int64) int64 {
	if value := os.Getenv(key); value != "" {
		if int64Value, err := strconv.ParseInt(value, 10, 64); err == nil {
			return int64Value
		}
	}
	return defaultValue
}

func getEnvAsBoolOrDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}
