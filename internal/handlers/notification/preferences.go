package notification

import (
	"database/sql"
	"net/http"

	"matching-api/internal/database"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// GetPreferences retrieves notification preferences for the user
// @Summary Get notification preferences
// @Description Retrieve the current user's notification preferences
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=object{preferences=models.NotificationPreferences}} "Notification preferences retrieved"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/preferences [get]
func (h *Handler) GetPreferences(w http.ResponseWriter, r *http.Request) {
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

	// Query user preferences
	var preferences models.NotificationPreferences
	query := `
		SELECT id, user_id, push_enabled, email_enabled, match_notifications, 
		       message_notifications, profile_notifications, marketing_emails, 
		       created_at, updated_at
		FROM notification_preferences
		WHERE user_id = $1
	`

	err := database.DB.QueryRow(query, user.UserID).Scan(
		&preferences.ID, &preferences.UserID, &preferences.PushEnabled,
		&preferences.EmailEnabled, &preferences.MatchNotifications,
		&preferences.MessageNotifications, &preferences.ProfileNotifications,
		&preferences.MarketingEmails, &preferences.CreatedAt, &preferences.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Create default preferences if none exist
			preferences = models.NotificationPreferences{
				UserID:               user.UserID,
				PushEnabled:          true,
				EmailEnabled:         true,
				MatchNotifications:   true,
				MessageNotifications: true,
				ProfileNotifications: true,
				MarketingEmails:      false,
			}

			// Insert default preferences
			insertQuery := `
				INSERT INTO notification_preferences 
				(id, user_id, push_enabled, email_enabled, match_notifications, 
				 message_notifications, profile_notifications, marketing_emails, created_at, updated_at)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
				RETURNING id, created_at, updated_at
			`

			err = database.DB.QueryRow(insertQuery,
				preferences.UserID, preferences.PushEnabled, preferences.EmailEnabled,
				preferences.MatchNotifications, preferences.MessageNotifications,
				preferences.ProfileNotifications, preferences.MarketingEmails,
			).Scan(&preferences.ID, &preferences.CreatedAt, &preferences.UpdatedAt)

			if err != nil {
				utils.LogError("Error creating default notification preferences", err)
				utils.WriteInternalError(w, err)
				return
			}
		} else {
			utils.LogError("Error querying notification preferences", err)
			utils.WriteInternalError(w, err)
			return
		}
	}

	utils.WriteSuccessResponse(w, "Notification preferences retrieved", map[string]interface{}{
		"preferences": preferences,
	})
}

// UpdatePreferences updates notification preferences for the user
// @Summary Update notification preferences
// @Description Update the current user's notification preferences
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UpdateNotificationPreferencesRequest true "Notification preferences update data"
// @Success 200 {object} models.APIResponse "Notification preferences updated"
// @Failure 400 {object} models.ErrorResponse "Bad request - validation failed"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 404 {object} models.ErrorResponse "Notification preferences not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/preferences [put]
func (h *Handler) UpdatePreferences(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Parse and validate request body
	var req models.UpdateNotificationPreferencesRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// Build dynamic update query
	setParts := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.PushEnabled != nil {
		setParts = append(setParts, "push_enabled = $"+string(rune(argIndex+'0')))
		args = append(args, *req.PushEnabled)
		argIndex++
	}
	if req.EmailEnabled != nil {
		setParts = append(setParts, "email_enabled = $"+string(rune(argIndex+'0')))
		args = append(args, *req.EmailEnabled)
		argIndex++
	}
	if req.MatchNotifications != nil {
		setParts = append(setParts, "match_notifications = $"+string(rune(argIndex+'0')))
		args = append(args, *req.MatchNotifications)
		argIndex++
	}
	if req.MessageNotifications != nil {
		setParts = append(setParts, "message_notifications = $"+string(rune(argIndex+'0')))
		args = append(args, *req.MessageNotifications)
		argIndex++
	}
	if req.ProfileNotifications != nil {
		setParts = append(setParts, "profile_notifications = $"+string(rune(argIndex+'0')))
		args = append(args, *req.ProfileNotifications)
		argIndex++
	}
	if req.MarketingEmails != nil {
		setParts = append(setParts, "marketing_emails = $"+string(rune(argIndex+'0')))
		args = append(args, *req.MarketingEmails)
		argIndex++
	}

	if len(setParts) == 0 {
		utils.WriteErrorResponse(w, "No preferences to update", http.StatusBadRequest)
		return
	}

	// Add updated_at and user_id
	setParts = append(setParts, "updated_at = NOW()")
	args = append(args, user.UserID)

	query := "UPDATE notification_preferences SET " +
		setParts[0] + ", " + setParts[1] + " WHERE user_id = $" + string(rune(argIndex+'0'))

	result, err := database.DB.Exec(query, args...)
	if err != nil {
		utils.LogError("Error updating notification preferences", err)
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
		utils.WriteErrorResponse(w, "Notification preferences not found", http.StatusNotFound)
		return
	}

	utils.WriteSuccessResponse(w, "Notification preferences updated", nil)
}
