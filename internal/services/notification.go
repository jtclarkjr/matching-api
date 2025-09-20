package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"matching-api/internal/models"
)

// NotificationService handles all notification operations
type NotificationService struct {
	fcmServerKey string
	fcmURL      string
	emailAPIKey string
	emailAPIURL string
}

// NewNotificationService creates a new notification service
func NewNotificationService() *NotificationService {
	return &NotificationService{
		fcmServerKey: os.Getenv("FCM_SERVER_KEY"),
		fcmURL:      "https://fcm.googleapis.com/fcm/send",
		emailAPIKey:  os.Getenv("EMAIL_API_KEY"),
		emailAPIURL:  os.Getenv("EMAIL_API_URL"),
	}
}

// SendPushNotification sends a push notification via FCM/APNS
func (ns *NotificationService) SendPushNotification(deviceTokens []string, payload models.PushNotificationPayload) error {
	if ns.fcmServerKey == "" {
		log.Println("FCM server key not configured, skipping push notification")
		return nil
	}

	for _, token := range deviceTokens {
		if err := ns.sendFCMNotification(token, payload); err != nil {
			log.Printf("Failed to send push notification to token %s: %v", token, err)
			// Continue with other tokens even if one fails
		}
	}

	return nil
}

// sendFCMNotification sends notification via Firebase Cloud Messaging
func (ns *NotificationService) sendFCMNotification(token string, payload models.PushNotificationPayload) error {
	fcmPayload := map[string]interface{}{
		"to": token,
		"notification": map[string]interface{}{
			"title": payload.Title,
			"body":  payload.Body,
			"sound": payload.Sound,
		},
		"data": payload.Data,
	}

	if payload.Badge > 0 {
		fcmPayload["notification"].(map[string]interface{})["badge"] = payload.Badge
	}

	jsonData, err := json.Marshal(fcmPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal FCM payload: %v", err)
	}

	req, err := http.NewRequest("POST", ns.fcmURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create FCM request: %v", err)
	}

	req.Header.Set("Authorization", "key="+ns.fcmServerKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("FCM request failed: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing FCM response body: %v", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("FCM request failed with status: %d", resp.StatusCode)
	}

	log.Printf("Push notification sent successfully to token: %s", token)
	return nil
}

// SendEmailNotification sends an email notification
func (ns *NotificationService) SendEmailNotification(email models.EmailNotification) error {
	if ns.emailAPIKey == "" || ns.emailAPIURL == "" {
		log.Println("Email API not configured, skipping email notification")
		return nil
	}

	// Example using a generic email API (like SendGrid, Mailgun, etc.)
	emailPayload := map[string]interface{}{
		"to":       email.To,
		"subject":  email.Subject,
		"template": email.TemplateID,
		"data":     email.TemplateData,
	}

	jsonData, err := json.Marshal(emailPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal email payload: %v", err)
	}

	req, err := http.NewRequest("POST", ns.emailAPIURL+"/send", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create email request: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+ns.emailAPIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email request: %v", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Printf("Error closing email response body: %v", err)
		}
	}()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("email request failed with status: %d", resp.StatusCode)
	}

	log.Printf("Email notification sent successfully to: %s", email.To)
	return nil
}

// CreateNotification creates a new notification record
func (ns *NotificationService) CreateNotification(userID string, notifType models.NotificationType, data map[string]interface{}) (*models.Notification, error) {
	title, message := models.GetNotificationTemplate(notifType, data)

	notification := &models.Notification{
		ID:        generateUUID(),
		UserID:    userID,
		Type:      notifType,
		Title:     title,
		Message:   message,
		Data:      data,
		IsRead:    false,
		IsSent:    false,
		CreatedAt: time.Now(),
	}

	// In a real app, save to database here
	// For now, simulate saving
	log.Printf("Created notification: %s for user: %s", notification.Title, userID)

	return notification, nil
}

// SendMatchNotification sends notification when users match
func (ns *NotificationService) SendMatchNotification(user1ID, user2ID string, user1Name, user2Name string) error {
	// Send notification to both users
	data1 := map[string]interface{}{
		"match_name": user2Name,
		"match_id":   generateUUID(),
	}
	
	data2 := map[string]interface{}{
		"match_name": user1Name,
		"match_id":   generateUUID(),
	}

	// Create notifications
	notif1, err := ns.CreateNotification(user1ID, models.NotificationNewMatch, data1)
	if err != nil {
		return err
	}

	notif2, err := ns.CreateNotification(user2ID, models.NotificationNewMatch, data2)
	if err != nil {
		return err
	}

	// Send push notifications (simulate getting device tokens)
	deviceTokens1 := ns.getUserDeviceTokens(user1ID)
	deviceTokens2 := ns.getUserDeviceTokens(user2ID)

	payload1 := models.PushNotificationPayload{
		Title: notif1.Title,
		Body:  notif1.Message,
		Data:  notif1.Data,
		Badge: 1,
		Sound: "default",
	}

	payload2 := models.PushNotificationPayload{
		Title: notif2.Title,
		Body:  notif2.Message,
		Data:  notif2.Data,
		Badge: 1,
		Sound: "default",
	}

	// Send push notifications
	go func() {
		if err := ns.SendPushNotification(deviceTokens1, payload1); err != nil {
			log.Printf("Error sending push notification to user1: %v", err)
		}
	}()
	go func() {
		if err := ns.SendPushNotification(deviceTokens2, payload2); err != nil {
			log.Printf("Error sending push notification to user2: %v", err)
		}
	}()

	return nil
}

// SendMessageNotification sends notification for new messages
func (ns *NotificationService) SendMessageNotification(recipientID, senderID, senderName, messageContent string) error {
	data := map[string]interface{}{
		"sender_id":   senderID,
		"sender_name": senderName,
		"message":     messageContent,
	}

	notification, err := ns.CreateNotification(recipientID, models.NotificationNewMessage, data)
	if err != nil {
		return err
	}

	// Send push notification
	deviceTokens := ns.getUserDeviceTokens(recipientID)
	payload := models.PushNotificationPayload{
		Title: notification.Title,
		Body:  notification.Message,
		Data:  notification.Data,
		Badge: 1,
		Sound: "default",
	}

	go func() {
		if err := ns.SendPushNotification(deviceTokens, payload); err != nil {
			log.Printf("Error sending message notification: %v", err)
		}
	}()

	return nil
}

// SendSuperLikeNotification sends notification for super likes
func (ns *NotificationService) SendSuperLikeNotification(recipientID, senderID, senderName string) error {
	data := map[string]interface{}{
		"sender_id":   senderID,
		"user_name":   senderName,
		"action_type": "super_like",
	}

	notification, err := ns.CreateNotification(recipientID, models.NotificationSuperLike, data)
	if err != nil {
		return err
	}

	// Send push notification
	deviceTokens := ns.getUserDeviceTokens(recipientID)
	payload := models.PushNotificationPayload{
		Title: notification.Title,
		Body:  notification.Message,
		Data:  notification.Data,
		Badge: 1,
		Sound: "default",
	}

	go func() {
		if err := ns.SendPushNotification(deviceTokens, payload); err != nil {
			log.Printf("Error sending super like push notification: %v", err)
		}
	}()

	// Send email if enabled for this user
	if ns.isEmailEnabledForUser(recipientID) {
		email := models.EmailNotification{
			To:         ns.getUserEmail(recipientID),
			Subject:    "Someone Super Liked You!",
			TemplateID: "super_like",
			TemplateData: map[string]interface{}{
				"sender_name": senderName,
			},
		}
		go func() {
			if err := ns.SendEmailNotification(email); err != nil {
				log.Printf("Error sending super like email notification: %v", err)
			}
		}()
	}

	return nil
}

// MarkNotificationAsRead marks a notification as read
func (ns *NotificationService) MarkNotificationAsRead(notificationID string) error {
	// In a real app, update database
	log.Printf("Marked notification as read: %s", notificationID)
	return nil
}

// GetUnreadCount gets count of unread notifications for user
func (ns *NotificationService) GetUnreadCount(userID string) (int, error) {
	// In a real app, query database
	// For now, simulate
	return 5, nil
}

// Helper functions (simulate database operations)

func (ns *NotificationService) getUserDeviceTokens(userID string) []string {
	// In a real app, query device tokens from database
	// For simulation, return mock tokens
	return []string{
		"mock-device-token-1",
		"mock-device-token-2",
	}
}

func (ns *NotificationService) isEmailEnabledForUser(userID string) bool {
	// In a real app, check user's notification preferences
	// For simulation, return true
	return true
}

func (ns *NotificationService) getUserEmail(userID string) string {
	// In a real app, get user's email from database
	// For simulation, return mock email
	return "user@example.com"
}

// generateUUID generates a simple UUID (in production, use proper UUID library)
func generateUUID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}