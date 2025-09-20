package services

import (
	"log"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"matching-api/internal/models"
)

// AnalyticsService handles analytics tracking and reporting
type AnalyticsService struct {
	// In production, inject database and cache services
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService() *AnalyticsService {
	return &AnalyticsService{}
}

// TrackEvent tracks a user event
func (as *AnalyticsService) TrackEvent(userID *string, eventType models.EventType, eventData map[string]interface{}, r *http.Request) error {
	event := &models.AnalyticsEvent{
		ID:        uuid.New().String(),
		UserID:    userID,
		EventType: eventType,
		EventData: eventData,
		CreatedAt: time.Now(),
	}

	// Extract session ID from headers or context
	if sessionID := r.Header.Get("X-Session-ID"); sessionID != "" {
		event.SessionID = &sessionID
	}

	// Get IP address
	if ip := getClientIP(r); ip != "" {
		event.IPAddress = &ip
	}

	// Get user agent
	if userAgent := r.UserAgent(); userAgent != "" {
		event.UserAgent = &userAgent
	}

	// In production, save to database
	log.Printf("Analytics Event: %s - %s (User: %v)", eventType, models.GetEventDescription(eventType), userID)
	
	return nil
}

// TrackUserLogin tracks user login event
func (as *AnalyticsService) TrackUserLogin(userID string, r *http.Request) error {
	data := map[string]interface{}{
		"method": "jwt", // could be "google", "facebook", etc.
	}
	return as.TrackEvent(&userID, models.EventUserLogin, data, r)
}

// TrackUserRegistration tracks user registration
func (as *AnalyticsService) TrackUserRegistration(userID string, registrationMethod string, r *http.Request) error {
	data := map[string]interface{}{
		"method": registrationMethod,
	}
	return as.TrackEvent(&userID, models.EventUserRegistered, data, r)
}

// TrackSwipe tracks swipe events
func (as *AnalyticsService) TrackSwipe(userID, targetID string, action models.SwipeAction, r *http.Request) error {
	var eventType models.EventType
	switch action {
	case "left":
		eventType = models.EventSwipeLeft
	case "right":
		eventType = models.EventSwipeRight
	case "super":
		eventType = models.EventSuperLike
	}

	data := map[string]interface{}{
		"target_user_id": targetID,
		"swipe_action":   string(action),
	}
	return as.TrackEvent(&userID, eventType, data, r)
}

// TrackMatch tracks match creation
func (as *AnalyticsService) TrackMatch(user1ID, user2ID string, r *http.Request) error {
	// Track for both users
	data1 := map[string]interface{}{
		"matched_user_id": user2ID,
	}
	data2 := map[string]interface{}{
		"matched_user_id": user1ID,
	}

	if err := as.TrackEvent(&user1ID, models.EventMatchCreated, data1, r); err != nil {
		log.Printf("Error tracking match created event for user1: %v", err)
	}
	if err := as.TrackEvent(&user2ID, models.EventMatchCreated, data2, r); err != nil {
		log.Printf("Error tracking match created event for user2: %v", err)
	}
	
	return nil
}

// TrackMessage tracks message events
func (as *AnalyticsService) TrackMessage(senderID, recipientID, chatID string, messageType string, r *http.Request) error {
	data := map[string]interface{}{
		"recipient_id":   recipientID,
		"chat_id":        chatID,
		"message_type":   messageType,
	}
	return as.TrackEvent(&senderID, models.EventMessageSent, data, r)
}

// GetUserMetrics returns metrics for a specific user
func (as *AnalyticsService) GetUserMetrics(userID string) (*models.UserMetrics, error) {
	// In production, aggregate from database
	return &models.UserMetrics{
		UserID:              userID,
		TotalSwipes:         450,
		RightSwipes:         180,
		LeftSwipes:          270,
		SuperLikes:          15,
		TotalMatches:        45,
		ActiveMatches:       12,
		MessagesSent:        250,
		MessagesReceived:    280,
		ProfileViews:        1200,
		PhotosUploaded:      6,
		SessionCount:        89,
		TotalTimeSpent:      152400, // 42.33 hours
		LastActiveAt:        &[]time.Time{time.Now().Add(-2 * time.Hour)}[0],
		RegistrationDate:    time.Now().AddDate(0, -2, -15), // 2.5 months ago
		SwipeRightRate:      0.4,  // 40% right swipe rate
		MatchRate:           0.25, // 25% of right swipes result in matches
		MessageResponseRate: 0.78, // 78% message response rate
	}, nil
}

// GetAppMetrics returns overall app metrics
func (as *AnalyticsService) GetAppMetrics() (*models.AppMetrics, error) {
	// In production, aggregate from database and cache
	return &models.AppMetrics{
		TotalUsers:          125000,
		ActiveUsers24h:      15000,
		ActiveUsers7d:       45000,
		ActiveUsers30d:      85000,
		NewRegistrations24h: 250,
		TotalSwipes24h:      75000,
		TotalMatches24h:     3500,
		TotalMessages24h:    12000,
		AvgSessionLength:    420, // 7 minutes
		UserRetention: models.UserRetentionMetrics{
			Day1Retention:  0.65, // 65%
			Day7Retention:  0.35, // 35%
			Day30Retention: 0.18, // 18%
		},
		ConversionFunnel: models.ConversionFunnelMetrics{
			Registrations:   1000,
			ProfileComplete: 850, // 85%
			FirstSwipe:      720, // 72%
			FirstMatch:      540, // 54%
			FirstMessage:    320, // 32%
		},
	}, nil
}

// GetDashboardMetrics returns comprehensive dashboard metrics
func (as *AnalyticsService) GetDashboardMetrics() (*models.DashboardMetrics, error) {
	// TODO: Implement real data aggregation from database
	// This should pull from actual analytics tables and compute real-time metrics
	overview, _ := as.GetAppMetrics()
	
	return &models.DashboardMetrics{
		Overview: *overview,
		Engagement: models.EngagementMetrics{
			AvgSwipesPerUser:      45.5,
			AvgMatchesPerUser:     8.2,
			AvgMessagesPerUser:    25.7,
			AvgSessionsPerUser:    12.3,
			SwipeToMatchRate:      0.08,  // 8% swipes result in matches
			MatchToMessageRate:    0.65,  // 65% matches get messages
			MessageResponseRate:   0.72,  // 72% response rate
			ProfileCompletionRate: 0.85,  // 85% complete profiles
		},
		Revenue: models.RevenueMetrics{
			TotalRevenue24h:     12500.00,
			TotalRevenue30d:     385000.00,
			SubscriptionRevenue: 280000.00,
			PurchaseRevenue:     105000.00,
			ARPU:               4.52, // $4.52 per user
			ConversionRate:     0.045, // 4.5% convert to paid
			ChurnRate:          0.06,  // 6% monthly churn
		},
		Geographic: models.GeographicMetrics{
			TopCountries: []models.CountryMetric{
				{Country: "United States", UserCount: 45000, Percentage: 36.0},
				{Country: "Canada", UserCount: 18000, Percentage: 14.4},
				{Country: "United Kingdom", UserCount: 15000, Percentage: 12.0},
				{Country: "Australia", UserCount: 12000, Percentage: 9.6},
				{Country: "Germany", UserCount: 8000, Percentage: 6.4},
			},
			TopCities: []models.CityMetric{
				{City: "New York", UserCount: 8500, Percentage: 6.8},
				{City: "Los Angeles", UserCount: 7200, Percentage: 5.76},
				{City: "Toronto", UserCount: 6800, Percentage: 5.44},
				{City: "London", UserCount: 6200, Percentage: 4.96},
				{City: "Sydney", UserCount: 4800, Percentage: 3.84},
			},
		},
		Demographic: models.DemographicMetrics{
			AgeDistribution: []models.AgeGroup{
				{AgeRange: "18-24", UserCount: 35000, Percentage: 28.0},
				{AgeRange: "25-34", UserCount: 52000, Percentage: 41.6},
				{AgeRange: "35-44", UserCount: 25000, Percentage: 20.0},
				{AgeRange: "45-54", UserCount: 10000, Percentage: 8.0},
				{AgeRange: "55+", UserCount: 3000, Percentage: 2.4},
			},
			GenderDistribution: []models.GenderGroup{
				{Gender: "male", UserCount: 72000, Percentage: 57.6},
				{Gender: "female", UserCount: 50000, Percentage: 40.0},
				{Gender: "non-binary", UserCount: 3000, Percentage: 2.4},
			},
		},
		Trending: models.TrendingMetrics{
			UserGrowth: generateTrendingData(30, 1000, 1500), // 30 days, 1000-1500 new users/day
			Engagement: generateTrendingData(30, 0.35, 0.45), // 30 days, engagement rate
			Revenue: generateTrendingData(30, 10000, 15000),  // 30 days, revenue
			Retention: generateTrendingData(30, 0.15, 0.25),  // 30 days, retention rate
		},
	}, nil
}

// GetFunnelAnalysis returns conversion funnel analysis
func (as *AnalyticsService) GetFunnelAnalysis() (*models.FunnelAnalysis, error) {
	return &models.FunnelAnalysis{
		Steps: []models.FunnelStep{
			{StepName: "App Downloaded", UserCount: 10000, ConversionRate: 1.0, DropOffRate: 0.0},
			{StepName: "Registration Started", UserCount: 8500, ConversionRate: 0.85, DropOffRate: 0.15},
			{StepName: "Profile Created", UserCount: 7200, ConversionRate: 0.85, DropOffRate: 0.15},
			{StepName: "Photo Uploaded", UserCount: 6500, ConversionRate: 0.90, DropOffRate: 0.10},
			{StepName: "First Swipe", UserCount: 5800, ConversionRate: 0.89, DropOffRate: 0.11},
			{StepName: "First Match", UserCount: 4200, ConversionRate: 0.72, DropOffRate: 0.28},
			{StepName: "First Message", UserCount: 2800, ConversionRate: 0.67, DropOffRate: 0.33},
			{StepName: "Active User (7+ days)", UserCount: 1900, ConversionRate: 0.68, DropOffRate: 0.32},
		},
	}, nil
}

// GetEventSummary returns event summary for a time period
func (as *AnalyticsService) GetEventSummary(days int) ([]models.EventSummary, error) {
	// In production, aggregate from database
	return []models.EventSummary{
		{EventType: models.EventSwipeRight, Count: 45000, UniqueUsers: 8500},
		{EventType: models.EventSwipeLeft, Count: 67500, UniqueUsers: 8500},
		{EventType: models.EventMatchCreated, Count: 3500, UniqueUsers: 2800},
		{EventType: models.EventMessageSent, Count: 12000, UniqueUsers: 1800},
		{EventType: models.EventProfileViewed, Count: 25000, UniqueUsers: 6200},
		{EventType: models.EventUserLogin, Count: 28000, UniqueUsers: 15000},
	}, nil
}

// Helper functions

func getClientIP(r *http.Request) string {
	// Check for X-Forwarded-For header (load balancers, proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		return xff
	}
	
	// Check for X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}
	
	// Fall back to remote address
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}

func generateTrendingData(days int, min, max float64) []models.DataPoint {
	var data []models.DataPoint
	now := time.Now().Truncate(24 * time.Hour)
	
	for i := days; i > 0; i-- {
		date := now.AddDate(0, 0, -i)
		// Generate mock trending data within the range
		value := min + (max-min)*0.5 + (max-min)*0.3*(float64(i%7)/7.0) // Weekly pattern
		data = append(data, models.DataPoint{
			Date:  date,
			Value: value,
		})
	}
	
	return data
}