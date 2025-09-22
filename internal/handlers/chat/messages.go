package chat

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"matching-api/internal/database"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// GetMessages retrieves messages for a specific chat
func (h *Handler) GetMessages(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get chat ID from URL parameter
	chatID := chi.URLParam(r, "chatID")
	if chatID == "" {
		utils.WriteErrorResponse(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// First verify that the user has access to this chat
	var matchExists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM chats c
			JOIN matches m ON c.match_id = m.id
			WHERE c.id = $1 AND (m.user1_id = $2 OR m.user2_id = $2) AND m.is_active = true
		)
	`
	err := database.DB.QueryRow(checkQuery, chatID, user.UserID).Scan(&matchExists)
	if err != nil {
		utils.LogError("Error checking chat access", err)
		utils.WriteInternalError(w, err)
		return
	}

	if !matchExists {
		utils.WriteErrorResponse(w, "Chat not found or access denied", http.StatusNotFound)
		return
	}

	// Get pagination parameters
	limit := 50 // default limit
	offset := 0
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}
	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Query messages
	query := `
		SELECT 
			m.id, m.chat_id, m.sender_id, m.content, m.message_type, m.is_read, m.created_at,
			u.first_name, u.last_name
		FROM messages m
		JOIN users u ON m.sender_id = u.id
		WHERE m.chat_id = $1
		ORDER BY m.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := database.DB.Query(query, chatID, limit, offset)
	if err != nil {
		utils.LogError("Error querying messages", err)
		utils.WriteInternalError(w, err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.LogError("Error closing rows", err)
		}
	}()

	var messages []models.Message
	for rows.Next() {
		var message models.Message
		var senderFirstName, senderLastName string

		err := rows.Scan(
			&message.ID, &message.ChatID, &message.SenderID, &message.Content,
			&message.MessageType, &message.IsRead, &message.CreatedAt,
			&senderFirstName, &senderLastName,
		)
		if err != nil {
			utils.LogError("Error scanning message row", err)
			continue
		}

		messages = append(messages, message)
	}

	// Mark messages as read for the current user
	markReadQuery := `
		UPDATE messages 
		SET is_read = true 
		WHERE chat_id = $1 AND sender_id != $2 AND is_read = false
	`
	_, err = database.DB.Exec(markReadQuery, chatID, user.UserID)
	if err != nil {
		utils.LogError("Error marking messages as read", err)
	}

	utils.WriteSuccessResponse(w, "Messages retrieved successfully", map[string]interface{}{
		"messages": messages,
		"total":    len(messages),
		"limit":    limit,
		"offset":   offset,
	})
}

// SendMessage sends a new message in a chat
func (h *Handler) SendMessage(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get chat ID from URL parameter
	chatID := chi.URLParam(r, "chatID")
	if chatID == "" {
		utils.WriteErrorResponse(w, "Chat ID is required", http.StatusBadRequest)
		return
	}

	// Parse and validate request body
	var req models.MessageRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Set default message type
	if req.MessageType == "" {
		req.MessageType = "text"
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// First verify that the user has access to this chat
	var matchExists bool
	checkQuery := `
		SELECT EXISTS(
			SELECT 1 FROM chats c
			JOIN matches m ON c.match_id = m.id
			WHERE c.id = $1 AND (m.user1_id = $2 OR m.user2_id = $2) AND m.is_active = true
		)
	`
	err := database.DB.QueryRow(checkQuery, chatID, user.UserID).Scan(&matchExists)
	if err != nil {
		utils.LogError("Error checking chat access", err)
		utils.WriteInternalError(w, err)
		return
	}

	if !matchExists {
		utils.WriteErrorResponse(w, "Chat not found or access denied", http.StatusNotFound)
		return
	}

	// Insert the message
	var messageID string
	insertQuery := `
		INSERT INTO messages (id, chat_id, sender_id, content, message_type, is_read, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, false, NOW())
		RETURNING id, created_at
	`

	var createdAt interface{}
	err = database.DB.QueryRow(insertQuery, chatID, user.UserID, req.Content, req.MessageType).Scan(&messageID, &createdAt)
	if err != nil {
		utils.LogError("Error inserting message", err)
		utils.WriteInternalError(w, err)
		return
	}

	// Update chat's last_message_at
	updateChatQuery := `
		UPDATE chats 
		SET last_message_at = NOW(), updated_at = NOW()
		WHERE id = $1
	`
	_, err = database.DB.Exec(updateChatQuery, chatID)
	if err != nil {
		utils.LogError("Error updating chat timestamp", err)
	}

	// Create response message
	message := models.Message{
		ID:          messageID,
		ChatID:      chatID,
		SenderID:    user.UserID,
		Content:     req.Content,
		MessageType: req.MessageType,
		IsRead:      false,
	}

	// Broadcast message to WebSocket connections
	BroadcastMessage(chatID, &message)

	utils.WriteCreated(w, "Message sent successfully", map[string]interface{}{
		"message": message,
	})
}
