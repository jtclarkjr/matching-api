package models

import (
	"encoding/json"
	"fmt"
	"time"
)

// NotificationType represents different types of notifications
type NotificationType string

const (
	// Match notifications
	NotificationNewMatch      NotificationType = "new_match"
	NotificationSuperLike     NotificationType = "super_like"
	NotificationMatchExpiring NotificationType = "match_expiring"

	// Message notifications
	NotificationNewMessage NotificationType = "new_message"
	NotificationMessageRead NotificationType = "message_read"

	// Profile notifications
	NotificationProfileViewed NotificationType = "profile_viewed"
	NotificationPhotoLiked    NotificationType = "photo_liked"

	// System notifications
	NotificationAccountVerified NotificationType = "account_verified"
	NotificationAccountWarning  NotificationType = "account_warning"
	NotificationPromotional     NotificationType = "promotional"
)

// Notification represents a notification record
type Notification struct {
	ID        string                 `json:"id" db:"id"`
	UserID    string                 `json:"user_id" db:"user_id"`
	Type      NotificationType       `json:"type" db:"type"`
	Title     string                 `json:"title" db:"title"`
	Message   string                 `json:"message" db:"message"`
	Data      map[string]interface{} `json:"data,omitempty" db:"data"`
	IsRead    bool                   `json:"is_read" db:"is_read"`
	IsSent    bool                   `json:"is_sent" db:"is_sent"`
	CreatedAt time.Time              `json:"created_at" db:"created_at"`
}

// PushNotificationPayload represents FCM/APNS payload
type PushNotificationPayload struct {
	Title string                 `json:"title"`
	Body  string                 `json:"body"`
	Data  map[string]interface{} `json:"data,omitempty"`
	Badge int                    `json:"badge,omitempty"`
	Sound string                 `json:"sound,omitempty"`
}

// DeviceToken represents user's device token for push notifications
type DeviceToken struct {
	ID       string    `json:"id" db:"id"`
	UserID   string    `json:"user_id" db:"user_id"`
	Token    string    `json:"token" db:"token"`
	Platform string    `json:"platform" db:"platform"` // ios, android
	IsActive bool      `json:"is_active" db:"is_active"`
	LastUsed time.Time `json:"last_used" db:"last_used"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// EmailNotification represents email notification data
type EmailNotification struct {
	To          string                 `json:"to"`
	Subject     string                 `json:"subject"`
	TemplateID  string                 `json:"template_id"`
	TemplateData map[string]interface{} `json:"template_data"`
}

// NotificationPreferences represents user's notification preferences
type NotificationPreferences struct {
	ID                string `json:"id" db:"id"`
	UserID            string `json:"user_id" db:"user_id"`
	PushEnabled       bool   `json:"push_enabled" db:"push_enabled"`
	EmailEnabled      bool   `json:"email_enabled" db:"email_enabled"`
	MatchNotifications bool  `json:"match_notifications" db:"match_notifications"`
	MessageNotifications bool `json:"message_notifications" db:"message_notifications"`
	ProfileNotifications bool `json:"profile_notifications" db:"profile_notifications"`
	MarketingEmails   bool   `json:"marketing_emails" db:"marketing_emails"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// CreateNotificationRequest represents request to create notification
type CreateNotificationRequest struct {
	UserID  string                 `json:"user_id" validate:"required"`
	Type    NotificationType       `json:"type" validate:"required"`
	Title   string                 `json:"title" validate:"required,max=255"`
	Message string                 `json:"message" validate:"required,max=500"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

// UpdateNotificationPreferencesRequest represents notification preferences update
type UpdateNotificationPreferencesRequest struct {
	PushEnabled           *bool `json:"push_enabled,omitempty"`
	EmailEnabled          *bool `json:"email_enabled,omitempty"`
	MatchNotifications    *bool `json:"match_notifications,omitempty"`
	MessageNotifications  *bool `json:"message_notifications,omitempty"`
	ProfileNotifications  *bool `json:"profile_notifications,omitempty"`
	MarketingEmails       *bool `json:"marketing_emails,omitempty"`
}

// RegisterDeviceRequest represents device token registration
type RegisterDeviceRequest struct {
	Token    string `json:"token" validate:"required"`
	Platform string `json:"platform" validate:"required,oneof=ios android web"`
}

// NotificationStats represents notification statistics
type NotificationStats struct {
	TotalSent       int64 `json:"total_sent"`
	TotalDelivered  int64 `json:"total_delivered"`
	TotalRead       int64 `json:"total_read"`
	DeliveryRate    float64 `json:"delivery_rate"`
	ReadRate        float64 `json:"read_rate"`
}

// GetNotificationTemplate returns notification template based on type and data
func GetNotificationTemplate(notifType NotificationType, data map[string]interface{}) (string, string) {
	switch notifType {
	case NotificationNewMatch:
		if name, ok := data["match_name"].(string); ok {
			return "New Match!", fmt.Sprintf("You have a new match with %s! Start chatting now.", name)
		}
		return "New Match!", "You have a new match! Start chatting now."
	
	case NotificationSuperLike:
		if name, ok := data["user_name"].(string); ok {
			return "Super Like!", fmt.Sprintf("%s super liked you! Check them out.", name)
		}
		return "Super Like!", "Someone super liked you! Check them out."
	
	case NotificationNewMessage:
		if name, ok := data["sender_name"].(string); ok {
			return "New Message", fmt.Sprintf("%s sent you a message", name)
		}
		return "New Message", "You have a new message"
	
	case NotificationProfileViewed:
		return "Profile View", "Someone viewed your profile!"
	
	case NotificationPhotoLiked:
		return "Photo Liked", "Someone liked your photo!"
	
	case NotificationAccountVerified:
		return "Account Verified", "Your account has been verified! You'll now see a blue checkmark on your profile."
	
	case NotificationMatchExpiring:
		if name, ok := data["match_name"].(string); ok {
			return "Match Expiring Soon", fmt.Sprintf("Your match with %s expires in 24 hours. Send a message!", name)
		}
		return "Match Expiring Soon", "One of your matches expires soon. Send a message!"
	
	default:
		return "Notification", "You have a new notification"
	}
}

// MarshalData marshals notification data to JSON
func (n *Notification) MarshalData() ([]byte, error) {
	if n.Data == nil {
		return nil, nil
	}
	return json.Marshal(n.Data)
}

// UnmarshalData unmarshals JSON data to notification data
func (n *Notification) UnmarshalData(data []byte) error {
	if data == nil {
		return nil
	}
	return json.Unmarshal(data, &n.Data)
}