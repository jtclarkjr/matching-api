package models

import (
	"time"
)

// EventType represents different types of analytics events
type EventType string

const (
	// User engagement events
	EventUserLogin       EventType = "user_login"
	EventUserLogout      EventType = "user_logout"
	EventUserRegistered  EventType = "user_registered"
	EventProfileViewed   EventType = "profile_viewed"
	EventProfileUpdated  EventType = "profile_updated"
	EventPhotoUploaded   EventType = "photo_uploaded"
	EventPhotoDeleted    EventType = "photo_deleted"

	// Matching events
	EventSwipeLeft      EventType = "swipe_left"
	EventSwipeRight     EventType = "swipe_right"
	EventSuperLike      EventType = "super_like"
	EventMatchCreated   EventType = "match_created"
	EventMatchDeleted   EventType = "match_deleted"
	EventPotentialViewd EventType = "potential_viewed"

	// Messaging events
	EventMessageSent     EventType = "message_sent"
	EventMessageReceived EventType = "message_received"
	EventMessageRead     EventType = "message_read"
	EventChatOpened      EventType = "chat_opened"
	EventTypingStarted   EventType = "typing_started"

	// App usage events
	EventAppOpened     EventType = "app_opened"
	EventAppClosed     EventType = "app_closed"
	EventScreenViewed  EventType = "screen_viewed"
	EventFeatureUsed   EventType = "feature_used"
	EventSearchPerformed EventType = "search_performed"

	// Purchase events
	EventSubscriptionPurchased EventType = "subscription_purchased"
	EventBoostPurchased       EventType = "boost_purchased"
	EventSuperLikePurchased   EventType = "super_like_purchased"

	// Error events
	EventError     EventType = "error"
	EventException EventType = "exception"
)

// AnalyticsEvent represents an analytics event
type AnalyticsEvent struct {
	ID        string                 `json:"id" db:"id"`
	UserID    *string                `json:"user_id,omitempty" db:"user_id"`
	EventType EventType              `json:"event_type" db:"event_type"`
	EventData map[string]interface{} `json:"event_data,omitempty" db:"event_data"`
	SessionID *string                `json:"session_id,omitempty" db:"session_id"`
	IPAddress *string                `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent *string                `json:"user_agent,omitempty" db:"user_agent"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// UserMetrics represents aggregated user metrics
type UserMetrics struct {
	UserID              string    `json:"user_id"`
	TotalSwipes         int64     `json:"total_swipes"`
	RightSwipes         int64     `json:"right_swipes"`
	LeftSwipes          int64     `json:"left_swipes"`
	SuperLikes          int64     `json:"super_likes"`
	TotalMatches        int64     `json:"total_matches"`
	ActiveMatches       int64     `json:"active_matches"`
	MessagesSent        int64     `json:"messages_sent"`
	MessagesReceived    int64     `json:"messages_received"`
	ProfileViews        int64     `json:"profile_views"`
	PhotosUploaded      int64     `json:"photos_uploaded"`
	SessionCount        int64     `json:"session_count"`
	TotalTimeSpent      int64     `json:"total_time_spent"` // in seconds
	LastActiveAt        *time.Time `json:"last_active_at"`
	RegistrationDate    time.Time  `json:"registration_date"`
	SwipeRightRate      float64    `json:"swipe_right_rate"`
	MatchRate           float64    `json:"match_rate"`
	MessageResponseRate float64    `json:"message_response_rate"`
}

// AppMetrics represents overall app metrics
type AppMetrics struct {
	TotalUsers          int64   `json:"total_users"`
	ActiveUsers24h      int64   `json:"active_users_24h"`
	ActiveUsers7d       int64   `json:"active_users_7d"`
	ActiveUsers30d      int64   `json:"active_users_30d"`
	NewRegistrations24h int64   `json:"new_registrations_24h"`
	TotalSwipes24h      int64   `json:"total_swipes_24h"`
	TotalMatches24h     int64   `json:"total_matches_24h"`
	TotalMessages24h    int64   `json:"total_messages_24h"`
	AvgSessionLength    float64 `json:"avg_session_length"`
	UserRetention       UserRetentionMetrics `json:"user_retention"`
	ConversionFunnel    ConversionFunnelMetrics `json:"conversion_funnel"`
}

// UserRetentionMetrics represents user retention statistics
type UserRetentionMetrics struct {
	Day1Retention  float64 `json:"day1_retention"`
	Day7Retention  float64 `json:"day7_retention"`
	Day30Retention float64 `json:"day30_retention"`
}

// ConversionFunnelMetrics represents user conversion funnel
type ConversionFunnelMetrics struct {
	Registrations    int64 `json:"registrations"`
	ProfileComplete  int64 `json:"profile_complete"`
	FirstSwipe       int64 `json:"first_swipe"`
	FirstMatch       int64 `json:"first_match"`
	FirstMessage     int64 `json:"first_message"`
}

// TrackEventRequest represents request to track an event
type TrackEventRequest struct {
	EventType EventType              `json:"event_type" validate:"required"`
	EventData map[string]interface{} `json:"event_data,omitempty"`
	SessionID *string                `json:"session_id,omitempty"`
}

// DashboardMetrics represents metrics for admin dashboard
type DashboardMetrics struct {
	Overview    AppMetrics              `json:"overview"`
	Engagement  EngagementMetrics       `json:"engagement"`
	Revenue     RevenueMetrics          `json:"revenue"`
	Geographic  GeographicMetrics       `json:"geographic"`
	Demographic DemographicMetrics      `json:"demographic"`
	Trending    TrendingMetrics         `json:"trending"`
}

// EngagementMetrics represents user engagement statistics
type EngagementMetrics struct {
	AvgSwipesPerUser     float64 `json:"avg_swipes_per_user"`
	AvgMatchesPerUser    float64 `json:"avg_matches_per_user"`
	AvgMessagesPerUser   float64 `json:"avg_messages_per_user"`
	AvgSessionsPerUser   float64 `json:"avg_sessions_per_user"`
	SwipeToMatchRate     float64 `json:"swipe_to_match_rate"`
	MatchToMessageRate   float64 `json:"match_to_message_rate"`
	MessageResponseRate  float64 `json:"message_response_rate"`
	ProfileCompletionRate float64 `json:"profile_completion_rate"`
}

// RevenueMetrics represents revenue and monetization metrics
type RevenueMetrics struct {
	TotalRevenue24h      float64 `json:"total_revenue_24h"`
	TotalRevenue30d      float64 `json:"total_revenue_30d"`
	SubscriptionRevenue  float64 `json:"subscription_revenue"`
	PurchaseRevenue      float64 `json:"purchase_revenue"`
	ARPU                 float64 `json:"arpu"` // Average Revenue Per User
	ConversionRate       float64 `json:"conversion_rate"`
	ChurnRate           float64 `json:"churn_rate"`
}

// GeographicMetrics represents geographic distribution
type GeographicMetrics struct {
	TopCountries []CountryMetric `json:"top_countries"`
	TopCities    []CityMetric    `json:"top_cities"`
}

type CountryMetric struct {
	Country    string `json:"country"`
	UserCount  int64  `json:"user_count"`
	Percentage float64 `json:"percentage"`
}

type CityMetric struct {
	City       string `json:"city"`
	UserCount  int64  `json:"user_count"`
	Percentage float64 `json:"percentage"`
}

// DemographicMetrics represents user demographic breakdown
type DemographicMetrics struct {
	AgeDistribution    []AgeGroup    `json:"age_distribution"`
	GenderDistribution []GenderGroup `json:"gender_distribution"`
}

type AgeGroup struct {
	AgeRange   string  `json:"age_range"`
	UserCount  int64   `json:"user_count"`
	Percentage float64 `json:"percentage"`
}

type GenderGroup struct {
	Gender     string  `json:"gender"`
	UserCount  int64   `json:"user_count"`
	Percentage float64 `json:"percentage"`
}

// TrendingMetrics represents trending data over time
type TrendingMetrics struct {
	UserGrowth   []DataPoint `json:"user_growth"`
	Engagement   []DataPoint `json:"engagement"`
	Revenue      []DataPoint `json:"revenue"`
	Retention    []DataPoint `json:"retention"`
}

type DataPoint struct {
	Date  time.Time `json:"date"`
	Value float64   `json:"value"`
}

// FunnelAnalysis represents funnel analysis data
type FunnelAnalysis struct {
	Steps []FunnelStep `json:"steps"`
}

type FunnelStep struct {
	StepName       string  `json:"step_name"`
	UserCount      int64   `json:"user_count"`
	ConversionRate float64 `json:"conversion_rate"`
	DropOffRate    float64 `json:"drop_off_rate"`
}

// CohortAnalysis represents user cohort analysis
type CohortAnalysis struct {
	CohortSize int64       `json:"cohort_size"`
	Periods    []CohortPeriod `json:"periods"`
}

type CohortPeriod struct {
	Period         int     `json:"period"`
	ActiveUsers    int64   `json:"active_users"`
	RetentionRate  float64 `json:"retention_rate"`
}

// EventSummary represents event summary for a time period
type EventSummary struct {
	EventType EventType `json:"event_type"`
	Count     int64     `json:"count"`
	UniqueUsers int64   `json:"unique_users"`
}

// GetEventDescription returns a human-readable description of an event type
func GetEventDescription(eventType EventType) string {
	descriptions := map[EventType]string{
		EventUserLogin:         "User logged in",
		EventUserLogout:        "User logged out",
		EventUserRegistered:    "User registered",
		EventProfileViewed:     "Profile viewed",
		EventSwipeLeft:         "Swiped left (pass)",
		EventSwipeRight:        "Swiped right (like)",
		EventSuperLike:         "Super liked",
		EventMatchCreated:      "New match created",
		EventMessageSent:       "Message sent",
		EventMessageReceived:   "Message received",
		EventAppOpened:         "App opened",
		EventSubscriptionPurchased: "Subscription purchased",
	}
	
	if desc, exists := descriptions[eventType]; exists {
		return desc
	}
	return string(eventType)
}