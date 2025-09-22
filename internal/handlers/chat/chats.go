package chat

import (
	"net/http"

	"matching-api/internal/database"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// GetChats retrieves all chats for the authenticated user
func (h *Handler) GetChats(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// Query to get chats with match and user information
	query := `
		SELECT 
			c.id, c.match_id, c.last_message_at, c.created_at, c.updated_at,
			m.user1_id, m.user2_id,
			u1.first_name as user1_first_name, u1.last_name as user1_last_name,
			u2.first_name as user2_first_name, u2.last_name as user2_last_name,
			COALESCE(msg.id, '') as last_message_id,
			COALESCE(msg.content, '') as last_message_content,
			COALESCE(msg.message_type, '') as last_message_type,
			COALESCE(msg.created_at, '1970-01-01'::timestamp) as last_message_created_at
		FROM chats c
		JOIN matches m ON c.match_id = m.id
		LEFT JOIN users u1 ON m.user1_id = u1.id
		LEFT JOIN users u2 ON m.user2_id = u2.id
		LEFT JOIN messages msg ON msg.id = (
			SELECT id FROM messages 
			WHERE chat_id = c.id 
			ORDER BY created_at DESC 
			LIMIT 1
		)
		WHERE (m.user1_id = $1 OR m.user2_id = $1) AND m.is_active = true
		ORDER BY COALESCE(c.last_message_at, c.created_at) DESC
	`

	rows, err := database.DB.Query(query, user.UserID)
	if err != nil {
		utils.LogError("Error querying chats", err)
		utils.WriteInternalError(w, err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.LogError("Error closing rows", err)
		}
	}()

	var chats []models.Chat
	for rows.Next() {
		var chat models.Chat
		var match models.Match
		var user1FirstName, user1LastName, user2FirstName, user2LastName string
		var lastMessageID, lastMessageContent, lastMessageType string
		var lastMessageCreatedAt interface{}

		err := rows.Scan(
			&chat.ID, &chat.MatchID, &chat.LastMessageAt, &chat.CreatedAt, &chat.UpdatedAt,
			&match.User1ID, &match.User2ID,
			&user1FirstName, &user1LastName, &user2FirstName, &user2LastName,
			&lastMessageID, &lastMessageContent, &lastMessageType, &lastMessageCreatedAt,
		)
		if err != nil {
			utils.LogError("Error scanning chat row", err)
			continue
		}

		// Create user objects
		user1 := &models.User{
			ID:        match.User1ID,
			FirstName: user1FirstName,
			LastName:  user1LastName,
		}
		user2 := &models.User{
			ID:        match.User2ID,
			FirstName: user2FirstName,
			LastName:  user2LastName,
		}

		// Set match users
		match.User1 = user1
		match.User2 = user2

		// Add last message if exists
		if lastMessageID != "" {
			chat.LastMessage = &models.Message{
				ID:          lastMessageID,
				ChatID:      chat.ID,
				Content:     lastMessageContent,
				MessageType: lastMessageType,
			}
		}

		chats = append(chats, chat)
	}

	utils.WriteSuccessResponse(w, "Chats retrieved successfully", map[string]interface{}{
		"chats": chats,
		"total": len(chats),
	})
}
