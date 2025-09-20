package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"matching-api/internal/models"
)

// WriteJSONResponse writes a JSON response to the http.ResponseWriter
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			log.Printf("Error encoding JSON response: %v", err)
		}
	}
}

// WriteSuccessResponse writes a successful API response
func WriteSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	response := models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	WriteJSONResponse(w, http.StatusOK, response)
}

// WriteErrorResponse writes an error response
func WriteErrorResponse(w http.ResponseWriter, message string, statusCode int) {
	response := models.ErrorResponse{
		Success: false,
		Error:   message,
		Code:    statusCode,
	}
	WriteJSONResponse(w, statusCode, response)
}

// WriteValidationError writes a validation error response
func WriteValidationError(w http.ResponseWriter, err error) {
	WriteErrorResponse(w, "Validation error: "+err.Error(), http.StatusBadRequest)
}

// WriteInternalError writes an internal server error response
func WriteInternalError(w http.ResponseWriter, err error) {
	WriteErrorResponse(w, "Internal server error", http.StatusInternalServerError)
}

// WriteUnauthorized writes an unauthorized error response
func WriteUnauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Unauthorized"
	}
	WriteErrorResponse(w, message, http.StatusUnauthorized)
}

// WriteForbidden writes a forbidden error response
func WriteForbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Forbidden"
	}
	WriteErrorResponse(w, message, http.StatusForbidden)
}

// WriteNotFound writes a not found error response
func WriteNotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	WriteErrorResponse(w, message, http.StatusNotFound)
}

// WriteCreated writes a created response
func WriteCreated(w http.ResponseWriter, message string, data interface{}) {
	response := models.APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}
	WriteJSONResponse(w, http.StatusCreated, response)
}

// LogError logs an error message with context
func LogError(message string, err error) {
	if err != nil {
		log.Printf("%s: %v", message, err)
	} else {
		log.Printf("%s", message)
	}
}
