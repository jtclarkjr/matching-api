package models

import (
	"time"
)

// SwipeAction represents the direction of a swipe
type SwipeAction string

const (
	SwipeLeft  SwipeAction = "left"  // Pass/Dislike
	SwipeRight SwipeAction = "right" // Like
	SuperLike  SwipeAction = "super" // Super like
)

// Swipe represents a user's swipe on another user
type Swipe struct {
	ID         string      `json:"id" db:"id"`
	UserID     string      `json:"user_id" db:"user_id"`
	TargetID   string      `json:"target_id" db:"target_id"`
	Action     SwipeAction `json:"action" db:"action"`
	CreatedAt  time.Time   `json:"created_at" db:"created_at"`
}

// Match represents a mutual like between two users
type Match struct {
	ID        string    `json:"id" db:"id"`
	User1ID   string    `json:"user1_id" db:"user1_id"`
	User2ID   string    `json:"user2_id" db:"user2_id"`
	User1     *User     `json:"user1,omitempty"`
	User2     *User     `json:"user2,omitempty"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}

// Chat represents a conversation between matched users
type Chat struct {
	ID            string     `json:"id" db:"id"`
	MatchID       string     `json:"match_id" db:"match_id"`
	LastMessage   *Message   `json:"last_message,omitempty"`
	LastMessageAt *time.Time `json:"last_message_at,omitempty" db:"last_message_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`
}

// Message represents a chat message
type Message struct {
	ID        string    `json:"id" db:"id"`
	ChatID    string    `json:"chat_id" db:"chat_id"`
	SenderID  string    `json:"sender_id" db:"sender_id"`
	Content   string    `json:"content" db:"content"`
	MessageType string  `json:"message_type" db:"message_type"` // text, image, gif
	IsRead    bool      `json:"is_read" db:"is_read"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// SwipeRequest represents the request body for swiping
type SwipeRequest struct {
	TargetID string      `json:"target_id" validate:"required"`
	Action   SwipeAction `json:"action" validate:"required,oneof=left right super"`
}

// MessageRequest represents the request body for sending a message
type MessageRequest struct {
	Content     string `json:"content" validate:"required,max=500"`
	MessageType string `json:"message_type,omitempty" validate:"oneof=text image gif"`
}

// GetOtherUser returns the other user in a match (not the current user)
func (m *Match) GetOtherUser(currentUserID string) *User {
	if m.User1ID == currentUserID {
		return m.User2
	}
	return m.User1
}