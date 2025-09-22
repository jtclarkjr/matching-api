package notification

import (
	"encoding/json"
	"net/http"
	"strconv"

	"matching-api/internal/database"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// GetNotifications retrieves all notifications for the authenticated user
// @Summary Get user notifications
// @Description Retrieve user's notifications with pagination and filtering
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Number of notifications per page (max 50)" default(20)
// @Param type query string false "Filter by notification type"
// @Success 200 {object} models.APIResponse{data=object{notifications=[]models.Notification,page=int,limit=int,total=int}} "Notifications retrieved successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request - invalid pagination"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications [get]
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
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

	// Filter parameters
	unreadOnly := r.URL.Query().Get("unread_only") == "true"
	notificationType := r.URL.Query().Get("type")

	// Build query
	query := `
		SELECT id, user_id, type, title, message, data, is_read, is_sent, created_at
		FROM notifications
		WHERE user_id = $1
	`
	args := []interface{}{user.UserID}
	argIndex := 2

	if unreadOnly {
		query += " AND is_read = false"
	}

	if notificationType != "" {
		query += " AND type = $" + strconv.Itoa(argIndex)
		args = append(args, notificationType)
		argIndex++
	}

	query += " ORDER BY created_at DESC LIMIT $" + strconv.Itoa(argIndex) + " OFFSET $" + strconv.Itoa(argIndex+1)
	args = append(args, limit, offset)

	rows, err := database.DB.Query(query, args...)
	if err != nil {
		utils.LogError("Error querying notifications", err)
		utils.WriteInternalError(w, err)
		return
	}
	defer func() {
		if err := rows.Close(); err != nil {
			utils.LogError("Error closing rows", err)
		}
	}()

	var notifications []models.Notification
	for rows.Next() {
		var notification models.Notification
		var dataBytes []byte

		err := rows.Scan(
			&notification.ID, &notification.UserID, &notification.Type,
			&notification.Title, &notification.Message, &dataBytes,
			&notification.IsRead, &notification.IsSent, &notification.CreatedAt,
		)
		if err != nil {
			utils.LogError("Error scanning notification row", err)
			continue
		}

		// Parse JSON data
		if dataBytes != nil {
			if err := json.Unmarshal(dataBytes, &notification.Data); err != nil {
				utils.LogError("Error unmarshaling notification data", err)
			}
		}

		notifications = append(notifications, notification)
	}

	utils.WriteSuccessResponse(w, "Notifications retrieved successfully", map[string]interface{}{
		"notifications": notifications,
		"total":         len(notifications),
		"limit":         limit,
		"offset":        offset,
	})
}

// MarkAsRead marks a specific notification as read
// @Summary Mark notification as read
// @Description Mark a specific notification as read for the current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param notificationID path string true "Notification ID"
// @Success 200 {object} models.APIResponse "Notification marked as read"
// @Failure 400 {object} models.ErrorResponse "Bad request - missing notification ID"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 404 {object} models.ErrorResponse "Notification not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/{notificationID}/read [put]
func (h *Handler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get notification ID from URL parameter
	notificationID := r.URL.Query().Get("notificationID")
	if notificationID == "" {
		utils.WriteErrorResponse(w, "Notification ID is required", http.StatusBadRequest)
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// Update notification as read
	query := `
		UPDATE notifications 
		SET is_read = true 
		WHERE id = $1 AND user_id = $2 AND is_read = false
	`

	result, err := database.DB.Exec(query, notificationID, user.UserID)
	if err != nil {
		utils.LogError("Error marking notification as read", err)
		utils.WriteInternalError(w, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.LogError("Error getting rows affected", err)
		utils.WriteInternalError(w, err)
		return
	}

	if rowsAffected == 0 {
		utils.WriteErrorResponse(w, "Notification not found or already read", http.StatusNotFound)
		return
	}

	utils.WriteSuccessResponse(w, "Notification marked as read", nil)
}

// MarkAllAsRead marks all notifications as read for the user
// @Summary Mark all notifications as read
// @Description Mark all notifications as read for the current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=object{notifications_updated=int}} "All notifications marked as read"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/read-all [put]
func (h *Handler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
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

	// Update all notifications as read
	query := `
		UPDATE notifications 
		SET is_read = true 
		WHERE user_id = $1 AND is_read = false
	`

	result, err := database.DB.Exec(query, user.UserID)
	if err != nil {
		utils.LogError("Error marking all notifications as read", err)
		utils.WriteInternalError(w, err)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		utils.LogError("Error getting rows affected", err)
		utils.WriteInternalError(w, err)
		return
	}

	utils.WriteSuccessResponse(w, "All notifications marked as read", map[string]interface{}{
		"notifications_updated": rowsAffected,
	})
}

// GetUnreadCount returns the count of unread notifications
// @Summary Get unread notification count
// @Description Get the count of unread notifications for the current user
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=object{unread_count=int}} "Unread count retrieved successfully"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/unread-count [get]
func (h *Handler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
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

	// Count unread notifications
	var unreadCount int
	query := `
		SELECT COUNT(*)
		FROM notifications
		WHERE user_id = $1 AND is_read = false
	`

	err := database.DB.QueryRow(query, user.UserID).Scan(&unreadCount)
	if err != nil {
		utils.LogError("Error counting unread notifications", err)
		utils.WriteInternalError(w, err)
		return
	}

	utils.WriteSuccessResponse(w, "Unread count retrieved", map[string]interface{}{
		"unread_count": unreadCount,
	})
}
