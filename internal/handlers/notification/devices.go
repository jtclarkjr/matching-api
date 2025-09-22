package notification

import (
	"net/http"

	"matching-api/internal/database"
	"matching-api/internal/middleware"
	"matching-api/internal/models"
	"matching-api/pkg/utils"
)

// RegisterDevice registers a device for push notifications
// @Summary Register device for push notifications
// @Description Register a device token to receive push notifications
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.RegisterDeviceRequest true "Device registration data"
// @Success 201 {object} models.APIResponse{data=models.DeviceToken} "Device registered successfully"
// @Failure 400 {object} models.ErrorResponse "Bad request - validation failed"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/devices [post]
func (h *Handler) RegisterDevice(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Parse and validate request body
	var req models.RegisterDeviceRequest
	if err := utils.ParseAndValidateJSON(r, &req); err != nil {
		utils.WriteValidationError(w, err)
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// Check if device token already exists for this user
	var existingID string
	checkQuery := `
		SELECT id FROM device_tokens 
		WHERE user_id = $1 AND token = $2
	`
	err := database.DB.QueryRow(checkQuery, user.UserID, req.Token).Scan(&existingID)

	if err == nil {
		// Token already exists, update it
		updateQuery := `
			UPDATE device_tokens 
			SET platform = $1, is_active = true, last_used = NOW()
			WHERE id = $2
		`
		_, err = database.DB.Exec(updateQuery, req.Platform, existingID)
		if err != nil {
			utils.LogError("Error updating existing device token", err)
			utils.WriteInternalError(w, err)
			return
		}

		utils.WriteSuccessResponse(w, "Device token updated", map[string]interface{}{
			"device_id": existingID,
			"status":    "updated",
		})
		return
	}

	// Insert new device token
	var deviceID string
	insertQuery := `
		INSERT INTO device_tokens (id, user_id, token, platform, is_active, last_used, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, true, NOW(), NOW())
		RETURNING id
	`
	err = database.DB.QueryRow(insertQuery, user.UserID, req.Token, req.Platform).Scan(&deviceID)
	if err != nil {
		utils.LogError("Error registering device token", err)
		utils.WriteInternalError(w, err)
		return
	}

	utils.WriteCreated(w, "Device registered for notifications", map[string]interface{}{
		"device_id": deviceID,
		"status":    "created",
	})
}

// UnregisterDevice unregisters a device from push notifications
// @Summary Unregister device from push notifications
// @Description Remove a device token from receiving push notifications
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param tokenID path string true "Device token ID"
// @Success 200 {object} models.APIResponse "Device unregistered from notifications"
// @Failure 400 {object} models.ErrorResponse "Bad request - missing token ID"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 404 {object} models.ErrorResponse "Device token not found"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/devices/{tokenID} [delete]
func (h *Handler) UnregisterDevice(w http.ResponseWriter, r *http.Request) {
	// Get user from context
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		utils.WriteUnauthorized(w, "User not found in context")
		return
	}

	// Get token ID from URL parameter
	tokenID := r.URL.Query().Get("tokenID")
	if tokenID == "" {
		utils.WriteErrorResponse(w, "Token ID is required", http.StatusBadRequest)
		return
	}

	// Get database connection
	if database.DB == nil {
		utils.WriteInternalError(w, nil)
		return
	}

	// Delete or deactivate the device token
	deleteQuery := `
		UPDATE device_tokens 
		SET is_active = false 
		WHERE id = $1 AND user_id = $2
	`

	result, err := database.DB.Exec(deleteQuery, tokenID, user.UserID)
	if err != nil {
		utils.LogError("Error unregistering device token", err)
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
		utils.WriteErrorResponse(w, "Device token not found", http.StatusNotFound)
		return
	}

	utils.WriteSuccessResponse(w, "Device unregistered from notifications", nil)
}

// TestNotification sends a test notification
// @Summary Send test notification
// @Description Send a test notification to verify notification setup
// @Tags Notifications
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.APIResponse{data=object{notification_id=string,message=string}} "Test notification sent"
// @Failure 401 {object} models.ErrorResponse "Unauthorized - invalid token"
// @Failure 500 {object} models.ErrorResponse "Internal server error"
// @Router /notifications/test [post]
func (h *Handler) TestNotification(w http.ResponseWriter, r *http.Request) {
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

	// Create a test notification in the database
	insertQuery := `
		INSERT INTO notifications (id, user_id, type, title, message, data, is_read, is_sent, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, false, true, NOW())
		RETURNING id
	`

	var notificationID string
	testData := `{"test": true, "timestamp": "` + "now" + `"}`

	err := database.DB.QueryRow(insertQuery,
		user.UserID,
		"promotional",
		"Test Notification",
		"This is a test notification to verify your notification settings are working correctly.",
		testData,
	).Scan(&notificationID)

	if err != nil {
		utils.LogError("Error creating test notification", err)
		utils.WriteInternalError(w, err)
		return
	}

	// In a real implementation, you would also send push notifications to registered devices
	// For now, we'll just create the database record

	utils.WriteSuccessResponse(w, "Test notification sent", map[string]interface{}{
		"notification_id": notificationID,
		"message":         "Test notification created successfully. Check your notifications list.",
	})
}
